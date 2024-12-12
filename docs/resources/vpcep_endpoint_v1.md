---
subcategory: "VPC Endpoint (VPCEP)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_vpcep_endpoint_v1"
sidebar_current: "docs-opentelekomcloud-resource-vpcep-endpoint-v1"
description: |-
  Manages a VPCEP Endpoint resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for VPCEP you can get at
[documentation portal](https://docs.otc.t-systems.com/vpc-endpoint/api-ref/apis/apis_for_managing_vpc_endpoints)

# opentelekomcloud_vpcep_endpoint_v1

Manages a VPC Endpoint v1 resource within OpenTelekomCloud.

## Example Usage

```hcl
data "opentelekomcloud_vpc_subnet_v1" "shared_subnet" {
  name = "test-subnet"
}

resource "opentelekomcloud_lb_loadbalancer_v2" "lb_1" {
  vip_subnet_id = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.subnet_id
}

resource "opentelekomcloud_vpcep_service_v1" "service" {
  name        = "service_1"
  port_id     = opentelekomcloud_lb_loadbalancer_v2.lb_1.vip_port_id
  vpc_id      = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  server_type = "LB"

  port {
    client_port = 80
    server_port = 8080
  }

  tags = {
    "key" : "value",
  }
}

resource "opentelekomcloud_vpcep_endpoint_v1" "endpoint" {
  service_id = opentelekomcloud_vpcep_service_v1.service.id
  vpc_id     = opentelekomcloud_vpcep_service_v1.service.vpc_id
  subnet_id  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.id
  port_ip    = "192.168.0.12"
  enable_dns = true
  whitelist = [
    "127.0.0.1"
  ]

  tags = {
    "fizz" : "buzz"
  }
}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required, String, ForceNew) Specifies the ID of the VPC endpoint service.

* `vpc_id` - (Required, String, ForceNew) Specifies the ID of the VPC (OpenStack router) where the VPC endpoint is to be created.

* `subnet_id` - (Optional, String, ForceNew) The value must be the ID of the subnet (OpenStack network) created in the VPC specified
  by `vpc_id` and in the format of the UUID.
  This parameter is mandatory only if you create a VPC endpoint for connecting to an interface VPC endpoint service.

~>
The CIDR block of the VPC subnet cannot overlap with `198.19.128.0/20`. The destination address of the custom route in
the VPC route table cannot overlap with the CIDR block `198.19.128.0/20`.

* `enable_dns` - (Optional, Bool, ForceNew) Specifies whether to create a private domain name. The default value is `false`.

* `description` - (Optional, String, ForceNew) Specifies the description of the VPC endpoint. The value can contain
  characters such as letters and digits, but cannot contain less than signs (<) and great than signs (>).

* `route_tables` - (Optional, List, ForceNew) Lists the IDs of route tables.

* `port_ip` - (Optional, String, ForceNew) Specifies the IP address for accessing the associated VPC endpoint service.

* `whitelist` - (Optional, List, ForceNew) Specifies an array of whitelisted IPs for controlling access to the VPC endpoint.
  ``IPv4 addresses`` or ``CIDR blocks`` can be specified to control access when you create a VPC endpoint.
  This parameter is mandatory only when you create a ``VPC endpoint`` for connecting to an interface VPC endpoint service.

* `enable_whitelist` - (Optional, Bool, ForceNew) Specifies whether to enable access control.
  This parameter is available only if you create a ``VPC endpoint`` for connecting to an interface VPC endpoint service.

* `tags` - (Optional, Map) The key/value pairs to associate with the VPC endpoint.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of VPC endpoint.

* `marker_id` - Specifies the packet ID of the VPC endpoint.

* `service_name` - Specifies the name of the VPC endpoint service.

* `service_type` - Specifies the type of the VPC endpoint service that is associated with the VPC endpoint.

* `dns_names` - Specifies the domain name for accessing the associated VPC endpoint service.
  This parameter is only available when `enable_dns` is set to `true`.

* `project_id` - Specifies the project ID.

* `status` - The status of the VPC endpoint. The value can be `pendingAcceptance`, `creating`, `accepted`,
    `rejected`, `failed`, `deleting`.

## Import

VPC endpoint can be imported using the `id`, e.g.

```sh
terraform import opentelekomcloud_vpcep_endpoint_v1.endpoint 71ba78a2-d847-4882-8fd0-42c5854c1cbc
```
