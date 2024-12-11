package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/obs"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getOBSBucketObjectAclResourceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.NewObjectStorageClient(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OBS client: %s", err)
	}
	params := &obs.GetObjectAclInput{
		Bucket: state.Primary.Attributes["bucket"],
		Key:    state.Primary.ID,
	}
	return client.GetObjectAcl(params)
}

func TestAccOBSBucketObjectAcl_basic(t *testing.T) {
	var obj interface{}

	bucketName := fmt.Sprintf("bucket-%s", acctest.RandString(3))
	rName := "opentelekomcloud_obs_bucket_object_acl.test"

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getOBSBucketObjectAclResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testOBSBucketObjectAcl_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "bucket", bucketName),
					resource.TestCheckResourceAttr(rName, "key", "test-key"),
					resource.TestCheckResourceAttr(rName, "public_permission.0.access_to_acl.0", "READ_ACP"),
					resource.TestCheckResourceAttr(rName, "public_permission.0.access_to_acl.1", "WRITE_ACP"),
					resource.TestCheckResourceAttr(rName, "account_permission.#", "2"),
					resource.TestCheckResourceAttr(rName, "public_permission.#", "1"),
					resource.TestCheckResourceAttr(rName, "owner_permission.#", "1"),
				),
			},
			{
				Config: testOBSBucketObjectAcl_basicUpdate(bucketName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "bucket", bucketName),
					resource.TestCheckResourceAttr(rName, "key", "test-key"),
					resource.TestCheckResourceAttr(rName, "account_permission.0.access_to_acl.0", "READ_ACP"),
					resource.TestCheckResourceAttr(rName, "account_permission.0.account_id", "1000010022"),
					resource.TestCheckResourceAttr(rName, "public_permission.0.access_to_acl.0", "WRITE_ACP"),
					resource.TestCheckResourceAttr(rName, "account_permission.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_permission.#", "1"),
					resource.TestCheckResourceAttr(rName, "owner_permission.#", "1"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testObsBucketObjectAclImportState(rName),
			},
		},
	})
}

func testOBSBucketObjectAcl_base(bucketName string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "%s"
  storage_class = "STANDARD"
  acl           = "private"
}

resource "opentelekomcloud_obs_bucket_object" "object" {
  bucket  = opentelekomcloud_obs_bucket.bucket.bucket
  key     = "test-key"
  content = "some_bucket_content"
}

`, bucketName)
}

func testOBSBucketObjectAcl_basic(bucketName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_obs_bucket_object_acl" "test" {
  bucket = opentelekomcloud_obs_bucket.bucket.bucket
  key    = opentelekomcloud_obs_bucket_object.object.key

  account_permission {
    access_to_object = ["READ"]
    access_to_acl    = ["READ_ACP", "WRITE_ACP"]
    account_id       = "1000010020"
  }

  account_permission {
    access_to_object = ["READ"]
    access_to_acl    = ["READ_ACP"]
    account_id       = "1000010021"
  }

  public_permission {
    access_to_acl = ["READ_ACP", "WRITE_ACP"]
  }
}
`, testOBSBucketObjectAcl_base(bucketName))
}

func testOBSBucketObjectAcl_basicUpdate(bucketName string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_obs_bucket_object_acl" "test" {
  bucket = opentelekomcloud_obs_bucket.bucket.bucket
  key    = opentelekomcloud_obs_bucket_object.object.key

  account_permission {
    access_to_acl = ["READ_ACP"]
    account_id    = "1000010022"
  }

  public_permission {
    access_to_acl = ["WRITE_ACP"]
  }
}
`, testOBSBucketObjectAcl_base(bucketName))
}

func testObsBucketObjectAclImportState(name string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("resource (%s) not found: %s", name, rs)
		}

		bucket := rs.Primary.Attributes["bucket"]
		if bucket == "" {
			return "", fmt.Errorf("attribute (bucket) of Resource (%s) not found: %s", name, rs)
		}
		return fmt.Sprintf("%s/%s", bucket, rs.Primary.ID), nil
	}
}
