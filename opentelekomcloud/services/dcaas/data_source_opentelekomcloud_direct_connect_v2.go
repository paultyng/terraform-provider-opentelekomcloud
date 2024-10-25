package dcaas

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dcaas "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v2/direct-connect"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceDirectConnectV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDirectConnectV2Read,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"bandwidth": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"port_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"interface_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"redundant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosting_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"charge_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"order_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"apply_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"applicant": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mobile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cable_label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_port_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"onestop_product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"building_line_product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_onestop_product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"period_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"period_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vgw_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDirectConnectV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DCaaSV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	var ID string
	if v, ok := d.GetOk("id"); ok {
		ID = v.(string)
	}

	directConnect, err := dcaas.Get(client, ID)
	if err != nil {
		return fmterr.Errorf("error reading direct connect: %s", err)
	}
	log.Printf("[DEBUG] Direct Connect read result: %#v", directConnect)

	d.SetId(directConnect.ID)

	mErr := multierror.Append(err,
		d.Set("tenant_id", directConnect.TenantID),
		d.Set("name", directConnect.Name),
		d.Set("description", directConnect.Description),
		d.Set("port_type", directConnect.PortType),
		d.Set("bandwidth", directConnect.Bandwidth),
		d.Set("location", directConnect.Location),
		d.Set("peer_location", directConnect.PeerLocation),
		d.Set("device_id", directConnect.DeviceID),
		d.Set("interface_name", directConnect.InterfaceName),
		d.Set("redundant_id", directConnect.RedundantID),
		d.Set("provider_name", directConnect.Provider),
		d.Set("provider_status", directConnect.ProviderStatus),
		d.Set("type", directConnect.Type),
		d.Set("hosting_id", directConnect.HostingID),
		d.Set("vlan", directConnect.VLAN),
		d.Set("charge_mode", directConnect.ChargeMode),
		d.Set("apply_time", directConnect.ApplyTime),
		d.Set("create_time", directConnect.CreateTime),
		d.Set("delete_time", directConnect.DeleteTime),
		d.Set("order_id", directConnect.OrderID),
		d.Set("product_id", directConnect.ProductID),
		d.Set("status", directConnect.Status),
		d.Set("admin_state_up", directConnect.AdminStateUp),
		d.Set("spec_code", directConnect.SpecCode),
		d.Set("applicant", directConnect.Applicant),
		d.Set("mobile", directConnect.Mobile),
		d.Set("email", directConnect.Email),
		d.Set("region_id", directConnect.RegionID),
		d.Set("service_key", directConnect.ServiceKey),
		d.Set("cable_label", directConnect.CableLabel),
		d.Set("peer_port_type", directConnect.PeerPortType),
		d.Set("peer_provider", directConnect.PeerProvider),
		d.Set("onestop_product_id", directConnect.OnestopProductID),
		d.Set("building_line_product_id", directConnect.BuildingLineProductID),
		d.Set("last_onestop_product_id", directConnect.LastOnestopProductID),
		d.Set("period_type", directConnect.PeriodType),
		d.Set("period_num", directConnect.PeriodNum),
		d.Set("reason", directConnect.Reason),
		d.Set("vgw_type", directConnect.VGWType),
		d.Set("lag_id", directConnect.LagID),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
