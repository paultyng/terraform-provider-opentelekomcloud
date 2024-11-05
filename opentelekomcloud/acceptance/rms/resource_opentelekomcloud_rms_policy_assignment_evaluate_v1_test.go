package rms

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
)

func TestAccPolicyAssignmentEvaluate_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("rms-test")
	basicConfig := testAccPolicyAssignment_ecsConfig(name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAssignmentEvaluate_basic(basicConfig, name),
			},
		},
	})
}

func testAccPolicyAssignmentEvaluate_basic(basicConfig, name string) string {
	return fmt.Sprintf(
		`
%[1]s

data "opentelekomcloud_rms_policy_definitions_v1" "test" {
  name = "allowed-ecs-flavors"
}

data "opentelekomcloud_compute_flavor_v2" "test" {
  name = "s3.large.2"
}

resource "opentelekomcloud_rms_policy_assignment_v1" "test" {
  name                 = "%[2]s"
  description          = "Noncompliant ECS"
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

resource "opentelekomcloud_rms_policy_assignment_evaluate_v1" "test" {
  policy_assignment_id = opentelekomcloud_rms_policy_assignment_v1.test.id
}
`, basicConfig, name, env.OS_REGION_NAME)
}
