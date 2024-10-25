package dcaas

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	dceg "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v2/dc-endpoint-group"
	virtual_gateway "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v2/virtual-gateway"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVirtualGatewayV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualGatewayV2Create,
		ReadContext:   resourceVirtualGatewayV2Read,
		UpdateContext: resourceVirtualGatewayV2Update,
		DeleteContext: resourceVirtualGatewayV2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"local_ep_group":    localGroupSchema(),
			"local_ep_group_v6": localGroupSchema(),
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[^<>]*$`),
						"The angle brackets (< and >) are not allowed."),
					validation.StringLenBetween(0, 128),
				),
			},
			"asn": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"redundant_device_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_ep_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_ep_group_ipv6_id": {
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

func localGroupSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		AtLeastOneOf: []string{"local_ep_group_v6", "local_ep_group"},
		Optional:     true,
		Computed:     true,
		MaxItems:     1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"description": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"endpoints": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  "cidr",
				},
			},
		},
	}
}

func createGroup(client *golangsdk.ServiceClient, d *schema.ResourceData, group string) (*dceg.DCEndpointGroup, error) {
	projectId := d.Get("project_id").(string)
	if projectId == "" {
		projectId = client.ProjectID
	}
	egRaw := d.Get(group).([]interface{})
	if len(egRaw) == 1 {
		rawMap := egRaw[0].(map[string]interface{})
		createOpts := dceg.CreateOpts{
			Name:        rawMap["name"].(string),
			TenantId:    projectId,
			Description: rawMap["description"].(string),
			Endpoints:   GetEndpoints(rawMap["endpoints"].([]interface{})),
			Type:        rawMap["type"].(string),
		}
		egResp, err := dceg.Create(client, createOpts)
		if err != nil {
			return nil, err
		}
		return egResp, nil
	}
	return nil, nil
}

func resourceLocalEgCreate(client *golangsdk.ServiceClient, d *schema.ResourceData) (eg, egV6 *dceg.DCEndpointGroup, err error) {
	egRaw, err := createGroup(client, d, "local_ep_group")
	if err != nil {
		return nil, nil, err
	}

	egV6Raw, err := createGroup(client, d, "local_ep_group_v6")
	if err != nil {
		return nil, nil, err
	}

	eg, egV6 = egRaw, egV6Raw

	return eg, egV6, err
}

func resourceVirtualGatewayV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	eg, egV6, err := resourceLocalEgCreate(client, d)
	if err != nil {
		return fmterr.Errorf("error creating DC endpoint group: %s", err)
	}

	opts := virtual_gateway.CreateOpts{
		VpcId:             d.Get("vpc_id").(string),
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		BgpAsn:            d.Get("asn").(int),
		DeviceId:          d.Get("device_id").(string),
		RedundantDeviceId: d.Get("redundant_device_id").(string),
		Type:              "default",
		ProjectId:         d.Get("project_id").(string),
	}

	if eg != nil {
		opts.LocalEndpointGroupId = eg.ID
	}

	if egV6 != nil {
		opts.LocalEndpointGroupIpv6Id = egV6.ID
	}

	vg, err := virtual_gateway.Create(client, opts)
	if err != nil {
		return diag.Errorf("error creating opentelekomcloud virtual gateway: %s", err)
	}
	d.SetId(vg.ID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceVirtualGatewayV2Read(clientCtx, d, meta)
}

func resourceVirtualGatewayV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	vg, err := virtual_gateway.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "virtual gateway")
	}

	eg, egV6, err := getLocalGroups(client, *vg)
	if err != nil {
		return diag.Errorf("error querying local groups: %s", err)
	}

	mErr := multierror.Append(nil,
		d.Set("vpc_id", vg.VPCID),
		d.Set("local_ep_group_id", vg.LocalEPGroupID),
		d.Set("local_ep_group_ipv6_id", vg.LocalEPGroupIPv6ID),
		d.Set("name", vg.Name),
		d.Set("description", vg.Description),
		d.Set("asn", vg.BGPASN),
		d.Set("device_id", vg.DeviceID),
		d.Set("redundant_device_id", vg.RedundantDeviceID),
		d.Set("project_id", vg.TenantID),
		d.Set("local_ep_group", eg),
		d.Set("local_ep_group_v6", egV6),
		d.Set("status", vg.Status),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error saving opentelekomcloud virtual gateway fields: %s", err)
	}
	return nil
}

func resourceVirtualGatewayV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	if d.HasChange("local_ep_group") {
		err := updateLocalGroup(client, d, "local_ep_group")
		if err != nil {
			return nil
		}
	}

	if d.HasChange("local_ep_group_v6") {
		err := updateLocalGroup(client, d, "local_ep_group_v6")
		if err != nil {
			return nil
		}
	}

	opts := virtual_gateway.UpdateOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}
	err = virtual_gateway.Update(client, d.Id(), opts)
	if err != nil {
		return diag.Errorf("error updating opentelekomcloud virtual gateway (%s): %s", d.Id(), err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceVirtualGatewayV2Read(clientCtx, d, meta)
}

func resourceVirtualGatewayV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	vg, err := virtual_gateway.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "virtual gateway")
	}

	eg, egV6 := vg.LocalEPGroupID, vg.LocalEPGroupIPv6ID

	err = virtual_gateway.Delete(client, d.Id())
	if err != nil {
		return diag.Errorf("error deleting opentelekomcloud virtual gateway (%s): %s", d.Id(), err)
	}

	if eg != "" {
		err = dceg.Delete(client, eg)
		if err != nil {
			return fmterr.Errorf("error deleting DC endpoint group: %s", err)
		}
	}

	if egV6 != "" {
		err = dceg.Delete(client, egV6)
		if err != nil {
			return fmterr.Errorf("error deleting DC endpoint group: %s", err)
		}
	}

	return nil
}

func flattenGroup(client *golangsdk.ServiceClient, groupId string) (group []map[string]interface{}, err error) {
	eg, err := dceg.Get(client, groupId)
	if err != nil {
		return nil, fmt.Errorf("error reading DC endpoint group: %s", err)
	}
	log.Printf("[DEBUG] DC endpoint group V2 read: %+v", eg)
	group = []map[string]interface{}{
		{
			"name":        eg.Name,
			"description": eg.Description,
			"endpoints":   eg.Endpoints,
			"type":        eg.Type,
		},
	}
	return
}

func getLocalGroups(client *golangsdk.ServiceClient, gateway virtual_gateway.VirtualGateway) (eg, egV6 []map[string]interface{}, err error) {
	if gateway.LocalEPGroupID != "" {
		egResp, err := flattenGroup(client, gateway.LocalEPGroupID)
		if err != nil {
			return nil, nil, err
		}
		eg = egResp
	}
	if gateway.LocalEPGroupIPv6ID != "" {
		egV6Resp, err := flattenGroup(client, gateway.LocalEPGroupIPv6ID)
		if err != nil {
			return nil, nil, err
		}
		egV6 = egV6Resp
	}
	return
}

func updateLocalGroup(client *golangsdk.ServiceClient, d *schema.ResourceData, group string) diag.Diagnostics {
	newEg, err := createGroup(client, d, group)
	if err != nil {
		return fmterr.Errorf("error creating new local DC endpoint group: %s", err)
	}

	vg, err := virtual_gateway.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "virtual gateway")
	}

	var (
		opts       virtual_gateway.UpdateOpts
		oldGroupId string
	)

	if group == "local_ep_group" {
		opts.LocalEndpointGroupId = newEg.ID
		oldGroupId = vg.LocalEPGroupID
	} else {
		opts.LocalEndpointGroupIpv6Id = newEg.ID
		oldGroupId = vg.LocalEPGroupIPv6ID
	}

	err = virtual_gateway.Update(client, d.Id(), opts)
	if err != nil {
		return diag.Errorf("error updating opentelekomcloud virtual gateway (%s): %s", d.Id(), err)
	}

	err = dceg.Delete(client, oldGroupId)
	if err != nil {
		return fmterr.Errorf("error deleting old local DC endpoint group: %s", err)
	}
	return nil
}
