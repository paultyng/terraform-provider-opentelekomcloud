package acceptance

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/backups"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getBackupResourceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.RdsV3Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating RDS v3 client: %s", err)
	}

	opts := backups.ListOpts{
		InstanceID: state.Primary.Attributes["instance_id"],
		BackupID:   state.Primary.ID,
	}

	backupList, err := backups.List(client, opts)

	if err != nil {
		return nil, fmt.Errorf("error listing backups: %s", err)
	}

	return backupList[0], nil
}

func TestAccResourceRDSV3Backup_basic(t *testing.T) {
	var (
		obj     backups.Backup
		rName   = "opentelekomcloud_rds_backup_v3.test"
		postfix = acctest.RandString(3)
	)

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getBackupResourceFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckRdsBackupV3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRDSV3BackupBasic(postfix),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					testAccCheckRdsBackupV3Exists(rName),
					resource.TestCheckResourceAttr(rName, "name", "tf_rds_backup_"+postfix),
					resource.TestCheckResourceAttr(rName, "type", "manual"),
					resource.TestCheckResourceAttr(rName, "status", "COMPLETED"),
					resource.TestCheckResourceAttrSet(rName, "id"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccBackupImportStateFunc(rName),
			},
		},
	})
}

func testAccBackupImportStateFunc(rsName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		var instanceId, backupId string
		rs, ok := s.RootModule().Resources[rsName]
		if !ok {
			return "", fmt.Errorf("the resource (%s) of RDF backup is not found in the tfstate", rsName)
		}
		instanceId = rs.Primary.Attributes["instance_id"]
		backupId = rs.Primary.ID
		if instanceId == "" || backupId == "" {
			return "", fmt.Errorf("some import IDs are missing: "+
				"'<instance_id>/<id>', but got '%s/%s'",
				instanceId, backupId)
		}
		return fmt.Sprintf("%s/%s", instanceId, backupId), nil
	}
}

func testAccCheckRdsBackupV3Destroy(s *terraform.State) error {
	const backupDeleteRetryTimeout = 5 * time.Minute

	config := common.TestAccProvider.Meta().(*cfg.Config)
	client, err := config.RdsV3Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating RDSv3 client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_rds_backup_v3" {
			continue
		}

		err = resource.RetryContext(context.Background(), backupDeleteRetryTimeout, func() *resource.RetryError {
			backupList, err := backups.List(client, backups.ListOpts{
				BackupID:   rs.Primary.ID,
				InstanceID: rs.Primary.Attributes["instance_id"],
			})

			if err != nil && !strings.Contains(err.Error(), "The backup file does not exist") {
				return resource.NonRetryableError(fmt.Errorf("error listing backups: %s", err))
			}

			for _, backup := range backupList {
				if backup.ID == rs.Primary.ID {
					if backup.Status == "DELETING" {
						return resource.RetryableError(fmt.Errorf("backup is still in a deleting state"))
					}
					return resource.NonRetryableError(fmt.Errorf("backup still exists"))
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckRdsBackupV3Exists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.RdsV3Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating RDSv3 client: %s", err)
		}

		backup, err := backups.List(client, backups.ListOpts{
			BackupID:   rs.Primary.ID,
			InstanceID: rs.Primary.Attributes["instance_id"],
		})
		if err != nil {
			return fmt.Errorf("error getting backup %s: %s", rs.Primary.ID, err)
		}

		if backup == nil {
			return fmt.Errorf("backup not found with id: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceRDSV3BackupBasic(postfix string) string {
	return fmt.Sprintf(`
%s
%s

resource "opentelekomcloud_rds_instance_v3" "instance" {
  name              = "tf_rds_instance_%s"
  availability_zone = ["%s"]
  db {
    password = "Postgres!120521"
    type     = "PostgreSQL"
    version  = "16"
    port     = "8635"
  }
  security_group_id = data.opentelekomcloud_networking_secgroup_v2.default_secgroup.id
  subnet_id         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  vpc_id            = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  volume {
    type = "COMMON"
    size = 40
  }
  flavor = "rds.pg.c2.large"
}

resource "opentelekomcloud_rds_backup_v3" "test" {
  instance_id = opentelekomcloud_rds_instance_v3.instance.id
  name        = "tf_rds_backup_%s"
}

`, common.DataSourceSecGroupDefault, common.DataSourceSubnet, postfix, env.OS_AVAILABILITY_ZONE, postfix)
}
