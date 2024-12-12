package vpcep

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/tools"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vpcep/v1/services"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceVPCEPServiceName = "opentelekomcloud_vpcep_service_v1.service"

func getVPCServiceFunc(config *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := config.VpcEpV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating VPCEP v1 client: %s", err)
	}
	return services.Get(client, state.Primary.ID)
}

func TestVPCEPService_basic(t *testing.T) {
	var svc services.Service
	srvName := tools.RandomString("tf-test-", 4)
	srvName2 := tools.RandomString("tf-test-", 4)

	rc := common.InitResourceCheck(
		resourceVPCEPServiceName,
		&svc,
		getVPCServiceFunc,
	)

	t.Parallel()
	quotas.BookOne(t, serviceQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testServiceBasic(srvName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "name", srvName),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "port.#", "1"),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "server_type", "LB"),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "service_type", "interface"),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "tags.key", "value"),
				),
			},
			{
				Config: testServiceUpdated(srvName2),
				Check: resource.ComposeTestCheckFunc(
					checkServiceIDPersist(resourceVPCEPServiceName, &svc),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "name", srvName2),
					resource.TestCheckResourceAttr(resourceVPCEPServiceName, "port.#", "2"),
				),
			},
			{
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      resourceVPCEPServiceName,
			},
		},
	})
}

func checkServiceIDPersist(name string, svc *services.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID != svc.ID {
			return fmt.Errorf("service ID changed")
		}
		return nil
	}
}

func testServiceBasic(name string) string {
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
  description = "test description update"

  approval_enabled = false

  port {
    client_port = 80
    server_port = 8080
  }

  tags = {
    "key" : "value",
  }
  whitelist = ["698f9bf85ca9437a9b2f41132ab3aa0e"]
}
`, common.DataSourceSubnet, name)
}

func testServiceUpdated(name string) string {
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
  description = "test description updated"

  approval_enabled = true

  port {
    client_port = 80
    server_port = 8080
  }

  port {
    client_port = 81
    server_port = 8081
    protocol    = "TCP"
  }

  tags = {
    "key" : "value",
  }
  whitelist = ["698f9bf85ca9437a9b2f41132ab3aa0e", "e8df38eb4e4f4f148e06d8db527059c7"]
}
`, common.DataSourceSubnet, name)
}
