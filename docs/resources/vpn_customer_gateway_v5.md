---
subcategory: "Virtual Private Network (VPN)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_enterprise_vpn_customer_gateway_v5"
sidebar_current: "docs-opentelekomcloud-resource-enterprise-vpn-customer-gateway-v5"
description: |-
  Manages a Enterprise VPN Customer Gateway Service resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for EVPN you can get at
[documentation portal](https://docs.otc.t-systems.com/virtual-private-network/api-ref/api_reference_enterprise_edition_vpn/apis_of_enterprise_edition_vpn/customer_gateway/index.html)

# opentelekomcloud_enterprise_vpn_customer_gateway_v5

Manages a VPN customer gateway resource within OpenTelekomCloud.

## Example Usage

### Manages a common VPN customer gateway

```hcl
variable "name" {}
variable "id_value" {}

resource "opentelekomcloud_enterprise_vpn_customer_gateway_v5" "test" {
  name     = var.name
  id_value = var.id_value
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The customer gateway name.
  The valid length is limited from `1` to `64`, only letters, digits, hyphens (-) and underscores (_) are allowed.

* `id_value` - (Required, String, ForceNew) Specifies the identifier of a customer gateway.
  When `id_type` is set to `ip`, the value is an IPv4 address in dotted decimal notation, for example, 192.168.45.7.
  When `id_type` is set to `fqdn`, the value is a string of characters that can contain uppercase letters, lowercase letters,
  digits, and special characters. Spaces and the following special characters are not supported: & < > [ ] \ ?.

  Changing this parameter will create a new resource.

* `id_type` - (Optional, String, ForceNew) Specifies the identifier type of customer gateway.
  The value can be `ip` or `fqdn`. The default value is `ip`.

* `asn` - (Optional, Int, ForceNew) The BGP ASN number of the customer gateway.
  The value ranges from `1` to `4,294,967,295`, the default value is `65,000`.
  Set this parameter to `0` when `id_type` is set to `fqdn`.

  Changing this parameter will create a new resource.

* `tags` - (Optional, Map) Specifies the tags of the customer gateway.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID.

* `created_at` - The create time.

* `updated_at` - The update time.

* `route_mode` - Specifies the routing mode.

* `ip` - Specifies the IP address of the customer gateway.

* `region` - Specifies the region in which resource is created.

## Import

The gateway can be imported using the `id`, e.g.

```bash
$ terraform import opentelekomcloud_enterprise_vpn_customer_gateway_v5.cgw <id>
```
