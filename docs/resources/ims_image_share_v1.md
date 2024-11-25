---
subcategory: "Image Management Service (IMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ims_image_share_v1"
sidebar_current: "docs-opentelekomcloud-resource-ims-image-share-v1"
description: |-
Manages a IMS Image Share resource within OpenTelekomCloud.
---

# opentelekomcloud_ims_image_share_v1

Manages an IMS image share resource within OpenTelekomCloud.

## Example Usage

```hcl
variable "source_image_id" {}
variable "target_project_ids" {}

resource "opentelekomcloud_ims_image_share_v1" "share" {
  source_image_id    = var.source_image_id
  target_project_ids = var.target_project_ids
}
```

## Argument Reference

The following arguments are supported:
* `source_image_id` - (Required, String, ForceNew) Specifies the ID of the source image. The source image must be in the
  same region as the current resource.

  Changing this parameter will create a new resource.

* `target_project_ids` - (Required, List) Specifies the IDs of the target projects.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID, same as `source_image_id`.

* `region` - The region in which resource is located.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 5 minutes.
* `delete` - Default is 5 minutes.
