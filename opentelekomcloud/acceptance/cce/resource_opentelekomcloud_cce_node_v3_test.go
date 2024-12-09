package acceptance

import (
	"fmt"
	"testing"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/cce/shared"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const (
	resourceNameNode  = "opentelekomcloud_cce_node_v3.node_1"
	resourceNameNode2 = "opentelekomcloud_cce_node_v3.node_2"
	resourceNameNode3 = "opentelekomcloud_cce_node_v3.node_3"
)

func getCceNodeResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.CceV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating CCE v3 Client: %s", err)
	}
	return nodes.Get(client, state.Primary.Attributes["cluster_id"], state.Primary.ID)
}

func TestAccResourceCCENodesV3Basic(t *testing.T) {
	var node nodes.Nodes

	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	ip, _ := cidr.Host(shared.SubnetNet, 200)
	privateIP := ip.String()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3Basic(privateIP),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "name", "test-node"),
					resource.TestCheckResourceAttr(resourceNameNode, "flavor_id", "s2.large.2"),
					resource.TestCheckResourceAttr(resourceNameNode, "os", "EulerOS 2.9"),
					resource.TestCheckResourceAttr(resourceNameNode, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceNameNode, "data_volumes.0.extend_params.useType", "docker"),
				),
			},
			{
				Config: testAccCCENodeV3Update(privateIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameNode, "name", "test-node2"),
					resource.TestCheckResourceAttr(resourceNameNode, "data_volumes.0.extend_params.useType", "docker"),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3Agency(t *testing.T) {
	var node nodes.Nodes

	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	ip, _ := cidr.Host(shared.SubnetNet, 200)
	privateIP := ip.String()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3Agency(privateIP),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "name", "test-node"),
					resource.TestCheckResourceAttr(resourceNameNode, "flavor_id", "s2.large.2"),
					resource.TestCheckResourceAttr(resourceNameNode, "os", "EulerOS 2.9"),
					resource.TestCheckResourceAttr(resourceNameNode, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceNameNode, "data_volumes.0.extend_params.useType", "docker"),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3Multiple(t *testing.T) {
	t.Parallel()
	shared.BookCluster(t)
	quotas.BookMany(t, singleNodeQuotas.X(2))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3Multiple,
			},
		},
	})
}

func TestAccResourceCCENodesV3Timeout(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()
	quotas.BookMany(t, singleNodeQuotas)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3Timeout,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}
func TestAccResourceCCENodesV3OS(t *testing.T) {
	var node2 nodes.Nodes
	var node3 nodes.Nodes

	rc2 := common.InitResourceCheck(
		resourceNameNode2,
		&node2,
		getCceNodeResourceFunc,
	)

	rc3 := common.InitResourceCheck(
		resourceNameNode3,
		&node3,
		getCceNodeResourceFunc,
	)

	t.Parallel()
	shared.BookCluster(t)
	quotas.BookMany(t, singleNodeQuotas.X(2))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCCENodeV3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3OS,
				Check: resource.ComposeTestCheckFunc(
					rc2.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode2, "os", "EulerOS 2.9"),
					rc3.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode3, "os", "Ubuntu 22.04"),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3BandWidthResize(t *testing.T) {
	var node nodes.Nodes

	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()
	qts := quotas.MultipleQuotas{{Q: quotas.FloatingIP, Count: 1}}
	qts = append(qts, singleNodeQuotas...)
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3Ip,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "iptype", "5_bgp"),
					resource.TestCheckResourceAttr(resourceNameNode, "sharetype", "PER"),
					resource.TestCheckResourceAttr(resourceNameNode, "bandwidth_charge_mode", "traffic"),
					resource.TestCheckResourceAttr(resourceNameNode, "bandwidth_size", "100"),
				),
			},
			{
				Config: testAccCCENodeV3BandWidthResize,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "bandwidth_size", "10"),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3_eipIds(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	qts := []*quotas.ExpectedQuota{{Q: quotas.FloatingIP, Count: 2}}
	qts = append(qts, singleNodeQuotas...)
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3IpIDs,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
			{
				Config: testAccCCENodeV3IpIDsUnset,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3IpSetNull(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	qts := []*quotas.ExpectedQuota{{Q: quotas.FloatingIP, Count: 2}}
	qts = append(qts, singleNodeQuotas...)
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3Ip,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "iptype", "5_bgp"),
					resource.TestCheckResourceAttr(resourceNameNode, "sharetype", "PER"),
					resource.TestCheckResourceAttr(resourceNameNode, "bandwidth_charge_mode", "traffic"),
				),
			},
			{
				Config: testAccCCENodeV3IpUnset,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3IpCreate(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()
	qts := []*quotas.ExpectedQuota{{Q: quotas.FloatingIP, Count: 1}}
	qts = append(qts, singleNodeQuotas...)
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3IpUnset,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
			{
				Config: testAccCCENodeV3Ip,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3IpWithExtendedParameters(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()
	qts := []*quotas.ExpectedQuota{{Q: quotas.FloatingIP, Count: 2}}
	qts = append(qts, singleNodeQuotas...)
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3IpParams,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "iptype", "5_bgp"),
					resource.TestCheckResourceAttr(resourceNameNode, "sharetype", "PER"),
					resource.TestCheckResourceAttr(resourceNameNode, "bandwidth_charge_mode", "traffic"),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3IpNulls(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	quotas.BookMany(t, singleNodeQuotas)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3IpNull,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3EncryptedVolume(t *testing.T) {
	var node nodes.Nodes

	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	quotas.BookMany(t, singleNodeQuotas)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3EncryptedVolume,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "data_volumes.0.kms_id", env.OS_KMS_ID),
					resource.TestCheckResourceAttr(resourceNameNode, "root_volume.0.kms_id", env.OS_KMS_ID),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3TaintsK8sTags(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	quotas.BookMany(t, singleNodeQuotas)

	ip, _ := cidr.Host(shared.SubnetNet, 15)
	privateIP := ip.String()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3TaintsK8sTags(privateIP),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceNameNode, "taints.0.key", "dedicated"),
					resource.TestCheckResourceAttr(resourceNameNode, "taints.0.value", "database"),
					resource.TestCheckResourceAttr(resourceNameNode, "taints.0.effect", "NoSchedule"),
					resource.TestCheckResourceAttr(resourceNameNode, "k8s_tags.app", "sometag"),
				),
			},
		},
	})
}

