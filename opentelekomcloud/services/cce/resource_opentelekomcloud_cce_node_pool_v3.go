package cce

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodepools"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

const (
	createError = "error creating Open Telekom Cloud CCE Node Pool: %w"
	setError    = "error setting %s for CCE Node Pool: %w"
)

func ResourceCCENodePoolV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCENodePoolV3Create,
		ReadContext:   resourceCCENodePoolV3Read,
		UpdateContext: resourceCCENodePoolV3Update,
		DeleteContext: resourceCCENodePoolV3Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),

			// used for cluster waiting
			Default: schema.DefaultTimeout(15 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: common.ImportByPath("cluster_id", "id"),
		},

		CustomizeDiff: common.MultipleCustomizeDiffs(
			common.ValidateVolumeType("root_volume.*.volumetype"),
			common.ValidateVolumeType("data_volumes.*.volumetype"),
			common.ValidateSubnet("subnet_id"),
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "random",
			},
			"os": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"root_volume": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0xa, 0x8000),
						},
						"volumetype": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"kms_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							DefaultFunc: schema.EnvDefaultFunc("OS_KMS_ID", nil),
						},
						"extend_param": {
							Type:       schema.TypeString,
							Optional:   true,
							ForceNew:   true,
							Deprecated: "use extend_params instead",
						},
						"extend_params": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					}},
			},
			"data_volumes": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0x64, 0x8000),
						},
						"volumetype": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"kms_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							DefaultFunc: schema.EnvDefaultFunc("OS_KMS_ID", nil),
						},
						"extend_param": {
							Type:       schema.TypeString,
							Optional:   true,
							ForceNew:   true,
							Deprecated: "use extend_params instead",
						},
						"extend_params": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					}},
			},
			"storage": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateJsonString,
				StateFunc: func(v interface{}) string {
					jsonString, _ := common.NormalizeJsonString(v)
					return jsonString
				},
			},
			"initial_node_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"k8s_tags": {
				Type:         schema.TypeMap,
				Optional:     true,
				Computed:     true,
				ValidateFunc: common.ValidateK8sTagsMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user_tags": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"taints": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"effect": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"NoSchedule", "PreferNoSchedule", "NoExecute",
							}, false),
						},
					}},
			},
			"runtime": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"docker", "containerd",
				}, false),
			},
			"key_pair": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"password", "key_pair"},
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"password", "key_pair"},
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"preinstall": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				StateFunc: common.GetHashOrEmpty,
			},
			"postinstall": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				StateFunc: common.GetHashOrEmpty,
			},
			"max_pods": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"agency_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"docker_base_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"docker_lvm_config_override": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"scale_enable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"min_node_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_node_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"scale_down_cooldown_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"server_group_reference": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCCENodePoolUserTags(d *schema.ResourceData) []tags.ResourceTag {
	tagRaw := d.Get("user_tags").(map[string]interface{})
	return common.ExpandResourceTags(tagRaw)
}

