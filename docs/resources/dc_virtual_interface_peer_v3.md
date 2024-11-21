---
subcategory: "Direct Connect (DCaaS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dc_virtual_interface_peer_v3"
sidebar_current: "docs-opentelekomcloud-resource-dc-virtual-interface-peer-v3"
description: |-
  Manages a Direct Connect Virtual Interface Peer v3 resource within OpenTelekomCloud.
---


# opentelekomcloud_dc_virtual_interface_peer_v3

Manages a virtual interface peer v3 resource within OpenTelekomCloud.

-> **NOTE:** Direct Connect v3 API that are used in this resource officially supported only on SwissCloud now.

## Example Usage

```hcl
variable "virtual_interface_id" {}

resource "opentelekomcloud_dc_virtual_interface_peer_v3" "vp" {
  vif_id            = var.virtual_interface_id
  name              = "my_peer"
  address_family    = "ipv6"
  route_mode        = "static"
  remote_ep_group   = ["fd00:0:0:0:0:0:0:0/64"]
  description       = "ipv6 peer"
  local_gateway_ip  = "FD00::1/64"
  remote_gateway_ip = "FD00::2/64"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) Specifies the name of the virtual interface peer.

* `description` - (Required, String) Provides supplementary information about the virtual interface peer.

* `address_family` - (Required, String, ForceNew) The address family type of the virtual interface, which can be `IPv4` or `IPv6`.

* `local_gateway_ip` - (Required, String, ForceNew) The address of the virtual interface peer used on the cloud.

* `remote_gateway_ip` - (Required, String, ForceNew) The address of the virtual interface peer used in the on-premises data center.

* `route_mode` - (Optional, String, ForceNew) The routing mode, which can be `static` or `bgp`.

* `bgp_asn` - (Optional, String, ForceNew) The ASN of the BGP peer.

* `bgp_md5` - (Optional, String, ForceNew) The MD5 password of the BGP peer.

* `remote_ep_group` - (Optional, List) The remote subnet list, which records the CIDR blocks used in the on-premises data center.

* `vif_id` - (Required, String, ForceNew) Specifies the ID of the virtual interface corresponding to the virtual interface peer.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The VIF peer resource ID.

* `region` - The region where the virtual interface is located.

* `project_id` - The project where the virtual interface is located.

* `device_id` - The ID of the device that the virtual interface peer belongs to.

* `enable_bfd` - BFD status.

* `enable_nqa` - NQA status.

* `bgp_route_limit` - The BGP route configuration.

* `bgp_status` - The BGP protocol status of the virtual interface peer.

* `status` - The status of the virtual interface peer.

* `receive_route_num` - The number of received BGP routes if `bgp` routing is used.

* `service_ep_group` - The list of public network addresses that can be accessed by the on-premises data center.

## Import

Virtual interface peers can be imported using their `id`, e.g.

```shell
$ terraform import opentelekomcloud_dc_virtual_interface_peer_v3.vi e41748a0-aed9-463e-9817-5c6162265d11
```
