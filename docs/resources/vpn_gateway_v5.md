---
subcategory: "Virtual Private Network (VPN)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_enterprise_vpn_gateway_v5"
sidebar_current: "docs-opentelekomcloud-resource-enterprise-vpn-gateway-v5"
description: |-
  Manages a Enterprise VPN Gateway Service resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for EVPN you can get at
[documentation portal](https://docs.otc.t-systems.com/virtual-private-network/api-ref/api_reference_enterprise_edition_vpn/apis_of_enterprise_edition_vpn/vpn_gateway/index.html)


# opentelekomcloud_enterprise_vpn_gateway_v5

Manages a VPN gateway resource within OpenTelekomCloud.

## Example Usage

### Basic Usage

```hcl
variable "name" {}

resource "opentelekomcloud_enterprise_vpn_gateway_v5" "gw_1" {
  name           = var.name
  vpc_id         = opentelekomcloud_vpc_v1.vpc.id
  local_subnets  = [opentelekomcloud_vpc_subnet_v1.subnet.cidr]
  connect_subnet = opentelekomcloud_vpc_subnet_v1.subnet.id

  availability_zones = [
    "eu-de-01",
    "eu-de-02"
  ]

  eip1 {
    id = opentelekomcloud_vpc_eip_v1.eip_1.id
  }

  eip2 {
    id = opentelekomcloud_vpc_eip_v1.eip_2.id
  }

  tags = {
    key = "val"
    foo = "bar"
  }
}
```

### Creating a VPN gateway with creating new EIPs

```hcl
variable "name" {}

resource "opentelekomcloud_enterprise_vpn_gateway_v5" "gw_1" {
  name           = var.name
  ha_mode        = "active-standby"
  vpc_id         = opentelekomcloud_vpc_v1.vpc.id
  local_subnets  = [opentelekomcloud_vpc_subnet_v1.subnet.cidr]
  connect_subnet = opentelekomcloud_vpc_subnet_v1.subnet.id

  availability_zones = [
    "eu-de-01",
    "eu-de-02"
  ]

  eip1 {
    bandwidth_name = "evpn-gw-bw-1"
    type           = "5_bgp"
    bandwidth_size = 5
    charge_mode    = "traffic"
  }

  eip2 {
    bandwidth_name = "evpn-gw-bw-2"
    type           = "5_bgp"
    bandwidth_size = 5
    charge_mode    = "traffic"
  }
}

```

### Creating a private VPN gateway with Enterprise Router

```hcl
variable "name" {}
variable "er_id" {}

resource "opentelekomcloud_enterprise_vpn_gateway_v5" "gw_1" {
  name            = var.name
  network_type    = "private"
  attachment_type = "er"
  er_id           = var.er_id

  availability_zones = [
    "eu-de-01",
    "eu-de-02"
  ]

  access_vpc_id    = opentelekomcloud_vpc_v1.vpc_er.id
  access_subnet_id = opentelekomcloud_vpc_subnet_v1.subnet_er.id

  access_private_ip_1 = "172.16.0.99"
  access_private_ip_2 = "172.16.0.100"
}
```

## Argument Reference

The following arguments are supported:
* `name` - (Required, String) The name of the VPN gateway.
  The valid length is limited from `1` to `64`, only letters, digits, hyphens (-) and underscores (_) are allowed.

* `availability_zones` - (Required, List, ForceNew) The list of availability zone IDs.
  Changing this parameter will create a new resource.

* `flavor` - (Optional, String, ForceNew) The flavor of the VPN gateway.
  The value can be `Basic`, `Professional1`, `Professional2`. Defaults to `Professional1`.
  Changing this parameter will create a new resource.

* `attachment_type` - (Optional, String, ForceNew) The attachment type. The value can be `vpc` and `er`.
  Defaults to `vpc`.
  Changing this parameter will create a new resource.

* `network_type` - (Optional, String, ForceNew) The network type. The value can be `public` and `private`.
  Defaults to `public`.
  Changing this parameter will create a new resource.

* `vpc_id` - (Optional, String, ForceNew) The ID of the VPC to which the VPN gateway is connected.
  This parameter is mandatory when `attachment_type` is `vpc`.
  Changing this parameter will create a new resource.

* `local_subnets` - (Optional, List) The list of local subnets.
  This parameter is mandatory when `attachment_type` is `vpc`.

* `connect_subnet` - (Optional, String, ForceNew) The Network ID of the VPC subnet used by the VPN gateway.
  This parameter is mandatory when `attachment_type` is `vpc`.
  Changing this parameter will create a new resource.

* `er_id` - (Optional, String, ForceNew) The enterprise router ID to attach with to VPN gateway.
  This parameter is mandatory when `attachment_type` is `er`.
  Changing this parameter will create a new resource.

* `ha_mode` - (Optional, String, ForceNew) The HA mode of VPN gateway. Valid values are `active-active` and
  `active-standby`. The default value is `active-active`.
  Changing this parameter will create a new resource.

* `eip1` - (Optional, List) The master 1 IP in active-active VPN gateway or the master IP
  in active-standby VPN gateway. This parameter is mandatory when `network_type` is `public` or left empty.
  The [object](#GwCreateRequestEip) structure is documented below.

* `eip2` - (Optional, List, ForceNew) The master 2 IP in active-active VPN gateway or the slave IP
  in active-standby VPN gateway. This parameter is mandatory when `network_type` is **public** or left empty.
  The [object](#GwCreateRequestEip) structure is documented below.

* `access_vpc_id` - (Optional, String, ForceNew) The access VPC ID.
  The default value is the value of `vpc_id`.
  Changing this parameter will create a new resource.

* `access_subnet_id` - (Optional, String, ForceNew) The access subnet ID.
  The default value is the value of `connect_subnet`.
  Changing this parameter will create a new resource.

* `access_private_ip_1` - (Optional, String, ForceNew) The private IP 1 in private network type VPN gateway.
  It is the master IP 1 in `active-active` HA mode, and the master IP in `active-standby` HA mode.
  Must declare the `access_private_ip_2` at the same time, and can not use the same IP value.
  Changing this parameter will create a new resource.

* `access_private_ip_2` - (Optional, String, ForceNew) The private IP 2 in private network type VPN gateway.
  It is the master IP 2 in `active-active` HA mode, and the slave IP in `active-standby` HA mode.
  Must declare the `access_private_ip_1` at the same time, and can not use the same IP value.
  Changing this parameter will create a new resource.

* `asn` - (Optional, Int, ForceNew) The ASN number of BGP. The value ranges from `1` to `4,294,967,295`.
  Defaults to `64,512`.
  Changing this parameter will create a new resource.

<a name="GwCreateRequestEip"></a>
The `eip1` or `eip2` block supports:

* `id` - (Optional, String, ForceNew) The public IP ID.
  Changing this parameter will create a new resource.

* `type` - (Optional, String, ForceNew) The EIP type.
  Changing this parameter will create a new resource.

* `bandwidth_name` - (Optional, String, ForceNew) The bandwidth name.
  The valid length is limited from `1` to `64`, only letters, digits, hyphens (-) and underscores (_) are allowed.
  Changing this parameter will create a new resource.

* `bandwidth_size` - (Optional, Int, ForceNew) Bandwidth size in Mbit/s. When the `flavor` is `Basic`, the value
  cannot be greater than `100`. When the `flavor` is `Professional1`, the value cannot be greater than `300`.
  When the `flavor` is `Professional2`, the value cannot be greater than `1,000`.
  Changing this parameter will create a new resource.

* `charge_mode` - (Optional, String, ForceNew) The charge mode of the bandwidth. The value can be `bandwidth` and `traffic`.
  Changing this parameter will create a new resource.

  ~> You can use `id` to specify an existing EIP or use `type`, `bandwidth_name`, `bandwidth_size` and `charge_mode` to
    create a new EIP.

* `tags` - (Optional, Map) Specifies the tags of the VPN gateway.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPN gateway

* `status` - The status of VPN gateway.

* `created_at` - The create time.

* `updated_at` - The update time.

* `used_connection_group` - The number of used connection groups.

* `used_connection_number` - The number of used connections.

* `er_attachment_id` - The ER attachment ID.

* `region` - Specifies the region in which resource is created.

* `eip1` - The master 1 IP in active-active VPN gateway or the master IP in active-standby VPN gateway.
  The [object](#GatewayGetResponseEip) structure is documented below.

* `eip2` - The master 2 IP in active-active VPN gateway or the slave IP in active-standby VPN gateway.
  The [object](#GatewayGetResponseEip) structure is documented below.

<a name="GatewayGetResponseEip"></a>
The `eip1` or `eip2` block supports:

* `id` - The public IP ID.

* `bandwidth_id` - The bandwidth ID.

* `ip_address` - The public IP address.

* `ip_version` - Specifies the EIP version.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 10 minutes.
* `update` - Default is 10 minutes.
* `delete` - Default is 10 minutes.

## Import

The gateway can be imported using the `id`, e.g.

```bash
$ terraform import opentelekomcloud_enterprise_vpn_gateway_v5.test <id>
```
