package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/apigw/v2/gateway"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getInstanceFeatureFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.APIGWV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating APIG v2 client: %s", err)
	}
	gatewayId := state.Primary.Attributes["gateway_id"]
	features, err := gateway.ListGatewayFeatures(client, gateway.ListFeaturesOpts{
		GatewayID: gatewayId,
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}
	if len(features) < 1 {
		return nil, err
	}
	var f gateway.FeatureResp
	for _, feature := range features {
		if feature.Name == state.Primary.ID {
			f = feature
		}
	}
	return f, err
}

func TestAccInstanceFeature_basic(t *testing.T) {
	var (
		feature gateway.FeatureResp
		rName   = "opentelekomcloud_apigw_gateway_feature_v2.feat"
	)

	rc := common.InitResourceCheck(
		rName,
		&feature,
		getInstanceFeatureFunc,
	)

	// Avoid CheckDestroy because this resource already exists and does not need to be deleted.
	// lintignore:AT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckApigw(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceFeature_basic(),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", "ratelimit"),
					resource.TestCheckResourceAttr(rName, "enabled", "true"),
					resource.TestCheckResourceAttr(rName, "config", "{\"api_limits\":200}"),
				),
			},
			{
				Config: testAccInstanceFeature_basicUpdate(),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", "ratelimit"),
					resource.TestCheckResourceAttr(rName, "config", "{\"api_limits\":300}"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccInstanceFeatureResourceImportStateFunc(rName),
			},
		},
	})
}

func testAccInstanceFeatureResourceImportStateFunc(rName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return "", fmt.Errorf("resource (%s) not found: %s", rName, rs)
		}
		gatewayId := rs.Primary.Attributes["gateway_id"]
		featureName := rs.Primary.ID
		if gatewayId == "" || featureName == "" {
			return "", fmt.Errorf("missing some attributes, want '<gateway_id>/<name>', but '%s/%s'",
				gatewayId, featureName)
		}
		return fmt.Sprintf("%s/%s", gatewayId, featureName), nil
	}
}

func testAccInstanceFeature_basic() string {
	return fmt.Sprintf(`
resource "opentelekomcloud_apigw_gateway_feature_v2" "feat" {
  gateway_id = "%[1]s"
  name       = "ratelimit"
  enabled    = true

  config = jsonencode({
    api_limits = 200
  })
}
`, env.OS_APIGW_GATEWAY_ID)
}

func testAccInstanceFeature_basicUpdate() string {
	return fmt.Sprintf(`
resource "opentelekomcloud_apigw_gateway_feature_v2" "feat" {
  gateway_id = "%[1]s"
  name       = "ratelimit"
  enabled    = true

  config = jsonencode({
    api_limits = 300
  })
}
`, env.OS_APIGW_GATEWAY_ID)
}
