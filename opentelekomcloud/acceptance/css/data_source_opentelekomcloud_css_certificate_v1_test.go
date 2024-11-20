package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const dataCssCertificateName = "data.opentelekomcloud_css_certificate_v1.cert"

func TestAccCSSCertificateV1DataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCSSCertificateV1DataSource,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCSSCertificateV1DataSourceID(dataCssCertificateName),
					resource.TestCheckResourceAttrSet(dataCssCertificateName, "certificate"),
					resource.TestCheckResourceAttrSet(dataCssCertificateName, "region"),
					resource.TestCheckResourceAttrSet(dataCssCertificateName, "project_id"),
				),
			},
		},
	})
}
func testAccCheckCSSCertificateV1DataSourceID(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find backup data source: %s ", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("backup data source ID not set ")
		}

		return nil
	}
}

const testAccCSSCertificateV1DataSource = `
data "opentelekomcloud_css_certificate_v1" "cert" {}
`
