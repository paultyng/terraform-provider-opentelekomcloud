package css

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/snapshots"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceCssSnapshotConfigurationV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCssSnapshotConfigurationV1,
		ReadContext:   readCssSnapshotConfigurationV1,
		UpdateContext: updateCssSnapshotConfigurationV1,
		DeleteContext: deleteCssSnapshotConfigurationV1,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"automatic": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"configuration", "creation_policy"},
			},
			"configuration": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"automatic"},
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"agency": {
							Type:     schema.TypeString,
							Required: true,
						},
						"kms_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"base_path": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"creation_policy": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"automatic"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"period": {
							Type:     schema.TypeString,
							Required: true,
						},
						"keepday": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"delete_auto": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func createCssSnapshotConfigurationV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	clusterID := d.Get("cluster_id").(string)
	d.SetId(clusterID)

	if d.Get("automatic").(bool) {
		if err := snapshots.Enable(client, d.Id()); err != nil {
			return fmterr.Errorf("error using automatic config for snapshots: %w", err)
		}
		return nil
	}

	if d.Get("configuration.#") != 0 {
		if err := updateSnapshotConfiguration(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.Get("creation_policy.#") != 0 {
		if err := updateSnapshotPolicy(client, d); err != nil {
			return diag.FromErr(err)
		}
	}
	return readCssSnapshotConfigurationV1(ctx, d, meta)
}

func readCssSnapshotConfigurationV1(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}
	clusterID := d.Id()

	info, err := snapshots.PolicyGet(client, clusterID)
	if err != nil {
		return fmterr.Errorf("error retrieving CSS cluster automatic snapshot configuration")
	}
	configuration := []map[string]interface{}{{
		"bucket":    info.Bucket,
		"agency":    info.Agency,
		"kms_id":    info.SnapshotCmkID,
		"base_path": info.BasePath,
	}}
	creation := []map[string]interface{}{{
		"prefix":      info.Prefix,
		"period":      info.Period,
		"keepday":     info.KeepDay,
		"enable":      info.Enable == "true",
		"delete_auto": d.Get("creation_policy.0.delete_auto"),
	}}
	mErr := multierror.Append(
		d.Set("configuration", configuration),
		d.Set("creation_policy", creation),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		fmterr.Errorf("error setting snapshot configuration fields: %w", err)
	}

	return nil
}

func updateCssSnapshotConfigurationV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	if d.Get("automatic").(bool) {
		if err := snapshots.Enable(client, d.Id()); err != nil {
			return fmterr.Errorf("error using automatic config for snapshots: %w", err)
		}
		return nil
	}
	if d.HasChange("configuration") {
		if err := updateSnapshotConfiguration(client, d); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("creation_policy") {
		if err := updateSnapshotPolicy(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return readCssSnapshotConfigurationV1(ctx, d, meta)
}

func deleteCssSnapshotConfigurationV1(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}
	clusterID := d.Id()

	if err := snapshots.Disable(client, clusterID); err != nil {
		return fmterr.Errorf("error disabling automatic snapshots: %w", err)
	}

	return nil
}

func updateSnapshotPolicy(client *golangsdk.ServiceClient, d *schema.ResourceData) error {
	policyOpts := snapshots.PolicyCreateOpts{
		Prefix:     d.Get("creation_policy.0.prefix").(string),
		Period:     d.Get("creation_policy.0.period").(string),
		KeepDay:    d.Get("creation_policy.0.keepday").(int),
		Enable:     strconv.FormatBool(d.Get("creation_policy.0.enable").(bool)),
		DeleteAuto: strconv.FormatBool(d.Get("creation_policy.0.delete_auto").(bool)),
	}

	clusterID := d.Get("cluster_id").(string)
	err := snapshots.PolicyCreate(client, policyOpts, clusterID)
	if err != nil {
		return fmt.Errorf("error creating snapshot creating policy: %w", err)
	}
	return nil
}

func updateSnapshotConfiguration(client *golangsdk.ServiceClient, d *schema.ResourceData) error {
	opts := snapshots.UpdateConfigurationOpts{
		Bucket:        d.Get("configuration.0.bucket").(string),
		Agency:        d.Get("configuration.0.agency").(string),
		BasePath:      d.Get("configuration.0.base_path").(string),
		SnapshotCmkID: d.Get("configuration.0.kms_id").(string),
	}
	err := snapshots.UpdateConfiguration(client, d.Id(), opts)
	if err != nil {
		return fmt.Errorf("error updating cluster automatic snapshot configuration: %w", err)
	}
	return nil
}
