---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_instance_v3"
sidebar_current: "docs-opentelekomcloud-resource-rds-instance-v3"
description: |-
  Manages an RDS Instance v3 resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RDS instance rule you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/db_instance_management)

# opentelekomcloud_rds_instance_v3

Manages RDS instance v3 resource.

## Example Usage

### Create a single db instance

```hcl
resource "opentelekomcloud_networking_secgroup_v2" "secgroup" {
  name        = "terraform_test_security_group"
  description = "terraform security group acceptance test"
}

resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "terraform_test_rds_instance"
  availability_zone = [var.availability_zone]

  db {
    password = "P@ssw0rd1!9851"
    type     = "PostgreSQL"
    version  = "9.5"
    port     = "8635"
  }

  security_group_id = opentelekomcloud_networking_secgroup_v2.secgroup.id
  subnet_id         = var.subnet_id
  vpc_id            = var.vpc_id
  flavor            = "rds.pg.c2.medium"

  volume {
    type = "CLOUDSSD"
    size = 100
  }

  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }

  tags = {
    foo = "bar"
    key = "value"
  }
}
```

### Create a primary/standby db instance

```hcl
resource "opentelekomcloud_networking_secgroup_v2" "secgroup" {
  name        = "terraform_test_security_group"
  description = "terraform security group acceptance test"
}

resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "terraform_test_rds_instance"
  availability_zone = [var.availability_zone_1, var.availability_zone_2]

  db {
    password = "P@ssw0rd1!9851"
    type     = "PostgreSQL"
    version  = "9.5"
    port     = "8635"
  }
  security_group_id   = opentelekomcloud_networking_secgroup_v2.secgroup.id
  subnet_id           = var.subnet_id
  vpc_id              = var.vpc_id
  flavor              = "rds.pg.s1.medium.ha"
  ha_replication_mode = "async"

  volume {
    type = "CLOUDSSD"
    size = 100
  }

  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }

  tags = {
    foo = "bar"
    key = "value"
  }
}
```

### Create a db instance with public IP

```hcl
resource "opentelekomcloud_networking_secgroup_v2" "secgroup" {
  name        = "terraform_test_security_group"
  description = "terraform security group acceptance test"
}

resource "opentelekomcloud_compute_floatingip_v2" "ip" {}

resource "opentelekomcloud_rds_instance_v3" "instance" {
  availability_zone = [
    var.availability_zone_1,
    var.availability_zone_2
  ]
  db {
    password = "Telekom!120521"
    type     = "PostgreSQL"
    version  = "9.5"
    port     = "8635"
  }
  name              = "terraform_test_rds_instance"
  security_group_id = opentelekomcloud_networking_secgroup_v2.secgroup.id
  subnet_id         = var.subnet_id
  vpc_id            = var.vpc_id
  volume {
    type = "CLOUDSSD"
    size = 100
  }
  flavor              = "rds.pg.s1.medium.ha"
  ha_replication_mode = "async"
  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }
  public_ips = [
    opentelekomcloud_compute_floatingip_v2.ip.address
  ]
  tags = {
    foo = "bar"
    key = "value"
  }
}
```

### Create a single db instance with encrypted volume

```hcl
resource "opentelekomcloud_kms_key_v1" "key" {
  key_alias       = "key_1"
  key_description = "first test key"
  is_enabled      = true
}

resource "opentelekomcloud_networking_secgroup_v2" "secgroup" {
  name        = "terraform_test_security_group"
  description = "terraform security group acceptance test"
}

resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "terraform_test_rds_instance"
  availability_zone = [var.availability_zone]

  security_group_id = opentelekomcloud_networking_secgroup_v2.secgroup.id
  subnet_id         = var.subnet_id
  vpc_id            = var.vpc_id
  flavor            = "rds.pg.c2.medium"

  db {
    password = "P@ssw0rd1!9851"
    type     = "PostgreSQL"
    version  = "9.5"
    port     = "8635"
  }

  volume {
    disk_encryption_id = opentelekomcloud_kms_key_v1.key.id
    type               = "CLOUDSSD"
    size               = 100
  }
  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }
}
```

### Overriding parameters from template

