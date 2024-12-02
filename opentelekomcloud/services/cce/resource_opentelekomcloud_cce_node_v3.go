package cce

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
	nodesv1 "github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v1/nodes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/floatingips"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/bandwidths"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/eips"
	"github.com/unknwon/com"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/vpc"
)

var (
	predefinedTags = []string{
		"beta.kubernetes.io/arch",
		"beta.kubernetes.io/instance-type",
		"beta.kubernetes.io/os",
		"failure-domain.beta.kubernetes.io/region",
		"failure-domain.beta.kubernetes.io/zone",
		"kubernetes.io/arch",
		"kubernetes.io/hostname",
		"kubernetes.io/os",
		"node.kubernetes.io/baremetal",
		"node.kubernetes.io/subnetid",
		"node.kubernetes.io/container-engine",
		"node.kubernetes.io/instance-type",
		"os.architecture",
		"os.name",
		"os.version",
		"topology.kubernetes.io/region",
		"topology.kubernetes.io/zone",
	}

	predefinedTaints = []string{
		"node.kubernetes.io/unreachable",
		"node.cloudprovider.kubernetes.io/shutdown",
	}
)

func ResourceCCENodeV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCENodeV3Create,
		ReadContext:   resourceCCENodeV3Read,
		UpdateContext: resourceCCENodeV3Update,
		DeleteContext: resourceCCENodeV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: common.MultipleCustomizeDiffs(
			common.ValidateVolumeType("root_volume.*.volumetype"),
			common.ValidateVolumeType("data_volumes.*.volumetype"),
			common.ValidateSubnet("subnet_id"),
		),

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:          schema.TypeMap,
				ConflictsWith: []string{"tags"},
				Optional:      true,
				ForceNew:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_pair": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"os": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"root_volume": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
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
							Computed: true,
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
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
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
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					}},
			},
			"eip_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				ConflictsWith: []string{
					"iptype", "bandwidth_charge_mode", "bandwidth_size", "sharetype",
				},
			},
			"eip_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"eip_ids"},
			},
			"iptype": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"eip_ids"},
				Computed:      true,
			},
			"bandwidth_charge_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"eip_ids"},
				Computed:      true,
			},
			"sharetype": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"eip_ids"},
				Computed:      true,
			},
			"bandwidth_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"eip_ids"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"billing_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"extend_param_charging_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"ecs_performance_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"order_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"max_pods": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"agency_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"preinstall": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return common.InstallScriptHashSum(v)
					default:
						return ""
					}
				},
			},
			"postinstall": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return common.InstallScriptHashSum(v)
					default:
						return ""
					}
				},
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
			"dedicated_host_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": {
				Type:          schema.TypeMap,
				ConflictsWith: []string{"labels"},
				Optional:      true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"k8s_tags": {
				Type:         schema.TypeMap,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateK8sTagsMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"taints": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
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
					},
				},
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
		},
	}
}

