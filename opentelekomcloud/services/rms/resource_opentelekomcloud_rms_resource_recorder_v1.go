package rms

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/recorder"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceRmsResourceRecorderV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecorderUpdate,
		UpdateContext: resourceRecorderUpdate,
		ReadContext:   resourceRecorderRead,
		DeleteContext: resourceRecorderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agency_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"selector": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_supported": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"resource_types": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
					},
				},
				Required: true,
			},
			"obs_channel": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bucket_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Optional:     true,
				AtLeastOneOf: []string{"smn_channel", "obs_channel"},
			},
			"smn_channel": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topic_urn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Optional:     true,
				AtLeastOneOf: []string{"smn_channel", "obs_channel"},
			},
			"retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceRecorderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	domainID := client.DomainID

	if domainID == "" {
		domainID = config.DomainClient.AKSKAuthOptions.DomainID
	}

	resTypesRaw := d.Get("selector.0.resource_types").(*schema.Set).List()

	resTypes := make([]string, len(resTypesRaw))

	for i, v := range resTypesRaw {
		resTypes[i] = v.(string)
	}

	updateOpts := recorder.UpdateOpts{
		DomainId:   domainID,
		AgencyName: d.Get("agency_name").(string),
		Channel:    recorder.ChannelConfigBody{},
		Selector: recorder.SelectorConfigBody{
			AllSupported:  d.Get("selector.0.all_supported").(bool),
			ResourceTypes: resTypes,
		},
	}

	if bucketRaw, ok := d.GetOk("obs_channel.0.bucket"); ok && bucketRaw != nil {
		updateOpts.Channel.Obs = &recorder.TrackerObsConfigBody{
			BucketName: d.Get("obs_channel.0.bucket").(string),
			RegionId:   d.Get("obs_channel.0.region").(string),
		}
		if prefixRaw, ok := d.GetOk("obs_channel.0.bucket_prefix"); ok {
			prefix := prefixRaw.(string)
			updateOpts.Channel.Obs.BucketPrefix = &prefix
		}
	}

	if topicRaw, ok := d.GetOk("smn_channel.0.topic_urn"); ok && topicRaw != nil {
		projectId := client.ProjectID
		if projectId == "" {
			projectId = config.TenantID
		}
		updateOpts.Channel.Smn = &recorder.TrackerSMNConfigBody{
			ProjectId: projectId,
			RegionId:  config.GetRegion(d),
			TopicUrn:  d.Get("smn_channel.0.topic_urn").(string),
		}
	}

	log.Printf("[DEBUG] the RMS recorder request options: %#v", updateOpts)

	err = recorder.UpdateRecorder(client, updateOpts)
	if err != nil {
		return diag.Errorf("error creating or updating RMS recorder: %s", err)
	}

	d.SetId(domainID)

	clientCtx := common.CtxWithClient(ctx, client, errCreationRMSV1Client)
	return resourceRecorderRead(clientCtx, d, meta)
}

func resourceRecorderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	resp, err := recorder.GetRecorder(client, d.Id())

	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving RMS recorder")
	}

	mErr := multierror.Append(nil,
		d.Set("agency_name", resp.AgencyName),
		d.Set("retention_period", resp.RetentionPeriod),
		d.Set("selector", flattenSelector(resp.Selector)),
		d.Set("obs_channel", flattenObs(resp.Channel)),
		d.Set("smn_channel", flattenSmn(resp.Channel)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenSelector(resp recorder.SelectorConfigBody) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"all_supported":  resp.AllSupported,
			"resource_types": resp.ResourceTypes,
		},
	}
}

func flattenObs(resp recorder.ChannelConfigBody) []interface{} {
	if resp.Obs == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"bucket":        resp.Obs.BucketName,
			"region":        resp.Obs.RegionId,
			"bucket_prefix": resp.Obs.BucketPrefix,
		},
	}
}

func flattenSmn(resp recorder.ChannelConfigBody) []interface{} {
	if resp.Smn == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"topic_urn":  resp.Smn.TopicUrn,
			"region":     resp.Smn.RegionId,
			"project_id": resp.Smn.ProjectId,
		},
	}
}

func resourceRecorderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, rmsClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.RmsV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationRMSV1Client, err)
	}

	err = recorder.DeleteRecorder(client, d.Id())
	if err != nil {
		return diag.Errorf("error deleting RMS recorder: %s", err)
	}

	return nil
}
