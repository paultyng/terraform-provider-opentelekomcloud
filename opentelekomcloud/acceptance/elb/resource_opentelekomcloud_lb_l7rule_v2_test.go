package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	l7rules "github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/lbaas_v2/l7policies"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func TestAccLBV2L7Rule_basic(t *testing.T) {
	var l7rule l7rules.Rule

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckLBV2L7RuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLBV2L7RuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBV2L7RuleExists("opentelekomcloud_lb_l7rule_v2.l7rule_1", &l7rule),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "type", "PATH"),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "compare_type", "EQUAL_TO"),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "value", "/api"),
					resource.TestMatchResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "listener_id",
						regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
					resource.TestMatchResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "l7policy_id",
						regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
				),
			},
			{
				Config: testAccCheckLBV2L7RuleConfig_update2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBV2L7RuleExists("opentelekomcloud_lb_l7rule_v2.l7rule_1", &l7rule),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "type", "PATH"),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "compare_type", "STARTS_WITH"),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "key", ""),
					resource.TestCheckResourceAttr(
						"opentelekomcloud_lb_l7rule_v2.l7rule_1", "value", "/images"),
				),
			},
		},
	})
}

func testAccCheckLBV2L7RuleDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	lbClient, err := config.NetworkingV2Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud load balancing client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_lb_l7rule_v2" {
			continue
		}

		l7policyID := ""
		for k, v := range rs.Primary.Attributes {
			if k == "l7policy_id" {
				l7policyID = v
				break
			}
		}

		if l7policyID == "" {
			return fmt.Errorf("Unable to find l7policy_id")
		}

		_, err := l7rules.GetRule(lbClient, l7policyID, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("L7 Rule still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckLBV2L7RuleExists(n string, l7rule *l7rules.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		lbClient, err := config.NetworkingV2Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating OpenTelekomCloud load balancing client: %s", err)
		}

		l7policyID := ""
		for k, v := range rs.Primary.Attributes {
			if k == "l7policy_id" {
				l7policyID = v
				break
			}
		}

		if l7policyID == "" {
			return fmt.Errorf("Unable to find l7policy_id")
		}

		found, err := l7rules.GetRule(lbClient, l7policyID, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Policy not found")
		}

		*l7rule = *found

		return nil
	}
}

var testAccCheckLBV2L7RuleConfig = fmt.Sprintf(`
resource "opentelekomcloud_lb_loadbalancer_v2" "loadbalancer_1" {
  name = "loadbalancer_1"
  vip_subnet_id = "%s"
}

resource "opentelekomcloud_lb_listener_v2" "listener_1" {
  name = "listener_1"
  protocol = "HTTP"
  protocol_port = 8080
  loadbalancer_id = opentelekomcloud_lb_loadbalancer_v2.loadbalancer_1.id
}

resource "opentelekomcloud_lb_pool_v2" "pool_1" {
  name            = "pool_1"
  protocol        = "HTTP"
  lb_method       = "ROUND_ROBIN"
  loadbalancer_id = opentelekomcloud_lb_loadbalancer_v2.loadbalancer_1.id
}

resource "opentelekomcloud_lb_l7policy_v2" "l7policy_1" {
  name         = "test"
  action       = "REDIRECT_TO_POOL"
  description  = "test description"
  position     = 1
  listener_id  = opentelekomcloud_lb_listener_v2.listener_1.id
  redirect_pool_id = opentelekomcloud_lb_pool_v2.pool_1.id
}
`, env.OS_SUBNET_ID)

var testAccCheckLBV2L7RuleConfig_basic = fmt.Sprintf(`
%s

resource "opentelekomcloud_lb_l7rule_v2" "l7rule_1" {
  l7policy_id  = opentelekomcloud_lb_l7policy_v2.l7policy_1.id
  type         = "PATH"
  compare_type = "EQUAL_TO"
  value        = "/api"
}
`, testAccCheckLBV2L7RuleConfig)

var testAccCheckLBV2L7RuleConfig_update2 = fmt.Sprintf(`
%s

resource "opentelekomcloud_lb_l7rule_v2" "l7rule_1" {
  l7policy_id  = opentelekomcloud_lb_l7policy_v2.l7policy_1.id
  type         = "PATH"
  compare_type = "STARTS_WITH"
  value        = "/images"
}
`, testAccCheckLBV2L7RuleConfig)
