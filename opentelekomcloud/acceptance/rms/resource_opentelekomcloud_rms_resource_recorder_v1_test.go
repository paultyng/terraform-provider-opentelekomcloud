package rms

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/recorder"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getRecorderResourceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.RmsV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating RMS V1 client: %s", err)
	}

	return recorder.GetRecorder(client, state.Primary.ID)
}

func TestAccRecorder_basic(t *testing.T) {
	var obj interface{}

	name := "rms-test-" + acctest.RandString(5)
	rName := "opentelekomcloud_rms_resource_recorder_v1.test"

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getRecorderResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testRecorder_with_obs_partial(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "agency_name", "rms_tracker_agency"),
					resource.TestCheckResourceAttr(rName, "selector.0.all_supported", "false"),
					resource.TestCheckResourceAttr(rName, "selector.0.resource_types.#", "4"),
					resource.TestCheckResourceAttrSet(rName, "obs_channel.0.region"),
					resource.TestCheckResourceAttrPair(rName, "obs_channel.0.bucket",
						"opentelekomcloud_obs_bucket.test", "bucket"),
				),
			},
			{
				Config: testRecorder_with_obs_all(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rName, "agency_name", "rms_tracker_agency"),
					resource.TestCheckResourceAttr(rName, "selector.0.all_supported", "true"),
					resource.TestCheckResourceAttr(rName, "selector.0.resource_types.#", "0"),
				),
			},
			{
				Config: testRecorder_with_smn(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "agency_name", "rms_tracker_agency"),
					resource.TestCheckResourceAttr(rName, "selector.0.all_supported", "true"),
					resource.TestCheckResourceAttrSet(rName, "smn_channel.0.region"),
					resource.TestCheckResourceAttrSet(rName, "smn_channel.0.project_id"),
					resource.TestCheckResourceAttrPair(rName, "smn_channel.0.topic_urn",
						"opentelekomcloud_smn_topic_v2.test", "topic_urn"),
				),
			},
			{
				Config: testRecorder_with_all(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(rName, "smn_channel.0.topic_urn",
						"opentelekomcloud_smn_topic_v2.test", "topic_urn"),
					resource.TestCheckResourceAttrPair(rName, "obs_channel.0.bucket",
						"opentelekomcloud_obs_bucket.test", "bucket"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testRecorder_base(name string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_obs_bucket" "test" {
  bucket        = "%[1]s"
  storage_class = "STANDARD"
  acl           = "private"
  force_destroy = true

  tags = {
    env = "rms_recorder_channel"
    key = "value"
  }
}

resource "opentelekomcloud_smn_topic_v2" "test" {
  name = "%[1]s"

  tags = {
    env = "rms_recorder_channel"
    key = "value"
  }
}
`, name)
}

func testRecorder_with_obs_partial(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_rms_resource_recorder_v1" "test" {
  agency_name = "rms_tracker_agency"

  selector {
    all_supported  = false
    resource_types = ["vpc.vpcs", "rds.instances", "dms.kafkas", "dms.queues"]
  }

  obs_channel {
    bucket = opentelekomcloud_obs_bucket.test.id
    region = "%s"
  }
}
`, testRecorder_base(name), env.OS_REGION_NAME)
}

func testRecorder_with_obs_all(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_rms_resource_recorder_v1" "test" {
  agency_name = "rms_tracker_agency"

  selector {
    all_supported = true
  }

  obs_channel {
    bucket = opentelekomcloud_obs_bucket.test.id
    region = "%s"
  }
}
`, testRecorder_base(name), env.OS_REGION_NAME)
}

func testRecorder_with_smn(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_rms_resource_recorder_v1" "test" {
  agency_name = "rms_tracker_agency"

  selector {
    all_supported = true
  }

  smn_channel {
    topic_urn = opentelekomcloud_smn_topic_v2.test.topic_urn
  }
}
`, testRecorder_base(name))
}

func testRecorder_with_all(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_rms_resource_recorder_v1" "test" {
  agency_name = "rms_tracker_agency"

  selector {
    all_supported = true
  }

  obs_channel {
    bucket = opentelekomcloud_obs_bucket.test.id
    region = "%s"
  }
  smn_channel {
    topic_urn = opentelekomcloud_smn_topic_v2.test.topic_urn
  }
}
`, testRecorder_base(name), env.OS_REGION_NAME)
}
