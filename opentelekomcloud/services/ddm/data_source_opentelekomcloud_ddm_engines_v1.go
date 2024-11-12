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

func DataSourceDdmEnginesV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDdmEnginesV1Read,

		Schema: map[string]*schema.Schema{
			"engines": {
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
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zones": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"favored": {
										Type:     schema.TypeBool,
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

func dataSourceDdmEnginesV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	enginesRaw, err := ddmv2instances.QueryEngineInfo(client, ddmv2instances.QueryEngineOpts{})
	if err != nil {
		return fmterr.Errorf("error fetching DDM engines: %w", err)
	}

	log.Printf("[DEBUG] Retrieved DDM engines info: %#v", enginesRaw)

	d.SetId(config.GetRegion(d))

	mErr := multierror.Append(nil,
		d.Set("region", d.Id()),
	)

	var enginesList []map[string]interface{}
	for _, engineRaw := range enginesRaw {
		engine := make(map[string]interface{})
		engine["id"] = engineRaw.ID
		engine["name"] = engineRaw.Name
		engine["version"] = engineRaw.Version
		var azList []map[string]interface{}
		for _, azRaw := range engineRaw.SupportAzs {
			az := make(map[string]interface{})
			az["name"] = azRaw.Name
			az["code"] = azRaw.Code
			az["favored"] = azRaw.Favored
			azList = append(azList, az)
		}
		engine["availability_zones"] = azList
		enginesList = append(enginesList, engine)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("engines", enginesList),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}
