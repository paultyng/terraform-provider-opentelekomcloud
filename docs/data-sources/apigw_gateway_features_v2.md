---
subcategory: "APIGW"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_apigw_gateway_features_v2"
sidebar_current: "docs-opentelekomcloud-datasource-apigw-gateway-features-v2"
description: |-
  Get the all APIGW gateway features from OpenTelekomCloud
---

Up-to-date reference of API arguments for API Gateway environment variable service you can get at
[documentation portal](https://docs.otc.t-systems.com/api-gateway/api-ref/dedicated_gateway_apis_v2/gateway_feature_management/querying_gateway_features.html)

# opentelekomcloud_apigw_gateway_features_v2

Use this data source to get the list of the features under the APIGW gateway within OpenTelekomCloud.

## Example Usage

```hcl
variable gateway_id {}

data "opentelekomcloud_apigw_gateway_features_v2" "ft" {
  gateway_id = var.gateway_id
}
```

## Argument Reference

The following arguments are supported:

* `gateway_id` - (Required, String) Specified the ID of the dedicated gateway to which the features belong.

* `name` - (Optional, String) Specified the name of the feature.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID.

* `region` - The region in which to query the data source.

* `features` - All instance features that match the filter parameters.
  The [features](#instance_features) structure is documented below.

<a name="instance_features"></a>
The `features` block supports:

* `id` - The ID of the feature.

* `name` - The name of the feature.

* `enabled` - Whether the feature is enabled.

* `config` - The detailed configuration of the instance feature.

* `updated_at` - The latest update time of the feature, in RFC3339 format.
