package dcaas

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	virtual_interface "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/virtual-interface"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVirtualInterfacePeerV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualInterfacePeerV3Create,
		ReadContext:   resourceVirtualInterfacePeerV3Read,
		UpdateContext: resourceVirtualInterfacePeerV3Update,
		DeleteContext: resourceVirtualInterfacePeerV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualInterfacePeerV3ImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"address_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"local_gateway_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_gateway_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"bgp_md5": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"remote_ep_group": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vif_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
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
			"receive_route_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVirtualInterfacePeerV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	opts := virtual_interface.CreatePeerOpts{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		LocalGatewayIP:  d.Get("local_gateway_ip").(string),
		RemoteGatewayIP: d.Get("remote_gateway_ip").(string),
		AddressFamily:   d.Get("address_family").(string),
		RouteMode:       d.Get("route_mode").(string),
		BGPASN:          d.Get("bgp_asn").(int),
		BGPMD5:          d.Get("bgp_md5").(string),
		RemoteEpGroup:   common.ExpandToStringList(d.Get("remote_ep_group").([]interface{})),
		VifId:           d.Get("vif_id").(string),
	}
	peer, err := virtual_interface.CreatePeer(client, opts)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud DC virtual interface peer v3: %s", err)
	}
	d.SetId(peer.ID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceVirtualInterfacePeerV3Read(clientCtx, d, meta)
}

func resourceVirtualInterfacePeerV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}
	vi, err := virtual_interface.Get(client, d.Get("vif_id").(string))
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving OpenTelekomCloud DC virtual interface v3")
	}
	log.Printf("[DEBUG] The response of OpenTelekomCloud DC virtual interface v3 is: %#v", vi)

	peer := getVifPeer(vi.VifPeers, d)
	mErr := multierror.Append(nil,
		d.Set("name", peer.Name),
		d.Set("description", peer.Description),
		d.Set("address_family", peer.AddressFamily),
		d.Set("local_gateway_ip", peer.LocalGatewayIp),
		d.Set("remote_gateway_ip", peer.RemoteGatewayIp),
		d.Set("route_mode", peer.RouteMode),
		d.Set("bgp_asn", peer.BgpAsn),
		d.Set("bgp_md5", peer.BgpMd5),
		d.Set("device_id", peer.DeviceId),
		d.Set("enable_bfd", peer.EnableBfd),
		d.Set("enable_nqa", peer.EnableNqa),
		d.Set("bgp_route_limit", peer.BgpRouteLimit),
		d.Set("bgp_status", peer.BgpStatus),
		d.Set("status", peer.Status),
		d.Set("vif_id", peer.VifId),
		d.Set("receive_route_num", peer.ReceiveRouteNum),
		d.Set("remote_ep_group", peer.RemoteEpGroup),
		d.Set("service_ep_group", peer.ServiceEpGroup),
		d.Set("project_id", peer.TenantId),
		d.Set("region", config.GetRegion(d)),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error saving OpenTelekomCloud DC virtual interface peer v3 fields: %s", err)
	}
	return nil
}

func getVifPeer(vifPeers []virtual_interface.VifPeer, d *schema.ResourceData) virtual_interface.VifPeer {
	var vifPeer virtual_interface.VifPeer
	for _, v := range vifPeers {
		if v.ID == d.Id() {
			return v
		}
	}
	return vifPeer
}

func resourceVirtualInterfacePeerV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	if d.HasChanges("name", "description", "remote_ep_group") {
		opts := virtual_interface.UpdatePeerOpts{
			Name:          d.Get("name").(string),
			Description:   d.Get("description").(string),
			RemoteEpGroup: common.ExpandToStringList(d.Get("remote_ep_group").([]interface{})),
		}

		_, err := virtual_interface.UpdatePeer(client, d.Id(), opts)
		if err != nil {
			return diag.Errorf("error updating of the OpenTelekomCloud DC virtual interface peer v3 (%s): %s", d.Id(), err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceVirtualInterfacePeerV3Read(clientCtx, d, meta)
}

func resourceVirtualInterfacePeerV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	err = virtual_interface.DeletePeer(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error deleting OpenTelekomCloud DC virtual interface peer v3")
	}

	return nil
}

func resourceVirtualInterfacePeerV3ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*cfg.Config)
	client, err := config.DCaaSV3Client(config.GetRegion(d))
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud DCaaS v3 client: %s", err)
	}
	viList, err := virtual_interface.List(client, virtual_interface.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error getting OpenTelekomCloud DCaaS v3 virtual interface: %s", err)
	}
	var peer *virtual_interface.VifPeer
	for _, v := range viList {
		for _, p := range v.VifPeers {
			if p.ID == d.Id() {
				peer = &p
			}
		}
	}
	if peer == nil {
		return nil, fmt.Errorf("no resource found for provided import ID")
	}
	mErr := multierror.Append(nil,
		d.Set("name", peer.Name),
		d.Set("description", peer.Description),
		d.Set("address_family", peer.AddressFamily),
		d.Set("local_gateway_ip", peer.LocalGatewayIp),
		d.Set("remote_gateway_ip", peer.RemoteGatewayIp),
		d.Set("route_mode", peer.RouteMode),
		d.Set("bgp_asn", peer.BgpAsn),
		d.Set("bgp_md5", peer.BgpMd5),
		d.Set("device_id", peer.DeviceId),
		d.Set("enable_bfd", peer.EnableBfd),
		d.Set("enable_nqa", peer.EnableNqa),
		d.Set("bgp_route_limit", peer.BgpRouteLimit),
		d.Set("bgp_status", peer.BgpStatus),
		d.Set("status", peer.Status),
		d.Set("vif_id", peer.VifId),
		d.Set("receive_route_num", peer.ReceiveRouteNum),
		d.Set("remote_ep_group", peer.RemoteEpGroup),
		d.Set("service_ep_group", peer.ServiceEpGroup),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return []*schema.ResourceData{d}, fmt.Errorf("error saving OpenTelekomCloud DCaaS v3 virtual interface peer resource fields during import: %s", err)
	}
	return []*schema.ResourceData{d}, nil
}
