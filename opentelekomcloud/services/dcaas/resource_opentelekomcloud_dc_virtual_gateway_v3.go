package dcaas

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	gateway "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/virtual-gateway"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVirtualGatewayV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualGatewayV3Create,
		ReadContext:   resourceVirtualGatewayV3Read,
		UpdateContext: resourceVirtualGatewayV3Update,
		DeleteContext: resourceVirtualGatewayV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"local_ep_group": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"local_ep_group_ipv6": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asn": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVirtualGatewayV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}
	opts := gateway.CreateOpts{
		VpcId:            d.Get("vpc_id").(string),
		LocalEpGroup:     common.ExpandToStringList(d.Get("local_ep_group").([]interface{})),
		LocalEpGroupIpv6: common.ExpandToStringList(d.Get("local_ep_group_ipv6").([]interface{})),
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		BgpAsn:           d.Get("asn").(int),
	}

	gw, err := gateway.Create(client, opts)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud DC virtual gateway v3: %s", err)
	}
	d.SetId(gw.ID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceVirtualGatewayV3Read(clientCtx, d, meta)
}

func resourceVirtualGatewayV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	gw, err := gateway.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving OpenTelekomCloud DC virtual gateway v3")
	}
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("vpc_id", gw.VpcId),
		d.Set("local_ep_group", gw.LocalEpGroup),
		d.Set("local_ep_group_ipv6", gw.LocalEpGroupIpv6),
		d.Set("name", gw.Name),
		d.Set("description", gw.Description),
		d.Set("asn", gw.BgpAsn),
		d.Set("status", gw.Status),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error saving OpenTelekomCloud DC virtual gateway v3 fields: %s", err)
	}
	return nil
}

func resourceVirtualGatewayV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	updateVirtualGatewayChanges := []string{
		"name",
		"description",
		"local_ep_group",
		"local_ep_group_ipv6",
	}

	if d.HasChanges(updateVirtualGatewayChanges...) {
		opts := gateway.UpdateOpts{
			Name:             d.Get("name").(string),
			Description:      pointerto.String(d.Get("description").(string)),
			LocalEpGroup:     common.ExpandToStringList(d.Get("local_ep_group").([]interface{})),
			LocalEpGroupIpv6: common.ExpandToStringList(d.Get("local_ep_group_ipv6").([]interface{})),
		}
		_, err = gateway.Update(client, d.Id(), opts)
		if err != nil {
			return diag.Errorf("error updating OpenTelekomCloud DC virtual gateway v3 (%s): %s", d.Id(), err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceVirtualGatewayV3Read(clientCtx, d, meta)
}

func resourceVirtualGatewayV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	err = gateway.Delete(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error deleting OpenTelekomCloud DC virtual gateway v3")
	}

	return nil
}
