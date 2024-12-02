---
subcategory: "APIGW"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_apigw_gateway_routes_v2"
sidebar_current: "docs-opentelekomcloud-resource-apigw-gateway-routes-v2"
description: |-
  Manages a APIGW gateway routes resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for API Gateway environment variable service you can get at
[documentation portal](https://docs.otc.t-systems.com/api-gateway/api-ref/dedicated_gateway_apis_v2/gateway_feature_management/configuring_a_feature_for_a_gateway.html)

# opentelekomcloud_apigw_gateway_routes_v2

Manages a APIGW gateway routes resource within OpenTelekomCloud.

## Example Usage

```hcl
variable "gateway_id" {}

resource "opentelekomcloud_apigw_gateway_routes_v2" "rt" {
  gateway_id = var.gateway_id
  nexthops   = ["172.16.3.0/24", "172.16.7.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `gateway_id` - (Required, String, ForceNew) Specifies the ID of the dedicated gateway to which the routes belong.
  Changing this will create a new resource.

* `nexthops` - (Required, List) Specifies the configuration of the next-hop routes.

-> The network segment of the next hop cannot overlap with the network segment of the APIGW gateway.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID (gateway ID).

* `region` - The region where the dedicated gateway and routes are located.

## Import

Routes can be imported using their related dedicated instance ID (`gateway_id`), e.g.

```bash
$ terraform import opentelekomcloud_apigw_gateway_routes_v2.rt 628001b3c5eg6d3e91a8da530f46427y
```
