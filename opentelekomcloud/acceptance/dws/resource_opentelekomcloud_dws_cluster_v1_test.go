package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dws/v1/cluster"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceInstanceName = "opentelekomcloud_dws_cluster_v1.cluster_1"

func TestAccDwsClusterV1_basic(t *testing.T) {
	var cls cluster.ClusterDetail
	var clusterName = fmt.Sprintf("dws_cluster_%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDwsV1ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDwsV1ClusterBasic(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDwsV1ClusterExists(resourceInstanceName, &cls),
					resource.TestCheckResourceAttr(resourceInstanceName, "name", clusterName),
					resource.TestCheckResourceAttr(resourceInstanceName, "number_of_node", "3"),
				),
			},
			{
				Config: testAccDwsV1ClusterUpdated(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceInstanceName, "name", clusterName),
					resource.TestCheckResourceAttr(resourceInstanceName, "number_of_node", "6"),
				),
			},
			{
				ResourceName:      resourceInstanceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"user_pwd", "number_of_cn",
				},
			},
		},
	})
}

func testAccCheckDwsV1ClusterDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.DwsV1Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating DWSv1 client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_dws_cluster_v1" {
			continue
		}

		_, err := cluster.ListClusterDetails(client, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("DWS cluster still exists")
		}
	}
	return nil
}

func testAccCheckDwsV1ClusterExists(n string, cls *cluster.ClusterDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		dwsClient, err := config.DwsV1Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating DWSv1 client: %w", err)
		}

		v, err := cluster.ListClusterDetails(dwsClient, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting cluster (%s): %w", rs.Primary.ID, err)
		}

		if v.Id != rs.Primary.ID {
			return fmt.Errorf("DWS cluster not found")
		}
		*cls = *v
		return nil
	}
}

func testAccDwsV1ClusterBasic(clusterName string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_dws_cluster_v1" "cluster_1" {
  name              = "%s"
  user_name         = "dbadmin"
  user_pwd          = "#dbadmin12345"
  node_type         = "dws.m3.xlarge"
  number_of_node    = 3
  network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  availability_zone = "%s"
  port              = 8899

  public_ip {
    public_bind_type = "auto_assign"
  }
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, clusterName, env.OS_AVAILABILITY_ZONE)
}

// extend not stable, skip this for now
func testAccDwsV1ClusterUpdated(clusterName string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_dws_cluster_v1" "cluster_1" {
  name              = "%s"
  user_name         = "dbadmin"
  user_pwd          = "#dbadmin12345"
  node_type         = "dws.m3.xlarge"
  number_of_node    = 6
  network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  availability_zone = "%s"
  port              = 8899

  public_ip {
    public_bind_type = "auto_assign"
  }
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, clusterName, env.OS_AVAILABILITY_ZONE)
}
