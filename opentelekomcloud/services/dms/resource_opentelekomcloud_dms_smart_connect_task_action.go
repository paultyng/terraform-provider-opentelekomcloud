package dms

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/smart_connect"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDmsSmartConnectTaskAction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsV2SmartConnectTaskActionCreate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"task_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"restart", "resume", "pause",
				}, true),
			},
			"task_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDmsV2SmartConnectTaskActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dmsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	instanceId := d.Get("instance_id").(string)
	taskId := d.Get("task_id").(string)
	action := d.Get("action").(string)

	switch action {
	case "pause":
		err = smart_connect.PauseTask(client, instanceId, taskId)
		if err != nil {
			return diag.Errorf("error pausing DMS kafka smart connect task: %v", err)
		}
	case "resume":
		err = smart_connect.RestartTask(client, instanceId, taskId)
		if err != nil {
			return diag.Errorf("error resuming DMS kafka smart connect task: %v", err)
		}
	case "restart":
		err = smart_connect.StartOrRestartTask(client, instanceId, taskId)
		if err != nil {
			return diag.Errorf("error resuming DMS kafka smart connect task: %v", err)
		}
	default:
		return diag.Errorf("invalid action")
	}

	getTask, err := smart_connect.GetTask(client, d.Get("instance_id").(string), d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving DMS kafka smart connect task")
	}

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("status", getTask.Status),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}
