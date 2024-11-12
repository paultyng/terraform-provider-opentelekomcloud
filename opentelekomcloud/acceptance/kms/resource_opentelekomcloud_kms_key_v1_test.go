package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/kms"
)

func TestAccKmsKeyV1_basic(t *testing.T) {
	var key keys.Key
	createName := fmt.Sprintf("kms_%s", acctest.RandString(5))
	updateName := fmt.Sprintf("kms_updated_%s", acctest.RandString(5))
	resourceName := "opentelekomcloud_kms_key_v1.key_1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckKmsV1KeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsV1Key_basic(createName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_alias", createName),
					resource.TestCheckResourceAttr(resourceName, "tags.muh", "value-create"),
				),
			},
			{
				Config: testAccKmsV1Key_update(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_alias", updateName),
					resource.TestCheckResourceAttr(resourceName, "key_description", "key update description"),
					resource.TestCheckResourceAttr(resourceName, "tags.muh", "value-update"),
				),
			},
		},
	})
}

func testAccCheckKmsV1KeyDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.KmsKeyV1Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud KMSv1 client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_kms_key_v1" {
			continue
		}
		v, err := keys.Get(client, rs.Primary.ID)
		if err != nil {
			return err
		}
		if v.KeyState != "4" {
			return fmt.Errorf("key still exists")
		}
	}
	return nil
}

func testAccCheckKmsV1KeyExists(n string, key *keys.Key) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.KmsKeyV1Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud KMSv1 client: %s", err)
		}
		found, err := keys.Get(client, rs.Primary.ID)
		if err != nil {
			return err
		}
		if found.KeyID != rs.Primary.ID {
			return fmt.Errorf("key not found")
		}

		*key = *found
		return nil
	}
}

func TestAccKmsKey_isEnabled(t *testing.T) {
	var key1, key2, key3 keys.Key
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "opentelekomcloud_kms_key_v1.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckKmsV1KeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsKey_enabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckKmsKeyIsEnabled(&key1, true),
				),
			},
			{
				Config: testAccKmsKey_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					testAccCheckKmsKeyIsEnabled(&key2, false),
				),
			},
			{
				Config: testAccKmsKey_enabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckKmsKeyIsEnabled(&key3, true),
				),
			},
		},
	})
}

func testAccCheckKmsKeyIsEnabled(key *keys.Key, isEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if (key.KeyState == kms.EnabledState) != isEnabled {
			return fmt.Errorf("expected key %s to have is_enabled=%t, given %s",
				key.KeyID, isEnabled, key.KeyState)
		}

		return nil
	}
}

func TestAccKmsKey_rotation(t *testing.T) {
	var key keys.Key
	createName := fmt.Sprintf("kms_%s", acctest.RandString(5))
	resourceName := "opentelekomcloud_kms_key_v1.key_1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckKmsV1KeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsKey_rotation(createName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_alias", createName),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rotation_interval", "183"),
				),
			},
		},
	})
}

func TestAccKmsKey_cancelDeletion(t *testing.T) {
	var key keys.Key
	createName := "test_key_gopher"
	resourceName := "opentelekomcloud_kms_key_v1.key_1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckKmsV1KeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsV1Key_cancelDeletion(createName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_alias", createName),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
				),
			},
		},
	})
}

func TestAccKmsKey_cancelDeletionWithRotation(t *testing.T) {
	var key keys.Key
	createName := "test_key_gopher_2"
	resourceName := "opentelekomcloud_kms_key_v1.key_1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckKmsV1KeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsV1Key_cancelDeletionWithRotation(createName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsV1KeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_alias", createName),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
				),
			},
		},
	})
}

func testAccKmsV1Key_basic(rName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "key_1" {
  key_alias = "%s"
  tags = {
    muh = "value-create"
    kuh = "value-create"
  }
}
`, rName)
}

func testAccKmsV1Key_update(rName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "key_1" {
  key_alias       = "%s"
  key_description = "key update description"
  tags = {
    muh = "value-update"
  }
}
`, rName)
}

func testAccKmsKey_enabled(prefix string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "bar" {
  key_alias       = "tf-acc-test-kms-key-%[1]s"
  key_description = "Terraform acc test is_enabled %[1]s"
  pending_days    = "7"
}
`, prefix)
}

func testAccKmsKey_disabled(prefix string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "bar" {
  key_description = "Terraform acc test is_enabled %[1]s"
  pending_days    = "7"
  key_alias       = "tf-acc-test-kms-key-%[1]s"
  is_enabled      = false
}
`, prefix)
}

func testAccKmsKey_rotation(prefix string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "key_1" {
  key_alias         = "%s"
  pending_days      = "7"
  rotation_enabled  = true
  rotation_interval = 183
}`, prefix)
}

func testAccKmsV1Key_cancelDeletion(rName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "key_1" {
  key_alias             = "%s"
  allow_cancel_deletion = true
  key_description       = "some description"
  tags = {
    muh = "value-create"
    kuh = "value-create"
  }
}
`, rName)
}

func testAccKmsV1Key_cancelDeletionWithRotation(rName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_v1" "key_1" {
  key_alias             = "%s"
  key_description       = "A test key"
  rotation_enabled      = true
  rotation_interval     = 90
  allow_cancel_deletion = true
}
`, rName)
}
