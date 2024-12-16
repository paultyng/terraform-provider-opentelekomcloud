package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/topics"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceTopicV2Name = "opentelekomcloud_dms_topic_v2.topic_1"

func geDmsTopicFunc(config *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := config.DmsV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating DMS v2 client: %s", err)
	}
	var fTopic topics.Topic

	v, err := topics.List(client, state.Primary.Attributes["instance_id"])
	if err != nil {
		return nil, fmt.Errorf("provided topic doesn't exist")
	}
	found := false

	for _, topic := range v.Topics {
		if topic.Name == state.Primary.ID {
			fTopic = topic
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("provided topic doesn't exist")
	}

	return fTopic, nil
}

func TestAccDmsTopicsV2_basic(t *testing.T) {
	var instance topics.Topic
	var instanceName = fmt.Sprintf("dms_instance_%s", acctest.RandString(5))
	var topicName = fmt.Sprintf("topic_instance_%s", acctest.RandString(5))

	rc := common.InitResourceCheck(
		resourceTopicV2Name,
		&instance,
		geDmsTopicFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDmsV1InstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsV2TopicBasic(instanceName, topicName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "partition", "200"),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "replication", "3"),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "sync_replication", "true"),
					resource.TestCheckResourceAttr(resourceTopicV2Name, "retention_time", "600"),
				),
			},
			{
				ResourceName:      resourceTopicV2Name,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccDmsTopicImportStateFunc(resourceTopicV2Name),
			},
		},
	})
}

func testAccDmsTopicImportStateFunc(rName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return "", fmt.Errorf("resource (%s) not found: %s", rName, rs)
		}
		if rs.Primary.Attributes["instance_id"] == "" {
			return "", fmt.Errorf("invalid format specified for import ID, want '<instance_id>/<name>', but '%s/%s'",
				rs.Primary.Attributes["instance_id"], rs.Primary.ID)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["instance_id"], rs.Primary.ID), nil
	}
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
