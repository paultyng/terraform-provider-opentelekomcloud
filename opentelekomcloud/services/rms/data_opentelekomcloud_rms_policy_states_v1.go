package rms

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/compliance"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourcePolicyStates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePolicyStatesRead,

		Schema: map[string]*schema.Schema{
			"policy_assignment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"compliance_state": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"states": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_provider": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"trigger_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"compliance_state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_assignment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_assignment_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_definition_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"evaluation_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func dataSourcePolicyStatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	var policyStates []compliance.PolicyState
	if _, ok := d.GetOk("policy_assignment_id"); ok {
		policyStates, err = compliance.ListAllRuleCompliance(client, compliance.ListAllComplianceOpts{
			DomainId: GetRmsDomainId(client, config),
			PolicyId: d.Get("policy_assignment_id").(string),
		})

		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		policyStates, err = compliance.ListAllUserCompliance(client, compliance.ListAllUserComplianceOpts{
			DomainId: GetRmsDomainId(client, config),
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	filteredStates := filterPolicyStates(policyStates, d)

	if len(filteredStates) == 0 {
		return diag.FromErr(fmt.Errorf("policy states not found"))
	}

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}
	d.SetId(uuid)

	states := make([]map[string]interface{}, len(filteredStates))
	for i, state := range filteredStates {
		states[i] = map[string]interface{}{
			"domain_id":              state.DomainID,
			"region_id":              state.RegionID,
			"resource_id":            state.ResourceID,
			"resource_name":          state.ResourceName,
			"resource_provider":      state.ResourceProvider,
			"resource_type":          state.ResourceType,
			"trigger_type":           state.TriggerType,
			"compliance_state":       state.ComplianceState,
			"policy_assignment_id":   state.PolicyAssignmentID,
			"policy_assignment_name": state.PolicyAssignmentName,
			"policy_definition_id":   state.PolicyDefinitionID,
			"evaluation_time":        state.EvaluationTime,
		}
	}

	mErr := multierror.Append(
		nil,
		d.Set("states", states),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func filterPolicyStates(states []compliance.PolicyState, d *schema.ResourceData) []compliance.PolicyState {
	var filtered []compliance.PolicyState

	for _, state := range states {
		if v, ok := d.GetOk("compliance_state"); ok && v.(string) != state.ComplianceState {
			continue
		}

		if v, ok := d.GetOk("resource_name"); ok && v.(string) != state.ResourceName {
			continue
		}

		if v, ok := d.GetOk("resource_id"); ok && v.(string) != state.ResourceID {
			continue
		}

		filtered = append(filtered, state)
	}

	if len(filtered) == 0 {
		return states
	}

	return filtered
}
