package dns

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/zones"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

var serviceMap = map[string]string{
	"public":  "DNS-public_zone",
	"private": "DNS-private_zone",
}

func ResourceDNSZoneV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSZoneV2Create,
		ReadContext:   resourceDNSZoneV2Read,
		UpdateContext: resourceDNSZoneV2Update,
		DeleteContext: resourceDNSZoneV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDnsZoneV2ImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: common.SuppressEqualZoneNames,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "public",
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": common.TagsSchema(),
			"router": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"router_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"router_region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"masters": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDNSRouter(d *schema.ResourceData) map[string]string {
	router := d.Get("router").(*schema.Set).List()

	if len(router) > 0 {
		mp := make(map[string]string)
		c := router[0].(map[string]interface{})

		if val, ok := c["router_id"]; ok {
			mp["router_id"] = val.(string)
		}
		if val, ok := c["router_region"]; ok {
			mp["router_region"] = val.(string)
		}
		return mp
	}
	return nil
}

func resourceDNSZoneV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(region)
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}
	zone_type := d.Get("type").(string)
	router := d.Get("router").(*schema.Set).List()

	// Dirty hack for nl region without it impossible to create public zone
	nlClientOverride(d, region, client)

	// router is required when creating private zone
	if zone_type == "private" {
		if len(router) < 1 {
			return fmterr.Errorf("the argument (router) is required when creating OpenTelekomCloud DNS private zone")
		}
	}
	vs := common.MapResourceProp(d, "value_specs")
	// Add zone_type to the list.  We do this to keep GopherCloud OpenTelekomCloud standard.
	vs["zone_type"] = zone_type
	vs["router"] = resourceDNSRouter(d)

	createOpts := ZoneCreateOpts{
		zones.CreateOpts{
			Name:        d.Get("name").(string),
			TTL:         d.Get("ttl").(int),
			Email:       d.Get("email").(string),
			Description: d.Get("description").(string),
		},
		vs,
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	n, err := zones.Create(client, createOpts).Extract()
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud DNS zone: %s", logHttpError(err))
	}

	log.Printf("[DEBUG] Waiting for DNS Zone (%s) to become available", n.ID)
	stateConf := &resource.StateChangeConf{
		Target:       []string{"ACTIVE"},
		Pending:      []string{"PENDING"},
		Refresh:      waitForDNSZone(client, n.ID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        5 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf(
			"Error waiting for DNS Zone (%s) to become ACTIVE: %s",
			n.ID, err)
	}

	// router length >1 when creating private zone
	if zone_type == "private" {
		// AssociateZone for the other routers
		routerList := getDNSRouters(d)
		if len(routerList) > 1 {
			for i := range routerList {
				// Skip the first router
				if i > 0 {
					log.Printf("[DEBUG] Creating AssociateZone Options: %#v", routerList[i])
					_, err := zones.AssociateZone(client, n.ID, routerList[i]).Extract()
					if err != nil {
						return fmterr.Errorf("error AssociateZone: %s", err)
					}

					log.Printf("[DEBUG] Waiting for AssociateZone (%s) to Router (%s) become ACTIVE",
						n.ID, routerList[i].RouterID)
					stateRouterConf := &resource.StateChangeConf{
						Target:       []string{"ACTIVE"},
						Pending:      []string{"PENDING"},
						Refresh:      waitForDNSZoneRouter(client, n.ID, routerList[i].RouterID),
						Timeout:      d.Timeout(schema.TimeoutCreate),
						Delay:        5 * time.Second,
						MinTimeout:   3 * time.Second,
						PollInterval: 2 * time.Second,
					}

					_, err = stateRouterConf.WaitForStateContext(ctx)
					if err != nil {
						return fmterr.Errorf("error waiting for AssociateZone (%s) to Router (%s) become ACTIVE: %s",
							n.ID, routerList[i].RouterID, err)
					}
				} else {
					log.Printf("[DEBUG] First Router Options: %#v", routerList[i])
				}
			}
		}
	}

	d.SetId(n.ID)

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		taglist := common.ExpandResourceTags(tagRaw)
		if tagErr := tags.Create(client, serviceMap[zone_type], n.ID, taglist).ExtractErr(); tagErr != nil {
			return fmterr.Errorf("error setting tags of DNS zone %s: %s", n.ID, tagErr)
		}
	}

	log.Printf("[DEBUG] Created OpenTelekomCloud DNS Zone %s: %#v", n.ID, n)
	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceDNSZoneV2Read(clientCtx, d, meta)
}

