package vpcep

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/tools"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vpcep/v1/endpoints"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceEndpointName = "opentelekomcloud_vpcep_endpoint_v1.endpoint"

func getVPCEndpointFunc(config *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := config.VpcEpV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating VPCEP v1 client: %s", err)
	}
	return endpoints.Get(client, state.Primary.ID)
}

func TestVPCEndpoint_basic(t *testing.T) {
	var ep endpoints.Endpoint
	name := tools.RandomString("tf-test-ep-", 4)

	rc := common.InitResourceCheck(
		resourceEndpointName,
		&ep,
		getVPCEndpointFunc,
	)

	t.Parallel()
	quotas.BookMany(t, endpointQuotas())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testEndpointBasic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceEndpointName, "tags.fizz", "buzz"),
					resource.TestCheckResourceAttr(resourceEndpointName, "enable_dns", "true"),
					resource.TestCheckResourceAttr(resourceEndpointName, "dns_names.#", "1"),
					resource.TestCheckResourceAttr(resourceEndpointName, "service_name", name),
				),
			},
			{
				Config: testEndpointBasic_Update(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceEndpointName, "status", "accepted"),
					resource.TestCheckResourceAttr(resourceEndpointName, "tags.owner", "tf-acc-update"),
					resource.TestCheckResourceAttr(resourceEndpointName, "tags.foo", "bar"),
				),
			},
			{
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      resourceEndpointName,
			},
		},
	})
}

func testEndpointBasic(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_vpcep_endpoint_v1" "endpoint" {
  service_id = opentelekomcloud_vpcep_service_v1.service.id
  vpc_id     = opentelekomcloud_vpcep_service_v1.service.vpc_id
  subnet_id  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.id
  port_ip    = cidrhost(data.opentelekomcloud_vpc_subnet_v1.shared_subnet.cidr, 32)
  enable_dns = true

  tags = {
    "fizz" : "buzz"
  }
}
`, testServiceBasic(name))
}

func testEndpointBasic_Update(rName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_vpcep_endpoint_v1" "endpoint" {
  service_id = opentelekomcloud_vpcep_service_v1.service.id
  vpc_id     = opentelekomcloud_vpcep_service_v1.service.vpc_id
  subnet_id  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.id
  port_ip    = cidrhost(data.opentelekomcloud_vpc_subnet_v1.shared_subnet.cidr, 32)
  enable_dns = true

  tags = {
    owner = "tf-acc-update"
    foo   = "bar"
  }
}
`, testServiceBasic(rName))
}
