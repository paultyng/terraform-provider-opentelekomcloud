package ddm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"

	ddmv1instances "github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/instances"
	ddmv2instances "github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v2/instances"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v3/accounts"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDdmInstanceV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDdmInstanceV1Create,
		ReadContext:   resourceDdmInstanceV1Read,
		UpdateContext: resourceDdmInstanceV1Update,
		DeleteContext: resourceDdmInstanceV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common.ValidateDDMName,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"node_num": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"engine_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zones": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"subnet_id": {
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
			"param_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsUUID,
			},
			"time_zone": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateUTCOffset,
			},
			"username": {
				Type:         schema.TypeString,
				Sensitive:    true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateDDMUsername,
			},
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			"purge_rds_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_port": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
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
			"availability_zone": {
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

func resourceDdmInstanceV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	instanceDetails := ddmv1instances.CreateInstanceDetail{
		Name:              d.Get("name").(string),
		FlavorId:          d.Get("flavor_id").(string),
		NodeNum:           d.Get("node_num").(int),
		EngineId:          d.Get("engine_id").(string),
		AvailableZones:    resourceDDMAvailabilityZones(d),
		VpcId:             d.Get("vpc_id").(string),
		SecurityGroupId:   d.Get("security_group_id").(string),
		SubnetId:          d.Get("subnet_id").(string),
		ParamGroupId:      d.Get("param_group_id").(string),
		TimeZone:          d.Get("time_zone").(string),
		AdminUserName:     d.Get("username").(string),
		AdminUserPassword: d.Get("password").(string),
	}
	createOpts := ddmv1instances.CreateOpts{
		Instance: instanceDetails,
	}

	ddmInstance, err := ddmv1instances.Create(client, createOpts)
	if err != nil {
		return fmterr.Errorf("error getting OpenTelekomCloud DDM instance from result: %w", err)
	}
	log.Printf("[DEBUG] Create instance %s: %#v", ddmInstance.Id, ddmInstance)

	d.SetId(ddmInstance.Id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATE", "CREATING", "SET_CONFIGURATION", "RESTARTING"},
		Target:     []string{"RUNNING"},
		Refresh:    instanceStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for OpenTelekomCloud DDM instance (%s) to become ready: %w", d.Id(), err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV1)
	return resourceDdmInstanceV1Read(clientCtx, d, meta)
}

func resourceDdmInstanceV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	instance, err := ddmv1instances.QueryInstanceDetails(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error fetching DDM instance: %w", err)
	}

	log.Printf("[DEBUG] Retrieved instance %s: %#v", d.Id(), instance)

	mErr := multierror.Append(nil,
		d.Set("region", d.Get("region").(string)),
		d.Set("name", instance.Name),
		d.Set("status", instance.Status),
		d.Set("vpc_id", instance.VpcId),
		d.Set("subnet_id", instance.SubnetId),
		d.Set("security_group_id", instance.SecurityGroupId),
		d.Set("username", instance.AdminUserName),
		d.Set("availability_zone", instance.AvailableZone),
		d.Set("node_num", instance.NodeCount),
		d.Set("access_ip", instance.AccessIp),
		d.Set("access_port", instance.AccessPort),
		d.Set("node_status", instance.NodeStatus),
		d.Set("created_at", instance.Created),
		d.Set("updated_at", instance.Updated),
		d.Set("time_zone", d.Get("time_zone").(string)),
	)

	var nodesList []map[string]interface{}
	for _, nodeObj := range instance.Nodes {
		node := make(map[string]interface{})
		node["ip"] = nodeObj.IP
		node["port"] = nodeObj.Port
		node["status"] = nodeObj.Status
		nodesList = append(nodesList, node)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("nodes", nodesList),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceDdmInstanceV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	clientV1, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	clientV2, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	if d.HasChange("node_num") {
		err = resourceDDMScaling(clientV1, clientV2, d, ctx)
		if err != nil {
			return fmterr.Errorf("error in OpenTelekomCloud DDM instance scaling: %w", err)
		}
	}

	if d.HasChange("name") {
		_, newNameRaw := d.GetChange("name")
		newName := newNameRaw.(string)
		log.Printf("[DEBUG] Renaming instance %s: %s", d.Id(), newName)
		_, err = ddmv1instances.Rename(clientV1, d.Id(), newName)
		if err != nil {
			return fmterr.Errorf("error renaming OpenTelekomCloud DDM instance: %w", err)
		}
	}

	if d.HasChange("security_group_id") {
		_, newSecGroupRaw := d.GetChange("security_group_id")
		newSecGroup := newSecGroupRaw.(string)
		modifySecurityGroupOpts := ddmv1instances.ModifySecurityGroupOpts{
			SecurityGroupId: newSecGroup,
		}
		_, err := ddmv1instances.ModifySecurityGroup(clientV1, d.Id(), modifySecurityGroupOpts)
		if err != nil {
			return fmterr.Errorf("error modifying OpenTelekomCloud DDM instance security group: %w", err)
		}
	}

	if d.HasChange("password") {
		clientV3, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
			return config.DdmV3Client(config.GetRegion(d))
		})
		if err != nil {
			return fmterr.Errorf(errCreationV3Client, err)
		}

		_, err = accounts.ManageAdminPass(clientV3, d.Id(), accounts.ManageAdminPassOpts{
			Name:     d.Get("username").(string),
			Password: d.Get("password").(string),
		})
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud instance password: %w", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, clientV1, keyClientV1)
	return resourceDdmInstanceV1Read(clientCtx, d, meta)
}

