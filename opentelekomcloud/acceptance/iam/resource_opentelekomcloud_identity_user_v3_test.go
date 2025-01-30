package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	oldusers "github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/users"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3.0/users"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceName = "opentelekomcloud_identity_user_v3.user_1"

func getIdentityUserResourceFunc(c *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := c.IdentityV30Client()
	if err != nil {
		return nil, fmt.Errorf("error creating IAM client: %s", err)
	}
	return users.GetUser(client, state.Primary.ID)
}

func TestAccIdentityV3User_basic(t *testing.T) {
	var user users.User
	userName := acctest.RandomWithPrefix("tf-user")
	rc := common.InitResourceCheck(
		resourceName,
		&user,
		getIdentityUserResourceFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckAdminOnly(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3UserBasic(userName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPtr(resourceName, "name", &user.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "email", "test@acme.org"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
			{
				Config: testAccIdentityV3UserUpdate(userName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPtr(resourceName, "name", &user.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "email", "test2@acme.org"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"send_welcome_email",
				},
			},
		},
	})
}

func TestAccCheckIAMV3EmailValidation(t *testing.T) {
	var name = fmt.Sprintf("ACPTTEST%s.com.", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIdentityV3UserWrongEmail(name),
				ExpectError: regexp.MustCompile(`Error: "email" doesn't comply with email standards+`),
			},
		},
	})
}

func TestAccCheckIAMV3SendEmailValidation(t *testing.T) {
	var name = fmt.Sprintf("ACPTTEST%s.com.", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIdentityV3UserWrongSendEmail(name),
				ExpectError: regexp.MustCompile(`"send_welcome_email":+`),
			},
		},
	})
}

func TestAccIdentityV3User_protection(t *testing.T) {
	var user users.User
	userName := acctest.RandomWithPrefix("tf-user")
	rc := common.InitResourceCheck(
		resourceName,
		&user,
		getIdentityUserResourceFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckAdminOnly(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3UserProtectionConfig(userName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPtr(resourceName, "name", &user.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "email", "test@acme.org"),
					resource.TestCheckResourceAttr(resourceName, "login_protection.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "login_protection.0.verification_method", "email"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
			{
				Config: testAccIdentityV3UserProtectionConfigUpdate(userName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPtr(resourceName, "name", &user.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "email", "test2@acme.org"),
					resource.TestCheckResourceAttr(resourceName, "login_protection.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
			{
				Config: testAccIdentityV3UserProtectionConfigRemove(userName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPtr(resourceName, "name", &user.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "email", "test2@acme.org"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
			{
				Config: testAccIdentityV3UserProtectionConfigReturn(userName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPtr(resourceName, "name", &user.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "email", "test2@acme.org"),
					resource.TestCheckResourceAttr(resourceName, "login_protection.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "login_protection.0.verification_method", "vmfa"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
		},
	})
}

func testAccCheckIdentityV3UserDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.IdentityV3Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomcloud IdentityV3 client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_identity_user_v3" {
			continue
		}

		_, err := oldusers.Get(client, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("user still exists")
		}
	}

	return nil
}

func testAccIdentityV3UserBasic(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name               = "%s"
  password           = "password123@!"
  enabled            = true
  email              = "test@acme.org"
  send_welcome_email = true
}
  `, userName)
}

func testAccIdentityV3UserUpdate(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  enabled  = false
  password = "password123@!"
  email    = "tEst2@acme.org"
}
  `, userName)
}

func testAccIdentityV3UserWrongEmail(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  enabled  = false
  password = "password123@!"
  email    = "tEst2@.org"
}
  `, userName)
}

func testAccIdentityV3UserWrongSendEmail(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name               = "%s"
  enabled            = false
  password           = "password123@!"
  send_welcome_email = true
}
  `, userName)
}

func testAccIdentityV3UserProtectionConfig(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name               = "%s"
  password           = "password123@!"
  enabled            = true
  email              = "test@acme.org"
  send_welcome_email = true

  login_protection {
    enabled             = true
    verification_method = "email"
  }
}
  `, userName)
}

func testAccIdentityV3UserProtectionConfigUpdate(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  enabled  = false
  password = "password123@!"
  email    = "tEst2@acme.org"

  login_protection {
    enabled             = false
    verification_method = "email"
  }
}
  `, userName)
}

func testAccIdentityV3UserProtectionConfigRemove(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  enabled  = false
  password = "password123@!"
  email    = "tEst2@acme.org"
}
  `, userName)
}

func testAccIdentityV3UserProtectionConfigReturn(userName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  enabled  = false
  password = "password123@!"
  email    = "tEst2@acme.org"

  login_protection {
    enabled             = true
    verification_method = "vmfa"
  }
}
  `, userName)
}
