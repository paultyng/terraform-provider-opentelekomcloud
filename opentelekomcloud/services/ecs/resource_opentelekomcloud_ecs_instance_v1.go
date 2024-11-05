package ecs

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/secgroups"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/servers"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ecs/v1/cloudservers"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/evs/v3/volumes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/ports"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceEcsInstanceV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEcsInstanceV1Create,
		ReadContext:   resourceEcsInstanceV1Read,
		UpdateContext: resourceEcsInstanceV1Update,
		DeleteContext: resourceEcsInstanceV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: common.MultipleCustomizeDiffs(
			common.ValidateVPC("vpc_id"),
			common.ValidateVolumeType("system_disk_type"),
			common.ValidateVolumeType("data_disks.*.type"),
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// just stash the hash for state & diff comparisons
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return common.InstallScriptHashSum(v)
					default:
						return ""
					}
				},
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"nics": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 12,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"system_disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"system_disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"system_disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"system_disk_kms_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"data_disks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 23,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"kms_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							DefaultFunc: schema.EnvDefaultFunc("OS_KMS_ID", nil),
						},
						"snapshot_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"volumes_attached": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"kms_id": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("OS_KMS_ID", nil),
						},
						"snapshot_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"os_scheduler_hints": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"tenancy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"dedicated", "shared",
							}, true),
						},
						"dedicated_host_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"tags": common.TagsSchema(),
			"auto_recovery": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"delete_disks_on_termination": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceEcsInstanceV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	createOpts := &cloudservers.CreateOpts{
		Name:             d.Get("name").(string),
		ImageRef:         d.Get("image_id").(string),
		FlavorRef:        d.Get("flavor").(string),
		KeyName:          d.Get("key_name").(string),
		VpcId:            d.Get("vpc_id").(string),
		SecurityGroups:   resourceInstanceSecGroupsV1(d),
		AvailabilityZone: pointerto.String(d.Get("availability_zone").(string)),
		Nics:             resourceInstanceNicsV1(d),
		RootVolume:       resourceInstanceRootVolumeV1(d),
		DataVolumes:      resourceInstanceDataVolumesV1(d),
		AdminPass:        d.Get("password").(string),
		UserData:         []byte(d.Get("user_data").(string)),
		SchedulerHints:   resourceInstanceOsSchedulerHints(d),
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	jobResponse, err := cloudservers.Create(client, createOpts).ExtractJobResponse()
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud server: %w", err)
	}

	timeout := int(d.Timeout(schema.TimeoutCreate) / time.Second)
	if err := cloudservers.WaitForJobSuccess(client, timeout, jobResponse.JobID); err != nil {
		return diag.FromErr(err)
	}

	serverID, err := cloudservers.GetJobEntity(client, jobResponse.JobID, "server_id")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serverID.(string))

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		tagList := common.ExpandResourceTags(tagRaw)
		if err := tags.Create(client, "cloudservers", d.Id(), tagList).ExtractErr(); err != nil {
			return fmterr.Errorf("error setting tags of CloudServer: %w", err)
		}
	}

	if common.HasFilledOpt(d, "auto_recovery") {
		ar := d.Get("auto_recovery").(bool)
		log.Printf("[DEBUG] Set auto recovery of instance to %t", ar)

		if err := setAutoRecoveryForInstance(ctx, d, meta, d.Id(), ar); err != nil {
			log.Printf("[WARN] Error setting auto recovery of CloudServer: %s", err)
		}
	}

	return resourceEcsInstanceV1Read(ctx, d, meta)
}

