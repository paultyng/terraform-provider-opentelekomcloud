package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodepools"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/cce/shared"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
	ecs "github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/ecs"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const nodePoolResourceName = "opentelekomcloud_cce_node_pool_v3.node_pool"

func getNodePoolFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.CceV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud CCE client: %s", err)
	}
	return nodepools.Get(client, state.Primary.Attributes["cluster_id"], state.Primary.ID)
}

func TestAccCCENodePoolsV3_basic(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)

	t.Parallel()
	qts := []*quotas.ExpectedQuota{
		{Q: quotas.Server, Count: 2},
		{Q: quotas.Volume, Count: 2},
		{Q: quotas.VolumeSize, Count: 40 + 100},
	}
	qts = append(qts, ecs.QuotasForFlavor("s2.large.2")...)
	quotas.BookMany(t, qts)
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3Basic,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(nodePoolResourceName, "name", "opentelekomcloud-cce-node-pool"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "flavor", "s2.large.2"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "os", "EulerOS 2.9"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "k8s_tags.kubelet.kubernetes.io/namespace", "muh"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "data_volumes.0.extend_params.useType", "docker"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "taints.0.key", "example.com/node"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "taints.0.value", "infra"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "taints.0.effect", "NoSchedule"),
				),
			},
			{
				Config: testAccCCENodePoolV3Update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(nodePoolResourceName, "initial_node_count", "2"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "k8s_tags.kubelet.kubernetes.io/namespace", "kuh"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "data_volumes.0.extend_params.useType", "docker"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "taints.0.key", "example-updated.com/node"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "taints.0.value", "infra"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "taints.0.effect", "NoExecute"),
				),
			},
			{
				ResourceName:      nodePoolResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccCCENodePoolV3ImportStateIdFunc(),
				ImportStateVerifyIgnore: []string{
					"max_node_count", "min_node_count", "priority",
					"scale_down_cooldown_time", "initial_node_count",
					"root_volume", "taints",
				},
			},
		},
	})
}

func TestAccCCENodePoolsV3_agency(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	qts := []*quotas.ExpectedQuota{
		{Q: quotas.Server, Count: 2},
		{Q: quotas.Volume, Count: 2},
		{Q: quotas.VolumeSize, Count: 40 + 100},
	}
	qts = append(qts, ecs.QuotasForFlavor("s2.large.2")...)
	quotas.BookMany(t, qts)
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3Agency,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(nodePoolResourceName, "name", "opentelekomcloud-cce-node-pool"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "flavor", "s2.large.2"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "os", "EulerOS 2.9"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "k8s_tags.kubelet.kubernetes.io/namespace", "muh"),
					resource.TestCheckResourceAttr(nodePoolResourceName, "data_volumes.0.extend_params.useType", "docker"),
				),
			},
		},
	})
}

func TestAccCCENodePoolsV3_SecurityGroupIds(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	qts := []*quotas.ExpectedQuota{
		{Q: quotas.Server, Count: 2},
		{Q: quotas.Volume, Count: 2},
		{Q: quotas.VolumeSize, Count: 40 + 100},
	}
	qts = append(qts, ecs.QuotasForFlavor("s2.large.2")...)
	quotas.BookMany(t, qts)
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3SecurityGroupIds,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(nodePoolResourceName, "name", "opentelekomcloud-cce-node-pool"),
				),
			},
		},
	})
}

func testAccCCENodePoolV3ImportStateIdFunc() resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		var clusterID string
		var nodePoolID string
		for _, rs := range s.RootModule().Resources {
			if rs.Type == "opentelekomcloud_cce_cluster_v3" {
				clusterID = rs.Primary.ID
			} else if rs.Type == "opentelekomcloud_cce_node_pool_v3" {
				nodePoolID = rs.Primary.ID
			}
		}
		if clusterID == "" || nodePoolID == "" {
			return "", fmt.Errorf("resource not found: %s/%s", clusterID, nodePoolID)
		}
		return fmt.Sprintf("%s/%s", clusterID, nodePoolID), nil
	}
}

func TestAccCCENodePoolsV3_randomAZ(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3RandomAZ,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(nodePoolResourceName, "availability_zone", "random"),
				),
			},
		},
	})
}

func TestAccCCENodePoolsV3EncryptedVolume(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3Encrypted,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(nodePoolResourceName, "data_volumes.0.kms_id", env.OS_KMS_ID),
					resource.TestCheckResourceAttr(nodePoolResourceName, "root_volume.0.kms_id", env.OS_KMS_ID),
				),
			},
		},
	})
}

