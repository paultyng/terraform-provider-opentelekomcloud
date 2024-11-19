package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const resourceReassignPartitionsV2Name = "opentelekomcloud_dms_reassign_partitions_v2.rp_1"

func TestAccDmsReassignPartitionsV2_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsV2ReassignPartitionsBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceReassignPartitionsV2Name, "instance_id"),
				),
			},
		},
	})
}

func testAccDmsV2ReassignPartitionsBasic() string {
	var instanceName = fmt.Sprintf("dms_test_rp_instance_%s", acctest.RandString(5))
	var topicName = fmt.Sprintf("dms_test_rp_topic_%s", acctest.RandString(5))

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

resource "opentelekomcloud_dms_topic_v1" "topic_1" {
  instance_id      = opentelekomcloud_dms_instance_v2.instance_1.id
  name             = "%s"
  partition        = 10
  replication      = 2
  sync_replication = true
  retention_time   = 720
}

resource "opentelekomcloud_dms_reassign_partitions_v2" "rp_1" {
  instance_id = opentelekomcloud_dms_instance_v2.instance_1.id
  time_estimate = false
  reassignments {
    topic = opentelekomcloud_dms_topic_v1.topic_1.name
  }
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, instanceName, topicName)
}
