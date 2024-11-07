---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_advanced_query_schemas_v1"
sidebar_current: "docs-opentelekomcloud-datasource-rms-advanced-query-schemas-v1"
description: |-
  Manages an RMS Advanced Query Schemas data source within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RMS Advanced Query Schemas you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/advanced_queries/querying_schemas.html)


# opentelekomcloud_rms_advanced_query_schemas_v1

Use this data source to get the list of RMS advanced query schemas.

## Example Usage

```hcl
data "opentelekomcloud_rms_advanced_query_schemas_v1" "test" {
  type = "aad.instances"
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Optional, String) Specifies the type of the schema.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID.

* `schemas` - The list of schema.

  The [schemas](#schemas_struct) structure is documented below.

<a name="schemas_struct"></a>
The `schemas` block supports:

* `type` - The schema type.

* `schema` - The schema detail.