func TestAccCCENodePoolsV3ExtendParams(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3ExtendParams,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccCCENodePoolsV3Storage(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3Storage,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func TestAccCCENodePoolsV3StorageJsonEncode(t *testing.T) {
	var nodePool nodepools.NodePool
	rc := common.InitResourceCheck(
		nodePoolResourceName,
		&nodePool,
		getNodePoolFunc,
	)
	t.Parallel()
	shared.BookCluster(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCCEKeyPairPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCENodePoolV3StorageJsonEncode,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

var testAccCCENodePoolV3Basic = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  availability_zone  = "%s"
  key_pair           = "%s"
  runtime            = "containerd"

  scale_enable             = false
  min_node_count           = 1
  max_node_count           = 3
  scale_down_cooldown_time = 6
  priority                 = 1

  root_volume {
    size       = 40
    volumetype = "SSD"
  }

  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  k8s_tags = {
    "kubelet.kubernetes.io/namespace" = "muh"
  }

  taints {
    key    = "example.com/node"
    value  = "infra"
    effect = "NoSchedule"
  }
}`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodePoolV3Update = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 2
  availability_zone  = "%s"
  key_pair           = "%s"

  scale_enable             = true
  min_node_count           = 2
  max_node_count           = 9
  scale_down_cooldown_time = 100
  priority                 = 1

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  k8s_tags = {
    "kubelet.kubernetes.io/namespace" = "kuh"
  }

  taints {
    key    = "example-updated.com/node"
    value  = "infra"
    effect = "NoExecute"
  }
}`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodePoolV3RandomAZ = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  key_pair           = "%s"
  availability_zone  = "random"

  scale_enable             = false
  min_node_count           = 0
  max_node_count           = 0
  scale_down_cooldown_time = 0
  priority                 = 0

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
  }
}`, shared.DataSourceCluster, env.OS_KEYPAIR_NAME)

var testAccCCENodePoolV3Encrypted = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  key_pair           = "%s"
  availability_zone  = "random"

  scale_enable             = false
  min_node_count           = 0
  max_node_count           = 0
  scale_down_cooldown_time = 0
  priority                 = 0

  root_volume {
    size       = 40
    volumetype = "SSD"
    kms_id     = "%s"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
    kms_id     = "%s"
  }
}`, shared.DataSourceCluster, env.OS_KEYPAIR_NAME, env.OS_KMS_ID, env.OS_KMS_ID)

var testAccCCENodePoolV3ExtendParams = fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  key_pair           = "%s"
  availability_zone  = "random"

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
  }

  max_pods         = 16
  docker_base_size = 32
}`, shared.DataSourceCluster, env.OS_KEYPAIR_NAME)

var testAccCCENodePoolV3Agency = fmt.Sprintf(`
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

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  availability_zone  = "%s"
  key_pair           = "%s"
  runtime            = "containerd"
  agency_name        = opentelekomcloud_identity_agency_v3.agency_1.name

  scale_enable             = false
  min_node_count           = 1
  max_node_count           = 3
  scale_down_cooldown_time = 6
  priority                 = 1

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  k8s_tags = {
    "kubelet.kubernetes.io/namespace" = "muh"
  }
}`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)

var testAccCCENodePoolV3Storage = fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  key_pair           = "%[2]s"
  availability_zone  = "random"

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
  }

  storage = <<EOF
        {
    "storageSelectors": [
        {
            "name": "cceUse",
            "storageType": "evs",
            "matchLabels": {
                "size": "100",
                "volumeType": "SSD",
                "count": "1",
				"metadataEncrypted": "1",
				"metadataCmkid": "%[3]s"
            }
        }
    ],
    "storageGroups": [
        {
            "name": "vgpaas",
            "selectorNames": [
                "cceUse"
            ],
            "cceManaged": true,
            "virtualSpaces": [
                {
                    "name": "runtime",
                    "size": "90%%"
                },
                {
                    "name": "kubernetes",
                    "size": "10%%"
                }
            ]
        }
    ]
}
EOF

  max_pods         = 16
  docker_base_size = 32
}`, shared.DataSourceCluster, env.OS_KEYPAIR_NAME, env.OS_KMS_ID)

var testAccCCENodePoolV3StorageJsonEncode = fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  key_pair           = "%[2]s"
  availability_zone  = "random"

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
  }

  storage = jsonencode(
    {
      "storageSelectors" : [
        {
          "name" : "cceUse",
          "storageType" : "evs",
          "matchLabels" : {
            "size" : "100",
            "volumeType" : "SSD",
            "count" : "1",
            "metadataEncrypted" : "1",
            "metadataCmkid" : "%[3]s"
          }
        }
      ],
      "storageGroups" : [
        {
          "name" : "vgpaas",
          "selectorNames" : [
            "cceUse"
          ],
          "cceManaged" : true,
          "virtualSpaces" : [
            {
              "name" : "runtime",
              "size" : "90%%"
            },
            {
              "name" : "kubernetes",
              "size" : "10%%"
            }
          ]
        }
      ]
  })

  max_pods         = 16
  docker_base_size = 32
}`, shared.DataSourceCluster, env.OS_KEYPAIR_NAME, env.OS_KMS_ID)

var testAccCCENodePoolV3SecurityGroupIds = fmt.Sprintf(`
%s

resource "opentelekomcloud_networking_secgroup_v2" "test" {
  name        = "secgroup_cce_nodepool_1"
  description = "My cce security group"
}

resource "opentelekomcloud_networking_secgroup_v2" "test2" {
  name        = "secgroup_cce_nodepool_2"
  description = "My cce modepool security group"
}

resource "opentelekomcloud_cce_node_pool_v3" "node_pool" {
  cluster_id         = data.opentelekomcloud_cce_cluster_v3.cluster.id
  name               = "opentelekomcloud-cce-node-pool"
  os                 = "EulerOS 2.9"
  flavor             = "s2.large.2"
  initial_node_count = 1
  availability_zone  = "%s"
  key_pair           = "%s"
  runtime            = "containerd"

  security_group_ids = [opentelekomcloud_networking_secgroup_v2.test.id, opentelekomcloud_networking_secgroup_v2.test2.id]

  scale_enable             = false
  min_node_count           = 1
  max_node_count           = 3
  scale_down_cooldown_time = 6
  priority                 = 1

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }

  k8s_tags = {
    "kubelet.kubernetes.io/namespace" = "muh"
  }
}`, shared.DataSourceCluster, env.OS_AVAILABILITY_ZONE, env.OS_KEYPAIR_NAME)
