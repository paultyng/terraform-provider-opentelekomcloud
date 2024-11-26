---
subcategory: "Image Management Service (IMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ims_image_share_accept_v1"
sidebar_current: "docs-opentelekomcloud-resource-ims-image-share-accept-v1"
description: |-
  Manages an IMS Image Share Accept resource within OpenTelekomCloud.
---

# opentelekomcloud_ims_image_share_accept_v1

Manages an IMS image share accept resource within OpenTelekomCloud.

-> Creating resource means accepting shared image, while destroying resource means rejecting shared image.

## Example Usage

```hcl
variable "shared_image_id" {}

resource "opentelekomcloud_ims_image_share_accept_v1" "acc" {
  image_id = var.shared_image_id
}
```

## Argument Reference

The following arguments are supported:
* `image_id` - (Required, String, ForceNew) Specifies the ID of the image.

  Changing this parameter will create a new resource.

* `vault_id` - (Optional, String, ForceNew) Specifies the ID of a vault. This parameter is mandatory if you want
  to accept a shared full-ECS image created from a CBR backup.

  Changing this parameter will create a new resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID in UUID format.

* `region` - The region in which resource is located.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 5 minutes.
* `delete` - Default is 5 minutes.
