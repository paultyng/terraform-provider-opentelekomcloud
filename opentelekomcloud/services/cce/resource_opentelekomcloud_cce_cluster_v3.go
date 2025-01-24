package cce

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/addons"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/subnets"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/vpcs"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/security/groups"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

var (
	// Cluster name is 4 to 128 characters starting with a letter and not ending with a hyphen (-).
	// Only lowercase letters, digits, and hyphens (-) are allowed
	clusterNameRegex = regexp.MustCompile("^[a-z][a-z0-9-]{2,126}[a-z0-9]$")
)

func ResourceCCEClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCEClusterV3Create,
		ReadContext:   resourceCCEClusterV3Read,
		UpdateContext: resourceCCEClusterV3Update,
		DeleteContext: resourceCCEClusterV3Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			validateCCEClusterNetwork,
			validateAuthProxy,
		),

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(clusterNameRegex, "Invalid cluster name. "+
					"Cluster name should be 4 to 128 characters starting with a letter and not ending with a hyphen (-). "+
					"Only lowercase letters, digits, and hyphens (-) are allowed."),
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_version": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				DiffSuppressFunc: common.SuppressSmartVersionDiff,
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"billing_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
			"highway_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"extend_param": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"enable_volume_encryption": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"container_network_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_network_cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"eni_subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				RequiredWith: []string{"eni_subnet_cidr"},
			},
			"eni_subnet_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				RequiredWith: []string{"eni_subnet_id"},
			},
			"authentication_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "x509",
			},
			"masters": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				MaxItems:      3,
				ConflictsWith: []string{"multi_az"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
					},
				},
			},
			"authenticating_proxy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ca": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"cert": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"private_key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"authenticating_proxy_ca": {
				Type:       schema.TypeString,
				Optional:   true,
				ForceNew:   true,
				Deprecated: "Please use `authenticating_proxy` instead",
			},
			"kubernetes_svc_ip_range": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"kube_proxy_mode": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ipvs", "iptables"}, true),
			},
			"multi_az": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"masters"},
			},
			"eip": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: common.ValidateIP,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"internal": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_otc": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ignore_certificate_clusters_data": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"certificate_clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"server": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"certificate_authority_data": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ignore_certificate_users_data": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"certificate_users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_certificate_data": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key_data": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"no_addons": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"ignore_addons": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"no_addons"},
			},
			"installed_addons": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"security_group_control": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_node": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_efs": associateDeleteSchema,
			"delete_eni": associateDeleteSchema,
			"delete_evs": associateDeleteSchema,
			"delete_net": associateDeleteSchema,
			"delete_obs": associateDeleteSchema,
			"delete_sfs": associateDeleteSchema,
			"delete_all_storage": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"true", "try", "false",
				}, true),
			},
			"delete_all_network": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"true", "try", "false",
				}, true),
			},
		},
	}
}

var associateDeleteSchema *schema.Schema = &schema.Schema{
	Type:     schema.TypeString,
	Optional: true,
	ValidateFunc: validation.StringInSlice([]string{
		"true", "try", "false",
	}, true),
	ConflictsWith: []string{"delete_all_storage", "delete_all_network"},
}

func resourceClusterMasters(d *schema.ResourceData) ([]clusters.MasterSpec, error) {
	if v, ok := d.GetOk("masters"); ok {
		flavorId := d.Get("flavor_id").(string)
		mastersRaw := v.([]interface{})
		if strings.Contains(flavorId, "s1") && len(mastersRaw) != 1 {
			return nil, fmt.Errorf("error creating CCE cluster: "+
				"single-master cluster need 1 az for master node, but got %d", len(mastersRaw))
		}
		if strings.Contains(flavorId, "s2") && len(mastersRaw) != 3 {
			return nil, fmt.Errorf("error creating CCE cluster: "+
				"high-availability cluster need 3 az for master nodes, but got %d", len(mastersRaw))
		}
		masters := make([]clusters.MasterSpec, len(mastersRaw))
		for i, raw := range mastersRaw {
			rawMap := raw.(map[string]interface{})
			masters[i] = clusters.MasterSpec{
				AvailabilityZone: rawMap["availability_zone"].(string),
			}
		}
		return masters, nil
	}

	return nil, nil
}

