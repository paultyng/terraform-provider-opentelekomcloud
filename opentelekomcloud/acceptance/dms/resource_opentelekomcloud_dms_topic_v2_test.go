package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const resourceTopicV2Name = "opentelekomcloud_dms_topic_v2.topic_1"

func TestAccDmsTopicsV2_basic(t *testing.T) {
	var instanceName = fmt.Sprintf("dms_instance_%s", acctest.RandString(5))
	var topicName = fmt.Sprintf("topic_instance_%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDmsV1InstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsV2TopicBasic(instanceName, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceTopicV2Name, "partition", "200"),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "replication", "3"),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "sync_replication", "true"),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "retention_time", "600"),
				),
			},
		},
	})
}

func testAccDmsV2TopicBasic(instanceName string, topicName string) string {
	return fmt.Sprintf(`
%s

%s

data "opentelekomcloud_dms_az_v1" "az_1" {}

data "opentelekomcloud_dms_product_v1" "product_1" {
  engine        = "kafka"
  instance_type = "cluster"
  version       = "2.3.0"
}

resource "opentelekomcloud_dms_instance_v1" "instance_1" {
  name              = "%s"
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

resource "opentelekomcloud_dms_topic_v2" "topic_1" {
  instance_id      = resource.opentelekomcloud_dms_instance_v1.instance_1.id
  name             = "%s"
  partition        = 200
  replication      = 3
  sync_replication = true
  retention_time   = 600
}`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, instanceName, topicName)
}
