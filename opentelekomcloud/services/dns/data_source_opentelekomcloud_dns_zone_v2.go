package dns

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/zones"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceDNSZoneV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSZoneV2Read,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"serial": {
				Type:     schema.TypeInt,
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
			"masters": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"links": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": common.TagsSchema(),
		},
	}
}

func dataSourceDNSZoneV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.DnsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	listOpts := zones.ListOpts{}

	if v, ok := d.GetOk("name"); ok {
		listOpts.Name = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		listOpts.Description = v.(string)
	}

	if v, ok := d.GetOk("email"); ok {
		listOpts.Email = v.(string)
	}

	if v, ok := d.GetOk("status"); ok {
		listOpts.Status = v.(string)
	}

	if v, ok := d.GetOk("ttl"); ok {
		listOpts.TTL = v.(int)
	}

	if v, ok := d.GetOk("zone_type"); ok {
		listOpts.Type = v.(string)
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.(map[string]interface{})
		listOpts.Tags = common.ExpandResourceTags(tags)
	}

	// Initial attempt to retrieve zones
	pages, err := zones.List(client, listOpts).AllPages()
	if err != nil {
		return fmterr.Errorf("unable to retrieve zones: %w", err)
	}

	allZones, err := zones.ExtractZones(pages)
	if err != nil {
		return fmterr.Errorf("unable to extract zones: %w", err)
	}
	if allZones == nil {
		// Check if the region is "eu-nl" and retry with "eu-de"
		if config.GetRegion(d) == "eu-nl" {
			replaceNlEndpoint(client)
			// Retry the request after updating the endpoint
			pages, err = zones.List(client, listOpts).AllPages()
			if err != nil {
				return fmterr.Errorf("unable to retrieve zones: %w", err)
			}
			allZones, err = zones.ExtractZones(pages)
			if err != nil {
				return fmterr.Errorf("unable to extract zones: %w", err)
			}
		}
	}

	if len(allZones) < 1 {
		return fmterr.Errorf("your query returned no results. " +
			"Please change your search criteria and try again")
	}

	if len(allZones) > 1 {
		return fmterr.Errorf("your query returned more than one result. " +
			"Please try a more specific search criteria")
	}

	zone := allZones[0]

	log.Printf("[DEBUG] Retrieved DNS Zone %s: %+v", zone.ID, zone)
	d.SetId(zone.ID)

	mErr := multierror.Append(
		d.Set("name", zone.Name),
		d.Set("pool_id", zone.PoolID),
		d.Set("email", zone.Email),
		d.Set("description", zone.Description),
		d.Set("status", zone.Status),
		d.Set("zone_type", zone.ZoneType),
		d.Set("ttl", zone.TTL),
		d.Set("serial", zone.Serial),
		d.Set("created_at", zone.CreatedAt),
		d.Set("updated_at", zone.UpdatedAt),
		d.Set("links", zone.Links),
		d.Set("masters", zone.Masters),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
