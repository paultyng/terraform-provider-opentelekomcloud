---
subcategory: "Distributed Database Middleware (DDM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ddm_engines_v1"
sidebar_current: "docs-opentelekomcloud-datasource-ddm-engines-v1"
description: |-
  Get DDM engines from OpenTelekomCloud.
---

Up-to-date reference of API arguments for DDM engines you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-database-middleware/api-ref/apis_recommended/ddm_instances/querying_ddm_engine_information.html)

# opentelekomcloud_ddm_engines_v1

Use this data source to get info of OpenTelekomCloud DDM engines.

## Example Usage
```hcl
data "opentelekomcloud_ddm_engines_v1" "engine_list" {
}
```


## Attributes Reference

The following attributes are exported:

* `region` - Indicates the region of the DDM engines.
* `engines` - Indicates the DDM engines information. Structure is documented below.

The `engines` block contains:

  - `id` - Indicates the DDM engine ID.
  - `name` - Indicates the DDM engine name.
  - `version` - Indicates the DDM engine version.
  - `availability_zones` - Indicates the supported availability zones. Structure is documented below

The `availability_zones` block contains:

  - `name` - Indicates the AZ name.
  - `code` - Indicates the AZ code.
  - `favored` - Indicates whether current AZ is favored.
