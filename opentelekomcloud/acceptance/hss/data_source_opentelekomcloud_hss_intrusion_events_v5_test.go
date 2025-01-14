package hss

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceEvents_basic(t *testing.T) {
	dataSource := "data.opentelekomcloud_hss_intrusion_events_v5.events"
	dc := common.InitDataSourceCheck(dataSource)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceIntrusionEvents_basic(),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestCheckResourceAttrSet(dataSource, "events.#"),
				),
			},
		},
	})
}

func testDataSourceIntrusionEvents_basic() string {
	return `

data "opentelekomcloud_hss_intrusion_events_v5" "events" {
  category = "host"
}
`
}
