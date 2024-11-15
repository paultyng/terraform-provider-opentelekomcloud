package hss

import (
	"context"
	"fmt"
	"log"
	"strconv"

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

func DataSourceHostGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHostGroupsRead,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"host_num": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"risk_host_num": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"unprotect_host_num": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"groups": {
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
						"host_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"risk_host_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unprotect_host_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"host_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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

func dataSourceHostGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	allHostGroups, err := queryHostGroups(client, d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error querying OpenTelekomCloud HSS host groups: %s", err)
	}

	if len(allHostGroups) == 0 {
		log.Printf("[DEBUG] No groups in OpenTelekomCloud found")
	}

	uuId, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}

	d.SetId(uuId)
	targetGroups := filterHostGroups(allHostGroups, d)
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("groups", flattenHostGroups(targetGroups)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func filterHostGroups(groups []host.HostGroupResp, d *schema.ResourceData) []host.HostGroupResp {
	if len(groups) == 0 {
		return nil
	}

	rst := make([]host.HostGroupResp, 0, len(groups))
	for _, v := range groups {
		if groupID, ok := d.GetOk("group_id"); ok &&
			fmt.Sprint(groupID) != v.ID {
			continue
		}

		if hostNum, ok := d.GetOk("host_num"); ok &&
			hostNum.(string) != strconv.Itoa(v.HostNum) {
			continue
		}

		if riskHostNum, ok := d.GetOk("risk_host_num"); ok &&
			riskHostNum.(string) != strconv.Itoa(v.RiskHostNum) {
			continue
		}

		if unprotectHostNum, ok := d.GetOk("unprotect_host_num"); ok &&
			unprotectHostNum.(string) != strconv.Itoa(v.UnprotectHostNum) {
			continue
		}

		rst = append(rst, v)
	}

	return rst
}

func flattenHostGroups(groups []host.HostGroupResp) []interface{} {
	if len(groups) == 0 {
		return nil
	}

	rst := make([]interface{}, 0, len(groups))
	for _, v := range groups {
		rst = append(rst, map[string]interface{}{
			"id":                 v.ID,
			"name":               v.Name,
			"host_num":           v.HostNum,
			"risk_host_num":      v.RiskHostNum,
			"unprotect_host_num": v.UnprotectHostNum,
			"host_ids":           v.HostIds,
		})
	}

	return rst
}
