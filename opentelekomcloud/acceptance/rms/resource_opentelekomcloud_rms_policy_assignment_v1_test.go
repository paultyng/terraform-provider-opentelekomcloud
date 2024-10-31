package rms

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/compliance"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/rms"
)

var (
	statusReg = regexp.MustCompile(`^(Enabled|Evaluating)$`)
)

func getPolicyAssignmentResourceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.RmsV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating RMS V1 client: %s", err)
	}

	domainId := rms.GetRmsDomainId(client, conf)

	return compliance.GetRule(client, domainId, state.Primary.ID)
}

func TestAccPolicyAssignment_basic(t *testing.T) {
	var (
		obj compliance.PolicyRule

		rName       = "opentelekomcloud_rms_policy_assignment_v1.test"
		name        = "rms-test" + acctest.RandString(5)
		basicConfig = testAccPolicyAssignment_ecsConfig(name)
	)

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getPolicyAssignmentResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAssignment_basic(basicConfig, name, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "type", rms.AssignmentTypeBuiltin),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttrPair(rName, "policy_definition_id",
						"data.opentelekomcloud_rms_policy_definitions_v1.test", "definitions.0.id"),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.region", env.OS_REGION_NAME),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.resource_provider", "ecs"),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.resource_type", "cloudservers"),
					resource.TestCheckResourceAttrPair(rName, "policy_filter.0.resource_id",
						"opentelekomcloud_compute_instance_v2.test", "id"),
					resource.TestCheckResourceAttr(rName, "status", "Disabled"),
					resource.TestCheckResourceAttrSet(rName, "parameters.listOfAllowedFlavors"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttrSet(rName, "updated_at"),
				),
			},
			{
				Config: testAccPolicyAssignment_basic(basicConfig, name, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestMatchResourceAttr(rName, "status", statusReg),
				),
			},
			{
				Config: testAccPolicyAssignment_basicUpdate(basicConfig, name, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "type", rms.AssignmentTypeBuiltin),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttrPair(rName, "policy_definition_id",
						"data.opentelekomcloud_rms_policy_definitions_v1.test", "definitions.0.id"),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.region", env.OS_REGION_NAME),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.resource_provider", "ecs"),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.resource_type", "cloudservers"),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.tag_key", "foo"),
					resource.TestCheckResourceAttr(rName, "policy_filter.0.tag_value", "bar"),
					resource.TestMatchResourceAttr(rName, "status", statusReg),
					resource.TestCheckResourceAttrSet(rName, "parameters.listOfAllowedFlavors"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttrSet(rName, "updated_at"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPolicyAssignment_ecsConfig(name string) string {
	return fmt.Sprintf(`


resource "opentelekomcloud_vpc_v1" "test" {
  name = "%[1]s"
  cidr = "192.168.0.0/16"
}

resource "opentelekomcloud_vpc_subnet_v1" "test" {
  name       = "%[1]s"
  vpc_id     = opentelekomcloud_vpc_v1.test.id
  cidr       = cidrsubnet(opentelekomcloud_vpc_v1.test.cidr, 4, 1)
  gateway_ip = cidrhost(cidrsubnet(opentelekomcloud_vpc_v1.test.cidr, 4, 1), 1)
}

resource "opentelekomcloud_networking_secgroup_v2" "test" {
  name                 = "%[1]s"
  delete_default_rules = true
}

resource "opentelekomcloud_compute_instance_v2" "test" {
  name              = "%[1]s"
  image_name        = "Standard_Debian_11_latest"
  flavor_name       = "s3.large.2"
  security_groups   = [opentelekomcloud_networking_secgroup_v2.test.name]
  availability_zone = "eu-de-01"

  network {
    uuid = opentelekomcloud_vpc_subnet_v1.test.id
  }
}
`, name)
}

func testAccPolicyAssignment_basic(basicConfig, name, status string) string {
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
  description          = "Test description"
  policy_definition_id = try(data.opentelekomcloud_rms_policy_definitions_v1.test.definitions[0].id, "")
  status               = "%[3]s"

  policy_filter {
    region            = "%[4]s"
    resource_provider = "ecs"
    resource_type     = "cloudservers"
    resource_id       = opentelekomcloud_compute_instance_v2.test.id
  }

  parameters = {
    listOfAllowedFlavors = "[\"${data.opentelekomcloud_compute_flavor_v2.test.id}\"]"
  }
}
`, basicConfig, name, status, env.OS_REGION_NAME)
}

func testAccPolicyAssignment_basicUpdate(basicConfig, name, status string) string {
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
  description          = "Test description"
  policy_definition_id = try(data.opentelekomcloud_rms_policy_definitions_v1.test.definitions[0].id, "")
  status               = "%[3]s"

  policy_filter {
    region            = "%[4]s"
    resource_provider = "ecs"
    resource_type     = "cloudservers"
    tag_key           = "foo"
    tag_value         = "bar"
  }

  parameters = {
    listOfAllowedFlavors = "[\"${data.opentelekomcloud_compute_flavor_v2.test.id}\"]"
  }
}
`, basicConfig, name, status, env.OS_REGION_NAME)
}

