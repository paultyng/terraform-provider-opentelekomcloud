package rms

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceRmsAdvancedQueries_basic(t *testing.T) {
	dataSource1 := "data.opentelekomcloud_rms_advanced_queries_v1.basic"
	dataSource2 := "data.opentelekomcloud_rms_advanced_queries_v1.filter_by_name"
	rName := acctest.RandomWithPrefix("rms-test")
	dc1 := common.InitDataSourceCheck(dataSource1)
	dc2 := common.InitDataSourceCheck(dataSource2)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceDataSourceRmsAdvancedQueries_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					dc1.CheckResourceExists(),
					dc2.CheckResourceExists(),
					resource.TestCheckOutput("is_results_not_empty", "true"),
					resource.TestCheckOutput("is_name_filter_useful", "true"),
				),
			},
		},
	})
}

func testDataSourceDataSourceRmsAdvancedQueries_basic(name string) string {
	return fmt.Sprintf(`
%[1]s

data "opentelekomcloud_rms_advanced_queries_v1" "basic" {
  depends_on = [opentelekomcloud_rms_advanced_query_v1.test]
}

data "opentelekomcloud_rms_advanced_queries_v1" "filter_by_name" {
  name = "%[2]s"

  depends_on = [opentelekomcloud_rms_advanced_query_v1.test]
}

locals {
  name_filter_result = [for v in data.opentelekomcloud_rms_advanced_queries_v1.filter_by_name.queries[*].name : v == "%[2]s"]
}

output "is_results_not_empty" {
  value = length(data.opentelekomcloud_rms_advanced_queries_v1.basic.queries) > 0
}

output "is_name_filter_useful" {
  value = alltrue(local.name_filter_result) && length(local.name_filter_result) > 0
}
`, testAdvancedQuery_basic(name), name)
}
