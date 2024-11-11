package ddm

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	ddmv2instances "github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v2/instances"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceDdmFlavorsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDdmFlavorsV1Read,

		Schema: map[string]*schema.Schema{
			"engine_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"flavor_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"flavors": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"iaas_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cpu": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"memory": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"max_connections": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"server_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"architecture": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"az_status": {
										Type:     schema.TypeMap,
										Computed: true,
									},
								},
							},
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

func dataSourceDdmFlavorsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	queryNodeClassesResponse, err := ddmv2instances.QueryNodeClasses(client, ddmv2instances.QueryNodeClassesOpts{
		EngineId: d.Get("engine_id").(string),
	})
	if err != nil {
		return fmterr.Errorf("error fetching DDM node classes: %w", err)
	}

	flavorGroupsRaw := queryNodeClassesResponse.ComputeFlavorGroups
	log.Printf("[DEBUG] Retrieved DDM node class info: %#v", flavorGroupsRaw)

	d.SetId(d.Get("engine_id").(string))
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("engine_id", d.Id()),
	)

	var flavourGroups []map[string]interface{}
	for _, flavorGroupRaw := range flavorGroupsRaw {
		flavorGroup := make(map[string]interface{})
		flavorGroup["type"] = flavorGroupRaw.GroupType
		var flavorList []map[string]interface{}
		for _, flavorRaw := range flavorGroupRaw.ComputeFlavors {
			flavor := make(map[string]interface{})
			flavor["id"] = flavorRaw.ID
			flavor["type_code"] = flavorRaw.TypeCode
			flavor["code"] = flavorRaw.Code
			flavor["iaas_code"] = flavorRaw.IaaSCode
			flavor["cpu"] = flavorRaw.CPU
			flavor["memory"] = flavorRaw.Mem
			flavor["max_connections"] = flavorRaw.MaxConnections
			flavor["server_type"] = flavorRaw.ServerType
			flavor["architecture"] = flavorRaw.Architecture
			flavor["az_status"] = flavorRaw.AZStatus
			flavorList = append(flavorList, flavor)
		}
		flavorGroup["flavors"] = flavorList
		flavourGroups = append(flavourGroups, flavorGroup)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("flavor_groups", flavourGroups),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}
