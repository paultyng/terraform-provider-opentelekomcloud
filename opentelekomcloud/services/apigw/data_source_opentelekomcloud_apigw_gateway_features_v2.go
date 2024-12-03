package apigw

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/apigw/v2/gateway"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceGatewayFeaturesV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstanceFeaturesV2Read,

		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"features": {
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
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"config": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
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

func dataSourceInstanceFeaturesV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	opts := gateway.ListFeaturesOpts{
		GatewayID: d.Get("gateway_id").(string),
		// Default value of parameter 'limit' is 20, parameter 'offset' is an invalid parameter.
		// If we omit it, we can only obtain 20 features, other features will be lost.
		Limit: 500,
	}
	features, err := gateway.ListGatewayFeatures(client, opts)
	if err != nil {
		return diag.Errorf("error querying OpenTelekomCloud APIGW v2 gateway feature list: %s", err)
	}

	dataSourceId, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}
	d.SetId(dataSourceId)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("features", filterInstanceFeatures(flattenInstanceFeatures(features), d)),
	)
	return diag.FromErr(mErr.ErrorOrNil())
}

func filterInstanceFeatures(all []interface{}, d *schema.ResourceData) []interface{} {
	rst := make([]interface{}, 0, len(all))
	for _, v := range all {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if param, ok := d.GetOk("name"); ok {
			name, ok := vMap["name"].(string)
			if !ok || fmt.Sprint(param) != name {
				continue
			}
		}
		rst = append(rst, vMap)
	}
	return rst
}

func flattenInstanceFeatures(features []gateway.FeatureResp) []interface{} {
	if len(features) < 1 {
		return nil
	}

	result := make([]interface{}, 0, len(features))
	for _, feature := range features {
		updateTime := common.ConvertTimeStrToNanoTimestamp(feature.UpdatedAt)
		result = append(result, map[string]interface{}{
			"id":      feature.ID,
			"name":    feature.Name,
			"enabled": feature.Enabled,
			"config":  feature.Config,
			// If this feature has not been configured, the time format is "0001-01-01T00:00:00Z",
			// the corresponding timestamp is a negative, and this format is uniformly processed as an empty string.
			"updated_at": common.FormatTimeStampRFC3339(updateTime/1000, false),
		})
	}
	return result
}
