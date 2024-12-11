---
subcategory: "Object Storage Service (OBS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_obs_bucket_object_acl"
sidebar_current: "docs-opentelekomcloud-resource-obs-bucket-object-acl"
description: |-
  Manages a OBS Bucket Object ACL resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for OBS bucket object you can get at
[documentation portal](https://docs.otc.t-systems.com/object-storage-service/api-ref/apis/operations_on_objects)

# opentelekomcloud_obs_bucket_object_acl

Manages an OBS bucket object acl resource within OpenTelekomCloud.

-> **NOTE:** When creating or updating the OBS bucket object acl, the original object acl will be overwritten. When
deleting the OBS bucket object acl, only the owner permissions will be retained, and the other permissions will be
removed.

## Example Usage

```hcl
variable "bucket" {}
variable "key" {}
variable "account1" {}
variable "account2" {}

resource "opentelekomcloud_obs_bucket_acl" "test" {
  bucket = var.bucket
  key    = var.key

  account_permission {
    access_to_object = ["READ"]
    access_to_acl    = ["READ_ACP", "WRITE_ACP"]
    account_id       = var.account1
  }

  account_permission {
    access_to_object = ["READ"]
    access_to_acl    = ["READ_ACP"]
    account_id       = var.account2
  }

  public_permission {
    access_to_acl = ["READ_ACP", "WRITE_ACP"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, String, ForceNew) Specifies the name of the bucket which the object belongs to.

  Changing this parameter will create a new resource.

* `key` - (Required, String, ForceNew) Specifies the name of the object to which to set the acl.

  Changing this parameter will create a new resource.

* `public_permission` - (Optional, List) Specifies the object public permission.
  The [permission_struct](#OBSBucketObjectAcl_permission_struct) structure is documented below.

* `account_permission` - (Optional, List) Specifies the object account permissions.
  The [account_permission_struct](#OBSBucketObjectAcl_account_permission_struct) structure is documented below.

<a name="OBSBucketObjectAcl_permission_struct"></a>
The `permission_struct` block supports:

* `access_to_object` - (Optional, List) Specifies the access to object. Only **READ** supported.

* `access_to_acl` - (Optional, List) Specifies the access to acl. Valid values are **READ_ACP** and **WRITE_ACP**.

<a name="OBSBucketObjectAcl_account_permission_struct"></a>
The `account_permission_struct` block supports:

* `account_id` - (Required, String) Specifies the account id to authorize. The account id cannot be the object owner,
  and must be unique.

* `access_to_object` - (Optional, List) Specifies the access to object. Only **READ** supported.

* `access_to_acl` - (Optional, List) Specifies the access to acl. Valid values are **READ_ACP** and **WRITE_ACP**.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket object key.

* `region` - The region of the bucket.

* `owner_permission` - The object owner permission information.
  The [owner_permission_struct](#OBSBucketObjectAcl_owner_permission_struct) structure is documented below.

<a name="OBSBucketObjectAcl_owner_permission_struct"></a>
The `owner_permission_struct` block supports:

* `access_to_object` - The owner object permissions.

* `access_to_acl` - The owner acl permissions.

## Import

The obs bucket object acl can be imported using `bucket` and `key`, separated by a slash, e.g.

```bash
$ terraform import opentelekomcloud_obs_bucket_acl.test <bucket>/<key>
```
