---
subcategory: "Object Storage Service (OBS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_obs_bucket_replication"
sidebar_current: "docs-opentelekomcloud-resource-obs-bucket-replication"
description: |-
  Manages a OBS Bucket Replication resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for OBS bucket cross-region replication you can get at
[documentation portal](https://docs.otc.t-systems.com/object-storage-service/api-ref/apis/advanced_bucket_settings/configuring_cross-region_replication_for_a_bucket.html)

# opentelekomcloud_obs_bucket_replication

Manages an OBS bucket **Cross-Region Replication** resource within OpenTelekomCloud.

-> **NOTE:** The source bucket and destination bucket must belong to the same account. More cross-Region replication
constraints see [Cross-Region replication](https://docs.otc.t-systems.com/object-storage-service/api-ref/apis/advanced_bucket_settings/configuring_cross-region_replication_for_a_bucket.html#obs-04-0046)

## Example Usage

### Replicate all objects

```hcl
variable "bucket" {}
variable "destination_bucket" {}
variable "agency" {}

resource "opentelekomcloud_obs_bucket_replication" "test" {
  bucket             = var.bucket
  destination_bucket = var.destination_bucket
  agency             = var.agency

  rule {
    history_enabled = false
    delete_data     = false
  }
}
```

### Replicate objects matched by prefix

```hcl
variable "bucket" {}
variable "destination_bucket" {}
variable "agency" {}

resource "opentelekomcloud_obs_bucket_replication" "test" {
  bucket             = var.bucket
  destination_bucket = var.destination_bucket
  agency             = var.agency

  rule {
    prefix = "log"
  }

  rule {
    prefix          = "imgs/"
    storage_class   = "COLD"
    enabled         = true
    history_enabled = false
    delete_data     = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, ForceNew) Specifies the region in which to create the resource.
  If omitted, the provider-level region will be used.

  Changing this parameter will create a new resource.

* `bucket` - (Required, ForceNew) Specifies the name of the source bucket.

  Changing this parameter will create a new resource.

* `destination_bucket` - (Required) Specifies the name of the destination bucket.

  -> **NOTE:** The destination bucket cannot be in the region where the source bucket resides.
  Some regions do not support cross regional replication. More constraints information see:
  [Cross-Region replication](https://docs.otc.t-systems.com/object-storage-service/api-ref/apis/advanced_bucket_settings/configuring_cross-region_replication_for_a_bucket.html#obs-04-0046)

* `agency` - (Required) Specifies the IAM agency name applied to the cross-region replication.

  -> **NOTE:** The IAM agency is a cloud service agency of OBS. Which must has the **OBS Administrator** permission.

* `rule` - (Optional) Specifies the configurations of object cross-region replication management.
  The structure is documented below.

The `rule` block supports:

* `prefix` - (Optional) Specifies the prefix of an object key name, applicable to one or more objects.
  The maximum length of a prefix is 1024 characters.
  Duplicated prefixes are not supported. If omitted, all objects in the bucket will be managed by the lifecycle rule.
  To copy a folder, end the prefix with a slash (/), for example, imgs/.

* `storage_class` - (Optional) Specifies the storage class for replicated objects. Valid values are `STANDARD`,
  `WARM` (Infrequent Access) and `COLD` (Archive).
  If omitted, the storage class of object copies is the same as that of objects in the source bucket.

* `enabled` - (Optional) Specifies cross-region replication rule status. Defaults to `true`.

* `history_enabled` - (Optional) Specifies cross-region replication history rule status. Defaults to `false`.
  If the value is `true`, historical objects meeting this rule are copied.

* `delete_data` - (Optional) Specifies cross-region replication object deletion operations status. Defaults to `false`.
  If the value is `true`, object deletion for the source bucket will be replicated to the destination bucket.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.
* `rule/id` - The ID of a rule in UUID format.

## Import

The obs bucket cross-region replication can be imported using the `bucket`, e.g.

```bash
$ terraform import opentelekomcloud_obs_bucket_replication.test <bucket-name>
```
