package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3.0/security"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceProtectionPolicyName = "opentelekomcloud_identity_protection_policy_v3.pol_1"

func getIAMProtectionFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.IdentityV30Client()
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud APIG v2 client: %s", err)
	}

	return security.GetOperationProtectionPolicy(c, state.Primary.ID)
}

func TestAccIdentityV3Protection_basic(t *testing.T) {
	var policy security.ProtectionPolicy
	rc := common.InitResourceCheck(
		resourceProtectionPolicyName,
		&policy,
		getIAMProtectionFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckAdminOnly(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3ProtectionPolicyBasic,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "enable_operation_protection_policy", "true"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.access_key", "false"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.password", "false"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.email", "false"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.mobile", "false"),
				),
			},
			{
				Config: testAccIdentityV3ProtectionPolicyUpdate,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "enable_operation_protection_policy", "false"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.access_key", "true"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.password", "true"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.email", "true"),
					resource.TestCheckResourceAttr(resourceProtectionPolicyName, "self_management.0.mobile", "true"),
				),
			},
			{
				ResourceName:      resourceProtectionPolicyName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccIdentityV3ProtectionPolicyBasic = `
resource "opentelekomcloud_identity_protection_policy_v3" "pol_1" {
  enable_operation_protection_policy = true
  self_management {
    access_key = false
    password   = false
    email      = false
    mobile     = false
  }
}
`

const testAccIdentityV3ProtectionPolicyUpdate = `
resource "opentelekomcloud_identity_protection_policy_v3" "pol_1" {
  enable_operation_protection_policy = false
  self_management {
    access_key = true
    password   = true
    email      = true
    mobile     = true
  }
}
`
