package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/instances"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const ddmInstanceResourceName = "opentelekomcloud_ddm_instance_v1.instance_1"

func getDDMInstanceResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.DdmV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating SDRS Client: %s", err)
	}
	return instances.QueryInstanceDetails(client, state.Primary.ID)
}

func TestAccDdmInstancesV1_basic(t *testing.T) {
	var instance instances.QueryInstanceDetailsResponse
	rc := common.InitResourceCheck(
		ddmInstanceResourceName,
		&instance,
		getDDMInstanceResourceFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDdmInstanceV1Basic,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "name", "ddm-instance"),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "node_num", "2"),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "username", "test_user"),
				),
			},
			{
				Config: testAccDdmInstanceV1ScaleUp,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "name", "ddm-instance-scale-up"),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "node_num", "3"),
				),
			},
			{
				Config: testAccDdmInstanceV1ScaleDown,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "name", "ddm-instance-scale-down"),
					resource.TestCheckResourceAttr(ddmInstanceResourceName, "node_num", "1"),
				),
			},
			{
				ResourceName:      ddmInstanceResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"availability_zones",
					"flavor_id",
					"engine_id",
					"time_zone",
					"password",
					"param_group_id",
					"purge_rds_on_delete",
				},
			},
		},
	})
}

var testAccDdmInstanceV1Basic = fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name              = "ddm-instance"
  availability_zones = ["%s"]
  flavor_id         = "941b5a6d-3485-329e-902c-ffd49d352f16"
  node_num          = 2
  engine_id         = "367b68a3-b48b-3d8a-b3a1-4c463a75a4b4"
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  time_zone = "UTC+01:00"
  username = "test_user"
  password = "test!-acc-Password-V1!"
}
`, common.DataSourceSubnet, common.DataSourceSecGroupDefault, env.OS_AVAILABILITY_ZONE)

var testAccDdmInstanceV1ScaleUp = fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name              = "ddm-instance-scale-up"
  availability_zones = ["%s"]
  flavor_id         = "941b5a6d-3485-329e-902c-ffd49d352f16"
  node_num          = 3
  engine_id         = "367b68a3-b48b-3d8a-b3a1-4c463a75a4b4"
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  time_zone = "UTC+01:00"
  username = "test_user"
  password = "test!-acc-Password-V1!"
}
`, common.DataSourceSubnet, common.DataSourceSecGroupDefault, env.OS_AVAILABILITY_ZONE)

var testAccDdmInstanceV1ScaleDown = fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name              = "ddm-instance-scale-down"
  availability_zones = ["%s"]
  flavor_id         = "941b5a6d-3485-329e-902c-ffd49d352f16"
  node_num          = 1
  engine_id         = "367b68a3-b48b-3d8a-b3a1-4c463a75a4b4"
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  time_zone = "UTC+01:00"
  username = "test_user"
  password = "test!-acc-Password-V2!"
  purge_rds_on_delete = true
}
`, common.DataSourceSubnet, common.DataSourceSecGroupDefault, env.OS_AVAILABILITY_ZONE)
