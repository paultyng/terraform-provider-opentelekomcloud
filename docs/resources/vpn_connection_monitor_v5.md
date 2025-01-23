---
subcategory: "Virtual Private Network (VPN)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_enterprise_vpn_connection_monitor_v5"
sidebar_current: "docs-opentelekomcloud-resource-enterprise-vpn-connection-monitor-v5"
description: |-
  Manages a Enterprise VPN Connection Monitoring Service resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for EVPN connection monitor you can get at
[documentation portal](https://docs.otc.t-systems.com/virtual-private-network/api-ref/api_reference_enterprise_edition_vpn/apis_of_enterprise_edition_vpn/vpn_connection_monitoring/index.html)

# opentelekomcloud_enterprise_vpn_connection_monitor_v5

Manages a VPN connection monitoring resource within OpenTelekomCloud.

## Example Usage

```hcl
variable "connection_id" {}

resource "opentelekomcloud_enterprise_vpn_connection_monitor_v5" "test" {
  connection_id = var.connection_id
}
```

## Argument Reference

The following arguments are supported:

* `connection_id` - (Required, String, ForceNew) Specifies the ID of the VPN connection to monitor.

  Changing this parameter will create a new resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID.

* `source_ip` - The source IP address of the VPN connection.

* `destination_ip` - The destination IP address of the VPN connection.

* `status` - The status of the connection monitor.

* `region` - Specifies the region in which resource is created.

## Import

The monitor can be imported using the `id`, e.g.

```bash
$ terraform import opentelekomcloud_enterprise_vpn_connection_monitor_v5.test <id>
```
