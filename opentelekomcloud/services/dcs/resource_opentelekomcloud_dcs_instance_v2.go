package dcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmespath/go-jmespath"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v2/configs"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v2/instance"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v2/others"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v2/ssl"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v2/whitelists"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDcsInstanceV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDcsInstancesV2Create,
		ReadContext:   resourceDcsInstancesV2Read,
		UpdateContext: resourceDcsInstancesV2Update,
		DeleteContext: resourceDcsInstancesV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Redis",
				}, true),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"capacity": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
			},

			"availability_zones": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"whitelist"},
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"access_user": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			"enable_whitelist": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"whitelist": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip_list": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ssl_enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"maintain_begin": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"maintain_end"},
			},
			"maintain_end": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"backup_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"save_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 7),
						},
						"backup_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "manual"}, false),
						},
						"begin_at": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(
								regexp.MustCompile(`^([0-1]\d|2[0-3]):00-([0-1]\d|2[0-3]):00$`),
								"format must be HH:00-HH:00",
							),
						},
						"period_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "weekly",
							ValidateFunc: validation.StringInSlice([]string{"weekly"}, false),
						},
						"backup_at": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeInt,
								ValidateFunc: validation.IntBetween(1, 7),
							},
						},
					},
				},
			},
			"rename_commands": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"template_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"parameters": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},
			"tags": common.TagsSchema(),
			"deleted_nodes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MaxItems: 1,
			},
			"reserved_ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"used_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launched_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth_info": {
				Type:     schema.TypeList,
				Elem:     bandwidthSchema(),
				Computed: true,
			},
			"cache_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replica_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"readonly_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transparent_client_ip_enable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"product_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sharding_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func bandwidthSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"bandwidth": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"begin_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"current_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expand_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"expand_effect_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"expand_interval_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_expand_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"next_expand_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"task_running": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
	return &sc
}

func buildBackupPolicyParams(d *schema.ResourceData) *instance.InstanceBackupPolicyOpts {
	backupPolicyList := d.Get("backup_policy").([]interface{})
	if len(backupPolicyList) == 0 {
		return nil
	}
	backupPolicy := backupPolicyList[0].(map[string]interface{})
	backupType := backupPolicy["backup_type"].(string)
	if len(backupType) == 0 || backupType == "manual" {
		return nil
	}
	// build backup policy options
	backupAt := common.ExpandToIntList(backupPolicy["backup_at"].([]interface{}))
	backupPlan := instance.BackupPlan{
		BackupAt:   backupAt,
		PeriodType: backupPolicy["period_type"].(string),
		BeginAt:    backupPolicy["begin_at"].(string),
	}
	backupPolicyOpts := &instance.InstanceBackupPolicyOpts{
		BackupType:           backupPolicy["backup_type"].(string),
		SaveDays:             backupPolicy["save_days"].(int),
		PeriodicalBackupPlan: &backupPlan,
	}
	return backupPolicyOpts
}

func resourceDcsInstancesCheck(d *schema.ResourceData) error {
	engineVersion := d.Get("engine_version").(string)
	secGroupID := d.Get("security_group_id").(string)

	if _, ok := redisEngineVersion[engineVersion]; ok {
		if secGroupID != "" {
			return fmt.Errorf("security_group_id is not supported for Redis 4.0, 5.0 and 6.0. " +
				"please configure the whitelists alternatively")
		}
	} else if engineVersion == "3.0" {
		if secGroupID == "" {
			return fmt.Errorf("security_group_id is mandatory for this DCS instance")
		}
	}

	return nil
}

func buildDcsTagsParams(tagsMap map[string]interface{}) []tags.ResourceTag {
	tagArr := make([]tags.ResourceTag, 0, len(tagsMap))
	for k, v := range tagsMap {
		tag := tags.ResourceTag{
			Key:   k,
			Value: v.(string),
		}
		tagArr = append(tagArr, tag)
	}
	return tagArr
}

