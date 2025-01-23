---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_instance_v1"
sidebar_current: "docs-opentelekomcloud-resource-rds-instance-v1"
description: |-
  Manages an RDS Instance v1 resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RDS instance you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v1_to_be_discarded/db_instance_management)

**DEPRECATED**
# opentelekomcloud_rds_instance_v1

Manages RDS instance v1 resource within OpenTelekomCloud.

## Example Usage

### Creating a PostgreSQL RDS instance

```hcl
data "opentelekomcloud_rds_flavors_v1" "flavor" {
  datastore_name    = "PostgreSQL"
  datastore_version = "16"
  speccode          = "rds.pg.x1.xlarge.4"
}

resource "opentelekomcloud_compute_secgroup_v2" "secgrp_rds" {
  name        = "secgrp-rds-instance"
  description = "Rds Security Group"
}

resource "opentelekomcloud_rds_instance_v1" "instance" {
  name             = "rds-instance"
  availabilityzone = "eu-de-01"

  datastore {
    type    = "PostgreSQL"
    version = "16"
  }

  flavorref = data.opentelekomcloud_rds_flavors_v1.flavor.id
  volume {
    type = "CLOUDSSD"
    size = 200
  }
  vpc    = "c1095fe7-03df-4205-ad2d-6f4c181d436e"
  dbrtpd = "P@ssw0rd1!9851"
  dbport = "8635"

  nics {
    subnetid = "b65f8d25-c533-47e2-8601-cfaa265a3e3e"
  }
  securitygroup {
    id = opentelekomcloud_compute_secgroup_v2.secgrp_rds.id
  }
  backupstrategy {
    starttime = "04:00:00"
    keepdays  = 4
  }
  ha {
    enable          = true
    replicationmode = "async"
  }
  tag = {
    foo = "bar"
    key = "value"
  }
  depends_on = ["opentelekomcloud_compute_secgroup_v2.secgrp_rds"]
}
```

### Creating a SQLServer RDS instance
```hcl
data "opentelekomcloud_rds_flavors_v1" "flavor" {
  datastore_name    = "SQLServer"
  datastore_version = "2022_SE"
  speccode          = "rds.mssql.c2.xlarge"
}

resource "opentelekomcloud_compute_secgroup_v2" "secgrp_rds" {
  name        = "secgrp-rds-instance"
  description = "Rds Security Group"
}

resource "opentelekomcloud_rds_instance_v1" "instance" {
  name             = "rds-instance"
  availabilityzone = "eu-de-01"
  flavorref        = data.opentelekomcloud_rds_flavors_v1.flavor.id
  vpc              = "c1095fe7-03df-4205-ad2d-6f4c181d436e"
  dbport           = "8635"
  dbrtpd           = "P@ssw0rd1!9851"

  datastore {
    type    = "SQLServer"
    version = "2022_SE"
  }
  volume {
    type = "COMMON"
    size = 200
  }
  nics {
    subnetid = "b65f8d25-c533-47e2-8601-cfaa265a3e3e"
  }
  securitygroup {
    id = opentelekomcloud_compute_secgroup_v2.secgrp_rds.id
  }
  backupstrategy {
    starttime = "04:00:00"
    keepdays  = 4
  }
  depends_on = ["opentelekomcloud_compute_secgroup_v2.secgrp_rds"]
}
```