func TestAccResourceCCENodesV3_extendParams(t *testing.T) {
	var node nodes.Nodes
	rc := common.InitResourceCheck(
		resourceNameNode,
		&node,
		getCceNodeResourceFunc,
	)

	shared.BookCluster(t)
	t.Parallel()

	quotas.BookMany(t, singleNodeQuotas)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodeV3ExtendParams,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func testAccCheckCCENodeV3Destroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.CceV3Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud CCE client: %s", err)
	}

	var clusterID string
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "opentelekomcloud_cce_cluster_v3" {
			clusterID = rs.Primary.ID
		}

		if rs.Type != "opentelekomcloud_cce_node_v3" {
			continue
		}

		_, err := nodes.Get(client, clusterID, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("node still exists")
		}
	}

	return nil
}

var testAccCCENodeV3OS = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_2" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node-euler"
  flavor_id  = "s2.large.2"
  os         = "EulerOS 2.9"

  availability_zone = "%[2]s"
  key_pair          = "%[3]s"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}

resource "opentelekomcloud_cce_node_v3" "node_3" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node-ubuntu"
  flavor_id  = "s2.large.2"
  os         = "Ubuntu 22.04"

  availability_zone = "%[2]s"
  key_pair          = "%[3]s"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

func testAccCCENodeV3Basic(privateIP string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node"
  flavor_id  = "s2.large.2"

  availability_zone = "%s"
  key_pair          = "%s"
  runtime           = "containerd"
  os                = "EulerOS 2.9"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  private_ip = "%s"
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME, privateIP)
}

func testAccCCENodeV3Update(privateIP string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node2"
  flavor_id  = "s2.large.2"

  availability_zone = "%s"
  key_pair          = "%s"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  private_ip = "%s"
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME, privateIP)
}

var testAccCCENodeV3Multiple = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node"
  flavor_id  = "s2.large.2"
  count      = 2

  availability_zone = "%s"
  key_pair          = "%s"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3Timeout = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "test-node1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }
  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3Ip = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "cce-node-1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  bandwidth_size = 100

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3BandWidthResize = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "cce-node-1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  bandwidth_size = 10

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3IpUnset = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "cce-node-1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3IpParams = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "cce-node-1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  bandwidth_size = 100
  sharetype      = "PER"
  iptype         = "5_bgp"

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3IpNull = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "cce-node-1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  bandwidth_size = null
  sharetype      = null
  iptype         = null

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3IpIDs = fmt.Sprintf(`
%s

resource "opentelekomcloud_networking_floatingip_v2" "fip_1" {}
resource "opentelekomcloud_networking_floatingip_v2" "fip_2" {}

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "cce-node-1"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  eip_ids = [opentelekomcloud_networking_floatingip_v2.fip_1.id]

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3IpIDsUnset = fmt.Sprintf(`
%s

resource "opentelekomcloud_networking_floatingip_v2" "fip_1" {}
resource "opentelekomcloud_networking_floatingip_v2" "fip_2" {}

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  eip_ids = [opentelekomcloud_networking_floatingip_v2.fip_2.id]

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodeV3EncryptedVolume = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"

  root_volume {
    size       = 40
    volumetype = "SATA"
    kms_id     = "%s"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
    kms_id     = "%s"
  }
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME, env.OS_KMS_ID, env.OS_KMS_ID)

func testAccCCENodeV3TaintsK8sTags(privateIP string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id        = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name              = "test-node"
  flavor_id         = "s2.large.2"
  availability_zone = "%s"
  key_pair          = "%s"
  root_volume {
    size       = 40
    volumetype = "SATA"
  }
  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
  taints {
    key    = "dedicated"
    value  = "database"
    effect = "NoSchedule"
  }

  k8s_tags = {
    "app" = "sometag"
  }
  private_ip = "%s"
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME, privateIP)
}

var testAccCCENodeV3ExtendParams = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node"
  flavor_id  = "s2.large.2"

  availability_zone = "%s"
  key_pair          = "%s"

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }

  max_pods         = 16
  docker_base_size = 30
}`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

func testAccCCENodeV3Agency(privateIP string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_identity_agency_v3" "agency_1" {
  name                  = "test-agency-cce"
  delegated_domain_name = "op_svc_evs"
  project_role {
    project = "eu-de"
    roles = [
      "KMS Administrator",
    ]
  }
}

resource "opentelekomcloud_cce_node_v3" "node_1" {
  cluster_id = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name       = "test-node"
  flavor_id  = "s2.large.2"

  availability_zone = "%s"
  key_pair          = "%s"
  runtime           = "containerd"
  os                = "EulerOS 2.9"
  agency_name       = opentelekomcloud_identity_agency_v3.agency_1.name

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  private_ip = "%s"
}
`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME, privateIP)
}
