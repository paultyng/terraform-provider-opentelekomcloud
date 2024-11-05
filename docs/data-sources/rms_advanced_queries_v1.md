---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_advanced_queries_v1"
sidebar_current: "docs-opentelekomcloud-datasource-rms-advanced-queries-v1"
description: |-
  Manages an RMS Advanced Queries data source within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RMS Advanced Queries you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/advanced_queries/listing_advanced_queries.html#rms-04-0703)


# opentelekomcloud_rms_advanced_queries_v1

Use this data source to get the list of RMS advanced queries.

## Example Usage

```hcl
variable "advanced_query_name" {}

data "opentelekomcloud_rms_advanced_queries_v1" "test" {
  name = var.advanced_query_name
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, String) Specifies the advanced query name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID.

* `region` - The queries region.

* `queries` - The list of advanced queries.

  The [queries](#queries_struct) structure is documented below.

<a name="queries_struct"></a>
The `queries` block supports:

* `name` - The advanced query name.

* `id` - The advanced query ID.

* `type` - The advanced query type.

* `description` - The advanced query description.

* `expression` - The advanced query expression.

* `created_at` - The creation time of the advanced query.

* `updated_at` - The latest update time of the advanced query.
