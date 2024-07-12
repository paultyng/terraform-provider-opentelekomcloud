package migrations

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	apis "github.com/opentelekomcloud/gophertelekomcloud/openstack/dataarts/v1.1/cluster"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceClusterV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterV1Create,
		ReadContext:   resourceClusterV1Read,
		DeleteContext: resourceClusterV1Delete,

		Importer: &schema.ResourceImporter{
			//
			// StateContext: resourceClusterV1ImportState,
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"language": {
				Type: schema.TypeString,
				// TODO: add to the documentation a warning: "All changes in parameters will cause the cluster recreation."
				ForceNew: true,
				Required: true,
			},
			"auto_remind": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"schedule_boot_time": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"is_schedule_boot_off": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     instanceSchema(),
			},
			"datastore": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  string(ClusterTypeCDM),
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"extended_properties": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workspace": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"resource": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"trial": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"schedule_off_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"sys_tags": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"is_auto_off": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_endpoint_domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"flavor_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"endpoint_domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"eip_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"cluster_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func instanceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(FlavorTypeSmall),
					string(FlavorTypeMedium),
					string(FlavorTypeLarge),
					string(FlavorTypeXLarge),
				}, false),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(ClusterTypeCDM),
				}, false),
			},
			"nics": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group": {
							Type:     schema.TypeString,
							Required: true,
						},
						"net": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"vm_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"group": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"manage_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"traffic_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"shard_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"manage_fix_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"internal_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
		},
	}

}

func buildCluster(cluster *schema.Set) apis.Cluster {
	c := cluster.List()[0].(map[string]interface{})
	i := c["instances"].([]interface{})
	ds := c["datastore"].(map[string]interface{})
	ep := c["extended_properties"].(map[string]interface{})
	st := c["sys_tags"].([]interface{})

	return apis.Cluster{
		ScheduleBootTime:   c["schedule_boot_time"].(string),
		IsScheduleBootOff:  pointerto.Bool(c["is_schedule_boot_off"].(bool)),
		Instances:          buildInstances(i),
		DataStore:          buildDatastore(ds),
		ExtendedProperties: buildExtendedProperties(ep),
		ScheduleOffTime:    c["schedule_off_time"].(string),
		VpcId:              c["vpc_id"].(string),
		Name:               c["name"].(string),
		SysTags:            buildSysTags(st),
		IsAutoOff:          c["is_auto_off"].(bool),
	}
}

func buildInstances(instances []interface{}) []apis.Instance {
	if len(instances) == 0 {
		return nil
	}
	insSlice := make([]apis.Instance, len(instances))
	for _, i := range instances {
		ins := i.(map[string]interface{})
		insSlice = append(insSlice, apis.Instance{
			AZ:        ins["availability_zone"].(string),
			Nics:      buildNics(ins["nics"].([]interface{})),
			FlavorRef: ins["flavor"].(string),
			Type:      ins["type"].(string),
		})
	}
	return insSlice
}

func buildNics(nics []interface{}) []apis.Nic {
	if len(nics) == 0 {
		return nil
	}
	nicSlice := make([]apis.Nic, len(nics))
	for _, nic := range nics {
		n := nic.(map[string]interface{})
		nicSlice = append(nicSlice, apis.Nic{
			SecurityGroupId: n["security_group"].(string),
			NetId:           n["net"].(string),
		})
	}

	return nicSlice
}

func buildDatastore(ds map[string]interface{}) *apis.Datastore {

	return &apis.Datastore{
		Type:    ds["type"].(string),
		Version: ds["version"].(string),
	}
}

func buildExtendedProperties(ep map[string]interface{}) *apis.ExtendedProp {

	return &apis.ExtendedProp{
		WorkSpaceId: ep["workspace"].(string),
		ResourceId:  ep["resource"].(string),
		Trial:       ep["trial"].(string),
	}
}
func buildSysTags(st []interface{}) []tags.ResourceTag {
	if len(st) < 1 {
		return nil
	}

	ts := make([]tags.ResourceTag, len(st))

	for _, t := range st {
		tag := t.(map[string]interface{})
		ts = append(ts, tags.ResourceTag{
			Key:   tag["key"].(string),
			Value: tag["value"].(string),
		})
	}

	return ts
}

