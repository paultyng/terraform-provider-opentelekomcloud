package hss

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

func TestAccDataSourceQuotas_basic(t *testing.T) {
	var (
		dataSource = "data.opentelekomcloud_hss_quotas_v5.test"
		dc         = common.InitDataSourceCheck(dataSource)
		name       = fmt.Sprintf("hss-acc-api%s", acctest.RandString(5))
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceQuotas_basic(name),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.#"),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.0.id"),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.0.version"),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.0.status"),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.0.used_status"),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.0.charging_mode"),
					resource.TestCheckResourceAttrSet(dataSource, "quotas.0.shared_quota"),
				),
			},
		},
	})
}

func testDataSourceQuotas_basic(name string) string {
	return fmt.Sprintf(`
%[1]s

data "opentelekomcloud_hss_quotas_v5" "test" {
  depends_on = [opentelekomcloud_hss_host_protection_v5.protection]
}
`, testAccHostProtection_basic(name))
}
