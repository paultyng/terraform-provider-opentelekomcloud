---
subcategory: "Direct Connect (DCaaS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dc_virtual_gateway_v3"
sidebar_current: "docs-opentelekomcloud-resource-dc-virtual-gateway-v3"
description: |-
  Manages a Direct Connect Virtual Gateway v3 resource within OpenTelekomCloud.
---

# opentelekomcloud_dc_virtual_gateway_v3

Manages a virtual gateway v3 resource within OpenTelekomCloud.

-> **NOTE:** Direct Connect v3 API that are used in this resource officially supported only on SwissCloud now.

## Example Usage

```hcl
variable "vpc_id" {}
variable "vpc_cidr" {}
variable "gateway_name" {}

resource "opentelekomcloud_dc_virtual_gateway_v3" "gw" {
  vpc_id      = var.vpc_id
  name        = var.gateway_name
  description = "my gateway"

  local_ep_group = [
    var.vpc_cidr,
  ]
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, String, ForceNew) Specifies the ID of the VPC connected to the virtual gateway.
  Changing this will create a new resource.

* `local_ep_group` - (Required, List) Specifies the list of IPv4 subnets from the virtual gateway to access cloud
  services, which is usually the CIDR block of the VPC.

* `local_ep_group_ipv6` - (Optional, List) Specifies the IPv6 subnets of the associated VPC that can be accessed over the virtual gateway.

* `name` - (Required, String) Specifies the name of the virtual gateway.
  The valid length is limited from `3` to `64`, only chinese and english letters, digits, hyphens (-), underscores (_)
  and dots (.) are allowed.
  The Chinese characters must be in `UTF-8` or `Unicode` format.

* `description` - (Optional, String) Specifies the description of the virtual gateway.
  The description contain a maximum of 128 characters and the angle brackets (< and >) are not allowed.
  Chinese characters must be in `UTF-8` or `Unicode` format.

* `asn` - (Optional, Int, ForceNew) Specifies the local BGP ASN of the virtual gateway.
  The valid value is range from `1` to `4,294,967,295`.
  Changing this will create a new resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual gateway.

* `region` - The region where the virtual gateway is located.

* `status` - The current status of the virtual gateway.

## Import

Virtual gateways can be imported using their `id`, e.g.

```shell
$ terraform import opentelekomcloud_dc_virtual_gateway_v3.gw e41748a0-aed9-463e-9817-5c6162265d10
```
