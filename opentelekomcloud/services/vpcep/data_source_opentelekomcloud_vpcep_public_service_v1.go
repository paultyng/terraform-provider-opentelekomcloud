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

func DataSourceVPCEPPublicServiceV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCEPPublicServiceV1Read,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_charge": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEPPublicServiceV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	client, err := config.VpcEpV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	opts := services.ListPublicOpts{
		Name: d.Get("name").(string),
		ID:   d.Get("id").(string),
	}

	svcs, err := services.ListPublic(client, opts)
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
		d.Set("name", svc.ServiceName),
		d.Set("service_type", svc.ServiceType),
		d.Set("owner", svc.Owner),
		d.Set("is_charge", svc.IsCharge),
		d.Set("created_at", svc.CreatedAt),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting public service fields: %w", err)
	}

	return nil
}
