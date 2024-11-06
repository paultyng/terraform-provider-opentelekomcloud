package rms

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/advanced"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceAdvancedQueryV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAdvancedQueryRead,

		Schema: map[string]*schema.Schema{
			"expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_info": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"select_fields": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func dataSourceAdvancedQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	resp, err := advanced.RunQuery(client, advanced.RunQueryOpts{
		DomainId:   GetRmsDomainId(client, config),
		Expression: d.Get("expression").(string),
	})

	if err != nil {
		return diag.Errorf("error getting the advanced query from server: %s", err)
	}

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}
	d.SetId(uuid)

	mErr := multierror.Append(
		nil,
		d.Set("results", flattenResults(resp.Results)),
		d.Set("query_info", flattenQueryInfo(&resp.QueryInfo)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenResults(results []interface{}) []map[string]string {
	if len(results) == 0 {
		return []map[string]string{}
	}

	flattenedResults := make([]map[string]string, 0, len(results))

	for _, item := range results {
		if item == nil {
			continue
		}

		if itemMap, ok := item.(map[string]interface{}); ok {
			stringMap := make(map[string]string)
			for key, value := range itemMap {
				stringMap[key] = value.(string)
			}
			flattenedResults = append(flattenedResults, stringMap)
		}
	}

	return flattenedResults
}

func flattenQueryInfo(queryInfo *advanced.QueryInfo) []map[string]interface{} {
	if queryInfo == nil {
		return []map[string]interface{}{}
	}

	return []map[string]interface{}{
		{
			"select_fields": queryInfo.SelectFields,
		},
	}
}
