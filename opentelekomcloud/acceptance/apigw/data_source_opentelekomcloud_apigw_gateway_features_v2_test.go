package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
)

func TestAccDataSourceInstanceFeatures_basic(t *testing.T) {
	var (
		rName = "data.opentelekomcloud_apigw_gateway_features_v2.test"
		dc    = common.InitDataSourceCheck(rName)

		byName   = "data.opentelekomcloud_apigw_gateway_features_v2.filter_by_name"
		dcByName = common.InitDataSourceCheck(byName)

		byNotFoundName   = "data.opentelekomcloud_apigw_gateway_features_v2.filter_by_not_found_name"
		dcByNotFoundName = common.InitDataSourceCheck(byNotFoundName)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckApigw(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceInstanceFeatures_basic(),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestMatchResourceAttr(rName, "features.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
					dcByName.CheckResourceExists(),
					resource.TestMatchResourceAttr(byName, "features.0.updated_at",
						regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}?(Z|([+-]\d{2}:\d{2}))$`)),
					resource.TestCheckOutput("is_name_filter_useful", "true"),
					dcByNotFoundName.CheckResourceExists(),
					resource.TestCheckOutput("is_name_not_found_filter_useful", "true"),
				),
			},
		},
	})
}

func testAccDataSourceInstanceFeatures_basic() string {
	return fmt.Sprintf(`
locals {
  gateway_id = "%[1]s"
}

data "opentelekomcloud_apigw_gateway_features_v2" "test" {
  gateway_id = local.gateway_id
}

# Filter by name
locals {
  feature_name = data.opentelekomcloud_apigw_gateway_features_v2.test.features[0].name
}

data "opentelekomcloud_apigw_gateway_features_v2" "filter_by_name" {
  gateway_id = local.gateway_id
  name       = local.feature_name
}

locals {
  name_filter_result = [
    for v in data.opentelekomcloud_apigw_gateway_features_v2.filter_by_name.features[*].name : v == local.feature_name
  ]
}

output "is_name_filter_useful" {
  value = length(local.name_filter_result) > 0 && alltrue(local.name_filter_result)
}

# Filter by name (not found)
locals {
  not_found_name = "not_found"
}

data "opentelekomcloud_apigw_gateway_features_v2" "filter_by_not_found_name" {
  gateway_id = local.gateway_id
  name       = local.not_found_name
}

locals {
  not_found_name_filter_result = [
    for v in data.opentelekomcloud_apigw_gateway_features_v2.filter_by_not_found_name.features[*].name : strcontains(v, local.not_found_name)
  ]
}

output "is_name_not_found_filter_useful" {
  value = length(local.not_found_name_filter_result) == 0
}
`, env.OS_APIGW_GATEWAY_ID)
}
