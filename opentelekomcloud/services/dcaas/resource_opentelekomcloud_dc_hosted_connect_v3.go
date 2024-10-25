package dcaas

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	hosted_connect "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcaas/v3/hosted-connect"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceHostedConnectV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostedConnectV3Create,
		UpdateContext: resourceHostedConnectV3Update,
		ReadContext:   resourceHostedConnectV3Read,
		DeleteContext: resourceHostedConnectV3Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"hosting_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"resource_tenant_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"peer_location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHostedConnectV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	opts := hosted_connect.CreateOpts{
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		Bandwidth:        d.Get("bandwidth").(int),
		HostingID:        d.Get("hosting_id").(string),
		Vlan:             d.Get("vlan").(int),
		ResourceTenantId: d.Get("resource_tenant_id").(string),
		PeerLocation:     d.Get("peer_location").(string),
	}

	hc, err := hosted_connect.Create(client, opts)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud DC hosted connect v3: %s", err)
	}

	log.Printf("[DEBUG] Waiting for OpenTelekomCloud DC hosted connect v3 (%s) to become available", hc.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"PENDING"},
		Target:     []string{"COMPLETED"},
		Refresh:    WaitForHostedConnectActive(client, hc.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud DC hosted connect v3: %s", err)
	}
	d.SetId(hc.ID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceHostedConnectV3Read(clientCtx, d, meta)
}

func resourceHostedConnectV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	hc, err := hosted_connect.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "hosted connect")
	}

	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("name", hc.Name),
		d.Set("description", hc.Description),
		d.Set("bandwidth", hc.Bandwidth),
		d.Set("hosting_id", hc.HostingId),
		d.Set("vlan", hc.Vlan),
		d.Set("resource_tenant_id", hc.TenantID),
		d.Set("peer_location", hc.PeerLocation),
		d.Set("status", hc.Status),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceHostedConnectV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	updateHostedConnectChanges := []string{
		"name",
		"description",
		"bandwidth",
		"peer_location",
	}

	if d.HasChanges(updateHostedConnectChanges...) {
		opts := hosted_connect.UpdateOpts{
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			Bandwidth:    d.Get("bandwidth").(int),
			PeerLocation: d.Get("peer_location").(string),
		}
		_, err = hosted_connect.Update(client, d.Id(), opts)
		if err != nil {
			return diag.Errorf("error updating OpenTelekomCloud DC hosted connect v3 (%s): %s", d.Id(), err)
		}

		log.Printf("[DEBUG] Waiting for OpenTelekomCloud DC hosted connect v3 (%s) to become available", d.Id())
		stateConf := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{"COMPLETED"},
			Refresh:    WaitForHostedConnectActive(client, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud DC hosted connect v3: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceHostedConnectV3Read(clientCtx, d, meta)
}

func resourceHostedConnectV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.DCaaSV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateClientV3, err)
	}

	err = hosted_connect.Delete(client, d.Id())
	if err != nil {
		return diag.Errorf("error deleting OpenTelekomCloud DC hosted connect v3 (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for OpenTelekomCloud DC hosted connect v3 (%s) to become deleted", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    WaitForHostedConnectDelete(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error updating OpenTelekomCloud DC hosted connect v3: %s", err)
	}
	return nil
}

func WaitForHostedConnectActive(client *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := hosted_connect.Get(client, id)
		if err != nil {
			return nil, "", fmt.Errorf("error waiting for OpenTelekomCloud DC hosted connect v3 to become active: %w", err)
		}

		if n.Status == "ACTIVE" {
			return n, "COMPLETED", nil
		}
		return n, "PENDING", nil
	}
}

func WaitForHostedConnectDelete(client *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete OpenTelekomCloud DC hosted connect v3 %s.\n", id)

		r, err := hosted_connect.Get(client, id)

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[DEBUG] Successfully deleted OpenTelekomCloud DC hosted connect v3 %s", id)
				return r, "DELETED", nil
			}
			return r, "DELETING", err
		}

		log.Printf("[DEBUG] OpenTelekomCloud DC hosted connect v3 %s still available.\n", id)
		return r, r.Status, nil
	}
}