func resourceDNSZoneV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(region)
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	// Dirty hack for nl region without it impossible to create public zone
	nlClientOverride(d, region, client)

	n, err := zones.Get(client, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "zone")
	}

	log.Printf("[DEBUG] Retrieved Zone %s: %#v", d.Id(), n)

	mErr := multierror.Append(
		d.Set("name", n.Name),
		d.Set("email", n.Email),
		d.Set("description", n.Description),
		d.Set("ttl", n.TTL),
		d.Set("type", n.ZoneType),
		d.Set("region", config.GetRegion(d)),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("masters", n.Masters); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving masters to state for OpenTelekomCloud DNS zone (%s): %s", d.Id(), err)
	}

	// save tags
	resourceTags, err := tags.Get(client, serviceMap[n.ZoneType], d.Id()).Extract()
	if err != nil {
		return fmterr.Errorf("error fetching OpenTelekomCloud DNS zone tags: %s", err)
	}

	tagmap := common.TagsToMap(resourceTags)
	if err := d.Set("tags", tagmap); err != nil {
		return fmterr.Errorf("error saving tags for OpenTelekomCloud DNS zone %s: %s", d.Id(), err)
	}

	return nil
}

func resourceDNSZoneV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(region)
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}
	zone_type := d.Get("type").(string)
	router := d.Get("router").(*schema.Set).List()

	// Dirty hack for nl region without it impossible to create public zone
	nlClientOverride(d, region, client)

	// router is required when updating private zone
	if zone_type == "private" {
		if len(router) < 1 {
			return fmterr.Errorf("the argument (router) is required when updating OpenTelekomCloud DNS private zone")
		}
	}

	var updateOpts zones.UpdateOpts
	if d.HasChange("email") {
		updateOpts.Email = d.Get("email").(string)
	}
	if d.HasChange("ttl") {
		updateOpts.TTL = d.Get("ttl").(int)
	}
	if d.HasChange("description") {
		updateOpts.Description = d.Get("description").(string)
	}

	log.Printf("[DEBUG] Updating Zone %s with options: %#v", d.Id(), updateOpts)

	_, err = zones.Update(client, d.Id(), updateOpts).Extract()
	if err != nil {
		return fmterr.Errorf("error updating OpenTelekomCloud DNS Zone: %s", logHttpError(err))
	}

	log.Printf("[DEBUG] Waiting for DNS Zone (%s) to update", d.Id())
	stateConf := &resource.StateChangeConf{
		Target:       []string{"ACTIVE"},
		Pending:      []string{"PENDING"},
		Refresh:      waitForDNSZone(client, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        5 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for DNS zone to be active: %w", err)
	}

	if d.HasChange("router") {
		// when updating private zone
		if zone_type == "private" {
			associateList, disassociateList, err := resourceGetDNSRouters(client, d)
			if err != nil {
				return fmterr.Errorf("error getting OpenTelekomCloud DNS Zone Router: %s", err)
			}
			if len(associateList) > 0 {
				// AssociateZone
				for i := range associateList {
					log.Printf("[DEBUG] Updating AssociateZone Options: %#v", associateList[i])
					_, err := zones.AssociateZone(client, d.Id(), associateList[i]).Extract()
					if err != nil {
						return fmterr.Errorf("error AssociateZone: %s", err)
					}

					log.Printf("[DEBUG] Waiting for AssociateZone (%s) to Router (%s) become ACTIVE",
						d.Id(), associateList[i].RouterID)
					stateRouterConf := &resource.StateChangeConf{
						Target:       []string{"ACTIVE"},
						Pending:      []string{"PENDING"},
						Refresh:      waitForDNSZoneRouter(client, d.Id(), associateList[i].RouterID),
						Timeout:      d.Timeout(schema.TimeoutUpdate),
						Delay:        5 * time.Second,
						MinTimeout:   3 * time.Second,
						PollInterval: 2 * time.Second,
					}

					_, err = stateRouterConf.WaitForStateContext(ctx)
					if err != nil {
						return fmterr.Errorf("error waiting for AssociateZone (%s) to Router (%s) become ACTIVE: %s",
							d.Id(), associateList[i].RouterID, err)
					}
				}
			}
			if len(disassociateList) > 0 {
				// DisassociateZone
				for j := range disassociateList {
					log.Printf("[DEBUG] Updating DisassociateZone Options: %#v", disassociateList[j])
					_, err := zones.DisassociateZone(client, d.Id(), disassociateList[j]).Extract()
					if err != nil {
						return fmterr.Errorf("error DisassociateZone: %s", err)
					}

					log.Printf("[DEBUG] Waiting for DisassociateZone (%s) to Router (%s) become DELETED",
						d.Id(), disassociateList[j].RouterID)
					stateRouterConf := &resource.StateChangeConf{
						Target:       []string{"DELETED"},
						Pending:      []string{"ACTIVE", "PENDING", "ERROR"},
						Refresh:      waitForDNSZoneRouter(client, d.Id(), disassociateList[j].RouterID),
						Timeout:      d.Timeout(schema.TimeoutUpdate),
						Delay:        5 * time.Second,
						MinTimeout:   3 * time.Second,
						PollInterval: 2 * time.Second,
					}

					_, err = stateRouterConf.WaitForStateContext(ctx)
					if err != nil {
						return fmterr.Errorf("error waiting for DisassociateZone (%s) to Router (%s) become DELETED: %s",
							d.Id(), disassociateList[j].RouterID, err)
					}
				}
			}
		}
	}

	// update tags
	tagErr := common.UpdateResourceTags(client, d, serviceMap[zone_type], d.Id())
	if tagErr != nil {
		return fmterr.Errorf("error updating tags of DNS zone %s: %s", d.Id(), tagErr)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceDNSZoneV2Read(clientCtx, d, meta)
}

func resourceDNSZoneV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(region)
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	// Dirty hack for nl region without it impossible to create public zone
	nlClientOverride(d, region, client)

	_, err = zones.Delete(client, d.Id()).Extract()
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud DNS Zone: %s", err)
	}

	log.Printf("[DEBUG] Waiting for DNS Zone (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Target: []string{"DELETED"},
		// we allow to try to delete ERROR zone
		Pending:      []string{"ACTIVE", "PENDING", "ERROR"},
		Refresh:      waitForDNSZone(client, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        5 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf(
			"Error waiting for DNS Zone (%s) to delete: %s",
			d.Id(), err)
	}

	d.SetId("")
	return nil
}