func resourceCCENodeLabelsV2(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("labels").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func resourceCCENodeAnnotationsV2(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("annotations").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func resourceCCENodeTags(d *schema.ResourceData) []tags.ResourceTag {
	tagRaw := d.Get("tags").(map[string]interface{})
	return common.ExpandResourceTags(tagRaw)
}

func resourceCCENodeTaints(d *schema.ResourceData) []nodes.TaintSpec {
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

func resourceCCENodeK8sTags(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("k8s_tags").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func resourceCCEDataVolume(d *schema.ResourceData) []nodes.VolumeSpec {
	volumeRaw := d.Get("data_volumes").([]interface{})
	volumes := make([]nodes.VolumeSpec, len(volumeRaw))
	for i, raw := range volumeRaw {
		rawMap := raw.(map[string]interface{})
		volumes[i] = nodes.VolumeSpec{
			Size:        rawMap["size"].(int),
			VolumeType:  rawMap["volumetype"].(string),
			ExtendParam: rawMap["extend_params"].(map[string]interface{}),
		}
		if kmsID := rawMap["kms_id"]; kmsID != "" {
			volumes[i].Metadata = map[string]interface{}{
				"__system__cmkid":     kmsID,
				"__system__encrypted": "1",
			}
		}
	}
	return volumes
}

func resourceCCERootVolume(d *schema.ResourceData) nodes.VolumeSpec {
	var nics nodes.VolumeSpec
	nicsRaw := d.Get("root_volume").([]interface{})
	if len(nicsRaw) == 1 {
		rawMap := nicsRaw[0].(map[string]interface{})
		nics.Size = rawMap["size"].(int)
		nics.VolumeType = rawMap["volumetype"].(string)
		nics.ExtendParam = rawMap["extend_params"].(map[string]interface{})
		if kmsID := rawMap["kms_id"]; kmsID != "" {
			nics.Metadata = map[string]interface{}{
				"__system__cmkid":     kmsID,
				"__system__encrypted": "1",
			}
		}
	}
	return nics
}

func resourceCCEEipIDs(d *schema.ResourceData) []string {
	rawID := d.Get("eip_ids").(*schema.Set)
	id := make([]string, rawID.Len())
	for i, raw := range rawID.List() {
		id[i] = raw.(string)
	}
	return id
}

func explainNodesJob(job *nodes.Job) string {
	return fmt.Sprintf(`Job %s in status "%s":
Reason: %s
Message: %s
`, job.Metadata.ID, job.Status.Phase, job.Status.Reason, job.Status.Message)
}

func resourceCCENodeV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// eip_count and bandwidth_size parameters must be set simultaneously
	bandwidthSize := d.Get("bandwidth_size").(int)
	eipCount := d.Get("eip_count").(int)
	if bandwidthSize > 0 && eipCount == 0 {
		eipCount = 1
		checkCCENodeV3PublicIpParams(d)
	}

	createOpts := nodes.CreateOpts{
		Kind:       "Node",
		ApiVersion: "v3",
		Metadata: nodes.CreateMetaData{
			Name:        d.Get("name").(string),
			Labels:      resourceCCENodeLabelsV2(d),
			Annotations: resourceCCENodeAnnotationsV2(d),
		},
		Spec: nodes.Spec{
			Flavor:      d.Get("flavor_id").(string),
			Az:          d.Get("availability_zone").(string),
			Os:          d.Get("os").(string),
			Login:       nodes.LoginSpec{SshKey: d.Get("key_pair").(string)},
			RootVolume:  resourceCCERootVolume(d),
			DataVolumes: resourceCCEDataVolume(d),
			PublicIP: nodes.PublicIPSpec{
				Ids:   resourceCCEEipIDs(d),
				Count: eipCount,
				Eip: nodes.EipSpec{
					IpType: d.Get("iptype").(string),
					Bandwidth: nodes.BandwidthOpts{
						ChargeMode: d.Get("bandwidth_charge_mode").(string),
						Size:       d.Get("bandwidth_size").(int),
						ShareType:  d.Get("sharetype").(string),
					},
				},
			},
			DedicatedHostID: d.Get("dedicated_host_id").(string),
			NodeNicSpec: nodes.NodeNicSpec{
				PrimaryNic: nodes.PrimaryNic{
					SubnetId: d.Get("subnet_id").(string),
				},
			},
			BillingMode: d.Get("billing_mode").(int),
			Count:       1,
			ExtendParam: nodes.ExtendParam{
				ChargingMode:            d.Get("extend_param_charging_mode").(int),
				EcsPerformanceType:      d.Get("ecs_performance_type").(string),
				OrderID:                 d.Get("order_id").(string),
				ProductID:               d.Get("product_id").(string),
				PublicKey:               d.Get("public_key").(string),
				MaxPods:                 d.Get("max_pods").(int),
				PreInstall:              base64PreInstall,
				PostInstall:             base64PostInstall,
				DockerBaseSize:          d.Get("docker_base_size").(int),
				DockerLVMConfigOverride: d.Get("docker_lvm_config_override").(string),
				AgencyName:              d.Get("agency_name").(string),
			},
			UserTags: resourceCCENodeTags(d),
			K8sTags:  resourceCCENodeK8sTags(d),
			Taints:   resourceCCENodeTaints(d),
		},
	}

	if v, ok := d.GetOk("runtime"); ok {
		createOpts.Spec.Runtime = nodes.RuntimeSpec{
			Name: v.(string),
		}
	}

	if ip := d.Get("private_ip").(string); ip != "" {
		createOpts.Spec.NodeNicSpec.PrimaryNic.FixedIPs = []string{ip}
	}

	clusterID := d.Get("cluster_id").(string)
	stateCluster := &resource.StateChangeConf{
		Target:     []string{"Available"},
		Refresh:    waitForClusterAvailable(client, clusterID),
		Timeout:    15 * time.Minute,
		Delay:      15 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateCluster.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting cluster to become available: %w", err)
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	node, err := nodes.Create(client, clusterID, createOpts).Extract()
	switch err.(type) {
	case golangsdk.ErrDefault403:
		retryNode, err := recursiveCreate(ctx, client, createOpts, clusterID)
		if err == "fail" {
			return fmterr.Errorf("error creating OpenTelekomCloud Node")
		}
		node = retryNode
	case nil:
		break
	default:
		return fmterr.Errorf("error creating OpenTelekomCloud Node: %s", err)
	}

	nodeID, err := getNodeIDFromJob(ctx, client, node.Status.JobID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmterr.Errorf("error fetching OpenTelekomCloud Job Details: %s", err)
	}

	log.Printf("[DEBUG] Waiting for CCE Node (%s) to become available", node.Metadata.Name)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Build", "Installing"},
		Target:     []string{"Active"},
		Refresh:    waitForCceNodeActive(client, clusterID, nodeID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud CCE Node: %s", err)
	}

	d.SetId(nodeID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceCCENodeV3Read(clientCtx, d, meta)
}

// getNodeIDFromJob wait until job starts (status Running) and returns Node ID
func getNodeIDFromJob(ctx context.Context, client *golangsdk.ServiceClient, jobID string, timeout time.Duration) (string, error) {
	job, err := nodes.GetJobDetails(client, jobID).ExtractJob()
	if err != nil {
		return "", fmt.Errorf("error fetching OpenTelekomCloud Job Details: %s", err)
	}
	jobResourceId := job.Spec.SubJobs[0].Metadata.ID

	stateConf := &resource.StateChangeConf{
		Pending: []string{"Initializing"},
		Target:  []string{"Running"},
		Refresh: func() (interface{}, string, error) {
			subJob, err := nodes.GetJobDetails(client, jobResourceId).ExtractJob()
			if err != nil {
				return nil, "ERROR", fmt.Errorf("error fetching OpenTelekomCloud Job Details: %s", err)
			}
			return subJob, subJob.Status.Phase, nil
		},
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	j, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return "", fmt.Errorf("error creating OpenTelekomCloud CCE Node: %s", err)
	}
	log.Printf("job: %+v", j)
	subJob := j.(*nodes.Job)

	var nodeID string
	for _, s := range subJob.Spec.SubJobs {
		if s.Spec.Type == "CreateNodeVM" {
			nodeID = s.Spec.ResourceID
			break
		}
	}

	if nodeID == "" {
		return "", fmt.Errorf("can't find node ID in job response\n%s", explainNodesJob(subJob))
	}
	return nodeID, nil
}

func resourceCCENodeV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	clusterID := d.Get("cluster_id").(string)
	node, err := nodes.Get(client, clusterID, d.Id()).Extract()
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}
		return fmterr.Errorf("error retrieving OpenTelekomCloud Node: %w", err)
	}

	serverID := node.Status.ServerID
	mErr := multierror.Append(
		d.Set("region", config.GetRegion(d)),
		d.Set("name", node.Metadata.Name),
		d.Set("flavor_id", node.Spec.Flavor),
		d.Set("os", node.Spec.Os),
		d.Set("availability_zone", node.Spec.Az),
		d.Set("billing_mode", node.Spec.BillingMode),
		d.Set("key_pair", node.Spec.Login.SshKey),
		d.Set("server_id", serverID),
		d.Set("private_ip", node.Status.PrivateIP),
		d.Set("public_ip", node.Status.PublicIP),
		d.Set("status", node.Status.Phase),
		d.Set("subnet_id", node.Spec.NodeNicSpec.PrimaryNic.SubnetId),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving main conf to state for OpenTelekomCloud Node (%s): %w", d.Id(), err)
	}

	if node.Spec.Runtime.Name != "" {
		mErr = multierror.Append(mErr, d.Set("runtime", node.Spec.Runtime.Name))
	}

	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("Error setting 'runtime' for OpenTelekomCloud Node (%s): %w", d.Id(), err)
	}

	var volumes []map[string]interface{}
	for _, dataVolume := range node.Spec.DataVolumes {
		volume := make(map[string]interface{})
		volume["size"] = dataVolume.Size
		volume["volumetype"] = dataVolume.VolumeType
		volume["extend_params"] = dataVolume.ExtendParam
		volume["extend_param"] = ""
		if dataVolume.Metadata != nil {
			volume["kms_id"] = dataVolume.Metadata["__system__cmkid"]
		}
		volumes = append(volumes, volume)
	}
	if err := d.Set("data_volumes", volumes); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving dataVolumes to state for OpenTelekomCloud Node (%s): %w", d.Id(), err)
	}

	rootVolume := []map[string]interface{}{
		{
			"size":          node.Spec.RootVolume.Size,
			"volumetype":    node.Spec.RootVolume.VolumeType,
			"extend_params": node.Spec.RootVolume.ExtendParam,
			"extend_param":  "",
		},
	}
	if node.Spec.RootVolume.Metadata != nil {
		rootVolume[0]["kms_id"] = node.Spec.RootVolume.Metadata["__system__cmkid"]
	}

	if err := d.Set("root_volume", rootVolume); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving root Volume to state for OpenTelekomCloud Node (%s): %w", d.Id(), err)
	}

	// fetch tags from ECS instance
	computeClient, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud ComputeV1 client: %w", err)
	}

	resourceTags, err := tags.Get(computeClient, "cloudservers", serverID).Extract()
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}
		return fmterr.Errorf("error fetching OpenTelekomCloud instance tags: %w", err)
	}

	tagMap := common.TagsToMap(resourceTags)
	// ignore "CCE-Dynamic-Provisioning-Node"
	delete(tagMap, "CCE-Dynamic-Provisioning-Node")
	delete(tagMap, "CCE-Cluster-ID")
	if err := d.Set("tags", tagMap); err != nil {
		return fmterr.Errorf("error saving tags of CCE node: %w", err)
	}

	if err := setK8sNodeFields(d, config, clusterID, node.Status.PrivateIP); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// cceV1Client for swiss
