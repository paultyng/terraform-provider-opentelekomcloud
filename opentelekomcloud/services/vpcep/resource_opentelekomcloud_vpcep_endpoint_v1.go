package vpcep

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vpcep/v1/endpoints"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVPCEPEndpointV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCEPEndpointCreate,
		ReadContext:   resourceVPCEPEndpointRead,
		UpdateContext: resourceVPCEPEndpointUpdate,
		DeleteContext: resourceVPCEPEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enable_dns": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags": {
				Type:         schema.TypeMap,
				Optional:     true,
				Computed:     true,
				ValidateFunc: common.ValidateTags,
			},
			"route_tables": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"port_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"enable_whitelist": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"whitelist": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"policy_statement": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					equal, _ := common.CompareJsonTemplateAreEquivalent(old, new)
					return equal
				},
			},
			"marker_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"dns_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_type": {
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

func buildPolicyStatement(d *schema.ResourceData) ([]endpoints.PolicyStatement, error) {
	if d.Get("policy_statement").(string) == "" {
		return nil, nil
	}

	var statements []endpoints.PolicyStatement
	err := json.Unmarshal([]byte(d.Get("policy_statement").(string)), &statements)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling policy, please check the format of the policy statement: %s", err)
	}
	return statements, nil
}

func resourceVPCEPEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	client, err := config.VpcEpV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClient)

	policyStatementOpts, err := buildPolicyStatement(d)
	if err != nil {
		return diag.FromErr(err)
	}

	opts := endpoints.CreateOpts{
		NetworkID:   d.Get("subnet_id").(string),
		ServiceID:   d.Get("service_id").(string),
		VpcId:       d.Get("vpc_id").(string),
		PortIP:      d.Get("port_ip").(string),
		EnableDNS:   d.Get("enable_dns").(bool),
		Description: d.Get("description").(string),
		Tags: common.ExpandResourceTags(
			d.Get("tags").(map[string]interface{}),
		),
		RouteTables: common.ExpandToStringSlice(
			d.Get("route_tables").(*schema.Set).List(),
		),
		Whitelist: common.ExpandToStringSlice(
			d.Get("whitelist").(*schema.Set).List(),
		),
		PolicyStatement: policyStatementOpts,
	}
	if v, ok := d.GetOk("enable_whitelist"); ok {
		enable := v.(bool)
		opts.EnableWhitelist = &enable
	}
	created, err := endpoints.Create(client, opts)
	if err != nil {
		return fmterr.Errorf("error creating VPC Endpoint: %w", err)
	}
	d.SetId(created.ID)

	stateConf := &resource.StateChangeConf{
		Pending: []string{string(endpoints.StatusCreating)},
		Target:  []string{string(endpoints.StatusAccepted), string(endpoints.StatusPendingAcceptance)},
		Refresh: refreshVPCEndpoint(client, created.ID),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for VPC endpoint to be created: %w", err)
	}

	return resourceVPCEPEndpointRead(clientCtx, d, meta)
}

func refreshVPCEndpoint(client *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ep, err := endpoints.Get(client, id)
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return nil, "", nil
			}
			return nil, "", err
		}
		return ep, string(ep.Status), nil
	}
}

func resourceVPCEPEndpointRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	client, err := config.VpcEpV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	endpoint, err := endpoints.Get(client, d.Id())
	if err != nil {
		return fmterr.Errorf("error getting VPC Endpoint: %w", err)
	}

	policyStatements, err := json.Marshal(endpoint.PolicyStatement)
	if err != nil {
		return diag.Errorf("error marshaling policy statement: %s", err)
	}

	mErr := multierror.Append(
		d.Set("service_id", endpoint.ServiceID),
		d.Set("service_name", onlyServiceName(endpoint.ServiceName)),
		d.Set("service_type", endpoint.ServiceType),
		d.Set("project_id", endpoint.ProjectID),
		d.Set("enable_dns", endpoint.EnableDNS),
		d.Set("dns_names", endpoint.DNSNames),
		d.Set("port_ip", endpoint.IP),
		d.Set("enable_whitelist", endpoint.EnableWhitelist),
		d.Set("whitelist", endpoint.Whitelist),
		d.Set("vpc_id", endpoint.VpcID),
		d.Set("subnet_id", endpoint.NetworkID),
		d.Set("marker_id", endpoint.MarkerID),
		d.Set("tags", common.TagsToMap(endpoint.Tags)),
		d.Set("policy_statement", string(policyStatements)),
		d.Set("description", endpoint.Description),
		d.Set("status", endpoint.Status),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting VPC endpoint fields: %w", err)
	}

	return nil
}

func resourceVPCEPEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.VpcEpV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	if d.HasChange("tags") {
		tagErr := common.UpdateResourceTags(client, d, "endpoint", d.Id())
		if tagErr != nil {
			return diag.Errorf("error updating tags of VPC endpoint %s: %s", d.Id(), tagErr)
		}
	}

	return resourceVPCEPEndpointRead(ctx, d, meta)
}

func resourceVPCEPEndpointDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	client, err := config.VpcEpV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(ErrClientCreate, err)
	}

	err = endpoints.Delete(client, d.Id())
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			return nil
		}
		return fmterr.Errorf("error deleting VPC endpoint: %w", err)
	}
	err = endpoints.WaitForEndpointStatus(
		client, d.Id(), "", timeoutSeconds(d, schema.TimeoutDelete),
	)
	if err != nil {
		return fmterr.Errorf("error waiting for VPC EP endpoint to become deleted: %w", err)
	}

	return nil
}