func resourceDdmInstanceV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	log.Printf("[DEBUG] Deleting OpenTelekomCloud DDM Instance %s", d.Id())

	deleteRdsData := d.Get("purge_rds_on_delete").(bool)
	_, err = ddmv1instances.Delete(client, d.Id(), deleteRdsData)
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud DDM instance: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceDDMAvailabilityZones(d *schema.ResourceData) []string {
	azRaw := d.Get("availability_zones").([]interface{})
	zones := make([]string, 0)
	for _, v := range azRaw {
		zones = append(zones, v.(string))
	}
	return zones
}

func resourceDDMScaling(clientV1 *golangsdk.ServiceClient, clientV2 *golangsdk.ServiceClient, d *schema.ResourceData, ctx context.Context) error {
	oldNodeNumRaw, newNodeNumRaw := d.GetChange("node_num")
	oldNodeNum := oldNodeNumRaw.(int)
	newNodeNum := newNodeNumRaw.(int)
	if oldNodeNum < newNodeNum {
		log.Printf("[DEBUG] Scaling up OpenTelekomCloud DDM Instance %s", d.Id())
		scaleOutOpts := ddmv2instances.ScaleOutOpts{
			FlavorId:   d.Get("flavor_id").(string),
			NodeNumber: newNodeNum - oldNodeNum,
		}
		_, err := ddmv2instances.ScaleOut(clientV2, d.Id(), scaleOutOpts)
		if err != nil {
			return fmt.Errorf("error scaling up OpenTelekomCloudDDM instance: %w", err)
		}
		stateConf := &resource.StateChangeConf{
			Pending:    []string{"CREATING", "SET_CONFIGURATION", "RESTARTING", "GROWING"},
			Target:     []string{"RUNNING"},
			Refresh:    instanceStateRefreshFunc(clientV1, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      15 * time.Second,
			MinTimeout: 10 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmt.Errorf("error waiting for OpenTelekomCloud DDM instance (%s) to become ready during scale up: %w", d.Id(), err)
		}
	} else {
		log.Printf("[DEBUG] Scaling down Instance %s", d.Id())
		if oldNodeNum-newNodeNum < 1 {
			return fmt.Errorf("error scaling down OpenTelekomCloud DDM instance: %s\n num_nodes needs to be 1 or greater", d.Id())
		}
		scaleInOpts := ddmv2instances.ScaleInOpts{
			NodeNumber: oldNodeNum - newNodeNum,
		}
		_, err := ddmv2instances.ScaleIn(clientV2, d.Id(), scaleInOpts)
		if err != nil {
			return fmt.Errorf("error scaling down DDM instance: %w", err)
		}
		stateConf := &resource.StateChangeConf{
			Pending:    []string{"CREATING", "SET_CONFIGURATION", "RESTARTING", "REDUCING"},
			Target:     []string{"RUNNING"},
			Refresh:    instanceStateRefreshFunc(clientV1, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      15 * time.Second,
			MinTimeout: 10 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmt.Errorf("error waiting for OpenTelekomCloud DDM instance (%s) to become ready during scale down: %w", d.Id(), err)
		}
	}
	return nil
}
