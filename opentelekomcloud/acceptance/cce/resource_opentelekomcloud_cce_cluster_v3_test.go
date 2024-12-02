package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceClusterName = "opentelekomcloud_cce_cluster_v3.cluster_1"

func getCceClusterResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.CceV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating CCE v3 Client: %s", err)
	}
	return clusters.Get(client, state.Primary.ID)
}

func TestAccCCEClusterV3_basic(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()

	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3Basic(clusterName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceClusterName, "name", clusterName),
					resource.TestCheckResourceAttr(resourceClusterName, "status", "Available"),
					resource.TestCheckResourceAttr(resourceClusterName, "cluster_type", "VirtualMachine"),
					resource.TestCheckResourceAttr(resourceClusterName, "flavor_id", "cce.s1.small"),
					resource.TestCheckResourceAttr(resourceClusterName, "container_network_type", "overlay_l2"),
					resource.TestCheckResourceAttr(resourceClusterName, "authentication_mode", "x509"),
					resource.TestCheckResourceAttr(resourceClusterName, "kube_proxy_mode", "ipvs"),
					resource.TestCheckResourceAttr(resourceClusterName, "kubernetes_svc_ip_range", "10.247.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceClusterName, "security_group_control"),
					resource.TestCheckResourceAttrSet(resourceClusterName, "security_group_node"),
					resource.TestCheckResourceAttr(resourceClusterName, "certificate_clusters.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "certificate_clusters.0.name", "internalCluster"),
					resource.TestCheckResourceAttr(resourceClusterName, "certificate_users.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "certificate_users.0.name", "user"),
					resource.TestCheckResourceAttr(resourceClusterName, "enable_volume_encryption", "true"),
				),
			},
			{
				Config: testAccCCEClusterV3Update(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceClusterName, "description", "new description"),
					resource.TestCheckResourceAttr(resourceClusterName, "kube_proxy_mode", "ipvs"),
					resource.TestCheckResourceAttr(resourceClusterName, "enable_volume_encryption", "true"),
				),
			},
		},
	})
}

func TestAccCCEClusterV3_turbo_basic(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()

	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterTurboV3Basic(clusterName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceClusterName, "name", clusterName),
					resource.TestCheckResourceAttr(resourceClusterName, "status", "Available"),
					resource.TestCheckResourceAttr(resourceClusterName, "cluster_type", "VirtualMachine"),
					resource.TestCheckResourceAttr(resourceClusterName, "flavor_id", "cce.s1.small"),
					resource.TestCheckResourceAttr(resourceClusterName, "container_network_type", "eni"),
					resource.TestCheckResourceAttr(resourceClusterName, "authentication_mode", "x509"),
					resource.TestCheckResourceAttr(resourceClusterName, "kube_proxy_mode", "iptables"),
					resource.TestCheckResourceAttr(resourceClusterName, "kubernetes_svc_ip_range", "10.247.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceClusterName, "security_group_control"),
					resource.TestCheckResourceAttrSet(resourceClusterName, "security_group_node"),
				),
			},
		},
	})
}

func TestAccCCEClusterV3_importBasic(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()
	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3BasicSG(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceClusterName, "security_group_id")),
			},
			{
				ResourceName:      resourceClusterName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_version", "installed_addons", "ignore_addons",
				},
			},
		},
	})
}

func TestAccCCEClusterV3_invalidNetwork(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()

	quotas.BookOne(t, quotas.CCEClusterQuota)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testAccCCEClusterV3InvalidSubnet(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`can't find subnet.+`),
			},
			{
				Config:      testAccCCEClusterV3InvalidVPC(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`can't find VPC.+`),
			},
			{
				Config:             testAccCCEClusterV3Computed(clusterName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCCEClusterV3_proxyAuth(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()

	quotas.BookOne(t, quotas.CCEClusterQuota)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3AuthProxy(clusterName),
				Check:  resource.TestCheckResourceAttr(resourceClusterName, "authentication_mode", "authenticating_proxy"),
			},
		},
	})
}

func TestAccCCEClusterV3_timeout(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()
	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3Timeout(clusterName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceClusterName, "authentication_mode", "rbac"),
				),
			},
		},
	})
}

func TestAccCCEClusterV3NoAddons(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	t.Parallel()
	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3NoAddons(randClusterName()),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceClusterName, "installed_addons.#", "0"),
				),
			},
		},
	})
}

func TestAccCCEClusterV3NoUserClusterDataSet(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	t.Parallel()

	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3NoUserClusterData(randClusterName()),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceClusterName, "certificate_clusters.#", "0"),
					resource.TestCheckResourceAttr(resourceClusterName, "certificate_users.#", "0"),
				),
			},
		},
	})
}