func buildApiCreateOpts(d *schema.ResourceData) (apis.CreateOpts, error) {

	opts := apis.CreateOpts{
		Cluster:    buildCluster(d.Get("cluster").(*schema.Set)),
		AutoRemind: d.Get("auto_remind").(bool),
		PhoneNum:   d.Get("phone_number").(string),
		Email:      d.Get("email").(string),
		XLang:      d.Get("language").(string),
	}

	log.Printf("[DEBUG] The Migration Cluster Creation Opts are : %+v", opts)
	return opts, nil
}

func resourceClusterV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DataArtsMigrationsV1Client(config.GetProjectName(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	opts, err := buildApiCreateOpts(d)
	if err != nil {
		return diag.Errorf("unable to build the OpenTelekomCloud DataArts Migrations Cluster API create opts: %s", err)
	}
	api, err := apis.Create(client, opts)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud DataArts Migrations Cluster API: %s", err)
	}
	d.SetId(api.Id)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV1)
	return resourceClusterV1Read(clientCtx, d, meta)
}

func resourceClusterV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DataArtsMigrationsV1Client(config.GetProjectName(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	resp, err := apis.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "OpenTelekomCloud DataArts Migrations Cluster")
	}

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("public_endpoint", resp.PublicEndpoint),
		d.Set("instances", flattenInstances(resp.Instances)),
		d.Set("security_group_id", resp.SecurityGroupId),
		d.Set("vpc_id", resp.VpcId),
		d.Set("subnet_id", resp.SubnetId),
		d.Set("datastore", flattenDatastore(resp.Datastore)),
		d.Set("is_auto_off", resp.IsAutoOff),
		d.Set("public_endpoint_domain_name", resp.PublicEndpointDomainName),
		d.Set("flavor_name", resp.FlavorName),
		d.Set("availability_zone", resp.AzName),
		d.Set("endpoint_domain_name", resp.EndpointDomainName),
		d.Set("is_schedule_boot_off", resp.IsScheduleBootOff),
		d.Set("namespace", resp.Namespace),
		d.Set("eip_id", resp.EipId),
		d.Set("db_user", resp.DbUser),
		d.Set("links", flattenLinks(resp.Links)),
		d.Set("cluster_mode", resp.ClusterMode),
		d.Set("task", flattenTask(resp.Task)),
		d.Set("created", resp.Created),
		d.Set("status_detail", resp.StatusDetail),
		d.Set("config_status", resp.ConfigStatus),
		d.Set("actionProgress", flattenActionProgress(resp.ActionProgress)),
		d.Set("name", resp.Name),
		d.Set("cluster_id", resp.Id),
		d.Set("is_frozen", resp.IsFrozen),
		d.Set("actions", resp.Actions),
		d.Set("updated", resp.Updated),
		d.Set("status", analyseClusterStatus(resp.Status)),
	)
	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error saving  OpenTelekomCloud DataArts Migrations Cluster API fields: %s", err)
	}
	return nil
}

