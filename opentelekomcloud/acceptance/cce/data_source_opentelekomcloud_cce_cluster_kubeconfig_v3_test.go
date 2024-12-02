package acceptance

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
)

const dataSourceName = "data.opentelekomcloud_cce_cluster_kubeconfig_v3.this"

func TestAccCCEKubeConfigV3DataSource_basic(t *testing.T) {
	var cceName = fmt.Sprintf("cce-test-%s", acctest.RandString(5))

	t.Parallel()
	quotas.BookOne(t, quotas.CCEClusterQuota)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCCEClusterKubeconfigV3DataSourceInit(cceName),
			},
			{
				Config: testAccCCEClusterKubeconfigV3DataSourceBasic(cceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCCEClusterKubeconfigV3DataSourceID(dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "kubeconfig"),
				),
			},
			{
				Config: testAccCCEClusterKubeconfigV3DataSourceExpiryDate(cceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCCEClusterKubeconfigV3DataSourceID(dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "kubeconfig"),
				),
			},
		},
	})
}

func testAccCheckCCEClusterKubeconfigV3DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find kubeconfig data source: %s ", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("kubeconfig data source ID not set ")
		}

		return nil
	}
}

func testAccCCEClusterKubeconfigV3DataSourceInit(cceName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                   = "%s"
  cluster_type           = "VirtualMachine"
  flavor_id              = "cce.s1.small"
  vpc_id                 = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id              = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type = "overlay_l2"
}
`, common.DataSourceSubnet, cceName)
}

func testAccCCEClusterKubeconfigV3DataSourceBasic(cceName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                   = "%s"
  cluster_type           = "VirtualMachine"
  flavor_id              = "cce.s1.small"
  vpc_id                 = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id              = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type = "overlay_l2"
}

data "opentelekomcloud_cce_cluster_kubeconfig_v3" "this" {
  cluster_id = opentelekomcloud_cce_cluster_v3.cluster_1.id
}
`, common.DataSourceSubnet, cceName)
}

func testAccCCEClusterKubeconfigV3DataSourceExpiryDate(cceName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                   = "%s"
  cluster_type           = "VirtualMachine"
  flavor_id              = "cce.s1.small"
  vpc_id                 = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id              = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type = "overlay_l2"
}

data "opentelekomcloud_cce_cluster_kubeconfig_v3" "this" {
  cluster_id  = opentelekomcloud_cce_cluster_v3.cluster_1.id
  expiry_date = "%s"
}
`, common.DataSourceSubnet, cceName, GetDatePlusTwo())
}

// GetDatePlusTwo returns the date for today + 2 days in the YYYY-MM-DD format.
func GetDatePlusTwo() string {
	today := time.Now()
	futureDate := today.AddDate(0, 0, 2)
	formattedDate := futureDate.Format("2006-01-02")
	return formattedDate
}
