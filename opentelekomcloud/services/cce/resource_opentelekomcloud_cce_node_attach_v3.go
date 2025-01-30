package cce

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceCCENodeV3Attach() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCENodeV3AttachCreate,
		ReadContext:   resourceCCENodeV3Read,
		UpdateContext: resourceCCENodeV3AttachUpdate,
		DeleteContext: resourceCCENodeV3AttachDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"os": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_pair": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"password", "key_pair"},
				ForceNew:     true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"max_pods": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"lvm_config": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"storage",
				},
			},
			"docker_base_size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"runtime": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"system_disk_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"preinstall": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"postinstall": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage": resourceNodeStorageUpdatableSchema(),
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
						},
					}},
			},
			"k8s_tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// (node/ecs_tags)
			"tags": common.TagsSchema(),
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_volume": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volumetype": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"extend_params": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"kms_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"extend_param": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"data_volumes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volumetype": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"extend_params": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"kms_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dss_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hw_passthrough": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"extend_param": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "use extend_params instead",
						},
					},
				},
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"billing_mode": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceNodeStorageUpdatableSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"selectors": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"type": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "evs",
							},
							"match_label_size": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"match_label_volume_type": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"match_label_metadata_encrypted": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"match_label_metadata_cmkid": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"match_label_count": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"groups": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"cce_managed": {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"selector_names": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"virtual_spaces": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"size": {
											Type:     schema.TypeString,
											Required: true,
										},
										"lvm_lv_type": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"lvm_path": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"runtime_lv_type": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceNodeAttachServerConfig(d *schema.ResourceData) *nodes.ReinstallServerConfig {
	var res nodes.ReinstallServerConfig
	if common.HasFilledOpt(d, "tags") {
		res.UserTags = buildResourceNodeTags(d)
	}

	if common.HasFilledOpt(d, "image_id") || common.HasFilledOpt(d, "system_disk_kms_key_id") {
		rootVolume := nodes.ReinstallVolumeSpec{
			ImageID: d.Get("image_id").(string),
			CmkID:   d.Get("system_disk_kms_key_id").(string),
		}
		res.RootVolume = &rootVolume
	}

	return &res
}

func buildResourceNodeTags(d *schema.ResourceData) []tags.ResourceTag {
	tagRaw := d.Get("tags").(map[string]interface{})
	return common.ExpandResourceTags(tagRaw)
}

func resourceNodeAttachVolumeConfig(d *schema.ResourceData) *nodes.ReinstallVolumeConfig {
	if v, ok := d.GetOk("lvm_config"); ok {
		volumeConfig := nodes.ReinstallVolumeConfig{
			LvmConfig: v.(string),
		}
		return &volumeConfig
	}

	if _, ok := d.GetOk("storage"); ok {
		volumeConfig := nodes.ReinstallVolumeConfig{
			Storage: buildResourceNodeStorage(d),
		}
		return &volumeConfig
	}
	return nil
}

func resourceNodeAttachRuntimeConfig(d *schema.ResourceData) *nodes.ReinstallRuntimeConfig {
	var res nodes.ReinstallRuntimeConfig

	if v, ok := d.GetOk("docker_base_size"); ok {
		res.DockerBaseSize = v.(int)
	}

	if v, ok := d.GetOk("runtime"); ok {
		res.Runtime = &nodes.RuntimeSpec{
			Name: v.(string),
		}
	}

	return &res
}

func buildResourceNodeK8sTags(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("k8s_tags").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func buildResourceNodeTaint(d *schema.ResourceData) []nodes.TaintSpec {
	taintRaw := d.Get("taints").([]interface{})
	taints := make([]nodes.TaintSpec, len(taintRaw))
	for i, raw := range taintRaw {
		rawMap := raw.(map[string]interface{})
		taints[i] = nodes.TaintSpec{
			Key:    rawMap["key"].(string),
			Value:  rawMap["value"].(string),
			Effect: rawMap["effect"].(string),
		}
	}
	return taints
}

func resourceNodeAttachK8sOptions(d *schema.ResourceData) *nodes.ReinstallK8sOptionsConfig {
	if common.HasFilledOpt(d, "labels") || common.HasFilledOpt(d, "taints") || common.HasFilledOpt(d, "max_pods") {
		k8sOptions := nodes.ReinstallK8sOptionsConfig{
			Labels:  buildResourceNodeK8sTags(d),
			Taints:  buildResourceNodeTaint(d),
			MaxPods: d.Get("max_pods").(int),
		}
		return &k8sOptions
	}

	return nil
}

func resourceNodeAttachLifecycle(d *schema.ResourceData) *nodes.NodeLifecycleConfig {
	if common.HasFilledOpt(d, "preinstall") || common.HasFilledOpt(d, "postinstall") {
		lifecycle := nodes.NodeLifecycleConfig{
			PreInstall:  d.Get("preinstall").(string),
			PostInstall: d.Get("postinstall").(string),
		}
		return &lifecycle
	}
	return nil
}

func buildNodeAttachCreateOpts(d *schema.ResourceData) (*nodes.AcceptOpts, error) {
	result := nodes.AcceptOpts{
		Kind:       "List",
		ApiVersion: "v3",
		NodeList: []nodes.AddNode{
			{
				ServerID: d.Get("server_id").(string),
				Spec: nodes.ReinstallNodeSpec{
					OS:            d.Get("os").(string),
					Name:          d.Get("name").(string),
					ServerConfig:  resourceNodeAttachServerConfig(d),
					VolumeConfig:  resourceNodeAttachVolumeConfig(d),
					RuntimeConfig: resourceNodeAttachRuntimeConfig(d),
					K8sOptions:    resourceNodeAttachK8sOptions(d),
					Lifecycle:     resourceNodeAttachLifecycle(d),
				},
			},
		},
	}

	log.Printf("[DEBUG] Add node Options: %#v", result)
	// Add loginSpec here so it wouldn't go in the above log entry
	loginSpec, err := buildResourceNodeLoginSpec(d)
	if err != nil {
		diag.FromErr(err)
	}
	result.NodeList[0].Spec.Login = loginSpec
	return &result, nil
}

func resourceCCENodeV3AttachCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	// Wait for the cce cluster to become available
	clusterID := d.Get("cluster_id").(string)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating"},
		Target:     []string{"Available"},
		Refresh:    WaitForCCEClusterActive(client, clusterID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for CCE cluster to become available: %s", err)
	}

	addOpts, err := buildNodeAttachCreateOpts(d)
	addOpts.ClusterID = clusterID
	if err != nil {
		return diag.Errorf("error creating AddOpts structure of 'Add' method for CCE node attach: %s", err)
	}
	resp, err := nodes.Accept(client, *addOpts)
	if err != nil {
		return diag.Errorf("error adding node to the cluster (%s): %s", clusterID, err)
	}

	nodeID, err := getNodeIDFromJob(ctx, client, resp, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(nodeID)

	stateConf = &resource.StateChangeConf{
		// The statuses of pending phase includes "Build" and "Installing".
		Pending:      []string{"Installing"},
		Target:       []string{"Active"},
		Refresh:      waitForCceNodeActive(client, clusterID, nodeID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        20 * time.Second,
		PollInterval: 20 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for CCE node attach to the cluster: %s", err)
	}

	return resourceCCENodeV3Read(ctx, d, meta)
}

func buildNodeAttachUpdateOpts(d *schema.ResourceData) (*nodes.ResetOpts, error) {
	result := nodes.ResetOpts{
		Kind:       "List",
		ApiVersion: "v3",
		NodeList: []nodes.ResetNode{
			{
				NodeID: d.Id(),
				Spec: nodes.ReinstallNodeSpec{
					OS:            d.Get("os").(string),
					Name:          d.Get("name").(string),
					ServerConfig:  resourceNodeAttachServerConfig(d),
					VolumeConfig:  resourceNodeAttachVolumeConfig(d),
					RuntimeConfig: resourceNodeAttachRuntimeConfig(d),
					K8sOptions:    resourceNodeAttachK8sOptions(d),
					Lifecycle:     resourceNodeAttachLifecycle(d),
				},
			},
		},
	}

	log.Printf("[DEBUG] Add node Options: %#v", result)
	// Add loginSpec here so it wouldn't go in the above log entry
	loginSpec, err := buildResourceNodeLoginSpec(d)
	if err != nil {
		diag.FromErr(err)
	}
	result.NodeList[0].Spec.Login = loginSpec
	return &result, nil
}

func resourceCCENodeV3AttachUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	if d.HasChanges("name", "tags") {
		return resourceCCENodeV3Update(ctx, d, config)
	}

	if err != nil {
		return diag.Errorf("error creating CCE client: %s", err)
	}
	clusterID := d.Get("cluster_id").(string)

	resetOpts, err := buildNodeAttachUpdateOpts(d)
	resetOpts.ClusterID = clusterID
	if err != nil {
		return diag.Errorf("error creating ResetOpts structure of 'Reset' method for CCE node attach: %s", err)
	}
	resp, err := nodes.Reset(client, *resetOpts)
	if err != nil {
		return diag.Errorf("error resetting node: %s", err)
	}

	nodeID, err := getNodeIDFromJob(ctx, client, resp, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(nodeID)

	stateConf := &resource.StateChangeConf{
		// The statuses of pending phase includes "Build" and "Installing".
		Pending:      []string{"Installing"},
		Target:       []string{"Active"},
		Refresh:      waitForCceNodeActive(client, clusterID, nodeID),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        20 * time.Second,
		PollInterval: 20 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for CCE Node reset complete: %s", err)
	}

	return resourceCCENodeV3Read(ctx, d, config)
}

func resourceCCENodeV3AttachDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	clusterID := d.Get("cluster_id").(string)

	var removeOpts nodes.RemoveNodesOpts
	var loginSpec nodes.LoginSpec

	loginSpec, err = buildResourceNodeLoginSpec(d)
	if err != nil {
		diag.FromErr(err)
	}
	removeOpts.Spec.Login = loginSpec
	removeOpts.ClusterID = clusterID

	nodeItem := nodes.NodeItem{
		UID: d.Id(),
	}
	removeOpts.Spec.Nodes = append(removeOpts.Spec.Nodes, nodeItem)

	_, err = nodes.Remove(client, removeOpts)
	if err != nil {
		return diag.Errorf("error removing CCE node: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Deleting"},
		Target:     []string{"Deleted"},
		Refresh:    waitForCceNodeDelete(client, clusterID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error deleting CCE Node: %s", err)
	}
	return nil
}
