package rds

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/subnets"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/ports"
	tag "github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v1/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/backups"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/configurations"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/flavors"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/instances"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/security"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceRdsInstanceV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRdsInstanceV3Create,
		ReadContext:   resourceRdsInstanceV3Read,
		UpdateContext: resourceRdsInstanceV3Update,
		DeleteContext: resourceRdsInstanceV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			// validateRDSv3Flavor("flavor"),
			common.ValidateSubnet("subnet_id"),
			common.ValidateVPC("vpc_id"),
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"restore_point": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"restore_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							ExactlyOneOf: []string{"restore_point.0.backup_id"},
						},
						"backup_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ExactlyOneOf: []string{"restore_point.0.restore_time"},
						},
					},
				},
			},
			"restore_from_backup": {
				Type:       schema.TypeList,
				Optional:   true,
				MaxItems:   1,
				Computed:   false,
				Deprecated: "Use `restore_point` instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_instance_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"backup_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"restore_time": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"db": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Required:  true,
							ForceNew:  true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true, // can't be set in case of restored backup
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"PostgreSQL", "MySQL", "SQLServer",
							}, false),
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true, // can't be set in case of restored backup
							Computed: true,
							ForceNew: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
							Optional: true,
						},
						"user_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"disk_encryption_id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
							ForceNew: true,
						},
						"limit_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							RequiredWith: []string{"volume.0.trigger_threshold"},
						},
						"trigger_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							RequiredWith: []string{"volume.0.limit_size"},
						},
					},
				},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"backup_strategy": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"keep_days": {
							Type:     schema.TypeInt,
							Computed: true,
							Optional: true,
						},
					},
				},
			},
			"ha_replication_mode": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"tag": {
				Type:          schema.TypeMap,
				Optional:      true,
				ValidateFunc:  common.ValidateTags,
				Deprecated:    "Please use `tags` instead",
				ConflictsWith: []string{"tags"},
			},
			"tags": {
				Type:          schema.TypeMap,
				Optional:      true,
				ValidateFunc:  common.ValidateTags,
				ConflictsWith: []string{"tag"},
			},
			"param_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"public_ips": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: common.ValidateIP,
				},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ssl_enable": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lower_case_table_names": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: false,
				Optional: true,
			},
			"restored_backup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"autoscaling_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceRDSDataStore(d *schema.ResourceData) *instances.Datastore {
	dataStoreRaw := d.Get("db").([]interface{})[0].(map[string]interface{})
	dataStore := instances.Datastore{
		Type:    dataStoreRaw["type"].(string),
		Version: dataStoreRaw["version"].(string),
	}
	return &dataStore
}

func resourceRDSVolume(d *schema.ResourceData) *instances.Volume {
	volumeRaw := d.Get("volume").([]interface{})[0].(map[string]interface{})
	volume := instances.Volume{
		Type: volumeRaw["type"].(string),
		Size: volumeRaw["size"].(int),
	}
	return &volume
}

func resourceRDSBackupStrategy(d *schema.ResourceData) *instances.BackupStrategy {
	backupStrategyRaw := d.Get("backup_strategy").([]interface{})
	if len(backupStrategyRaw) == 0 {
		return nil
	}
	backupStrategyInfo := backupStrategyRaw[0].(map[string]interface{})
	backupStrategy := instances.BackupStrategy{
		StartTime: backupStrategyInfo["start_time"].(string),
		KeepDays:  backupStrategyInfo["keep_days"].(int),
	}
	return &backupStrategy
}

func resourceRDSHa(d *schema.ResourceData) *instances.Ha {
	replicationMode := d.Get("ha_replication_mode").(string)
	if replicationMode == "" {
		return nil
	}
	ha := instances.Ha{
		Mode:            "Ha",
		ReplicationMode: replicationMode,
	}
	return &ha
}

func resourceRDSChangeMode() *instances.ChargeInfo {
	chargeInfo := instances.ChargeInfo{
		ChargeMode: "postPaid",
	}
	return &chargeInfo
}

func resourceRDSDbInfo(d *schema.ResourceData) map[string]interface{} {
	dbRaw := d.Get("db").([]interface{})[0].(map[string]interface{})
	return dbRaw
}

func resourceRDSAvailabilityZones(d *schema.ResourceData) string {
	azRaw := d.Get("availability_zone").([]interface{})
	zones := make([]string, 0)
	for _, v := range azRaw {
		zones = append(zones, v.(string))
	}
	zone := strings.Join(zones, ",")
	return zone
}

