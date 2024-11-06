---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_advanced_query_v1"
sidebar_current: "docs-opentelekomcloud-datasource-rms-advanced-query-v1"
description: |-
  Manages an RMS Advanced Query data source within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RMS Advanced Query you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/advanced_queries/running_advanced_queries.html#rms-04-0701-response-queryinfo)


# opentelekomcloud_rms_advanced_query_v1

Use this data source to do an RMS advanced query.

## Example Usage

```hcl
variable "exression" {}

data "opentelekomcloud_rms_advanced_query_v1" "test" {
  exression = var.exression
}
```

## Argument Reference

The following arguments are supported:

* `expression` - (Required, String) Specifies the expression of the query.

  For example, **select name, id from tracked_resources where provider = 'ecs' and type = 'cloudservers'**

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID.

* `results` - The list of query results.

* `query_info` - The query info.

  The [query_info](#query_info) structure is documented below.

<a name="query_info"></a>
The `query_info` block supports:

* `select_fields` - The list of select fields.
