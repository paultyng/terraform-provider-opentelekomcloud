package hss

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceHosts_basic(t *testing.T) {
	dataSource := "data.opentelekomcloud_hss_hosts_v5.hosts"
	dc := common.InitDataSourceCheck(dataSource)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceHosts_basic(),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.#"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.id"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.name"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.status"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.os_type"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.agent_status"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.protect_status"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.detect_result"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.asset_value"),
					resource.TestCheckResourceAttrSet(dataSource, "hosts.0.private_ip"),

					resource.TestCheckOutput("is_host_id_filter_useful", "true"),
					resource.TestCheckOutput("is_agent_status_filter_useful", "true"),
					resource.TestCheckOutput("not_found_validation_pass", "true"),
				),
			},
		},
	})
}

func testDataSourceHosts_basic() string {
	return `

data "opentelekomcloud_hss_hosts_v5" "hosts" {}

# Filter using host ID.
locals {
  host_id = data.opentelekomcloud_hss_hosts_v5.hosts.hosts[0].id
}

data "opentelekomcloud_hss_hosts_v5" "host_id_filter" {
  host_id = local.host_id
}

output "is_host_id_filter_useful" {
  value = length(data.opentelekomcloud_hss_hosts_v5.host_id_filter.hosts) > 0 && alltrue(
    [for v in data.opentelekomcloud_hss_hosts_v5.host_id_filter.hosts[*].id : v == local.host_id]
  )
}

# Filter using agent_status
locals {
  agent_status = data.opentelekomcloud_hss_hosts_v5.hosts.hosts[0].agent_status
}

data "opentelekomcloud_hss_hosts_v5" "agent_status_filter" {
  agent_status = local.agent_status
}

output "is_agent_status_filter_useful" {
  value = length(data.opentelekomcloud_hss_hosts_v5.agent_status_filter.hosts) > 0 && alltrue(
    [for v in data.opentelekomcloud_hss_hosts_v5.agent_status_filter.hosts[*].agent_status : v == local.agent_status]
  )
}

# Filter using non existent name.
data "opentelekomcloud_hss_hosts_v5" "not_found" {
  name = "resource_not_found"
}

output "not_found_validation_pass" {
  value = length(data.opentelekomcloud_hss_hosts_v5.not_found.hosts) == 0
}
`
}
