package hss

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/host"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceHosts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHostsRead,

		Schema: map[string]*schema.Schema{
			"host_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"os_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"agent_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protect_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protect_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protect_charging_mode": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"detect_result": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asset_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"os_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"agent_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"agent_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protect_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protect_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protect_charging_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"detect_result": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"asset_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"asset_risk_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vulnerability_risk_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"baseline_risk_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"intrusion_risk_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
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

func dataSourceHostsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	opts := host.ListHostOpts{
		Limit:         200,
		Version:       d.Get("protect_version").(string),
		AgentStatus:   d.Get("agent_status").(string),
		DetectResult:  d.Get("detect_result").(string),
		HostName:      d.Get("name").(string),
		HostID:        d.Get("host_id").(string),
		HostStatus:    d.Get("status").(string),
		OsType:        d.Get("os_type").(string),
		ProtectStatus: d.Get("protect_status").(string),
		GroupId:       d.Get("group_id").(string),
		PolicyGroupId: d.Get("policy_group_id").(string),
		ChargingMode:  d.Get("protect_charging_mode").(string),
		Refresh:       false,
		AboveVersion:  false,
		AssetValue:    d.Get("asset_value").(string),
	}
	allHosts, err := host.ListHost(client, opts)
	if err != nil {
		return diag.Errorf("unable to list OpenTelekomCloud HSS hosts: %s", err)
	}

	if len(allHosts) == 0 {
		log.Printf("[DEBUG] No hosts in OpenTelekomCloud found")
	}

	uuId, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}

	d.SetId(uuId)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("hosts", flattenHosts(allHosts)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenHosts(hosts []host.HostResp) []interface{} {
	if len(hosts) == 0 {
		return nil
	}

	rst := make([]interface{}, 0, len(hosts))
	for _, v := range hosts {
		rst = append(rst, map[string]interface{}{
			"id":                     v.ID,
			"name":                   v.Name,
			"status":                 v.HostStatus,
			"os_type":                v.OsType,
			"agent_id":               v.AgentId,
			"agent_status":           v.AgentStatus,
			"protect_status":         v.ProtectStatus,
			"protect_version":        v.Version,
			"protect_charging_mode":  v.ChargingMode,
			"resource_id":            v.ResourceId,
			"detect_result":          v.DetectResult,
			"group_id":               v.GroupId,
			"policy_group_id":        v.PolicyGroupId,
			"asset_value":            v.AssetValue,
			"private_ip":             v.PrivateIp,
			"public_ip":              v.PublicIp,
			"asset_risk_num":         v.Asset,
			"vulnerability_risk_num": v.Vulnerability,
			"baseline_risk_num":      v.Baseline,
			"intrusion_risk_num":     v.Intrusion,
		})
	}

	return rst
}
