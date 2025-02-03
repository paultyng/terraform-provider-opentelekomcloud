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

* `notification_name` - (Required, String) The name of event notification rule. Only letters, digits
  and underscores (_) are allowed.

* `operation_type` - (Required, String) The operation type of event rules.

  Possible values:
  * `complete` - Any operation will trigger notification.

  * `customized` - Only selected operations will trigger notification.

* `topic_id` - (Optional, String) Specifies SMN topic URN that will be used for events notification.

* `status` - (Optional, String) Specifies whether SMN topic is `enabled` or `disabled`.

* `filter` - (Optional, List) Specifies the filtering rules for notification.
  The [filter](#CTS_Notification_Filter) structure is documented below.

* `operations` - (Optional, List) Specifies an array of operations that will trigger notifications.
  The [operations](#CTS_Notification_Operations) structure is documented below.

* `notify_user_list` - (Optional) Specifies the list of users whose operations will trigger notifications.
   Currently, up to 50 users in 10 user groups can be configured. Supported fields:

* `user_group` - (Required) Specifies the IAM user group.

* `user_list` - (Required) Specifies the list with IAM users which belong to `user_group`.

<a name="CTS_Notification_Filter"></a>
The `filter` block supports:

* `condition` - (Required, String) Specifies the relationship between multiple rules. The valid values are as follows:
    + **AND**: Effective after all filtering conditions are met.
    + **OR**: Effective when any one of the conditions is met.

* `rule` - (Required, List) Specifies an array of filtering rules. It consists of three parts,
  the first part is the **key**, the second part is the **rule**, and the third part is the **value**,
  the format is: **key != value**.
    + The **key** can be: **api_version**, **code**, **trace_rating**, **trace_type**, **resource_id** and
      **resource_name**.
      When the key is **api_version**, the value needs to follow the regular constraint: **^ (a-zA-Z0-9_ -.) {1,64}$**.
      When the key is **code**, the length range of value is from `1` to `256`.
      When the key is **trace_rating**, the value can be **normal**, **warning** or **incident**.
      When the key is **trace_type**, the value can be **ConsoleAction**, **ApiCall** or **SystemAction**.
      When the key is **resource_id**, the length range of value is from `1` to `350`.
      When the key is **resource_name**, the length range of value is from `1` to `256`.
    + The **rule** can be: **!=** or **=**.

<a name="CTS_Notification_Operations"></a>
The `operations` block supports:

* `service_type` - (Required, String) Specifies the cloud service. Every service should be provided separately, the value
  must be the acronym of a cloud service that has been connected with CTS.

* `resource_type` - (Required, String) Specifies the resource type of custom operation.

* `trace_names` - (Required, List) Specifies the list with trace names of custom operation.

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
