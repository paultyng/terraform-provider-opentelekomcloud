---
subcategory: "Relational Database Service (RDS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rds_backup_v3"
sidebar_current: "docs-opentelekomcloud-resource-rds-backup-v3"
description: |-
  Manages an RDS Backup resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RDS backup rule you can get at
[documentation portal](https://docs.otc.t-systems.com/relational-database-service/api-ref/api_v3_recommended/backup_and_restoration)

# opentelekomcloud_rds_backup_v3

Manages a manual RDS backup.

## Example Usage

### Create a basic RDS backup

```hcl
resource "opentelekomcloud_rds_backup_v3" "test" {
  instance_id = opentelekomcloud_rds_instance_v3.instance.id
  name        = "rds-backup-test-01"
}
```

### Create a specific RDS databases backup for Microsoft SQL Server

```hcl
resource "opentelekomcloud_rds_backup_v3" "test" {
  instance_id = opentelekomcloud_rds_instance_v3.instance.id
  name        = "rds-backup-test-01"
  databases   = ["test", "test2"]
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) The ID of the RDS instance to which the backup belongs.

* `name` - (Required, String, ForceNew) The name of the backup.

* `databases` - (Optional, List, ForceNew) Specifies a list of self-built Microsoft SQL Server databases that are partially backed up.
                (Only Microsoft SQL Server support partial backups.)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the backup.

* `begin_time` - Indicates the backup start time in the "yyyy-mm-ddThh:mm:ssZ" format,
                 where "T" indicates the start time of the time field, and "Z" indicates the time zone offset.

* `status` - Indicates the backup status. Values:
             - BUILDING: Backup in progress
             - COMPLETED: Backup completed
             - FAILED: Backup failed
             - DELETING: Backup being deleted

* `type` - Indicates the backup type. Values:
           - auto: automated full backup
           - manual: manual full backup
           - fragment: differential full backup
           - incremental: automated incremental backup

## Import

RDS backup can be imported using related RDS `instance_id` and their `backup_id`, separated by the slashes, e.g.

```bash
$ terraform import opentelekomcloud_rds_backup_v3.backup <instance_id>/<backup_id>
```