```hcl
resource "opentelekomcloud_networking_secgroup_v2" "sg" {
  name = "sg-rds-test"
}

resource "opentelekomcloud_rds_parametergroup_v3" "pg" {
  name = "pg-rds-test"
  values = {
    autocommit = "OFF"
  }
  datastore {
    type    = "postgresql"
    version = "10"
  }
}

resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "tf_rds_instance_%s"
  availability_zone = [var.availability_zone]

  db {
    password = "Postgres!120521"
    type     = "PostgreSQL"
    version  = "10"
    port     = "8635"
  }

  security_group_id = opentelekomcloud_networking_secgroup_v2.sg.id
  subnet_id         = var.subnet_id
  vpc_id            = var.vpc_id
  flavor            = "rds.pg.c2.medium"
  volume {
    type = "CLOUDSSD"
    size = 40
  }

  parameters = {
    max_connections = "37",
  }
}
```

### Restore backup to a new instance

```hcl
data "opentelekomcloud_rds_backup_v3" "backup" {
  instance_id = var.rds_instance_id
  type        = "auto"
}

resource "opentelekomcloud_rds_instance_v3" "from_backup" {
  name              = "instance-restored"
  availability_zone = opentelekomcloud_rds_instance_v3.instance.availability_zone
  flavor            = "rds.pg.c2.medium"

  restore_point {
    instance_id = data.opentelekomcloud_rds_backup_v3.backup.instance_id
    backup_id   = data.opentelekomcloud_rds_backup_v3.backup.id
  }

  db {
    password = "Postgres!120521"
    type     = "PostgreSQL"
    version  = "10"
    port     = "8635"
  }
  security_group_id = var.security_group_id
  subnet_id         = var.os_network_id
  vpc_id            = var.os_router_id
  volume {
    type = "CLOUDSSD"
    size = 40
  }
  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }
}
```

### Override timezone parameter (MySQL, for other DBs parameter can have different name)

