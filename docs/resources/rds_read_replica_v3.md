---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_read_replica_v3"
sidebar_current: "docs-opentelekomcloud-resource-rds-read-replica-v3"
description: |-
  Manages an RDS Read Replica resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RDS replica you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/db_instance_management)

# opentelekomcloud_rds_read_replica_v3

Manages a RDSv3 read replica resource within OpenTelekomCloud.

## Example Usage

```hcl
resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "test-instance"
  availability_zone = var.az_main
  db {
    password = var.db_password
    type     = "PostgreSQL"
    version  = "16"
    port     = "8635"
  }
  security_group_id = var.sg_id
  subnet_id         = var.os_network_id
  vpc_id            = var.os_router_id
  flavor            = "rds.pg.n1.medium.4.rr"

  volume {
    type = "CLOUDSSD"
    size = 40
  }
  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }

  tag = {
    created = "terraform"
  }
}

resource "opentelekomcloud_rds_read_replica_v3" "replica" {
  name          = "test-replica"
  replica_of_id = opentelekomcloud_rds_instance_v3.instance.id
  flavor_ref    = "${opentelekomcloud_rds_instance_v3.instance.flavor}.rr"

  availability_zone = var.az_replica

  volume {
    type = "CLOUDSSD"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - Specifies the DB replica instance name. The DB instance name of the same type must be unique for the same
  tenant. The value must be `4` to `64` characters in length and start with a letter. It is case-sensitive and can
  contain only letters, digits, hyphens (`-`), and underscores  (`_`). Changing this parameter will create a new
  resource.

* `replica_of_id` - Specifies ID of the replicated instance. Changing this parameter will create a new resource.

* `flavor` - Specifies the specification code. Read replica flavors ends with `.rr`.

* `region` - (Optional) Specifies the region of the replica instance. Changing this parameter will create a new
  resource.

* `public_ips` - (Optional) Specifies floating IP to be assigned to the instance.
  This should be a list with single element only.

* `ssl_enable` - (Optional) Specifies whether SSL should be enabled for MySql instances.

* `volume` - Specifies the volume information. Structure is documented below.

The `volume` block supports:

* `disk_encryption_id` - (Optional) Specifies the key ID for disk encryption. Changing this parameter will create a new
  resource.

* `type` - Specifies the volume type. Changing this parameter will create a new resource. Its value can be any of the
  following and is case-sensitive.
    * `COMMON`: indicates the SATA type.
    * `ULTRAHIGH`: indicates the SSD type.
    * `ESSD`: indicates the extreme SSD type.

-> Note
  The MySQL and PostgreSQL supports CLOUDSSD and ESSD, SQL Server supports only ESSD

## Attributes Reference

The following attributes are exported:

* `id` - ID of the read replica instance.

* `db` - Indicates the database information. Structure is documented below.

* `volume` - See Argument Reference above. The `volume` block also contains:

    * `size` - Indicates the volume size. Same as replicated instance disk size.

* `security_group_id` - Indicates the security group which the replica instance belongs to.

* `subnet_id` - Indicates the subnet id (OpenStack network ID).

* `vpc_id` - Indicates the VPC ID (OpenStack router ID).

* `private_ips` - Indicates the private IP address list.

The `db` block supports:

* `port` - Indicates the database port information.

* `type` - Indicates the DB engine. Value: `MySQL`, `PostgreSQL`, `SQLServer`

* `version` - Indicates the database version.

* `user_name` - Indicates the default user name of database.

## Import

Read replicas can be imported using the `id`, e.g.

```sh
terraform import opentelekomcloud_rds_read_replica_v3.rr_1 1a8efa8c-342a-40f0-bc8f-3d27bd603661
```
