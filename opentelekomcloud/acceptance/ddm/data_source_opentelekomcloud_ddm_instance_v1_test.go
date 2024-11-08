package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
)

const dataInstanceName = "data.opentelekomcloud_ddm_instance_v1.instance"

func TestAccDDMInstanceV1DataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDDMInstanceV1DataSourceBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDDMInstanceV1DataSourceID(dataInstanceName),
					resource.TestCheckResourceAttr(dataInstanceName, "name", "ddm-instance"),
					resource.TestCheckResourceAttrSet(dataInstanceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccCheckDDMInstanceV1DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find instances data source: %s ", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DDM instance data source ID not set ")
		}

		return nil
	}
}

var testAccDDMInstanceV1DataSourceBasic = fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name               = "ddm-instance"
  availability_zones = ["%s"]
  flavor_id          = "941b5a6d-3485-329e-902c-ffd49d352f16"
  node_num           = 2
  engine_id          = "367b68a3-b48b-3d8a-b3a1-4c463a75a4b4"
  vpc_id             = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id          = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  security_group_id  = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  time_zone          = "UTC+01:00"
  username           = "test_user"
  password           = "test!-acc-Password-V1!"
}

data "opentelekomcloud_ddm_instance_v1" "instance" {
  instance_id = opentelekomcloud_ddm_instance_v1.instance_1.id
}
`, common.DataSourceSubnet, common.DataSourceSecGroupDefault, env.OS_AVAILABILITY_ZONE)
