package dms

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/instances/lifecycle"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDmsInstancesV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsInstancesV2Create,
		ReadContext:   resourceDmsInstancesV2Read,
		UpdateContext: resourceDmsInstancesV2Update,
		DeleteContext: resourceDmsInstancesV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(50 * time.Minute),
			Update: schema.DefaultTimeout(50 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		DeprecationMessage: "Please use `opentelekomcloud_dms_dedicated_instance_v2` resource instead",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(4, 64),
					validation.StringMatch(
						regexp.MustCompile(`^[\w\-.]+$`),
						"Only lowercase letters, digits, periods (.), underscores (_), and hyphens (-) are allowed.",
					),
					validation.StringDoesNotMatch(
						regexp.MustCompile(`_{3,}?|\.{2,}?|-{2,}?`),
						"Periods, underscores, and hyphens cannot be placed next to each other. A maximum of two consecutive underscores are allowed.",
					),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"kafka",
				}, false),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_space": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"partition_num": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"password": {
				Type:         schema.TypeString,
				Sensitive:    true,
				ForceNew:     true,
				Optional:     true,
				RequiredWith: []string{"access_user"},
			},
			"access_user": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"password"},
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"security_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"available_zones": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"maintain_begin": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"maintain_end": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_publicip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"publicip_id": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				RequiredWith: []string{"enable_publicip"},
			},
			"public_bandwidth": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"retention_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"produce_reject", "time_base",
				}, false),
			},
			"disk_encrypted_enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"disk_encrypted_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"disk_encrypted_enable"},
			},
			"specification": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"storage_spec_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
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
			"connect_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resource_spec_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"used_storage_space": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_storage_space": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"public_connect_address": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"storage_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_access_enabled": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ssl_enable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": common.TagsSchema(),
		},
	}
}

func resourceDmsInstancesV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DmsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	sslEnable := false
	if d.Get("access_user").(string) != "" || d.Get("password").(string) != "" {
		sslEnable = true
	}

	createOpts := lifecycle.CreateDeprOpts{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Engine:          d.Get("engine").(string),
		EngineVersion:   d.Get("engine_version").(string),
		StorageSpace:    d.Get("storage_space").(int),
		Password:        d.Get("password").(string),
		AccessUser:      d.Get("access_user").(string),
		VpcID:           d.Get("vpc_id").(string),
		SecurityGroupID: d.Get("security_group_id").(string),
		SubnetID:        d.Get("subnet_id").(string),
		AvailableZones:  common.GetAllAvailableZones(d),
		ProductID:       d.Get("product_id").(string),
		MaintainBegin:   d.Get("maintain_begin").(string),
		MaintainEnd:     d.Get("maintain_end").(string),
		PartitionNum:    d.Get("partition_num").(int),
		Specification:   d.Get("specification").(string),
		StorageSpecCode: d.Get("storage_spec_code").(string),
		RetentionPolicy: d.Get("retention_policy").(string),
		SslEnable:       &sslEnable,
	}

	if d.Get("enable_publicip").(bool) {
		createOpts.EnablePublicIP = true
		rawIpList := d.Get("publicip_id").([]interface{})
		var ipList []string
		for _, ip := range rawIpList {
			ipList = append(ipList, ip.(string))
		}
		createOpts.PublicIpID = strings.Join(ipList, ",")
	}

	if d.Get("disk_encrypted_enable").(bool) {
		diskEncryptedEnable := true
		createOpts.DiskEncryptedEnable = &diskEncryptedEnable
		createOpts.DiskEncryptedKey = d.Get("disk_encrypted_key").(string)
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	v, err := lifecycle.CreateDepr(client, createOpts)
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud DMSv2 instance: %w", err)
	}
	log.Printf("[INFO] instance ID: %s", v.InstanceID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"RUNNING"},
		Refresh:    instancesV2StateRefreshFunc(client, v.InstanceID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for instance (%s) to become ready: %w", v.InstanceID, err)
	}

	// Store the instance ID now
	d.SetId(v.InstanceID)

	// Tag assignment during instance creation doesn't work therefore
	// tags are assigned via separate request
	if rawTags, ok := d.GetOk("tags"); ok {
		tagList := common.ExpandResourceTags(rawTags.(map[string]interface{}))
		err := tags.Create(client, "kafka", d.Id(), tagList).ExtractErr()
		if err != nil {
			return fmterr.Errorf("error assigning tags for instance (%s) : %w", v.InstanceID, err)
		}
	}

	return resourceDmsInstancesV2Read(ctx, d, meta)
}

