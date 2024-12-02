package apigw

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/apigw/v2/gateway"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceGatewayFeatureV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceFeatureV2Create,
		ReadContext:   resourceInstanceFeatureV2Read,
		UpdateContext: resourceInstanceFeatureV2Update,
		DeleteContext: resourceInstanceFeatureV2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceInstanceFeatureImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"config": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func updateFeatureConfiguration(ctx context.Context, client *golangsdk.ServiceClient, d *schema.ResourceData, gatewayId, name string) (*gateway.FeatureResp, error) {
	var resp *gateway.FeatureResp
	var reqErr error
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, reqErr = gateway.ConfigureFeature(client, gateway.FeatureOpts{
			GatewayID: gatewayId,
			Name:      name,
			Enable:    pointerto.Bool(d.Get("enabled").(bool)),
			Config:    d.Get("config").(string),
		})
		isRetry, err := handleOperationError409(reqErr)
		if isRetry {
			// lintignore:R018
			time.Sleep(30 * time.Second)
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	return resp, err
}

func resourceInstanceFeatureV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}
	gatewayId := d.Get("gateway_id").(string)
	name := d.Get("name").(string)
	feature, err := updateFeatureConfiguration(ctx, client, d, gatewayId, name)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud APIGW v2 gateway feature: %s", err)
	}

	d.SetId(feature.Name)
	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceInstanceFeatureV2Read(clientCtx, d, meta)
}

func handleOperationError409(err error) (bool, error) {
	if err == nil {
		return false, nil
	}
	if errCode, ok := err.(golangsdk.ErrUnexpectedResponseCode); ok && errCode.Actual == 409 {
		var apiError interface{}
		if jsonErr := json.Unmarshal(errCode.Body, &apiError); jsonErr != nil {
			return false, jsonErr
		}

		errCode, searchErr := jmespath.Search("error_code", apiError)
		if searchErr != nil {
			return false, err
		}

		// APIG.3711: A configuration parameter can be modified only once per minute.
		if errCode == "APIG.3711" {
			return true, err
		}
	}
	return false, err
}

func resourceInstanceFeatureV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	gatewayId := d.Get("gateway_id").(string)
	features, err := gateway.ListGatewayFeatures(client, gateway.ListFeaturesOpts{
		GatewayID: gatewayId,
		Limit:     500,
	})
	if err != nil {
		// When instance ID not exist, status code is 404, error code id APIG.3030
		return common.CheckDeletedDiag(d, err, "Instance feature configuration")
	}
	if len(features) < 1 {
		return diag.Errorf("error getting OpenTelekomCloud APIGW v2 gateway features: %s", err)
	}
	var f gateway.FeatureResp
	for _, feature := range features {
		if feature.Name == d.Id() {
			f = feature
		}
	}
	mErr := multierror.Append(nil,
		d.Set("name", f.Name),
		d.Set("enabled", f.Enabled),
		d.Set("config", f.Config),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceInstanceFeatureV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	gatewayId := d.Get("gateway_id").(string)
	_, err = updateFeatureConfiguration(ctx, client, d, gatewayId, d.Id())
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud APIGW v2 gateway feature: %s", err)
	}
	if err != nil {
		return diag.Errorf("error updating feature (%s) under specified instance (%s): %s", d.Id(), gatewayId, err)
	}
	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceInstanceFeatureV2Read(clientCtx, d, meta)
}

func resourceInstanceFeatureV2Delete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceInstanceFeatureImportState(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	importedId := d.Id()
	parts := strings.Split(importedId, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format specified for import ID, want <gateway_id>/<name>, but got '%s'", importedId)
	}

	mErr := multierror.Append(
		d.Set("gateway_id", parts[0]),
	)
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, mErr.ErrorOrNil()
}
