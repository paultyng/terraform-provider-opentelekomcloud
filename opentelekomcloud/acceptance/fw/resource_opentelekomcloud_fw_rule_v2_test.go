package acceptance

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/fwaas_v2/rules"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func TestAccFWRuleV2_basic(t *testing.T) {
	rule1 := &rules.Rule{
		Name:      "rule_1",
		Protocol:  "udp",
		Action:    "deny",
		IPVersion: 4,
		Enabled:   true,
	}

	rule2 := &rules.Rule{
		Name:                 "rule_1",
		Protocol:             "udp",
		Action:               "deny",
		Description:          "Terraform accept test",
		IPVersion:            4,
		SourceIPAddress:      "1.2.3.4",
		DestinationIPAddress: "4.3.2.0/24",
		SourcePort:           "444",
		DestinationPort:      "555",
		Enabled:              true,
	}

	rule3 := &rules.Rule{
		Name:                 "rule_1",
		Protocol:             "tcp",
		Action:               "allow",
		Description:          "Terraform accept test updated",
		IPVersion:            4,
		SourceIPAddress:      "1.2.3.0/24",
		DestinationIPAddress: "4.3.2.8",
		SourcePort:           "666",
		DestinationPort:      "777",
		Enabled:              false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckFWRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFWRuleV2_basic_1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFWRuleV2Exists("opentelekomcloud_fw_rule_v2.rule_1", rule1),
				),
			},
			{
				Config: testAccFWRuleV2_basic_2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFWRuleV2Exists("opentelekomcloud_fw_rule_v2.rule_1", rule2),
				),
			},
			{
				Config: testAccFWRuleV2_basic_3,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFWRuleV2Exists("opentelekomcloud_fw_rule_v2.rule_1", rule3),
				),
			},
		},
	})
}

func TestAccFWRuleV2_anyProtocol(t *testing.T) {
	rule := &rules.Rule{
		Name:            "rule_1",
		Description:     "Allow any protocol",
		Protocol:        "",
		Action:          "allow",
		IPVersion:       4,
		SourceIPAddress: "192.168.199.0/24",
		Enabled:         true,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckFWRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFWRuleV2_anyProtocol,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFWRuleV2Exists("opentelekomcloud_fw_rule_v2.rule_1", rule),
				),
			},
		},
	})
}

func TestAccFWRuleV2_TCPtoICMP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckFWRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFWRuleV2_TCP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "protocol", "tcp"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "destination_port", "23"),
				),
			},
			{
				Config: testAccFWRuleV2_ICMP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "protocol", "icmp"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "destination_port", ""),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "source_port", ""),
				),
			},
			{
				Config: testAccFWRuleV2_IPv6,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_2", "protocol", "tcp"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_2", "destination_ip_address", "2001:db8::"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_2", "source_port", ""),
				),
			},
		},
	})
}

func TestAccFWRuleV2_TCPUpdate(t *testing.T) {
	var description = "allow ssh from world"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckFWRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFWRuleV2_TCPDescription,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "protocol", "tcp"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "description", description),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "destination_port", "22"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "source_port", "22"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "action", "allow"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "source_ip_address", "0.0.0.0/0"),
				),
			},
			{
				Config: testAccFWRuleV2_TCPDescriptionUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "protocol", "tcp"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "description", description+" updated"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "destination_port", "22"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "source_port", "22"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "action", "allow"),
					resource.TestCheckResourceAttr("opentelekomcloud_fw_rule_v2.rule_1", "source_ip_address", "0.0.0.0/0"),
				),
			},
		},
	})
}

// Customer issue https://github.com/opentelekomcloud/terraform-provider-opentelekomcloud/issues/2711
func TestAccFWRuleV2_emptySourcePort(t *testing.T) {
	rname := "opentelekomcloud_fw_rule_v2.egress_test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckFWRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFWRuleV2emptySourcePort(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rname, "protocol", "udp"),
				),
			},
			{
				Config: testAccFWRuleV2emptySourcePort_update(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rname, "protocol", "udp"),
				),
			},
		},
	})
}

func testAccCheckFWRuleV2Destroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud networking client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_firewall_rule" {
			continue
		}
		_, err = rules.Get(networkingClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("firewall rule (%s) still exists.", rs.Primary.ID)
		}
		if _, ok := err.(golangsdk.ErrDefault404); !ok {
			return err
		}
	}
	return nil
}

func testAccCheckFWRuleV2Exists(n string, expected *rules.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		networkingClient, err := config.NetworkingV2Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud networking client: %s", err)
		}

		var found *rules.Rule
		for i := 0; i < 5; i++ {
			// Firewall rule creation is asynchronous. Retry some times
			// if we get a 404 error. Fail on any other error.
			found, err = rules.Get(networkingClient, rs.Primary.ID).Extract()
			if err != nil {
				if _, ok := err.(golangsdk.ErrDefault404); ok {
					time.Sleep(time.Second)
					continue
				}
				return err
			}
			break
		}

		expected.ID = found.ID
		// Erase the tenant id because we don't want to compare
		// it as long it is not present in the expected
		found.TenantID = ""

		if !reflect.DeepEqual(expected, found) {
			return fmt.Errorf("expected:\n%#v\nFound:\n%#v", expected, found)
		}

		return nil
	}
}

