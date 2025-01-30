package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/cce/shared"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getAttachedNodeFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.CceV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating CCE v3 client: %s", err)
	}
	return nodes.Get(client, state.Primary.Attributes["cluster_id"], state.Primary.ID)
}

func TestAccNodeAttach_basic(t *testing.T) {
	var (
		node nodes.Nodes

		name         = fmt.Sprintf("cce-node-%s", acctest.RandString(5))
		updateName   = fmt.Sprintf("cce-node-updated-%s", acctest.RandString(5))
		resourceName = "opentelekomcloud_cce_node_attach_v3.test"

		baseConfig = testAccNodeAttach_base()

		rc = common.InitResourceCheck(
			resourceName,
			&node,
			getAttachedNodeFunc,
		)
	)

	shared.BookCluster(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccNodeAttach_basic_step1(baseConfig, name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
					resource.TestCheckResourceAttr(resourceName, "os", "EulerOS 2.9"),
				),
			},
			{
				Config: testAccNodeAttach_basic_step2(baseConfig, updateName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", updateName),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar_update"),
					resource.TestCheckResourceAttr(resourceName, "tags.key_update", "value_update"),
				),
			},
			{
				Config: testAccNodeAttach_basic_step3(baseConfig, updateName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "os", "Ubuntu 22.04"),
				),
			},
		},
	})
}

func testAccNodeAttach_base() string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_ecs_instance_v1" "instance_1" {
  name     = "server_1"
  image_id = "8e36cc3c-2823-456c-b6d8-0b7496a04e28"
  flavor   = "s3.xlarge.1"
  vpc_id   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id

  nics {
    network_id = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  }

  availability_zone = "%[2]s"
  data_disks {
    type = "SSD"
    size = 60
  }

  password                    = "Password@123"
  delete_disks_on_termination = true

  lifecycle {
    ignore_changes = [
      name,
      image_id,
      password,
      key_name,
      tags,
      nics
    ]
  }
}

`, common.DataSourceSubnet, env.OS_AVAILABILITY_ZONE)
}

func testAccNodeAttach_basic_step1(baseConfig, name string) string {
	return fmt.Sprintf(`
%[1]s

%[2]s

resource "opentelekomcloud_cce_node_attach_v3" "test" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  server_id  = opentelekomcloud_ecs_instance_v1.instance_1.id
  key_pair   = "%[3]s"
  os         = "EulerOS 2.9"
  name       = "%[4]s"

  max_pods         = 20
  docker_base_size = 10
  lvm_config       = "dockerThinpool=vgpaas/90%%VG;kubernetesLV=vgpaas/10%%VG"

  k8s_tags = {
    test_key = "test_value"
  }

  taints {
    key    = "test_key"
    value  = "test_value"
    effect = "NoSchedule"
  }

  tags = {
    foo = "bar"
    key = "value"
  }
}
`, baseConfig, shared.DataSourceCluster, env.OS_KEYPAIR_NAME, name)
}

func testAccNodeAttach_basic_step2(baseConfig, name string) string {
	return fmt.Sprintf(`
%[1]s

%[2]s

resource "opentelekomcloud_cce_node_attach_v3" "test" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  server_id  = opentelekomcloud_ecs_instance_v1.instance_1.id
  key_pair   = "%[3]s"
  os         = "EulerOS 2.9"
  name       = "%[4]s"

  max_pods         = 20
  docker_base_size = 10
  lvm_config       = "dockerThinpool=vgpaas/90%%VG;kubernetesLV=vgpaas/10%%VG"

  k8s_tags = {
    test_key = "test_value"
  }

  taints {
    key    = "test_key"
    value  = "test_value"
    effect = "NoSchedule"
  }

  tags = {
    foo        = "bar_update"
    key_update = "value_update"
  }
}
`, baseConfig, shared.DataSourceCluster, env.OS_KEYPAIR_NAME, name)
}

func testAccNodeAttach_basic_step3(baseConfig, name string) string {
	return fmt.Sprintf(`
%[1]s

%[2]s

resource "opentelekomcloud_cce_node_attach_v3" "test" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  server_id  = opentelekomcloud_ecs_instance_v1.instance_1.id
  key_pair   = "%[3]s"
  os         = "Ubuntu 22.04"
  name       = "%[4]s"

  max_pods         = 20
  docker_base_size = 10
  lvm_config       = "dockerThinpool=vgpaas/90%%VG;kubernetesLV=vgpaas/10%%VG"

  k8s_tags = {
    test_key = "test_value"
  }

  taints {
    key    = "test_key"
    value  = "test_value"
    effect = "NoSchedule"
  }

  tags = {
    foo        = "bar_update"
    key_update = "value_update"
  }
}
`, baseConfig, shared.DataSourceCluster, env.OS_KEYPAIR_NAME, name)
}
