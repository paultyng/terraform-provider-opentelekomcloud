---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_flavors_v3"
sidebar_current: "docs-opentelekomcloud-datasource-rds-flavors-v3"
description: |-
  Get available RDSv3 flavors from OpenTelekomCloud
---

Up-to-date reference of API arguments for RDSv3 flavor you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/querying_database_specifications.html)

# opentelekomcloud_rds_flavors_v3

Use this data source to get available OpenTelekomCloud RDSv3 flavors.

## Example Usage

```hcl
data "opentelekomcloud_rds_flavors_v3" "flavor" {
  db_type       = "PostgreSQL"
  db_version    = "16"
  instance_mode = "ha"
}
```

## Argument Reference

* `db_type` - (Required) Specifies the DB engine. Possible values are: `MySQL`, `PostgreSQL`, `SQLServer`.

* `db_version` - (Required) Specifies the database version.
  * MySQL: 8.0, 5.7, and 5.6
  * PostgreSQL: 11 through 16
  * Microsoft SQL Server: 2017 (Enterprise/Standard) through 2022 (Enterprise/Standard)

* `instance_mode` - (Required) The mode of instance. Possible values are:
  `ha` indicates primary/standby instance, `single` indicates single instance
  and `replica` indicates read-replica instance.

## Attributes Reference

In addition, the following attributes are exported:

* `flavors` - Indicates the `flavors` information. Structure is documented below.

The `flavors` block contains:

* `name` - The name of the rds flavor.

* `vcpus` - Indicates the CPU size.

* `memory` - Indicates the memory size in GB.

* `mode` - Indicates the DB instance type.

* `az_status` - Indicates the status of the AZ to which the DB instance specifications belong.
