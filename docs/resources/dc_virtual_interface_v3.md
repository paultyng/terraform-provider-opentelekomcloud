---
subcategory: "Direct Connect (DCaaS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dc_virtual_interface_v3"
sidebar_current: "docs-opentelekomcloud-resource-dc-virtual-interface-v3"
description: |-
  Manages a Direct Connect Virtual Interface v3 resource within OpenTelekomCloud.
---


# opentelekomcloud_dc_virtual_interface_v3

Manages a virtual interface v3 resource within OpenTelekomCloud.

-> **NOTE:** Direct Connect v3 API that are used in this resource officially supported only on SwissCloud now.

## Example Usage

```hcl
variable "direct_connect_id" {}
variable "gateway_id" {}
variable "interface_name" {}

resource "opentelekomcloud_dc_virtual_interface_v3" "vi" {
  direct_connect_id = var.direct_connect_id
  vgw_id            = var.gateway_id
  name              = var.interface_name
  type              = "private"
  route_mode        = "static"
  vlan              = 522
  bandwidth         = 5

  remote_ep_group = [
    "1.1.1.0/30",
  ]

  address_family       = "ipv4"
  local_gateway_v4_ip  = "1.1.1.1/30"
  remote_gateway_v4_ip = "1.1.1.2/30"
}
```

## Argument Reference

The following arguments are supported:

* `direct_connect_id` - (Required, String, ForceNew) Specifies the ID of the direct connection associated with the
  virtual interface.
  Changing this will create a new resource.

* `vgw_id` - (Required, String, ForceNew) Specifies the ID of the virtual gateway to which the virtual interface is
  connected.
  Changing this will create a new resource.

* `name` - (Required, String) Specifies the name of the virtual interface.

* `type` - (Required, String, ForceNew) Specifies the type of the virtual interface.
  The valid value is `private`.
  Changing this will create a new resource.

* `route_mode` - (Required, String, ForceNew) Specifies the route mode of the virtual interface.
  The valid values are `static` and `bgp`.
  Changing this will create a new resource.

* `vlan` - (Required, Int, ForceNew) Specifies the customer VLAN to be connected.
  The valid value is range from `0` to `3,999`.
  Changing this will create a new resource.

* `bandwidth` - (Required, Int) Specifies the bandwidth of the virtual interface. The unit is Mbit/s.
  The size range depends on the direct connection.

* `remote_ep_group` - (Required, List) Specifies the CIDR list of remote subnets.
  A CIDR that contains CIDRs of local subnet (corresponding to the parameter `local_gateway_v4_ip` or
  `local_gateway_v6_ip`) and remote subnet (corresponding to the parameter `remote_gateway_v4_ip` or
  `remote_gateway_v6_ip`) must exist in the list.

* `service_ep_group` - (Optional, List) Specifies the subnets that access Internet services through a connection.

* `description` - (Optional, String) Specifies the description of the virtual interface.
  The description contain a maximum of `128` characters and the angle brackets (`<` and `>`) are not allowed.
  Chinese characters must be in `UTF-8` or `Unicode` format.

* `local_gateway_v4_ip` - (Optional, String, ForceNew) Specifies the IPv4 address of the virtual interface in cloud
  side. This parameter is mandatory if `address_family` is set to `ipv4`.
  Changing this will create a new resource.

* `remote_gateway_v4_ip` - (Optional, String, ForceNew) Specifies the IPv4 address of the virtual interface in client
  side. This parameter is mandatory if `address_family` is set to `ipv4`.
  Changing this will create a new resource.

* `address_family` - (Optional, String, ForceNew) Specifies the service type of the virtual interface.
  The valid values are `ipv4` and `ipv6`. Defaults to `ipv4`.
  Changing this will create a new resource.

* `local_gateway_v6_ip` - (Optional, String, ForceNew) Specifies the IPv6 address of the virtual interface in cloud
  side. This parameter is mandatory if `address_family` is set to `ipv6`.
  Changing this will create a new resource.

