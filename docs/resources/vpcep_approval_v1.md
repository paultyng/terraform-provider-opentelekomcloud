---
subcategory: "VPC Endpoint (VPCEP)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_vpcep_approval_v1"
sidebar_current: "docs-opentelekomcloud-resource-vpcep-approval-v1"
description: |-
Manages a VPCEP Endpoint resource within OpenTelekomCloud.
---


# opentelekomcloud_vpcep_approval_v1

Provides a resource to manage the VPC endpoint connections.

## Example Usage

```hcl
variable "service_vpc_id" {}
variable "vm_port" {}
variable "vpc_id" {}
variable "subnet_id" {}

resource "opentelekomcloud_vpcep_service_v1" "srv" {
  name        = "demo-service"
  server_type = "VM"
  vpc_id      = var.service_vpc_id
  port_id     = var.vm_port

  approval_enabled = true

  port {
    server_port = 8080
    client_port = 80
  }
}

resource "opentelekomcloud_vpcep_endpoint_v1" "ep" {
  service_id = opentelekomcloud_vpcep_service_v1.srv.id
  vpc_id     = var.vpc_id
  subnet_id  = var.subnet_id
  enable_dns = true

  lifecycle {
    # enable_dns and ip_address are not assigned until connecting to the service
    ignore_changes = [
      enable_dns,
      ip_address
    ]
  }
}

resource "opentelekomcloud_vpcep_approval_v1" "approval" {
  service_id = opentelekomcloud_vpcep_service_v1.srv.id
  endpoints  = [opentelekomcloud_vpcep_endpoint_v1.ep.id]
}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required, String, ForceNew) Specifies the ID of the VPC endpoint service. Changing this creates a new
  resource.

* `endpoints` - (Required, List) Specifies the list of VPC endpoint IDs which accepted to connect to VPC endpoint
  service. The VPC endpoints will be rejected when the resource was destroyed.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID in UUID format which equals to the ID of the VPC endpoint service.

* `connections` - An array of VPC endpoints connect to the VPC endpoint service. Structure is documented below.
  + `endpoint_id` - The unique ID of the VPC endpoint.
  + `packet_id` - The packet ID of the VPC endpoint.
  + `domain_id` - The user's domain ID.
  + `status` - The connection status of the VPC endpoint.
  + `description` - The description of the VPC endpoint service connection.

* `region` - The VPC endpoint service region.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 10 minute.
* `delete` - Default is 3 minute.

## Import

VPC endpoint approval can be imported using the `id`, e.g.

```bash
$ terraform import opentelekomcloud_vpcep_approval_v1.apr <id>
```
