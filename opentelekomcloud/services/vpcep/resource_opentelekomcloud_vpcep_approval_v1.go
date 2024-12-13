package vpcep

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
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vpcep/v1/services"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVPCEPApprovalV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCEndpointApprovalCreate,
		ReadContext:   resourceVPCEndpointApprovalRead,
		UpdateContext: resourceVPCEndpointApprovalUpdate,
		DeleteContext: resourceVPCEndpointApprovalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"endpoints": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"connections": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"packet_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"domain_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
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

func resourceVPCEndpointApprovalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClient, func() (*golangsdk.ServiceClient, error) {
		return config.VpcEpV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	// check status of the VPC endpoint service
	serviceID := d.Get("service_id").(string)
	n, err := services.Get(client, serviceID)
	if err != nil {
		return diag.Errorf("error retrieving VPC endpoint service %s: %s", serviceID, err)
	}
	if n.Status != "available" {
		return diag.Errorf("the status of VPC endpoint service is %s, expected to be available", n.Status)
	}

	raw := d.Get("endpoints").(*schema.Set).List()
	err = doConnectionAction(ctx, d, client, serviceID, actionReceive, raw)
	if err != nil {
		return diag.Errorf("error receiving connections to VPC endpoint service %s: %s", serviceID, err)
	}

	d.SetId(serviceID)

	clientCtx := common.CtxWithClient(ctx, client, keyClient)
	return resourceVPCEndpointApprovalRead(clientCtx, d, meta)
}

func resourceVPCEndpointApprovalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClient, func() (*golangsdk.ServiceClient, error) {
		return config.VpcEpV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	serviceID := d.Id()
	connections, err := flattenVPCEndpointConnections(client, serviceID)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "VPC endpoint service connection")
	}
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("connections", connections),
		d.Set("service_id", serviceID),
	)
	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceVPCEndpointApprovalUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClient, func() (*golangsdk.ServiceClient, error) {
		return config.VpcEpV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	if d.HasChange("endpoints") {
		oldVal, newVal := d.GetChange("endpoints")
		oldConnSet := oldVal.(*schema.Set)
		newConnSet := newVal.(*schema.Set)
		received := newConnSet.Difference(oldConnSet)
		rejected := oldConnSet.Difference(newConnSet)

		serviceID := d.Get("service_id").(string)
		err = doConnectionAction(ctx, d, client, serviceID, actionReceive, received.List())
		if err != nil {
			return diag.Errorf("error receiving connections to VPC endpoint service %s: %s", serviceID, err)
		}

		err = doConnectionAction(ctx, d, client, serviceID, actionReject, rejected.List())
		if err != nil {
			return diag.Errorf("error rejecting connections to VPC endpoint service %s: %s", serviceID, err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClient)
	return resourceVPCEndpointApprovalRead(clientCtx, d, meta)
}

func resourceVPCEndpointApprovalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClient, func() (*golangsdk.ServiceClient, error) {
		return config.VpcEpV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	serviceID := d.Get("service_id").(string)
	raw := d.Get("endpoints").(*schema.Set).List()
	err = doConnectionAction(ctx, d, client, serviceID, actionReject, raw)
	if err != nil {
		return diag.Errorf("error rejecting connections to VPC endpoint service %s: %s", serviceID, err)
	}

	return nil
}

func doConnectionAction(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient, serviceID,
	action string, raw []interface{}) error {
	if len(raw) == 0 {
		return nil
	}

	if _, ok := approvalActionStatusMap[action]; !ok {
		return fmt.Errorf("approval action(%s) is invalid, only support %s or %s", action, actionReceive,
			actionReject)
	}

	targetStatus := approvalActionStatusMap[action]
	for _, v := range raw {
		// Each request accepts or rejects only one VPC endpoint
		epID := v.(string)
		connOpts := services.ActionOpts{
			Action:    action,
			Endpoints: []string{epID},
		}

		log.Printf("[DEBUG] %s to endpoint %s from VPC endpoint service %s", action, epID, serviceID)
		_, err := services.Action(client, serviceID, connOpts)
		if err != nil {
			return fmt.Errorf("error %s to endpoint %s from VPC endpoint service %s: %s", action, epID, serviceID, err)
		}

		log.Printf("[INFO] Waiting for VPC endpoint(%s) to become %s", epID, targetStatus)
		stateConf := &resource.StateChangeConf{
			Pending:      []string{"creating", "pendingAcceptance"},
			Target:       []string{targetStatus},
			Refresh:      waitForVPCEndpointConnected(client, serviceID, epID),
			Timeout:      d.Timeout(schema.TimeoutCreate),
			Delay:        3 * time.Second,
			PollInterval: 3 * time.Second,
		}

		_, stateErr := stateConf.WaitForStateContext(ctx)
		if stateErr != nil {
			return fmt.Errorf("error waiting for VPC endpoint(%s) to become %s: %s", epID, targetStatus,
				stateErr)
		}
	}

	return nil
}

func waitForVPCEndpointConnected(client *golangsdk.ServiceClient, serviceId, endpointId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		listOpts := services.ListConnectionsOpts{
			ID: endpointId,
		}
		connections, err := services.ListConnections(client, serviceId, listOpts)
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return connections, "deleted", nil
			}
			return connections, "error", err
		}
		if len(connections) == 1 && connections[0].ID == endpointId {
			return connections, connections[0].Status, nil
		}
		return connections, "deleted", nil
	}
}
