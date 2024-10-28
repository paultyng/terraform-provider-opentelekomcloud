---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_advanced_query_v1"
sidebar_current: "docs-opentelekomcloud-rms_advanced_query-v1"
description: |-
  Manages an RMS Advanced query resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RDS replica you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/advanced_queries/index.html)

# opentelekomcloud_rms_advanced_query_v1

Manages an RMS advanced query resource within OpenTelekomCloud.

## Example Usage

```hcl
resource "opentelekomcloud_rms_advanced_query_v1" "test" {
  name       = "advanced_query_name"
  expression = "select * from table_test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) Specifies the advanced query name. It contains 1 to 64 characters.

  Changing this parameter will create a new resource.

* `expression` - (Required, String) Specifies the advanced query expression. It contains 1 to 4096 characters.

* `description` - (Optional, String) Specifies the advanced query description. It contains 1 to 512 characters.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID.

* `type` - The resource type.

* `created_at` - The resource creation time.

* `updated_at` - The resource update time.

## Import

The RMS advanced query can be imported using the `id`, e.g.

```bash
$ terraform import opentelekomcloud_rms_advanced_query_v1.test <id>
```
