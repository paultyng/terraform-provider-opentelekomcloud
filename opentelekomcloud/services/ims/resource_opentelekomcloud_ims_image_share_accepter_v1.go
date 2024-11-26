package ims

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ims/v1/members"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceImsImageShareAcceptV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImsImageShareAcceptCreate,
		ReadContext:   resourceImsImageShareAcceptRead,
		DeleteContext: resourceImsImageShareAcceptDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vault_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceImsImageShareAcceptCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ImageV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud image client: %s", err)
	}

	jobId, err := members.BatchUpdateMembers(client, members.BatchUpdateMembersOpts{
		Images:    []string{d.Get("image_id").(string)},
		ProjectId: client.ProjectID,
		Status:    "accepted",
		VaultId:   d.Get("vault_id").(string),
	})
	if err != nil {
		return fmterr.Errorf("error requesting OpenTelekomCloud ims share accept: %w", err)
	}
	err = waitForImageShareOrAcceptJobSuccess(ctx, d, client, *jobId, schema.TimeoutCreate)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceId, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}
	d.SetId(resourceId)

	return resourceImsImageShareAcceptRead(ctx, d, meta)
}

func resourceImsImageShareAcceptRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceImsImageShareAcceptDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ImageV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud image client: %s", err)
	}

	jobId, err := members.BatchUpdateMembers(client, members.BatchUpdateMembersOpts{
		Images:    []string{d.Get("image_id").(string)},
		ProjectId: client.ProjectID,
		Status:    "rejected",
		VaultId:   d.Get("vault_id").(string),
	})
	if err != nil {
		return fmterr.Errorf("error requesting OpenTelekomCloud ims share reject: %w", err)
	}
	err = waitForImageShareOrAcceptJobSuccess(ctx, d, client, *jobId, schema.TimeoutCreate)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
