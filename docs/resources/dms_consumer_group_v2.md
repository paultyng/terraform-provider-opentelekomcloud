---
subcategory: "Distributed Message Service (DMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dms_consumer_group_v2"
sidebar_current: "docs-opentelekomcloud-resource-dms-consumer-group-v2"
description: |-
  Manages an up-to-date DMS Consumer Group v2 resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DMS instance management you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-message-service/api-ref/apis_v2_recommended/instance_management/index.html)

# opentelekomcloud_dms_consumer_group_v2

Manage DMS consumer group v2 resource within OpenTelekomCloud.

## Example Usage

```hcl
variable "instance_id" {}

resource "opentelekomcloud_dms_consumer_group_v2" "group_1" {
  instance_id = var.instance_id
  group_name  = "dms_consumer_group"
  description = "Sample consumer group"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) Specifies the ID of the DMS instance.

  Changing this parameter will create a new resource.

* `group_name` - (Required, String, ForceNew) Specifies the name of the DMS consumer group.

  Changing this parameter will create a new resource.

* `description` - (Optional, String, ForceNew) Specifies any description for the DMS consumer group.

  Changing this parameter will create a new resource.

## Attribute Reference

In addition to all arguments above, the following attribute is exported:

* `state` - Indicates the Consumer group status. The value can be: 
    * Dead: The consumer group has no members and no metadata.
    * Empty: The consumer group has metadata but has no members.
    * PreparingRebalance: The consumer group is to be rebalanced.
    * CompletingRebalance: All members have jointed the group.
    * Stable: Members in the consumer group can consume messages normally.

* `assignment_strategy` - Indicates the partition assignment policy.
* `coordinator_id` - Indicates the coordinator ID.
* `members` - Indicates the consumer list. The structure is documented below.
* `group_message_offsets` - Indicates the consumer offset. The structure is documented below.

The `members` block contains:

* `host` - Indicates the consumer address.
* `member_id` - Indicates the consumer ID.
* `client_id` - Indicates the client ID.
* `assignments` - Indicates the details about the partition assigned to the consumer. The structure is as follows:
  + `topic` - Indicates the topic name.
  + `partitions` - Indicates the partition list. 

The `group_message_offsets` block contains:

* `partition` - Indicates the partition number.
* `lag` - Indicates the number of remaining messages that can be retrieved, that is, the number of accumulated messages.
* `topic` - Indicates the topic name.
* `message_current_offset` - Indicates the consumer offset.
* `message_log_end_offset` - Indicates the log end offset (LEO).

## Import

DMS consumer groups can be imported using their `group_name` and related `instance_id`, separated by a slash, e.g.

```bash
$ terraform import opentelekomcloud_dms_consumer_group_v2.test_group <instance_id>/<group_name>
```
## Notes

But due to some attributes missing from the API response, it's required to ignore changes as below:

```hcl
resource "opentelekomcloud_dms_consumer_group_v2" "group_1" {
  # ...

  lifecycle {
    ignore_changes = [
      description,
    ]
  }
}
```
