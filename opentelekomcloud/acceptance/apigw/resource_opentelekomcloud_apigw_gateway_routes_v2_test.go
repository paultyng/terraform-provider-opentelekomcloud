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

func getInstanceRoutesFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.APIGWV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating APIG v2 client: %s", err)
	}
	opts := gateway.ListFeaturesOpts{
		GatewayID: state.Primary.ID,
		Limit:     500,
	}
	resp, err := gateway.ListGatewayFeatures(client, opts)
	if err != nil {
		return nil, fmt.Errorf("error querying feature list: %s", err)
	}

	for _, val := range resp {
		if val.Name == "route" {
			return val, nil
		}
	}
	return nil, fmt.Errorf("error querying feature: route")
}

func TestAccInstanceRoutes_basic(t *testing.T) {
	var (
		feature gateway.FeatureResp
		rName   = "opentelekomcloud_apigw_gateway_routes_v2.rt"
	)

	rc := common.InitResourceCheck(
		rName,
		&feature,
		getInstanceRoutesFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckApigw(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceRoutes_basic(),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(rName, "gateway_id", "opentelekomcloud_apigw_gateway_routes_v2.rt", "id"),
					resource.TestCheckResourceAttr(rName, "nexthops.#", "2"),
				),
			},
			{
				Config: testAccInstanceRoutes_basicUpdate(),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "nexthops.#", "2"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccInstanceRoutes_basic() string {
	return fmt.Sprintf(`
resource "opentelekomcloud_apigw_gateway_routes_v2" "rt" {
  gateway_id = "%s"
  nexthops   = ["172.16.128.0/20", "172.16.0.0/20"]
}
`, env.OS_APIGW_GATEWAY_ID)
}

func testAccInstanceRoutes_basicUpdate() string {
	return fmt.Sprintf(`
resource "opentelekomcloud_apigw_gateway_routes_v2" "rt" {
  gateway_id = "%s"
  nexthops   = ["172.16.64.0/20", "172.16.192.0/20"]
}
`, env.OS_APIGW_GATEWAY_ID)
}
