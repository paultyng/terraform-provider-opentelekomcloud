---
subcategory: "Cloud Trace Service (CTS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_cts_event_notification_v3"
sidebar_current: "docs-opentelekomcloud-resource-cts-event-notification-v3"
description: |-
  Manages a CTS Event Notification resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for CTS event notification you can get at
[documentation portal](https://docs.otc.t-systems.com/cloud-trace-service/api-ref/api_description/key_event_notification_management)

# opentelekomcloud_cts_event_notification_v3

Allows to send SMS, email, or HTTP/HTTPS notifications through pre-configured SMN topics to subscribers.

## Example Usage

### Event notification which delivers every tenant action to subscribers

```hcl
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "topic_1"
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "my_notification"
  operation_type    = "complete"
  topic_id          = opentelekomcloud_smn_topic_v2.topic_1.id
  status            = "enabled"
}
```

### Event notification with disabled SMN topic and filtering

```hcl
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "topic_1"
}

resource "opentelekomcloud_cts_event_notification_v3" "notification_v3" {
  notification_name = "my_notification"
  operation_type    = "complete"

  filter {
    condition = "AND"
    rule      = ["code = 200", "resource_name = test"]
  }
}
```

### Event notification with selected operations and users

```hcl
resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name = "topic_1"
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
  operations {
    resource_type = "evs"
    service_type  = "EVS"
    trace_names = ["createVolume",
    "deleteVolume"]
  }
  notify_user_list {
    user_group = "user_group"
    user_list  = ["user_one", "user_two"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `notification_name` - (Required) The name of event notification rule. Only letters, digits
  and underscores (_) are allowed.

* `operation_type` - (Required) The operation type of event rules.

  Possible values:
  * `complete` - Any operation will trigger notification.

  * `customized` - Only selected operations will trigger notification.

* `topic_id` - (Optional) Specifies SMN topic URN that will be used for events notification.

* `status` - (Optional) Specifies whether SMN topic is `enabled` or `disabled`.

* `operations` - (Optional) Specifies which operations are enabled in event notification rule. Can be only specified
  when `operation_type` is set to `customized`. Supported fields:

    * `service_type` - (Required) Specifies the cloud service. Every service should be provided separately, the value
    must be the acronym of a cloud service that has been connected with CTS.

    * `resource_type` - (Required) Specifies the resource type of custom operation.

    * `trace_names` - (Required) Specifies the list with trace names of custom operation.

* `notify_user_list` - (Optional) Specifies the list of users whose operations will trigger notifications.
   Currently, up to 50 users in 10 user groups can be configured. Supported fields:

  * `user_group` - (Required) Specifies the IAM user group.

  * `user_list` - (Required) Specifies the list with IAM users which belong to `user_group`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `notification_id` - Unique event notification id.

* `notification_type` - Specifies the notification type. Current `cts` version supports only `smn` type.

* `project_id` - Specifies project id of event notification rule.

* `create_time` - Specifies creation time of event notification rule.

## Import

CTS event notification can be imported using the `notification_id/notification_name`, e.g.

```shell
$ terraform import opentelekomcloud_cts_event_notification_v3.notification c1881895-cdcb-4d23-96cb-032e6a3ee667/test_event
```
