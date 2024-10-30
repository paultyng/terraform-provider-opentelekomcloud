package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/compliance"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

const (
	AssignmentTypeBuiltin = "builtin"
	AssignmentTypeCustom  = "custom"

	AssignmentStatusDisabled   = "Disabled"
	AssignmentStatusEnabled    = "Enabled"
	AssignmentStatusEvaluating = "Evaluating"
)

func ResourceRmsPolicyAssignmentV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyAssignmentCreate,
		ReadContext:   resourcePolicyAssignmentRead,
		UpdateContext: resourcePolicyAssignmentUpdate,
		DeleteContext: resourcePolicyAssignmentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_definition_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"period": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_provider": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tag_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tag_value": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"policy_filter.0.tag_key"},
						},
					},
				},
			},
			"custom_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_urn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"auth_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"auth_value": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringIsJSON,
							},
						},
					},
				},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsJSON,
				},
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func buildPolicyFilter(filters []interface{}) compliance.PolicyFilterDefinition {
	if len(filters) < 1 {
		return compliance.PolicyFilterDefinition{}
	}
	filter := filters[0].(map[string]interface{})
	return compliance.PolicyFilterDefinition{
		RegionID:         filter["region"].(string),
		ResourceProvider: filter["resource_provider"].(string),
		ResourceType:     filter["resource_type"].(string),
		ResourceID:       filter["resource_id"].(string),
		TagKey:           filter["tag_key"].(string),
		TagValue:         filter["tag_value"].(string),
	}
}

func buildCustomPolicy(policies []interface{}) (*compliance.CustomPolicy, error) {
	if len(policies) < 1 {
		return nil, nil
	}
	policy := policies[0].(map[string]interface{})
	result := compliance.CustomPolicy{
		FunctionUrn: policy["function_urn"].(string),
		AuthType:    policy["auth_type"].(string),
	}
	authValues := make(map[string]interface{})
	for k, jsonVal := range policy["auth_value"].(map[string]interface{}) {
		var value interface{}
		err := json.Unmarshal([]byte(jsonVal.(string)), &value)
		if err != nil {
			return &result, fmt.Errorf("error analyzing authorization value: %s", err)
		}
		authValues[k] = value
	}
	result.AuthValue = authValues

	return &result, nil
}

func buildRuleParameters(parameters map[string]interface{}) (map[string]compliance.PolicyParameter, error) {
	if len(parameters) < 1 {
		return nil, nil
	}
	result := make(map[string]compliance.PolicyParameter)
	for k, jsonVal := range parameters {
		var value interface{}
		err := json.Unmarshal([]byte(jsonVal.(string)), &value)
		if err != nil {
			return result, fmt.Errorf("error analyzing parameter value: %s", err)
		}
		result[k] = compliance.PolicyParameter{
			Value: value,
		}
	}
	return result, nil
}

func buildPolicyAssignmentCreateOpts(d *schema.ResourceData) (compliance.AddRuleOpts, error) {
	result := compliance.AddRuleOpts{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		PolicyAssignmentType: AssignmentTypeBuiltin,
		PolicyFilter:         buildPolicyFilter(d.Get("policy_filter").([]interface{})),
		PolicyDefinitionID:   d.Get("policy_definition_id").(string),
		Period:               d.Get("period").(string),
	}
	customPolicy, err := buildCustomPolicy(d.Get("custom_policy").([]interface{}))
	if err != nil {
		return result, err
	}
	result.CustomPolicy = customPolicy
	if customPolicy != nil {
		result.PolicyAssignmentType = AssignmentTypeCustom
	}

	parameters, err := buildRuleParameters(d.Get("parameters").(map[string]interface{}))
	if err != nil {
		return result, err
	}
	result.Parameters = parameters

	return result, nil
}

func updatePolicyAssignmentStatus(client *golangsdk.ServiceClient, domainId, assignmentId,
	statusConfig string) (err error) {
	switch statusConfig {
	case AssignmentStatusDisabled:
		err = compliance.DisableRule(client, domainId, assignmentId)
	case AssignmentStatusEnabled:
		err = compliance.EnableRule(client, domainId, assignmentId)
	}
	return
}

func policyAssignmentRefreshFunc(client *golangsdk.ServiceClient, domainId,
	assignmentId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := compliance.GetRule(client, domainId, assignmentId)
		if err != nil {
			return resp, "ERROR", err
		}
		return resp, resp.State, nil
	}
}

func resourcePolicyAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	opts, err := buildPolicyAssignmentCreateOpts(d)
	if err != nil {
		return diag.Errorf("error creating the create option structure of the RMS policy assignment: %s", err)
	}
	domainId := GetRmsDomainId(client, config)
	opts.DomainId = domainId

	resp, err := compliance.AddRule(client, opts)
	if err != nil {
		return diag.Errorf("error creating policy assignment resource: %s", err)
	}

	assignmentId := resp.ID
	d.SetId(assignmentId)

	// it will take too long time to become enabled when the resources are very huge.
	// so we wait for the enabled status only when user want to disable it during creating.
	if statusConfig := d.Get("status").(string); statusConfig == AssignmentStatusDisabled {
		log.Printf("[DEBUG] Waiting for the policy assignment (%s) status to become enabled, then disable it", assignmentId)
		stateConf := &resource.StateChangeConf{
			Pending:                   []string{AssignmentStatusDisabled, AssignmentStatusEvaluating},
			Target:                    []string{AssignmentStatusEnabled},
			Refresh:                   policyAssignmentRefreshFunc(client, domainId, assignmentId),
			Timeout:                   d.Timeout(schema.TimeoutCreate),
			Delay:                     10 * time.Second,
			PollInterval:              10 * time.Second,
			ContinuousTargetOccurence: 2,
		}
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("error waiting for the policy assignment (%s) status to become enabled: %s",
				assignmentId, err)
		}

		err = updatePolicyAssignmentStatus(client, domainId, assignmentId, statusConfig)
		if err != nil {
			return diag.Errorf("error disabling the status of the policy assignment: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, errCreationRMSV1Client)
	return resourcePolicyAssignmentRead(clientCtx, d, meta)
}

func flattenPolicyFilter(filter compliance.PolicyFilterDefinition) []map[string]interface{} {
	if reflect.DeepEqual(filter, compliance.PolicyFilterDefinition{}) {
		return nil
	}

	return []map[string]interface{}{
		{
			"region":            filter.RegionID,
			"resource_provider": filter.ResourceProvider,
			"resource_type":     filter.ResourceType,
			"resource_id":       filter.ResourceID,
			"tag_key":           filter.TagKey,
			"tag_value":         filter.TagValue,
		},
	}
}

func flattenCustomPolicy(customPolicy *compliance.CustomPolicy) ([]map[string]interface{}, error) {
	if customPolicy == nil {
		return nil, nil
	}

	authValues := make(map[string]interface{})
	for k, v := range customPolicy.AuthValue {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("generate json string failed: %s", err)
		}
		authValues[k] = string(jsonBytes)
	}
	return []map[string]interface{}{
		{
			"function_urn": customPolicy.FunctionUrn,
			"auth_type":    customPolicy.AuthType,
			"auth_value":   authValues,
		},
	}, nil
}

func flattenPolicyParameters(parameters map[string]compliance.PolicyParameter) (map[string]interface{},
	error) {
	if len(parameters) < 1 {
		return nil, nil
	}

	result := make(map[string]interface{})
	for k, v := range parameters {
		jsonBytes, err := json.Marshal(v.Value)
		if err != nil {
			return nil, fmt.Errorf("generate json string failed: %s", err)
		}
		result[k] = string(jsonBytes)
	}
	return result, nil
}

func resourcePolicyAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	domainId := GetRmsDomainId(client, config)
	assignmentId := d.Id()
	resp, err := compliance.GetRule(client, domainId, assignmentId)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "RMS policy assignment")
	}

	customPolicy, err := flattenCustomPolicy(resp.CustomPolicy)
	if err != nil {
		return diag.FromErr(err)
	}
	parameters, err := flattenPolicyParameters(resp.Parameters)
	if err != nil {
		return diag.FromErr(err)
	}
	mErr := multierror.Append(nil,
		d.Set("type", resp.PolicyAssignmentType),
		d.Set("name", resp.Name),
		d.Set("description", resp.Description),
		d.Set("policy_definition_id", resp.PolicyDefinitionID),
		d.Set("period", resp.Period),
		d.Set("policy_filter", flattenPolicyFilter(*resp.PolicyFilter)),
		d.Set("custom_policy", customPolicy),
		d.Set("parameters", parameters),
		d.Set("status", resp.State),
		d.Set("created_at", resp.Created),
		d.Set("updated_at", resp.Updated),
	)

	if mErr.ErrorOrNil() != nil {
		return diag.Errorf("error saving policy assignment resource (%s) fields: %s", assignmentId, mErr)
	}
	return nil
}

