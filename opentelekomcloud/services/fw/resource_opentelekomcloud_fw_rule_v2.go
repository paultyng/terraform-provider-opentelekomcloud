package fw

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/fwaas_v2/policies"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/fwaas_v2/rules"
	"github.com/opentelekomcloud/gophertelekomcloud/pagination"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceFWRuleV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFWRuleV2Create,
		ReadContext:   resourceFWRuleV2Read,
		UpdateContext: resourceFWRuleV2Update,
		DeleteContext: resourceFWRuleV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4,
			},
			"source_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_port": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_port": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFWRuleV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud networking client: %s", err)
	}

	enabled := d.Get("enabled").(bool)
	ipVersion := resourceFWRuleV2DetermineIPVersion(d.Get("ip_version").(int))
	protocol := resourceFWRuleV2DetermineProtocol(d.Get("protocol").(string))

	ruleConfiguration := RuleCreateOpts{
		rules.CreateOpts{
			Name:                 d.Get("name").(string),
			Description:          d.Get("description").(string),
			Protocol:             protocol,
			Action:               d.Get("action").(string),
			IPVersion:            ipVersion,
			SourceIPAddress:      d.Get("source_ip_address").(string),
			DestinationIPAddress: d.Get("destination_ip_address").(string),
			SourcePort:           d.Get("source_port").(string),
			DestinationPort:      d.Get("destination_port").(string),
			Enabled:              &enabled,
			TenantID:             d.Get("tenant_id").(string),
		},
		common.MapValueSpecs(d),
	}

	log.Printf("[DEBUG] Create firewall rule: %#v", ruleConfiguration)

	rule, err := rules.Create(networkingClient, ruleConfiguration).Extract()

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Firewall rule with id %s : %#v", rule.ID, rule)

	d.SetId(rule.ID)

	return resourceFWRuleV2Read(ctx, d, meta)
}

func resourceFWRuleV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Retrieve information about firewall rule: %s", d.Id())

	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud networking client: %s", err)
	}

	rule, err := rules.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "FW rule")
	}

	log.Printf("[DEBUG] Read OpenTelekomCloud Firewall Rule %s: %#v", d.Id(), rule)

	var protocol string
	if rule.Protocol == "" {
		protocol = "any"
	} else {
		protocol = rule.Protocol
	}

	mErr := multierror.Append(
		d.Set("action", rule.Action),
		d.Set("name", rule.Name),
		d.Set("description", rule.Description),
		d.Set("ip_version", rule.IPVersion),
		d.Set("source_ip_address", rule.SourceIPAddress),
		d.Set("destination_ip_address", rule.DestinationIPAddress),
		d.Set("source_port", rule.SourcePort),
		d.Set("destination_port", rule.DestinationPort),
		d.Set("enabled", rule.Enabled),
		d.Set("protocol", protocol),
		d.Set("region", config.GetRegion(d)),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFWRuleV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud networking client: %s", err)
	}

	var updateOpts rules.UpdateOpts
	if d.HasChange("name") {
		name := d.Get("name").(string)
		updateOpts.Name = &name
	}
	if d.HasChange("description") {
		description := d.Get("description").(string)
		updateOpts.Description = &description
	}
	if d.HasChange("destination_port") {
		destinationPort := d.Get("destination_port").(string)
		updateOpts.DestinationPort = &destinationPort
		if *updateOpts.DestinationPort == "" {
			updateOpts.DestinationPort = nil
		}
	}
	if d.HasChange("source_port") {
		sourcePort := d.Get("source_port").(string)
		updateOpts.SourcePort = &sourcePort
		if *updateOpts.SourcePort == "" {
			updateOpts.SourcePort = pointerto.String("0")
		}
	}
	if d.HasChange("protocol") {
		protocol := d.Get("protocol").(string)
		updateOpts.Protocol = &protocol
		if protocol == "icmp" {
			updateOpts.DestinationPort = nil
			updateOpts.SourcePort = nil
		}
	}
	if d.HasChange("action") {
		action := d.Get("action").(string)
		updateOpts.Action = &action
	}
	if d.HasChange("ip_version") {
		ipVersion := resourceFWRuleV2DetermineIPVersion(d.Get("ip_version").(int))
		updateOpts.IPVersion = &ipVersion
	}
	if d.HasChange("source_ip_address") {
		sourceIPAddress := d.Get("source_ip_address").(string)
		updateOpts.SourceIPAddress = &sourceIPAddress
	}
	if d.HasChange("destination_ip_address") {
		destinationIPAddress := d.Get("destination_ip_address").(string)
		updateOpts.DestinationIPAddress = &destinationIPAddress
	}
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		updateOpts.Enabled = &enabled
	}

	if d.Get("protocol").(string) != "icmp" && (updateOpts.DestinationPort == nil && updateOpts.SourcePort == nil) {
		updateOpts.DestinationPort = pointerto.String(d.Get("destination_port").(string))
		if d.Get("source_port").(string) != "" {
			updateOpts.SourcePort = pointerto.String(d.Get("source_port").(string))
		}
	}

	log.Printf("[DEBUG] Updating firewall rules: %#v", updateOpts)
	err = rules.Update(networkingClient, d.Id(), updateOpts).Err
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFWRuleV2Read(ctx, d, meta)
}

func resourceFWRuleV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Destroy firewall rule: %s", d.Id())

	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud networking client: %s", err)
	}

	rule, err := rules.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	policyID, err := assignedPolicyID(networkingClient, rule.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	if policyID != "" {
		_, err := policies.RemoveRule(networkingClient, policyID, rule.ID).Extract()
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(rules.Delete(networkingClient, d.Id()).Err)
}

func assignedPolicyID(nwClient *golangsdk.ServiceClient, ruleID string) (string, error) {
	pager := policies.List(nwClient, policies.ListOpts{})
	policyID := ""
	err := pager.EachPage(func(page pagination.Page) (b bool, err error) {
		policyList, err := policies.ExtractPolicies(page)
		if err != nil {
			return false, err
		}
		for _, policy := range policyList {
			for _, rule := range policy.Rules {
				if rule == ruleID {
					policyID = policy.ID
					return false, nil
				}
			}
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return policyID, nil
}

func resourceFWRuleV2DetermineIPVersion(ipv int) golangsdk.IPVersion {
	// Determine the IP Version
	var ipVersion golangsdk.IPVersion
	switch ipv {
	case 4:
		ipVersion = golangsdk.IPv4
	case 6:
		ipVersion = golangsdk.IPv6
	}

	return ipVersion
}

func resourceFWRuleV2DetermineProtocol(p string) rules.Protocol {
	var protocol rules.Protocol
	switch p {
	case "any":
		protocol = rules.ProtocolAny
	case "icmp":
		protocol = rules.ProtocolICMP
	case "tcp":
		protocol = rules.ProtocolTCP
	case "udp":
		protocol = rules.ProtocolUDP
	}

	return protocol
}
