package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const dataEnginesName = "data.opentelekomcloud_ddm_engines_v1.engine_list"

func TestAccDDMEnginesV1DataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDDMEnginesV1DataSourceBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataEnginesName, "engines.0.id"),
				),
			},
		},
	})
}

var testAccDDMEnginesV1DataSourceBasic = `
data "opentelekomcloud_ddm_engines_v1" "engine_list" {
}
`
