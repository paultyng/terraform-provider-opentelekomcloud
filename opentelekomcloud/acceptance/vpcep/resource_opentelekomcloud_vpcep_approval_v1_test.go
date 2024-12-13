package vpcep

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/tools"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vpcep/v1/endpoints"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccVPCEndpointApproval_Basic(t *testing.T) {
	var endpoint endpoints.Endpoint
	rName := tools.RandomString("tf-test-ep-", 4)
	resourceName := "opentelekomcloud_vpcep_approval_v1.approval"

	rc := common.InitResourceCheck(
		"opentelekomcloud_vpcep_endpoint_v1.endpoint",
		&endpoint,
		getVPCEndpointFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointApproval_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(resourceName, "id", "opentelekomcloud_vpcep_service_v1.service", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "connections.0.endpoint_id",
						"opentelekomcloud_vpcep_endpoint_v1.endpoint", "id"),
					resource.TestCheckResourceAttr(resourceName, "connections.0.status", "accepted"),
				),
			},
			{
				Config: testAccVPCEndpointApproval_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "connections.0.endpoint_id",
						"opentelekomcloud_vpcep_endpoint_v1.endpoint", "id"),
					resource.TestCheckResourceAttr(resourceName, "connections.0.status", "rejected"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVPCEndpointApproval_Base(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_lb_loadbalancer_v2" "lb_1" {
  vip_subnet_id = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.subnet_id
}

resource "opentelekomcloud_vpcep_service_v1" "service" {
  name        = "%s"
  port_id     = opentelekomcloud_lb_loadbalancer_v2.lb_1.vip_port_id
  vpc_id      = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  server_type = "LB"
  description = "test description"

  approval_enabled = true

  port {
    client_port = 80
    server_port = 8080
  }

  tags = {
    "key" : "value",
  }
  whitelist = ["698f9bf85ca9437a9b2f41132ab3aa0e"]
}

resource "opentelekomcloud_vpcep_endpoint_v1" "endpoint" {
  service_id = opentelekomcloud_vpcep_service_v1.service.id
  vpc_id     = opentelekomcloud_vpcep_service_v1.service.vpc_id
  subnet_id  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.id
  enable_dns = true

  tags = {
    "fizz" : "buzz"
  }

  lifecycle {
    ignore_changes = [enable_dns]
  }
}
`, common.DataSourceSubnet, name)
}

func testAccVPCEndpointApproval_Basic(rName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_vpcep_approval_v1" "approval" {
  service_id = opentelekomcloud_vpcep_service_v1.service.id
  endpoints  = [opentelekomcloud_vpcep_endpoint_v1.endpoint.id]
}
`, testAccVPCEndpointApproval_Base(rName))
}

func testAccVPCEndpointApproval_Update(rName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_vpcep_approval_v1" "approval" {
  service_id = opentelekomcloud_vpcep_service_v1.service.id
  endpoints  = []
}
`, testAccVPCEndpointApproval_Base(rName))
}