func resourceClusterLabelsV3(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("labels").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}
func resourceClusterAnnotationsV3(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("annotations").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}
func resourceClusterExtendParamV3(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("extend_param").(map[string]interface{}) {
		m[key] = val.(string)
	}
	if multiAZ, ok := d.GetOk("multi_az"); ok && multiAZ.(bool) {
		m["clusterAZ"] = "multi_az"
	}
	if eip, ok := d.GetOk("eip"); ok {
		m["clusterExternalIP"] = eip.(string)
	}
	return m
}

func resourceCCEClusterV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	if d.Get("eip").(string) != "" {
		fipId, err := resourceFloatingIPV2Exists(d, meta, d.Get("eip").(string))
		if err != nil {
			return fmterr.Errorf("error retrieving the eip: %w", err)
		}
		if fipId == "" {
			return fmterr.Errorf("the specified EIP %s does not exist", d.Get("eip").(string))
		}
	}

	authProxy := map[string]string{}
	if ca, ok := d.GetOk("authenticating_proxy_ca"); ok {
		authProxy = map[string]string{
			"ca": common.Base64IfNot(ca.(string)),
		}
	} else if _, ok := d.GetOk("authenticating_proxy"); ok {
		authProxy = getAuthProxy(d)
	}

	createOpts := clusters.CreateOpts{
		Kind:       "Cluster",
		ApiVersion: "v3",
		Metadata: clusters.CreateMetaData{
			Name:        d.Get("name").(string),
			Labels:      resourceClusterLabelsV3(d),
			Annotations: resourceClusterAnnotationsV3(d),
		},
		Spec: clusters.Spec{
			Type:        d.Get("cluster_type").(string),
			Flavor:      d.Get("flavor_id").(string),
			Version:     d.Get("cluster_version").(string),
			Description: d.Get("description").(string),
			HostNetwork: clusters.HostNetworkSpec{
				VpcId:           d.Get("vpc_id").(string),
				SubnetId:        d.Get("subnet_id").(string),
				HighwaySubnet:   d.Get("highway_subnet_id").(string),
				SecurityGroupId: d.Get("security_group_id").(string),
			},
			ContainerNetwork: clusters.ContainerNetworkSpec{
				Mode: d.Get("container_network_type").(string),
				Cidr: d.Get("container_network_cidr").(string),
			},
			Authentication: clusters.AuthenticationSpec{
				Mode:                d.Get("authentication_mode").(string),
				AuthenticatingProxy: authProxy,
			},
			BillingMode:                  d.Get("billing_mode").(int),
			ExtendParam:                  resourceClusterExtendParamV3(d),
			KubernetesSvcIpRange:         d.Get("kubernetes_svc_ip_range").(string),
			KubeProxyMode:                d.Get("kube_proxy_mode").(string),
			EnableMasterVolumeEncryption: pointerto.Bool(d.Get("enable_volume_encryption").(bool)),
		},
	}

	if _, ok := d.GetOk("eni_subnet_id"); ok {
		eniNetwork := clusters.EniNetworkSpec{
			SubnetId: d.Get("eni_subnet_id").(string),
			Cidr:     d.Get("eni_subnet_cidr").(string),
		}
		createOpts.Spec.EniNetwork = &eniNetwork
	}

	masters, err := resourceClusterMasters(d)
	if err != nil {
		return diag.FromErr(err)
	}
	createOpts.Spec.Masters = masters

	create, err := clusters.Create(client, createOpts)

	if err != nil {
		if isAuthRequired(err) {
			err = fmt.Errorf("CCE is not authorized, see `cce_cluster_v3` documentation for details")
		}
		return fmterr.Errorf("error creating opentelekomcloud Cluster: %w", err)
	}

	log.Printf("[DEBUG] Waiting for opentelekomcloud CCE cluster (%s) to become available", create.Metadata.Id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating"},
		Target:     []string{"Available"},
		Refresh:    WaitForCCEClusterActive(client, create.Metadata.Id),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud CCE cluster: %s", err)
	}
	d.SetId(create.Metadata.Id)

	if ignore := d.Get("ignore_addons").(bool); !ignore {
		if err := waitForInstalledAddons(ctx, d, config); err != nil {
			return fmterr.Errorf("error waiting for default addons to install")
		}
		if d.Get("no_addons").(bool) {
			if err := removeAddons(ctx, d, config); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceCCEClusterV3Read(clientCtx, d, meta)
}

func resourceCCEClusterV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	cluster, err := clusters.Get(client, d.Id())
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}
		return fmterr.Errorf("error retrieving opentelekomcloud CCE: %w", err)
	}

	eip := ""
	if cluster.Status.Endpoints[0].External != "" {
		endpointURL, err := url.Parse(cluster.Status.Endpoints[0].External)
		if err != nil {
			return fmterr.Errorf("error parsing endpoint URL: %w", err)
		}
		eip = endpointURL.Hostname()
	}

	mErr := multierror.Append(nil,
		d.Set("name", cluster.Metadata.Name),
		d.Set("status", cluster.Status.Phase),
		d.Set("flavor_id", cluster.Spec.Flavor),
		d.Set("cluster_type", cluster.Spec.Type),
		d.Set("cluster_version", cluster.Spec.Version),
		d.Set("description", cluster.Spec.Description),
		d.Set("billing_mode", cluster.Spec.BillingMode),
		d.Set("vpc_id", cluster.Spec.HostNetwork.VpcId),
		d.Set("subnet_id", cluster.Spec.HostNetwork.SubnetId),
		d.Set("security_group_id", cluster.Spec.HostNetwork.SecurityGroupId),
		d.Set("highway_subnet_id", cluster.Spec.HostNetwork.HighwaySubnet),
		d.Set("container_network_type", cluster.Spec.ContainerNetwork.Mode),
		d.Set("container_network_cidr", cluster.Spec.ContainerNetwork.Cidr),
		d.Set("eni_subnet_id", cluster.Spec.EniNetwork.SubnetId),
		d.Set("eni_subnet_cidr", cluster.Spec.EniNetwork.Cidr),
		d.Set("authentication_mode", cluster.Spec.Authentication.Mode),
		d.Set("kubernetes_svc_ip_range", cluster.Spec.KubernetesSvcIpRange),
		d.Set("kube_proxy_mode", cluster.Spec.KubeProxyMode),
		d.Set("internal", cluster.Status.Endpoints[0].Internal),
		d.Set("external", cluster.Status.Endpoints[0].External),
		d.Set("external_otc", cluster.Status.Endpoints[0].ExternalOTC),
		d.Set("region", config.GetRegion(d)),
		d.Set("eip", eip),
		d.Set("enable_volume_encryption", cluster.Spec.EnableMasterVolumeEncryption),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting cce cluster fields: %w", err)
	}

	cert, err := clusters.GetCert(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error retrieving opentelekomcloud CCE cluster cert: %w", err)
	}

	var clusterList []map[string]interface{}
	if !d.Get("ignore_certificate_clusters_data").(bool) {
		// Set Certificate Clusters
		for _, clusterObj := range cert.Clusters {
			clusterCert := map[string]interface{}{
				"name":                       clusterObj.Name,
				"server":                     clusterObj.Cluster.Server,
				"certificate_authority_data": clusterObj.Cluster.CertAuthorityData,
			}
			clusterList = append(clusterList, clusterCert)
		}
	}
	if err := d.Set("certificate_clusters", clusterList); err != nil {
		return diag.FromErr(err)
	}

	var userList []map[string]interface{}
	if !d.Get("ignore_certificate_users_data").(bool) {
		// Set Certificate Users
		for _, userObj := range cert.Users {
			userCert := map[string]interface{}{
				"name":                    userObj.Name,
				"client_certificate_data": userObj.User.ClientCertData,
				"client_key_data":         userObj.User.ClientKeyData,
			}
			userList = append(userList, userCert)
		}
	}
	if err := d.Set("certificate_users", userList); err != nil {
		return diag.FromErr(err)
	}

	if ignore := d.Get("ignore_addons").(bool); !ignore {
		instances, err := listInstalledAddons(d, config)
		if err != nil {
			return fmterr.Errorf("error listing installed addons: %w", err)
		}
		installedAddons := make([]string, len(instances.Items))
		for i, instance := range instances.Items {
			installedAddons[i] = instance.Metadata.Id
		}
		if err := d.Set("installed_addons", installedAddons); err != nil {
			return fmterr.Errorf("error setting installed addons: %w", err)
		}
	}

	nwV2Client, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud NetworkingV2 client: %w", err)
	}
	securityGroupPages, err := groups.List(nwV2Client, groups.ListOpts{}).AllPages()
	if err != nil {
		return diag.FromErr(err)
	}
	securityGroups, err := groups.ExtractGroups(securityGroupPages)
	if err != nil {
		return diag.FromErr(err)
	}

	var controlSecGroupID string
	var nodeSecGroupID string
	for _, v := range securityGroups {
		if controlSecGroupID != "" && nodeSecGroupID != "" {
			break
		}
		if !strings.Contains(v.Description, d.Id()) {
			continue
		}
		if strings.Contains(v.Description, "master port") {
			controlSecGroupID = v.ID
			continue
		}
		if strings.Contains(v.Description, "node") {
			nodeSecGroupID = v.ID
			continue
		}
	}

	// Set masters
	var masterList []map[string]interface{}
	for _, masterObj := range cluster.Spec.Masters {
		master := make(map[string]interface{})
		master["availability_zone"] = masterObj.AvailabilityZone
		masterList = append(masterList, master)
	}
	if err := d.Set("masters", masterList); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("security_group_control", controlSecGroupID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("security_group_node", nodeSecGroupID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCCEClusterV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	var updateOpts clusters.UpdateOpts

	if d.HasChange("description") {
		updateOpts.Spec.Description = d.Get("description").(string)
		_, err = clusters.Update(client, d.Id(), updateOpts)
		if err != nil {
			return fmterr.Errorf("error updating opentelekomcloud CCE: %w", err)
		}
	}

	if d.HasChange("eip") {
		oldEip, newEip := d.GetChange("eip")
		oldEipStr := oldEip.(string)
		newEipStr := newEip.(string)
		var fipId string
		if newEipStr != "" {
			fipId, err = resourceFloatingIPV2Exists(d, meta, newEipStr)
			if err != nil {
				return fmterr.Errorf("error retrieving the eip: %w", err)
			}
			if fipId == "" {
				return fmterr.Errorf("the specified EIP %s does not exist", newEipStr)
			}
		}
		if oldEipStr != "" {
			updateIpOpts := clusters.UpdateIpOpts{
				Action: "unbind",
			}
			err = clusters.UpdateMasterIp(client, d.Id(), updateIpOpts)
			if err != nil {
				return fmterr.Errorf("error unbinding EIP to opentelekomcloud CCE: %w", err)
			}
		}
		if newEipStr != "" {
			updateIpOpts := clusters.UpdateIpOpts{
				Action:    "bind",
				ElasticIp: newEipStr,
			}
			updateIpOpts.Spec.ID = fipId
			err = clusters.UpdateMasterIp(client, d.Id(), updateIpOpts)
			if err != nil {
				return fmterr.Errorf("error binding EIP to opentelekomcloud CCE: %w", err)
			}
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceCCEClusterV3Read(clientCtx, d, meta)
}

func resourceCCEClusterV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.CceV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	deleteOpts := clusters.DeleteQueryParams{}
	var deleteAll bool
	if v, ok := d.GetOk("delete_all_storage"); ok && v.(string) != "false" {
		deleteOpt := d.Get("delete_all_storage").(string)
		deleteOpts.DeleteEfs = deleteOpt
		deleteOpts.DeleteEvs = deleteOpt
		deleteOpts.DeleteObs = deleteOpt
		deleteOpts.DeleteSfs = deleteOpt
		deleteAll = true
	}
	if v, ok := d.GetOk("delete_all_network"); ok && v.(string) != "false" {
		deleteOpt := d.Get("delete_all_network").(string)
		deleteOpts.DeleteENI = deleteOpt
		deleteOpts.DeleteNet = deleteOpt
		deleteAll = true
	}

	if !deleteAll {
		deleteOpts.DeleteEfs = d.Get("delete_efs").(string)
		deleteOpts.DeleteENI = d.Get("delete_eni").(string)
		deleteOpts.DeleteEvs = d.Get("delete_evs").(string)
		deleteOpts.DeleteNet = d.Get("delete_net").(string)
		deleteOpts.DeleteObs = d.Get("delete_obs").(string)
		deleteOpts.DeleteSfs = d.Get("delete_sfs").(string)
	}

	err = clusters.Delete(client, d.Id(), deleteOpts)
	if err != nil {
		return fmterr.Errorf("error deleting opentelekomcloud CCE Cluster: %w", err)
	}
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Deleting", "Available", "Unavailable"},
		Target:     []string{"Deleted"},
		Refresh:    WaitForCCEClusterDelete(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return fmterr.Errorf("error deleting opentelekomcloud CCE cluster: %w", err)
	}

	d.SetId("")
	return nil
}

func WaitForCCEClusterActive(cceClient *golangsdk.ServiceClient, clusterId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := clusters.Get(cceClient, clusterId)
		if err != nil {
			return nil, "", fmt.Errorf("error waiting for CCE cluster to become active: %w", err)
		}

		return n, n.Status.Phase, nil
	}
}

func WaitForCCEClusterDelete(client *golangsdk.ServiceClient, clusterId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete  CCE cluster %s.\n", clusterId)

		r, err := clusters.Get(client, clusterId)

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[DEBUG] Successfully deleted opentelekomcloud CCE cluster %s", clusterId)
				return r, "Deleted", nil
			}
			return nil, "", fmt.Errorf("error waiting CCE cluster to become deleted: %w", err)
		}
		if r.Status.Phase == "Deleting" {
			return r, "Deleting", nil
		}
		log.Printf("[DEBUG] opentelekomcloud CCE cluster %s still available.\n", clusterId)
		return r, "Available", nil
	}
}

