package dns

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/recordsets"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/zones"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDNSRecordSetV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSRecordSetV2Create,
		ReadContext:   resourceDNSRecordSetV2Read,
		UpdateContext: resourceDNSRecordSetV2Update,
		DeleteContext: resourceDNSRecordSetV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: common.ImportAsManaged,
		},

		CustomizeDiff: useSharedRecordSet,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: common.SuppressEqualZoneNames,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"records": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  300,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"tags": common.TagsSchema(),

			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func getRecordSetCreateOpts(d cfg.SchemaOrDiff) RecordSetCreateOpts {
	recordSetType := d.Get("type").(string)
	recordsRaw := d.Get("records").(*schema.Set).List()
	records := make([]string, len(recordsRaw))
	if recordSetType == "TXT" {
		for i, record := range recordsRaw {
			records[i] = fmt.Sprintf("\"%s\"", record.(string))
		}
	} else {
		for i, record := range recordsRaw {
			records[i] = record.(string)
		}
	}

	return RecordSetCreateOpts{
		recordsets.CreateOpts{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Records:     records,
			TTL:         d.Get("ttl").(int),
			Type:        recordSetType,
		},
		common.MapValueSpecs(d),
	}
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func resourceDNSRecordSetV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	zoneID := d.Get("zone_id").(string)

	shared := d.Get("shared").(bool)
	if shared {
		log.Printf("[DEBUG] Using non-managed DNS record set, skipping creation")
		id, _ := getExistingRecordSetID(d, meta)
		d.SetId(fmt.Sprintf("%s/%s", zoneID, id))
		clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
		return resourceDNSRecordSetV2Read(clientCtx, d, meta)
	}

	createOpts := getRecordSetCreateOpts(d)

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	recordSet, err := recordsets.Create(client, zoneID, createOpts).Extract()
	if err != nil {
		// Check if the region is "eu-nl" and retry with "eu-de"
		if config.GetRegion(d) == "eu-nl" {
			replaceNlEndpoint(client)
			// Retry the request after updating the endpoint
			recordSet, err = recordsets.Create(client, zoneID, createOpts).Extract()
			if err != nil {
				return fmterr.Errorf("error creating OpenTelekomCloud DNS record set: %s", err)
			}
		} else {
			return fmterr.Errorf("error creating OpenTelekomCloud DNS record set: %s", err)
		}
	}

	log.Printf("[DEBUG] Waiting for DNS record set (%s) to become available", recordSet.ID)
	stateConf := &resource.StateChangeConf{
		Target:       []string{"ACTIVE"},
		Pending:      []string{"PENDING"},
		Refresh:      waitForDNSRecordSet(client, zoneID, recordSet.ID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        5 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf(
			"error waiting for record set (%s) to become ACTIVE for creation: %s",
			recordSet.ID, err)
	}

	id := fmt.Sprintf("%s/%s", zoneID, recordSet.ID)
	d.SetId(id)

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		resourceType, err := getDNSRecordSetResourceType(client, zoneID)
		if err != nil {
			return fmterr.Errorf("error getting resource type of DNS record set %s: %s", recordSet.ID, err)
		}

		tagList := common.ExpandResourceTags(tagRaw)
		if tagErr := tags.Create(client, resourceType, recordSet.ID, tagList).ExtractErr(); tagErr != nil {
			return fmterr.Errorf("error setting tags of DNS record set %s: %s", recordSet.ID, tagErr)
		}
	}

	log.Printf("[DEBUG] Created OpenTelekomCloud DNS record set %s: %#v", recordSet.ID, recordSet)
	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceDNSRecordSetV2Read(clientCtx, d, meta)
}

func resourceDNSRecordSetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	// Obtain relevant info from parsing the ID
	zoneID, recordsetID, err := ParseDNSV2RecordSetID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	n, err := recordsets.Get(client, zoneID, recordsetID).Extract()
	if err != nil {
		// Check if the region is "eu-nl" and retry with "eu-de"
		if config.GetRegion(d) == "eu-nl" {
			replaceNlEndpoint(client)
			// Retry the request after updating the endpoint
			n, err = recordsets.Get(client, zoneID, recordsetID).Extract()
			if err != nil {
				return common.CheckDeletedDiag(d, err, "record_set")
			}
		} else {
			return common.CheckDeletedDiag(d, err, "record_set")
		}
	}

	records := make([]string, len(n.Records))
	if n.Type == "TXT" {
		for i, record := range n.Records {
			records[i] = trimQuotes(record)
		}
	} else {
		copy(records, n.Records)
	}

	log.Printf("[DEBUG] Retrieved  record set %s: %#v", recordsetID, n)

	mErr := multierror.Append(
		d.Set("name", n.Name),
		d.Set("description", n.Description),
		d.Set("ttl", n.TTL),
		d.Set("type", n.Type),
		d.Set("records", records),
		d.Set("region", config.GetRegion(d)),
		d.Set("zone_id", zoneID),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf(
			"error saving records to state for OpenTelekomCloud DNS record set (%s): %s", d.Id(), err)
	}

	// save tags
	resourceType, err := getDNSRecordSetResourceType(client, zoneID)
	if err != nil {
		return fmterr.Errorf("error getting resource type of DNS record set %s: %s", recordsetID, err)
	}
	resourceTags, err := tags.Get(client, resourceType, recordsetID).Extract()
	if err != nil {
		return fmterr.Errorf("error fetching OpenTelekomCloud DNS record set tags: %s", err)
	}

	tagmap := common.TagsToMap(resourceTags)
	if err := d.Set("tags", tagmap); err != nil {
		return fmterr.Errorf("error saving tags for OpenTelekomCloud DNS record set %s: %s", recordsetID, err)
	}

	return nil
}

func resourceDNSRecordSetV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	var updateOpts recordsets.UpdateOpts
	if d.HasChange("ttl") {
		updateOpts.TTL = d.Get("ttl").(int)
	}

	// `records` is required attribute for update request
	recordsRaw := d.Get("records").(*schema.Set).List()
	records := make([]string, len(recordsRaw))
	if d.Get("type").(string) == "TXT" {
		for i, record := range recordsRaw {
			records[i] = fmt.Sprintf("\"%s\"", record.(string))
		}
	} else {
		for i, record := range recordsRaw {
			records[i] = record.(string)
		}
	}
	updateOpts.Records = records

	if d.HasChange("description") {
		updateOpts.Description = d.Get("description").(string)
	}

	// Obtain relevant info from parsing the ID
	zoneID, recordsetID, err := ParseDNSV2RecordSetID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Updating  record set %s with options: %#v", recordsetID, updateOpts)

	_, err = recordsets.Update(client, zoneID, recordsetID, updateOpts).Extract()
	if err != nil {
		// Check if the region is "eu-nl" and retry with "eu-de"
		if config.GetRegion(d) == "eu-nl" {
			replaceNlEndpoint(client)
			// Retry the request after updating the endpoint
			_, err = recordsets.Update(client, zoneID, recordsetID, updateOpts).Extract()
			if err != nil {
				return fmterr.Errorf("error updating OpenTelekomCloud DNS record set: %s", err)
			}
		} else {
			return fmterr.Errorf("error updating OpenTelekomCloud DNS record set: %s", err)
		}
	}

	log.Printf("[DEBUG] Waiting for DNS record set (%s) to update", recordsetID)
	stateConf := &resource.StateChangeConf{
		Target:       []string{"ACTIVE"},
		Pending:      []string{"PENDING"},
		Refresh:      waitForDNSRecordSet(client, zoneID, recordsetID),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        5 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf(
			"error waiting for record set (%s) to become ACTIVE for updation: %s",
			recordsetID, err)
	}

	// update tags
	resourceType, err := getDNSRecordSetResourceType(client, zoneID)
	if err != nil {
		return fmterr.Errorf("error getting resource type of DNS record set %s: %s", d.Id(), err)
	}

	tagErr := common.UpdateResourceTags(client, d, resourceType, recordsetID)
	if tagErr != nil {
		return fmterr.Errorf("error updating tags of DNS record set %s: %s", d.Id(), tagErr)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceDNSRecordSetV2Read(clientCtx, d, meta)
}

