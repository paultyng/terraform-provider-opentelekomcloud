package rms

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceRmsAdvancedQuerySchemas_basic(t *testing.T) {
	dataSource1 := "data.opentelekomcloud_rms_advanced_query_schemas_v1.basic"
	dataSource2 := "data.opentelekomcloud_rms_advanced_query_schemas_v1.filter_by_type"
	dc1 := common.InitDataSourceCheck(dataSource1)
	dc2 := common.InitDataSourceCheck(dataSource2)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceDataSourceRmsAdvancedQuerySchemas_basic,
				Check: resource.ComposeTestCheckFunc(
					dc1.CheckResourceExists(),
					dc2.CheckResourceExists(),
					resource.TestCheckOutput("is_results_not_empty", "true"),
					resource.TestCheckOutput("is_type_filter_useful", "true"),
				),
			},
		},
	})
}

const testDataSourceDataSourceRmsAdvancedQuerySchemas_basic = `
data "opentelekomcloud_rms_advanced_query_schemas_v1" "basic" {}

data "opentelekomcloud_rms_advanced_query_schemas_v1" "filter_by_type" {
  type = "ecs.cloudservers"
}

locals {
  type_filter_result = [for v in data.opentelekomcloud_rms_advanced_query_schemas_v1.filter_by_type.schemas[*].type : v == "ecs.cloudservers"]
}

output "is_results_not_empty" {
  value = length(data.opentelekomcloud_rms_advanced_query_schemas_v1.basic.schemas) > 0
}

output "is_type_filter_useful" {
  value = alltrue(local.type_filter_result) && length(local.type_filter_result) > 0
}
`