func resourceRdsInstanceV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.RdsV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	if _, ok := d.GetOk("restore_from_backup.0.source_instance_id"); ok {
		return fmterr.Errorf("point in time restoration can be only produced on existing instance")
	}

	dbInfo := resourceRDSDbInfo(d)
	volumeInfo := d.Get("volume").([]interface{})[0].(map[string]interface{})
	dbPort := dbInfo["port"].(int)
	var dbPortString string
	if dbPort != 0 {
		dbPortString = strconv.Itoa(dbInfo["port"].(int))
	} else {
		dbPortString = ""
	}

	datastore := resourceRDSDataStore(d)
	var r *instances.CreateRds
	if _, ok := d.GetOk("restore_point"); ok {
		restoreOpts := backups.RestoreToNewOpts{
			Name:             d.Get("name").(string),
			Ha:               resourceRDSHa(d),
			ConfigurationId:  d.Get("param_group_id").(string),
			Port:             dbPortString,
			Password:         dbInfo["password"].(string),
			BackupStrategy:   resourceRDSBackupStrategy(d),
			DiskEncryptionId: volumeInfo["disk_encryption_id"].(string),
			FlavorRef:        d.Get("flavor").(string),
			Volume:           resourceRDSVolume(d),
			AvailabilityZone: resourceRDSAvailabilityZones(d),
			VpcId:            d.Get("vpc_id").(string),
			SubnetId:         d.Get("subnet_id").(string),
			SecurityGroupId:  d.Get("security_group_id").(string),
			RestorePoint:     resourceRestorePoint(d),
		}
		if ok := d.Get("lower_case_table_names").(string); ok != "" {
			lowerCase := &instances.Param{
				LowerCaseTableNames: ok,
			}
			restoreOpts.UnchangeableParam = lowerCase
		}
		restored, err := backups.RestoreToNew(client, restoreOpts)
		if err != nil {
			return fmterr.Errorf("error creating new RDSv3 instance from backup: %w", err)
		}
		r = restored
	} else {
		createOpts := instances.CreateRdsOpts{
			Name:             d.Get("name").(string),
			Datastore:        datastore,
			Ha:               resourceRDSHa(d),
			ConfigurationId:  d.Get("param_group_id").(string),
			Port:             dbPortString,
			Password:         dbInfo["password"].(string),
			BackupStrategy:   resourceRDSBackupStrategy(d),
			DiskEncryptionId: volumeInfo["disk_encryption_id"].(string),
			FlavorRef:        d.Get("flavor").(string),
			Volume:           resourceRDSVolume(d),
			Region:           config.GetRegion(d),
			AvailabilityZone: resourceRDSAvailabilityZones(d),
			VpcId:            d.Get("vpc_id").(string),
			SubnetId:         d.Get("subnet_id").(string),
			SecurityGroupId:  d.Get("security_group_id").(string),
			ChargeInfo:       resourceRDSChangeMode(),
		}
		if ok := d.Get("lower_case_table_names").(string); ok != "" {
			lowerCase := &instances.Param{
				LowerCaseTableNames: ok,
			}
			createOpts.UnchangeableParam = lowerCase
		}
		created, err := instances.Create(client, createOpts)
		if err != nil {
			return fmterr.Errorf("error creating new RDSv3 instance: %w", err)
		}
		r = created
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	if err := instances.WaitForJobCompleted(client, int(timeout.Seconds()), r.JobId); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Instance.Id)

	if common.HasFilledOpt(d, "tag") {
		rdsInstance, err := GetRdsInstance(client, r.Instance.Id)
		if err != nil {
			return diag.FromErr(err)
		}
		nodeID := getMasterID(rdsInstance.Nodes)

		if nodeID == "" {
			log.Printf("[WARN] Error setting tag(key/value) of instance: %s", r.Instance.Id)
			return nil
		}
		tagClient, err := config.RdsTagV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud RDSv1 tag client: %s", err)
		}
		tagMap := d.Get("tag").(map[string]interface{})
		log.Printf("[DEBUG] Setting tag(key/value): %v", tagMap)
		for key, val := range tagMap {
			tagOpts := tag.CreateOpts{
				Key:   key,
				Value: val.(string),
			}
			err = tag.Create(tagClient, nodeID, tagOpts).ExtractErr()
			if err != nil {
				log.Printf("[WARN] Error setting tag(key/value) of instance %s, err: %s", r.Instance.Id, err)
			}
		}
	}

	if common.HasFilledOpt(d, "tags") {
		tagRaw := d.Get("tags").(map[string]interface{})
		if len(tagRaw) > 0 {
			tagList := common.ExpandResourceTags(tagRaw)
			if err := tags.Create(client, "instances", r.Instance.Id, tagList).ExtractErr(); err != nil {
				return fmterr.Errorf("error setting tags of RDSv3 instance: %w", err)
			}
		}
	}

	ip := getPublicIP(d)
	if ip != "" {
		rdsInstance, err := GetRdsInstance(client, d.Id())
		if err != nil {
			return fmterr.Errorf("error fetching RDS instance to set EIP: %s", err)
		}
		nw, err := config.NetworkingV2Client(config.GetRegion(d))
		if err != nil {
			return diag.FromErr(err)
		}
		subnetID, err := getSubnetSubnetID(d, config)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := assignEipToInstance(nw, ip, rdsInstance.PrivateIps[0], subnetID); err != nil {
			log.Printf("[WARN] failed to assign public IP: %s", err)
		}
	}

	templateRestart, err := assureTemplateApplied(d, client)
	if err != nil {
		return fmterr.Errorf("error making sure configuration template is applied: %w", err)
	}

	paramRestart := false
	if _, ok := d.GetOk("parameters"); ok {
		stateConf := &resource.StateChangeConf{
			Pending:      []string{"PENDING"},
			Target:       []string{"SUCCESS"},
			Refresh:      waitForParameterApply(d, client),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			PollInterval: 10 * time.Second,
		}

		result, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		paramRestart = result.(bool)
	}

	if templateRestart || paramRestart {
		if err := instances.WaitForStateAvailable(client, 1200, d.Id()); err != nil {
			return fmterr.Errorf("error waiting for instance to become available: %w", err)
		}
		if err := restartInstance(d, client); err != nil {
			return diag.FromErr(err)
		}
	}

	if sslEnable := d.Get("ssl_enable").(bool); sslEnable {
		if dbType := d.Get("db.0.type").(string); strings.ToLower(dbType) == "mysql" {
			updateOpts := security.SwitchSslOpts{
				SslOption:  sslEnable,
				InstanceId: d.Id(),
			}
			log.Printf("[DEBUG] Update opts of SSL configuration: %+v", updateOpts)
			err := security.SwitchSsl(client, updateOpts)
			if err != nil {
				return fmterr.Errorf("error updating instance SSL configuration: %s ", err)
			}
			stateConf := &resource.StateChangeConf{
				Pending:      []string{"PENDING"},
				Target:       []string{"SUCCESS"},
				Refresh:      waitForSSLEnable(d, client),
				Timeout:      d.Timeout(schema.TimeoutCreate),
				PollInterval: 5 * time.Second,
			}

			_, err = stateConf.WaitForStateContext(ctx)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if size := d.Get("volume.0.limit_size").(int); size > 0 {
		if err = enableVolumeAutoExpand(ctx, d, client, d.Id(), size); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRdsInstanceV3Read(ctx, d, meta)
}

func resourceRestorePoint(d *schema.ResourceData) backups.RestorePoint {
	restorePoint := backups.RestorePoint{
		InstanceID: d.Get("restore_point.0.instance_id").(string),
	}
	if tm, ok := d.GetOk("restore_point.0.restore_time"); ok {
		restorePoint.RestoreTime = tm.(int)
		restorePoint.Type = backups.TypeTimestamp
	} else if id, ok := d.GetOk("restore_point.0.backup_id"); ok {
		restorePoint.BackupID = id.(string)
		restorePoint.Type = backups.TypeBackup
	}
	return restorePoint
}

func assureTemplateApplied(d *schema.ResourceData, client *golangsdk.ServiceClient) (bool, error) {
	templateID := d.Get("param_group_id").(string)
	if templateID == "" {
		return false, nil
	}

	applied, err := configurations.Get(client, templateID)
	if err != nil {
		return false, fmt.Errorf("error getting parameter template %s: %w", templateID, err)
	}
	// convert to the map
	appliedParams := make(map[string]configurations.Parameter, len(applied.Parameters))
	for _, param := range applied.Parameters {
		appliedParams[param.Name] = param
	}

	current, err := configurations.GetForInstance(client, d.Id())
	if err != nil {
		return false, fmt.Errorf("error getting configuration of instance %s: %w", d.Id(), err)
	}

	needsReapply := false
	for _, val := range current.Parameters {
		param, ok := appliedParams[val.Name]
		if !ok { // Then it's not from the template
			continue
		}
		if val.Value != param.Value {
			needsReapply = true
			break // that's enough
		}
	}
	if !needsReapply {
		return false, nil
	}

	return applyTemplate(d, client)
}

func restartInstance(d *schema.ResourceData, client *golangsdk.ServiceClient) error {
	job, err := instances.Restart(client, instances.RestartOpts{
		InstanceId: d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error restarting RDS instance: %w", err)
	}
	timeout := d.Timeout(schema.TimeoutCreate)
	if err := instances.WaitForJobCompleted(client, int(timeout.Seconds()), *job); err != nil {
		return fmt.Errorf("error waiting for instance to reboot: %w", err)
	}
	if err := instances.WaitForStateAvailable(client, int(timeout.Seconds()), d.Id()); err != nil {
		return fmt.Errorf("error waiting for instance to become available: %w", err)
	}
	return nil
}

func applyTemplate(d *schema.ResourceData, client *golangsdk.ServiceClient) (bool, error) {
	templateID := d.Get("param_group_id").(string)
	applyResult, err := configurations.Apply(client, configurations.ApplyOpts{
		InstanceIDs: []string{d.Id()},
		ConfigId:    templateID,
	})
	if err != nil {
		return false, fmt.Errorf("error applying configuration %s to instance %s: %w", templateID, d.Id(), err)
	}
	restartRequired := false
	switch l := len(applyResult.ApplyResults); l {
	case 0:
		return false, fmt.Errorf("empty appply results")
	case 1:
		result := applyResult.ApplyResults[0]
		if !result.Success {
			return false, fmt.Errorf("unsuccessful apply of template instance %s", result.InstanceID)
		}
		restartRequired = result.RestartRequired
	default:
		return false, fmt.Errorf("more that one apply result returned: %#v", applyResult.ApplyResults)
	}

	waitSeconds := int(d.Timeout(schema.TimeoutCreate).Seconds())
	if err := instances.WaitForStateAvailable(client, waitSeconds, d.Id()); err != nil {
		return false, err
	}

	return restartRequired, nil
}

func GetRdsInstance(client *golangsdk.ServiceClient, rdsId string) (*instances.InstanceResponse, error) {
	listOpts := instances.ListOpts{
		Id: rdsId,
	}
	n, err := instances.List(client, listOpts)
	if err != nil {
		return nil, err
	}

	if len(n.Instances) == 0 {
		return nil, nil
	}
	return &n.Instances[0], nil
}

func getPrivateIP(d *schema.ResourceData) string {
	return d.Get("private_ips").([]interface{})[0].(string)
}

func getPublicIP(d *schema.ResourceData) string {
	publicIpRaw := d.Get("public_ips").([]interface{})
	if len(publicIpRaw) > 0 {
		return publicIpRaw[0].(string)
	}
	return ""
}

func findFloatingIP(client *golangsdk.ServiceClient, address string) (id string, err error) {
	var opts = floatingips.ListOpts{FloatingIP: address}

	pgFIP, err := floatingips.List(client, opts).AllPages()
	if err != nil {
		return
	}
	floatingIPs, err := floatingips.ExtractFloatingIPs(pgFIP)
	if err != nil {
		return
	}
	if len(floatingIPs) == 0 {
		return
	}

	for _, ip := range floatingIPs {
		if address != ip.FloatingIP {
			continue
		}
		return ip.ID, nil
	}
	return
}

// find assigned port
func findPort(client *golangsdk.ServiceClient, privateIP string, subnetID string) (id string, err error) {
	pg, err := ports.List(client, nil).AllPages()
	if err != nil {
		return
	}

	portList, err := ports.ExtractPorts(pg)
	if err != nil {
		return
	}

	for _, port := range portList {
		if len(port.FixedIPs) > 0 {
			address := port.FixedIPs[0]
			if address.IPAddress == privateIP && address.SubnetID == subnetID {
				id = port.ID
				return
			}
		}
	}
	return
}

func assignEipToInstance(client *golangsdk.ServiceClient, publicIP, privateIP, subnetID string) error {
	portID, err := findPort(client, privateIP, subnetID)
	if err != nil {
		return err
	}

	ipID, err := findFloatingIP(client, publicIP)
	if err != nil {
		return err
	}
	return floatingips.Update(client, ipID, floatingips.UpdateOpts{PortID: &portID}).Err
}

func getSubnetSubnetID(d *schema.ResourceData, config *cfg.Config) (id string, err error) {
	subnetClient, err := config.NetworkingV1Client(config.GetRegion(d))
	if err != nil {
		err = fmt.Errorf("[WARN] Failed to create VPC client")
		return
	}
	sn, err := subnets.Get(subnetClient, d.Get("subnet_id").(string)).Extract()
	if err != nil {
		return
	}
	id = sn.SubnetID
	return
}

func unAssignEipFromInstance(client *golangsdk.ServiceClient, oldPublicIP string) error {
	ipID, err := findFloatingIP(client, oldPublicIP)
	if err != nil {
		return err
	}
	if ipID == "" {
		return nil
	}
	return floatingips.Update(client, ipID, floatingips.UpdateOpts{PortID: nil}).Err
}

func resourceRdsInstanceV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.RdsV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	if d.HasChange("restore_from_backup") {
		rawPitr := d.Get("restore_from_backup").([]interface{})
		if len(rawPitr) > 0 {
			pitr := rawPitr[0].(map[string]interface{})
			pitrOpts := backups.RestorePITROpts{
				Source: backups.Source{
					BackupID:    pitr["backup_id"].(string),
					InstanceID:  pitr["source_instance_id"].(string),
					RestoreTime: int64(pitr["restore_time"].(int)),
					Type:        pitr["type"].(string),
				},
				Target: backups.Target{
					InstanceID: d.Id(),
				},
			}
			_, err = backups.RestorePITR(client, pitrOpts)
			if err != nil {
				return fmterr.Errorf("error in point in time restoration: %s ", err)
			}
			if err := instances.WaitForStateAvailable(client, 1200, d.Id()); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("restore_point") {
		rawPitr := d.Get("restore_point").([]interface{})
		if len(rawPitr) > 0 {
			pitr := rawPitr[0].(map[string]interface{})
			pitrOpts := backups.RestorePITROpts{
				Source: backups.Source{
					BackupID:    pitr["backup_id"].(string),
					InstanceID:  pitr["instance_id"].(string),
					RestoreTime: int64(pitr["restore_time"].(int)),
				},
				Target: backups.Target{
					InstanceID: d.Id(),
				},
			}
			if pitrOpts.Source.BackupID != "" {
				pitrOpts.Source.Type = "backup"
			} else {
				pitrOpts.Source.Type = "timestamp"
			}

			_, err = backups.RestorePITR(client, pitrOpts)
			if err != nil {
				return fmterr.Errorf("error in point in time restoration: %s ", err)
			}

			// Additional sleep is required to handle state transitions during PITR operations.
			// During PITR application, the backend may undergo 2 sequential state changes instead of 1.
			// Current waitForStateAvailable function terminates after detecting the first "Available" state.
			time.Sleep(20 * time.Second)
			if err := instances.WaitForStateAvailable(client, 1200, d.Id()); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("security_group_id") {
		updateOpts := security.SetSecurityGroupOpts{
			InstanceId:      d.Id(),
			SecurityGroupId: d.Get("security_group_id").(string),
		}
		_, err := security.SetSecurityGroup(client, updateOpts)
		if err != nil {
			return fmterr.Errorf("error updating instance security group: %s ", err)
		}
	}

	var updateBackupOpts backups.UpdateOpts

	if d.HasChange("backup_strategy") {
		backupRaw := resourceRDSBackupStrategy(d)
		updateBackupOpts.InstanceId = d.Id()
		if backupRaw.KeepDays != 0 {
			updateBackupOpts.KeepDays = &backupRaw.KeepDays
			updateBackupOpts.StartTime = backupRaw.StartTime
			updateBackupOpts.Period = "1,2,3,4,5,6,7"
			log.Printf("[DEBUG] updateOpts: %#v", updateBackupOpts)
		} else {
			updateBackupOpts.KeepDays = pointerto.Int(0)
		}
		if err = backups.Update(client, updateBackupOpts); err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud RDSv3 Instance: %s", err)
		}
	}

	// Fetching node id
	var nodeID string
	v, err := GetRdsInstance(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	nodeID = getMasterID(v.Nodes)
	if nodeID == "" {
		log.Printf("[WARN] Error fetching node id of instance:%s", d.Id())
		return nil
	}

	if d.HasChange("tag") {
		oldTagRaw, newTagRaw := d.GetChange("tag")
		oldTag := oldTagRaw.(map[string]interface{})
		newTag := newTagRaw.(map[string]interface{})
		create, remove := diffTagsRDS(oldTag, newTag)
		tagClient, err := config.RdsTagV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud RDSv3 tag client: %s ", err)
		}

		if len(remove) > 0 {
			for _, opts := range remove {
				err = tag.Delete(tagClient, nodeID, opts).ExtractErr()
				if err != nil {
					log.Printf("[WARN] Error deleting tag(key/value) of instance: %s, err: %s", d.Id(), err)
				}
			}
		}
		if len(create) > 0 {
			for _, opts := range create {
				err = tag.Create(tagClient, nodeID, opts).ExtractErr()
				if err != nil {
					log.Printf("[WARN] Error setting tag(key/value) of instance: %s, err: %s", d.Id(), err)
				}
			}
		}
	}
	if d.HasChange("tags") {
		if err := common.UpdateResourceTags(client, d, "instances", d.Id()); err != nil {
			return fmterr.Errorf("error updating tags of RDSv3 instance %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("flavor") {
		_, newFlavor := d.GetChange("flavor")

		updateFlavorOpts := instances.ResizeOpts{
			InstanceId: d.Id(),
			SpecCode:   newFlavor.(string),
		}

		log.Printf("Update flavor could be done only in status `available`")
		if err := instances.WaitForStateAvailable(client, 1200, d.Id()); err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[DEBUG] Update flavor: %s", newFlavor.(string))
		_, err = instances.Resize(client, updateFlavorOpts)
		if err != nil {
			return fmterr.Errorf("error updating instance Flavor from result: %s", err)
		}

		log.Printf("Waiting for RDSv3 become in status `available`")
		if err := instances.WaitForStateAvailable(client, 1200, d.Id()); err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[DEBUG] Successfully updated instance %s flavor: %s", d.Id(), d.Get("flavor").(string))
	}

	if d.HasChange("volume.0.size") {
		_, newVolume := d.GetChange("volume")
		volume := make(map[string]interface{})
		volumeRaw := newVolume.([]interface{})
		log.Printf("[DEBUG] volumeRaw: %+v", volumeRaw)
		if len(volumeRaw) == 1 {
			if m, ok := volumeRaw[0].(map[string]interface{}); ok {
				volume["size"] = m["size"].(int)
			}
		}
		log.Printf("[DEBUG] volume: %+v", volume)
		updateOpts := instances.EnlargeVolumeRdsOpts{
			InstanceId: d.Id(),
			Size:       volume["size"].(int),
		}

		log.Printf("Update volume size could be done only in status `available`")
		if err := instances.WaitForStateAvailable(client, 1200, d.Id()); err != nil {
			return diag.FromErr(err)
		}

		updateResult, err := instances.EnlargeVolume(client, updateOpts)
		if err != nil {
			return fmterr.Errorf("error updating instance volume from result: %s", err)
		}
		timeout := d.Timeout(schema.TimeoutUpdate)
		if err := instances.WaitForJobCompleted(client, int(timeout.Seconds()), *updateResult); err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[DEBUG] Successfully updated instance %s volume: %+v", d.Id(), volume)
	}

	if d.HasChange("public_ips") {
		nwClient, err := config.NetworkingV2Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating networking V2 client: %w", err)
		}
		oldPublicIps, newPublicIps := d.GetChange("public_ips")
		oldIPs := oldPublicIps.([]interface{})
		newIPs := newPublicIps.([]interface{})
		switch len(newIPs) {
		case 0:
			err := unAssignEipFromInstance(nwClient, oldIPs[0].(string)) // if it becomes 0, it was 1 before
			if err != nil {
				return diag.FromErr(err)
			}
		case 1:
			if len(oldIPs) > 0 {
				err := unAssignEipFromInstance(nwClient, oldIPs[0].(string))
				if err != nil {
					return diag.FromErr(err)
				}
			}
			privateIP := getPrivateIP(d)
			subnetID, err := getSubnetSubnetID(d, config)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := assignEipToInstance(nwClient, newIPs[0].(string), privateIP, subnetID); err != nil {
				return diag.FromErr(err)
			}
		default:
			return fmterr.Errorf("RDS instance can't have more than one public IP")
		}
	}

	var restartRequired bool

	if d.HasChange("param_group_id") {
		newParamGroupID := d.Get("param_group_id").(string)
		if len(newParamGroupID) == 0 {
			return fmterr.Errorf("you can't remove `param_group_id` without recreation")
		}
		templateRestart, err := applyTemplate(d, client)
		if err != nil {
			return fmterr.Errorf("error applying parameter template: %w", err)
		}
		restartRequired = restartRequired || templateRestart
	}

	if d.HasChange("parameters") {
		paramRestart, err := updateInstanceParameters(d, client)
		if err != nil {
			return fmterr.Errorf("error applying parameters to the instance: %w", err)
		}
		restartRequired = restartRequired || paramRestart
	}

	if d.HasChange("db.0.port") {
		udpateOpts := security.UpdatePortOpts{
			Port:       int32(d.Get("db.0.port").(int)),
			InstanceId: d.Id(),
		}
		log.Printf("[DEBUG] Update opts of Database port: %+v", udpateOpts)
		_, err := security.UpdatePort(client, udpateOpts)
		if err != nil {
			return fmterr.Errorf("error updating instance database port: %s ", err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:      []string{"MODIFYING DATABASE PORT"},
			Target:       []string{"ACTIVE"},
			Refresh:      rdsInstanceStateRefreshFunc(client, d.Id()),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			Delay:        5 * time.Second,
			PollInterval: 10 * time.Second,
		}
		if _, err = stateConf.WaitForStateContext(ctx); err != nil {
			return fmterr.Errorf("error waiting for RDS instance (%s) creation completed: %s", d.Id(), err)
		}
		restartRequired = true
	}

	err = instances.WaitForStateAvailable(client, 1200, d.Id())
	if err != nil {
		return fmterr.Errorf("error waiting for instance to become available: %w", err)
	}

	if restartRequired {
		if err := restartInstance(d, client); err != nil {
			return diag.FromErr(err)
		}
		waitSeconds := int(d.Timeout(schema.TimeoutUpdate).Seconds())
		if err := instances.WaitForStateAvailable(client, waitSeconds, d.Id()); err != nil {
			return fmterr.Errorf("error waiting for instance to become available: %w", err)
		}
	}

	if d.HasChange("ssl_enable") {
		if dbType := d.Get("db.0.type").(string); strings.ToLower(dbType) == "mysql" {
			updateOpts := security.SwitchSslOpts{
				SslOption:  d.Get("ssl_enable").(bool),
				InstanceId: d.Id(),
			}
			log.Printf("[DEBUG] Update opts of SSL configuration: %+v", updateOpts)
			err := security.SwitchSsl(client, updateOpts)
			if err != nil {
				return fmterr.Errorf("error updating instance SSL configuration: %s ", err)
			}
			return nil
		} else {
			return diag.Errorf("only MySQL database support SSL enable and disable")
		}
	}

	if err = updateVolumeAutoExpand(ctx, d, client, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceRdsInstanceV3Read(clientCtx, d, meta)
}

func getMasterID(nodes []instances.Nodes) (nodeID string) {
	for _, node := range nodes {
		if node.Role == "master" {
			nodeID = node.Id
		}
	}
	return
}

func resourceRdsInstanceV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.RdsV3Client(region)
	})
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	rdsInstance, err := GetRdsInstance(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error fetching RDS instance: %s", err)
	}
	if rdsInstance == nil {
		d.SetId("")
		return nil
	}

	me := multierror.Append(
		d.Set("flavor", rdsInstance.FlavorRef),
		d.Set("name", rdsInstance.Name),
		d.Set("security_group_id", rdsInstance.SecurityGroupId),
		d.Set("subnet_id", rdsInstance.SubnetId),
		d.Set("vpc_id", rdsInstance.VpcId),
		d.Set("created", rdsInstance.Created),
		d.Set("ha_replication_mode", rdsInstance.Ha.ReplicationMode),
		d.Set("lower_case_table_names", d.Get("lower_case_table_names").(string)),
		d.Set("ssl_enable", *rdsInstance.EnableSSL),
	)

	if v, ok := d.GetOk("restore_from_backup"); ok {
		rawPitr := v.([]interface{})
		if len(rawPitr) > 0 {
			backupId := rawPitr[0].(map[string]interface{})["backup_id"]
			if backupId != "" {
				me = multierror.Append(me, d.Set("restored_backup_id", backupId))
			}
		}
	}

	if me.ErrorOrNil() != nil {
		return diag.FromErr(me)
	}

	var nodesList []map[string]interface{}
	for _, nodeObj := range rdsInstance.Nodes {
		node := make(map[string]interface{})
		node["id"] = nodeObj.Id
		node["role"] = nodeObj.Role
		node["name"] = nodeObj.Name
		node["availability_zone"] = nodeObj.AvailabilityZone
		node["status"] = nodeObj.Status
		nodesList = append(nodesList, node)
	}

	if err := d.Set("nodes", nodesList); err != nil {
		return fmterr.Errorf("error setting node list: %s", err)
	}

	var availabilityZones []string
	switch n := len(rdsInstance.Nodes); n {
	case 1:
		availabilityZones = []string{
			rdsInstance.Nodes[0].AvailabilityZone,
		}
	case 2:
		if rdsInstance.Nodes[0].Role == "master" {
			availabilityZones = []string{
				rdsInstance.Nodes[0].AvailabilityZone,
				rdsInstance.Nodes[1].AvailabilityZone,
			}
		} else {
			availabilityZones = []string{
				rdsInstance.Nodes[1].AvailabilityZone,
				rdsInstance.Nodes[0].AvailabilityZone,
			}
		}
	default:
		fmterr.Errorf("RDSv3 instance expects 1 or 2 nodes, but got %d", n)
	}

	if err := d.Set("availability_zones", availabilityZones); err != nil {
		return diag.FromErr(err)
	}

	var backupStrategyList []map[string]interface{}
	backupStrategy := make(map[string]interface{})
	backupStrategy["keep_days"] = rdsInstance.BackupStrategy.KeepDays
	// if `keep_days` is set to 0 backend returns empty `start_time`
	// which forces instance update
	if backupStrategy["keep_days"] != 0 {
		backupStrategy["start_time"] = rdsInstance.BackupStrategy.StartTime
	} else {
		backupRaw := resourceRDSBackupStrategy(d)
		backupStrategy["start_time"] = backupRaw.StartTime
	}
	backupStrategyList = append(backupStrategyList, backupStrategy)
	if err := d.Set("backup_strategy", backupStrategyList); err != nil {
		return fmterr.Errorf("error setting backup strategy: %s", err)
	}

	dbRaw := d.Get("db").([]interface{})
	dbInfo := make(map[string]interface{})
	if len(dbRaw) != 0 {
		dbInfo = dbRaw[0].(map[string]interface{})
	}
	dbInfo["type"] = rdsInstance.DataStore.Type
	// backwards compatibility for minor versions on Swiss
	if region != "eu-ch2" || (region == "eu-ch2" && checkMinorVersion(dbInfo)) {
		dbInfo["version"] = rdsInstance.DataStore.Version
	}
	dbInfo["port"] = rdsInstance.Port
	dbInfo["user_name"] = rdsInstance.DbUserName
	dbList := []interface{}{dbInfo}
	if err = d.Set("db", dbList); err != nil {
		return diag.FromErr(err)
	}

	var volumeList []map[string]interface{}
	volume := make(map[string]interface{})
	volume["size"] = rdsInstance.Volume.Size
	volume["type"] = rdsInstance.Volume.Type
	volume["disk_encryption_id"] = rdsInstance.DiskEncryptionId

	if region != "eu-ch2" {
		resp, err := instances.GetAutoScaling(client, d.Id())
		if err != nil {
			log.Printf("[ERROR] error query automatic expansion configuration of the instance storage: %s", err)
		} else if resp.SwitchOption {
			err = d.Set("autoscaling_enabled", true)
			if err != nil {
				return diag.FromErr(err)
			}
			volume["limit_size"] = resp.LimitSize
			volume["trigger_threshold"] = resp.TriggerThreshold
		}
	}

	// in case when autoscaling was enabled even once size will be overwritten by data from schema
	if as, ok := d.GetOk("autoscaling_enabled"); ok && as.(bool) {
		volume["size"] = d.Get("volume.0.size").(int)
	}

	volumeList = append(volumeList, volume)
	if err = d.Set("volume", volumeList); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("private_ips", rdsInstance.PrivateIps); err != nil {
		return diag.FromErr(err)
	}

	publicIp := getPublicIP(d)
	if publicIp != "" {
		if err = d.Set("public_ips", []string{publicIp}); err != nil {
			return diag.FromErr(err)
		}
	}

	var tagParamName string
	// set instance tags
	if _, ok := d.GetOk("tags"); ok {
		tagParamName = "tags"
	} else if _, ok := d.GetOk("tag"); ok {
		tagParamName = "tag"
	}
	if tagParamName == "tag" {
		// set instance tag
		var nodeID string
		nodes := d.Get("nodes").([]interface{})
		for _, node := range nodes {
			nodeObj := node.(map[string]interface{})
			if nodeObj["role"].(string) == "master" {
				nodeID = nodeObj["id"].(string)
			}
		}

		if nodeID == "" {
			log.Printf("[WARN] Error fetching node id of instance: %s", d.Id())
			return nil
		}
		tagClient, err := config.RdsTagV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud rds tag client: %#v", err)
		}
		tagList, err := tag.Get(tagClient, nodeID).Extract()
		if err != nil {
			return fmterr.Errorf("error fetching OpenTelekomCloud rds instance tags: %s", err)
		}
		tagMap := make(map[string]string)
		for _, val := range tagList.Tags {
			tagMap[val.Key] = val.Value
		}
		if err := d.Set("tag", tagMap); err != nil {
			return fmterr.Errorf("[DEBUG] Error saving tag to state for OpenTelekomCloud rds instance (%s): %s", d.Id(), err)
		}
	} else if tagParamName == "tags" {
		tagsMap := common.TagsToMap(rdsInstance.Tags)
		if err := d.Set("tags", tagsMap); err != nil {
			return fmterr.Errorf("error saving tags for OpenTelekomCloud RDSv3 instance: %s", err)
		}
	}

	return nil
}

func resourceRdsInstanceV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.RdsV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	log.Printf("[DEBUG] Deleting Instance %s", d.Id())

	_, err = instances.Delete(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud RDSv3 instance: %s", err)
	}

	d.SetId("")
	return nil
}

//nolint:all
func validateRDSv3Flavor(argName string) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		config, ok := meta.(*cfg.Config)
		if !ok {
			return fmt.Errorf("error retreiving configuration: can't convert %v to Config", meta)
		}

		client, err := config.RdsV3Client(config.GetRegion(d))
		if err != nil {
			return fmt.Errorf(errCreateClient, err)
		}
		dataStoreInfo := d.Get("db").([]interface{})[0].(map[string]interface{})
		flavor := d.Get(argName).(string)

		listOpts := flavors.ListOpts{
			VersionName:  dataStoreInfo["version"].(string),
			DatabaseName: dataStoreInfo["type"].(string),
		}
		flavorList, err := flavors.ListFlavors(client, listOpts)
		if err != nil {
			return fmt.Errorf("unable to get flavor pages: %w", err)
		}

		var matches = false
		for _, flavorItem := range flavorList {
			if flavorItem.SpecCode == flavor {
				matches = true
				break
			}
		}

		if !matches {
			return fmt.Errorf("can't find flavor `%s`", flavor)
		}

		return nil
	}
}

func updateVolumeAutoExpand(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient,
	instanceID string) error {
	if !d.HasChanges("volume.0.limit_size", "volume.0.trigger_threshold") {
		return nil
	}

	limitSize := d.Get("volume.0.limit_size").(int)
	if limitSize > 0 {
		if err := enableVolumeAutoExpand(ctx, d, client, instanceID, limitSize); err != nil {
			return err
		}
	} else {
		if err := disableVolumeAutoExpand(ctx, d.Timeout(schema.TimeoutUpdate), client, instanceID); err != nil {
			return err
		}
	}
	return nil
}

func disableVolumeAutoExpand(ctx context.Context, timeout time.Duration, client *golangsdk.ServiceClient,
	instanceID string) error {
	retryFunc := func() (interface{}, bool, error) {
		err := instances.ManageAutoScaling(client, instanceID, instances.ScalingOpts{
			SwitchOption: false,
		})
		retry, err := handleMultiOperationsError(err)
		return nil, retry, err
	}
	_, err := common.RetryContextWithWaitForState(&common.RetryContextWithWaitForStateParam{
		Ctx:          ctx,
		RetryFunc:    retryFunc,
		WaitFunc:     rdsInstanceStateRefreshFunc(client, instanceID),
		WaitTarget:   []string{"ACTIVE"},
		Timeout:      timeout,
		DelayTimeout: 10 * time.Second,
		PollInterval: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("an error occurred while disable automatic expansion of instance storage: %v", err)
	}
	return nil
}

func updateInstanceParameters(d *schema.ResourceData, client *golangsdk.ServiceClient) (bool, error) {
	opts := configurations.UpdateInstanceConfigurationOpts{
		Values:     d.Get("parameters").(map[string]interface{}),
		InstanceId: d.Id(),
	}
	_, err := configurations.UpdateInstanceConfiguration(client, opts)
	if err != nil {
		return false, err
	}
	return true, err
}

func rdsInstanceStateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := GetRdsInstance(client, instanceID)
		if err != nil {
			return nil, "Error retrieving RDSv3 Instance", err
		}
		if instance.Id == "" {
			return instance, "DELETED", nil
		}

		return instance, instance.Status, nil
	}
}

