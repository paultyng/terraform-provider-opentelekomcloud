package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/kms"
)

func getKmsKeyMaterialResourceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.KmsKeyV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating KMS client: %s", err)
	}

	key, err := keys.Get(client, state.Primary.ID)

	if key.KeyState == kms.PendingDeletionState || key.KeyState == kms.WaitingImportState {
		return nil, golangsdk.ErrDefault404{}
	}

	return key, err
}

func TestAccKmsKeyMaterial_Symmetric(t *testing.T) {
	var resourceName = "opentelekomcloud_kms_key_material_v1.test"
	var key keys.Key

	rc := common.InitResourceCheck(
		resourceName,
		&key,
		getKmsKeyMaterialResourceFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// The key status must be pending import.
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccKmsKeyMaterial_Symmetric(t),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "key_state", "2"),
					resource.TestCheckResourceAttr(resourceName, "expiration_time", "2999886177"),
					resource.TestCheckResourceAttr(resourceName, "region", env.OS_REGION_NAME),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"import_token", "encrypted_key_material",
				},
			},
		},
	})
}

func testAccKmsKeyMaterial_Symmetric(t *testing.T) string {
	client, err := common.TestAccProvider.Meta().(*cfg.Config).KmsKeyV1Client(env.OS_REGION_NAME)
	if err != nil {
		t.Fatalf("error creating KMS client: %s", err)
	}

	opts := keys.GetCMKImportOpts{
		KeyId:             env.OS_KMS_ID,
		WrappingAlgorithm: "RSAES_PKCS1_V1_5",
	}

	params, err := keys.GetCMKImport(client, opts)
	if err != nil {
		t.Fatalf("error getting CMK import parameters: %s", err)
	}

	keyMaterial, err := generateKeyMaterial(params.PublicKey)
	if err != nil {
		t.Fatalf("error generating key material: %s", err)
	}

	return fmt.Sprintf(`
resource "opentelekomcloud_kms_key_material_v1" "test" {
  key_id                 = "%[1]s"
  import_token           = "%[2]s"
  encrypted_key_material = "%[3]s"
  expiration_time        = "2999886177"
}
`, env.OS_KMS_ID, params.ImportToken, keyMaterial)
}
