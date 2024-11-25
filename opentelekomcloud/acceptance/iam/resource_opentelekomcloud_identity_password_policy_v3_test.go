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

const resourcePasswordPolicyName = "opentelekomcloud_identity_password_policy_v3.pol_1"

func TestAccIdentityV3Password_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckAdminOnly(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckIdentityV3PasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3PasswordPolicyBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3PasswordPolicyExists(resourcePasswordPolicyName),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "password_validity_period", "179"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "maximum_consecutive_identical_chars", "0"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "minimum_password_length", "6"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "minimum_password_age", "0"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "number_of_recent_passwords_disallowed", "0"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "password_not_username_or_invert", "true"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "password_char_combination", "4"),
				),
			},
			{
				Config: testAccIdentityV3PasswordPolicyUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3PasswordPolicyExists(resourcePasswordPolicyName),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "password_validity_period", "180"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "maximum_consecutive_identical_chars", "0"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "minimum_password_length", "6"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "minimum_password_age", "0"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "number_of_recent_passwords_disallowed", "0"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "password_not_username_or_invert", "true"),
					resource.TestCheckResourceAttr(resourcePasswordPolicyName, "password_char_combination", "3"),
				),
			},
			{
				ResourceName:      resourcePasswordPolicyName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIdentityV3PasswordPolicyDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.IdentityV30Client()
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomcloud IdentityV3 client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_identity_password_policy_v3" {
			continue
		}

		_, err := security.GetPasswordPolicy(client, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching the IAM account password policy")
		}
	}

	return nil
}

func testAccCheckIdentityV3PasswordPolicyExists(n string) resource.TestCheckFunc {
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

		_, err = security.GetPasswordPolicy(client, rs.Primary.ID)
		return err
	}
}

const testAccIdentityV3PasswordPolicyBasic = `
resource "opentelekomcloud_identity_password_policy_v3" "pol_1" {
  maximum_consecutive_identical_chars   = 0
  minimum_password_length               = 6
  minimum_password_age                  = 0
  number_of_recent_passwords_disallowed = 0
  password_not_username_or_invert       = true
  password_validity_period              = 179
  password_char_combination             = 4
}
`

const testAccIdentityV3PasswordPolicyUpdate = `
resource "opentelekomcloud_identity_password_policy_v3" "pol_1" {
  maximum_consecutive_identical_chars   = 0
  minimum_password_length               = 6
  minimum_password_age                  = 0
  number_of_recent_passwords_disallowed = 0
  password_not_username_or_invert       = true
  password_validity_period              = 180
  password_char_combination             = 3
}
`