const testAccFWRuleV2_basic_1 = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  name     = "rule_1"
  protocol = "udp"
  action   = "deny"
}
`

const testAccFWRuleV2_basic_2 = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  name                   = "rule_1"
  description            = "Terraform accept test"
  protocol               = "udp"
  action                 = "deny"
  ip_version             = 4
  source_ip_address      = "1.2.3.4"
  destination_ip_address = "4.3.2.0/24"
  source_port            = "444"
  destination_port       = "555"
  enabled                = true
}
`

const testAccFWRuleV2_basic_3 = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  name                   = "rule_1"
  description            = "Terraform accept test updated"
  protocol               = "tcp"
  action                 = "allow"
  ip_version             = 4
  source_ip_address      = "1.2.3.0/24"
  destination_ip_address = "4.3.2.8"
  source_port            = "666"
  destination_port       = "777"
  enabled                = false
}
`

const testAccFWRuleV2_anyProtocol = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  name              = "rule_1"
  description       = "Allow any protocol"
  protocol          = "any"
  action            = "allow"
  ip_version        = 4
  source_ip_address = "192.168.199.0/24"
  enabled           = true
}
`

const testAccFWRuleV2_TCP = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  name             = "rule_1"
  description      = "Tcp protocol"
  protocol         = "tcp"
  ip_version       = 4
  enabled          = true
  action           = "deny"
  destination_port = "23"
}
`

const testAccFWRuleV2_ICMP = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  name        = "rule_1"
  description = "ICMP protocol"
  protocol    = "icmp"
  ip_version  = 4
  enabled     = true
  action      = "allow"
}
`

const testAccFWRuleV2_IPv6 = `
resource "opentelekomcloud_fw_rule_v2" "rule_2" {
  name        = "rule_1"
  description = "Ipv6 deny"
  protocol    = "tcp"
  ip_version  = 6
  enabled     = true
  action      = "deny"

  destination_ip_address = "2001:db8::"
}
`

const testAccFWRuleV2_TCPDescription = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  description            = "allow ssh from world"
  action                 = "allow"
  protocol               = "tcp"
  source_ip_address      = "0.0.0.0/0"
  destination_ip_address = "192.168.100.0/29"
  source_port            = "22"
  destination_port       = "22"

  enabled = "true"
}
`

const testAccFWRuleV2_TCPDescriptionUpdate = `
resource "opentelekomcloud_fw_rule_v2" "rule_1" {
  description            = "allow ssh from world updated"
  action                 = "allow"
  protocol               = "tcp"
  source_ip_address      = "0.0.0.0/0"
  destination_ip_address = "192.168.100.0/29"
  source_port            = "22"
  destination_port       = "22"

  enabled = "true"
}
`

func testAccFWRuleV2emptySourcePort() string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_fw_rule_v2" "egress_test" {

  description            = "egress test"
  action                 = "allow"
  protocol               = "udp"
  source_ip_address      = "192.168.1.0/29"
  destination_ip_address = "0.0.0.0/0"
  destination_port       = "1234"

  enabled = "true"
}

resource "opentelekomcloud_fw_policy_v2" "egress" {
  name = "egress"

  rules = [opentelekomcloud_fw_rule_v2.egress_test.id]
}

data "opentelekomcloud_networking_port_v2" "this" {
  network_id   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  device_owner = "network:router_interface_distributed"
}

resource "opentelekomcloud_fw_firewall_group_v2" "this" {
  name             = "test"
  egress_policy_id = opentelekomcloud_fw_policy_v2.egress.id
  ports = [
    data.opentelekomcloud_networking_port_v2.this.id
  ]
}
`, common.DataSourceSubnet)
}

func testAccFWRuleV2emptySourcePort_update() string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_fw_rule_v2" "egress_test" {

  description            = "egress test"
  action                 = "allow"
  protocol               = "udp"
  source_ip_address      = "192.168.1.0/24"
  destination_ip_address = "0.0.0.0/0"
  destination_port       = "1234"

  enabled = "true"
}

resource "opentelekomcloud_fw_policy_v2" "egress" {
  name = "egress"

  rules = [opentelekomcloud_fw_rule_v2.egress_test.id]
}

data "opentelekomcloud_networking_port_v2" "this" {
  network_id   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  device_owner = "network:router_interface_distributed"
}

resource "opentelekomcloud_fw_firewall_group_v2" "this" {
  name             = "test"
  egress_policy_id = opentelekomcloud_fw_policy_v2.egress.id
  ports = [
    data.opentelekomcloud_networking_port_v2.this.id
  ]
}
`, common.DataSourceSubnet)
}
