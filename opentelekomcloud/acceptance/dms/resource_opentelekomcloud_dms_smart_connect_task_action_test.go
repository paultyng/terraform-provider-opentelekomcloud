package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
)

func TestAccDmsKafkav2SmartConnectTaskActionV2_basic(t *testing.T) {
	var obj interface{}
	rName := fmt.Sprintf("dms-acc-api%s", acctest.RandString(5))
	resourceName := "opentelekomcloud_dms_smart_connect_task_v2.test"

	rc := common.InitResourceCheck(
		resourceName,
		&obj,
		getDmsKafkav2SmartConnectTaskResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			common.TestAccPreCheckOBS(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDmsKafkav2SmartConnectTaskAction_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func testDmsKafkav2SmartConnectTaskAction_basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_obs_bucket" "test" {
  bucket        = "%[2]s"
  storage_class = "STANDARD"
  acl           = "private"
}

resource "opentelekomcloud_dms_smart_connect_v2" "test" {
  instance_id       = opentelekomcloud_dms_dedicated_instance_v2.test.id
  storage_spec_code = "dms.physical.storage.high.v2"
  node_count        = 2
  bandwidth         = "100MB"
}

resource "opentelekomcloud_dms_topic_v1" "test" {
  instance_id    = opentelekomcloud_dms_dedicated_instance_v2.test.id
  name           = "%[2]s"
  partition      = 10
  retention_time = 36
}

resource "opentelekomcloud_dms_smart_connect_task_v2" "test" {
  depends_on = [opentelekomcloud_dms_smart_connect_v2.test, opentelekomcloud_dms_topic_v1.test]

  instance_id      = opentelekomcloud_dms_dedicated_instance_v2.test.id
  task_name        = "%[2]s"
  destination_type = "OBS_SINK"
  topics           = [opentelekomcloud_dms_topic_v1.test.name]

  destination_task {
    consumer_strategy     = "latest"
    destination_file_type = "TEXT"
    access_key            = "%[3]s"
    secret_key            = "%[4]s"
    obs_bucket_name       = opentelekomcloud_obs_bucket.test.bucket
    partition_format      = "yyyy/MM/dd/HH/mm"
    record_delimiter      = ";"
    deliver_time_interval = 300
  }
}

resource "opentelekomcloud_dms_smart_connect_task_action_v2" "test" {
  depends_on = [opentelekomcloud_dms_smart_connect_task_v2.test]

  instance_id = opentelekomcloud_dms_dedicated_instance_v2.test.id
  task_id     = opentelekomcloud_dms_smart_connect_task_v2.test.id
  action      = "pause"
}
`, testAccKafkaInstance_newFormat(rName), rName, env.OS_ACCESS_KEY, env.OS_SECRET_KEY)
}