func resourceEcsInstanceV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	server, err := cloudservers.Get(client, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "CloudServer")
	}
	if server.Status == "DELETED" {
		d.SetId("")
		return nil
	}

	mErr := multierror.Append(
		d.Set("name", server.Name),
		d.Set("image_id", server.Image.ID),
		d.Set("flavor", server.Flavor.ID),
		d.Set("password", d.Get("password")),
		d.Set("key_name", server.KeyName),
		d.Set("vpc_id", server.Metadata.VpcID),
		d.Set("availability_zone", server.AvailabilityZone),
	)
	var secGrpIDs []string
	for _, sg := range server.SecurityGroups {
		secGrpIDs = append(secGrpIDs, sg.ID)
	}
	mErr = multierror.Append(mErr,
		d.Set("security_groups", secGrpIDs),
	)
	evsClient, err := config.BlockStorageV3Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud EVSv3 client: %w", err)
	}
	var volumeList []map[string]interface{}
	for _, v := range server.VolumeAttached {
		disk, err := volumes.Get(evsClient, v.ID).Extract()
		if err != nil {
			return diag.FromErr(err)
		}
		if disk.Bootable == "true" {
			mErr = multierror.Append(mErr,
				d.Set("system_disk_id", disk.ID),
				d.Set("system_disk_size", disk.Size),
				d.Set("system_disk_type", disk.VolumeType),
				d.Set("system_disk_kms_id", disk.Metadata["__system__cmkid"]),
			)
			continue
		}
		dataVolume := map[string]interface{}{
			"id":          disk.ID,
			"type":        disk.VolumeType,
			"size":        disk.Size,
			"kms_id":      disk.Metadata["__system__cmkid"],
			"snapshot_id": disk.SnapshotID,
		}
		volumeList = append(volumeList, dataVolume)
	}
	mErr = multierror.Append(mErr,
		d.Set("volumes_attached", volumeList),
	)

	// Get the instance network and address information
	nics := flattenInstanceNicsV1(d, meta, server.Addresses)
	mErr = multierror.Append(mErr,
		d.Set("nics", nics),
	)

	schedulerHints := convertOsSchedulerHintsToSchedulerHintsV1(server.OsSchedulerHints)
	osSchedulerHints := []map[string]interface{}{
		{
			"group":             schedulerHints.Group,
			"tenancy":           schedulerHints.Tenancy,
			"dedicated_host_id": schedulerHints.DedicatedHostID,
		},
	}
	mErr = multierror.Append(mErr, d.Set("os_scheduler_hints", osSchedulerHints))

	// save tags
	resourceTags, err := tags.Get(client, "cloudservers", d.Id()).Extract()
	if err != nil {
		return fmterr.Errorf("error fetching OpenTelekomCloud CloudServers tags: %w", err)
	}
	tagMap := common.TagsToMap(resourceTags)
	if err := d.Set("tags", tagMap); err != nil {
		return fmterr.Errorf("error saving tags for OpenTelekomCloud CloudServers: %w", err)
	}

	ar, err := resourceECSAutoRecoveryV1Read(ctx, d, meta, d.Id())
	if err != nil && !common.IsResourceNotFound(err) {
		return fmterr.Errorf("error reading auto recovery of instance: %w", err)
	}
	mErr = multierror.Append(mErr,
		d.Set("auto_recovery", ar),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting ECS attributes: %w", err)
	}

	return nil
}

