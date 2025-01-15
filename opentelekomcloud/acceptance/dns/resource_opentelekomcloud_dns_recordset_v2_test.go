package acceptance

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/recordsets"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/dns"
)

const resourceRecordSetName = "opentelekomcloud_dns_recordset_v2.recordset_1"

func randomZoneName() string {
	// TODO: why does back-end convert name to lowercase?
	return fmt.Sprintf("acpttest-zone-%s.com.", acctest.RandString(5))
}

func getDnsRecordSetFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.DnsV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DNS v2 client: %s", err)
	}
	zoneID, recordsetID, err := dns.ParseDNSV2RecordSetID(state.Primary.ID)
	if err != nil {
		return nil, err
	}
	rs, err := recordsets.Get(c, zoneID, recordsetID).Extract()
	if err != nil {
		if env.OS_REGION_NAME == "eu-nl" {
			c.Endpoint = strings.Replace(c.Endpoint, "eu-nl", "eu-de", 1)
			c.ResourceBase = strings.Replace(c.ResourceBase, "eu-nl", "eu-de", 1)
			rs, err = recordsets.Get(c, zoneID, recordsetID).Extract()
			if err != nil {
				return nil, err
			}
		}
	}
	return rs, err
}

func TestAccDNSV2RecordSet_basic(t *testing.T) {
	var recordset recordsets.RecordSet
	zoneName := randomZoneName()

	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetBasic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceRecordSetName, "name", zoneName),
					resource.TestCheckResourceAttr(resourceRecordSetName, "description", "a record set"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "type", "A"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "ttl", "3000"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "tags.key", "value"),
				),
			},
			{
				Config: testAccDNSV2RecordSetUpdate(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRecordSetName, "ttl", "6000"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "tags.key", "value_updated"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "description", "an updated record set"),
				),
			},
			{
				ResourceName:      resourceRecordSetName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSV2RecordSet_unDotted(t *testing.T) {
	zoneName := randomZoneName()
	zoneName = strings.TrimSuffix(zoneName, ".")
	var recordset recordsets.RecordSet
	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetBasic(zoneName),
			},
		},
	})
}

// TestAccDNSV2RecordSet_childFirst covers #847
func TestAccDNSV2RecordSet_childFirst(t *testing.T) {
	zoneName := randomZoneName()
	var recordset recordsets.RecordSet
	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetChildFirst1(zoneName),
			},
			{
				Config: testAccDNSV2RecordSetChildFirst2(zoneName),
			},
		},
	})
}

func TestAccDNSV2RecordSet_readTTL(t *testing.T) {
	var recordset recordsets.RecordSet
	zoneName := randomZoneName()
	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetReadTTL(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestMatchResourceAttr(resourceRecordSetName, "ttl", regexp.MustCompile("^[0-9]+$")),
				),
			},
		},
	})
}

func TestAccDNSV2RecordSet_timeout(t *testing.T) {
	var recordset recordsets.RecordSet
	zoneName := randomZoneName()
	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetTimeout(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccDNSV2RecordSet_shared(t *testing.T) {
	zoneName := randomZoneName()
	resourceRecordSet2Name := "opentelekomcloud_dns_recordset_v2.recordset_2"
	var recordset recordsets.RecordSet
	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetBasic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRecordSetName, "shared", "false"),
				),
			},
			{
				Config: testAccDNSV2RecordSetReuse(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRecordSetName, "shared", "false"),
					resource.TestCheckResourceAttr(resourceRecordSet2Name, "shared", "true"),
				),
			},
			{
				Config: testAccDNSV2RecordSetBasic(zoneName),
			},
		},
	})
}

func TestAccDNSV2RecordSet_txt(t *testing.T) {
	var recordset recordsets.RecordSet
	zoneName := randomZoneName()
	rc := common.InitResourceCheck(
		resourceRecordSetName,
		&recordset,
		getDnsRecordSetFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSV2RecordSetTxt(zoneName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceRecordSetName, "name", zoneName),
					resource.TestCheckResourceAttr(resourceRecordSetName, "description", "a record set"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "type", "TXT"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "ttl", "300"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "records.0", "v=spf1 include:my.example.try.com -all"),
				),
			},
			{
				Config: testAccDNSV2RecordSetTxtUpdate(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRecordSetName, "ttl", "3000"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "description", "an updated record set"),
					resource.TestCheckResourceAttr(resourceRecordSetName, "records.0", "v=spf1 include:my.example.try.com -none"),
				),
			},
		},
	})
}

func testAccDNSV2RecordSetBasic(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "a zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "%[1]s"
  type        = "A"
  description = "a record set"
  ttl         = 3000
  records     = ["10.1.0.0"]

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, zoneName)
}

func testAccDNSV2RecordSetUpdate(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%s"
  email       = "email2@example.com"
  description = "an updated zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "%s"
  type        = "A"
  description = "an updated record set"
  ttl         = 6000
  records     = ["10.1.0.1"]

  tags = {
    foo = "bar"
    key = "value_updated"
  }
}
`, zoneName, zoneName)
}

func testAccDNSV2RecordSetReadTTL(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "an updated zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id = opentelekomcloud_dns_zone_v2.zone_1.id
  name    = "%[1]s"
  type    = "A"
  records = ["10.1.0.2"]
}
`, zoneName)
}

func testAccDNSV2RecordSetTimeout(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "an updated zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id = opentelekomcloud_dns_zone_v2.zone_1.id
  name    = "%[1]s"
  type    = "A"
  ttl     = 3000
  records = ["10.1.0.3", "10.1.0.2"]

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}
`, zoneName)
}

func testAccDNSV2RecordSetReuse(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "a zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "%[1]s"
  type        = "A"
  description = "a record set"
  ttl         = 3000
  records     = ["10.1.0.0"]

  tags = {
    foo = "bar"
    key = "value"
  }
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_2" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "%[1]s"
  type        = "A"
  description = "a record set"
  ttl         = 3000
  records     = ["10.1.0.0"]

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, zoneName)
}

func testAccDNSV2RecordSetChildFirst1(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "a zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "test.test.%[1]s"
  type        = "A"
  description = "a record set"
  ttl         = 3000
  records     = ["10.1.0.0"]

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, zoneName)
}

func testAccDNSV2RecordSetChildFirst2(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "a zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "test.test.%[1]s"
  type        = "A"
  description = "a record set"
  ttl         = 3000
  records     = ["10.1.0.0"]
}
resource "opentelekomcloud_dns_recordset_v2" "recordset_sup" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "test.%[1]s"
  type        = "A"
  description = "a parent record set"
  ttl         = 3000
  records     = ["10.1.0.0"]
}
`, zoneName)
}

func testAccDNSV2RecordSetTxt(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "a zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "%[1]s"
  type        = "TXT"
  description = "a record set"
  ttl         = 300
  records     = ["v=spf1 include:my.example.try.com -all"]

}
`, zoneName)
}

func testAccDNSV2RecordSetTxtUpdate(zoneName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_dns_zone_v2" "zone_1" {
  name        = "%[1]s"
  email       = "email2@example.com"
  description = "a zone"
  ttl         = 6000
}

resource "opentelekomcloud_dns_recordset_v2" "recordset_1" {
  zone_id     = opentelekomcloud_dns_zone_v2.zone_1.id
  name        = "%[1]s"
  type        = "TXT"
  description = "an updated record set"
  ttl         = 3000
  records     = ["v=spf1 include:my.example.try.com -none"]

}
`, zoneName)
}
