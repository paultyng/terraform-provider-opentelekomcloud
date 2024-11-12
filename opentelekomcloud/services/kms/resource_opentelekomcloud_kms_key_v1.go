package kms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

const (
	WaitingForEnableState = "1"
	EnabledState          = "2"
	DisabledState         = "3"
	PendingDeletionState  = "4"
	WaitingImportState    = "5"
)

func ResourceKmsKeyV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKmsKeyV1Create,
		ReadContext:   resourceKmsKeyV1Read,
		UpdateContext: resourceKmsKeyV1Update,
		DeleteContext: resourceKmsKeyV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key_alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"realm": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_cancel_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"scheduled_deletion_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"default_key_flag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pending_days": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "7",
			},
			"rotation_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"rotation_enabled"},
				ValidateFunc: validation.IntBetween(30, 365),
			},
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"rotation_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": common.TagsSchema(),
		},
	}
}

func resourceKmsKeyValidation(d *schema.ResourceData) error {
	_, rotationEnabled := d.GetOk("rotation_enabled")
	_, hasInterval := d.GetOk("rotation_interval")

	if !rotationEnabled && hasInterval {
		return fmt.Errorf("invalid arguments: rotation_interval is only valid when rotation is enabled")
	}
	return nil
}

func resourceKmsKeyV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	if err := resourceKmsKeyValidation(d); err != nil {
		return fmterr.Errorf("error validating KMS key: %s", err)
	}

	createOpts := keys.CreateOpts{
		KeyAlias:       d.Get("key_alias").(string),
		KeyDescription: d.Get("key_description").(string),
		Realm:          d.Get("realm").(string),
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	key, err := keys.Create(client, createOpts)
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud key: %s", err)
	}
	log.Printf("[INFO] Key ID: %s", key.KeyID)

	if d.Get("allow_cancel_deletion").(bool) {
		keyGet, err := keys.Get(client, key.KeyID)
		if err != nil {
			return diag.FromErr(err)
		}
		if keyGet.KeyState == PendingDeletionState {
			_, err = keys.CancelDelete(client, key.KeyID)
			if err != nil {
				return fmterr.Errorf("error disabling deletion of key: %s", err)
			}

			key, err := keys.EnableKey(client, key.KeyID)
			if err != nil {
				return fmterr.Errorf("error enabling key: %s", err)
			}
			if key.KeyState != EnabledState {
				return fmterr.Errorf("error enabling key, the key state is: %s", key.KeyState)
			}
		}
	}

	// Wait for the key to become enabled.
	log.Printf("[DEBUG] Waiting for key (%s) to become enabled", key.KeyID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{WaitingForEnableState, DisabledState},
		Target:       []string{EnabledState},
		Refresh:      keyV1StateRefreshFunc(client, key.KeyID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        10 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for key (%s) to become ready: %s", key.KeyID, err)
	}

	if !d.Get("is_enabled").(bool) {
		disableKey, err := keys.DisableKey(client, key.KeyID)
		if err != nil {
			return fmterr.Errorf("error disabling key: %s", err)
		}

		if disableKey.KeyState != DisabledState {
			return fmterr.Errorf("error disabling key, the key state is: %s", disableKey.KeyState)
		}
	}

	// enable rotation and change interval if necessary
	if _, ok := d.GetOk("rotation_enabled"); ok {
		rotationOpts := keys.RotationOpts{
			KeyID: key.KeyID,
		}

		keyRotation, err := keys.GetKeyRotationStatus(client, rotationOpts)
		if err != nil {
			return fmterr.Errorf("failed to fetch KMS key rotation status: %s", err)
		}
		if !keyRotation.Enabled {
			err := keys.EnableKeyRotation(client, key.KeyID)
			if err != nil {
				return fmterr.Errorf("failed to enable KMS key rotation: %s", err)
			}

			if i, ok := d.GetOk("rotation_interval"); ok {
				rotationOpts := keys.RotationOpts{
					KeyID:    key.KeyID,
					Interval: i.(int),
				}
				err := keys.UpdateKeyRotationInterval(client, rotationOpts)
				if err != nil {
					return fmterr.Errorf("failed to change KMS key rotation interval: %s", err)
				}
			}
		}
	}

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		tagList := common.ExpandResourceTags(tagRaw)
		if err := tags.Create(client, "kms", key.KeyID, tagList).ExtractErr(); err != nil {
			return fmterr.Errorf("error setting tags of KMS: %s", err)
		}
	}

	// Store the key ID now
	d.SetId(key.KeyID)
	clientCtx := common.CtxWithClient(ctx, client, keyClientV1)
	return resourceKmsKeyV1Read(clientCtx, d, meta)
}

