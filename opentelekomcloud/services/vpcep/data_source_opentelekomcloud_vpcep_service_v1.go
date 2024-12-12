package vpcep

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vpcep/v1/services"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceVPCEPServiceV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCEPServiceV1Read,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vip_port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"approval_enabled": {
				Type:     schema.TypeBool,
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
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"server_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": common.TagsSchema(),
			"connection_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tcp_proxy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEPServiceV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	client, err := config.VpcEpV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	opts := services.ListOpts{
		Name:   d.Get("name").(string),
		ID:     d.Get("id").(string),
		Status: services.Status(d.Get("status").(string)),
	}

	svcs, err := services.List(client, opts)
	if err != nil {
		return fmterr.Errorf("error listing VPCEP public services: %w", err)
	}

	if len(svcs) > 1 {
		return common.DataSourceTooManyDiag
	}
	if len(svcs) < 1 {
		return common.DataSourceTooFewDiag
	}

	svc := svcs[0]

	d.SetId(svc.ID)
	mErr := multierror.Append(
		d.Set("name", onlyServiceName(svc.ServiceName)),
		d.Set("port_id", svc.PortID),
		d.Set("vip_port_id", svc.VIPPortID),
		d.Set("server_type", svc.ServerType),
		d.Set("vpc_id", svc.VpcID),
		d.Set("approval_enabled", svc.ApprovalEnabled),
		d.Set("service_type", svc.ServiceType),
		d.Set("created_at", svc.CreatedAt),
		d.Set("updated_at", svc.UpdatedAt),
		d.Set("project_id", svc.ProjectID),
		d.Set("port", portsSlice(svc.Ports)),
		d.Set("tags", common.TagsToMap(svc.Tags)),
		d.Set("connection_count", svc.ConnectionCount),
		d.Set("tcp_proxy", svc.TCPProxy),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting public service fields: %w", err)
	}

	return nil
}