func resourceEcsInstanceV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ComputeV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud ComputeV2 client: %w", err)
	}

	var updateOpts servers.UpdateOpts
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}

	if updateOpts != (servers.UpdateOpts{}) {
		_, err := servers.Update(client, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud server: %w", err)
		}
	}

	if d.HasChange("security_groups") {
		oldSGRaw, newSGRaw := d.GetChange("security_groups")
		oldSGSet := oldSGRaw.(*schema.Set)
		newSGSet := newSGRaw.(*schema.Set)
		secGroupsToAdd := newSGSet.Difference(oldSGSet)
		secGroupsToRemove := oldSGSet.Difference(newSGSet)

		log.Printf("[DEBUG] Security groups to add: %v", secGroupsToAdd)

		log.Printf("[DEBUG] Security groups to remove: %v", secGroupsToRemove)

		for _, sg := range secGroupsToRemove.List() {
			err := secgroups.RemoveServer(client, d.Id(), sg.(string)).ExtractErr()
			if err != nil && err.Error() != "EOF" {
				if _, ok := err.(golangsdk.ErrDefault404); ok {
					continue
				}
				return fmterr.Errorf("error removing security group (%s) from OpenTelekomCloud server (%s): %w", sg, d.Id(), err)
			}
			log.Printf("[DEBUG] Removed security group (%s) from instance (%s)", sg, d.Id())
		}

		for _, sg := range secGroupsToAdd.List() {
			err := secgroups.AddServer(client, d.Id(), sg.(string)).ExtractErr()
			if err != nil && err.Error() != "EOF" {
				return fmterr.Errorf("error adding security group (%s) to OpenTelekomCloud server (%s): %w", sg, d.Id(), err)
			}
			log.Printf("[DEBUG] Added security group (%s) to instance (%s)", sg, d.Id())
		}
	}

	if d.HasChange("flavor") {
		newFlavorId := d.Get("flavor").(string)

		resizeOpts := &servers.ResizeOpts{
			FlavorRef: newFlavorId,
		}
		log.Printf("[DEBUG] Resize configuration: %#v", resizeOpts)

		if err := servers.Resize(client, d.Id(), resizeOpts).ExtractErr(); err != nil {
			return fmterr.Errorf("error resizing OpenTelekomCloud server: %w", err)
		}

		// Wait for the instance to finish resizing.
		log.Printf("[DEBUG] Waiting for instance (%s) to finish resizing", d.Id())

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"RESIZE"},
			Target:     []string{"VERIFY_RESIZE"},
			Refresh:    ServerV2StateRefreshFunc(client, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmterr.Errorf("error waiting for instance (%s) to resize: %w", d.Id(), err)
		}

		// Confirm resize.
		log.Printf("[DEBUG] Confirming resize")
		if err := servers.ConfirmResize(client, d.Id()).ExtractErr(); err != nil {
			return fmterr.Errorf("error confirming resize of OpenTelekomCloud server: %w", err)
		}

		stateConf = &resource.StateChangeConf{
			Pending:    []string{"VERIFY_RESIZE"},
			Target:     []string{"ACTIVE"},
			Refresh:    ServerV2StateRefreshFunc(client, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmterr.Errorf("error waiting for instance (%s) to confirm resize: %w", d.Id(), err)
		}
	}

	// update tags
	if d.HasChange("tags") {
		computeClient, err := config.ComputeV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf(errCreateClient, err)
		}
		if err := common.UpdateResourceTags(computeClient, d, "cloudservers", d.Id()); err != nil {
			return fmterr.Errorf("error updating tags of CloudServer %s: %w", d.Id(), err)
		}
	}

	if d.HasChange("auto_recovery") {
		ar := d.Get("auto_recovery").(bool)
		log.Printf("[DEBUG] Update auto recovery of instance to %t", ar)
		if err := setAutoRecoveryForInstance(ctx, d, meta, d.Id(), ar); err != nil {
			return fmterr.Errorf("error updating auto recovery of CloudServer: %w", err)
		}
	}

	return resourceEcsInstanceV1Read(ctx, d, meta)
}