func cceV1Client(config *cfg.Config, region string) (*golangsdk.ServiceClient, error) {
	v1Client, err := config.CceV1Client(region)
	if err == nil { // for eu-de
		return v1Client, nil
	}
	client, err := config.CceV3Client(region)
	if err != nil {
		return nil, fmt.Errorf("both v1 and v3 clients are not available for %s region: %w", region, err)
	}
	client.ResourceBase = client.Endpoint + "api/v1/"

	return client, nil
}

func setK8sNodeFields(d *schema.ResourceData, config *cfg.Config, clusterID, privateIP string) error {
	client, err := cceV1Client(config, config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud CCEv1 client: %w", err)
	}
	k8Node, err := nodesv1.Get(client, clusterID, privateIP).Extract()
	if err != nil {
		log.Printf("[WARN] error retrieving CCE node: %s", err.Error())
		return nil
	}
	taints := make([]interface{}, 0)
	for _, value := range k8Node.Spec.Taints {
		if com.IsSliceContainsStr(predefinedTaints, value.Key) {
			continue
		}
		taints = append(taints, map[string]interface{}{
			"key":    value.Key,
			"value":  value.Value,
			"effect": value.Effect,
		})
	}
	if err := d.Set("taints", taints); err != nil {
		return fmt.Errorf("error setting taints for CCE Node: %w", err)
	}
	k8sTags := make(map[string]interface{})
	for key, value := range k8Node.Metadata.Labels {
		if com.IsSliceContainsStr(predefinedTags, key) {
			continue
		}
		k8sTags[key] = value
	}
	if err := d.Set("k8s_tags", k8sTags); err != nil {
		return fmt.Errorf("error setting k8s_tags for CCE Node: %w", err)
	}
	return nil
}

