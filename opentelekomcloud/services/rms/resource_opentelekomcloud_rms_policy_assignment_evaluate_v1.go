package rms

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/compliance"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceRmsPolicyAssignmentEvalV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyAssignmentEvaluateCreate,
		ReadContext:   resourcePolicyAssignmentEvaluateRead,
		DeleteContext: resourcePolicyAssignmentEvaluateDelete,

		Schema: map[string]*schema.Schema{
			"policy_assignment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAssignmentEvaluateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	domainId := GetRmsDomainId(client, config)

	err = compliance.RunEval(client, domainId, d.Get("policy_assignment_id").(string))

	if err != nil {
		return diag.Errorf("error creating RMS policy assignment evaluate: %s", err)
	}

	d.SetId(d.Get("policy_assignment_id").(string))

	return resourcePolicyAssignmentEvaluateRead(ctx, d, meta)
}

func resourcePolicyAssignmentEvaluateRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourcePolicyAssignmentEvaluateDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	errorMsg := "Deleting policy assignment evaluate is not supported. The policy assignment evaluate is only removed from the state."
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  errorMsg,
		},
	}
}