func resourceCCENodePoolV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	var base64PreInstall, base64PostInstall string
	if v, ok := d.GetOk("preinstall"); ok {
		base64PreInstall = common.InstallScriptEncode(v.(string))
	}
	if v, ok := d.GetOk("postinstall"); ok {
		base64PostInstall = common.InstallScriptEncode(v.(string))
	}
	var loginSpec nodes.LoginSpec
	if common.HasFilledOpt(d, "key_pair") {
		loginSpec = nodes.LoginSpec{SshKey: d.Get("key_pair").(string)}
	}
	if common.HasFilledOpt(d, "password") {
		loginSpec = nodes.LoginSpec{
			UserPassword: nodes.UserPassword{
				Username: "root",
				Password: d.Get("password").(string),
			},
		}
	}

	createOpts := nodepools.CreateOpts{
		Kind:       "NodePool",
		ApiVersion: "v3",
		Metadata: nodepools.CreateMetaData{
			Name: d.Get("name").(string),
		},
		Spec: nodepools.CreateSpec{
			Type:             "vm",
			InitialNodeCount: d.Get("initial_node_count").(int),
			Autoscaling: nodepools.AutoscalingSpec{
				Enable:                d.Get("scale_enable").(bool),
				MinNodeCount:          d.Get("min_node_count").(int),
				MaxNodeCount:          d.Get("max_node_count").(int),
				ScaleDownCooldownTime: d.Get("scale_down_cooldown_time").(int),
				Priority:              d.Get("priority").(int),
			},
			NodeManagement: nodepools.NodeManagementSpec{
				ServerGroupReference: d.Get("server_group_reference").(string),
			},
			CustomSecurityGroupIds: common.ExpandToStringList(d.Get("security_group_ids").([]interface{})),
			NodeTemplate: nodes.Spec{
				Flavor:      d.Get("flavor").(string),
				Az:          d.Get("availability_zone").(string),
				Os:          d.Get("os").(string),
				Login:       loginSpec,
				RootVolume:  resourceCCERootVolume(d),
				DataVolumes: resourceCCEDataVolume(d),
				Count:       1,
				NodeNicSpec: nodes.NodeNicSpec{
					PrimaryNic: nodes.PrimaryNic{
						SubnetId: d.Get("subnet_id").(string),
					},
				},
				ExtendParam: nodes.ExtendParam{
					MaxPods:                 d.Get("max_pods").(int),
					PreInstall:              base64PreInstall,
					PostInstall:             base64PostInstall,
					DockerBaseSize:          d.Get("docker_base_size").(int),
					DockerLVMConfigOverride: d.Get("docker_lvm_config_override").(string),
					AgencyName:              d.Get("agency_name").(string),
				},
				Taints:   resourceCCENodeTaints(d),
				K8sTags:  resourceCCENodeK8sTags(d),
				UserTags: resourceCCENodePoolUserTags(d),
			},
		},
	}

	if storageJsonRaw, ok := d.GetOk("storage"); ok {
		var storage nodes.Storage
		err = json.Unmarshal([]byte(storageJsonRaw.(string)), &storage)
		if err != nil {
			return fmterr.Errorf("error unmarshalling flavor json %s", err)
		}
		createOpts.Spec.NodeTemplate.Storage = &storage
	}

	if v, ok := d.GetOk("runtime"); ok {
		createOpts.Spec.NodeTemplate.Runtime = nodes.RuntimeSpec{
			Name: v.(string),
		}
	}

	clusterID := d.Get("cluster_id").(string)
	clusterStateConf := &resource.StateChangeConf{
		Target:     []string{"Available"},
		Refresh:    waitForClusterAvailable(client, clusterID),
		Timeout:    d.Timeout(schema.TimeoutDefault),
		Delay:      15 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := clusterStateConf.WaitForStateContext(ctx); err != nil {
		return fmterr.Errorf("error waiting for cluster to be available: %w", err)
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	pool, err := nodepools.Create(client, clusterID, createOpts)
	switch err.(type) {
	case golangsdk.ErrDefault403:
		if _, err := clusterStateConf.WaitForStateContext(ctx); err != nil {
			return fmterr.Errorf("error waiting for cluster to be available: %w", err)
		}
		retried, err := nodepools.Create(client, clusterID, createOpts)
		if err != nil {
			return fmterr.Errorf(createError, err)
		}
		pool = retried
	case nil:
		break
	default:
		return fmterr.Errorf(createError, err)
	}
	d.SetId(pool.Metadata.Id)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Synchronizing", "Synchronized"},
		Target:       []string{""},
		Refresh:      waitForCceNodePoolActive(client, clusterID, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        120 * time.Second,
		PollInterval: 20 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmterr.Errorf(createError, err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceCCENodePoolV3Read(clientCtx, d, meta)
}

func resourceCCENodePoolV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	clusterID := d.Get("cluster_id").(string)
	s, err := nodepools.Get(client, clusterID, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "CCE Node Pool")
	}

	rootVolume := []map[string]interface{}{
		{
			"size":          s.Spec.NodeTemplate.RootVolume.Size,
			"volumetype":    s.Spec.NodeTemplate.RootVolume.VolumeType,
			"extend_params": s.Spec.NodeTemplate.RootVolume.ExtendParam,
			"extend_param":  "",
		},
	}
	if s.Spec.NodeTemplate.RootVolume.Metadata != nil {
		rootVolume[0]["kms_id"] = s.Spec.NodeTemplate.RootVolume.Metadata["__system__cmkid"]
	}

	mErr := multierror.Append(
		d.Set("name", s.Metadata.Name),
		d.Set("flavor", s.Spec.NodeTemplate.Flavor),
		d.Set("availability_zone", s.Spec.NodeTemplate.Az),
		d.Set("os", s.Spec.NodeTemplate.Os),
		d.Set("key_pair", s.Spec.NodeTemplate.Login.SshKey),
		d.Set("scale_enable", s.Spec.Autoscaling.Enable),
		d.Set("max_pods", s.Spec.NodeTemplate.ExtendParam.MaxPods),
		d.Set("subnet_id", s.Spec.NodeTemplate.NodeNicSpec.PrimaryNic.SubnetId),
		d.Set("security_group_ids", s.Spec.CustomSecurityGroupIds),
		d.Set("root_volume", rootVolume),
		d.Set("status", s.Status.Phase),
	)

	if s.Spec.NodeTemplate.Runtime.Name == "null" {
		if v, ok := d.GetOk("runtime"); ok {
			mErr = multierror.Append(mErr, d.Set("runtime", v.(string)))
		}
	} else {
		mErr = multierror.Append(mErr, d.Set("runtime", "docker"))
	}

	if s.Spec.NodeTemplate.Runtime.Name != "" {
		mErr = multierror.Append(mErr, d.Set("runtime", s.Spec.NodeTemplate.Runtime.Name))
	}

	if s.Spec.Autoscaling.Enable {
		mErr = multierror.Append(mErr,
			d.Set("min_node_count", s.Spec.Autoscaling.MinNodeCount),
			d.Set("max_node_count", s.Spec.Autoscaling.MaxNodeCount),
			d.Set("scale_down_cooldown_time", s.Spec.Autoscaling.ScaleDownCooldownTime),
			d.Set("priority", s.Spec.Autoscaling.Priority),
		)
	}

	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf(setError, "attributes", mErr)
	}

	k8sTags := map[string]string{}
	for key, val := range s.Spec.NodeTemplate.K8sTags {
		if strings.Contains(key, "cce.cloud.com") {
			continue
		}
		k8sTags[key] = val
	}
	if err := d.Set("k8s_tags", k8sTags); err != nil {
		return fmterr.Errorf(setError, "k8s_tags", err)
	}

	var volumes []interface{}
	for _, pairObject := range s.Spec.NodeTemplate.DataVolumes {
		volume := map[string]interface{}{
			"size":          pairObject.Size,
			"volumetype":    pairObject.VolumeType,
			"extend_params": pairObject.ExtendParam,
			"extend_param":  "",
		}
		if pairObject.Metadata != nil {
			volume["kms_id"] = pairObject.Metadata["__system__cmkid"]
		}
		volumes = append(volumes, volume)
	}
	if err := d.Set("data_volumes", volumes); err != nil {
		return fmterr.Errorf(setError, "data_volumes", err)
	}

	return nil
}

