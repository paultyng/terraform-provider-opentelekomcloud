---
subcategory: "Distributed Message Service (DMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dms_smart_connect_task_action_v2"
sidebar_current: "docs-opentelekomcloud-resource-dms-smart-connect-task-action-v2"
description: |-
  Start or pause an up-to-date DMS Smart Connect Task v2 resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DMS instance you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-message-service/api-ref/apis_v2_recommended/smart_connect/index.html)

# opentelekomcloud_dms_smart_connect_task_action_v2

Start or pause a DMS kafka smart connect task resource within OpenTelekomCloud.

## Example Usage

### Pause a task

```hcl
variable "instance_id" {}
variable "task_id" {}

resource "opentelekomcloud_dms_smart_connect_task_action_v2" "test" {
  instance_id = var.instance_id
  task_id     = var.task_id
  action      = "pause"
}
```

### Resume a paused task

```hcl
variable "instance_id" {}
variable "task_id" {}

resource "opentelekomcloud_dms_smart_connect_task_action_v2" "test" {
  instance_id = var.instance_id
  task_id     = var.task_id
  action      = "resume"
}
```

### Start or restart a running or paused task

```hcl
variable "instance_id" {}
variable "task_id" {}

resource "opentelekomcloud_dms_smart_connect_task_action_v2" "test" {
  instance_id = var.instance_id
  task_id     = var.task_id
  action      = "restart"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) Specifies the kafka instance ID.

* `task_id` - (Required, String, ForceNew) Specifies the smart connect task ID.

* `action` - (Required, String, ForceNew) Specifies the action to be performed on the smart connect task.
  Supported values: `pause`, `resume`, `restart`

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `task_status` - Indicates the status of the smart connect task.

* `region` - The region in which the resource is created.


## Timeouts

This resource provides the following timeout configuration options:

* `create` - Default is 30 minutes.
