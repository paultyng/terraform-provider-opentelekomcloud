package rms

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
)

func TestAccDataSourceRmsPolicyStates_basic(t *testing.T) {
	dataSource1 := "data.opentelekomcloud_rms_policy_states_v1.basic"
	dataSource2 := "data.opentelekomcloud_rms_policy_states_v1.filter_by_compliance_state"
	dataSource3 := "data.opentelekomcloud_rms_policy_states_v1.filter_by_resource_name"
	dataSource4 := "data.opentelekomcloud_rms_policy_states_v1.filter_by_resource_id"
	dataSource5 := "data.opentelekomcloud_rms_policy_states_v1.filter_by_assignment_id"
	rName := acctest.RandomWithPrefix("rms-test-")
	dc1 := common.InitDataSourceCheck(dataSource1)
	dc2 := common.InitDataSourceCheck(dataSource2)
	dc3 := common.InitDataSourceCheck(dataSource3)
	dc4 := common.InitDataSourceCheck(dataSource4)
	dc5 := common.InitDataSourceCheck(dataSource5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceDataSourceRmsPolicyStates_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					dc1.CheckResourceExists(),
					dc2.CheckResourceExists(),
					dc3.CheckResourceExists(),
					dc4.CheckResourceExists(),
					dc5.CheckResourceExists(),
					resource.TestCheckOutput("is_results_not_empty", "true"),
					resource.TestCheckOutput("is_compliance_state_filter_useful", "true"),
					resource.TestCheckOutput("is_resource_name_filter_useful", "true"),
					resource.TestCheckOutput("is_resource_id_filter_useful", "true"),
					resource.TestCheckOutput("is_assignment_id_filter_useful", "true"),
				),
			},
		},
	})
}

func testDataSourceDataSourceRmsPolicyStates_base(name string) string {
	return fmt.Sprintf(`
%[1]s

data "opentelekomcloud_compute_flavor_v2" "test" {
  name = "s3.large.2"
}

data "opentelekomcloud_rms_policy_definitions_v1" "test" {
  name = "allowed-ecs-flavors"
}

resource "opentelekomcloud_rms_policy_assignment_v1" "test" {
  name                 = "%[2]s"
  description          = "An ECS is noncompliant if its flavor is not in the specified flavor list (filter by resource ID)."
  policy_definition_id = try(data.opentelekomcloud_rms_policy_definitions_v1.test.definitions[0].id, "")

  policy_filter {
    region            = "%[3]s"
    resource_provider = "ecs"
    resource_type     = "cloudservers"
    resource_id       = opentelekomcloud_compute_instance_v2.test.id
  }

  parameters = {
    listOfAllowedFlavors = "[\"${data.opentelekomcloud_compute_flavor_v2.test.id}\"]"
  }
}
`, testAccPolicyAssignment_ecsConfig(name), name, env.OS_REGION_NAME)
}

func testDataSourceDataSourceRmsPolicyStates_basic(name string) string {
	return fmt.Sprintf(`
%[1]s

data "opentelekomcloud_rms_policy_states_v1" "basic" {}

data "opentelekomcloud_rms_policy_states_v1" "filter_by_resource_name" {
  resource_name = "%[2]s"

  depends_on = [opentelekomcloud_compute_instance_v2.test]
}

data "opentelekomcloud_rms_policy_states_v1" "filter_by_resource_id" {
  resource_id = opentelekomcloud_compute_instance_v2.test.id
}

data "opentelekomcloud_rms_policy_states_v1" "filter_by_assignment_id" {
  policy_assignment_id = opentelekomcloud_rms_policy_assignment_v1.test.id
}

locals {
  compliance_state_result = [for v in data.opentelekomcloud_rms_policy_states_v1.filter_by_compliance_state.states[*].compliance_state : v == "Compliant"]

  resource_name_filter_result = [for v in data.opentelekomcloud_rms_policy_states_v1.filter_by_resource_name.states[*].resource_name : v == "%[2]s"]

  resource_id_filter_result = [
    for v in data.opentelekomcloud_rms_policy_states_v1.filter_by_resource_id.states[*].resource_id : v == opentelekomcloud_compute_instance_v2.test.id
  ]

  assignment_id_filter_result = [
    for v in data.opentelekomcloud_rms_policy_states_v1.filter_by_assignment_id.states[*].policy_assignment_id :
    v == opentelekomcloud_rms_policy_assignment_v1.test.id
  ]
}

output "is_results_not_empty" {
  value = length(data.opentelekomcloud_rms_policy_states_v1.basic.states) > 0
}

output "is_compliance_state_filter_useful" {
  value = alltrue(local.compliance_state_result) && length(local.compliance_state_result) > 0
}

output "is_resource_name_filter_useful" {
  value = alltrue(local.resource_name_filter_result) && length(local.resource_name_filter_result) > 0
}

output "is_resource_id_filter_useful" {
  value = alltrue(local.resource_id_filter_result) && length(local.resource_id_filter_result) > 0
}

output "is_assignment_id_filter_useful" {
  value = alltrue(local.assignment_id_filter_result) && length(local.assignment_id_filter_result) > 0
}


`, testDataSourceDataSourceRmsPolicyStates_base(name), name)
}