func resourceCCENodePoolV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	updateOpts := nodepools.UpdateOpts{
		Metadata: nodepools.UpdateMetaData{
			Name: d.Get("name").(string),
		},
		Spec: nodepools.UpdateSpec{
			InitialNodeCount: d.Get("initial_node_count").(int),
			Autoscaling: nodepools.UpdateAutoscalingSpec{
				Enable:                d.Get("scale_enable").(bool),
				MinNodeCount:          d.Get("min_node_count").(int),
				MaxNodeCount:          d.Get("max_node_count").(int),
				ScaleDownCooldownTime: d.Get("scale_down_cooldown_time").(int),
				Priority:              d.Get("priority").(int),
			},
			NodeTemplate: nodepools.UpdateNodeTemplate{
				K8sTags: resourceCCENodeK8sTags(d),
				Taints:  resourceCCENodeTaints(d),
			},
		},
	}

	clusterID := d.Get("cluster_id").(string)
	_, err = nodepools.Update(client, clusterID, d.Id(), updateOpts)
	if err != nil {
		return fmterr.Errorf("error updating Open Telekom Cloud CCE Node Pool: %w", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Synchronizing", "Synchronized"},
		Target:     []string{""},
		Refresh:    waitForCceNodePoolActive(client, clusterID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      15 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmterr.Errorf("error waiting for Open Telekom Cloud CCE Node Pool to update: %w", err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceCCENodePoolV3Read(clientCtx, d, meta)
}

func resourceCCENodePoolV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	clusterID := d.Get("cluster_id").(string)

	if err := nodepools.Delete(client, clusterID, d.Id()); err != nil {
		return fmterr.Errorf("error deleting Open Telekom Cloud CCE Node Pool: %w", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Deleting"},
		Target:       []string{"Deleted"},
		Refresh:      waitForCceNodePoolDelete(client, clusterID, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        60 * time.Second,
		PollInterval: 20 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for Open Telekom Cloud CCE Node Pool to be deleted: %w", err)
	}

	d.SetId("")
	return nil
}

func waitForCceNodePoolActive(cceClient *golangsdk.ServiceClient, clusterId, nodePoolId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := nodepools.Get(cceClient, clusterId, nodePoolId)
		if err != nil {
			return nil, "", err
		}
		return n, n.Status.Phase, nil
	}
}

func waitForCceNodePoolDelete(cceClient *golangsdk.ServiceClient, clusterID, nodePoolID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete Open Telekom Cloud CCE Node Pool %s.\n", nodePoolID)

		r, err := nodepools.Get(cceClient, clusterID, nodePoolID)

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[DEBUG] Successfully deleted Open Telekom Cloud CCE Node Pool %s", nodePoolID)
				return r, "Deleted", nil
			}
			return r, "Deleting", err
		}

		log.Printf("[DEBUG] Open Telekom Cloud Node Pool %s still available.\n", nodePoolID)
		return r, r.Status.Phase, nil
	}
}