func resourceDmsInstancesV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DmsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}
	v, err := lifecycle.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "DMS instance")
	}

	log.Printf("[DEBUG] DMS instance %s: %+v", d.Id(), v)

	mErr := multierror.Append(
		d.Set("name", v.Name),
		d.Set("engine", v.Engine),
		d.Set("engine_version", v.EngineVersion),
		d.Set("specification", v.Specification),
		d.Set("used_storage_space", v.UsedStorageSpace),
		d.Set("storage_space", v.TotalStorageSpace),
		d.Set("connect_address", v.ConnectAddress),
		d.Set("port", v.Port),
		d.Set("status", v.Status),
		d.Set("description", v.Description),
		d.Set("resource_spec_code", v.ResourceSpecCode),
		d.Set("type", v.Type),
		d.Set("vpc_id", v.VPCID),
		d.Set("vpc_name", v.VPCName),
		d.Set("created_at", v.CreatedAt),
		d.Set("product_id", v.ProductID),
		d.Set("security_group_id", v.SecurityGroupID),
		d.Set("security_group_name", v.SecurityGroupName),
		d.Set("subnet_id", v.SubnetID),
		d.Set("subnet_name", v.SubnetName),
		d.Set("user_id", v.UserID),
		d.Set("user_name", v.UserName),
		d.Set("access_user", v.AccessUser),
		d.Set("maintain_begin", strings.TrimSuffix(v.MaintainBegin, ":00")),
		d.Set("maintain_end", strings.TrimSuffix(v.MaintainEnd, ":00")),
		d.Set("retention_policy", v.RetentionPolicy),
		d.Set("enable_publicip", v.EnablePublicIP),
		d.Set("public_connect_address", flattenPublicIps(v.PublicConnectAddress)),
		d.Set("public_bandwidth", v.PublicBandWidth),
		d.Set("ssl_enable", v.SslEnable),
		d.Set("disk_encrypted_enable", v.DiskEncrypted),
		d.Set("disk_encrypted_key", v.DiskEncryptedKey),
		d.Set("subnet_cidr", v.SubnetCIDR),
		d.Set("total_storage_space", v.TotalStorageSpace),
		d.Set("storage_resource_id", v.StorageResourceID),
		d.Set("public_access_enabled", v.PublicAccessEnabled),
		d.Set("node_num", v.NodeNum),
	)

	if resourceTags, err := tags.Get(client, "kafka", d.Id()).Extract(); err == nil {
		tagMap := common.TagsToMap(resourceTags)
		if err = d.Set("tags", tagMap); err != nil {
			mErr = multierror.Append(mErr,
				fmt.Errorf("error saving tags to state for DMS kafka instance (%s): %s", d.Id(), err))
		}
	} else {
		log.Printf("[WARN] error fetching tags of DMS kafka instance (%s): %s", d.Id(), err)
	}

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDmsInstancesV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DmsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	var updateOpts lifecycle.UpdateOpts
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("description") {
		description := d.Get("description").(string)
		updateOpts.Description = &description
	}
	if d.HasChange("maintain_begin") {
		maintainBegin := d.Get("maintain_begin").(string)
		updateOpts.MaintainBegin = maintainBegin
	}
	if d.HasChange("maintain_end") {
		maintainEnd := d.Get("maintain_end").(string)
		updateOpts.MaintainEnd = maintainEnd
	}
	if d.HasChange("security_group_id") {
		securityGroupID := d.Get("security_group_id").(string)
		updateOpts.SecurityGroupID = securityGroupID
	}

	if _, err := lifecycle.Update(client, d.Id(), updateOpts); err != nil {
		return fmterr.Errorf("error updating OpenTelekomCloud DMSv2 Instance: %s", err)
	}

	if d.HasChange("tags") {
		if err = common.UpdateResourceTags(client, d, "kafka", d.Id()); err != nil {
			err = fmt.Errorf("error updating tags of Kafka instance: %s, err: %s",
				d.Id(), err)
			if err != nil {
				return diag.Errorf("error while updating DMSv2 tags: %s", err)
			}
		}
	}

	return resourceDmsInstancesV2Read(ctx, d, meta)
}

func resourceDmsInstancesV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DmsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	_, err = lifecycle.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "instance")
	}

	err = lifecycle.Delete(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud DMSv2 instance: %w", err)
	}

	// Wait for the instance to delete before moving on.
	log.Printf("[DEBUG] Waiting for instance (%s) to delete", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING", "RUNNING"},
		Target:     []string{"DELETED"},
		Refresh:    instancesV2StateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for instance (%s) to delete: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] DMS instance %s deactivated.", d.Id())
	d.SetId("")
	return nil
}

func instancesV2StateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := lifecycle.Get(client, instanceID)
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return v, "DELETED", nil
			}
			return nil, "", err
		}

		return v, v.Status, nil
	}
}

func flattenPublicIps(ips string) []string {
	return strings.Split(ips, ",")
}
