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

func DataSourceRmsAdvancedQueriesV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRmsAdvancedQueriesRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"queries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expression": {
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
				},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRmsAdvancedQueriesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	queries, err := advanced.ListQueries(client, advanced.ListQueriesOpts{
		DomainId: GetRmsDomainId(client, config),
		Name:     d.Get("name").(string),
	})

	if err != nil {
		return diag.Errorf("error getting the advanced query list from server: %s", err)
	}

	var flattenQueries []map[string]interface{}

	for _, item := range queries {
		query := map[string]interface{}{
			"name":        item.Name,
			"id":          item.Id,
			"type":        item.Type,
			"description": item.Description,
			"expression":  item.Expression,
			"created_at":  item.Created,
			"updated_at":  item.Updated,
		}
		flattenQueries = append(flattenQueries, query)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	mErr := multierror.Append(
		d.Set("region", config.GetRegion(d)),
		d.Set("queries", flattenQueries),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