func resourceEcsInstanceV1Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(errCreateClient, err)
	}

	var serverRequests []cloudservers.Server
	server := cloudservers.Server{
		Id: d.Id(),
	}
	serverRequests = append(serverRequests, server)

	deleteOpts := cloudservers.DeleteOpts{
		Servers:      serverRequests,
		DeleteVolume: d.Get("delete_disks_on_termination").(bool),
	}

	jobResponse, err := cloudservers.Delete(client, deleteOpts).ExtractJobResponse()
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud server: %w", err)
	}

	timeout := int(d.Timeout(schema.TimeoutDelete) / time.Second)
	if err := cloudservers.WaitForJobSuccess(client, timeout, jobResponse.JobID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceInstanceNicsV1(d *schema.ResourceData) []cloudservers.Nic {
	var nicRequests []cloudservers.Nic

	nics := d.Get("nics").([]interface{})
	for i := range nics {
		nic := nics[i].(map[string]interface{})
		nicRequest := cloudservers.Nic{
			SubnetId:  nic["network_id"].(string),
			IpAddress: nic["ip_address"].(string),
		}

		nicRequests = append(nicRequests, nicRequest)
	}
	return nicRequests
}

func resourceInstanceRootVolumeV1(d *schema.ResourceData) cloudservers.RootVolume {
	diskType := d.Get("system_disk_type").(string)
	if diskType == "" {
		diskType = "SATA"
	}
	volRequest := cloudservers.RootVolume{
		VolumeType: diskType,
		Size:       d.Get("system_disk_size").(int),
	}
	if kmsID := d.Get("system_disk_kms_id").(string); kmsID != "" {
		volRequest.Metadata = map[string]interface{}{
			"__system__cmkid":     kmsID,
			"__system__encrypted": "1",
		}
	}
	return volRequest
}

func resourceInstanceDataVolumesV1(d *schema.ResourceData) []cloudservers.DataVolume {
	var dataVolumes []cloudservers.DataVolume

	vols := d.Get("data_disks").([]interface{})
	for i := range vols {
		vol := vols[i].(map[string]interface{})
		volRequest := cloudservers.DataVolume{
			VolumeType: vol["type"].(string),
			Size:       vol["size"].(int),
		}
		if kmsID := vol["kms_id"]; kmsID != "" {
			volRequest.Metadata = map[string]interface{}{
				"__system__cmkid":     kmsID,
				"__system__encrypted": "1",
			}
		}
		if vol["snapshot_id"] != "" {
			extendParam := cloudservers.VolumeExtendParam{
				SnapshotId: vol["snapshot_id"].(string),
			}
			volRequest.Extendparam = &extendParam
		}

		dataVolumes = append(dataVolumes, volRequest)
	}
	return dataVolumes
}

func resourceInstanceOsSchedulerHints(d *schema.ResourceData) *cloudservers.SchedulerHints {
	rawOsSchedulerHintsList := d.Get("os_scheduler_hints").(*schema.Set).List()
	var schedulerHints cloudservers.SchedulerHints
	if len(rawOsSchedulerHintsList) > 0 {
		rawOsSchedulerHints := rawOsSchedulerHintsList[0].(map[string]interface{})
		schedulerHints = cloudservers.SchedulerHints{
			Group:           rawOsSchedulerHints["group"].(string),
			Tenancy:         rawOsSchedulerHints["tenancy"].(string),
			DedicatedHostID: rawOsSchedulerHints["dedicated_host_id"].(string),
		}
	}
	return &schedulerHints
}

// Function to convert OsSchedulerHints struct to SchedulerHints
func convertOsSchedulerHintsToSchedulerHintsV1(osSchedulerHints cloudservers.OsSchedulerHints) cloudservers.SchedulerHints {
	return cloudservers.SchedulerHints{
		Group:           firstOrEmpty(osSchedulerHints.Group),
		Tenancy:         firstOrEmpty(osSchedulerHints.Tenancy),
		DedicatedHostID: firstOrEmpty(osSchedulerHints.DedicatedHostID),
	}
}

// Helper function to return the first element or an empty string if the slice is empty
func firstOrEmpty(arr []string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}

func resourceInstanceSecGroupsV1(d *schema.ResourceData) []cloudservers.SecurityGroup {
	rawSecGroups := d.Get("security_groups").(*schema.Set).List()
	secGroups := make([]cloudservers.SecurityGroup, len(rawSecGroups))
	for i, raw := range rawSecGroups {
		secGroups[i] = cloudservers.SecurityGroup{
			ID: raw.(string),
		}
	}
	return secGroups
}

func flattenInstanceNicsV1(d *schema.ResourceData, meta interface{}, addresses map[string][]cloudservers.Address) []map[string]interface{} {
	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		log.Printf("error creating OpenTelekomCloud NetworkingV2 client: %s", err)
	}

	var network string
	var nics []map[string]interface{}
	// Loop through all networks and addresses.
	for _, addrs := range addresses {
		for _, addr := range addrs {
			// Skip if not fixed ip
			if addr.Type != "fixed" {
				continue
			}

			p, err := ports.Get(networkingClient, addr.PortID).Extract()
			if err != nil {
				network = ""
				log.Printf("[DEBUG] flattenInstanceNicsV1: failed to fetch port %s", addr.PortID)
			} else {
				network = p.NetworkID
			}

			v := map[string]interface{}{
				"network_id":  network,
				"ip_address":  addr.Addr,
				"mac_address": addr.MacAddr,
				"port_id":     addr.PortID,
				"type":        addr.Type,
			}
			nics = append(nics, v)
		}
	}

	log.Printf("[DEBUG] flattenInstanceNicsV1: %#v", nics)
	return nics
}
