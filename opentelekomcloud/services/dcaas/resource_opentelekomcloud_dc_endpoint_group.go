package dcaas

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	dceg "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v2/dc-endpoint-group"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDCEndpointGroupV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDCEndpointGroupV2Create,
		DeleteContext: resourceDCEndpointGroupV2Delete,
		ReadContext:   resourceDCEndpointGroupV2Read,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		DeprecationMessage: egDeprecated,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"endpoints": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDCEndpointGroupV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	createOpts := dceg.CreateOpts{
		Name:        d.Get("name").(string),
		TenantId:    d.Get("project_id").(string),
		Description: d.Get("description").(string),
		Endpoints:   GetEndpointsValues(d),
		Type:        d.Get("type").(string),
	}
	log.Printf("[DEBUG] DC endpoint group V2 createOpts: %+v", createOpts)

	eg, err := dceg.Create(client, createOpts)
	if err != nil {
		return fmterr.Errorf("error creating DC endpoint group: %s", err)
	}
	d.SetId(eg.ID)
	log.Printf("[DEBUG] DC endpoint group V2 created: %+v", eg)
	return resourceDCEndpointGroupV2Read(ctx, d, meta)
}

func resourceDCEndpointGroupV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	eg, err := dceg.Get(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error reading DC endpoint group: %s", err)
	}
	log.Printf("[DEBUG] DC endpoint group V2 read: %+v", eg)

	mErr := multierror.Append(
		d.Set("name", eg.Name),
		d.Set("project_id", eg.TenantId),
		d.Set("description", eg.Description),
		d.Set("endpoints", eg.Endpoints),
		d.Set("type", eg.Type),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDCEndpointGroupV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV2, err)
	}

	err = dceg.Delete(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error deleting DC endpoint group: %s", err)
	}
	log.Printf("[DEBUG] DC endpoint group V2 deleted: %s", d.Id())
	d.SetId("")
	return nil
}

func GetEndpointsValues(d *schema.ResourceData) []string {
	endpoints := make([]string, 0)
	for _, val := range d.Get("endpoints").([]interface{}) {
		endpoints = append(endpoints, val.(string))
	}
	return endpoints
}
