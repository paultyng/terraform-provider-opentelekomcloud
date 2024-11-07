package rms

import (
	"context"
	"encoding/json"
	"fmt"

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

func DataSourceRmsAdvancedQuerySchemasV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRmsAdvancedQuerySchemasRead,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schemas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schema": {
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

func dataSourceRmsAdvancedQuerySchemasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	res, err := advanced.ListSchemas(client, advanced.ListSchemasOpts{
		DomainId: GetRmsDomainId(client, config),
	})
	if err != nil {
		return diag.Errorf("error getting the advanced query schemas list from server: %s", err)
	}

	schemaType := d.Get("type").(string)

	var filteredSchemas []advanced.Schema
	if schemaType != "" {
		for _, schema := range res {
			if schema.Type == schemaType {
				filteredSchemas = append(filteredSchemas, schema)
			}
		}
	} else {
		filteredSchemas = res
	}

	mErr := &multierror.Error{}

	if err := d.Set("schemas", flattenSchemas(filteredSchemas)); err != nil {
		mErr = multierror.Append(mErr, fmt.Errorf("error setting schemas: %s", err))
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		mErr = multierror.Append(mErr, fmt.Errorf("error generating UUID: %s", err))
	}

	if mErr.ErrorOrNil() != nil {
		return diag.FromErr(mErr)
	}

	d.SetId(id)
	return nil
}

func flattenSchemas(schemas []advanced.Schema) []map[string]interface{} {
	if len(schemas) == 0 {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0, len(schemas))

	for _, schema := range schemas {
		schemaMap := map[string]interface{}{
			"type":   schema.Type,
			"schema": flattenSchemaValue(schema.Schema),
		}
		result = append(result, schemaMap)
	}

	return result
}

func flattenSchemaValue(schemaInterface interface{}) map[string]string {
	if schemaInterface == nil {
		return map[string]string{}
	}

	result := make(map[string]string)

	if schemaMap, ok := schemaInterface.(map[string]interface{}); ok {
		for key, value := range schemaMap {
			jsonBytes, err := json.Marshal(value)
			if err == nil {
				result[key] = string(jsonBytes)
			}
		}
	}

	return result
}
