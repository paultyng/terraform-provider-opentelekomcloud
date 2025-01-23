---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_flavors_v1"
sidebar_current: "docs-opentelekomcloud-datasource-rds-flavors-v1"
description: |-
  Get details about RDSv1 flavor from OpenTelekomCloud
---

Up-to-date reference of API arguments for RDSv1 flavor you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v1_to_be_discarded/db_instance_management/obtaining_all_db_instance_specifications.html#en-us-topic-0032347783)

**DEPRECATED**
# opentelekomcloud_rds_flavors_v1

Use this data source to get the ID of an available OpenTelekomCloud RDS flavor.

## Example Usage

```hcl
data "opentelekomcloud_rds_flavors_v1" "flavor" {
  datastore_name    = "PostgreSQL"
  datastore_version = "16"
  speccode          = "rds.pg.x1.xlarge.4"
}
```

## Argument Reference

* `datastore_name` - (Required) The datastore name of the rds.

* `datastore_version` - (Required) The datastore version of the rds.

* `speccode` - (Optional) The spec code of a rds flavor.

## Attributes Reference

`id` is set to the ID of the found rds flavor. In addition, the following attributes are exported:

* `datastore_name` - See Argument Reference above.

* `datastore_version` - See Argument Reference above.

* `speccode` - See Argument Reference above.

* `name` - The name of the rds flavor.

* `ram` - The name of the rds flavor.
