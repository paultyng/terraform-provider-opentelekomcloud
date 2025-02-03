package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cts/v3/keyevent"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/cts"
)

const notificationResource = "opentelekomcloud_cts_event_notification_v3.notification_v3"

func TestAccCTSEventNotificationV3_basic(t *testing.T) {
	var ctsEventResponse keyevent.NotificationResponse
	var topicName = fmt.Sprintf("terra-user-%s", acctest.RandString(5))
	var notificationName = fmt.Sprintf("terra_notification_%s", acctest.RandString(5))
	var notificationNameUpdated = fmt.Sprintf("terra_notification_updated_%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCTSEventNotificationV3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCTSEventNotificationV3Basic(topicName, notificationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCTSEventNotificationV3Exists(notificationResource, &ctsEventResponse, env.OS_TENANT_NAME),
					resource.TestCheckResourceAttr(notificationResource, "filter.0.condition", "AND"),
					resource.TestCheckResourceAttr(notificationResource, "filter.0.rule.#", "2"),
					resource.TestCheckResourceAttr(notificationResource, "notification_name", notificationName),
					resource.TestCheckResourceAttr(notificationResource, "status", "disabled"),
				),
			},
			{
				Config: testAccCTSEventNotificationV3Update(topicName, notificationNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCTSEventNotificationV3Exists(notificationResource, &ctsEventResponse, env.OS_TENANT_NAME),
					resource.TestCheckResourceAttr(notificationResource, "notification_name", notificationNameUpdated),
					resource.TestCheckResourceAttr(notificationResource, "status", "enabled"),
					resource.TestCheckResourceAttr(notificationResource, "filter.0.condition", "OR"),
					resource.TestCheckResourceAttr(notificationResource, "filter.0.rule.#", "3"),
					resource.TestCheckResourceAttr(notificationResource, "operations.0.resource_type", "evs"),
					resource.TestCheckResourceAttr(notificationResource, "operations.0.service_type", "EVS"),
					resource.TestCheckResourceAttr(notificationResource, "operations.0.trace_names.0", "createVolume"),
					resource.TestCheckResourceAttr(notificationResource, "operations.0.trace_names.1", "deleteVolume"),
				),
			},
		},
	})
}

func TestAccCTSEventNotificationV3_users(t *testing.T) {
	var ctsEventResponse keyevent.NotificationResponse
	var userOne = fmt.Sprintf("terra-user-%s", acctest.RandString(5))
	var userTwo = fmt.Sprintf("terra-user-%s", acctest.RandString(5))
	var groupName = fmt.Sprintf("terra-group-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCTSEventNotificationV3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCTSEventNotificationV3Users(groupName, userOne, userTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCTSEventNotificationV3Exists(notificationResource, &ctsEventResponse, env.OS_TENANT_NAME),
					resource.TestCheckResourceAttr(notificationResource, "notification_name", "test_user"),
					resource.TestCheckResourceAttr(notificationResource, "status", "disabled"),
					resource.TestCheckResourceAttr(notificationResource, "notification_name", "test_user"),
					resource.TestCheckResourceAttr(notificationResource, "notify_user_list.0.user_group", groupName),
					resource.TestCheckResourceAttr(notificationResource, "notify_user_list.0.user_list.0", userOne),
					resource.TestCheckResourceAttr(notificationResource, "notify_user_list.0.user_list.1", userTwo),
				),
			},
			{
				Config: testAccCTSEventNotificationV3UsersUpdate(groupName, userOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCTSEventNotificationV3Exists(notificationResource, &ctsEventResponse, env.OS_TENANT_NAME),
					resource.TestCheckResourceAttr(notificationResource, "notification_name", "test_user"),
					resource.TestCheckResourceAttr(notificationResource, "status", "enabled"),
					resource.TestCheckResourceAttr(notificationResource, "notify_user_list.0.user_group", groupName),
					resource.TestCheckResourceAttr(notificationResource, "notify_user_list.0.user_list.0", userOne),
				),
			},
		},
	})
}

func TestAccCTSEventNotificationV3_importBasic(t *testing.T) {
	var topicName = fmt.Sprintf("terra-user-%s", acctest.RandString(5))
	var notificationName = fmt.Sprintf("terra_notification_%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckCTSEventNotificationV3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCTSEventNotificationV3Basic(topicName, notificationName),
			},

			{
				ResourceName:      notificationResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccCheckCTSV3Validation(t *testing.T) {
	var topicName = fmt.Sprintf("terra-user-%s", acctest.RandString(5))
	var notificationName = fmt.Sprintf("terra_notification_%s", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCTSEventNotificationV3ValidationOperations(topicName, notificationName),
				ExpectError: regexp.MustCompile(`customized operations can't be used.+`),
			},
		},
	})
}

