package dcaas

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	virtual_interface "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/virtual-interface"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVirtualInterfaceV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualInterfaceV3Create,
		ReadContext:   resourceVirtualInterfaceV3Read,
		UpdateContext: resourceVirtualInterfaceV3Update,
		DeleteContext: resourceVirtualInterfaceV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"direct_connect_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vgw_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"bandwidth": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"remote_ep_group": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_ep_group": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"local_gateway_v4_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"remote_gateway_v4_ip"},
			},
			"remote_gateway_v4_ip": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"remote_gateway_v6_ip"},
			},
			"address_family": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"local_gateway_v6_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"remote_gateway_v6_ip"},
				ExactlyOneOf: []string{"local_gateway_v4_ip"},
			},
			"remote_gateway_v6_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"asn": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"bgp_md5": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enable_bfd": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"enable_nqa": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"resource_tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
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
			"vif_peers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tenant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"local_gateway_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"remote_gateway_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"route_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"bgp_asn": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"bgp_md5": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"remote_ep_group": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"service_ep_group": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"device_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_bfd": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enable_nqa": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"bgp_route_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"bgp_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vif_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"receive_route_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVirtualInterfaceV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	opts := virtual_interface.CreateOpts{
		DirectConnectID:   d.Get("direct_connect_id").(string),
		VgwId:             d.Get("vgw_id").(string),
		Type:              d.Get("type").(string),
		RouteMode:         d.Get("route_mode").(string),
		VLAN:              d.Get("vlan").(int),
		Bandwidth:         d.Get("bandwidth").(int),
		RemoteEpGroup:     common.ExpandToStringList(d.Get("remote_ep_group").([]interface{})),
		ServiceEpGroup:    common.ExpandToStringList(d.Get("service_ep_group").([]interface{})),
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		LocalGatewayV4IP:  d.Get("local_gateway_v4_ip").(string),
		RemoteGatewayV4IP: d.Get("remote_gateway_v4_ip").(string),
		AddressFamily:     d.Get("address_family").(string),
		LocalGatewayV6IP:  d.Get("local_gateway_v6_ip").(string),
		RemoteGatewayV6IP: d.Get("remote_gateway_v6_ip").(string),
		BGPASN:            d.Get("asn").(int),
		BGPMD5:            d.Get("bgp_md5").(string),
		EnableBfd:         d.Get("enable_bfd").(bool),
		EnableNqa:         d.Get("enable_nqa").(bool),
		LagId:             d.Get("lag_id").(string),
		ResourceTenantID:  d.Get("resource_tenant_id").(string),
	}
	vi, err := virtual_interface.Create(client, opts)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud DC virtual interface v3: %s", err)
	}
	d.SetId(vi.ID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceVirtualInterfaceV3Read(clientCtx, d, meta)
}

func resourceVirtualInterfaceV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	vi, err := virtual_interface.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving OpenTelekomCloud DC virtual interface v3")
	}
	log.Printf("[DEBUG] The response of OpenTelekomCloud DC virtual interface v3 is: %#v", vi)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("vgw_id", vi.VgwId),
		d.Set("type", vi.Type),
		d.Set("route_mode", vi.RouteMode),
		d.Set("vlan", vi.Vlan),
		d.Set("bandwidth", vi.Bandwidth),
		d.Set("remote_ep_group", vi.RemoteEpGroup),
		d.Set("service_ep_group", vi.ServiceEpGroup),
		d.Set("name", vi.Name),
		d.Set("description", vi.Description),
		d.Set("direct_connect_id", vi.DirectConnectId),
		d.Set("local_gateway_v4_ip", vi.LocalGatewayV4Ip),
		d.Set("remote_gateway_v4_ip", vi.RemoteGatewayV4Ip),
		d.Set("address_family", vi.AddressFamily),
		d.Set("local_gateway_v6_ip", vi.LocalGatewayV6Ip),
		d.Set("remote_gateway_v6_ip", vi.RemoteGatewayV6Ip),
		d.Set("asn", vi.BgpAsn),
		d.Set("bgp_md5", vi.BgpMd5),
		d.Set("enable_bfd", vi.EnableBfd),
		d.Set("enable_nqa", vi.EnableNqa),
		d.Set("lag_id", vi.LagId),
		d.Set("device_id", vi.DeviceId),
		d.Set("status", vi.Status),
		d.Set("created_at", vi.CreatedAt),
		d.Set("updated_at", vi.UpdatedAt),
		d.Set("vif_peers", flattenVifPeers(vi.VifPeers)),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error saving OpenTelekomCloud DC virtual interface v3 fields: %s", err)
	}
	return nil
}