func buildPolicyAssignmentUpdateOpts(d *schema.ResourceData) (compliance.UpdateRuleOpts, error) {
	result := compliance.UpdateRuleOpts{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		PolicyAssignmentType: AssignmentTypeBuiltin,
		PolicyFilter:         buildPolicyFilter(d.Get("policy_filter").([]interface{})),
		PolicyDefinitionID:   d.Get("policy_definition_id").(string),
		Period:               d.Get("period").(string),
	}
	customPolicy, err := buildCustomPolicy(d.Get("custom_policy").([]interface{}))
	if err != nil {
		return result, err
	}
	result.CustomPolicy = customPolicy
	if customPolicy != nil {
		result.PolicyAssignmentType = AssignmentTypeCustom
	}

	parameters, err := buildRuleParameters(d.Get("parameters").(map[string]interface{}))
	if err != nil {
		return result, err
	}
	result.Parameters = parameters

	return result, nil
}

func resourcePolicyAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	assignmentId := d.Id()
	domainId := GetRmsDomainId(client, config)

	if d.HasChange("status") {
		oldVal, newVal := d.GetChange("status")
		err = updatePolicyAssignmentStatus(client, domainId, d.Id(), d.Get("status").(string))
		if err != nil {
			return diag.Errorf("error updating the status of the policy assignment (%s): %s", assignmentId, err)
		}

		if newVal.(string) == AssignmentStatusEnabled {
			log.Printf("[DEBUG] Waiting for the policy assignment (%s) status to become %s.", assignmentId,
				strings.ToLower(newVal.(string)))
			stateConf := &resource.StateChangeConf{
				Pending:                   []string{oldVal.(string)},
				Target:                    []string{AssignmentStatusEvaluating, AssignmentStatusEnabled},
				Refresh:                   policyAssignmentRefreshFunc(client, domainId, assignmentId),
				Timeout:                   d.Timeout(schema.TimeoutUpdate),
				Delay:                     10 * time.Second,
				PollInterval:              10 * time.Second,
				ContinuousTargetOccurence: 2,
			}
			_, err = stateConf.WaitForStateContext(ctx)
			if err != nil {
				return diag.Errorf("error waiting for the policy assignment (%s) status to become %s: %s",
					assignmentId, strings.ToLower(newVal.(string)), err)
			}
		}
	}
	if d.HasChangeExcept("status") {
		opts, err := buildPolicyAssignmentUpdateOpts(d)
		if err != nil {
			return diag.Errorf("error creating the update option structure of the RMS policy assignment: %s", err)
		}

		opts.DomainId = domainId
		opts.PolicyAssignmentId = assignmentId

		_, err = compliance.UpdateRule(client, opts)
		if err != nil {
			return diag.Errorf("error updating policy assignment resource (%s): %s", assignmentId, err)
		}
		currentStatus := d.Get("status").(string)
		log.Printf("[DEBUG] Waiting for the policy assignment (%s) status to become %s.", assignmentId,
			strings.ToLower(currentStatus))
		stateConf := &resource.StateChangeConf{
			Target:                    []string{currentStatus},
			Refresh:                   policyAssignmentRefreshFunc(client, domainId, assignmentId),
			Timeout:                   d.Timeout(schema.TimeoutUpdate),
			Delay:                     10 * time.Second,
			PollInterval:              10 * time.Second,
			ContinuousTargetOccurence: 2,
		}
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("error waiting for the policy assignment (%s) status to become %s: %s",
				assignmentId, strings.ToLower(currentStatus), err)
		}
	}

	return resourcePolicyAssignmentRead(ctx, d, meta)
}

func resourcePolicyAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}
	var (
		assignmentId = d.Id()
		domainId     = GetRmsDomainId(client, config)
	)
	if d.Get("status").(string) == AssignmentStatusEnabled {
		err = compliance.DisableRule(client, domainId, assignmentId)
		if err != nil {
			return diag.Errorf("failed to disable the policy assignment (%s): %s", assignmentId, err)
		}
	}

	err = compliance.Delete(client, domainId, assignmentId)
	if err != nil {
		return diag.Errorf("error deleting the policy assignment (%s): %s", assignmentId, err)
	}
	return nil
}
