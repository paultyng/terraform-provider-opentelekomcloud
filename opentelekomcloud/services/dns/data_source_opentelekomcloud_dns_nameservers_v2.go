package dns

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/nameservers"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceDNSNameserversV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSNameserverRead,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"nameservers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDNSNameserverRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DnsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	var allNameservers []nameservers.Nameserver

	if v, ok := d.GetOk("zone_id"); ok {
		allNameservers, err = nameservers.List(client, v.(string)).Extract()
		if err != nil {
			// Check if the region is "eu-nl" and retry with "eu-de"
			if config.GetRegion(d) == "eu-nl" {
				replaceNlEndpoint(client)
				// Retry the request after updating the endpoint
				allNameservers, err = nameservers.List(client, v.(string)).Extract()
				if err != nil {
					return fmterr.Errorf("unable to retrieve nameservers: %w", err)
				}
			} else {
				return fmterr.Errorf("unable to retrieve nameservers: %w", err)
			}
		}
	}

	if len(allNameservers) < 1 {
		return common.DataSourceTooFewDiag
	}

	var nameserverList []map[string]interface{}
	for _, n := range allNameservers {
		nameserverEntry := map[string]interface{}{
			"hostname": n.Hostname,
			"priority": n.Priority,
		}
		nameserverList = append(nameserverList, nameserverEntry)
	}

	listID := nameserverList[0]["hostname"].(string)
	d.SetId(listID)

	mErr := multierror.Append(
		d.Set("nameservers", nameserverList),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func replaceNlEndpoint(client *golangsdk.ServiceClient) {
	client.Endpoint = strings.Replace(client.Endpoint, "eu-nl", "eu-de", 1)
	client.ResourceBase = strings.Replace(client.ResourceBase, "eu-nl", "eu-de", 1)
}