```hcl
resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "tf_rds_instance_%s"
  availability_zone = opentelekomcloud_rds_instance_v3.instance.availability_zone
  db {
    password = "MySql!112822"
    type     = "MySQL"
    version  = "8.0"
    port     = "8635"
  }
  param_group_id    = opentelekomcloud_rds_parametergroup_v3.pg_1.id
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  subnet_id         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  volume {
    type = "CLOUDSSD"
    size = 40
  }
  flavor = "rds.mysql.c2.medium"
  backup_strategy {
    start_time = "08:00-09:00"
    keep_days  = 1
  }
  tags = {
    muh = "value-create"
    kuh = "value-create"
  }
}


resource "opentelekomcloud_rds_parametergroup_v3" "pg_1" {
  name        = "pg_tmz"
  description = "time zone template"

  values = {
    time_zone = "Africa/Casablanca"
  }

  datastore {
    type    = "mysql"
    version = "8.0"
  }
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Specifies the AZ name. Changing this parameter will create a new resource.

* `db` - (Required, ForceNew) Specifies the database information. Structure is documented below. Changing this parameter will create a new resource.

* `flavor` - (Required) Specifies the specification code.
  Use data source [opentelekomcloud_rds_flavors_v3](../data-sources/rds_flavors_v3.md) to get a list of available flavor names.
  Examples could be `rds.pg.c2.medium` or   `rds.pg.c2.medium.ha` for HA clusters.

* `name` - (Required, ForceNew) Specifies the DB instance name. The DB instance name of the same type
  must be unique for the same tenant. The value must be 4 to 64
  characters in length and start with a letter. It is case-sensitive
  and can contain only letters, digits, hyphens (-), and underscores
  (_).  Changing this parameter will create a new resource.

* `security_group_id` - (Required) Specifies the security group which the RDS DB instance belongs to.
  Changing this parameter will create a new resource.

* `subnet_id` - (Required, ForceNew) Specifies the subnet id. Changing this parameter will create a new resource.

* `volume` - (Required, ForceNew) Specifies the volume information. Structure is documented below.

* `vpc_id` - (Required, ForceNew) Specifies the VPC ID. Changing this parameter will create a new resource.

* `backup_strategy` - (Optional) Specifies the advanced backup policy. Structure is documented below.

* `ha_replication_mode` - (Optional, ForceNew) Specifies the replication mode for the standby DB instance. For MySQL, the value
  is async or semisync. For PostgreSQL, the value is async or sync. For Microsoft SQL Server, the value is sync.
  Parameter is required for HA clusters.

-> Async indicates the asynchronous replication mode. `semisync` indicates the
  semi-synchronous replication mode. sync indicates the synchronous
  replication mode.  Changing this parameter will create a new resource.

* `param_group_id` - (Optional) Specifies the parameter group ID.

* `parameters` - (Optional) Map of additional configuration parameters. Values should be strings. Parameters set here
  overrides values from configuration template (parameter group).

* `lower_case_table_names` - (Optional, ForceNew) Specifies the case-sensitive state of the database table name,
  the default value is "1". Changing this parameter will create a new resource.
  + 0: Table names are stored as fixed and table names are case-sensitive.
  + 1: Table names will be stored in lower case and table names are not case-sensitive.

* `public_ips` - (Optional) Specifies floating IP to be assigned to the instance.
  This should be a list with single element only.

-> Setting public IP is done with assigning floating IP to internally
  created port. So RDS itself doesn't know about this assignment. This assignment
  won't show on the console.
  This argument will be ignored in future when RDSv3 API for EIP assignment will be implemented.

* `tag` - (Optional) Tags key/value pairs to associate with the instance. Deprecated, please use
  the `tags` instead.

* `tags` - (Optional) Tags key/value pairs to associate with the instance.

* `restore_point` - (Optional, ForceNew) Specifies the restoration information. By selecting this option you can either
  create a new RDS instance or restore backup from existing one. Structure is documented below.

* `restore_from_backup` **DEPRECATED**  - (Optional) Specifies whether to restore database to an instance described in current resource.
  Structure is documented below.
  Please use alternative parameter `restore_point`.

* `ssl_enable` - (Optional) Specifies whether SSL should be enabled for MySql instances.

The `db` block supports:

* `password` - (Required, ForceNew) Specifies the database password. The value cannot be
  empty and should contain 8 to 32 characters, including uppercase
  and lowercase letters, digits, and the following special
  characters: ~!@#%^*-_=+? You are advised to enter a strong
  password to improve security, preventing security risks such as
  brute force cracking.  Changing this parameter will create a new resource.

* `port` - (Optional) Specifies the database port information. The MySQL database port
  ranges from 1024 to 65535 (excluding 12017 and 33071, which are
  occupied by the RDS system and cannot be used). The PostgreSQL
  database port ranges from 2100 to 9500. The Microsoft SQL Server
  database port can be 1433 or ranges from 2100 to 9500, excluding
  5355 and 5985. If this parameter is not set, the default value is
  as follows: For MySQL, the default value is 3306. For PostgreSQL,
  the default value is 5432. For Microsoft SQL Server, the default
  value is 1433.  Changing this parameter will create a new resource.

* `type` - (Required, ForceNew) Specifies the DB engine. Value: MySQL, PostgreSQL, SQLServer. Changing this parameter will create a new resource.

* `version` - (Required, ForceNew) Specifies the database version. MySQL databases support MySQL 5.6
  and above. PostgreSQL databases support PostgreSQL 9.5 and above. Microsoft SQL Server
  databases support 2014 SE, 2016 SE, and above.
  Changing this parameter will create a new resource.

The `volume` block supports:

* `disk_encryption_id` - (Optional, ForceNew) Specifies the key ID for disk encryption. Changing this parameter will create a new resource.

* `size` - (Required) Specifies the volume size. Its value range is from 40 GB to 4000
  GB. The value must be a multiple of 10. Changing this resize the volume.

* `type` - (Required, ForceNew) Specifies the volume type. Changing this resize the volume. Its value can be any of the following
  and is case-sensitive:
  * COMMON: SATA storage.
  * ULTRAHIGH: ultra-high I/O storage.
  * CLOUDSSD: cloud SSD storage.
  * ESSD: extreme SSD storage.

-> Note
  The MySQL and PostgreSQL DB engines support the following volume types: CLOUDSSD and ESSD. ESSD is not supported for Single instance types for MySQL and PostgreSQL.
  The SQL Server engine supports the following volume types: COMMON, ULTRAHIGH, and ESSD.

~> **Warning** Specifying both `limit_size` and `trigger_threshold` will enable autoscaling for RDS instance.
  Once autoscaling is activated, the `size` parameter for the volume will be ignored to prevent discrepancies
  between configuration and actual state.

* `limit_size` - (Optional, Int) Specifies the upper limit of automatic expansion of storage, in GB.
  The value ranges from 40 GB to 4,000 GB and must be no less than the current storage of the instance.
  If this parameter is configured, `trigger_threshold` is mandatory.

* `trigger_threshold` - (Optional, Int) Specifies the threshold to trigger automatic expansion.
  If this parameter is configured, `limit_size` is mandatory.
  If the available storage drops to this threshold or `10` GB, the automatic expansion is triggered.
  The valid values are as follows:
    + **10**
    + **15**
    + **20**

The `backup_strategy` block supports:

* `keep_days` - (Optional) Specifies the retention days for specific backup files. The value
  range is from 0 to 732. If this parameter is not specified or set
  to 0, the automated backup policy is disabled. NOTICE:
  Primary/standby DB instances of Microsoft SQL Server do not
  support disabling the automated backup policy.

* `start_time` - (Required) Specifies the backup time window. Automated backups will be
  triggered during the backup time window. It must be a valid value in the &quot;hh:mm-HH:MM&quot;
  format. The current time is in the UTC format. The HH value must
  be 1 greater than the hh value. The values of mm and MM must be
  the same and must be set to any of the following: 00, 15, 30, or
  45. Example value: 08:15-09:15 23:00-00:00.

The `restore_point` block supports:

* `instance_id` - (Required, ForceNew) Specifies the original DB instance ID.

* `backup_id` - (Optional, ForceNew) Specifies the ID of the backup used to restore data.

* `restore_time` - (Optional, ForceNew) Specifies the time point of data restoration in the UNIX timestamp.
  The unit is millisecond and the time zone is UTC.

-> Exactly one of `backup_id` and `restore_time` needs to be set.

The `restore_from_backup` block supports:

* `source_instance_id` - (Required) Specifies the source instance ID.

* `type` - (Required) Specifies the restoration mode. The values can be:
    * `backup` - Indicates using backup files for restoration.
    In this mode `backup_id` is mandatory.
    * `timestamp` - Indicates the point-in-time restoration mode.
    In this mode `restore_time` is mandatory.

* `backup_id` - (Optional) Specifies the ID of the backup used to restore data.
  This parameter must be specified when the backup file is used for restoration.

* `restore_time` - (Optional) Specifies the time point of data restoration in the UNIX timestamp.
  The unit is millisecond and the time zone is UTC.

-> Exactly one of `backup_id` and `restore_time` needs to be set.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `availability_zones` - Indicates the instance AZs.

* `created` - Indicates the creation time.

* `restored_backup_id` - Indicates the backup ID in cases when instance was restored by using
  `restore_from_backup` block.

* `nodes` - Indicates the instance nodes information. Structure is documented below.

* `private_ips` - Indicates the private IP address list. It is a blank string until an
  ECS is created.

* `public_ips` - Indicates the public IP address list.

* `db` - See Argument Reference above. The `db` block additionally contains:

  * `user_name` - Indicates the default user name of database.

The `nodes` block contains:

* `availability_zone` - Indicates the AZ.

* `id` - Indicates the node ID.

* `name` - Indicates the node name.

* `role` - Indicates the node type. The value can be master or slave, indicating the primary node or standby node respectively.

* `status` - Indicates the node status.

* `autoscaling_enabled` - Indicates whether autoscaling was enabled for this resource.

## Timeouts

This resource provides the following timeouts configuration options:
- `create` - Default is 30 minute.

## Import

RDS instance can be imported using the `id`, e.g.

```sh
terraform import opentelekomcloud_rds_instance_v3.instance_1 7117d38e-4c8f-4624-a505-bd96b97d024c
```

## Notes

But due to some attributes missing from the API response, it's required to ignore changes as below.
Also, for some MySQL versions the `lower_case_table_names` should be also added.

```hcl
resource "opentelekomcloud_rds_instance_v3" "instance_1" {
  # ...

  lifecycle {
    ignore_changes = [
      "db",
      "availability_zone",
    ]
  }
}
```
