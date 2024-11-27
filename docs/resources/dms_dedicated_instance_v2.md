---
subcategory: "Distributed Message Service (DMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dms_dedicated_instance_v2"
sidebar_current: "docs-opentelekomcloud-resource-dms-dedicated-instance-v2"
description: |-
  Manages an up-to-date DMS Instance v2 resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DMS instance you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-message-service/api-ref/apis_v2_recommended/lifecycle_management)

# opentelekomcloud_dms_dedicated_instance_v2

Manages a DMS instance in the OpenTelekomCloud DMS Service (Kafka Premium/Platinum).
## Example Usage

### Create a cluster Kafka instance

```hcl
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}
variable "access_password" {}

variable "flavor_id" {
  default = "c6.2u4g.cluster"
}
variable "storage_spec_code" {
  default = "dms.physical.storage.ultra.v2"
}

data "opentelekomcloud_dms_flavor_v2" "test" {
  type              = "cluster"
  flavor_id         = var.flavor_id
  storage_spec_code = var.storage_spec_code
}

resource "opentelekomcloud_dms_dedicated_instance_v2" "test" {
  name              = "kafka_test"
  vpc_id            = var.vpc_id
  network_id        = var.subnet_id
  security_group_id = var.security_group_id

  flavor_id         = data.opentelekomcloud_dms_flavor_v2.test.flavor_id
  storage_spec_code = data.opentelekomcloud_dms_flavor_v2.test.flavors[0].ios[0].storage_spec_code
  available_zones   = [data.opentelekomcloud_dms_az_v1.az_1.id]
  engine_version    = "2.7"
  storage_space     = 600
  broker_num        = 3

  ssl_enable  = true
  access_user = "user"
  password    = var.access_password
}
```

### Create a single-node Kafka instance with floating ip

