---
subcategory: "Distributed Database Middleware (DDM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ddm_instance_v1"
sidebar_current: "docs-opentelekomcloud-resource-ddm-instance-v1"
description: |-
  Manages a DDM Instance resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DDS instance you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-database-middleware/api-ref/apis_recommended/ddm_instances)

# opentelekomcloud_ddm_instance_v1

Manages DDM instance resource within OpenTelekomCloud

## Example Usage: Creating a basic DDM instance with 2 nodes
```hcl
variable "flavor_id" {}
variable "engine_id" {}
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name                = "ddm-instance"
  availability_zones  = ["eu-de-01", "eu-de-02", "eu-de-03"]
  flavor_id           = var.flavor_id
  node_num            = 2
  engine_id           = var.engine_id
  vpc_id              = var.vpc_id
  subnet_id           = var.subnet_id
  security_group_id   = var.security_group.id
  purge_rds_on_delete = true
}
```

## Example Usage: Creating a DDM instance with custom credentials
```hcl
variable "flavor_id" {}
variable "engine_id" {}
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}
variable "username" {}
variable "password" {}

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name                = "ddm-instance"
  availability_zones  = ["eu-de-01", "eu-de-02", "eu-de-03"]
  flavor_id           = var.flavor_id
  node_num            = 2
  engine_id           = var.engine_id
  vpc_id              = var.vpc_id
  subnet_id           = var.subnet_id
  security_group_id   = var.security_group.id
  purge_rds_on_delete = true
  username            = var.username
  password            = var.password
}
```

## Example Usage: Creating a DDM instance with custom time zone
```hcl
variable "flavor_id" {}
variable "engine_id" {}
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}

resource "opentelekomcloud_ddm_instance_v1" "instance_1" {
  name                = "ddm-instance"
  availability_zones  = ["eu-de-01", "eu-de-02", "eu-de-03"]
  flavor_id           = var.flavor_id
  node_num            = 2
  engine_id           = var.engine_id
  vpc_id              = var.vpc_id
  subnet_id           = var.subnet_id
  security_group_id   = var.security_group.id
  purge_rds_on_delete = true
  time_zone           = "UTC+01:00"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, String, ForceNew) Specifies the region of the DDM instance.

* `name` - (Required, String) Specifies the DDM instance name. The DDM instance name of the same
  type is unique in the same tenant. It can be  4 to 64 characters long. It must start with a letter and it can only contain etters, digits, and hyphens (-).

* `availability_zones` - (Required, List, ForceNew) Specifies the list of availability zones.

* `vpc_id` - (Required, String, ForceNew) Specifies the VPC ID.

* `subnet_id` - (Required, String, ForceNew) Specifies the subnet Network ID.

* `security_group_id` - (Required, String) Specifies the security group ID of the DDM instance.

* `node_num` - (Required, Integer) Specifies the disk encryption ID of the instance.

* `flavor_id` - (Required, String, ForceNew) Specifies the flavor ID of the instance nodes.

* `engine_id` - (Required, String, ForceNew) Specifies the Engine ID of the instance.

* `time_zone` - (Optional, String, ForceNew) Specifies the timezone. Valid formats are `UTC+12:00`, `UTC+11:00`, ... ,`UTC+01:00`, `UTC`, `UTC-01:00`, ... , `UTC-11:00`, `UTC-12:00`

* `username` - (Optional, String, ForceNew) Specifies the Administrator username of the DDM instance. It can be 1 to 32 characters long and can contain letters, digits, and underscores (_). It must start with a letter.

* `password` - (Optional, String) Specifies the Administrator password of the DDM instance. it can be 8 to 32 characters long. It must be a combination of uppercase letters, lowercase letters, digits, and the following special characters: `~ ! @ # % ^ * - _ = + ?`. It must be a strong password to improve security and prevent security risks such as brute force cracking.

* `param_group_id` - (Optional, String, ForceNew) Specifies the parameters group ID.

* `purge_rds_on_delete` - (Optional, Boolean) Specifies whether data stored on the associated DB instances is deleted. The value can be: `true` or `false` (default).


## Attributes Reference

The following attributes are exported:

* `region` - See Argument Reference above.
* `name` - See Argument Reference above.
* `availability_zones` - See Argument Reference above.
* `vpc_id` - See Argument Reference above.
* `subnet_id` - See Argument Reference above.
* `security_group_id` - See Argument Reference above.
* `node_num` - See Argument Reference above.
* `flavor_id` - See Argument Reference above.
* `engine_id` - See Argument Reference above.
* `time_zone` - See Argument Reference above.
* `username` - See Argument Reference above.
* `password` - See Argument Reference above.
* `param_group_id` - See Argument Reference above.
* `purge_rds_on_delete` - See Argument Reference above.
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


## Import

DDMv1 Instance can be imported using the `id`, e.g.

```shell
terraform import opentelekomcloud_ddm_instance_v1.instance_1 c1851195-cdcb-4d23-96cb-032e6a3ee667
```

Following attributes are not properly imported.
* `availability_zones`
* `flavor_id`
* `engine_id`
* `time_zone`
* `password`
* `param_group_id`
* `purge_rds_on_delete`