func testAccCheckCTSEventNotificationV3Destroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	ctsClient, err := config.CtsV3Client(env.OS_TENANT_NAME)
	if err != nil {
		return fmt.Errorf("error creating cts client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_cts_event_notification_v3" {
			continue
		}

		_, eventName := cts.ExtractNotificationID(rs.Primary.ID)
		ctsN, err := keyevent.List(ctsClient, keyevent.ListNotificationsOpts{
			NotificationType: "smn",
			NotificationName: eventName,
		})
		if err == nil {
			return fmt.Errorf("error retrieving cts event notification: %w", err)
		}

		if ctsN != nil {
			return fmt.Errorf("failed to delete CTS event notification")
		}
	}

	return nil
}

func testAccCheckCTSEventNotificationV3Exists(n string, notification *keyevent.NotificationResponse, projectName cfg.ProjectName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.CtsV3Client(projectName)
		if err != nil {
			return fmt.Errorf("error creating cts client: %s", err)
		}
		_, eventName := cts.ExtractNotificationID(rs.Primary.ID)

		ctsN, err := keyevent.List(client, keyevent.ListNotificationsOpts{
			NotificationType: "smn",
			NotificationName: eventName,
		})
		if err != nil {
			return fmt.Errorf("error retrieving cts event notification: %w", err)
		}

		id := fmt.Sprintf("%s/%s", ctsN[0].NotificationId, ctsN[0].NotificationName)

		if id != rs.Primary.ID {
			return fmt.Errorf("CTS event notification not found")
		}

		notification = &ctsN[0]

		return nil
	}
}

func testAccCTSEventNotificationV3Basic(topic, notification string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "%s"
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "%s"
  operation_type    = "complete"
  topic_id          = opentelekomcloud_smn_topic_v2.topic_1.id
  status            = "disabled"

  filter {
    condition = "AND"
    rule      = ["code = 200", "resource_name = test"]
  }
}
`, topic, notification)
}

func testAccCTSEventNotificationV3Update(topic, notification string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "%s"
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "%s"
  operation_type    = "customized"
  topic_id          = opentelekomcloud_smn_topic_v2.topic_1.id
  status            = "enabled"
  operations {
    resource_type = "evs"
    service_type  = "EVS"
    trace_names = ["createVolume",
    "deleteVolume"]
  }
  operations {
    resource_type = "vpc"
    service_type  = "VPC"
    trace_names = ["deleteVpc",
    "createVpc"]
  }

  filter {
    condition = "OR"
    rule      = ["code = 400", "resource_name = name", "api_version = 1.0"]
  }
}
`, topic, notification)
}

func testAccCTSEventNotificationV3Users(group, userOne, userTwo string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "topic_1"
}

resource "opentelekomcloud_identity_group_v3" "group_1" {
  name = "%s"
}

resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  password = "password123@#"
  enabled  = true
}

resource "opentelekomcloud_identity_user_v3" "user_2" {
  name     = "%s"
  password = "password123@#"
  enabled  = true
}

resource "opentelekomcloud_identity_group_membership_v3" "membership_1" {
  group = opentelekomcloud_identity_group_v3.group_1.id
  users = [opentelekomcloud_identity_user_v3.user_1.id,
  opentelekomcloud_identity_user_v3.user_2.id]
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "test_user"
  operation_type    = "customized"
  topic_id          = opentelekomcloud_smn_topic_v2.topic_1.id
  status            = "disabled"
  operations {
    resource_type = "vpc"
    service_type  = "VPC"
    trace_names = ["deleteVpc",
    "createVpc"]
  }
  notify_user_list {
    user_group = "%s"
    user_list  = ["%s", "%s"]
  }
}
`, group, userOne, userTwo, group, userOne, userTwo)
}

func testAccCTSEventNotificationV3UsersUpdate(group, userOne string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "topic_1"
}

resource "opentelekomcloud_identity_group_v3" "group_1" {
  name = "%s"
}

resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "%s"
  password = "password123@#"
  enabled  = true
}

resource "opentelekomcloud_identity_group_membership_v3" "membership_1" {
  group = opentelekomcloud_identity_group_v3.group_1.id
  users = [opentelekomcloud_identity_user_v3.user_1.id]
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "test_user"
  operation_type    = "customized"
  topic_id          = opentelekomcloud_smn_topic_v2.topic_1.id
  status            = "enabled"
  operations {
    resource_type = "vpc"
    service_type  = "VPC"
    trace_names = ["deleteVpc",
    "createVpc"]
  }
  notify_user_list {
    user_group = "%s"
    user_list  = ["%s"]
  }
}
`, group, userOne, group, userOne)
}

func testAccCTSEventNotificationV3ValidationOperations(topic, notification string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "%s"
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "%s"
  operation_type    = "complete"
  topic_id          = opentelekomcloud_smn_topic_v2.topic_1.id
  status            = "enabled"
  operations {
    resource_type = "evs"
    service_type  = "EVS"
    trace_names = ["createVolume",
    "deleteVolume"]
  }
  operations {
    resource_type = "vpc"
    service_type  = "VPC"
    trace_names = ["deleteVpc",
    "createVpc"]
  }
}
`, topic, notification)
}
