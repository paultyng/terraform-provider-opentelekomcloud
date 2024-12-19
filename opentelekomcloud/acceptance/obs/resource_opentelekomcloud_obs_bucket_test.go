package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func TestAccObsBucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", testAccObsBucketName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "bucket_domain_name", testAccObsBucketDomainName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "region", env.OS_REGION_NAME),
				),
			},
			{
				Config: testAccObsBucketUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "WARM"),
				),
			},
			{
				Config: testAccObsBucketSSE(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.kms_key_id", env.OS_KMS_ID),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "kms"),
				),
			},
			{
				Config: testAccObsBucketUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
				),
			},
		},
	})
}

func TestAccObsBucket_tags(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.name", testAccObsBucketName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccObsBucket_versioning(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning", "true"),
				),
			},
			{
				Config: testAccObsBucketConfigWithDisableVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning", "false"),
				),
			},
		},
	})
}

func TestAccObsBucket_logging(t *testing.T) {
	rInt := acctest.RandInt()
	targetBucket := fmt.Sprintf("tf-test-log-bucket-%d", rInt)
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithLogging(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					testAccCheckObsBucketLogging(resourceName, targetBucket, "log/"),
				),
			},
		},
	})
}

func TestAccObsBucket_lifecycle(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.name", "rule1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.name", "rule2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.name", "rule3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.transition.0.days", "30"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.transition.1.days", "180"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.noncurrent_version_transition.0.days", "60"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.noncurrent_version_transition.1.days", "180"),
				),
			},
		},
	})
}

func TestAccObsBucket_website(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketWebsiteConfigWithRoutingRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", "error.html"),
				),
			},
		},
	})
}

func TestAccObsBucket_cors(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.1", "ETag"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
		},
	})
}

func TestAccObsBucket_notifications(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithNotifications(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_notifications.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_notifications.0.events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_notifications.0.filter_rule.#", "2"),
				),
			},
		},
	})
}

func TestAccObsBucket_pfs(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketPfs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parallel_fs", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_version", "3.0"),
				),
			},
		},
	})
}

func TestAccObsBucket_WormPolicy(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketWormPolicyBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "worm_policy.0.days", "15"),
				),
			},
			{
				Config: testAccObsBucketWormPolicyUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "worm_policy.0.years", "1"),
				),
			},
		},
	})
}

func TestAccObsBucket_UserDomainNames(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "opentelekomcloud_obs_bucket.bucket"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketUserDomainNamesBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_domain_names.0", "www.test-check.com"),
				),
			},
			{
				Config: testAccObsBucketUserDomainNamesUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_domain_names.0", "www.test.com"),
				),
			},
		},
	})
}

func TestAccOBSBucket_VersioningObjectLockValidation(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccObsBucketWormPolicyValidation(rInt),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`OBS requires versioning to be enabled.+`),
			},
		},
	})
}

func testAccCheckObsBucketDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.NewObjectStorageClient(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud OBS client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_obs_bucket" {
			continue
		}

		_, err := client.HeadBucket(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("bucket %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckObsBucketExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.NewObjectStorageClient(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud OBS client: %s", err)
		}

		_, err = client.HeadBucket(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("bucket not found: %v", err)
		}
		return nil
	}
}

func testAccCheckObsBucketLogging(name, target, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.NewObjectStorageClient(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud OBS client: %s", err)
		}

		output, err := client.GetBucketLoggingConfiguration(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting logging configuration of OBS bucket: %s", err)
		}

		if output.TargetBucket != target {
			return fmt.Errorf("%s.logging: Attribute 'target_bucket' expected %s, got %s",
				name, output.TargetBucket, target)
		}
		if output.TargetPrefix != prefix {
			return fmt.Errorf("%s.logging: Attribute 'target_prefix' expected %s, got %s",
				name, output.TargetPrefix, prefix)
		}

		return nil
	}
}

// These need a bit of randomness as the name can only be used once globally
func testAccObsBucketName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d", randInt)
}

func testAccObsBucketDomainName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d.obs.%s.otc.t-systems.com", randInt, env.OS_REGION_NAME)
}

func testAccObsBucketBasic(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  storage_class = "STANDARD"
  acl           = "private"
}
`, randInt)
}

func testAccObsBucketUpdate(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  storage_class = "WARM"
  acl           = "public-read"
  versioning    = true
}
`, randInt)
}

