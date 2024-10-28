package dcaas

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	virtual_interface "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/virtual-interface"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getVirtualInterfaceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.DCaaSV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DCaaS v3 client: %s", err)
	}
	return virtual_interface.Get(c, state.Primary.ID)
}

func TestAccVirtualInterface_basic(t *testing.T) {
	dcId := os.Getenv("OS_DIRECT_CONNECT_ID")
	if dcId == "" {
		t.Skip("OS_DIRECT_CONNECT_ID should be set for acceptance tests")
	}
	var (
		vi virtual_interface.VirtualInterface

		rName      = "opentelekomcloud_dc_virtual_interface_v3.vi"
		name       = fmt.Sprintf("dc_acc_vi%s", acctest.RandString(5))
		updateName = fmt.Sprintf("dc_acc_vi_up%s", acctest.RandString(5))
		vlan       = acctest.RandIntRange(1, 3999)
	)

	rc := common.InitResourceCheck(
		rName,
		&vi,
		getVirtualInterfaceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualInterface_basic(name, vlan),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "direct_connect_id", dcId),
					resource.TestCheckResourceAttrPair(rName, "vgw_id", "opentelekomcloud_dc_virtual_gateway_v3.gw", "id"),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "description", "Created by acc test"),
					resource.TestCheckResourceAttr(rName, "type", "private"),
					resource.TestCheckResourceAttr(rName, "route_mode", "static"),
					resource.TestCheckResourceAttr(rName, "vlan", fmt.Sprintf("%v", vlan)),
					resource.TestCheckResourceAttr(rName, "bandwidth", "5"),
					resource.TestCheckResourceAttr(rName, "enable_bfd", "false"),
					resource.TestCheckResourceAttr(rName, "enable_nqa", "false"),
					resource.TestCheckResourceAttr(rName, "remote_ep_group.0", "1.1.1.0/30"),
					resource.TestCheckResourceAttr(rName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(rName, "local_gateway_v4_ip", "1.1.1.1/30"),
					resource.TestCheckResourceAttr(rName, "remote_gateway_v4_ip", "1.1.1.2/30"),
					resource.TestCheckResourceAttrSet(rName, "device_id"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttrSet(rName, "status"),
					resource.TestCheckResourceAttr(rName, "vif_peers.#", "1"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.address_family"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.bgp_asn"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.bgp_route_limit"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.device_id"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.enable_bfd"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.enable_nqa"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.id"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.local_gateway_ip"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.name"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.receive_route_num"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.remote_ep_group.#"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.remote_gateway_ip"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.route_mode"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.status"),
					resource.TestCheckResourceAttrSet(rName, "vif_peers.0.vif_id"),
				),
			},
			{
				Config: testAccVirtualInterface_update(updateName, vlan),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "direct_connect_id", dcId),
					resource.TestCheckResourceAttrPair(rName, "vgw_id", "opentelekomcloud_dc_virtual_gateway_v3.gw", "id"),
					resource.TestCheckResourceAttr(rName, "name", updateName),
					resource.TestCheckResourceAttr(rName, "description", ""),
					resource.TestCheckResourceAttr(rName, "type", "private"),
					resource.TestCheckResourceAttr(rName, "route_mode", "static"),
					resource.TestCheckResourceAttr(rName, "vlan", fmt.Sprintf("%v", vlan)),
					resource.TestCheckResourceAttr(rName, "bandwidth", "10"),
					resource.TestCheckResourceAttr(rName, "enable_bfd", "false"),
					resource.TestCheckResourceAttr(rName, "enable_nqa", "false"),
					resource.TestCheckResourceAttr(rName, "remote_ep_group.0", "1.1.1.0/30"),
					resource.TestCheckResourceAttr(rName, "remote_ep_group.1", "1.1.2.0/30"),
					resource.TestCheckResourceAttr(rName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(rName, "local_gateway_v4_ip", "1.1.1.1/30"),
					resource.TestCheckResourceAttr(rName, "remote_gateway_v4_ip", "1.1.1.2/30"),
					resource.TestCheckResourceAttrSet(rName, "device_id"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttrSet(rName, "status"),
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

func testAccVirtualInterface_base(name string) string {
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

func testAccVirtualInterface_basic(name string, vlan int) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_dc_virtual_interface_v3" "vi" {
  direct_connect_id = "%[2]s"
  vgw_id            = opentelekomcloud_dc_virtual_gateway_v3.gw.id
  name              = "%[3]s"
  description       = "Created by acc test"
  type              = "private"
  route_mode        = "static"
  vlan              = %[4]d
  bandwidth         = 5

  remote_ep_group = [
    "1.1.1.0/30",
  ]

  address_family       = "ipv4"
  local_gateway_v4_ip  = "1.1.1.1/30"
  remote_gateway_v4_ip = "1.1.1.2/30"
}
`, testAccVirtualInterface_base(name), os.Getenv("OS_DIRECT_CONNECT_ID"), name, vlan)
}

func testAccVirtualInterface_update(name string, vlan int) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_dc_virtual_interface_v3" "vi" {
  direct_connect_id = "%[2]s"
  vgw_id            = opentelekomcloud_dc_virtual_gateway_v3.gw.id
  name              = "%[3]s"
  type              = "private"
  route_mode        = "static"
  vlan              = %[4]d
  bandwidth         = 10

  remote_ep_group = [
    "1.1.1.0/30",
    "1.1.2.0/30",
  ]

  address_family       = "ipv4"
  local_gateway_v4_ip  = "1.1.1.1/30"
  remote_gateway_v4_ip = "1.1.1.2/30"
}
`, testAccVirtualInterface_base(name), os.Getenv("OS_DIRECT_CONNECT_ID"), name, vlan)
}