```hcl
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}
variable "access_password" {}

data "opentelekomcloud_dms_az_v1" "az_1" {}

data "opentelekomcloud_dms_flavor_v2" "test" {
  type      = "single"
  flavor_id = "s6.2u4g.single.small"
}

locals {
  flavor = data.opentelekomcloud_dms_flavor_v2.test.flavors[0]
}

resource "opentelekomcloud_networking_floatingip_v2" "fip_1" {
}

resource "opentelekomcloud_dms_dedicated_instance_v2" "instance_1" {
  name = "kafka-single-node"

  vpc_id            = var.vpc_ip
  network_id        = var.subnet_id
  security_group_id = var.security_group_id

  flavor_id         = local.flavor.id
  storage_spec_code = local.flavor.ios[0].storage_spec_code
  available_zones   = [data.opentelekomcloud_dms_az_v1.az_1.id]
  engine_version    = "2.7"
  storage_space     = 300
  broker_num        = 1

  ssl_enable       = true
  access_user      = "user"
  password         = var.access_password
  maintain_begin   = "02:00"
  maintain_end     = "06:00"
  retention_policy = "time_base"

  enable_publicip = true
  publicip_id     = [opentelekomcloud_networking_floatingip_v2.fip_1.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) Specifies the name of the DMS Kafka instance. An instance name starts with a letter,
  consists of 4 to 64 characters, and supports only letters, digits, hyphens (-) and underscores (_).

* `description` - (Optional, String) Specifies the description of the DMS Kafka instance. It is a character string
  containing not more than 1,024 characters.

* `flavor_id` - (Required, String, ForceNew) Specifies the Kafka [flavor ID](https://docs.otc.t-systems.com/distributed-message-service/api-ref/apis_v2_recommended/other_apis/querying_product_specifications_list.html#listengineproducts,
  e.g. **c6.2u4g.cluster**.

* `engine_version` - (Required, String, ForceNew) Specifies the version of the Kafka engine,
  such as 1.1.0, 2.3.0, 2.7 or other supported versions. Changing this creates a new instance resource.

* `storage_spec_code` - (Required, String, ForceNew) Specifies the storage I/O specification.
  The valid values are as follows:
  + **dms.physical.storage.high.v2**: Type of the disk that uses high I/O.
  + **dms.physical.storage.ultra.v2**: Type of the disk that uses ultra-high I/O.

* `vpc_id` - (Required, String, ForceNew) Specifies the ID of a VPC. Changing this creates a new instance resource.

* `network_id` - (Required, String, ForceNew) Specifies the ID of a subnet. Changing this creates a new instance
  resource.

* `security_group_id` - (Required, String) Specifies the ID of a security group.

* `available_zones` - (Optional, List, ForceNew) Indicates the ID of an AZ. The parameter value can not be
  left blank or an empty array. For details, see section
  [Querying AZ Information](https://docs.otc.t-systems.com/distributed-message-service/api-ref/apis_v2_recommended/other_apis/listing_az_information.html#listavailablezones).

* `ipv6_enable` - (Optional, Bool, ForceNew) Specifies whether to enable IPv6. Defaults to **false**.
  Changing this creates a new instance resource.

* `arch_type` - (Optional, String, ForceNew) Specifies the CPU architecture. Valid value is **X86**.
  Changing this creates a new instance resource.

* `enable_publicip` - (Optional, Bool, ForceNew) Whether to enable public access. By default, public access is disabled.

* `enable_publicip` - (Optional, List, ForceNew) A list of IDs of the EIP bound to the instance.
This parameter is mandatory if `enable_publicip` is set to `true`.

* `disk_encrypted_enable` - (Optional, Bool, ForceNew) Indicates whether to enable disk encryption.

* `disk_encrypted_key` - (Optional, String, ForceNew) Disk encryption key. If disk encryption is not enabled, this parameter is left blank.

* `storage_space` - (Required, Int) Specifies the message storage capacity, the unit is GB.
  The storage spaces corresponding to the product IDs are as follows:
  + **c6.2u4g.cluster** (100MB bandwidth): `300` to `300,000` GB
  + **c6.4u8g.cluster** (300MB bandwidth): `300` to `600,000` GB
  + **c6.8u16g.cluster** (600MB bandwidth): `300` to `900,000` GB
  + **c6.12u12g.cluster**: `300` to `900,000` GB
  + **c6.16u32g.cluster** (1,200MB bandwidth): `300` to `900,000` GB

* `broker_num` - (Required, Int) Specifies the broker numbers.

* `new_tenant_ips` - (Optional, List) Specifies the IPv4 private IP addresses for the new brokers.

  -> The number of specified IP addresses must be less than or equal to the number of new brokers.

* `access_user` - (Optional, String, ForceNew) Specifies the username of SASL_SSL user. A username consists of 4
  to 64 characters and supports only letters, digits, and hyphens (-). Changing this creates a new instance resource.

* `password` - (Optional, String) Specifies the password of SASL_SSL user. A password must meet the following
  complexity requirements: Must be 8 to 32 characters long. Must contain at least 2 of the following character types:
  lowercase letters, uppercase letters, digits, and special characters (`~!@#$%^&*()-_=+\\|[{}]:'",<.>/?).

  -> **NOTE:** `access_user` and `password` is mandatory and available when `ssl_enable` is **true**.

* `security_protocol` - (Optional, String, ForceNew) Specifies the protocol to use after SASL is enabled. Value options:
  + **SASL_SSL**: Data is encrypted with SSL certificates for high-security transmission.
  + **SASL_PLAINTEXT**: Data is transmitted in plaintext with username and password authentication. This protocol only
    uses the SCRAM-SHA-512 mechanism and delivers high performance.

  Defaults to **SASL_SSL**. Changing this creates a new instance resource.

* `enabled_mechanisms` - (Optional, List, ForceNew) Specifies the authentication mechanisms to use after SASL is
  enabled. Value options:
  + **PLAIN**: Simple username and password verification.
  + **SCRAM-SHA-512**: User credential verification, which is more secure than **PLAIN**.

  Defaults to [**PLAIN**]. Changing this creates a new instance resource.

* `maintain_begin` - (Optional, String) Specifies the time at which a maintenance time window starts. Format: HH:mm. The
  start time and end time of a maintenance time window must indicate the time segment of a supported maintenance time
  window. The start time must be set to 22:00, 02:00, 06:00, 10:00, 14:00, or 18:00. Parameters `maintain_begin`
  and `maintain_end` must be set in pairs. If parameter `maintain_begin` is left blank, parameter `maintain_end` is also
  blank. In this case, the system automatically allocates the default start time 02:00.