func testAccObsBucketSSE(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  storage_class = "WARM"
  acl           = "public-read"
  force_destroy = true
  server_side_encryption {
    algorithm  = "kms"
    kms_key_id = "%s"
  }
}
`, randInt, env.OS_KMS_ID)
}

func testAccObsBucketConfigWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"

  tags = {
    name = "tf-test-bucket-%d"
    foo  = "bar"
    key1 = "value1"
  }
}
`, randInt, randInt)
}

func testAccObsBucketConfigWithVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket     = "tf-test-bucket-%d"
  acl        = "private"
  versioning = true
}
`, randInt)
}

func testAccObsBucketConfigWithDisableVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket     = "tf-test-bucket-%d"
  acl        = "private"
  versioning = false
}
`, randInt)
}

func testAccObsBucketConfigWithLogging(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "log_bucket" {
  bucket        = "tf-test-log-bucket-%d"
  acl           = "log-delivery-write"
  force_destroy = "true"
}
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"

  logging {
    target_bucket = opentelekomcloud_obs_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
`, randInt, randInt)
}

func testAccObsBucketConfigWithLifecycle(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket     = "tf-test-bucket-%d"
  acl        = "private"
  versioning = true

  lifecycle_rule {
    name    = "rule1"
    prefix  = "path1/"
    enabled = true

    expiration {
      days = 365
    }
  }
  lifecycle_rule {
    name    = "rule2"
    prefix  = "path2/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days          = 30
      storage_class = "WARM"
    }
    transition {
      days          = 180
      storage_class = "COLD"
    }
  }
  lifecycle_rule {
    name    = "rule3"
    prefix  = "path3/"
    enabled = true

    noncurrent_version_expiration {
      days = 365
    }

    noncurrent_version_transition {
      days          = 60
      storage_class = "WARM"
    }
    noncurrent_version_transition {
      days          = 180
      storage_class = "COLD"
    }
  }
}
`, randInt)
}

func testAccObsBucketWebsiteConfigWithRoutingRules(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
    routing_rules  = <<EOF
[{
	"Condition": {
		"KeyPrefixEquals": "docs/"
	},
	"Redirect": {
		"ReplaceKeyPrefixWith": "documents/"
	}
}]
EOF
  }
}
`, randInt)
}

func testAccObsBucketConfigWithCORS(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, randInt)
}

func testAccObsBucketConfigWithNotifications(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_smn_topic_v2" "topic" {
  name         = "obs-notifications"
  display_name = "The display name of topic_1"
}

resource "opentelekomcloud_smn_topic_attribute_v2" "policy" {
  topic_urn       = opentelekomcloud_smn_topic_v2.topic.id
  attribute_name  = "access_policy"
  topic_attribute = <<EOF
{
  "Version": "2016-09-07",
  "Id": "__default_policy_ID",
  "Statement": [
    {
      "Sid": "__service_pub_0",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "obs",
          "s3"
        ]
      },
      "Action": [
        "SMN:Publish",
        "SMN:QueryTopicDetail"
      ],
      "Resource": "${opentelekomcloud_smn_topic_v2.topic.id}"
    }
  ]
}
EOF

  depends_on = [opentelekomcloud_smn_topic_v2.topic]
}

resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  event_notifications {
    topic = opentelekomcloud_smn_topic_v2.topic.id
    events = [
      "ObjectCreated:*",
      "ObjectRemoved:*",
    ]
    filter_rule {
      name  = "prefix"
      value = "smn"
    }
    filter_rule {
      name  = "suffix"
      value = ".jpg"
    }
  }

  depends_on = [opentelekomcloud_smn_topic_attribute_v2.policy]
}
`, randInt)
}

func testAccObsBucketPfs(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket      = "tf-test-bucket-%d"
  parallel_fs = true
}
`, randInt)
}

func testAccObsBucketWormPolicyBasic(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket     = "tf-test-bucket-%d"
  versioning = true
  worm_policy {
    days = 15
  }
}
`, randInt)
}

func testAccObsBucketWormPolicyUpdate(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket     = "tf-test-bucket-%d"
  versioning = true
  worm_policy {
    years = 1
  }
}
`, randInt)
}

func testAccObsBucketWormPolicyValidation(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  worm_policy {
    years = 1
  }
}
`, randInt)
}

func testAccObsBucketUserDomainNamesBasic(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  user_domain_names = [
    "www.test-check.com",
    "www.test.com",
    "www.tf.com"
  ]
}
`, randInt)
}

func testAccObsBucketUserDomainNamesUpdate(randInt int) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  user_domain_names = [
    "www.test.com"
  ]
}
`, randInt)
}
