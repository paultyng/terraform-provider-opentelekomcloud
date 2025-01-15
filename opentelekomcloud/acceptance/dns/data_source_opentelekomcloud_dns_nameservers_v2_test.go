package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const dataDNSNameserverName = "data.opentelekomcloud_dns_nameservers_v2.nameservers"

func TestAccDNSV2NameserverDataSource_basic(t *testing.T) {
	zoneName := randomZoneName()
	dc := common.InitDataSourceCheck(dataZoneName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheckRequiredEnvVars(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      dc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2NameserverDataSource_nameserver(zoneName),
				Check:  resource.TestCheckResourceAttr(dataDNSNameserverName, "nameservers.0.hostname", "ns1.open-telekom-cloud.com."),
			},
		},
	})
}

func testAccDNSV2NameserverDataSource_nameserver(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email1@example.com"
  description = "test public zone"
  ttl         = 3000
  type        = "public"
}

data "opentelekomcloud_dns_nameservers_v2" "nameservers" {
  zone_id = opentelekomcloud_dns_zone_v2.zone_1.id
}
`, zoneName)
}
