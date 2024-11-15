package hss

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceHostGroups_basic(t *testing.T) {
	dataSource := "data.opentelekomcloud_hss_host_groups_v5.test"
	dc := common.InitDataSourceCheck(dataSource)
	name := fmt.Sprintf("hss-acc-api%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceHostGroups_basic(name),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestCheckResourceAttrSet(dataSource, "groups.#"),
					resource.TestCheckResourceAttrSet(dataSource, "groups.0.id"),
					resource.TestCheckResourceAttrSet(dataSource, "groups.0.name"),
					resource.TestCheckResourceAttrSet(dataSource, "groups.0.host_num"),
					resource.TestCheckResourceAttrSet(dataSource, "groups.0.host_ids.#"),

					resource.TestCheckOutput("is_group_id_filter_useful", "true"),
					resource.TestCheckOutput("is_host_num_filter_useful", "true"),
					resource.TestCheckOutput("not_found_validation_pass", "true"),
				),
			},
		},
	})
}

func testDataSourceHostGroups_basic(name string) string {
	hostGroupBasic := testAccHostGroup_basic(name)

	return fmt.Sprintf(`
%[1]s

data "opentelekomcloud_hss_host_groups_v5" "test" {
  depends_on = [opentelekomcloud_hss_host_group_v5.group]
}

# Filter using group ID.
locals {
  group_id = data.opentelekomcloud_hss_host_groups_v5.test.groups[0].id
}

data "opentelekomcloud_hss_host_groups_v5" "group_id_filter" {
  group_id = local.group_id
}

output "is_group_id_filter_useful" {
  value = length(data.opentelekomcloud_hss_host_groups_v5.group_id_filter.groups) > 0 && alltrue(
    [for v in data.opentelekomcloud_hss_host_groups_v5.group_id_filter.groups[*].id : v == local.group_id]
  )
}

# Filter using host_num.
locals {
  host_num = data.opentelekomcloud_hss_host_groups_v5.test.groups[0].host_num
}

data "opentelekomcloud_hss_host_groups_v5" "host_num_filter" {
  host_num = local.host_num
}

output "is_host_num_filter_useful" {
  value = length(data.opentelekomcloud_hss_host_groups_v5.host_num_filter.groups) > 0 && alltrue(
    [for v in data.opentelekomcloud_hss_host_groups_v5.host_num_filter.groups[*].host_num : v == local.host_num]
  )
}

# Filter using non existent name.
data "opentelekomcloud_hss_host_groups_v5" "not_found" {
  name = "resource_not_found"
}

output "not_found_validation_pass" {
  value = length(data.opentelekomcloud_hss_host_groups_v5.not_found.groups) == 0
}
`, hostGroupBasic)
}