func TestAccCCEClusterV3_withVersionDiff(t *testing.T) {
	var cluster clusters.Clusters
	rc := common.InitResourceCheck(
		resourceClusterName,
		&cluster,
		getCceClusterResourceFunc,
	)
	clusterName := randClusterName()
	t.Parallel()
	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterV3WithInvalidVersion(clusterName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceClusterName, "name", clusterName),
				),
			},
		},
	})
}

func testAccCCEClusterV3Basic(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                     = "%s"
  cluster_type             = "VirtualMachine"
  flavor_id                = "cce.s1.small"
  vpc_id                   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id                = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type   = "overlay_l2"
  kubernetes_svc_ip_range  = "10.247.0.0/16"
  ignore_addons            = true
  kube_proxy_mode          = "ipvs"
  enable_volume_encryption = true
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3BasicSG(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                     = "%s"
  cluster_type             = "VirtualMachine"
  flavor_id                = "cce.s1.small"
  vpc_id                   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id                = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type   = "overlay_l2"
  kubernetes_svc_ip_range  = "10.247.0.0/16"
  ignore_addons            = true
  kube_proxy_mode          = "ipvs"
  enable_volume_encryption = true
  security_group_id        = opentelekomcloud_networking_secgroup_v2.test.id
}

resource "opentelekomcloud_networking_secgroup_v2" "test" {
  name        = "secgroup_cce"
  description = "My cce security group"
}

`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterTurboV3Basic(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id               = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type  = "eni"
  kubernetes_svc_ip_range = "10.247.0.0/16"
  ignore_addons           = true
  eni_subnet_id           = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.subnet_id
  eni_subnet_cidr         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.cidr
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3Update(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                     = "%s"
  cluster_type             = "VirtualMachine"
  flavor_id                = "cce.s1.small"
  vpc_id                   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id                = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type   = "overlay_l2"
  description              = "new description"
  kubernetes_svc_ip_range  = "10.247.0.0/16"
  ignore_addons            = true
  kube_proxy_mode          = "ipvs"
  delete_all_storage       = "true"
  delete_all_network       = "true"
  enable_volume_encryption = true
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3Timeout(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                   = "%s"
  cluster_type           = "VirtualMachine"
  flavor_id              = "cce.s2.small"
  vpc_id                 = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id              = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type = "overlay_l2"
  authentication_mode    = "rbac"
  timeouts {
    create = "20m"
    delete = "10m"
  }

  multi_az = true
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3WithInvalidVersion(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                   = "%s"
  cluster_type           = "VirtualMachine"
  flavor_id              = "cce.s1.small"
  cluster_version        = "v1.19.8"
  vpc_id                 = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id              = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type = "overlay_l2"
  description            = "new description"
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3AuthProxy(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id               = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type  = "overlay_l2"
  kubernetes_svc_ip_range = "10.247.0.0/16"
  authentication_mode     = "authenticating_proxy"
  authenticating_proxy {
    ca          = <<EOT
-----BEGIN CERTIFICATE-----
MIICZjCCAc+gAwIBAgIUZtMIBg4MdR/h8yPITTx5+B0Xj0swDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMTA5MjExNDEyMThaFw0zMTA5
MTkxNDEyMThaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwgZ8wDQYJKoZIhvcNAQEB
BQADgY0AMIGJAoGBANSP54SyxRLkWjnGkLQcUOSxM6FUfk8k+PvSgq9xF0CeO8BR
NbY+y/+Jr9l1k0XRjAajIe/pdD4Yta24ox9yH5+ay+eKIFjdr2DAsyD5SSnW+A3o
e1eYAwZhWWgGRgq8VG6yik1vm1nsJW8crj+3k/N0kjiLYFGRTelpSThnesb3AgMB
AAGjUzBRMB0GA1UdDgQWBBT2daqVn+fyu7qoWflS7WwQ9EosODAfBgNVHSMEGDAW
gBT2daqVn+fyu7qoWflS7WwQ9EosODAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3
DQEBCwUAA4GBAEoaBV1HA5wSDEGu1blKUtda13IxBg3ZpHWxwOpDS3gwXYzEOPAL
jgKvjKkLJaJnTp8V+YtO1xBhB272ZgWbr22Cer8TQZchNc16I2qLp+O9AQuPqVYO
15xHZN4yCgCVYcSlUm/HW2tJ3lAmilxkEFvJJcK1uLh7vqMflmcPSLe5
-----END CERTIFICATE-----
EOT
    cert        = <<EOT
-----BEGIN CERTIFICATE-----
MIIByDCCATECFAzNYuav3B0dfSfIe7L8pDJ0t0LuMA0GCSqGSIb3DQEBCwUAMEUx
CzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQwHhcNMjEwOTIxMTQzMDM3WhcNMzEwOTE5MTQz
MDM3WjBFMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UE
CgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMFwwDQYJKoZIhvcNAQEBBQADSwAw
SAJBALcrRkYvf/pLJQQp21KCPk56AuWh0UxbMd75NOZQWrY1QFXTRJ0xYHAGa/LG
LgposjKO9BELu+AYe+UoIJVwdAkCAwEAATANBgkqhkiG9w0BAQsFAAOBgQBwHUAE
Pzs5sfqJgJe0YDNZSC+V/+NZ67AYTDBzYt35dNEqUxbnsOaoMjuWr5a9/mkcIHPT
aXLd7Y85Iko9GDXrW4Xw58ZAqQO6nLSTvZ3LX5zxpmIY8RKw31bTnZbJWnzlV5UU
AmXWUWcptcu8WYcWfgrzP8Kob5znC1svHsKm0g==
-----END CERTIFICATE-----
EOT
    private_key = <<EOT
-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBALcrRkYvf/pLJQQp21KCPk56AuWh0UxbMd75NOZQWrY1QFXTRJ0x
YHAGa/LGLgposjKO9BELu+AYe+UoIJVwdAkCAwEAAQJBAJXeLHOErdum3DSZ4r+R
nVUsc25bhhpJi3Z6xJOlL3NgoDaAWEapQZ+jGs/XPCu14Skxwy5s9wgXznsfxIav
qWECIQDeVgWmBcvNz2FmQD8V1pIfQoec3hpTH3bVA06Rhg0j7QIhANLnGVpiCI+s
Pgqeqr93J1HojrcD9u5C9kahdt57GgUNAiBI5E7pxVCx4uF90mZcVIKHeRpY1YAv
7ErbP0BM+XPpaQIgNaUu37yb7N+lEFJ3oCgQylbbJlZN0yEZP7IGaGTro2kCIQCc
qYYLFv6yuySapSHrdOaPXnXrhMY4BE0EpzAuh+opxw==
-----END RSA PRIVATE KEY-----
EOT
  }
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3InvalidSubnet(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id               = "abc"
  container_network_type  = "overlay_l2"
  kubernetes_svc_ip_range = "10.247.0.0/16"
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3InvalidVPC(clusterName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_vpc_v1" "vpc" {
  cidr = "192.168.0.0/16"
  name = "cce-test"
}

locals {
  subnet_cidr  = cidrsubnet(opentelekomcloud_vpc_v1.vpc.cidr, 8, 0)
  subnet_gw_ip = cidrhost(local.subnet_cidr, 1)
}

resource "opentelekomcloud_vpc_subnet_v1" "subnet" {
  cidr       = local.subnet_cidr
  gateway_ip = local.subnet_gw_ip
  name       = "cce-test"
  vpc_id     = opentelekomcloud_vpc_v1.vpc.id
}

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = "abc"
  subnet_id               = opentelekomcloud_vpc_subnet_v1.subnet.id
  container_network_type  = "overlay_l2"
  kubernetes_svc_ip_range = "10.247.0.0/16"
}
`, clusterName)
}

