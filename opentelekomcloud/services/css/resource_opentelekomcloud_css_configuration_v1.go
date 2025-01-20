package css

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	pc "github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/parameter-configuration"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceCssConfigurationV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCssConfigurationV1Update,
		UpdateContext: resourceCssConfigurationV1Update,
		ReadContext:   resourceCssConfigurationV1Read,
		DeleteContext: resourceCssConfigurationV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `The CSS cluster ID.`,
			},
			"http_cors_allow_credentials": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Whether to return the Access-Control-Allow-Credentials of the header during cross-domain access.`,
			},
			"http_cors_allow_origin": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Origin IP address allowed for cross-domain access, for example, **122.122.122.122:9200**.`,
			},
			"http_cors_max_age": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Cache duration of the browser. The cache is automatically cleared after the time range you specify.`,
			},
			"http_cors_allow_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Headers allowed for cross-domain access.`,
			},
			"http_cors_enabled": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Whether to allow cross-domain access.`,
			},
			"http_cors_allow_methods": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Methods allowed for cross-domain access.`,
			},
			"reindex_remote_whitelist": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Configured for migrating data from the current cluster to the target cluster through the reindex API.`,
			},
			"indices_queries_cache_size": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Cache size in the query phase. Value range: **1** to **100**.`,
			},
			"thread_pool_force_merge_size": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Queue size in the force merge thread pool.`,
			},
			"auto_create_index": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Whether to auto-create index.`,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCssConfigurationV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.CssV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	_, err = pc.Modify(client, buildUpdateConfigurationBodyParams(d), d.Get("cluster_id").(string))
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud CSS configuration: %s", err)
	}

	d.SetId(d.Get("cluster_id").(string))

	err = configurationWaitingForStateCompleted(ctx, d, meta, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.Errorf("error waiting for the OpenTelekomCloud CSS configuration (%s) update to complete: %s", d.Id(), err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV1)
	return resourceCssConfigurationV1Read(clientCtx, d, meta)
}

func buildUpdateConfigurationBodyParams(d *schema.ResourceData) pc.ModifyOpts {
	elasticsearchYml := make(map[string]interface{})

	common.SetIfNotEmpty(elasticsearchYml, "http.cors.allow-credentials", d.Get("http_cors_allow_credentials"))
	common.SetIfNotEmpty(elasticsearchYml, "http.cors.allow-origin", d.Get("http_cors_allow_origin"))
	common.SetIfNotEmpty(elasticsearchYml, "http.cors.max-age", d.Get("http_cors_max_age"))
	common.SetIfNotEmpty(elasticsearchYml, "http.cors.allow-headers", d.Get("http_cors_allow_headers"))
	common.SetIfNotEmpty(elasticsearchYml, "http.cors.enabled", d.Get("http_cors_enabled"))
	common.SetIfNotEmpty(elasticsearchYml, "http.cors.allow-methods", d.Get("http_cors_allow_methods"))
	common.SetIfNotEmpty(elasticsearchYml, "reindex.remote.whitelist", d.Get("reindex_remote_whitelist"))
	common.SetIfNotEmpty(elasticsearchYml, "indices.queries.cache.size", d.Get("indices_queries_cache_size"))
	common.SetIfNotEmpty(elasticsearchYml, "thread_pool.force_merge.size", d.Get("thread_pool_force_merge_size"))
	common.SetIfNotEmpty(elasticsearchYml, "action.auto_create_index", d.Get("auto_create_index"))

	opts := pc.ModifyOpts{
		Edit: map[string]interface{}{
			"modify": map[string]interface{}{
				"elasticsearch.yml": elasticsearchYml,
			},
		},
	}
	return opts
}

func resourceCssConfigurationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.CssV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	configurations, err := pc.List(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error getting OpenTelekomCloud CSS configurations")
	}

	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("cluster_id", d.Id()),
	)

	keyMappings := map[string]string{
		"http.cors.allow-credentials":  "http_cors_allow_credentials",
		"http.cors.cors.allow-origin":  "http_cors_allow_origin",
		"http.cors.max-age":            "http_cors_max_age",
		"http.cors.allow-headers":      "http_cors_allow_headers",
		"http.cors.enabled":            "http_cors_enabled",
		"http.cors.allow-methods":      "http_cors_allow_methods",
		"reindex.remote.whitelist":     "reindex_remote_whitelist",
		"indices.queries.cache.size":   "indices_queries_cache_size",
		"thread_pool.force_merge.size": "thread_pool_force_merge_size",
		"action.auto_create_index":     "auto_create_index",
	}
	for key, c := range configurations.Templates {
		if mappedKey, exists := keyMappings[key]; exists {
			mErr = multierror.Append(mErr, d.Set(mappedKey, c.Value))
		}
	}

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceCssConfigurationV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.CssV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}
	delMsg := "error deleting OpenTelekomCloud CSS configuration"
	_, err = pc.Modify(client, buildResetConfigurationBodyParams(), d.Id())
	if err != nil {
		// The cluster does not exist, http code is 403, key/value of error code is errCode/CSS.0015
		return common.CheckDeletedDiag(d,
			common.ConvertExpected403ErrInto404Err(err, "errCode", "CSS.0015"), delMsg)
	}

	err = configurationWaitingForStateCompleted(ctx, d, meta, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.Errorf("error waiting for the OpenTelekomCloud CSS configuration (%s) deletion to complete: %s", d.Id(), err)
	}
	return nil
}

// Reset to default value.
func buildResetConfigurationBodyParams() pc.ModifyOpts {
	opts := pc.ModifyOpts{
		Edit: map[string]interface{}{
			"reset": map[string]interface{}{
				"elasticsearch.yml": map[string]interface{}{
					"http.cors.allow-credentials":  "",
					"http.cors.allow-origin":       "",
					"http.cors.max-age":            "",
					"http.cors.allow-headers":      "",
					"http.cors.enabled":            "",
					"http.cors.allow-methods":      "",
					"reindex.remote.whitelist":     "",
					"indices.queries.cache.size":   "",
					"thread_pool.force_merge.size": "",
				},
			},
			"delete": map[string]interface{}{
				"elasticsearch.yml": map[string]interface{}{
					"action.auto_create_index": "",
				},
			},
		},
	}
	return opts
}

func configurationWaitingForStateCompleted(ctx context.Context, d *schema.ResourceData, meta interface{}, t time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PENDING"},
		Target:  []string{"COMPLETED"},
		Refresh: func() (interface{}, string, error) {
			config := meta.(*cfg.Config)
			client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
				return config.CssV1Client(config.GetRegion(d))
			})
			if err != nil {
				return nil, "ERROR", err
			}

			cfgList, err := pc.ListTask(client, d.Id())
			if err != nil {
				if _, ok := err.(golangsdk.ErrDefault404); ok {
					return cfgList, "COMPLETED", nil
				}

				return nil, "ERROR", err
			}

			for _, task := range cfgList {
				if task.Status == "running" {
					return cfgList, "PENDING", nil
				}
			}
			return cfgList, "COMPLETED", nil
		},
		Timeout:      t,
		Delay:        30 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