func resourceDNSRecordSetV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DnsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	shared := d.Get("shared").(bool)
	if shared {
		log.Printf("[DEBUG] Using non-managed DNS record set, skipping deletion")
		d.SetId("")
		return nil
	}

	// Obtain relevant info from parsing the ID
	zoneID, recordsetID, err := ParseDNSV2RecordSetID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = recordsets.Delete(client, zoneID, recordsetID).ExtractErr()
	if err != nil {
		// Check if the region is "eu-nl" and retry with "eu-de"
		if config.GetRegion(d) == "eu-nl" {
			replaceNlEndpoint(client)
			// Retry the request after updating the endpoint
			err = recordsets.Delete(client, zoneID, recordsetID).ExtractErr()
			if err != nil {
				return fmterr.Errorf("error deleting OpenTelekomCloud DNS record set: %s", err)
			}
		} else {
			return fmterr.Errorf("error deleting OpenTelekomCloud DNS record set: %s", err)
		}
	}

	log.Printf("[DEBUG] Waiting for DNS record set (%s) to be deleted", recordsetID)
	stateConf := &resource.StateChangeConf{
		Target:       []string{"DELETED"},
		Pending:      []string{"ACTIVE", "PENDING", "ERROR"},
		Refresh:      waitForDNSRecordSet(client, zoneID, recordsetID),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        5 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf(
			"error waiting for record set (%s) to become DELETED for deletion: %s",
			recordsetID, err)
	}

	d.SetId("")
	return nil
}

func waitForDNSRecordSet(client *golangsdk.ServiceClient, zoneID, recordsetId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		recordset, err := recordsets.Get(client, zoneID, recordsetId).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return recordset, "DELETED", nil
			}

			return nil, "", err
		}

		log.Printf("[DEBUG] OpenTelekomCloud DNS record set (%s) current status: %s", recordset.ID, recordset.Status)
		return recordset, parseStatus(recordset.Status), nil
	}
}

func ParseDNSV2RecordSetID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unable to determine DNS record set ID from raw ID: %s", id)
	}

	zoneID := idParts[0]
	recordsetID := idParts[1]

	return zoneID, recordsetID, nil
}

// get resource type of DNS record set from zone_id
func getDNSRecordSetResourceType(client *golangsdk.ServiceClient, zoneID string) (string, error) {
	zone, err := zones.Get(client, zoneID).Extract()
	if err != nil {
		return "", err
	}

	zoneType := zone.ZoneType
	if zoneType == "public" {
		return "DNS-public_recordset", nil
	} else if zoneType == "private" {
		return "DNS-private_recordset", nil
	}
	return "", fmt.Errorf("invalid zone type: %s", zoneType)
}

func getExistingRecordSetID(d cfg.SchemaOrDiff, meta interface{}) (id string, err error) {
	config := meta.(*cfg.Config)
	client, err := config.DnsV2Client(config.GetRegion(d))
	if err != nil {
		err = fmt.Errorf("error creating OpenTelekomCloud DNS client: %s", err)
		return
	}

	createOpts := getRecordSetCreateOpts(d)

	zoneID := d.Get("zone_id").(string)
	if zoneID == "" {
		return
	}

	listOpts := recordsets.ListOpts{
		Name:   createOpts.Name,
		TTL:    createOpts.TTL,
		Type:   createOpts.Type,
		ZoneID: zoneID,
	}

	allPages, err := recordsets.ListByZone(client, zoneID, listOpts).AllPages()
	if err != nil {
		// Check if the region is "eu-nl" and retry with "eu-de"
		if config.GetRegion(d) == "eu-nl" {
			replaceNlEndpoint(client)
			// Retry the request after updating the endpoint
			allPages, err = recordsets.ListByZone(client, zoneID, listOpts).AllPages()
			if err != nil {
				return "", fmt.Errorf("error listing record sets: %s", err)
			}
		} else {
			return "", fmt.Errorf("error listing record sets: %s", err)
		}
	}
	sets, err := recordsets.ExtractRecordSets(allPages)
	if err != nil {
		return "", fmt.Errorf("error extracting record sets: %s", err)
	}
	if len(sets) == 0 {
		return
	}
	expectedName := createOpts.Name
	if !strings.HasSuffix(expectedName, ".") {
		expectedName = fmt.Sprintf("%s.", expectedName)
	}
	for _, set := range sets {
		if set.Name == expectedName {
			id = set.ID
			return id, err
		}
	}

	return id, err
}

func useSharedRecordSet(_ context.Context, d *schema.ResourceDiff, meta interface{}) (err error) {
	if d.Id() != "" { // skip if not new resource
		return
	}

	if _, ok := d.GetOk("shared"); ok { // skip if shared is already set
		return
	}

	id, err := getExistingRecordSetID(d, meta)
	if id == "" {
		_ = d.SetNew("shared", false)
		return
	}

	_ = d.SetNew("shared", true)
	return
}
