package css

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/clusters"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/flavors"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceCssClusterV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCssClusterV1Create,
		ReadContext:   resourceCssClusterV1Read,
		UpdateContext: resourceCssClusterV1Update,
		DeleteContext: resourceCssClusterV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterV2ImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: checkCssClusterFlavorRestrictions,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"node_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flavor": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"network_info": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_id": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"security_group_id": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"vpc_id": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"volume": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"volume_type": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"encryption_key": {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "",
						},
					},
				},
			},
			"enable_https": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"enable_authority": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"admin_pass": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enable_authority"},
				ForceNew:     true,
			},
			"expect_node_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"datastore": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "elasticsearch",
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "7.6.2",
						},
					},
				},
			},
			"tags": {
				Type:         schema.TypeMap,
				Optional:     true,
				Computed:     true,
				ValidateFunc: common.ValidateTags,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCssClusterV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating CSS v1 client: %s", err)
	}

	opts := clusters.CreateOpts{
		Instance: &clusters.InstanceSpec{
			Flavor: d.Get("node_config.0.flavor").(string),
			Volume: &clusters.Volume{
				Type: d.Get("node_config.0.volume.0.volume_type").(string),
				Size: d.Get("node_config.0.volume.0.size").(int),
			},
			Nics: &clusters.Nics{
				VpcID:           d.Get("node_config.0.network_info.0.vpc_id").(string),
				SubnetID:        d.Get("node_config.0.network_info.0.network_id").(string),
				SecurityGroupID: d.Get("node_config.0.network_info.0.security_group_id").(string),
			},
			AvailabilityZone: d.Get("node_config.0.availability_zone").(string),
		},
		Name:        d.Get("name").(string),
		InstanceNum: d.Get("expect_node_num").(int),
		DiskEncryption: &clusters.DiskEncryption{
			Encrypted: "0",
		},
		AuthorityEnabled: d.Get("enable_authority").(bool),
		AdminPassword:    d.Get("admin_pass").(string),
		Tags:             common.ExpandResourceTags(d.Get("tags").(map[string]interface{})),
	}
	if enable, ok := d.GetOk("enable_https"); ok {
		opts.HttpsEnabled = fmt.Sprint(enable.(bool))
	}
	if cmkID, ok := d.GetOk("node_config.0.volume.0.encryption_key"); ok {
		opts.DiskEncryption = &clusters.DiskEncryption{
			Encrypted: "1",
			CmkID:     cmkID.(string),
		}
	}
	if count := d.Get("datastore.#").(int); count != 0 {
		opts.Datastore = &clusters.Datastore{
			Version: d.Get("datastore.0.version").(string),
			Type:    d.Get("datastore.0.type").(string),
		}
	} else {
		opts.Datastore = &defaultDatastore
	}

	created, err := clusters.Create(client, opts)
	if err != nil {
		return fmterr.Errorf("error creating CSS cluster: %s", err)
	}

	secondsWait := int(math.Round(d.Timeout(schema.TimeoutCreate).Seconds()))
	err = clusters.WaitForClusterOperationSucces(client, created.ID, secondsWait)
	if err != nil {
		return fmterr.Errorf("error waiting for CSS cluster to be running: %s", err)
	}

	d.SetId(created.ID)

	return resourceCssClusterV1Read(ctx, d, meta)
}

func resourceCssClusterV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating CSS v1 client: %s", err)
	}

	cluster, err := clusters.Get(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error reading cluster value: %s", err)
	}

	mErr := multierror.Append(
		d.Set("name", cluster.Name),
		d.Set("enable_https", cluster.HttpsEnabled),
		d.Set("enable_authority", cluster.AuthorityEnabled),
		d.Set("created", cluster.Created),
		d.Set("updated", cluster.Updated),
		d.Set("endpoint", cluster.Endpoint),
		d.Set("nodes", extractNodes(cluster)),
		d.Set("datastore", extractDatastore(cluster)),
		d.Set("tags", common.TagsToMap(cluster.Tags)),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func extractNodes(c *clusters.Cluster) []interface{} {
	nodes := make([]interface{}, len(c.Instances))
	for i, node := range c.Instances {
		nodes[i] = map[string]interface{}{
			"id":   node.ID,
			"name": node.Name,
			"type": node.Type,
		}
	}
	return nodes
}

func extractDatastore(c *clusters.Cluster) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":    c.Datastore.Type,
			"version": c.Datastore.Version,
		},
	}
}

