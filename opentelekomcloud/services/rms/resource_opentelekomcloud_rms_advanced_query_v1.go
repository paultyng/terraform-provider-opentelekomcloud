package rms

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/advanced"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceRmsAdvancedQueryV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAdvancedQueryCreate,
		UpdateContext: resourceAdvancedQueryUpdate,
		ReadContext:   resourceAdvancedQueryRead,
		DeleteContext: resourceAdvancedQueryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
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
		},
	}
}

func resourceAdvancedQueryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	domainID := GetRmsDomainId(client, config)

	createOpts := advanced.CreateQueryOpts{
		DomainId:    domainID,
		Name:        d.Get("name").(string),
		Expression:  d.Get("expression").(string),
		Description: d.Get("description").(string),
	}

	resp, err := advanced.CreateQuery(client, createOpts)
	if err != nil {
		return diag.Errorf("error creating RMS advanced query: %s", err)
	}

	d.SetId(resp.Id)

	clientCtx := common.CtxWithClient(ctx, client, errCreationRMSV1Client)
	return resourceAdvancedQueryRead(clientCtx, d, meta)
}

func resourceAdvancedQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	domainID := GetRmsDomainId(client, config)

	resp, err := advanced.GetQuery(client, domainID, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving RMS advanced query")
	}

	mErr := multierror.Append(nil,
		d.Set("name", resp.Name),
		d.Set("expression", resp.Expression),
		d.Set("description", resp.Description),
		d.Set("type", resp.Type),
		d.Set("created_at", resp.Created),
		d.Set("updated_at", resp.Updated),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceAdvancedQueryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	updateAdvancedQueryChanges := []string{
		"expression",
		"description",
	}

	if d.HasChanges(updateAdvancedQueryChanges...) {
		config := meta.(*cfg.Config)
		client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
			return config.RmsV1Client(config.GetRegion(d))
		})
		if err != nil {
			return fmterr.Errorf(errCreationRMSV1Client, err)
		}

		domainID := GetRmsDomainId(client, config)

		updateOpts := advanced.UpdateQueryOpts{
			DomainId:    domainID,
			QueryId:     d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Expression:  d.Get("expression").(string),
		}

		_, err = advanced.UpdateQuery(client, updateOpts)
		if err != nil {
			return diag.Errorf("error updating RMS advanced query: %s", err)
		}
	}
	return resourceAdvancedQueryRead(ctx, d, meta)
}

func resourceAdvancedQueryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	domainID := GetRmsDomainId(client, config)

	err = advanced.DeleteQuery(client, domainID, d.Id())
	if err != nil {
		return diag.Errorf("error deleting RMS advanced query: %s", err)
	}

	return nil
}
