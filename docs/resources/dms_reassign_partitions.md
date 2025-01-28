---
subcategory: "Distributed Message Service (DMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dms_reassign_partitions_v2"
sidebar_current: "docs-opentelekomcloud-resource-dms-reassign-partitions-v2"
description: |-
  Initiate partition reassignment for an up-to-date DMS topic within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DMS partition reassigning you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-message-service/api-ref/apis_v2_recommended/instance_management/initiating_partition_reassigning_for_a_kafka_instance.html)

# opentelekomcloud_dms_reassign_partitions_v2

Initiate partition reassignment for an up-to-date DMS topic resource within OpenTelekomCloud.

## Example Usage

```hcl
variable "instance_id" {}
variable "topic_name" {}

resource "opentelekomcloud_dms_reassign_partitions_v2" "rp_1" {
  instance_id   = var.instance_id
  throttle      = 1000000
  time_estimate = false
  reassignments {
    topic              = var.topic_name
    replication_factor = 3
    brokers            = [0, 1, 2]
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) Specifies the DMS instance ID.

* `reassignments` - (Required, List, ForceNew) Specifies the reassignment plan.
  The [reassignments](#dms_reassignments) structure is documented below.

* `throttle` - (Optional, Integer, ForceNew) Specifies the reassignment threshold.

* `is_schedule` - (Optional, Boolean, ForceNew) Specifies whether the task is scheduled. If **no**, `is_schedule` and `execute_at` can be left blank. If **yes**, `is_schedule is true` and `execute_at` must be specified.

* `execute_at` - (Optional, Integer, ForceNew) Specifies the schedule time. The value is a UNIX timestamp, in ms. 

* `time_estimate` - (Optional, Boolean, ForceNew) Specifies whether to perform time estimation or relabancing tasks. Set **true** to perform time estimation tasks and **false** to perform rebalancing tasks.

<a name="dms_reassignments"></a>
The `reassignments` block supports:

* `topic` - (Required, String, ForceNew) Specifies the topic name.

* `brokers` - (Optional, List, ForceNew) Specifies the list of brokers to which partitions are reassigned. 
  **Note:** This parameter is **mandatory** in automatic assignment.

* `replication_factor` - (Optional, Integer, ForceNew) Specifies the replication factor, which can be specified in automatic assignment.

* `assignments` - (Optional, List, ForceNew) Specifies the manually specified assignment plan.
  **Note:** The `brokers` parameter and `assignments` parameter cannot be empty at the same time.
  The [assignments](#dms_assignments) structure is documented below.

<a name="dms_assignments"></a>
The `assignments` block supports:

* `partition` - (Optional, Integer, ForceNew) Specifies the partition number in manual assignment.

* `partition_brokers` - (Optional, List, ForceNew) Specifies the list of brokers to be assigned to a partition in manual assignment.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `reassignment_time` - Indicates the estimated time, in seconds. Only reassignment_time is returned for a time estimation task.

* `region` - The region in which the resource is created.