### Creating a MySQL RDS instance
```hcl
data "opentelekomcloud_rds_flavors_v1" "flavor" {
  datastore_name    = "MySQL"
  datastore_version = "8.0"
  speccode          = "rds.mysql.n1.2xlarge.2"
}

resource "opentelekomcloud_compute_secgroup_v2" "secgrp_rds" {
  name        = "secgrp-rds-instance"
  description = "Rds Security Group"
}

resource "opentelekomcloud_rds_instance_v1" "instance" {
  name             = "rds-instance"
  availabilityzone = "eu-de-01"

  vpc       = "c1095fe7-03df-4205-ad2d-6f4c181d436e"
  dbport    = "8635"
  dbrtpd    = "P@ssw0rd1!9851"
  flavorref = data.opentelekomcloud_rds_flavors_v1.flavor.id
  datastore {
    type    = "MySQL"
    version = "8.0"
  }
  volume {
    type = "CLOUDSSD"
    size = 200
  }
  nics {
    subnetid = "b65f8d25-c533-47e2-8601-cfaa265a3e3e"
  }
  securitygroup {
    id = opentelekomcloud_compute_secgroup_v2.secgrp_rds.id
  }
  backupstrategy {
    starttime = "04:00:00"
    keepdays  = 4
  }
  ha {
    enable          = true
    replicationmode = "async"
  }
  depends_on = ["opentelekomcloud_compute_secgroup_v2.secgrp_rds"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the DB instance name. The DB instance name of
  the same type is unique in the same tenant. The changes of the instance name
  will be suppressed in HA scenario.

* `datastore` - (Required) Specifies database information. The structure is
  described below.

* `flavorref` - (Required) Specifies the specification ID (flavors.id in the
  response message in Obtaining All DB Instance Specifications). If you want
  to enable ha for the rds instance, a flavor with ha speccode is required.

* `volume` - (Required) Specifies the volume information. The structure is described
  below.

* `availabilityzone` - (Required) Specifies the ID of the AZ.

* `vpc` - (Required) Specifies the VPC ID. For details about how to obtain this
  parameter value, see section "Virtual Private Cloud" in the Virtual Private
  Cloud API Reference.

* `nics` - (Required) Specifies the nics information. For details about how
  to obtain this parameter value, see section "Subnet" in the Virtual Private
  Cloud API Reference. The structure is described below.

* `securitygroup` - (Required) Specifies the security group which the RDS DB
  instance belongs to. The structure is described below.

* `dbport` - (Optional) Specifies the database port number.

* `backupstrategy` - (Optional) Specifies the advanced backup policy. The structure
  is described below.

* `dbrtpd` - (Required) Specifies the password for user root of the database.

* `tag` - (Optional) Tags key/value pairs to associate with the instance.

* `ha` - (Optional) Specifies the parameters configured on HA and is used when
  creating HA DB instances. The structure is described below. NOTICE:
  RDS for Microsoft SQL Server does not support creating HA DB instances and
  this parameter is not involved.

The `datastore` block supports:

* `type` - (Required) Specifies the DB engine. Currently, MySQL, PostgreSQL, and
  Microsoft SQL Server are supported. The value is MySQL, PostgreSQL, or SQLServer.

* `version` - (Required) Specifies the DB instance version.

* Available value for attributes

| type       | version                                                                   |
|------------|---------------------------------------------------------------------------|
| PostgreSQL | 9.5.5 <br> 9.6.3 <br> 9.6.5                                               |
| MySQL      | 5.6.33 <br>5.6.30  <br>5.6.34 <br>5.6.35 <br>5.6.36 <br>5.7.17 <br>5.7.20 |
| SQLServer  | 2014 SP2 SE                                                               |

The `volume` block supports:

* `type` - (Required) Specifies the volume type. Valid value:
  It must be COMMON (SATA) or ULTRAHIGH (SSD) and is case-sensitive.

* `size` - (Required) Specifies the volume size.
  Its value must be a multiple of 10 and the value range is 100 GB to 2000 GB.

The `nics` block supports:

* `subnetId` - (Required) Specifies the subnet ID obtained from the VPC.

The `securitygroup ` block supports:

* `id` - (Required) Specifies the ID obtained from the securitygroup.

The `backupstrategy ` block supports:

* `starttime` - (Optional) Indicates the backup start time that has been set.
  The backup task will be triggered within one hour after the backup start time.
  Valid value: The value cannot be empty. It must use the hh:mm:ss format and
  must be valid. The current time is the UTC time.

* `keepdays` - (Optional) Specifies the number of days to retain the generated backup files.
  Its value range is 0 to 35. If this parameter is not specified or set to 0, the
  automated backup policy is disabled.

The `ha` block supports:

* `enable` - (Optional) Specifies the configured parameters on the HA.
  Valid value: The value is true or false. The value true indicates creating
  HA DB instances. The value false indicates creating a single DB instance.

* `replicationmode` - (Optional) Specifies the replication mode for the standby DB instance.
  The value cannot be empty.
  For MySQL, the value is async or semisync.
  For PostgreSQL, the value is async or sync.

## Attributes Reference

The following attributes are exported:

* `name` - See Argument Reference above.

* `flavorref` - See Argument Reference above.

* `volume` - See Argument Reference above.

* `availabilityzone` - See Argument Reference above.

* `vpc` - See Argument Reference above.

* `nics` - See Argument Reference above.

* `securitygroup` - See Argument Reference above.

* `dbport` - See Argument Reference above.

* `backupstrategy` - See Argument Reference above.

* `dbrtpd` - See Argument Reference above.

* `ha` - See Argument Reference above.

* `status` - Indicates the DB instance status.

* `hostname` - Indicates the instance connection address. It is a blank string.

* `type` - Indicates the DB instance type, which can be master or readreplica.

* `created` - Indicates the creation time in the following format: yyyy-mm-dd Thh:mm:ssZ.

* `updated` - Indicates the update time in the following format: yyyy-mm-dd Thh:mm:ssZ.

## Attributes Reference

The following attributes can be updated:

* `volume.size` - See Argument Reference above.

* `flavorref` - See Argument Reference above.

* `backupstrategy` - See Argument Reference above.
