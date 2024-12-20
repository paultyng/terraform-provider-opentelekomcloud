package acceptance

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
)

const bucketBasePath = "acc_css/css-snapshot"

func TestResourceCSSSnapshotConfigurationV1_basic(t *testing.T) {
	name := fmt.Sprintf("css-%s", acctest.RandString(10))
	resourceName := "opentelekomcloud_css_snapshot_configuration_v1.config"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			quotas.BookMany(t, sharedFlavorQuotas(t, 1, 100))
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceCSSSnapshotConfigurationV1Basic(name, bucketBasePath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "creation_policy.0.prefix", "snap"),
					resource.TestCheckResourceAttr(resourceName, "creation_policy.0.period", "16:00 GMT+01:00"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.base_path", bucketBasePath),
				),
			},
			{
				Config: testResourceCSSSnapshotConfigurationV1Updated(name, bucketBasePath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "creation_policy.0.prefix", "snapshot"),
					resource.TestCheckResourceAttr(resourceName, "creation_policy.0.period", "17:00 GMT+01:00"),
					resource.TestCheckResourceAttr(resourceName, "creation_policy.0.keepday", "2"),
				),
			},
		},
	})
}

func TestAccCheckCSSV1Validation(t *testing.T) {
	name := fmt.Sprintf("css-%s", acctest.RandString(10))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testResourceCSSSnapshotConfigurationV1Validation(name, bucketBasePath),
				ExpectError: regexp.MustCompile(`Conflicting configuration.+`),
			},
		},
	})
}

func getOsAgency() string {
	agency := os.Getenv("OS_CSS_OBS_AGENCY")
	if agency == "" {
		agency = "css_obs_agency"
	}
	return agency
}

var osAgency = getOsAgency()

func testResourceCSSSnapshotConfigurationV1Basic(name, bucketBasePath string) string {
	relatedConfig := testAccCssClusterV1Basic(name)
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "tf-snap-testing"
  force_destroy = true
}

resource "opentelekomcloud_css_snapshot_configuration_v1" "config" {
  cluster_id = opentelekomcloud_css_cluster_v1.cluster.id
  configuration {
    bucket    = opentelekomcloud_obs_bucket.bucket.bucket
    agency    = "%s"
    base_path = "%s"
  }
  creation_policy {
    prefix      = "snap"
    period      = "16:00 GMT+01:00"
    keepday     = 1
    enable      = true
    delete_auto = true
  }
}
`, relatedConfig, osAgency, bucketBasePath)
}

func testResourceCSSSnapshotConfigurationV1Updated(name, bucketBasePath string) string {
	relatedConfig := testAccCssClusterV1Basic(name)
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "tf-snap-testing"
  force_destroy = true
}

resource "opentelekomcloud_css_snapshot_configuration_v1" "config" {
  cluster_id = opentelekomcloud_css_cluster_v1.cluster.id
  configuration {
    bucket    = opentelekomcloud_obs_bucket.bucket.bucket
    agency    = "%s"
    base_path = "%s"
  }
  creation_policy {
    prefix      = "snapshot"
    period      = "17:00 GMT+01:00"
    keepday     = 2
    enable      = true
    delete_auto = true
  }
}
`, relatedConfig, osAgency, bucketBasePath)
}

func testResourceCSSSnapshotConfigurationV1Validation(name, bucketBasePath string) string {
	relatedConfig := testAccCssClusterV1Basic(name)
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "tf-snap-testing"
  force_destroy = true
}

resource "opentelekomcloud_css_snapshot_configuration_v1" "config" {
  cluster_id = opentelekomcloud_css_cluster_v1.cluster.id
  automatic  = true
  configuration {
    bucket    = opentelekomcloud_obs_bucket.bucket.bucket
    agency    = "%s"
    base_path = "%s"
  }
}
`, relatedConfig, osAgency, bucketBasePath)
}
