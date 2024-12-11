package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/clusters"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceClusterName = "opentelekomcloud_css_cluster_v1.cluster"

func TestAccCssClusterV1_basic(t *testing.T) {
	name := fmt.Sprintf("css-%s", acctest.RandString(10))
	var cluster clusters.Cluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			quotas.BookMany(t, sharedFlavorQuotas(t, 2, 40))
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterV1Basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "1"),
				),
			},
			{
				Config: testAccCssClusterV1Extend(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "expect_node_num", "2"),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "2"),
				),
			},
			{
				Config: testAccCssClusterV1Extend2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "expect_node_num", "2"),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "2"),
					resource.TestCheckResourceAttr(resourceClusterName, "node_config.0.volume.0.size", "200"),
				),
			},
		},
	})
}

func TestAccCssClusterV1_tags(t *testing.T) {
	name := fmt.Sprintf("css-%s", acctest.RandString(10))
	var cluster clusters.Cluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			quotas.BookMany(t, sharedFlavorQuotas(t, 2, 40))
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterV1Tags(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "tags.say", "hi"),
				),
			},
			{
				Config: testAccCssClusterV1TagsUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "tags.say", "hi"),
					resource.TestCheckResourceAttr(resourceClusterName, "tags.mao", "nihao"),
				),
			},
		},
	})
}

func TestAccCssClusterV1_Opensearch(t *testing.T) {
	name := fmt.Sprintf("css-%s", acctest.RandString(10))
	var cluster clusters.Cluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			quotas.BookMany(t, sharedFlavorQuotas(t, 2, 40))
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterV1Opensearch(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "datastore.0.version", "Opensearch_1.3.6"),
				),
			},
		},
	})
}

func TestAccCssClusterV1_validateDiskAndFlavor(t *testing.T) {
	name := fmt.Sprintf("css-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCssClusterV1TooSmall(name),
				ExpectError: regexp.MustCompile(`invalid disk size.+`),
				PlanOnly:    true,
			},
			{
				Config:      testAccCssClusterV1TooBig(name),
				ExpectError: regexp.MustCompile(`invalid disk size.+`),
				PlanOnly:    true,
			},
			{
				Config:      testAccCssClusterV1FlavorName(name),
				ExpectError: regexp.MustCompile(`can't find flavor with name: .+`),
				PlanOnly:    true,
			},
			{
				Config:             testAccCssClusterV1Basic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCssClusterV1_encrypted(t *testing.T) {
	var cluster clusters.Cluster
	name := fmt.Sprintf("css-%s", acctest.RandString(10))
	if env.OS_KMS_ID == "" {
		t.Skip("OS_KMS_ID is not set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			quotas.BookMany(t, sharedFlavorQuotas(t, 1, 40))
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterV1Encrypted(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(resourceClusterName, &cluster),
					resource.TestCheckResourceAttr(resourceClusterName, "nodes.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "node_config.0.volume.0.encryption_key", env.OS_KMS_ID),
				),
			},
		},
	})
}

func TestAccCSSInstanceV1_importBasic(t *testing.T) {
	var instanceName = fmt.Sprintf("css_%s", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterV1Basic(instanceName),
			},
			{
				ResourceName:      resourceClusterName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"admin_pass",
				},
			},
		},
	})
}

func testAccCheckCssClusterV1Destroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.CssV1Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating CSSv1 client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_css_cluster_v1" {
			continue
		}

		_, err := clusters.Get(client, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("cluster still exists")
		}
	}

	return nil
}

func testAccCheckCssClusterV1Exists(n string, cluster *clusters.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.CssV1Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating CSSv1 client: %w", err)
		}

		found, err := clusters.Get(client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("cluster not found")
		}

		*cluster = *found

		return nil
	}
}

func testAccCssClusterV1Basic(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 40
    }

    availability_zone = "%s"
  }

  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1Tags(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 40
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"

  tags = {
    say = "hi"
  }
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1TagsUpdate(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 40
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"

  tags = {
    say = "hi"
    mao = "nihao"
  }
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1TooSmall(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 1
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1TooBig(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 10000000
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1FlavorName(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.large.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 20
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1Extend(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 2
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 40
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1Extend2(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 2
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 200
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1Encrypted(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type    = "COMMON"
      size           = 40
      encryption_key = "%s"
    }

    availability_zone = "%s"
  }
  datastore {
    version = "7.6.2"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_KMS_ID, env.OS_AVAILABILITY_ZONE)
}

func testAccCssClusterV1Opensearch(name string) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name            = "%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
      network_id        = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
      vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
    }
    volume {
      volume_type = "COMMON"
      size        = 40
    }

    availability_zone = "%s"
  }
  datastore {
    version = "Opensearch_1.3.6"
  }
  enable_https     = true
  enable_authority = true
  admin_pass       = "QwertyUI!"
}
`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}
