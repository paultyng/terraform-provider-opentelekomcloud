package ims

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ims/v1/members"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ims/v1/others"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceImsImageShareV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImsImageShareCreate,
		UpdateContext: resourceImsImageShareUpdate,
		ReadContext:   resourceImsImageShareRead,
		DeleteContext: resourceImsImageShareDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"source_image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_project_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceImsImageShareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ImageV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud image client: %s", err)
	}

	sourceImageId := d.Get("source_image_id").(string)
	jobId, err := members.BatchAddMembers(client, members.BatchMembersOpts{
		Images:   []string{d.Get("source_image_id").(string)},
		Projects: common.ExpandToStringSlice(d.Get("target_project_ids").(*schema.Set).List()),
	})
	if err != nil {
		return fmterr.Errorf("error requesting share for private image: %w", err)
	}
	err = waitForImageShareOrAcceptJobSuccess(ctx, d, client, *jobId, schema.TimeoutCreate)
	if err != nil {
		return fmterr.Errorf("error while waiting share for private image to become active: %w", err)
	}
	d.SetId(sourceImageId)

	return resourceImsImageShareRead(ctx, d, meta)
}

func resourceImsImageShareRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceImsImageShareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ImageV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud image client: %s", err)
	}

	if d.HasChange("target_project_ids") {
		oProjectIdsRaw, nProjectIdsRaw := d.GetChange("target_project_ids")
		shareProjectIds := nProjectIdsRaw.(*schema.Set).Difference(oProjectIdsRaw.(*schema.Set))
		unShareProjectIds := oProjectIdsRaw.(*schema.Set).Difference(nProjectIdsRaw.(*schema.Set))
		if shareProjectIds.Len() > 0 {
			jobId, err := members.BatchAddMembers(client, members.BatchMembersOpts{
				Images:   []string{d.Id()},
				Projects: common.ExpandToStringSlice(shareProjectIds.List()),
			})
			if err != nil {
				return fmterr.Errorf("error requesting share for private image: %w", err)
			}
			err = waitForImageShareOrAcceptJobSuccess(ctx, d, client, *jobId, schema.TimeoutCreate)
			if err != nil {
				return fmterr.Errorf("error while waiting share for private image to become active: %w", err)
			}
		}
		if unShareProjectIds.Len() > 0 {
			jobId, err := members.BatchDeleteMembers(client, members.BatchMembersOpts{
				Images:   []string{d.Id()},
				Projects: common.ExpandToStringSlice(unShareProjectIds.List()),
			})
			if err != nil {
				return fmterr.Errorf("error requesting share for private image: %w", err)
			}
			err = waitForImageShareOrAcceptJobSuccess(ctx, d, client, *jobId, schema.TimeoutDelete)
			if err != nil {
				return fmterr.Errorf("error while waiting share for private image to become deleted: %w", err)
			}
		}
	}

	return resourceImsImageShareRead(ctx, d, meta)
}

func resourceImsImageShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ImageV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud image client: %s", err)
	}

	projectIds := d.Get("target_project_ids")
	jobId, err := members.BatchDeleteMembers(client, members.BatchMembersOpts{
		Images:   []string{d.Id()},
		Projects: common.ExpandToStringSlice(projectIds.(*schema.Set).List()),
	})
	if err != nil {
		return fmterr.Errorf("error requesting delete share for private image: %w", err)
	}
	err = waitForImageShareOrAcceptJobSuccess(ctx, d, client, *jobId, schema.TimeoutDelete)
	if err != nil {
		return fmterr.Errorf("error while waiting share for private image to become deleted: %w", err)
	}

	return nil
}

func waitForImageShareOrAcceptJobSuccess(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient,
	jobId, timeout string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"INIT", "RUNNING"},
		Target:     []string{"SUCCESS"},
		Refresh:    imageShareOrAcceptJobStatusRefreshFunc(jobId, client),
		Timeout:    d.Timeout(timeout),
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for job (%s) success: %s", jobId, err)
	}

	return nil
}

func imageShareOrAcceptJobStatusRefreshFunc(jobId string, client *golangsdk.ServiceClient) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := others.ShowJob(client, jobId)
		if err != nil {
			return nil, "", err
		}
		return n, n.Status, nil
	}
}
