package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/clusters"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getCssClusterFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.CssV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating CSS v1 client: %s", err)
	}
	return clusters.Get(client, state.Primary.ID)
}

func TestAccCssClusterRestart_basic(t *testing.T) {
	clusterID := os.Getenv("OS_CSS_CLUSTER_ID")
	if clusterID == "" {
		t.Skip("OS_CSS_CLUSTER_ID env var is not set")
	}
	resourceName := "opentelekomcloud_css_cluster_restart_v1.r"

	var obj clusters.Cluster
	rc := common.InitResourceCheck(
		resourceName,
		&obj,
		getCssClusterFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterRestart_basic(clusterID),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func testAccCssClusterRestart_basic(clusterId string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_css_cluster_restart_v1" "r" {
  cluster_id = "%s"
}
`, clusterId)
}