* `remote_gateway_v6_ip` - (Optional, String, ForceNew) Specifies the IPv6 address of the virtual interface in client
  side. This parameter is mandatory if `address_family` is set to `ipv6`.
  Changing this will create a new resource.

-> The CIDRs of `local_gateway_v4_ip` and `remote_gateway_v4_ip` (or `local_gateway_v6_ip` and `remote_gateway_v6_ip`)
  must be in the same subnet.

* `asn` - (Optional, Int, ForceNew) Specifies the local BGP ASN of the virtual interface.
  The valid value is range from `1` to `4,294,967,295`, except `64,512`.
  Changing this will create a new resource.

* `bgp_md5` - (Optional, String, ForceNew) Specifies the (MD5) password for the local BGP.
  Changing this will create a new resource.

* `enable_bfd` - (Optional, Bool) Specifies whether to enable the Bidirectional Forwarding Detection (BFD) function.
  Defaults to `false`. (This is a reserved parameter and is not supported currently.)

* `enable_nqa` - (Optional, Bool) Specifies whether to enable the Network Quality Analysis (NQA) function.
  Defaults to `false`. (This is a reserved parameter and is not supported currently.)

-> The values of parameter `enable_bfd` and `enable_nqa` cannot be `true` at the same time.

* `lag_id` - (Optional, String, ForceNew) Specifies the ID of the link aggregation group (LAG) associated with the
  virtual interface.
  Changing this will create a new resource.

* `resource_tenant_id` - (Optional, String, ForceNew) Specifies the project ID of another tenant in the same region
  which is used to create virtual interface across tenant. After the across tenant virtual interface is successfully
  created, the target tenant needs to accept the virtual interface request for the virtual interface to take effect.
  Changing this will create a new resource.

  -> 1. When `resource_tenant_id` is specified, `vgw_id` must be the target tenant virtual gateway id.
  <br/>2. When `resource_tenant_id` is specified, the tags can only be configured after the target tenant accepts the
  virtual interface request and the virtual interface takes effect.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual interface.

* `device_id` - The attributed device ID.

* `status` - The current status of the virtual interface.

* `created_at` - The creation time of the virtual interface

* `region` - The region where the virtual interface is located.

* `vif_peers` - The peer information of the virtual interface.
  The [vif_peers](#DCVirtualInterface_vif_peers) structure is documented below.

<a name="DCVirtualInterface_vif_peers"></a>
The `vif_peers` block supports:

* `id` - The VIF peer resource ID.

* `name` - The name of the virtual interface peer.

* `description` - The description of the virtual interface peer.

* `address_family` - The address family type of the virtual interface, which can be `IPv4` or `IPv6`.

* `local_gateway_ip` - The address of the virtual interface peer used on the cloud.

* `remote_gateway_ip` - The address of the virtual interface peer used in the on-premises data center.

* `route_mode` - The routing mode, which can be `static` or `bgp`.

* `bgp_asn` - The ASN of the BGP peer.

* `bgp_md5` - The MD5 password of the BGP peer.

* `device_id` - The ID of the device that the virtual interface peer belongs to.

* `enable_bfd` - Whether to enable BFD.

* `enable_nqa` - Whether to enable NQA.

* `bgp_route_limit` - The BGP route configuration.

* `bgp_status` - The BGP protocol status of the virtual interface peer. If the virtual interface peer uses `static`
  routing, the status is null.

* `status` - The status of the virtual interface peer.

* `vif_id` - The ID of the virtual interface corresponding to the virtual interface peer.

* `receive_route_num` - The number of received BGP routes if `bgp` routing is used. If `static` routing is used,
  this parameter is meaningless and the value is `-1`.

* `remote_ep_group` - The remote subnet list, which records the CIDR blocks used in the on-premises data center.

* `service_ep_group` - The list of public network addresses that can be accessed by the on-premises data center.

## Import

Virtual interfaces can be imported using their `id`, e.g.

```shell
$ terraform import opentelekomcloud_dc_virtual_interface_v3.vi e41748a0-aed9-463e-9817-5c6162265d10
```