func buildWhiteListParams(d *schema.ResourceData) whitelists.WhitelistOpts {
	enable := d.Get("enable_whitelist").(bool)
	groupList := d.Get("whitelist").(*schema.Set).List()

	groups := make([]whitelists.WhitelistGroupOpts, len(groupList))
	for i, v := range groupList {
		item := v.(map[string]interface{})
		groups[i] = whitelists.WhitelistGroupOpts{
			GroupName: item["group_name"].(string),
			IPList:    common.ExpandToStringList(item["ip_list"].([]interface{})),
		}
	}

	whitelistOpts := whitelists.WhitelistOpts{
		Enable: &enable,
		Groups: groups,
	}
	return whitelistOpts
}

func buildSslParam(enable bool) ssl.SslOpts {
	sslOpts := ssl.SslOpts{
		Enabled: &enable,
	}
	return sslOpts
}

func waitForWhiteListCompleted(ctx context.Context, c *golangsdk.ServiceClient, d *schema.ResourceData) error {
	enable := d.Get("enable_whitelist").(bool)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{strconv.FormatBool(!enable)},
		Target:       []string{strconv.FormatBool(enable)},
		Refresh:      refreshForWhiteListEnableStatus(c, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceDcsInstancesV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dcsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DcsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	if err = resourceDcsInstancesCheck(d); err != nil {
		return diag.FromErr(err)
	}

	// noPasswordAccess
	noPasswordAccess := true
	if d.Get("access_user").(string) != "" || d.Get("password").(string) != "" {
		noPasswordAccess = false
	}

	// azCodes
	var azCodes []string
	availabilityZones := d.Get("availability_zones")

	azCodes = common.ExpandToStringList(availabilityZones.([]interface{}))

	createOpts := instance.CreateOpts{
		Name:             d.Get("name").(string),
		Engine:           d.Get("engine").(string),
		EngineVersion:    d.Get("engine_version").(string),
		Capacity:         d.Get("capacity").(float64),
		InstanceNum:      1,
		SpecCode:         d.Get("flavor").(string),
		AzCodes:          azCodes,
		Port:             d.Get("port").(int),
		VpcId:            d.Get("vpc_id").(string),
		SubnetId:         d.Get("subnet_id").(string),
		SecurityGroupId:  d.Get("security_group_id").(string),
		Description:      d.Get("description").(string),
		PrivateIp:        d.Get("private_ip").(string),
		MaintainBegin:    d.Get("maintain_begin").(string),
		MaintainEnd:      d.Get("maintain_end").(string),
		NoPasswordAccess: &noPasswordAccess,
		AccessUser:       d.Get("access_user").(string),
		TemplateId:       d.Get("template_id").(string),
		Tags:             buildDcsTagsParams(d.Get("tags").(map[string]interface{})),
	}

	renameCmds := d.Get("rename_commands").(map[string]interface{})
	if len(renameCmds) > 0 {
		createOpts.RenameCommands = createRenameCommandsOpt(renameCmds)
	}

	backupPolicy := buildBackupPolicyParams(d)
	if backupPolicy != nil {
		createOpts.BackupPolicy = backupPolicy
	}
	log.Printf("[DEBUG] Create DCS instance options(hide password) : %#v", createOpts)

	createOpts.Password = d.Get("password").(string)

	r, err := instance.Create(client, createOpts)
	if err != nil || len(r) == 0 {
		return diag.Errorf("error in creating DCS instance : %s", err)
	}
	id := r[0].InstanceID
	d.SetId(id)

	err = waitForDcsInstanceCompleted(ctx, client, id, d.Timeout(schema.TimeoutCreate),
		[]string{"CREATING"}, []string{"RUNNING"})
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("whitelist").(*schema.Set).Len() > 0 {
		whitelistOpts := buildWhiteListParams(d)
		log.Printf("[DEBUG] Create whitelist options: %#v", whitelistOpts)

		err = whitelists.Put(client, id, whitelistOpts)
		if err != nil {
			return diag.Errorf("error creating whitelist for DCS instance (%s): %s", id, err)
		}

		err = waitForWhiteListCompleted(ctx, client, d)
		if err != nil {
			return diag.Errorf("Error while waiting to create DCS whitelist: %s", err)
		}
	}

	if v, ok := d.GetOk("parameters"); ok {
		parameters := v.(*schema.Set).List()
		err = updateParameters(ctx, d.Timeout(schema.TimeoutCreate), client, id, parameters)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if sslEnabled := d.Get("ssl_enable").(bool); sslEnabled {
		sslOpts := buildSslParam(sslEnabled)
		sslOpts.InstanceId = id
		_, err := ssl.Update(client, sslOpts)
		if err != nil {
			return diag.Errorf("error updating SSL for the instance (%s): %s", id, err)
		}

		err = waitForSslCompleted(ctx, client, d)
		if err != nil {
			return diag.Errorf("error waiting for updating SSL to complete: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, dcsClientV2)
	return resourceDcsInstancesV2Read(clientCtx, d, meta)
}

func updateParameters(ctx context.Context, timeout time.Duration, client *golangsdk.ServiceClient, instanceID string,
	parameters []interface{}) error {
	parameterOpts := buildUpdateParametersOpt(parameters)
	parameterOpts.InstanceId = instanceID
	retryFunc := func() (interface{}, bool, error) {
		log.Printf("[DEBUG] Update DCS instance parameters params: %#v", parameterOpts)
		err := configs.Update(client, parameterOpts)
		retry, err := handleOperationError(err)
		return nil, retry, err
	}
	_, err := common.RetryContextWithWaitForState(&common.RetryContextWithWaitForStateParam{
		Ctx:          ctx,
		RetryFunc:    retryFunc,
		WaitFunc:     refreshDcsInstanceState(client, instanceID),
		WaitTarget:   []string{"RUNNING"},
		Timeout:      timeout,
		DelayTimeout: 1 * time.Second,
		PollInterval: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("error modifying parameters for DCS instance (%s): %s", instanceID, err)
	}
	return nil
}

func buildUpdateParametersOpt(parameters []interface{}) configs.ModifyConfigOpt {
	parameterOpts := make([]configs.RedisConfigs, 0, len(parameters))
	for _, parameter := range parameters {
		if v, ok := parameter.(map[string]interface{}); ok {
			parameterOpts = append(parameterOpts, configs.RedisConfigs{
				ParamID:    v["id"].(string),
				ParamName:  v["name"].(string),
				ParamValue: v["value"].(string),
			})
		}
	}
	return configs.ModifyConfigOpt{RedisConfig: parameterOpts}
}

func getParameters(client *golangsdk.ServiceClient, instanceID string, parameters []interface{}) ([]map[string]interface{},
	error) {
	configParameters, err := configs.Get(client, instanceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching the DCS instance parameters (%s): %s", instanceID, err)
	}
	parametersMap := generateParametersMap(configParameters)
	var params []map[string]interface{}
	for _, parameter := range parameters {
		paramId := parameter.(map[string]interface{})["id"]
		if v, ok := parametersMap[paramId.(string)]; ok {
			params = append(params, map[string]interface{}{
				"id":    v.ParamID,
				"name":  v.ParamName,
				"value": v.ParamValue,
			})
		}
	}
	return params, nil
}

func createRenameCommandsOpt(renameCmds map[string]interface{}) instance.RenameCommand {
	renameCommands := instance.RenameCommand{}
	if v, ok := renameCmds["command"]; ok {
		renameCommands.Command = v.(string)
	}
	if v, ok := renameCmds["keys"]; ok {
		renameCommands.Keys = v.(string)
	}
	if v, ok := renameCmds["flushdb"]; ok {
		renameCommands.Flushdb = v.(string)
	}
	if v, ok := renameCmds["flushall"]; ok {
		renameCommands.Flushdb = v.(string)
	}
	if v, ok := renameCmds["hgetall"]; ok {
		renameCommands.Hgetall = v.(string)
	}
	return renameCommands
}

func waitForDcsInstanceCompleted(ctx context.Context, c *golangsdk.ServiceClient, id string, timeout time.Duration,
	padding []string, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   padding,
		Target:                    target,
		Refresh:                   refreshDcsInstanceState(c, id),
		Timeout:                   timeout,
		Delay:                     10 * time.Second,
		PollInterval:              10 * time.Second,
		ContinuousTargetOccurence: 2,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("[DEBUG] error while waiting to create/resize/delete DCS instance. %s : %v",
			id, err)
	}
	return nil
}

func refreshDcsInstanceState(c *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := instance.Get(c, id)
		if err != nil {
			err404 := golangsdk.ErrDefault404{}
			if errors.As(err, &err404) {
				return &(instance.DcsInstance{}), "DELETED", nil
			}
			return nil, "Error", err
		}
		return r, r.Status, nil
	}
}

func resourceDcsInstancesV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dcsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DcsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	r, err := instance.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "DCS instance")
	}
	log.Printf("[DEBUG] Get DCS instance : %#v", r)

	capacity := r.Capacity
	if capacity == 0 {
		capacity, _ = strconv.ParseFloat(r.CapacityMinor, floatBitSize)
	}

	securityGroupID := r.SecurityGroupId

	if securityGroupID == "securityGroupId" {
		securityGroupID = ""
	}

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("name", r.Name),
		d.Set("engine", r.Engine),
		d.Set("engine_version", r.EngineVersion),
		d.Set("capacity", capacity),
		d.Set("flavor", r.SpecCode),
		d.Set("availability_zones", r.AzCodes),
		d.Set("vpc_id", r.VpcId),
		d.Set("vpc_name", r.VpcName),
		d.Set("subnet_id", r.SubnetId),
		d.Set("subnet_name", r.SubnetName),
		d.Set("subnet_cidr", r.SubnetCidr),
		d.Set("security_group_id", securityGroupID),
		d.Set("security_group_name", r.SecurityGroupName),
		d.Set("description", r.Description),
		d.Set("private_ip", r.Ip),
		d.Set("maintain_begin", r.MaintainBegin),
		d.Set("maintain_end", r.MaintainEnd),
		d.Set("port", r.Port),
		d.Set("status", r.Status),
		d.Set("used_memory", r.UsedMemory),
		d.Set("max_memory", r.MaxMemory),
		d.Set("domain_name", r.DomainName),
		d.Set("user_id", r.UserId),
		d.Set("user_name", r.UserName),
		d.Set("access_user", r.AccessUser),
		d.Set("ssl_enable", r.EnableSsl),
		d.Set("created_at", r.CreatedAt),
		d.Set("launched_at", r.LaunchedAt),
		d.Set("cache_mode", r.CacheMode),
		d.Set("cpu_type", r.CpuType),
		d.Set("readonly_domain_name", r.ReadOnlyDomainName),
		d.Set("replica_count", r.ReplicaCount),
		d.Set("transparent_client_ip_enable", r.TransparentClientIpEnable),
		d.Set("bandwidth_info", setBandWidthInfo(&r.BandWidthDetail)),
		d.Set("product_type", r.ProductType),
		d.Set("sharding_count", r.ShardingCount),
	)

	if mErr.ErrorOrNil() != nil {
		return diag.Errorf("error setting DCS instance attributes: %s", mErr)
	}

	backupPolicy := r.BackupPolicy
	if len(backupPolicy.Policy.BackupType) > 0 {
		bakPolicy := []map[string]interface{}{
			{
				"backup_type": backupPolicy.Policy.BackupType,
				"save_days":   backupPolicy.Policy.SaveDays,
				"begin_at":    backupPolicy.Policy.PeriodicalBackupPlan.BeginAt,
				"period_type": backupPolicy.Policy.PeriodicalBackupPlan.PeriodType,
				"backup_at":   backupPolicy.Policy.PeriodicalBackupPlan.BackupAt,
			},
		}
		mErr = multierror.Append(mErr, d.Set("backup_policy", bakPolicy))
	}

	// set tags
	if resourceTags, err := tags.Get(client, "instances", d.Id()).Extract(); err == nil {
		tagMap := common.TagsToMap(resourceTags)
		if err := d.Set("tags", tagMap); err != nil {
			return diag.Errorf("[DEBUG] error saving tag to state for DCS instance (%s): %s", d.Id(), err)
		}
	} else {
		log.Printf("[WARN] fetching tags of DCS instance failed: %s", err)
	}

	wList, err := whitelists.Get(client, d.Id())
	if err != nil || wList == nil || len(wList.Groups) == 0 {
		log.Printf("error fetching whitelists for DCS instance, error: %s", err)
		mErr = multierror.Append(
			mErr,
			d.Set("enable_whitelist", true),
		)
		return diag.FromErr(mErr.ErrorOrNil())
	}

	log.Printf("[DEBUG] Find DCS instance white list : %#v", wList.Groups)
	whiteList := make([]map[string]interface{}, len(wList.Groups))
	for i, group := range wList.Groups {
		whiteList[i] = map[string]interface{}{
			"group_name": group.GroupName,
			"ip_list":    group.IPList,
		}
	}
	mErr = multierror.Append(
		mErr,
		d.Set("whitelist", whiteList),
		d.Set("enable_whitelist", wList.Enable),
	)

	diagErr := setDcsInstanceParameters(d, client, d.Id())
	return append(diagErr, diag.FromErr(mErr.ErrorOrNil())...)
}

