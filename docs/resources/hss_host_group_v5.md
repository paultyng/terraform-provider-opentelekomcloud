---
subcategory: "Host Security Service (HSS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_hss_host_group_v5"
sidebar_current: "docs-opentelekomcloud-resource-hss-host-group-v5"
description: |-
  Manages an HSS host group Service resource within OpenTelekomCloud.
---

# opentelekomcloud_hss_host_group_v5

Manages an HSS host group resource within OpenTelekomCloud.

## Example Usage

### Create an HSS host group and bind ECS instances

```hcl
variable "host_group_name" {}
variable "host_ids" {
  type = list(string)
}

resource "opentelekomcloud_hss_host_group_v5" "test" {
  name     = var.host_group_name
  host_ids = var.host_ids
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) Specifies the name of the host group.
  The valid length is limited from `1` to `64`, only Chinese characters, English letters, digits, hyphens (-),
  underscores (_), dots (.), pluses (+) and asterisks (*) are allowed.
  The Chinese characters must be in `UTF-8` or `Unicode` format.

* `host_ids` - (Required, List) Specifies the list of host IDs.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID in UUID format.

* `host_num` - The total host number.

* `region` - The region where the host group is located.

* `risk_host_num` - The number of hosts at risk.

* `unprotect_host_num` - The number of unprotect hosts.

* `unprotect_host_ids` - The ID list of the unprotect hosts.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 30 minutes.
* `update` - Default is 30 minutes.

## Import

The host group resource can be imported using `id`, e.g.

```bash
$ terraform import opentelekomcloud_hss_host_group_v5.group <id>
```
