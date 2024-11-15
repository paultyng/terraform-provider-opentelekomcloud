package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccKmsImportMaterialParamsV1DataSource_basic(t *testing.T) {
	dataSourceImportName := "data.opentelekomcloud_kms_key_material_parameters_v1.test"
	kmsKeyId := os.Getenv("OS_KMS_ID")
	if kmsKeyId == "" {
		t.Skip("OS_KMS_ID env var is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsDataKeyV1ImportMaterialParams_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						dataSourceImportName, "key_id", kmsKeyId),
					resource.TestCheckResourceAttrSet(
						dataSourceImportName, "import_token"),
					resource.TestCheckResourceAttrSet(
						dataSourceImportName, "public_key"),
					resource.TestCheckResourceAttrSet(
						dataSourceImportName, "expiration_time"),
				),
			},
		},
	})
}

var testAccKmsDataKeyV1ImportMaterialParams_basic = fmt.Sprintf(`
data "opentelekomcloud_kms_key_material_parameters_v1" "test" {
  key_id             = "%s"
  wrapping_algorithm = "RSAES_PKCS1_V1_5"
}
`, env.OS_KMS_ID)
