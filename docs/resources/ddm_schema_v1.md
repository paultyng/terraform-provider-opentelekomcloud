---
subcategory: "Distributed Database Middleware (DDM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ddm_schema_v1"
sidebar_current: "docs-opentelekomcloud-resource-ddm-schema-v1"
description: |-
  Manages a DDM Schema resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DDM schema you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-database-middleware/api-ref/apis_recommended/schemas/)

# opentelekomcloud_ddm_instance_v1

Manages DDM schema resource within OpenTelekomCloud

## Example Usage: Creating a basic DDM schema
```hcl
variable "username" {}
variable "password" {}

resource "opentelekomcloud_ddm_schema_v1" "schema_1" {
  name         = "ddm_schema"
  instance_id  = "b4cd6aeb0b7445d3bf271457c6941544in09"
  shard_mode   = "cluster"
  shard_number = 8
  shard_unit   = 8
  rds {
    id             = "55d93e249b77461b81f990fa805db3f3in01"
    admin_username = var.username
    admin_password = var.password
  }
  purge_rds_on_delete = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) Specifies the DDM schema name. The DDM instance name of the same
  type is unique in the same tenant. It can be  2 to 48 characters long. It must start with a letter and it can only contain etters, digits, and underscores (_).

* `instance_id` - (Required, List, ForceNew) Specifies the DDM instance ID.

* `shard_mode` - (Required, String, ForceNew) Specifies the sharding mode of the schema. The values for this can be `cluster` or `single`. Cluster indicates that the schema is in sharded mode. Single indicates that the schema is in unsharded mode.

* `shard_number` - (Required, Integer, ForceNew) Specifies the number of shards in the same working mode. If `shard_unit` is not empty, the value is the **_product of shard_unit multiplied by the associated RDS instances_**. If shard_unit is left blank, the value must be **_greater than the number of associated RDS instances and less than or equal to the product of the associated RDS instances multiplied by 64_**.

* `shard_unit` - (Optional, Integer, ForceNew) Specifies the Number of shards per RDS instance. The value is 1 if the schema is unsharded. The value ranges from 1 to 64 if the schema is sharded.

* `purge_rds_on_delete` - (Optional, Integer, ForceNew) Specifies whether data stored on the associated DB instances is deleted. The value can be: `true` or `false` (default)

* `rds` - (Required, List, ForceNew) Specifies the rds instance information. The structure is described below.

The `rds` block supports:

- `id` - (Required, String, ForceNew) Specifies the ID of the rds instance.

- `admin_username` - (Required, String, ForceNew) Specifies the username of RDS admin.

- `admin_password` - (Required, String, ForceNew) Specifies the password of RDS admin.


`NOTE:` Currently DDM schema supports only MySQL RDS databases. Also the parameter, `lower_case_table_names`, must be set to 1 in RDS (on console, Table Name: Case insensitive).

## Attributes Reference

The following attributes are exported:

* `region` - The region of the DDM instance.
* `name` - See Argument Reference above.
* `instance_id` - See Argument Reference above.
* `shard_mode` - See Argument Reference above.
* `shard_number` - See Argument Reference above.
* `shard_unit` - See Argument Reference above.
* `rds` - See Argument Reference above.
* `purge_rds_on_delete` - See Argument Reference above.
* `status` - (String) Indicates the DDM schema status.
* `created_at` - (uint64) Indicates the creation time.
* `updated_at` - (uint64) Indicates the update time.
* `data_vips` - (List) Indicates the IP address and port number for connecting to the schema.
* `used_rds` - (List) Indicates the associated RDS instances. The structure is described below.
* `databases` - (List) Indicates the Sharding information of the schema. The structure is described below.

The `used_rds` block contains:

  - `id` - (String) Indicates the RDS ID.
  - `name` - Indicates the RDS name.
  - `status` - Indicates the RDS status.

The `databases` block contains:

  - `db_slot` - (String) Indicates the Number of shards.
  - `name` - (String) Indicates the shard name.
  - `status` - (String) Indicates the shard status.
  - `created` - (uint64) Indicates the screation time.
  - `updated` - (uint64) Indicates the update time.
  - `id` - (String) ID of the RDS instance where the shard is located.
  - `rds_name` (String) Name of the RDS instance where the shard is located


## Import

DDMv1 Instance can be imported using the DDM instance ID, `instance_id` and DDM schema `name`, e.g.

```shell
terraform import opentelekomcloud_ddm_schema_v1.schema_1 b4cd6aeb0b7445d3bf271457c6941544in09/ddm_schema
```

## Notes

But due to some attributes missing from the API response, it's required to ignore changes as below:

```hcl
resource "opentelekomcloud_ddm_schema_v1" "schema_1" {
  # ...

  lifecycle {
    ignore_changes = [
      rds,      
      updated_at,
      purge_rds_on_delete
    ]
  }
}
```
