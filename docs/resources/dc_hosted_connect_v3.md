---
subcategory: "Direct Connect (DCaaS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dc_hosted_connect_v3"
sidebar_current: "docs-opentelekomcloud-resource-dc-hosted-connect-v3"
description: |-
  Manages a Direct Connect Hosted Connection resource within OpenTelekomCloud.
---

# opentelekomcloud_dc_hosted_connect_v3

Manages a hosted connection resource within OpenTelekomCloud.

-> **NOTE:** Direct Connect v3 API that are used in this resource officially supported only on SwissCloud now.

## Example Usage

```hcl
variable hosting_id {}

data "opentelekomcloud_identity_project_v3" "project" {
  name = "project"
}

resource "opentelekomcloud_dc_hosted_connect_v3" "hc" {
  name               = "hc"
  description        = "create"
  resource_tenant_id = data.opentelekomcloud_identity_project_v3.project.id
  hosting_id         = var.hosting_id
  vlan               = 441
  bandwidth          = 10
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, String) The name of the hosted connect.

* `bandwidth` - (Required, Int) The bandwidth size of the hosted connect in Mbit/s.

* `hosting_id` - (Required, String, ForceNew) The ID of the operations connection on which the hosted connect is created.

  Changing this parameter will create a new resource.

* `vlan` - (Required, Int, ForceNew) The VLAN allocated to the hosted connect.

  Changing this parameter will create a new resource.

* `resource_tenant_id` - (Required, String, ForceNew) The project ID of the specified tenant for whom a hosted connection is to be created.

  Changing this parameter will create a new resource.

* `description` - (Optional, String) The description of the hosted connect.

* `peer_location` - (Optional, String) The location of the on-premises facility at the other end of the connection.
  Specific to the street or data center name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID.

* `region` - Specifies the region in which to create the resource.

* `status` - The status of the hosted connect.
  The options are as follows:
  + **BUILD**: The hosted connect has been created.
  + **ACTIVE**: The associated virtual gateway is normal.
  + **DOWN**: The port used by the hosted connect is down, indicating that there may be line faults.
  + **ERROR**: The associated virtual gateway is abnormal.
  + **PENDING_DELETE**: The hosted connect is being deleted.
  + **PENDING_UPDATE**: The hosted connect is being updated.
  + **PENDING_CREATE**: The hosted connect is being created.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 20 minutes.
* `update` - Default is 20 minutes.
* `delete` - Default is 20 minutes.

## Import

The hosted connect can be imported using the `id`, e.g.

```bash
$ terraform import opentelekomcloud_dc_hosted_connect_v3.hc 6d7bdb34-9254-46ad-b9e0-c7edf7abf8bc
```