func nlClientOverride(d *schema.ResourceData, region string, client *golangsdk.ServiceClient) {
	if d.Get("type").(string) == "public" && region == "eu-nl" {
		client.Endpoint = strings.Replace(client.Endpoint, "eu-nl", "eu-de", 1)
		client.ResourceBase = strings.Replace(client.ResourceBase, "eu-nl", "eu-de", 1)
	}
}

func parseStatus(rawStatus string) string {
	// rawStatus maybe one of PENDING_CREATE, PENDING_UPDATE, PENDING_DELETE, ACTIVE, or ERROR
	splits := strings.Split(rawStatus, "_")
	return splits[0]
}

func waitForDNSZone(client *golangsdk.ServiceClient, zoneId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		zone, err := zones.Get(client, zoneId).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return zone, "DELETED", nil
			}

			return nil, "", err
		}

		log.Printf("[DEBUG] OpenTelekomCloud DNS Zone (%s) current status: %s", zone.ID, zone.Status)
		return zone, parseStatus(zone.Status), nil
	}
}

func getDNSRouters(d *schema.ResourceData) []zones.RouterOpts {
	router := d.Get("router").(*schema.Set).List()
	if len(router) > 0 {
		res := make([]zones.RouterOpts, len(router))
		for i := range router {
			ro := zones.RouterOpts{}
			c := router[i].(map[string]interface{})
			if val, ok := c["router_id"]; ok {
				ro.RouterID = val.(string)
			}
			if val, ok := c["router_region"]; ok {
				ro.RouterRegion = val.(string)
			}
			res[i] = ro
		}
		return res
	}
	return nil
}

func waitForDNSZoneRouter(client *golangsdk.ServiceClient, zoneId string, routerId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		zone, err := zones.Get(client, zoneId).Extract()
		if err != nil {
			return nil, "", err
		}
		for i := range zone.Routers {
			if routerId == zone.Routers[i].RouterID {
				log.Printf("[DEBUG] OpenTelekomCloud DNS Zone (%s) Router (%s) current status: %s",
					zoneId, routerId, zone.Routers[i].Status)
				return zone, parseStatus(zone.Routers[i].Status), nil
			}
		}
		return zone, "DELETED", nil
	}
}

