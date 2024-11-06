package rms

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceAdvancedQuery_basic(t *testing.T) {
	dataSource := "data.opentelekomcloud_rms_advanced_query_v1.test"
	dc := common.InitDataSourceCheck(dataSource)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceAdvancedQuery_basic,
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestCheckOutput("is_name_set", "true"),
					resource.TestCheckOutput("is_id_set", "true"),
					resource.TestCheckOutput("is_query_info_correct", "true"),
				),
			},
		},
	})
}

const testDataSourceAdvancedQuery_basic = `
data "opentelekomcloud_rms_advanced_query_v1" "test" {
  expression = "select name, id from tracked_resources where provider = 'ecs' and type = 'cloudservers'"
}

locals {
  name_set = [
    for v in data.opentelekomcloud_rms_advanced_query_v1.test.results[*].name : v != ""
  ]
  id_set = [
    for v in data.opentelekomcloud_rms_advanced_query_v1.test.results[*].id : v != ""
  ]
  query_info_correct = [
    for v in data.opentelekomcloud_rms_advanced_query_v1.test.query_info[*].select_fields :
    length(setsubtract(v, ["name", "id"])) == 0
  ]
}

output "is_name_set" {
  value = alltrue(local.name_set) && length(local.name_set) > 0
}

output "is_id_set" {
  value = alltrue(local.id_set) && length(local.id_set) > 0
}

output "is_query_info_correct" {
  value = alltrue(local.query_info_correct) && length(local.query_info_correct) > 0
}
`
