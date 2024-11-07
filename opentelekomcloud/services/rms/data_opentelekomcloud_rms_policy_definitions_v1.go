package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/compliance"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/helper/hashcode"
)

func DataSourcePolicyDefinitionsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePolicyDefinitionsRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_rule_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"trigger_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"keywords": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"definitions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_rule_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_rule": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"trigger_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keywords": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"parameters": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func filterPolicyDefinitionsByKeywords(definitions []compliance.PolicyDefinition,
	keywords []interface{}) []compliance.PolicyDefinition {
	if len(keywords) < 1 {
		return definitions
	}

	filter := common.ExpandToStringList(keywords)
	result := make([]compliance.PolicyDefinition, 0, len(definitions))
	for _, v := range definitions {
		if common.StrSliceContainsAnother(v.Keywords, filter) {
			result = append(result, v)
		}
	}
	return result
}

func flattenDefinitionParameters(parameters map[string]compliance.PolicyParameterDefinition) (
	map[string]interface{}, error) {
	if len(parameters) < 1 {
		return nil, nil
	}

	result := make(map[string]interface{})
	for k, v := range parameters {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("generate json string failed: %s", err)
		}
		result[k] = string(jsonBytes)
	}
	return result, nil
}

func filterPolicyDefinitions(definitions []compliance.PolicyDefinition,
	d *schema.ResourceData) ([]map[string]interface{}, []string, error) {
	filter := map[string]interface{}{
		"Name":           d.Get("name"),
		"PolicyType":     d.Get("policy_type"),
		"PolicyRuleType": d.Get("policy_rule_type"),
		"TriggerType":    d.Get("trigger_type"),
	}
	filtResult, err := common.FilterSliceWithField(definitions, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("filter component runtimes failed: %s", err)
	}
	log.Printf("[DEBUG] Filter %d policy definitions from server through options: %v", len(filtResult), filter)

	result := make([]map[string]interface{}, len(filtResult))
	ids := make([]string, len(filtResult))
	for i, val := range filtResult {
		definition := val.(compliance.PolicyDefinition)
		ids[i] = definition.ID
		dm := map[string]interface{}{
			"id":               definition.ID,
			"name":             definition.Name,
			"policy_type":      definition.PolicyType,
			"description":      definition.Description,
			"policy_rule_type": definition.PolicyRuleType,
			"policy_rule":      definition.PolicyRule,
			"trigger_type":     definition.TriggerType,
			"keywords":         definition.Keywords,
		}

		params, err := flattenDefinitionParameters(definition.Parameters)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to flatten definition parameters: %s", err)
		}
		dm["parameters"] = params

		jsonBytes, err := json.Marshal(definition.PolicyRule)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate json string: %s", err)
		}
		dm["policy_rule"] = string(jsonBytes)

		result[i] = dm
	}
	return result, ids, nil
}

func dataSourcePolicyDefinitionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	definitions, err := compliance.ListAllPolicies(client)
	if err != nil {
		return diag.Errorf("error getting the policy definition list form server: %s", err)
	}

	filterResult := filterPolicyDefinitionsByKeywords(definitions, d.Get("keywords").([]interface{}))
	dm, ids, err := filterPolicyDefinitions(filterResult, d)
	if err != nil {
		return diag.Errorf("error query policy definitions: %s", err)
	}
	d.SetId(hashcode.Strings(ids))

	if err = d.Set("definitions", dm); err != nil {
		return diag.Errorf("error saving the information of the policy definitions to state: %s", err)
	}
	return nil
}
