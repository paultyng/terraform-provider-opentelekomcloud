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

func TestAccIdentityV3Protection_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckAdminOnly(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckIdentityV3ProtectionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3ProtectionPolicyBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3ProtectionPolicyExists(resourceProtectionPolicyName),
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
					testAccCheckIdentityV3ProtectionPolicyExists(resourceProtectionPolicyName),
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

func testAccCheckIdentityV3ProtectionPolicyDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.IdentityV30Client()
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomcloud IdentityV3 client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_identity_protection_policy_v3" {
			continue
		}

		_, err := security.GetOperationProtectionPolicy(client, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching the IAM protection policy")
		}
	}

	return nil
}

func testAccCheckIdentityV3ProtectionPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.IdentityV30Client()
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomcloud IdentityV3 client: %w", err)
		}

		_, err = security.GetOperationProtectionPolicy(client, rs.Primary.ID)
		return err
	}
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