* `maintain_end` - (Optional, String) Specifies the time at which a maintenance time window ends. Format: HH:mm. The
  start time and end time of a maintenance time window must indicate the time segment of a supported maintenance time
  window. The end time is four hours later than the start time. For example, if the start time is 22:00, the end time is
  02:00. Parameters `maintain_begin`
  and `maintain_end` must be set in pairs. If parameter `maintain_end` is left blank, parameter
  `maintain_begin` is also blank. In this case, the system automatically allocates the default end time 06:00.

* `retention_policy` - (Optional, String) Specifies the action to be taken when the memory usage reaches the disk
  capacity threshold. The valid values are as follows:
  + **time_base**: Automatically delete the earliest messages.
  + **produce_reject**: Stop producing new messages.

* `ssl_enable` - (Optional, Bool, ForceNew) Specifies whether the Kafka SASL_SSL is enabled.
  Changing this creates a new resource.

* `tags` - (Optional, Map) The key/value pairs to associate with the DMS Kafka instance.

* `cross_vpc_accesses` - (Optional, List) Specifies the cross-VPC access information.
  The [object](#dms_cross_vpc_accesses) structure is documented below.

<a name="dms_cross_vpc_accesses"></a>
The `cross_vpc_accesses` block supports:

* `advertised_ip` - (Optional, String) The advertised IP Address or domain name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Specifies a resource ID in UUID format.
* `engine` - Indicates the message engine.
* `partition_num` - Indicates the number of partitions in Kafka instance.
* `used_storage_space` - Indicates the used message storage space. Unit: GB
* `port` - Indicates the port number of the DMS Kafka instance.
* `status` - Indicates the status of the DMS Kafka instance.
* `resource_spec_code` - Indicates a resource specifications identifier.
* `type` - Indicates the DMS Kafka instance type.
* `user_id` - Indicates the ID of the user who created the DMS Kafka instance
* `user_name` - Indicates the name of the user who created the DMS Kafka instance
* `connect_address` - Indicates the IP address of the DMS Kafka instance.
* `cross_vpc_accesses` - Indicates the Access information of cross-VPC. The structure is documented below.
* `public_ip_address` - Indicates the public IP addresses list of the instance.
* `connector_node_num` - Indicates the number of connector node.
* `storage_resource_id` - Indicates the storage resource ID.
* `storage_type` - Indicates the storage type.
* `created_at` - Indicates the create time.
* `cert_replaced` - Indicates whether the certificate can be replaced.
* `node_num` - Indicates the node quantity.
* `pod_connect_address` - Indicates the connection address on the tenant side.
* `public_bandwidth` - Indicates the public network access bandwidth.
* `ssl_two_way_enable` - Indicates whether to enable two-way authentication.
* `dumping` - Whether message dumping(smart connect) is enabled.
* `region` - The region in which DMS Kafka instance is created.

The `cross_vpc_accesses` block supports:

* `listener_ip` - The listener IP address.
* `port` - The port number.
* `port_id` - The port ID associated with the address.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 50 minutes.
* `update` - Default is 50 minutes.
* `delete` - Default is 15 minutes.

## Import

DMS Kafka instance can be imported using the instance id, e.g.

```
 $ terraform import opentelekomcloud_dms_dedicated_instance_v2.instance_1 8d3c7938-dc47-4937-a30f-c80de381c5e3
```

Note that the imported state may not be identical to your resource definition, due to some attributes missing from the
API response, security or some other reason. The missing attributes include:
`password`, `manager_password`, `public_ip_ids`, `security_protocol`, `enabled_mechanisms` and `arch_type`.
It is generally recommended running `terraform plan` after importing
a DMS Kafka instance. You can then decide if changes should be applied to the instance, or the resource definition
should be updated to align with the instance. Also, you can ignore changes as below.

```hcl
resource "opentelekomcloud_dms_dedicated_instance_v2" "instance_1" {
  lifecycle {
    ignore_changes = [
      "password",
      "used_storage_space",
      "cross_vpc_accesses",
      "publicip_id",
    ]
  }
}
```