func setDcsInstanceParameters(d *schema.ResourceData, client *golangsdk.ServiceClient,
	instanceID string) diag.Diagnostics {
	params, err := getParameters(client, instanceID, d.Get("parameters").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	if len(params) > 0 {
		if err = d.Set("parameters", params); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func generateParametersMap(configurations *configs.ConfigParam) map[string]configs.RedisConfigResult {
	parametersMap := make(map[string]configs.RedisConfigResult)
	for _, redisConfig := range configurations.RedisConfigs {
		parametersMap[redisConfig.ParamID] = redisConfig
	}
	return parametersMap
}

func resourceDcsInstancesV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dcsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DcsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	if d.HasChanges("port", "name", "description", "security_group_id", "backup_policy",
		"maintain_begin", "maintain_end", "rename_commands") {
		desc := d.Get("description").(string)
		securityGroupID := d.Get("security_group_id").(string)
		renameCommandsOpt := createRenameCommandsOpt(d.Get("rename_commands").(map[string]interface{}))
		opts := instance.ModifyInstanceOpt{
			InstanceId:      d.Id(),
			Name:            d.Get("name").(string),
			Port:            pointerto.Int(d.Get("port").(int)),
			Description:     &desc,
			MaintainBegin:   d.Get("maintain_begin").(string),
			MaintainEnd:     d.Get("maintain_end").(string),
			SecurityGroupId: &securityGroupID,
			BackupPolicy:    buildBackupPolicyParams(d),
			RenameCommands:  &renameCommandsOpt,
		}
		log.Printf("[DEBUG] Update DCS instance options : %#v", opts)

		err = instance.Update(client, opts)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("port") {
			err = waitForPortUpdated(ctx, client, d)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("password") {
		oldVal, newVal := d.GetChange("password")
		opts := instance.UpdatePasswordOpts{
			InstanceId:  d.Id(),
			OldPassword: oldVal.(string),
			NewPassword: newVal.(string),
		}
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err = instance.UpdatePassword(client, opts)
			isRetry, err := handleOperationError(err)
			if isRetry {
				return resource.RetryableError(err)
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = resizeDcsInstance(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("tags") {
		oldVal, newVal := d.GetChange("tags")
		err = updateDcsTags(client, d.Id(), oldVal.(map[string]interface{}), newVal.(map[string]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("whitelist", "enable_whitelist") {
		whitelistOpts := buildWhiteListParams(d)
		log.Printf("[DEBUG] Update DCS instance whitelist options: %#v", whitelistOpts)

		err = whitelists.Put(client, d.Id(), whitelistOpts)
		if err != nil {
			return diag.Errorf("error updating whitelist for instance (%s): %s", d.Id(), err)
		}

		err = waitForWhiteListCompleted(ctx, client, d)
		if err != nil {
			return diag.Errorf("error while waiting to create DCS whitelist: %s", err)
		}
	}

	if d.HasChange("parameters") {
		oRaw, nRaw := d.GetChange("parameters")
		changedParameters := nRaw.(*schema.Set).Difference(oRaw.(*schema.Set)).List()
		err = updateParameters(ctx, d.Timeout(schema.TimeoutUpdate), client, d.Id(), changedParameters)
		if err != nil {
			return diag.FromErr(err)
		}
		ctx = context.WithValue(ctx, ctxType("parametersChanged"), "true")
	}

	if d.HasChange("ssl_enable") {
		sslOpts := buildSslParam(d.Get("ssl_enable").(bool))
		sslOpts.InstanceId = d.Id()
		_, err = ssl.Update(client, sslOpts)
		if err != nil {
			return diag.Errorf("error updating SSL for the instance (%s): %s", d.Id(), err)
		}

		// wait for SSL updated
		err = waitForSslCompleted(ctx, client, d)
		if err != nil {
			return diag.Errorf("error waiting for updating SSL to complete: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, dcsClientV2)
	return resourceDcsInstancesV2Read(clientCtx, d, meta)
}

func waitForPortUpdated(ctx context.Context, c *golangsdk.ServiceClient, d *schema.ResourceData) error {
	op, np := d.GetChange("port")
	stateConf := &resource.StateChangeConf{
		Pending:      []string{strconv.Itoa(op.(int))},
		Target:       []string{strconv.Itoa(np.(int))},
		Refresh:      refreshDcsInstancePort(c, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("[DEBUG] error while waiting to create/resize/delete DCS instance. %s : %#v",
			d.Id(), err)
	}
	return nil
}

func refreshDcsInstancePort(c *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := instance.Get(c, id)
		if err != nil {
			return nil, "Error", err
		}
		return r, strconv.Itoa(r.Port), nil
	}
}

func resizeDcsInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dcsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DcsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmt.Errorf(errCreationClient, err)
	}

	if d.HasChanges("flavor", "capacity") {
		oVal, nVal := d.GetChange("flavor")
		oldSpecCode := oVal.(string)
		newSpecCode := nVal.(string)
		opts, err := buildResizeInstanceOpt(client, d, oldSpecCode, newSpecCode)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Resize DCS dcsInstance options : %#v", *opts)

		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			err = instance.Resize(client, *opts)
			isRetry, err := handleOperationError(err)
			if isRetry {
				return resource.RetryableError(err)
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error resize DCS dcsInstance: %s", err)
		}

		err = waitForDcsInstanceCompleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate),
			[]string{"EXTENDING", "RESTARTING"}, []string{"RUNNING"})
		if err != nil {
			return err
		}

		dcsInstance, err := instance.Get(client, d.Id())
		if err != nil {
			return common.CheckDeleted(d, err, "DCS instance")
		}
		if dcsInstance.SpecCode != d.Get("flavor").(string) {
			return fmt.Errorf("change flavor failed, after changed the DCS flavor still is: %s, expected: %s",
				dcsInstance.SpecCode, newSpecCode)
		}
	}
	return nil
}

func buildResizeInstanceOpt(client *golangsdk.ServiceClient, d *schema.ResourceData, oldSpecCode,
	newSpecCode string) (*instance.ResizeInstanceOpts, error) {
	opts := instance.ResizeInstanceOpts{
		InstanceId:  d.Id(),
		SpecCode:    newSpecCode,
		NewCapacity: d.Get("capacity").(float64),
	}

	if oldSpecCode == newSpecCode {
		return nil, fmt.Errorf("the param flavor is invalid")
	}
	oldFlavor, err := getFlavorBySpecCode(client, oldSpecCode)
	if err != nil {
		return nil, err
	}
	newFlavor, err := getFlavorBySpecCode(client, newSpecCode)
	if err != nil {
		return nil, err
	}
	changeType := getFlavorChangeType(oldFlavor, newFlavor)
	opts.ChangeType = changeType
	if changeType == "createReplication" {
		azCodes, err := getAzCode(d, client)
		if err != nil {
			return nil, err
		}
		opts.AvailableZones = azCodes
	}
	if changeType == "deleteReplication" {
		if newFlavor.CacheMode == "ha" {
			opts.NodeList = common.ExpandToStringList(d.Get("deleted_nodes").([]interface{}))
		} else if newFlavor.CacheMode == "cluster" {
			azCodes, err := getAzCode(d, client)
			if err != nil {
				return nil, err
			}
			opts.ReservedIp = common.ExpandToStringList(d.Get("reserved_ips").([]interface{}))
			opts.AvailableZones = azCodes
		}
	}
	return &opts, nil
}

func getFlavorChangeType(oldFlavor, newFlavor *others.Product) string {
	if oldFlavor.CacheMode != newFlavor.CacheMode {
		return "instanceType"
	}
	if oldFlavor.ReplicaCount < newFlavor.ReplicaCount {
		return "createReplication"
	}
	if oldFlavor.ReplicaCount > newFlavor.ReplicaCount {
		return "deleteReplication"
	}
	return ""
}

func getFlavorBySpecCode(client *golangsdk.ServiceClient, specCode string) (*others.Product, error) {
	list, err := others.ListFlavors(client, others.ListFlavorOpts{SpecCode: specCode})
	if err != nil {
		return nil, fmt.Errorf("error getting dcs flavors list by specCode %s: %s", specCode, err)
	}
	if len(list) < 1 {
		return nil, fmt.Errorf("the result queried by specCode(%s) is empty", specCode)
	}
	return &list[0], nil
}

func handleOperationError(err error) (bool, error) {
	if err == nil {
		return false, nil
	}
	if errCode, ok := err.(golangsdk.ErrDefault400); ok {
		var apiError interface{}
		if jsonErr := json.Unmarshal(errCode.Body, &apiError); jsonErr != nil {
			return false, jsonErr
		}
		errorCode, errorCodeErr := jmespath.Search("error_code", apiError)
		if errorCodeErr != nil {
			return false, fmt.Errorf("error parse errorCode from response body: %s", errorCodeErr)
		}
		// CBC.99003651: Another operation is being performed.
		if operateErrorCode[errorCode.(string)] || errorCode == "CBC.99003651" {
			return true, err
		}
	}
	return false, err
}

func resourceDcsInstancesV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dcsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DcsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	var retryFunc = func() (interface{}, bool, error) {
		err = instance.Delete(client, d.Id())
		retry, err := handleOperationError(err)
		return nil, retry, err
	}

	_, err = common.RetryContextWithWaitForState(&common.RetryContextWithWaitForStateParam{
		Ctx:          ctx,
		RetryFunc:    retryFunc,
		WaitFunc:     refreshDcsInstanceState(client, d.Id()),
		WaitTarget:   []string{"RUNNING"},
		Timeout:      d.Timeout(schema.TimeoutDelete),
		DelayTimeout: 1 * time.Second,
		PollInterval: 10 * time.Second,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// Waiting to delete success
	err = waitForDcsInstanceCompleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{"RUNNING"}, []string{"DELETED"})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func getAzCode(d *schema.ResourceData, client *golangsdk.ServiceClient) ([]string, error) {
	var azCodes []string
	availabilityZones := d.Get("availability_zones")
	azCodes = common.ExpandToStringList(availabilityZones.([]interface{}))

	return azCodes, nil
}

func waitForSslCompleted(ctx context.Context, c *golangsdk.ServiceClient, d *schema.ResourceData) error {
	enable := d.Get("ssl_enable").(bool)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{strconv.FormatBool(!enable)},
		Target:       []string{strconv.FormatBool(enable)},
		Refresh:      updateSslStatusRefreshFunc(c, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        2 * time.Second,
		PollInterval: 2 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func updateSslStatusRefreshFunc(c *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := ssl.Get(c, id)
		if err != nil {
			if res, ok := err.(golangsdk.ErrUnexpectedResponseCode); ok && res.Actual == 204 {
				return r, strconv.FormatBool(false), nil
			}
			return nil, "Error", err
		}
		return r, strconv.FormatBool(r.Enabled), nil
	}
}

func setBandWidthInfo(bandWidthInfo *instance.BandWidthInfo) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"bandwidth":            bandWidthInfo.BandWidth,
			"begin_time":           common.FormatTimeStampRFC3339(int64(bandWidthInfo.BeginTime)/1000, false),
			"current_time":         common.FormatTimeStampRFC3339(int64(bandWidthInfo.CurrentTime)/1000, false),
			"end_time":             common.FormatTimeStampRFC3339(int64(bandWidthInfo.EndTime)/1000, false),
			"expand_count":         bandWidthInfo.ExpandCount,
			"expand_effect_time":   bandWidthInfo.ExpandEffectTime,
			"expand_interval_time": bandWidthInfo.ExpandIntervalTime,
			"max_expand_count":     bandWidthInfo.MaxExpandCount,
			"next_expand_time":     common.FormatTimeStampRFC3339(int64(bandWidthInfo.NextExpandTime)/1000, false),
			"task_running":         bandWidthInfo.TaskRunning,
		},
	}
}