func resourceFloatingIPV2Exists(d *schema.ResourceData, meta interface{}, floatingIP string) (string, error) {
	config := meta.(*cfg.Config)
	networkClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return "", fmt.Errorf("error creating opentelekomcloud networking Client: %w", err)
	}
	listOpts := floatingips.ListOpts{
		FloatingIP: floatingIP,
	}
	allPages, err := floatingips.List(networkClient, listOpts).AllPages()
	if err != nil {
		return "", fmt.Errorf("error listing floating IPs: %w", err)
	}

	allFips, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return "", fmt.Errorf("error extracting floating IPs: %w", err)
	}

	if len(allFips) == 0 {
		return "", nil
	}

	return allFips[0].ID, nil
}

func validateCCEClusterNetwork(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	config, ok := meta.(*cfg.Config)
	if !ok {
		return fmt.Errorf("error retreiving configuration: can't convert %v to Config", meta)
	}
	vpcClient, err := config.NetworkingV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf(cceClientError, err)
	}

	if vpcID := d.Get("vpc_id").(string); vpcID != "" {
		if err = vpcs.Get(vpcClient, vpcID).Err; err != nil {
			return fmt.Errorf("can't find VPC `%s`: %w", vpcID, err)
		}
	}

	if subnetID := d.Get("subnet_id").(string); subnetID != "" {
		if err = subnets.Get(vpcClient, subnetID).Err; err != nil {
			return fmt.Errorf("can't find subnet `%s`: %w", subnetID, err)
		}
	}

	return nil
}