func resourceCCENodeV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	var updateOpts nodes.UpdateOpts

	if d.HasChange("name") {
		updateOpts.Metadata.Name = d.Get("name").(string)

		clusterID := d.Get("cluster_id").(string)
		_, err = nodes.Update(client, clusterID, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud CCE node: %s", err)
		}
	}

	// update tags
	if d.HasChange("tags") {
		computeV1Client, err := config.ComputeV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud ComputeV1 client: %s", err)
		}

		serverID := d.Get("server_id").(string)
		tagErr := common.UpdateResourceTags(computeV1Client, d, "cloudservers", serverID)
		if tagErr != nil {
			return fmterr.Errorf("error updating tags of CCE node %s: %s", d.Id(), tagErr)
		}
	}

	// release, change bandwidth size or create and assign ip
	if d.HasChange("bandwidth_size") {
		oldBandwidthSize, newBandWidthSize := d.GetChange("bandwidth_size")
		newBandWidth := newBandWidthSize.(int)
		oldBandwidth := oldBandwidthSize.(int)
		serverId := d.Get("server_id").(string)
		computeV2client, err := config.ComputeV2Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud ComputeV2 client: %w", err)
		}
		floatingIp, err := getCCENodeV3FloatingIp(computeV2client, serverId)
		if err != nil {
			return diag.FromErr(err)
		}
		if newBandWidth == 0 {
			err = deleteCCENodeV3FloatingIP(computeV2client, serverId, floatingIp.ID)
		} else {
			checkCCENodeV3PublicIpParams(d)
			if oldBandwidth > 0 {
				err = resizeCCENodeV3IpBandwidth(d, config, floatingIp.ID, newBandWidth)
			} else {
				err = createAndAssociateCCENodeV3FloatingIp(ctx, d, config, newBandWidth, serverId)
			}
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("eip_ids") {
		oldEipIdsRaw, newEipIdsRaw := d.GetChange("eip_ids")
		oldEipIds := oldEipIdsRaw.(*schema.Set).List()
		newEipIds := newEipIdsRaw.(*schema.Set).List()
		serverId := d.Get("server_id").(string)
		computeV2Client, err := config.ComputeV2Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud ComputeV2 client: %s", err)
		}
		if len(newEipIds) == 0 {
			if err := unbindCCENodeV3FloatingIP(computeV2Client, serverId, oldEipIds[0].(string)); err != nil {
				return diag.FromErr(err)
			}
		} else if len(oldEipIds) > 0 {
			err = reassignCCENodeV3Eip(computeV2Client, oldEipIds[0].(string), newEipIds[0].(string), serverId)
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceCCENodeV3Read(clientCtx, d, meta)
}

func resourceCCENodeV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	clusterID := d.Get("cluster_id").(string)
	if err := nodes.Delete(client, clusterID, d.Id()).ExtractErr(); err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud CCE Cluster: %w", err)
	}
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Deleting"},
		Target:     []string{"Deleted"},
		Refresh:    waitForCceNodeDelete(client, clusterID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud CCE Node: %w", err)
	}

	d.SetId("")
	return nil
}

