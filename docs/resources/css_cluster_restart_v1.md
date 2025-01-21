---
subcategory: "Cloud Search Service (CSS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_css_cluster_restart_v1"
sidebar_current: "docs-opentelekomcloud-resource-css-cluster-restart-v1"
description: |-
  Manages CSS cluster restart resource within OpenTelekomCloud.
---

# opentelekomcloud_css_cluster_restart_v1

Manages CSS cluster restart resource within OpenTelekomCloud.

## Example Usage

```hcl
variable "cluster_id" {}

resource "opentelekomcloud_css_cluster_restart_v1" "test" {
  cluster_id = var.cluster_id
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required, String, ForceNew) Specifies the ID of the CSS cluster.
  Changing this creates a new resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID.

* `region` - The region in which the resource created.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 60 minutes.
