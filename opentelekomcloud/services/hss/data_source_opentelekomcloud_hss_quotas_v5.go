package hss

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/quota"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceQuotas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceQuotasRead,

		Schema: map[string]*schema.Schema{
			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"used_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"host_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"charging_mode": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"quotas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"used_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"charging_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expire_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"shared_quota": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": common.TagsSchema(),
					},
				},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceQuotasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	opts := quota.ListOpts{
		Limit:        200,
		Version:      d.Get("version").(string),
		Category:     d.Get("category").(string),
		QuotaStatus:  d.Get("status").(string),
		UsedStatus:   d.Get("used_status").(string),
		HostName:     d.Get("host_name").(string),
		ResourceId:   d.Get("resource_id").(string),
		ChargingMode: d.Get("charging_mode").(string),
	}
	q, err := quota.List(client, opts)
	if err != nil {
		return diag.Errorf("unable to list OpenTelekomCloud HSS quotas: %s", err)
	}

	if len(q) == 0 {
		log.Printf("[DEBUG] No quotas in OpenTelekomCloud found")
	}

	uuId, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}

	d.SetId(uuId)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("quotas", flattenQuotas(q)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenQuotas(quotas []quota.QuotaResp) []interface{} {
	if len(quotas) == 0 {
		return nil
	}

	rst := make([]interface{}, 0, len(quotas))
	for _, v := range quotas {
		rst = append(rst, map[string]interface{}{
			"id":            v.ResourceId,
			"version":       v.Version,
			"status":        v.QuotaStatus,
			"used_status":   v.UsedStatus,
			"host_id":       v.HostId,
			"host_name":     v.HostName,
			"charging_mode": v.ChargingMode,
			"expire_time":   common.FormatTimeStampRFC3339(v.ExpireTime/1000, false),
			"shared_quota":  v.SharedQuota,
			"tags":          common.TagsToMap(v.Tags),
		})
	}

	return rst
}
