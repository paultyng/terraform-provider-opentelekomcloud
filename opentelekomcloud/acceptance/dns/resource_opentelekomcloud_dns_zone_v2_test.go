package acceptance

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/zones"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceZoneName = "opentelekomcloud_dns_zone_v2.zone_1"

func getDnsZoneFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.DnsV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DNS v2 client: %s", err)
	}
	if state.Primary.Attributes["type"] == "public" && env.OS_REGION_NAME == "eu-nl" {
		c.Endpoint = strings.Replace(c.Endpoint, "eu-nl", "eu-de", 1)
		c.ResourceBase = strings.Replace(c.ResourceBase, "eu-nl", "eu-de", 1)
	}
	return zones.Get(c, state.Primary.ID).Extract()
}

func TestAccDNSV2Zone_basic(t *testing.T) {
	// TODO: Why does it lowercase names in back-end?
	var (
		zone     zones.Zone
		zoneName = fmt.Sprintf("accbasictest%s.com.", acctest.RandString(5))
	)

	rc := common.InitResourceCheck(
		resourceZoneName,
		&zone,
		getDnsZoneFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2ZoneBasic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceZoneName, "description", "a public zone"),
					resource.TestCheckResourceAttr(resourceZoneName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceZoneName, "tags.key", "value"),
				),
			},
			{
				ResourceName:      resourceZoneName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDNSV2ZoneUpdate(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceZoneName, "name", zoneName),
					resource.TestCheckResourceAttr(resourceZoneName, "email", "email2@example.com"),
					resource.TestCheckResourceAttr(resourceZoneName, "ttl", "6000"),
					// TODO: research why this is blank...
					// resource.TestCheckResourceAttr(resourceZoneName, "type", "PRIMARY"),
					resource.TestCheckResourceAttr(resourceZoneName, "description", "an updated zone"),
					resource.TestCheckResourceAttr(resourceZoneName, "tags.key", "value_updated"),
				),
			},
		},
	})
}

func TestAccDNSV2Zone_unDotted(t *testing.T) {
	var (
		zone     zones.Zone
		zoneName = randomZoneName()
	)

	rc := common.InitResourceCheck(
		resourceZoneName,
		&zone,
		getDnsZoneFunc,
	)
	zoneName = strings.TrimSuffix(zoneName, ".")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2ZoneBasic(zoneName),
			},
		},
	})
}

func TestAccDNSV2Zone_private(t *testing.T) {
	var (
		zone     zones.Zone
		zoneName = fmt.Sprintf("accprivatetest%s.com.", acctest.RandString(5))
	)

	rc := common.InitResourceCheck(
		resourceZoneName,
		&zone,
		getDnsZoneFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2ZonePrivate(zoneName),
				// ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceZoneName, "description", "a private zone"),
					resource.TestCheckResourceAttr(resourceZoneName, "type", "private"),
					resource.TestCheckResourceAttr(resourceZoneName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceZoneName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccDNSV2Zone_readTTL(t *testing.T) {
	var (
		zone     zones.Zone
		zoneName = fmt.Sprintf("accttltest%s.com.", acctest.RandString(5))
	)

	rc := common.InitResourceCheck(
		resourceZoneName,
		&zone,
		getDnsZoneFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2ZoneReadTTL(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestMatchResourceAttr(resourceZoneName, "ttl", regexp.MustCompile("^[0-9]+$")),
				),
			},
		},
	})
}

func TestAccDNSV2Zone_timeout(t *testing.T) {
	var zone zones.Zone
	var zoneName = fmt.Sprintf("acctimeouttest%s.com.", acctest.RandString(5))
	rc := common.InitResourceCheck(
		resourceZoneName,
		&zone,
		getDnsZoneFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2ZoneTimeout(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccCheckDNSV2Zone_routerValidation(t *testing.T) {
	var zoneName = fmt.Sprintf("ACPTTEST%s.com.", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDNSV2ZoneWrongRouterSetting(zoneName),
				ExpectError: regexp.MustCompile(`region is invalid.+`),
			},
		},
	})
}

func testAccDNSV2ZoneBasic(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email1@example.com"
  description = "a public zone"
  ttl         = 3000
  type        = "public"

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, zoneName)
}

func testAccDNSV2ZonePrivate(zoneName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email1@example.com"
  description = "a private zone"
  ttl         = 3000
  type        = "private"

  router {
    router_id     = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    router_region = "%s"
  }
  tags = {
    foo = "bar"
    key = "value"
  }
}
`, common.DataSourceSubnet, zoneName, env.OS_REGION_NAME)
}

func testAccDNSV2ZoneUpdate(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email2@example.com"
  description = "an updated zone"
  ttl         = 6000
  type        = "public"

  tags = {
    foo = "bar"
    key = "value_updated"
  }
}
`, zoneName)
}

func testAccDNSV2ZoneReadTTL(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name  = "%s"
  email = "email1@example.com"
}
`, zoneName)
}

func testAccDNSV2ZoneTimeout(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name  = "%s"
  email = "email@example.com"
  ttl   = 3000

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}
`, zoneName)
}

func testAccDNSV2ZoneWrongRouterSetting(zoneName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email1@example.com"
  description = "a private zone"
  ttl         = 3000
  type        = "private"

  router {
    router_id     = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    router_region = "BAD"
  }
  tags = {
    foo = "bar"
    key = "value"
  }
}
`, common.DataSourceSubnet, zoneName)
}
