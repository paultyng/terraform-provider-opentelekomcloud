---
subcategory: "Distributed Database Middleware (DDM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ddm_instance_v1"
sidebar_current: "docs-opentelekomcloud-datasource-ddm-instance-v1"
description: |-
  Get a DDM Instance resource from OpenTelekomCloud.
---

Up-to-date reference of API arguments for DDM instance you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-database-middleware/api-ref/apis_recommended/ddm_instances/querying_details_of_a_ddm_instance.html)

# opentelekomcloud_ddm_instance_v1

Use this data source to get info of the OpenTelekomCloud DDM instance.

## Example Usage
```hcl
variable "instance_id" {}

data "opentelekomcloud_ddm_instance_v1" "instance" {
  instance_id = var.instance_id
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String) Specifies the DDM instance ID.

## Attributes Reference

The following attributes are exported:

* `instance_id` - See Argument Reference above.
* `region` - Indicates the region of the DDM instance.
* `name` - Indicates the name of DDM instance.
* `vpc_id` - Indicates the VPC ID.
* `subnet_id` - Indicates the subnet Network ID.
* `security_group_id` - Indicates the security group ID of the DDM instance.
* `node_num` - Indicates the disk encryption ID of the instance.
* `username` - Indicates the Administrator username of the DDM instance.
* `status` - Indicates the DDM instance status.
* `access_IP` - Indicates the DDM access IP.
* `access_port` - Indicates the DDM access port.
* `created_at` - Indicates the creation time.
* `updated_at` - Indicates the update time.
* `availability_zone` - Indicates the availability zone of DDM instance.
* `node_status` - Indicates the DDM nodes status.
* `nodes` - Indicates the instance nodes information. Structure is documented below.

The `nodes` block contains:

  - `ip` - Indicates the node IP.
  - `port` - Indicates the node port.
  - `status` - Indicates the node status.
