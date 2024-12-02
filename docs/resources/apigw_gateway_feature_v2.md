---
subcategory: "APIGW"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_apigw_gateway_feature_v2"
sidebar_current: "docs-opentelekomcloud-resource-apigw-gateway-feature-v2"
description: |-
  Manages a APIGW gateway feature resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for API Gateway environment variable service you can get at
[documentation portal](https://docs.otc.t-systems.com/api-gateway/api-ref/dedicated_gateway_apis_v2/gateway_feature_management/configuring_a_feature_for_a_gateway.html)

# opentelekomcloud_apigw_gateway_feature_v2

Manages an APIGW gateway feature resource within OpenTelekomCloud.

-> For various types of feature parameter configurations, please refer to the
   [documentation](https://docs.otc.t-systems.com/api-gateway/api-ref/appendix/supported_features.html#apig-api-20200402).

## Example Usage

```hcl
variable "gateway_id" {}

resource "opentelekomcloud_apigw_gateway_feature_v2" "feat" {
  gateway_id = var.gateway_id
  name       = "ratelimit"
  enabled    = true

  config = jsonencode({
    api_limits = 300
  })
}
```

## Argument Reference

The following arguments are supported:

* `gateway_id` - (Required, String, ForceNew) Specified the ID of the dedicated gateway to which the feature belongs.
  Changing this creates a new resource.

* `name` - (Required, String, ForceNew) Specified the name of the feature.
  Changing this creates a new resource.

* `enabled` - (Optional, Bool) Specified whether to enable the feature. Default value is `false`.

* `config` - (Optional, String) Specified the detailed configuration of the feature.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID. The value is the feature name.

* `region` - The region in which to create the resource.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 5 minutes.

## Import

The resource can be imported using `gateway_id` and `name`, separated by a slash (/), e.g.

```bash
$ terraform import opentelekomcloud_apigw_gateway_feature_v2.feat <gateway_id>/<name>
```
