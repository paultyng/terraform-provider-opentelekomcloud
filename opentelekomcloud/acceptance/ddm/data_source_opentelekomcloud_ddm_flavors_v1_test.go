package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const dataFlavorsName = "data.opentelekomcloud_ddm_flavors_v1.flavor_list"

func TestAccDDMFlavorsV1DataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDDMFlavorsV1DataSourceBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataFlavorsName, "flavor_groups.0.type"),
				),
			},
		},
	})
}

var testAccDDMFlavorsV1DataSourceBasic = `
data "opentelekomcloud_ddm_engines_v1" "engine_list" {
}

data "opentelekomcloud_ddm_flavors_v1" "flavor_list" {
  engine_id = data.opentelekomcloud_ddm_engines_v1.engine_list.engines.0.id
}
`