func TestAccPolicyAssignment_custom(t *testing.T) {
	var (
		obj compliance.PolicyRule

		rName        = "opentelekomcloud_rms_policy_assignment_v1.test"
		name         = "rms-test-" + acctest.RandString(5)
		customConfig = testAccPolicyAssignment_customConfig(name)
	)

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getPolicyAssignmentResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAssignment_custom(customConfig, name, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "type", rms.AssignmentTypeCustom),
					resource.TestCheckResourceAttr(rName, "description", "Test description"),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "status", "Disabled"),
					resource.TestCheckResourceAttr(rName, "parameters.string_test", "\"string_value\""),
					resource.TestCheckResourceAttr(rName, "parameters.array_test", "[\"array_element\"]"),
					resource.TestCheckResourceAttr(rName, "parameters.object_test", "{\"terraform_version\":\"1.xx.x\"}"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttrSet(rName, "updated_at"),
				),
			},
			{
				Config: testAccPolicyAssignment_custom(customConfig, name, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestMatchResourceAttr(rName, "status", statusReg),
				),
			},
			{
				Config: testAccPolicyAssignment_customUpdate(customConfig, name, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "parameters.string_test", "\"update_string_value\""),
					resource.TestCheckResourceAttr(rName, "parameters.update_array_test", "[\"array_element\"]"),
					resource.TestCheckResourceAttr(rName, "parameters.object_test", "{\"update_terraform_version\":\"1.xx.xx\"}"),
					resource.TestMatchResourceAttr(rName, "status", statusReg),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPolicyAssignment_customConfig(name string) string {
	customConfig := testAccPolicyAssignment_ecsConfig(name)

	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_fgs_function_v2" "test" {
  name                  = "%[2]s"
  code_type             = "inline"
  handler               = "index.handler"
  runtime               = "Node.js10.16"
  functiongraph_version = "v2"
  app                   = "default"
  memory_size           = 128
  timeout               = 3
}
`, customConfig, name)
}

func testAccPolicyAssignment_custom(customConfig, name, status string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_rms_policy_assignment_v1" "test" {
  name        = "%[2]s"
  description = "Test description"
  status      = "%[3]s"

  custom_policy {
    function_urn = "${opentelekomcloud_fgs_function_v2.test.urn}:${opentelekomcloud_fgs_function_v2.test.version}"
    auth_type    = "agency"
    auth_value = {
      agency_name = "\"rms_tracker_agency\""
    }
  }

  parameters = {
    string_test = "\"string_value\""
    array_test  = "[\"array_element\"]"
    object_test = "{\"terraform_version\":\"1.xx.x\"}"
  }
}
`, customConfig, name, status)
}

func testAccPolicyAssignment_customUpdate(customConfig, name, status string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_rms_policy_assignment_v1" "test" {
  name        = "%[2]s"
  description = "Test description"
  status      = "%[3]s"

  custom_policy {
    function_urn = "${opentelekomcloud_fgs_function_v2.test.urn}:${opentelekomcloud_fgs_function_v2.test.version}"
    auth_type    = "agency"
    auth_value = {
      agency_name = "\"rms_tracker_agency\""
    }
  }

  parameters = {
    string_test       = "\"update_string_value\""
    update_array_test = "[\"array_element\"]"
    object_test       = "{\"update_terraform_version\":\"1.xx.xx\"}"
  }
}
`, customConfig, name, status)
}