func listInstalledAddons(d *schema.ResourceData, config *cfg.Config) (*addons.AddonInstanceList, error) {
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return nil, fmt.Errorf("error creating CCE Addon client: %w", logHttpError(err))
	}
	return addons.ListAddonInstances(client, d.Id())
}

func waitForInstalledAddons(ctx context.Context, d *schema.ResourceData, config *cfg.Config) error {
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating CCE Addon client: %w", logHttpError(err))
	}
	// First wait for addons to be assigned
	stateConfExist := &resource.StateChangeConf{
		Pending:    []string{"Deleted"},
		Target:     []string{"Available"},
		Refresh:    waitForCCEClusterAddonsState(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      30 * time.Second,
		MinTimeout: 1 * time.Minute,
	}

	if _, err := stateConfExist.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for addons to be assigned: %w", err)
	}

	return nil
}

func removeAddons(ctx context.Context, d *schema.ResourceData, config *cfg.Config) error {
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating CCE Addon client: %w", logHttpError(err))
	}

	instances, err := addons.ListAddonInstances(client, d.Id())
	if err != nil {
		return fmt.Errorf("error listing cluster addons: %w", err)
	}
	for _, addon := range instances.Items {
		addonID := addon.Metadata.Id
		stateConfAddonReady := &resource.StateChangeConf{
			Pending:    []string{"installing"},
			Target:     []string{"running", "available", "abnormal"},
			Refresh:    waitForCCEClusterAddonActive(client, addonID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      30 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := stateConfAddonReady.WaitForStateContext(ctx); err != nil {
			return fmt.Errorf("error waiting for addons to be installed: %w", err)
		}
	}

	for _, instance := range instances.Items {
		addonID := instance.Metadata.Id
		if err := addons.Delete(client, addonID, d.Id()); err != nil {
			return fmt.Errorf("error deleting cluster addon %s/%s: %w", d.Id(), addonID, err)
		}
	}
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Available"},
		Target:     []string{"Deleted"},
		Refresh:    waitForCCEClusterAddonsState(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for addons to be removed: %w", err)
	}

	return nil
}

func waitForCCEClusterAddonsState(client *golangsdk.ServiceClient, clusterID string) resource.StateRefreshFunc {
	return func() (r interface{}, s string, err error) {
		instances, err := addons.ListAddonInstances(client, clusterID)
		if err != nil {
			return nil, "", fmt.Errorf("error listing cluster addons: %w", err)
		}
		if len(instances.Items) > 0 {
			return instances, "Available", nil
		}
		return instances, "Deleted", nil
	}
}

func waitForCCEClusterAddonActive(client *golangsdk.ServiceClient, id, clusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := addons.Get(client, id, clusterID)
		if err != nil {
			return nil, "", err
		}

		return n, n.Status.Status, nil
	}
}