func flattenVifPeers(vifPeers []virtual_interface.VifPeer) []interface{} {
	if vifPeers == nil {
		return nil
	}

	rst := make([]interface{}, 0, len(vifPeers))
	for _, v := range vifPeers {
		rst = append(rst, map[string]interface{}{
			"id":                v.ID,
			"name":              v.Name,
			"description":       v.Description,
			"address_family":    v.AddressFamily,
			"local_gateway_ip":  v.LocalGatewayIp,
			"remote_gateway_ip": v.RemoteGatewayIp,
			"route_mode":        v.RouteMode,
			"bgp_asn":           v.BgpAsn,
			"bgp_md5":           v.BgpMd5,
			"device_id":         v.DeviceId,
			"enable_bfd":        v.EnableBfd,
			"enable_nqa":        v.EnableNqa,
			"bgp_route_limit":   v.BgpRouteLimit,
			"bgp_status":        v.BgpStatus,
			"status":            v.Status,
			"vif_id":            v.VifId,
			"receive_route_num": v.ReceiveRouteNum,
			"remote_ep_group":   v.RemoteEpGroup,
			"service_ep_group":  v.ServiceEpGroup,
		})
	}
	return rst
}

func closeVirtualInterfaceNetworkDetection(client *golangsdk.ServiceClient, d *schema.ResourceData) error {
	opts := virtual_interface.UpdateOpts{}

	// At the same time, only one of BFD and NQA is enabled.
	if d.HasChange("enable_bfd") && !d.Get("enable_bfd").(bool) {
		opts.EnableBfd = pointerto.Bool(false)
	} else if d.HasChange("enable_nqa") && !d.Get("enable_nqa").(bool) {
		opts.EnableNqa = pointerto.Bool(false)
	}
	if reflect.DeepEqual(opts, virtual_interface.UpdateOpts{}) {
		return nil
	}

	_, err := virtual_interface.Update(client, d.Id(), opts)
	if err != nil {
		return fmt.Errorf("error closing network detection of the OpenTelekomCloud DC virtual interface v3 (%s): %s", d.Id(), err)
	}
	return nil
}

func openVirtualInterfaceNetworkDetection(client *golangsdk.ServiceClient, d *schema.ResourceData) error {
	opts := virtual_interface.UpdateOpts{}
	detectionOpened := false

	if d.HasChange("enable_bfd") && d.Get("enable_bfd").(bool) {
		detectionOpened = true
		opts.EnableBfd = pointerto.Bool(true)
	}
	if d.HasChange("enable_nqa") && d.Get("enable_nqa").(bool) {
		// The enable requests of BFD and NQA cannot be sent at the same time.
		if detectionOpened {
			return fmt.Errorf("BFD and NQA cannot be enabled at the same time")
		}
		opts.EnableNqa = pointerto.Bool(true)
	}
	if reflect.DeepEqual(opts, virtual_interface.UpdateOpts{}) {
		return nil
	}

	_, err := virtual_interface.Update(client, d.Id(), opts)
	if err != nil {
		return fmt.Errorf("error opening network detection of the OpenTelekomCloud DC virtual interface v3 (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceVirtualInterfaceV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	if d.HasChanges("name", "description", "bandwidth", "remote_ep_group") {
		opts := virtual_interface.UpdateOpts{
			Name:           d.Get("name").(string),
			Description:    pointerto.String(d.Get("description").(string)),
			Bandwidth:      d.Get("bandwidth").(int),
			RemoteEpGroup:  common.ExpandToStringList(d.Get("remote_ep_group").([]interface{})),
			ServiceEpGroup: common.ExpandToStringList(d.Get("service_ep_group").([]interface{})),
		}

		_, err := virtual_interface.Update(client, d.Id(), opts)
		if err != nil {
			return diag.Errorf("error closing network detection of the OpenTelekomCloud DC virtual interface v3 (%s): %s", d.Id(), err)
		}
	}
	if d.HasChanges("enable_bfd", "enable_nqa") {
		// BFD and NQA cannot be enabled at the same time.
		// When BFD (NQA) is enabled and NQA (BFD) is disabled,
		// we need to disable BFD (NQA) first, and then enable NQA (BFD).
		// If to disable and enable requests are sent at the same time, an error will be reported.
		if err = closeVirtualInterfaceNetworkDetection(client, d); err != nil {
			return diag.FromErr(err)
		}
		if err = openVirtualInterfaceNetworkDetection(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceVirtualInterfaceV3Read(clientCtx, d, meta)
}

func resourceVirtualInterfaceV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	err = virtual_interface.Delete(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error deleting OpenTelekomCloud DC virtual interface v3")
	}

	return nil
}
