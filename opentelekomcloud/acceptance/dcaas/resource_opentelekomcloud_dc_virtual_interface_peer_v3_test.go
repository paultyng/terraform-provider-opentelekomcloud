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

func getVirtualInterfacePeerFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.DCaaSV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DCaaS v3 client: %s", err)
	}
	vi, err := virtual_interface.Get(c, state.Primary.Attributes["vif_id"])
	if err != nil {
		return nil, fmt.Errorf("error getting OpenTelekomCloud DCaaS v3 virtual interface: %s", err)
	}
	for _, v := range vi.VifPeers {
		if v.ID == state.Primary.ID {
			return v, err
		}
	}
	return nil, fmt.Errorf("error OpenTelekomCloud DCaaS v3 virtual interface peer not found: %s", err)
}

func TestAccVirtualInterfacePeer_basic(t *testing.T) {
	dcId := os.Getenv("OS_DIRECT_CONNECT_ID")
	if dcId == "" {
		t.Skip("OS_DIRECT_CONNECT_ID should be set for acceptance tests")
	}
	var (
		vp virtual_interface.VifPeer

		rName      = "opentelekomcloud_dc_virtual_interface_peer_v3.vp"
		name       = fmt.Sprintf("dc_acc_vp%s", acctest.RandString(5))
		updateName = fmt.Sprintf("dc_acc_vp_up%s", acctest.RandString(5))
		vlan       = acctest.RandIntRange(1, 3999)
	)

	rc := common.InitResourceCheck(
		rName,
		&vp,
		getVirtualInterfacePeerFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualInterfaceVif_basic(name, vlan),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(rName, "description", "ipv6 peer"),
					resource.TestCheckResourceAttrSet(rName, "remote_ep_group.#"),
					resource.TestCheckResourceAttr(rName, "remote_ep_group.0", "fd00:0:0:0:0:0:0:0/64"),
					resource.TestCheckResourceAttrSet(rName, "remote_gateway_ip"),
					resource.TestCheckResourceAttrSet(rName, "route_mode"),
					resource.TestCheckResourceAttrSet(rName, "status"),
					resource.TestCheckResourceAttrSet(rName, "vif_id"),
				),
			},
			{
				Config: testAccVirtualInterfaceVif_update(updateName, vlan),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(rName, "description", "ipv6 peer updated"),
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

func testAccVirtualInterfaceVif_base(name string, vlan int) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_dc_virtual_gateway_v3" "gw" {
  vpc_id      = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  name        = "%[2]s"
  description = "Created by acc test"

  local_ep_group = [
    "192.168.0.0/16",
  ]
  local_ep_group_ipv6 = [
    "FD00::/64",
  ]
}

resource "opentelekomcloud_dc_virtual_interface_v3" "vi" {
  direct_connect_id = "%[3]s"
  vgw_id            = opentelekomcloud_dc_virtual_gateway_v3.gw.id
  name              = "%[2]s"
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

`, common.DataSourceSubnet, name, os.Getenv("OS_DIRECT_CONNECT_ID"), vlan)
}

func testAccVirtualInterfaceVif_basic(name string, vlan int) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_dc_virtual_interface_peer_v3" "vp" {
  vif_id            = opentelekomcloud_dc_virtual_interface_v3.vi.id
  name              = "%[2]s"
  address_family    = "ipv6"
  route_mode        = "static"
  remote_ep_group   = ["fd00:0:0:0:0:0:0:0/64"]
  description       = "ipv6 peer"
  local_gateway_ip  = "FD00::1/64"
  remote_gateway_ip = "FD00::2/64"
}
`, testAccVirtualInterfaceVif_base(name, vlan), name)
}

func testAccVirtualInterfaceVif_update(name string, vlan int) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_dc_virtual_interface_peer_v3" "vp" {
  vif_id            = opentelekomcloud_dc_virtual_interface_v3.vi.id
  name              = "%[2]s"
  address_family    = "ipv6"
  route_mode        = "static"
  remote_ep_group   = ["fd00:0:0:0:0:0:0:0/64"]
  description       = "ipv6 peer updated"
  local_gateway_ip  = "FD00::1/64"
  remote_gateway_ip = "FD00::2/64"
}
`, testAccVirtualInterfaceVif_base(name, vlan), name)
}
