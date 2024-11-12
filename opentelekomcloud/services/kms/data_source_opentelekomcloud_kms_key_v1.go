package kms

import (
	"context"
	"log"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceKmsKeyV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKmsKeyV1Read,

		Schema: map[string]*schema.Schema{
			"key_alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"realm": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					WaitingForEnableState,
					EnabledState,
					DisabledState,
					PendingDeletionState,
					WaitingImportState,
				}, true),
			},
			"default_key_flag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scheduled_deletion_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKmsKeyV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.KmsKeyV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	isListKey := true
	nextMarker := ""
	var allKeys []keys.Key
	for isListKey {
		req := keys.ListOpts{
			KeyState: d.Get("key_state").(string),
			Marker:   nextMarker,
		}

		v, err := keys.List(client, req)
		if err != nil {
			return diag.FromErr(err)
		}

		isListKey = v.Truncated == "true"
		nextMarker = v.NextMarker
		allKeys = append(allKeys, v.KeyDetails...)
	}

	keyProperties := map[string]string{}
	if v, ok := d.GetOk("key_description"); ok {
		keyProperties["KeyDescription"] = v.(string)
	}
	if v, ok := d.GetOk("key_id"); ok {
		keyProperties["KeyID"] = v.(string)
	}
	if v, ok := d.GetOk("realm"); ok {
		keyProperties["Realm"] = v.(string)
	}
	if v, ok := d.GetOk("key_alias"); ok {
		keyProperties["KeyAlias"] = v.(string)
	}
	if v, ok := d.GetOk("default_key_flag"); ok {
		keyProperties["DefaultKeyFlag"] = v.(string)
	}
	if v, ok := d.GetOk("domain_id"); ok {
		keyProperties["DomainID"] = v.(string)
	}
	if v, ok := d.GetOk("origin"); ok {
		keyProperties["Origin"] = v.(string)
	}

	if len(allKeys) > 1 && len(keyProperties) > 0 {
		var filteredKeys []keys.Key
		for _, key := range allKeys {
			match := true
			for searchKey, searchValue := range keyProperties {
				r := reflect.ValueOf(&key)
				f := reflect.Indirect(r).FieldByName(searchKey)
				if !f.IsValid() {
					match = false
					break
				}

				keyValue := f.String()
				if searchValue != keyValue {
					match = false
					break
				}
			}

			if match {
				filteredKeys = append(filteredKeys, key)
			}
		}
		allKeys = filteredKeys
	}

	if len(allKeys) < 1 {
		return fmterr.Errorf("your query returned no results. " +
			"Please change your search criteria and try again.")
	}

	if len(allKeys) > 1 {
		return fmterr.Errorf("your query returned more than one result. " +
			"Please try a more specific search criteria.")
	}

	key := allKeys[0]
	log.Printf("[DEBUG] KMS key: %+v", key)

	d.SetId(key.KeyID)
	mErr := multierror.Append(
		d.Set("key_id", key.KeyID),
		d.Set("domain_id", key.DomainID),
		d.Set("key_alias", key.KeyAlias),
		d.Set("realm", key.Realm),
		d.Set("key_description", key.KeyDescription),
		d.Set("creation_date", key.CreationDate),
		d.Set("scheduled_deletion_date", key.ScheduledDeletionDate),
		d.Set("key_state", key.KeyState),
		d.Set("default_key_flag", key.DefaultKeyFlag),
		d.Set("expiration_time", key.ExpirationTime),
		d.Set("origin", key.Origin),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