func resourceKmsKeyV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	key, err := keys.Get(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Kms key %s: %+v", d.Id(), key)
	if key.KeyState == PendingDeletionState {
		log.Printf("[WARN] Removing KMS key %s because it's already gone", d.Id())
		d.SetId("")
		return nil
	}

	mErr := multierror.Append(nil,
		d.Set("domain_id", key.DomainID),
		d.Set("key_alias", key.KeyAlias),
		d.Set("realm", key.Realm),
		d.Set("key_description", key.KeyDescription),
		d.Set("creation_date", key.CreationDate),
		d.Set("scheduled_deletion_date", key.ScheduledDeletionDate),
		d.Set("is_enabled", key.KeyState == EnabledState),
		d.Set("default_key_flag", key.DefaultKeyFlag),
		d.Set("expiration_time", key.ExpirationTime),
		d.Set("origin", key.Origin),
	)

	if mErr.ErrorOrNil() != nil {
		return diag.FromErr(mErr)
	}

	// save tags
	resourceTags, err := tags.Get(client, "kms", d.Id()).Extract()
	if err != nil {
		return fmterr.Errorf("error fetching OpenTelekomCloud KMS tags: %s", err)
	}
	tagMap := common.TagsToMap(resourceTags)
	if err := d.Set("tags", tagMap); err != nil {
		return fmterr.Errorf("error saving tags for OpenTelekomCloud KMS: %s", err)
	}

	// save rotation status
	rotationOpts := keys.RotationOpts{
		KeyID: key.KeyID,
	}
	r, err := keys.GetKeyRotationStatus(client, rotationOpts)
	if err == nil {
		_ = d.Set("rotation_enabled", r.Enabled)
		_ = d.Set("rotation_interval", r.Interval)
		_ = d.Set("rotation_number", r.NumberOfRotations)
	} else {
		log.Printf("[WARN] error fetching details about KMS key rotation: %s", err)
	}

	return nil
}

func resourceKmsKeyV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	if d.HasChange("key_alias") {
		updateAliasOpts := keys.UpdateAliasOpts{
			KeyID:    d.Id(),
			KeyAlias: d.Get("key_alias").(string),
		}
		_, err = keys.UpdateAlias(client, updateAliasOpts)
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud key: %s", err)
		}
	}

	if d.HasChange("key_description") {
		updateDesOpts := keys.UpdateDesOpts{
			KeyID:          d.Id(),
			KeyDescription: d.Get("key_description").(string),
		}
		_, err = keys.UpdateDes(client, updateDesOpts)
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud key: %s", err)
		}
	}

	if d.HasChange("is_enabled") {
		key, err := keys.Get(client, d.Id())
		if err != nil {
			return fmterr.Errorf("describeKey got an error: %s", err)
		}

		if d.Get("is_enabled").(bool) && key.KeyState == DisabledState {
			key, err := keys.EnableKey(client, d.Id())
			if err != nil {
				return fmterr.Errorf("error enabling key: %s", err)
			}
			if key.KeyState != EnabledState {
				return fmterr.Errorf("error enabling key, the key state is: %s", key.KeyState)
			}
		}

		if !d.Get("is_enabled").(bool) && key.KeyState == EnabledState {
			key, err := keys.DisableKey(client, d.Id())
			if err != nil {
				return fmterr.Errorf("error disabling key: %s", err)
			}
			if key.KeyState != DisabledState {
				return fmterr.Errorf("error disabling key, the key state is: %s", key.KeyState)
			}
		}
	}

	// update tags
	if d.HasChange("tags") {
		if err := common.UpdateResourceTags(client, d, "kms", d.Id()); err != nil {
			return fmterr.Errorf("error updating tags of KMS %s: %s", d.Id(), err)
		}
	}

	_, rotationEnabled := d.GetOk("rotation_enabled")
	if d.HasChange("rotation_enabled") {
		var rotationErr error
		if rotationEnabled {
			rotationErr = keys.EnableKeyRotation(client, d.Id())
		} else {
			rotationErr = keys.DisableKeyRotation(client, d.Id())
		}

		if rotationErr != nil {
			return fmterr.Errorf("failed to update key rotation status: %s", err)
		}
	}

	if rotationEnabled && d.HasChange("rotation_interval") {
		intervalOpts := keys.RotationOpts{
			KeyID:    d.Id(),
			Interval: d.Get("rotation_interval").(int),
		}
		err := keys.UpdateKeyRotationInterval(client, intervalOpts)
		if err != nil {
			return fmterr.Errorf("failed to change key rotation interval: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV1)
	return resourceKmsKeyV1Read(clientCtx, d, meta)
}

func resourceKmsKeyV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	key, err := keys.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "key")
	}

	deleteOpts := keys.DeleteOpts{
		KeyID: d.Id(),
	}
	if v, ok := d.GetOk("pending_days"); ok {
		deleteOpts.PendingDays = v.(string)
	}

	// It's possible that this key was used as a boot device and is currently
	// in a pending deletion state from when the instance was terminated.
	// If this is true, just move on. It'll eventually delete.
	if key.KeyState != PendingDeletionState {
		rotationOpts := keys.RotationOpts{
			KeyID: d.Id(),
		}
		keyRotation, err := keys.GetKeyRotationStatus(client, rotationOpts)
		if err != nil {
			return fmterr.Errorf("failed to fetch KMS key rotation status: %s", err)
		}
		if keyRotation.Enabled {
			err := keys.DisableKeyRotation(client, d.Id())
			if err != nil {
				return fmterr.Errorf("failed to disable KMS key rotation: %s", err)
			}
		}

		keyDel, err := keys.Delete(client, deleteOpts)
		if err != nil {
			return diag.FromErr(err)
		}

		if keyDel.KeyState != PendingDeletionState {
			return fmterr.Errorf("failed to delete key")
		}
	}

	log.Printf("[DEBUG] KMS Key %s deactivated.", d.Id())
	d.SetId("")
	return nil
}

func keyV1StateRefreshFunc(client *golangsdk.ServiceClient, keyID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := keys.Get(client, keyID)
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return v, PendingDeletionState, nil
			}
			return nil, "", err
		}

		return v, v.KeyState, nil
	}
}
