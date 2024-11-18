package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/instances/management"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceGroupV2Name = "opentelekomcloud_dms_consumer_group_v2.group_1"

func getDmsConsumerGroupResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.DmsV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating DMS V2 Client: %s", err)
	}
	getResp, err := management.GetConsumerGroup(client, state.Primary.Attributes["instance_id"], state.Primary.Attributes["group_name"])
	if err != nil {
		return nil, fmt.Errorf("error fetching dms group: %s", err)
	}
	return getResp.Group, nil
}

func TestAccDmsConsumerGroupV2_basic(t *testing.T) {
	var dmsConsumerGroup management.Group
	rc := common.InitResourceCheck(
		resourceGroupV2Name,
		&dmsConsumerGroup,
		getDmsConsumerGroupResourceFunc,
	)

	var groupName = fmt.Sprintf("dms_consumer_group_%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDmsV2ConsumerGroupBasic(groupName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceGroupV2Name, "group_name", groupName),
				),
			},
			{
				ResourceName:      resourceGroupV2Name,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description",
				},
			},
		},
	})
}

func testAccDmsV2ConsumerGroupBasic(groupName string) string {
	return fmt.Sprintf(`
%s

%s

data "opentelekomcloud_dms_az_v1" "az_1" {}

data "opentelekomcloud_dms_product_v1" "product_1" {
  engine        = "kafka"
  instance_type = "cluster"
  version       = "2.3.0"
}

resource "opentelekomcloud_dms_instance_v2" "instance_1" {
  name              = "dms_test_cg_instance"
  engine            = "kafka"
  storage_space     = data.opentelekomcloud_dms_product_v1.product_1.storage
  access_user       = "user"
  password          = "Dmstest@123"
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  subnet_id         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  available_zones   = [data.opentelekomcloud_dms_az_v1.az_1.id]
  product_id        = data.opentelekomcloud_dms_product_v1.product_1.id
  engine_version    = data.opentelekomcloud_dms_product_v1.product_1.version
  storage_spec_code = data.opentelekomcloud_dms_product_v1.product_1.storage_spec_code
}

resource "opentelekomcloud_dms_consumer_group_v2" "group_1" {
  instance_id = opentelekomcloud_dms_instance_v2.instance_1.id
  group_name  = "%s"
  description = "Test consumer group"
}


`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, groupName)
}