func deleteCCENodeV3FloatingIP(client *golangsdk.ServiceClient, serverId string, floatingIpId string) error {
	err := unbindCCENodeV3FloatingIP(client, serverId, floatingIpId)
	if err != nil {
		return fmt.Errorf("error unbind floatingip from the node: %s", err)
	}
	err = floatingips.Delete(client, floatingIpId).ExtractErr()
	if err != nil {
		return fmt.Errorf("error delete floatingip: %s", err)
	}
	return nil
}

func unbindCCENodeV3FloatingIP(client *golangsdk.ServiceClient, serverId string, floatingIpId string) error {
	eip, err := floatingips.Get(client, floatingIpId).Extract()
	if err != nil {
		return fmt.Errorf("error get eip by id: %s", err)
	}

	disassociateOpts := floatingips.DisassociateOpts{
		FloatingIP: eip.IP,
	}

	err = floatingips.DisassociateInstance(client, serverId, disassociateOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("error unassign floating IP from CCE Node")
	}
	return nil
}

func resizeCCENodeV3IpBandwidth(d *schema.ResourceData, meta interface{}, eipId string, newSize int) error {
	config := meta.(*cfg.Config)
	nwClient, err := config.NetworkingV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud NetworkingV1 client: %s", err)
	}
	elasticIp, err := eips.Get(nwClient, eipId).Extract()
	if err != nil {
		return err
	}

	updateOpts := bandwidths.UpdateOpts{Size: newSize}

	_, err = bandwidths.Update(nwClient, elasticIp.BandwidthID, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("error updating bandwidth size: %s", err)
	}
	return nil
}

func createAndAssociateCCENodeV3FloatingIp(ctx context.Context, d *schema.ResourceData, meta interface{}, size int, serverId string) error {
	config := meta.(*cfg.Config)
	nwClient, err := config.NetworkingV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud NetworkingV1 client: %s", err)
	}
	createEipOpts := vpc.EIPCreateOpts{
		ApplyOpts: eips.ApplyOpts{
			IP: eips.PublicIpOpts{
				Type: d.Get("iptype").(string),
			},
			Bandwidth: eips.BandwidthOpts{
				Name:       "bandwidth-cce-node-1",
				Size:       size,
				ShareType:  d.Get("sharetype").(string),
				ChargeMode: d.Get("bandwidth_charge_mode").(string),
			},
		},
	}

	eip, err := eips.Apply(nwClient, createEipOpts).Extract()
	if err != nil {
		return fmt.Errorf("error updating bandwidth size: %s", err)
	}

	err = vpc.WaitForEIPActive(ctx, nwClient, eip.ID, time.Minute*10)
	if err != nil {
		return fmt.Errorf("error waiting for EIP (%s) to become ready: %s", eip.ID, err)
	}

	computeClient, err := config.ComputeV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud ComputeV2 client: %s", err)
	}
	associateOpts := floatingips.AssociateOpts{
		FloatingIP: eip.PublicAddress,
	}
	if err := floatingips.AssociateInstance(computeClient, serverId, associateOpts).ExtractErr(); err != nil {
		return fmt.Errorf("error associating CCE Node to publicIp: %s", err)
	}

	return nil
}

