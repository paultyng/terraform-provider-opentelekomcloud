---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_parametergroup_v3"
sidebar_current: "docs-opentelekomcloud-resource-rds-parametergroup-v3"
description: |-
  Manages an RDS Parameter Group resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RDS parameter group rule you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/parameter_configuration)

# opentelekomcloud_rds_parametergroup_v3

Manages a RDSv3 parametergroup resource within OpenTelekomCloud.

-> **NOTE:** When you create a PostgreSQL parameter template, some specification parameters do not take effect and are
invisible after the parameter template is created. For more information see [Parameter Template Constraints](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/parameter_configuration/creating_a_parameter_template.html#constraints).

These parameters can be directly applied in `opentelekomcloud_rds_instance_v3` resource by providing a `parameters` argument.

## Example Usage

```hcl
resource "opentelekomcloud_rds_parametergroup_v3" "pg_1" {
  name        = "pg_1"
  description = "some description here"

  values = {
    max_connections = "10"
    autocommit      = "OFF"
  }
  datastore {
    type    = "mysql"
    version = "8.0"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The parameter group name. It contains a maximum of 64 characters.

* `description` - (Optional) The parameter group description. It contains a maximum of 256 characters
  and cannot contain the following special characters: `>!<"&'=` the value is left blank by default.

* `values` - (Optional) Parameter group values key/value pairs defined by users based on the default parameter groups.

* `datastore` - (Required) Database object. The database object structure is documented below. Changing this creates a new parameter group.

The `datastore` block supports:

* `type` - (Required) Specifies the DB engine. Currently, MySQL, PostgreSQL and MS SQLServer are supported.
  The value is case-insensitive and can be `mysql`, `postgresql` or `sqlserver`.

* `version` - (Required) Specifies the database version.
  * MySQL databases support MySQL `5.6`, `5.7`, `8.0`. Example value: `8.0`.
  * PostgreSQL databases support PostgreSQL `11`, `12` and above. Example value: `16`.
  * Microsoft SQL Server databases support `2017_SE`, `2017_EE`, and above. Example value: `2022_EE`.


## Attributes Reference

The following attributes are exported:

* `id` -  ID of the parameter group.

* `configuration_parameters` - Indicates the parameter configuration defined by users based on the default parameters groups.

* `name` - Indicates the parameter name.

* `value` - Indicates the parameter value.

* `restart_required` - Indicates whether a restart is required.

* `readonly` - Indicates whether the parameter is read-only.

* `value_range` - Indicates the parameter value range.

* `type` - Indicates the parameter type.

* `description` - Indicates the parameter description.

* `created` - Indicates the creation time in the following format: `yyyy-MM-ddTHH:mm:ssZ`.

* `updated` - Indicates the update time in the following format: `yyyy-MM-ddTHH:mm:ssZ`.

## Import

Parameter groups can be imported using the `id`, e.g.

```sh
terraform import opentelekomcloud_rds_parametergroup_v3.pg_1 7117d38e-4c8f-4624-a505-bd96b97d024c
```