func testAccCCEClusterV3Computed(clusterName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_vpc_v1" "vpc" {
  cidr = "192.168.0.0/16"
  name = "cce-test"
}

locals {
  subnet_cidr  = cidrsubnet(opentelekomcloud_vpc_v1.vpc.cidr, 8, 0)
  subnet_gw_ip = cidrhost(local.subnet_cidr, 1)
}

resource "opentelekomcloud_vpc_subnet_v1" "subnet" {
  cidr       = local.subnet_cidr
  gateway_ip = local.subnet_gw_ip
  name       = "cce-test"
  vpc_id     = opentelekomcloud_vpc_v1.vpc.id
}

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = opentelekomcloud_vpc_v1.vpc.id
  subnet_id               = opentelekomcloud_vpc_subnet_v1.subnet.id
  container_network_type  = "overlay_l2"
  kubernetes_svc_ip_range = "10.247.0.0/16"
}
`, clusterName)
}

func testAccCCEClusterV3NoAddons(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id               = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type  = "overlay_l2"
  kubernetes_svc_ip_range = "10.247.0.0/16"
  no_addons               = true
}
`, common.DataSourceSubnet, clusterName)
}

func testAccCCEClusterV3NoUserClusterData(clusterName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "%s"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id               = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type  = "overlay_l2"
  kubernetes_svc_ip_range = "10.247.0.0/16"

  ignore_certificate_clusters_data = true
  ignore_certificate_users_data    = true
}
`, common.DataSourceSubnet, clusterName)
}

func randClusterName() string {
	return fmt.Sprintf("cce-%s", acctest.RandString(5))
}
