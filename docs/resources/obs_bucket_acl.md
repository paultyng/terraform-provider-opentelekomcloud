---
subcategory: "Object Storage Service (OBS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_obs_bucket_acl"
sidebar_current: "docs-opentelekomcloud-resource-obs-bucket-acl"
description: |-
  Manages a OBS Bucket ACL resource within OpenTelekomCloud.
---


# opentelekomcloud_obs_bucket_acl

Manages an OBS bucket acl resource within OpenTelekomCloud.

-> **NOTE:** When creating or updating the OBS bucket acl, the original bucket acl will be overwritten. When deleting
the OBS bucket acl, the full permissions of the bucket owner will be set, and the other permissions will be removed.

## Example Usage

```hcl
variable "bucket" {}
variable "account1" {}
variable "account2" {}

resource "opentelekomcloud_obs_bucket_acl" "test" {
  bucket = var.bucket

  owner_permission {
    access_to_bucket = ["READ", "WRITE"]
    access_to_acl    = ["READ_ACP", "WRITE_ACP"]
  }

  account_permission {
    access_to_bucket = ["READ", "WRITE"]
    access_to_acl    = ["READ_ACP", "WRITE_ACP"]
    account_id       = var.account1
  }

  account_permission {
    access_to_bucket = ["READ"]
    access_to_acl    = ["READ_ACP", "WRITE_ACP"]
    account_id       = var.account2
  }

  public_permission {
    access_to_bucket = ["READ", "WRITE"]
  }

  log_delivery_user_permission {
    access_to_bucket = ["READ", "WRITE"]
    access_to_acl    = ["READ_ACP", "WRITE_ACP"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, String, ForceNew) Specifies the name of the bucket to which to set the acl.

  Changing this parameter will create a new resource.

* `owner_permission` - (Optional, List) Specifies the bucket owner permission. If omitted, the current obs bucket acl
  owner permission will not be changed.
  The [permission_struct](#OBSBucketAcl_permission_struct) structure is documented below.

* `public_permission` - (Optional, List) Specifies the public permission.
  The [permission_struct](#OBSBucketAcl_permission_struct) structure is documented below.

* `log_delivery_user_permission` - (Optional, List) Specifies the log delivery user permission.
  The [permission_struct](#OBSBucketAcl_permission_struct) structure is documented below.

* `account_permission` - (Optional, List) Specifies the account permissions.
  The [account_permission_struct](#OBSBucketAcl_account_permission_struct) structure is documented below.

<a name="OBSBucketAcl_permission_struct"></a>
The `permission_struct` block supports:

* `access_to_bucket` - (Optional, List) Specifies the access to bucket. Valid values are **READ** and **WRITE**.

* `access_to_acl` - (Optional, List) Specifies the access to acl. Valid values are **READ_ACP** and **WRITE_ACP**.

<a name="OBSBucketAcl_account_permission_struct"></a>
The `account_permission_struct` block supports:

* `access_to_bucket` - (Optional, List) Specifies the access to bucket. Valid values are **READ** and **WRITE**.

* `access_to_acl` - (Optional, List) Specifies the access to acl. Valid values are **READ_ACP** and **WRITE_ACP**.

* `account_id` - (Required, String) Specifies the account id to authorize. The account id cannot be the bucket owner,
  and must be unique.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.


* `region` - The region in which resource is created.

## Import

The obs bucket acl can be imported using the `bucket`, e.g.

```bash
$ terraform import opentelekomcloud_obs_bucket_acl.test <bucket-name>
```
