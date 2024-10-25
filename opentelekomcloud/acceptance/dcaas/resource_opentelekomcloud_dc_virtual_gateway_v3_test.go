package dcaas

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	gateway "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/virtual-gateway"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getVirtualGatewayFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.DCaaSV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DCaaS v3 client: %s", err)
	}
	return gateway.Get(c, state.Primary.ID)
}

func TestAccVirtualGateway_basic(t *testing.T) {
	var (
		gw gateway.VirtualGateway

		rName      = "opentelekomcloud_dc_virtual_gateway_v3.gw"
		name       = fmt.Sprintf("dc_acc_gw%s", acctest.RandString(5))
		updateName = fmt.Sprintf("dc_acc_gw_up%s", acctest.RandString(5))
	)

	rc := common.InitResourceCheck(
		rName,
		&gw,
		getVirtualGatewayFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayV3_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "local_ep_group.0", "192.168.0.0/16"),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "description", "Created by acc test"),
					resource.TestCheckResourceAttrSet(rName, "asn"),
					resource.TestCheckResourceAttrSet(rName, "status"),
				),
			},
			{
				Config: testAccVirtualGatewayV3_update(updateName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "local_ep_group.0", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(rName, "name", updateName),
					resource.TestCheckResourceAttr(rName, "description", ""),
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

//  private_ip         = cidrhost(data.opentelekomcloud_vpc_subnet_v1.shared_subnet.cidr, 6)

func testAccVirtualGatewayV3_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_dc_virtual_gateway_v3" "gw" {
  vpc_id      = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  name        = "%s"
  description = "Created by acc test"

  local_ep_group = [
    "192.168.0.0/16",
  ]
}
`, common.DataSourceSubnet, name)
}

func testAccVirtualGatewayV3_update(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_dc_virtual_gateway_v3" "gw" {
  vpc_id = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  name   = "%s"

  local_ep_group = [
    "192.168.0.0/24",
  ]
}
`, common.DataSourceSubnet, name)
}
