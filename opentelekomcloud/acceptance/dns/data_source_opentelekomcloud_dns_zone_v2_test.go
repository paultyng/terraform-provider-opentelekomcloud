package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
)

const dataZoneName = "data.opentelekomcloud_dns_zone_v2.z1"

func TestAccOpenStackDNSZoneV2DataSource_basic(t *testing.T) {
	zone := randomZoneName()
	randZoneTag := fmt.Sprintf("value-%s", acctest.RandString(5))
	dc := common.InitDataSourceCheck(dataZoneName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheckRequiredEnvVars(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      dc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenStackDNSZoneV2DataSourceZone(zone, randZoneTag),
			},
			{
				Config: testAccOpenStackDNSZoneV2DataSourceBasic(zone, randZoneTag),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					testAccCheckDNSZoneV2DataSourceID(dataZoneName),
					resource.TestCheckResourceAttr(dataZoneName, "name", zone),
					resource.TestCheckResourceAttr(dataZoneName, "ttl", "3000"),
					resource.TestCheckResourceAttr(dataZoneName, "zone_type", "public"),
				),
			},
			{
				Config: testAccOpenStackDNSZoneV2DataSourceZone(zone, randZoneTag),
			},
		},
	})
}

func TestAccOpenStackDNSZoneV2DataSource_byTag(t *testing.T) {
	zone1 := randomZoneName()
	zone2 := randomZoneName()
	randZoneTag := fmt.Sprintf("value-%s", acctest.RandString(5))
	dc := common.InitDataSourceCheck(dataZoneName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheckRequiredEnvVars(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      dc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenStackDNSZoneV2DataSourceZone2(zone1, zone2, randZoneTag),
			},
			{
				Config: testAccOpenStackDNSZoneV2DataSourceByTag(zone1, zone2, randZoneTag),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneV2DataSourceID(dataZoneName),
				),
			},
		},
	})
}

func TestAccOpenStackDNSZoneV2DataSource_private(t *testing.T) {
	zone1 := randomZoneName()
	dc := common.InitDataSourceCheck(dataZoneName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheckRequiredEnvVars(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      dc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2ZonePrivate(zone1),
			},
			{
				Config: testAccOpenStackDNSZoneV2DataSourcePrivate(zone1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneV2DataSourceID(dataZoneName),
				),
			},
		},
	})
}

func testAccCheckDNSZoneV2DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find DNS Zone data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DNS Zone data source ID not set")
		}

		return nil
	}
}

const zoneTemplate = `
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email1@example.com"
  description = "a public zone"
  ttl         = 3000
  type        = "public"

  tags = {
    key = "%s"
  }
}
`

const zoneTemplateNoTags = `
resource "opentelekomcloud_dns_zone_v2" "zone_2" {
  name        = "%s"
  email       = "email1@example.com"
  description = "a public zone"
  ttl         = 3000
  type        = "public"
}
`

func testAccOpenStackDNSZoneV2DataSourceZone(zoneName, zoneTagValue string) string {
	return fmt.Sprintf(zoneTemplate, zoneName, zoneTagValue)
}

func testAccOpenStackDNSZoneV2DataSourceBasic(zoneName, zoneTagValue string) string {
	base := testAccOpenStackDNSZoneV2DataSourceZone(zoneName, zoneTagValue)
	return fmt.Sprintf(`
%s
data "opentelekomcloud_dns_zone_v2" "z1" {
  name = "%s"
}
`, base, zoneName)
}

func testAccOpenStackDNSZoneV2DataSourceZone2(zoneName, zone2Name, zoneTagValue string) string {
	zone1 := testAccOpenStackDNSZoneV2DataSourceZone(zoneName, zoneTagValue)
	zone2 := fmt.Sprintf(zoneTemplateNoTags, zone2Name)
	return fmt.Sprintf(`
%s
%s
`, zone1, zone2)
}

func testAccOpenStackDNSZoneV2DataSourceByTag(zoneName, zone2Name, zoneTagValue string) string {
	base := testAccOpenStackDNSZoneV2DataSourceZone2(zoneName, zone2Name, zoneTagValue)
	return fmt.Sprintf(`
%s
data "opentelekomcloud_dns_zone_v2" "z1" {
  tags = {
    key = "%s"
  }
}
`, base, zoneTagValue)
}

func testAccOpenStackDNSZoneV2DataSourcePrivate(zoneName string) string {
	base := testAccDNSV2ZonePrivate(zoneName)
	return fmt.Sprintf(`
%s

data "opentelekomcloud_dns_zone_v2" "z1" {
  name      = "%s"
  zone_type = "private"
}
`, base, zoneName)
}
