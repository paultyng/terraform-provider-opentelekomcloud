package vpcep

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/tools"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
)

const dataSourceServiceName = "data.opentelekomcloud_vpcep_service_v1.service"

func TestDataSourceVPCEPService(t *testing.T) {
	dc := common.InitDataSourceCheck(dataSourceServiceName)
	name := tools.RandomString("tf-test-", 4)
	t.Parallel()
	quotas.BookOne(t, serviceQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      dc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testServiceBasic(name),
			},
			{
				Config: testServiceDSBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceServiceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceServiceName, "port.#", "1"),
					resource.TestCheckResourceAttr(dataSourceServiceName, "server_type", "LB"),
					resource.TestCheckResourceAttr(dataSourceServiceName, "service_type", "interface"),
					resource.TestCheckResourceAttr(dataSourceServiceName, "tags.key", "value"),
					resource.TestCheckResourceAttr(dataSourceServiceName, "connection_count", "0"),
				),
			},
		},
	})
}

func testServiceDSBasic(name string) string {
	return fmt.Sprintf(`
%s

data "opentelekomcloud_vpcep_service_v1" "service" {
  name = "%s"
}
`, testServiceBasic(name), name)
}