func resourceCssClusterV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating CSS v1 client: %s", err)
	}

	if d.HasChange("tags") {
		if err := common.UpdateResourceTags(client, d, "css-cluster", d.Id()); err != nil {
			return fmterr.Errorf("error updating tags of CSS cluster %s: %s", d.Id(), err)
		}
	}

	if !d.HasChange("expect_node_num") && !d.HasChange("node_config.0.volume.0.size") {
		return nil
	}

	oldNode, newNode := d.GetChange("expect_node_num")
	nodeDiff := newNode.(int) - oldNode.(int)
	if nodeDiff < 0 {
		return fmterr.Errorf("invalid number of new nodes: %d", nodeDiff)
	}

	oldSize, newSize := d.GetChange("node_config.0.volume.0.size")
	sizeDiff := newSize.(int) - oldSize.(int)

	switch {
	case sizeDiff < 0:
		return fmterr.Errorf("invalid number of new volume size: %d", sizeDiff)
	case sizeDiff == 0:
		_, err = clusters.ExtendCluster(client, d.Id(), clusters.ClusterExtendCommonOpts{
			ModifySize: nodeDiff,
		})
	default:
		_, err = clusters.ExtendCluster(client, d.Id(), []clusters.ClusterExtendSpecialOpts{
			{
				Type:     "ess",
				NodeSize: nodeDiff,
				DiskSize: sizeDiff,
			},
		})
	}

	if err != nil {
		return fmterr.Errorf("error extending cluster: %s", err)
	}

	secondsWait := int(math.Round(d.Timeout(schema.TimeoutUpdate).Seconds()))
	if err := clusters.WaitForClusterToExtend(client, d.Id(), secondsWait); err != nil {
		state, _ := clusters.Get(client, d.Id())
		if state != nil {
			return fmterr.Errorf("error waiting cluster to extend: %s\nFail reason: %+v", err, state.FailedReasons)
		}
		return fmterr.Errorf("error waiting cluster to extend: %s", err)
	}

	return resourceCssClusterV1Read(ctx, d, meta)
}

func resourceCssClusterV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating CSS v1 client: %s", err)
	}

	if err := clusters.Delete(client, d.Id()); err != nil {
		return fmterr.Errorf("error deleting cluster: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Target:     []string{},
		Pending:    []string{clusterStateAvailable},
		Refresh:    resourceCssClusterV1StateRefresh(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for cluster to be deleted: %s", err)
	}
	return nil
}

const (
	clusterStateAvailable = "AVAILABLE"
)

func resourceCssClusterV1StateRefresh(client *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (result interface{}, state string, err error) {
		cluster, err := clusters.Get(client, id)
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return nil, "", nil
			}
			return nil, "", err
		}
		return cluster, clusterStateAvailable, nil
	}
}

func checkCssClusterFlavorRestrictions(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating CSS v1 client: %s", err)
	}

	flavorName := d.Get("node_config.0.flavor").(string)
	size := d.Get("node_config.0.volume.0.size").(int)

	versions, err := flavors.List(client)
	if err != nil {
		return fmt.Errorf("error extracting flavor list: %s", err)
	}
	flavor := flavors.FindFlavor(versions, flavors.FilterOpts{
		FlavorName: flavorName,
	})
	if flavor == nil {
		return fmt.Errorf("can't find flavor with name: %s", flavorName)
	}

	if size < flavor.DiskMin || size > flavor.DiskMax {
		return fmt.Errorf("invalid disk size, `%s` support disk from %dGB to %dGB",
			flavorName, flavor.DiskMin, flavor.DiskMax)
	}

	return nil
}

func resourceClusterV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return nil, fmt.Errorf("error creating CSS v1 client: %s", err)
	}

	if diagRead := resourceCssClusterV1Read(ctx, d, meta); diagRead.HasError() {
		return nil, fmt.Errorf("error reading opentelekomcloud_css_cluster_v1 %s: %s", d.Id(), diagRead[0].Summary)
	}

	cluster, err := clusters.Get(client, d.Id())
	if err != nil {
		return nil, common.CheckDeleted(d, err, "Cluster")
	}

	var nodes []map[string]interface{}
	for _, instance := range cluster.Instances {
		volume := map[string]interface{}{}
		volume["volume_type"] = instance.Volume.Type
		volume["size"] = instance.Volume.Size
		volume["encryption_key"] = cluster.CmkID
		volumeList := []interface{}{volume}

		network := map[string]interface{}{}
		network["network_id"] = cluster.SubnetID
		network["security_group_id"] = cluster.SecurityGroupID
		network["vpc_id"] = cluster.VpcID
		networkList := []interface{}{network}

		mapping := map[string]interface{}{
			"availability_zone": instance.AvailabilityZone,
			"flavor":            instance.SpecCode,
			"volume":            volumeList,
			"network_info":      networkList,
		}
		nodes = append(nodes, mapping)
	}

	mErr := multierror.Append(nil,
		d.Set("name", cluster.Name),
		d.Set("enable_https", cluster.HttpsEnabled),
		d.Set("enable_authority", cluster.AuthorityEnabled),
		d.Set("expect_node_num", len(cluster.Instances)),
		d.Set("node_config", nodes),
		d.Set("datastore", extractDatastore(cluster)),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return nil, fmt.Errorf("error setting addon attributes: %w", err)
	}

	log.Printf("[DEBUG] Retrieved CSS cluster %s during the import: %#v", d.Id(), cluster)

	return []*schema.ResourceData{d}, nil
}
