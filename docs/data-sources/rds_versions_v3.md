---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_versions_v3"
sidebar_current: "docs-opentelekomcloud-datasource-rds-versions-v3"
description: |-
  Get available RDSv3 versions from OpenTelekomCloud
---

Up-to-date reference of API arguments for RDSv3 versions you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/querying_version_information_about_a_db_engine.html)

# opentelekomcloud_rds_versions_v3

Use this data source to get available OpenTelekomCloud rds versions.

## Example Usage

```hcl
data "opentelekomcloud_rds_versions_v3" "versions" {
  database_name = "mysql"
}
```

## Argument Reference

* `database_name` - (Required) Specifies the DB engine. Value: MySQL, PostgreSQL, SQLServer. Case-insensitive.

## Attributes Reference

In addition, the following attributes are exported:

* `versions` - List of version names, sorted by a version (higher to lower). Example: `["16", "14", "14", "13"]`.
