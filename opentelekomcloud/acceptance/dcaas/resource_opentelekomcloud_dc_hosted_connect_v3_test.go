package dcaas

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hosted_connect "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/hosted-connect"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceDCHostedConnectName = "opentelekomcloud_dc_hosted_connect_v3.hc"

func getHostedConnectResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := cfg.DCaaSV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DCaaS v3 client: %s", err)
	}
	return hosted_connect.Get(c, state.Primary.ID)
}

func TestAccHostedConnectV3_basic(t *testing.T) {
	var hc hosted_connect.HostedConnect
	name := fmt.Sprintf("dc_acc_hc%s", acctest.RandString(5))
	rc := common.InitResourceCheck(
		resourceDCHostedConnectName,
		&hc,
		getHostedConnectResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckDcHostedConnection(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testHostedConnectV3_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "name", name),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "description", "create"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "bandwidth", "10"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "hosting_id", env.OS_DC_HOSTING_ID),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "vlan", "441"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "status", "ACTIVE"),
				),
			},
			{
				Config: testHostedConnectV3_update(name + "update"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "name", name+"update"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "description", "update"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "bandwidth", "12"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "hosting_id", env.OS_DC_HOSTING_ID),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "vlan", "441"),
					resource.TestCheckResourceAttr(resourceDCHostedConnectName, "status", "ACTIVE"),
				),
			},
			{
				ResourceName:      resourceDCHostedConnectName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testHostedConnectV3_basic(name string) string {
	return fmt.Sprintf(`

data "opentelekomcloud_identity_project_v3" "project" {
  name = "eu-ch2"
}

resource "opentelekomcloud_dc_hosted_connect_v3" "hc" {
  name               = "%s"
  description        = "create"
  resource_tenant_id = data.opentelekomcloud_identity_project_v3.project.id
  hosting_id         = "%s"
  vlan               = 441
  bandwidth          = 10
}
`, name, env.OS_DC_HOSTING_ID)
}

func testHostedConnectV3_update(name string) string {
	return fmt.Sprintf(`
data "opentelekomcloud_identity_project_v3" "project" {
  name = "eu-ch2"
}

resource "opentelekomcloud_dc_hosted_connect_v3" "hc" {
  name               = "%s"
  description        = "update"
  resource_tenant_id = data.opentelekomcloud_identity_project_v3.project.id
  hosting_id         = "%s"
  vlan               = 441
  bandwidth          = 12
}
`, name, env.OS_DC_HOSTING_ID)
}