func resourceGetDNSRouters(client *golangsdk.ServiceClient, d *schema.ResourceData) ([]zones.RouterOpts, []zones.RouterOpts, error) {
	// get zone info from api
	n, err := zones.Get(client, d.Id()).Extract()
	if err != nil {
		return nil, nil, common.CheckDeleted(d, err, "zone")
	}
	// get routers from local
	localRouters := getDNSRouters(d)

	// get associateMap
	associateMap := make(map[string]zones.RouterOpts)
	for _, local := range localRouters {
		// Check if local is found in api
		found := false
		for _, raw := range n.Routers {
			if local.RouterID == raw.RouterID {
				found = true
				break
			}
		}
		// If local is not found in api
		if !found {
			associateMap[local.RouterID] = local
		}
	}

	// convert associateMap to associateList
	associateList := make([]zones.RouterOpts, len(associateMap))
	var i = 0
	for _, associateRouter := range associateMap {
		associateList[i] = associateRouter
		i++
	}

	// get disassociateMap
	disassociateMap := make(map[string]zones.RouterOpts)
	for _, raw := range n.Routers {
		// Check if api is found in local
		found := false
		for _, local := range localRouters {
			if raw.RouterID == local.RouterID {
				found = true
				break
			}
		}
		// If api is not found in local
		if !found {
			disassociateMap[raw.RouterID] = zones.RouterOpts{
				RouterID:     raw.RouterID,
				RouterRegion: raw.RouterRegion,
			}
		}
	}

	// convert disassociateMap to disassociateList
	disassociateList := make([]zones.RouterOpts, len(disassociateMap))
	var j = 0
	for _, disassociateRouter := range disassociateMap {
		disassociateList[j] = disassociateRouter
		j++
	}

	return associateList, disassociateList, nil
}

func logHttpError(err error) error {
	if httpErr, ok := err.(golangsdk.ErrDefault500); ok {
		return fmt.Errorf("%s\n %s", httpErr.Error(), httpErr.Body)
	}
	return err
}

func resourceDnsZoneV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(region)
	})
	if err != nil {
		return []*schema.ResourceData{d}, fmt.Errorf(errCreationClient, err)
	}
	_, err = zones.Get(client, d.Id()).Extract()
	if err != nil {
		// Check if the region is "eu-nl" and attempt to retry with "eu-de"
		if region == "eu-nl" {
			client.Endpoint = strings.Replace(client.Endpoint, "eu-nl", "eu-de", 1)
			client.ResourceBase = strings.Replace(client.ResourceBase, "eu-nl", "eu-de", 1)
			zoneId := d.Id()
			n, err := zones.Get(client, zoneId).Extract()
			if err != nil {
				return []*schema.ResourceData{d}, fmt.Errorf("error getting OpenTelekomCloud DNS zone from server: %s", err)
			}
			mErr := multierror.Append(
				d.Set("name", n.Name),
				d.Set("email", n.Email),
				d.Set("description", n.Description),
				d.Set("ttl", n.TTL),
				d.Set("type", n.ZoneType),
				d.Set("region", config.GetRegion(d)),
			)
			if err := mErr.ErrorOrNil(); err != nil {
				return []*schema.ResourceData{d}, fmt.Errorf("error saving OpenTelekomCloud DNS attributes to state: %s", err)
			}
			if err = d.Set("masters", n.Masters); err != nil {
				return []*schema.ResourceData{d}, fmt.Errorf("error saving masters to state for OpenTelekomCloud DNS zone (%s): %s", d.Id(), err)
			}
			resourceTags, err := tags.Get(client, serviceMap[n.ZoneType], d.Id()).Extract()
			if err != nil {
				return []*schema.ResourceData{d}, fmt.Errorf("error fetching OpenTelekomCloud DNS zone tags: %s", err)
			}
			tagmap := common.TagsToMap(resourceTags)
			if err := d.Set("tags", tagmap); err != nil {
				return []*schema.ResourceData{d}, fmt.Errorf("error saving tags for OpenTelekomCloud DNS zone %s: %s", d.Id(), err)
			}
			return []*schema.ResourceData{d}, nil
		}
		return []*schema.ResourceData{d}, fmt.Errorf("error getting OpenTelekomCloud DNS zone from server: %s", err)
	}

	return []*schema.ResourceData{d}, nil
}