func reassignCCENodeV3Eip(client *golangsdk.ServiceClient, oldEipId string, newEipId string, serverId string) error {
	oldEip, err := floatingips.Get(client, oldEipId).Extract()
	if err != nil {
		return fmt.Errorf("error get eip by id: %s", err)
	}
	newEip, err := floatingips.Get(client, newEipId).Extract()
	if err != nil {
		return fmt.Errorf("error get eip by id: %s", err)
	}

	disassociateOpts := floatingips.DisassociateOpts{
		FloatingIP: oldEip.IP,
	}
	err = floatingips.DisassociateInstance(client, serverId, disassociateOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("error unassign floating IP from CCE Node")
	}

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: newEip.IP,
	}
	err = floatingips.AssociateInstance(client, serverId, associateOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("error assign floating IP to CCE Node")
	}
	return nil
}

func getCCENodeV3FloatingIp(client *golangsdk.ServiceClient, serverId string) (*floatingips.FloatingIP, error) {
	fipPages, err := floatingips.List(client).AllPages()
	if err != nil {
		return nil, err
	}
	fips, err := floatingips.ExtractFloatingIPs(fipPages)
	if err != nil {
		return nil, err
	}
	var floatingIp floatingips.FloatingIP
	for _, ip := range fips {
		if ip.InstanceID == serverId {
			floatingIp = ip
		}
	}
	return &floatingIp, nil
}

func checkCCENodeV3PublicIpParams(d *schema.ResourceData) {
	if d.Get("bandwidth_charge_mode").(string) == "" {
		_ = d.Set("bandwidth_charge_mode", "traffic")
	}
	if d.Get("sharetype").(string) == "" {
		_ = d.Set("sharetype", "PER")
	}
	if d.Get("iptype").(string) == "" {
		_ = d.Set("iptype", "5_bgp")
	}
}

func waitForCceNodeActive(cceClient *golangsdk.ServiceClient, clusterId, nodeId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := nodes.Get(cceClient, clusterId, nodeId).Extract()
		if err != nil {
			return nil, "", err
		}

		return n, n.Status.Phase, nil
	}
}

func waitForCceNodeDelete(cceClient *golangsdk.ServiceClient, clusterId, nodeId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete OpenTelekomCloud CCE Node %s.\n", nodeId)

		r, err := nodes.Get(cceClient, clusterId, nodeId).Extract()

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[DEBUG] Successfully deleted OpenTelekomCloud CCE Node %s", nodeId)
				return r, "Deleted", nil
			}
			return r, "Deleting", err
		}

		log.Printf("[DEBUG] OpenTelekomCloud CCE Node %s still available.\n", nodeId)
		return r, r.Status.Phase, nil
	}
}

func waitForClusterAvailable(cceClient *golangsdk.ServiceClient, clusterId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[INFO] Waiting for OpenTelekomCloud Cluster to be available %s.\n", clusterId)
		n, err := clusters.Get(cceClient, clusterId)

		if err != nil {
			return nil, "", err
		}

		return n, n.Status.Phase, nil
	}
}

func recursiveCreate(ctx context.Context, client *golangsdk.ServiceClient, opts nodes.CreateOptsBuilder, clusterID string) (*nodes.Nodes, string) {
	stateCluster := &resource.StateChangeConf{
		Target:     []string{"Available"},
		Refresh:    waitForClusterAvailable(client, clusterID),
		Timeout:    15 * time.Minute,
		Delay:      15 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, stateErr := stateCluster.WaitForStateContext(ctx)
	if stateErr != nil {
		log.Printf("[INFO] Cluster Unavailable %s.\n", stateErr)
	}
	node, err := nodes.Create(client, clusterID, opts).Extract()
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault403); ok {
			return recursiveCreate(ctx, client, opts, clusterID)
		}
		return node, "fail"
	}
	return node, "success"
}
