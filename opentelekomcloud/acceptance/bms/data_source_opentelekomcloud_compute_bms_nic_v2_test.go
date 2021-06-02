package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
)

func TestAccOTCBMSNicV2DataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckBMSNic(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenTelekomCloudBMSNicV2DataSource_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBMSNicV2DataSourceID("data.opentelekomcloud_compute_bms_nic_v2.nic_1"),
					resource.TestCheckResourceAttr("data.opentelekomcloud_compute_bms_nic_v2.nic_1", "status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccCheckBMSNicV2DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find nic data source: %s ", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("nic data source ID not set ")
		}

		return nil
	}
}

var testAccOpenTelekomCloudBMSNicV2DataSource_basic = fmt.Sprintf(`
resource "opentelekomcloud_compute_instance_v2" "instance_1" {
  name              = "BMSinstance_1"
  image_id          = "%s"
  security_groups   = ["default"]
  availability_zone = "%s"
  flavor_id         = "physical.o2.medium"
  flavor_name       = "physical.o2.medium"
  metadata          = {
    foo = "bar"
  }
  network {
    uuid = "%s"
  }
}
data "opentelekomcloud_compute_bms_nic_v2" "nic_1" {
  server_id = opentelekomcloud_compute_instance_v2.instance_1.id
}
`, env.OS_IMAGE_ID, env.OS_AVAILABILITY_ZONE, env.OS_NETWORK_ID)

func testAccPreCheckBMSNic(t *testing.T) {
	common.TestAccPreCheckRequiredEnvVars(t)

	if env.OS_NIC_ID == "" {
		t.Skip("OS_NIC_ID must be set for NIC acceptance tests")
	}
}
