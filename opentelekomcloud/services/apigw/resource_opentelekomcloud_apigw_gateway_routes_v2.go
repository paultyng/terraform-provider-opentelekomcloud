package apigw

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/apigw/v2/gateway"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceGatewayRoutesV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceRoutesV2Create,
		ReadContext:   resourceInstanceRoutesV2Read,
		UpdateContext: resourceInstanceRoutesV2Update,
		DeleteContext: resourceInstanceRoutesV2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceInstanceRoutesImportState,
		},

		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"nexthops": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func modifyInstanceRoutes(client *golangsdk.ServiceClient, gatewayId string, routes []interface{}) error {
	routeConfig := map[string]interface{}{
		"user_routes": routes,
	}
	routeBytes, err := json.Marshal(routeConfig)
	if err != nil {
		return fmt.Errorf("error parsing routes configuration: %s", err)
	}
	opts := gateway.FeatureOpts{
		GatewayID: gatewayId,
		Name:      "route",
		Enable:    pointerto.Bool(true),
		Config:    string(routeBytes),
	}
	log.Printf("[DEBUG] The modify options of the gateway routes is: %#v", opts)
	_, err = gateway.ConfigureFeature(client, opts)
	if err != nil {
		return err
	}
	return nil
}

func resourceInstanceRoutesV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	var (
		gatewayId = d.Get("gateway_id").(string)
		routes    = d.Get("nexthops").(*schema.Set)
	)
	if err := modifyInstanceRoutes(client, gatewayId, routes.List()); err != nil {
		return diag.Errorf("error creating OpenTelekomCloud APIGW v2 gateway routes: %v", err)
	}
	d.SetId(gatewayId)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceInstanceRoutesV2Read(clientCtx, d, meta)
}

func resourceInstanceRoutesV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	gatewayId := d.Get("gateway_id").(string)

	opts := gateway.ListFeaturesOpts{
		GatewayID: gatewayId,
		// Default value of parameter 'limit' is 20, parameter 'offset' is an invalid parameter.
		// If we omit it, we can only obtain 20 features, other features will be lost.
		Limit: 500,
	}
	resp, err := gateway.ListGatewayFeatures(client, opts)
	if err != nil {
		return diag.Errorf("error querying OpenTelekomCloud APIGW v2 gateway feature list: %s", err)
	}
	log.Printf("[DEBUG] The feature list is: %v", resp)

	var routeConfig string
	for _, val := range resp {
		if val.Name == "route" {
			routeConfig = val.Config
			break
		}
	}
	var result RouteConfig
	err = json.Unmarshal([]byte(routeConfig), &result)
	if err != nil {
		return diag.Errorf("error analyzing routes configuration: %s", err)
	}
	if len(result.UserRoutes) < 1 {
		return common.CheckDeletedDiag(d, golangsdk.ErrDefault404{}, "Instance routes")
	}

	return diag.FromErr(d.Set("nexthops", result.UserRoutes))
}

func resourceInstanceRoutesV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	var (
		gatewayId = d.Get("gateway_id").(string)
		routes    = d.Get("nexthops").(*schema.Set)
	)
	if err := modifyInstanceRoutes(client, gatewayId, routes.List()); err != nil {
		return diag.Errorf("error updating OpenTelekomCloud APIGW v2 gateway routes: %v", err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceInstanceRoutesV2Read(clientCtx, d, meta)
}

func resourceInstanceRoutesV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	gatewayId := d.Get("gateway_id").(string)
	// The expression "{\"user_routes\":null}" has the same result as the expression"{\"user_routes\":[]}".
	if err := modifyInstanceRoutes(client, gatewayId, nil); err != nil {
		return diag.Errorf("error deleting OpenTelekomCloud APIGW v2 gateway routes: %v", err)
	}

	return nil
}

func resourceInstanceRoutesImportState(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	mErr := multierror.Append(nil, d.Set("gateway_id", d.Id()))
	return []*schema.ResourceData{d}, mErr.ErrorOrNil()
}
