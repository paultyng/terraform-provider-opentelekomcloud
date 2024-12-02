package cce

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceCCEClusterKubeConfigV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCCEClusterKubeConfigV3Read,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"expiry_date": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  common.ValidateRFC3339Timestamp,
				ConflictsWith: []string{"duration"},
			},
			"kubeconfig": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCCEClusterKubeConfigV3Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CceV3Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	clusterID := d.Get("cluster_id").(string)
	expiryDate := d.Get("expiry_date").(string)
	duration := -1
	if v, ok := d.GetOk("duration"); ok {
		duration = v.(int)
	}
	if expiryDate != "" {
		currentTime := time.Now()
		t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", expiryDate))
		if err != nil {
			return fmterr.Errorf("error Parsing Expiration Date: %s", err)
		}
		duration = int(t.Sub(currentTime).Hours() / 24)
	}
	expiryOpts := clusters.ExpirationOpts{
		Duration: duration,
	}

	kubeconfig, err := clusters.GetCertWithExpiration(client, clusterID, expiryOpts)
	if err != nil {
		return fmterr.Errorf("unable to retrieve cluster kubeconfig: %w", err)
	}

	d.SetId(clusterID)

	jsonStr, err := json.Marshal(kubeconfig)
	if err != nil {
		return fmterr.Errorf("unable to marshal kubeconfig: %w", err)
	}

	mErr := multierror.Append(nil,
		d.Set("cluster_id", clusterID),
		d.Set("duration", duration),
		d.Set("kubeconfig", string(jsonStr)),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