func resourceClusterV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DataArtsMigrationsV1Client(config.GetProjectName(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	deleteOpts := apis.DeleteOpts{}

	if lastBackups, ok := d.Get("keep_last_manual_backup").(int); ok {
		deleteOpts.KeepBackup = lastBackups

	}

	if _, err = apis.Delete(client, d.Id(), deleteOpts); err != nil {
		return diag.Errorf("unable to delete the OpenTelekomCloud DataArts Migrations API (%s): %s", d.Id(), err)
	}
	d.SetId("")

	return nil
}

func flattenInstances(reqParams []apis.DetailedInstances) []map[string]interface{} {
	if len(reqParams) < 1 {
		return nil
	}

	result := make([]map[string]interface{}, len(reqParams))
	for i, v := range reqParams {
		param := map[string]interface{}{
			"flavor":          flattenFlavor(v.Flavor),
			"volume":          flattenVolume(v.Volume),
			"status":          analyseStatus(v.Status),
			"actions":         v.Actions,
			"type":            v.Type,
			"id":              v.Id,
			"name":            v.Name,
			"is_frozen":       v.IsFrozen,
			"components":      v.Components,
			"config_status":   v.ConfigStatus,
			"role":            v.Role,
			"group":           v.Group,
			"links":           flattenLinks(v.Links),
			"params_group_id": v.ParamsGroupId,
			"public_ip":       v.PublicIp,
			"manage_ip":       v.ManageIp,
			"traffic_ip":      v.TrafficIp,
			"shard_id":        v.ShardId,
			"manage_fix_ip":   v.ManageFixIp,
			"private_ip":      v.PrivateIp,
			"internal_ip":     v.InternalIp,
			"resource":        flattenResource(v.Resource),
		}

		result[i] = param
	}
	return result

}

func flattenFlavor(f apis.Flavor) map[string]interface{} {
	fs := map[string]interface{}{
		"id": f.Id,
	}
	fs["links"] = flattenLinks(f.Links)

	return fs
}

func flattenLinks(links []apis.ClusterLinks) []map[string]interface{} {
	if len(links) < 1 {
		return nil
	}

	ls := make([]map[string]interface{}, 0)
	for _, v := range links {
		link := map[string]interface{}{
			"rel":  v.Rel,
			"href": v.Href,
		}
		ls = append(ls, link)
	}

	return ls
}

func flattenVolume(v apis.Volume) map[string]interface{} {
	return map[string]interface{}{
		"type": v.Type,
		"size": v.Size,
	}
}

func analyseStatus(s string) string {
	apiType := map[string]string{
		"100": "creating",
		"200": "normal",
		"300": "failed",
		"303": "failed to be created",
		"400": "deleted",
		"800": "frozen",
	}
	if v, ok := apiType[s]; ok {
		return v
	}
	return ""
}

func flattenResource(resources []apis.Resource) []map[string]interface{} {
	if len(resources) < 1 {
		return nil
	}

	rs := make([]map[string]interface{}, len(resources))
	for _, resource := range resources {
		rs = append(rs, map[string]interface{}{
			"resource_id":   resource.ResourceId,
			"resource_type": resource.ResourceType,
		})
	}
	return rs
}

func flattenCustomerConfig(cConfig apis.CustomerConfig) map[string]interface{} {
	return map[string]interface{}{
		"failure_remind":   cConfig.FailureRemind,
		"cluster_name":     cConfig.ClusterName,
		"service_provider": cConfig.ServiceProvider,
		"local_disk":       cConfig.LocalDisk,
		"ssl":              cConfig.Ssl,
		"create_from":      cConfig.CreateFrom,
		"resource_id":      cConfig.ResourceId,
		"flavor_type":      cConfig.FlavorType,
		"workSpace_id":     cConfig.WorkSpaceId,
		"trial":            cConfig.Trial,
	}
}

func flattenDatastore(d apis.Datastore) map[string]interface{} {
	return map[string]interface{}{
		"type":    d.Type,
		"version": d.Version,
	}
}

func flattenEndpointStatus(pStatus apis.PublicEndpointStatus) map[string]interface{} {
	return map[string]interface{}{
		"status":        pStatus.Status,
		"error_message": pStatus.ErrorMessage,
	}
}

func flattenFailedReasons(fReasons apis.FailedReasons) map[string]interface{} {
	return map[string]interface{}{
		"create_failed": map[string]interface{}{
			"error_code": fReasons.CreateFailed.ErrorCode,
			"error_msg":  fReasons.CreateFailed.ErrorMsg,
		},
	}
}

func flattenTask(cTask apis.ClusterTask) map[string]interface{} {
	return map[string]interface{}{
		"description": cTask.Description,
		"id":          cTask.Id,
		"name":        cTask.Name,
	}
}

func flattenActionProgress(ap apis.ActionProgress) map[string]interface{} {
	return map[string]interface{}{
		"creating":     ap.Creating,
		"growing":      ap.Growing,
		"restoring":    ap.Restoring,
		"snapshotting": ap.Snapshotting,
		"repairing":    ap.Repairing,
	}
}

func analyseClusterStatus(s string) string {
	apiType := map[string]string{
		"100": "creating",
		"200": "normal",
		"300": "failed",
		"303": "failed to be created",
		"400": "deleted",
		"800": "frozen",
		"900": "stopped",
		"910": "stopping",
		"920": "starting",
	}
	if v, ok := apiType[s]; ok {
		return v
	}

	return ""
}