func isAuthRequired(err error) bool {
	authRequiredRegex := regexp.MustCompile(`.+\sauthorize\sCCE.+`)

	if err400, ok := err.(golangsdk.ErrDefault400); ok {
		var body struct {
			Message string `json:"message"`
		}
		if jsonErr := json.Unmarshal(err400.Body, &body); jsonErr != nil {
			return false // return original error
		}

		if authRequiredRegex.MatchString(body.Message) {
			return true
		}
	}
	return false
}

func getAuthProxy(d *schema.ResourceData) map[string]string {
	if d.Get("authenticating_proxy.#").(int) == 0 {
		return nil
	}
	resMap := map[string]string{
		"ca":         common.Base64IfNot(d.Get("authenticating_proxy.0.ca").(string)),
		"cert":       common.Base64IfNot(d.Get("authenticating_proxy.0.cert").(string)),
		"privateKey": common.Base64IfNot(d.Get("authenticating_proxy.0.private_key").(string)),
	}
	return resMap
}

func validateAuthProxy(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if d.Get("authentication_mode") != "authenticating_proxy" {
		return nil
	}
	if d.Get("authenticating_proxy.#").(int) == 0 {
		return fmt.Errorf("`authenticating_proxy` fields needs to be set if auth mode is `authenticating_proxy`")
	}
	return nil
}
