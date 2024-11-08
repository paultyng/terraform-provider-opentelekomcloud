package ddm

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	ddmv1instances "github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/instances"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceDdmInstanceV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDdmInstanceV1Read,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceDdmInstanceV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	instance, err := ddmv1instances.QueryInstanceDetails(client, d.Get("instance_id").(string))
	if err != nil {
		return fmterr.Errorf("error fetching DDM instance: %w", err)
	}

	d.SetId(instance.Id)
	log.Printf("[DEBUG] Retrieved instance %s: %#v", d.Id(), instance)

	mErr := multierror.Append(nil,
		d.Set("instance_id", d.Id()),
		d.Set("region", config.GetRegion(d)),
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
