package css

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/clusters"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceCssClusterRestartV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCssClusterRestartV1Create,
		ReadContext:   resourceCssClusterRestartV1Read,
		DeleteContext: resourceCssClusterRestartV1Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCssClusterRestartV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.CssV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	clusterID := d.Get("cluster_id").(string)
	// Check whether the cluster status is available.
	secondsWait := int(math.Round(d.Timeout(schema.TimeoutCreate).Seconds()))
	err = checkClusterOperationCompleted(client, clusterID, secondsWait)
	if err != nil {
		return fmterr.Errorf("error waiting for OpenTelekomCloud CSS cluster to be ready: %s", err)
	}

	err = clusters.RestartCluster(client, clusterID)
	if err != nil {
		return diag.Errorf("error restart OpenTelekomCloud CSS cluster, err: %s", err)
	}

	// Check whether the cluster restart is complete
	err = checkClusterOperationCompleted(client, clusterID, secondsWait)
	if err != nil {
		return fmterr.Errorf("error waiting for OpenTelekomCloud CSS cluster to be ready: %s", err)
	}

	d.SetId(clusterID)

	return nil
}

func resourceCssClusterRestartV1Read(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceCssClusterRestartV1Delete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	errorMsg := "Deleting restart resource is not supported. The restart resource is only removed from the state," +
		" the cluster instance remains in the cloud."
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  errorMsg,
		},
	}
}

func checkClusterOperationCompleted(client *golangsdk.ServiceClient, id string, timeout int) error {
	return golangsdk.WaitFor(timeout, func() (bool, error) {
		cluster, err := clusters.Get(client, id)
		if err != nil {
			if _, ok := err.(golangsdk.BaseError); ok {
				return true, err
			}
			log.Printf("Error waiting for CSS cluster: %s", err)
			return false, nil
		}
		if cluster.Status != "200" {
			return false, nil
		}
		if len(cluster.Actions) > 0 {
			return false, nil
		}
		if cluster.Instances == nil {
			return false, nil
		}
		for _, v := range cluster.Instances {
			if v.Status != "200" {
				return false, nil
			}
		}
		return true, nil
	})
}
