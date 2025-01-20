package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	pc "github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/parameter-configuration"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getCssConfigurationV1ResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := cfg.CssV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating APIG v2 client: %s", err)
	}
	configurations, err := pc.List(c, state.Primary.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving OpenTelekomCloud CSS configuration: %s", err)
	}
	for _, template := range configurations.Templates {
		if template.Value != template.DefaultValue {
			return configurations, nil
		}
	}
	return nil, golangsdk.ErrDefault404{}
}

func TestAccCssConfiguration_basic(t *testing.T) {
	clusterID := os.Getenv("OS_CSS_CLUSTER_ID")
	if clusterID == "" {
		t.Skip("OS_CSS_CLUSTER_ID env var is not set")
	}

	var obj pc.Configurations
	rName := "opentelekomcloud_css_configuration_v1.config"
	rc := common.InitResourceCheck(
		rName,
		&obj,
		getCssConfigurationV1ResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testCssConfigurationV1_basic(clusterID),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "thread_pool_force_merge_size", "3"),
					resource.TestCheckResourceAttr(rName, "http_cors_allow_credentials", "true"),
				),
			},
			{
				Config: testCssConfigurationV1_update(clusterID),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "thread_pool_force_merge_size", "4"),
					resource.TestCheckResourceAttr(rName, "http_cors_allow_credentials", "true"),
					resource.TestCheckResourceAttr(rName, "http_cors_allow_headers", "X-Requested-With, Content-Type"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCssConfigurationV1_basic(clusterID string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_css_configuration_v1" "config" {
  cluster_id                   = "%s"
  thread_pool_force_merge_size = "3"
  http_cors_allow_credentials  = true
}
`, clusterID)
}

func testCssConfigurationV1_update(clusterID string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_css_configuration_v1" "config" {
  cluster_id                   = "%s"
  thread_pool_force_merge_size = "4"
  http_cors_allow_credentials  = true
  http_cors_allow_headers      = "X-Requested-With, Content-Type"
  auto_create_index            = true
}
`, clusterID)
}
