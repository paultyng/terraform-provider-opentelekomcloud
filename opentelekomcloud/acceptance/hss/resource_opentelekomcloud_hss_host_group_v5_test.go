package hss

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	group "github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/host"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/hss"
)

func getHostGroupFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.HssV5Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating HSS v5 client: %s", err)
	}
	return hss.QueryHostGroupById(client, state.Primary.ID)
}

func TestAccHostGroup_basic(t *testing.T) {
	var (
		gr *group.HostGroupResp

		name  = fmt.Sprintf("hss-acc-api%s", acctest.RandString(5))
		rName = "opentelekomcloud_hss_host_group_v5.group"
	)

	rc := common.InitResourceCheck(
		rName,
		&gr,
		getHostGroupFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccHostGroup_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "host_ids.#", "1"),
					resource.TestCheckResourceAttrSet(rName, "host_num"),
				),
			},
			{
				Config: testAccHostGroup_update(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name+"-update"),
					resource.TestCheckResourceAttr(rName, "host_ids.#", "2"),
					resource.TestCheckResourceAttrSet(rName, "host_num"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				// The field `unprotect_host_ids` will be filled in during the creation and editing operations.
				// We only need to add ignore to the test case and do not need to make special instructions in the document.
				ImportStateVerifyIgnore: []string{
					"unprotect_host_ids",
				},
			},
		},
	})
}

func testAccHostGroup_base(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_compute_instance_v2" "instance" {
  count = 2

  name              = "%s"
  description       = "my_desc"
  availability_zone = "%s"

  image_name = "Standard_Debian_11_latest"
  flavor_id  = "s3.large.2"

  metadata = {
    foo = "bar"
  }
  network {
    uuid = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  }

  tags = {
    muh = "value-create"
    kuh = "value-create"
    emp = ""
  }

  stop_before_destroy = true
}
`, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccHostGroup_basic(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_hss_host_group_v5" "group" {
  name     = "%[2]s"
  host_ids = slice(opentelekomcloud_compute_instance_v2.instance[*].id, 0, 1)
}
`, testAccHostGroup_base(name), name)
}

func testAccHostGroup_update(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_hss_host_group_v5" "group" {
  name     = "%[2]s-update"
  host_ids = opentelekomcloud_compute_instance_v2.instance[*].id
}
`, testAccHostGroup_base(name), name)
}