func waitForParameterApply(d *schema.ResourceData, client *golangsdk.ServiceClient) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := updateInstanceParameters(d, client)

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault403); ok {
				return r, "PENDING", nil
			}
			return nil, "", fmt.Errorf("error applying configuration parameters: %w", err)
		}

		return r, "SUCCESS", nil
	}
}

func waitForSSLEnable(d *schema.ResourceData, client *golangsdk.ServiceClient) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rdsInstance, err := GetRdsInstance(client, d.Id())
		if err != nil {
			return nil, "", fmt.Errorf("error fetching RDS instance SSL status: %s", err)
		}

		if *rdsInstance.EnableSSL {
			return rdsInstance, "SUCCESS", nil
		}

		return nil, "PENDING", nil
	}
}

func enableVolumeAutoExpand(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient,
	instanceID string, limitSize int) error {
	opts := instances.ScalingOpts{
		LimitSize:        pointerto.Int(limitSize),
		TriggerThreshold: pointerto.Int(d.Get("volume.0.trigger_threshold").(int)),
		SwitchOption:     true,
	}
	retryFunc := func() (interface{}, bool, error) {
		err := instances.ManageAutoScaling(client, d.Id(), opts)
		retry, err := handleMultiOperationsError(err)
		return nil, retry, err
	}
	_, err := common.RetryContextWithWaitForState(&common.RetryContextWithWaitForStateParam{
		Ctx:          ctx,
		RetryFunc:    retryFunc,
		WaitFunc:     rdsInstanceStateRefreshFunc(client, instanceID),
		WaitTarget:   []string{"ACTIVE"},
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		DelayTimeout: 10 * time.Second,
		PollInterval: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("an error occurred while enable automatic expansion of instance storage: %v", err)
	}
	return nil
}
