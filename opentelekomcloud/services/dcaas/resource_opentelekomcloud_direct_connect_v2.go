package dcaas

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	dcaas "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v2/direct-connect"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDirectConnectV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDirectConnectV2Create,
		ReadContext:   resourceDirectConnectV2Read,
		DeleteContext: resourceDirectConnectV2Delete,
		UpdateContext: resourceDirectConnectV2Update,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
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
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
				Optional: true,
				ForceNew: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"interface_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"redundant_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provider_status": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"hosting_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"charge_mode": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"order_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

func resourceDirectConnectV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	createOpts := dcaas.CreateOpts{
		Bandwidth:      d.Get("bandwidth").(int),
		PortType:       d.Get("port_type").(string),
		Location:       d.Get("location").(string),
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		PeerLocation:   d.Get("peer_location").(string),
		DeviceID:       d.Get("device_id").(string),
		InterfaceName:  d.Get("interface_name").(string),
		RedundantID:    d.Get("redundant_id").(string),
		Provider:       d.Get("provider_name").(string),
		ProviderStatus: d.Get("provider_status").(string),
		Type:           d.Get("type").(string),
		HostingID:      d.Get("hosting_id").(string),
		ChargeMode:     d.Get("charge_mode").(string),
		OrderID:        d.Get("order_id").(string),
		ProductID:      d.Get("product_id").(string),
		Status:         d.Get("status").(string),
		AdminStateUp:   d.Get("admin_state_up").(bool),
	}
	log.Printf("[DEBUG] Direct Connect create options: %#v", createOpts)

	directConnect, err := dcaas.Create(client, createOpts)
	if err != nil {
		return fmterr.Errorf("error creating direct connect: %s", err)
	}
	d.SetId(directConnect.ID)
	log.Printf("[INFO] Direct Connect created: %s", directConnect.ID)
	return resourceDirectConnectV2Read(ctx, d, meta)
}

func resourceDirectConnectV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	if d.HasChange("name") {
		opt := dcaas.UpdateOpts{
			Name: d.Get("name").(string),
		}
		err = dcaas.Update(client, d.Id(), opt)
		if err != nil {
			return fmterr.Errorf("error updating direct connect: %s", err)
		}
	}

	if d.HasChange("bandwidth") {
		opt := dcaas.UpdateOpts{
			Bandwidth: d.Get("bandwidth").(int),
		}
		err = dcaas.Update(client, d.Id(), opt)
		log.Printf("[INFO] Direct Connect updated: %s", d.Id())
		if err != nil {
			return fmterr.Errorf("error updating direct connect: %s", err)
		}
	}

	if d.HasChange("description") {
		opt := dcaas.UpdateOpts{
			Description: d.Get("description").(string),
		}
		err = dcaas.Update(client, d.Id(), opt)
		if err != nil {
			return fmterr.Errorf("error updating direct connect: %s", err)
		}
	}

	if d.HasChange("provider_status") {
		opt := dcaas.UpdateOpts{
			ProviderStatus: d.Get("provider_status").(string),
		}
		err = dcaas.Update(client, d.Id(), opt)
		if err != nil {
			return fmterr.Errorf("error updating direct connect: %s", err)
		}
	}

	return resourceDirectConnectV2Read(ctx, d, meta)
}

func resourceDirectConnectV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	directConnect, err := dcaas.Get(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error reading direct connect: %s", err)
	}
	log.Printf("[DEBUG] Direct Connect read result: %#v", directConnect)

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

func resourceDirectConnectV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	err = dcaas.Delete(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error deleting direct connect: %s", err)
	}
	log.Printf("[INFO] Direct Connect deleted: %s", d.Id())
	d.SetId("")
	return nil
}
