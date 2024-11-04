package ecs

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/blockstorage/v2/volumes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/tags"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/availabilityzones"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/keypairs"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/schedulerhints"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/secgroups"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/startstop"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/flavors"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/servers"
	ims2 "github.com/opentelekomcloud/gophertelekomcloud/openstack/ims/v2/images"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/helper/hashcode"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/ims"
	"golang.org/x/crypto/ssh"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func ResourceComputeInstanceV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeInstanceV2Create,
		ReadContext:   resourceComputeInstanceV2Read,
		UpdateContext: resourceComputeInstanceV2Update,
		DeleteContext: resourceComputeInstanceV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceComputeInstanceV2ImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
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
				Required: true,
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_IMAGE_ID", nil),
			},
			"image_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_IMAGE_NAME", nil),
			},
			"flavor_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_FLAVOR_ID", nil),
			},
			"flavor_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_FLAVOR_NAME", nil),
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
			"ssh_private_key_path": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Sensitive: true,
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
				ForceNew: true,
				Computed: true,
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"fixed_ip_v4": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"fixed_ip_v6": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"mac": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_network": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"config_drive": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"encrypted_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"admin_pass": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Sensitive: true,
			},
			"access_ip_v4": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			"access_ip_v6": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			"key_pair": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"block_device": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"destination_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"boot_index": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						"guest_format": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"scheduler_hints": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"different_host": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"same_host": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"query": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"target_cell": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"build_near_host_ip": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"tenancy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"deh_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				Set: resourceComputeSchedulerHintsHash,
			},
			"stop_before_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"power_state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "active",
				ValidateFunc: validation.StringInSlice([]string{
					"active", "shutoff",
				}, true),
				DiffSuppressFunc: suppressPowerStateDiffs,
			},
			"tags": common.TagsSchema(),
			"all_metadata": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"auto_recovery": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"volume_attached": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceComputeInstanceV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.ComputeV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateV2Client, err)
	}
	client.Microversion = microversion

	var createOpts servers.CreateOptsBuilder

	// Determines the Image ID using the following rules:
	// If a bootable block_device was specified, ignore the image altogether.
	// If an image_id was specified, use it.
	// If an image_name was specified, look up the image ID, report if error.
	imageID, err := getImageIDFromConfig(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	flavorID, err := getFlavorID(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// determine if block_device configuration is correct
	// this includes valid combinations and required attributes
	if err := checkBlockDeviceConfig(d); err != nil {
		return diag.FromErr(err)
	}

	// Build a list of networks with the information given upon creation.
	// Error out if an invalid network configuration was used.
	allInstanceNetworks, err := getAllInstanceNetworks(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build a []servers.Network to pass into the create options.
	networks := expandInstanceNetworks(allInstanceNetworks)

	configDrive := d.Get("config_drive").(bool)

	createOpts = &servers.CreateOpts{
		Name:             d.Get("name").(string),
		ImageRef:         imageID,
		FlavorRef:        flavorID,
		SecurityGroups:   resourceInstanceSecGroupsV2(d),
		AvailabilityZone: d.Get("availability_zone").(string),
		Networks:         networks,
		Metadata:         resourceInstanceMetadataV2(d),
		ConfigDrive:      &configDrive,
		AdminPass:        d.Get("admin_pass").(string),
		UserData:         []byte(d.Get("user_data").(string)),
		Description:      d.Get("description").(string),
	}

	if keyName, ok := d.Get("key_pair").(string); ok && keyName != "" {
		createOpts = &keypairs.CreateOptsExt{
			CreateOptsBuilder: createOpts,
			KeyName:           keyName,
		}
	}

	if vL, ok := d.GetOk("block_device"); ok {
		blockDevices, err := ResourceInstanceBlockDevicesV2(d, vL.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		createOpts = &bootfromvolume.CreateOptsExt{
			CreateOptsBuilder: createOpts,
			BlockDevice:       blockDevices,
		}
	}

	schedulerHintsRaw := d.Get("scheduler_hints").(*schema.Set).List()
	if len(schedulerHintsRaw) > 0 {
		log.Printf("[DEBUG] schedulerhints: %+v", schedulerHintsRaw)
		schedulerHints := resourceInstanceSchedulerHintsV2(d, schedulerHintsRaw[0].(map[string]interface{}))
		createOpts = &schedulerhints.CreateOptsExt{
			CreateOptsBuilder: createOpts,
			SchedulerHints:    schedulerHints,
		}
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	// If a block_device is used, use the bootfromvolume. Create function as it allows an empty ImageRef.
	// Otherwise, use the normal servers. Create function.
	var server *servers.Server
	if _, ok := d.GetOk("block_device"); ok {
		server, err = bootfromvolume.Create(client, createOpts).Extract()
	} else {
		server, err = servers.Create(client, createOpts).Extract()
	}

	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud server: %w", err)
	}
	log.Printf("[INFO] Instance ID: %s", server.ID)

	// Wait for the instance to become running so we can get some attributes
	// that aren't available until later.
	log.Printf("[DEBUG] Waiting for instance (%s) to become running", server.ID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"BUILD"},
		Target:       []string{"ACTIVE"},
		Refresh:      ServerV2StateRefreshFunc(client, server.ID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        10 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for instance (%s) to become ready: %s", server.ID, err)
	}

	// Store the ID now
	d.SetId(server.ID)

	vmState := d.Get("power_state").(string)
	if strings.ToLower(vmState) == "shutoff" {
		err = startstop.Stop(client, d.Id()).ExtractErr()
		if err != nil {
			return fmterr.Errorf("error stopping server instance: %w", err)
		}
		stopStateConf := &resource.StateChangeConf{
			Target:       []string{"SHUTOFF"},
			Refresh:      ServerV2StateRefreshFunc(client, d.Id()),
			Timeout:      d.Timeout(schema.TimeoutCreate),
			Delay:        10 * time.Second,
			MinTimeout:   3 * time.Second,
			PollInterval: 2 * time.Second,
		}

		log.Printf("[DEBUG] Waiting for instance (%s) to stop", d.Id())
		_, err = stopStateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmterr.Errorf("error waiting for instance (%s) to become inactive(shutoff): %w", d.Id(), err)
		}
	}

	if common.HasFilledOpt(d, "auto_recovery") {
		ar := d.Get("auto_recovery").(bool)
		log.Printf("[DEBUG] Set auto recovery of instance to %t", ar)
		err = setAutoRecoveryForInstance(ctx, d, meta, server.ID, ar)
		if err != nil {
			log.Printf("[WARN] Error setting auto recovery of instance: %s", err)
		}
	}

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		computeClient, err := config.ComputeV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud ComputeV1 client: %w", err)
		}
		tagList := common.ExpandResourceTags(tagRaw)
		if err := tags.Create(computeClient, "cloudservers", server.ID, tagList).ExtractErr(); err != nil {
			return fmterr.Errorf("error setting tags of CloudServer: %w", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceComputeInstanceV2Read(clientCtx, d, meta)
}

func resourceComputeInstanceV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.ComputeV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateV2Client, err)
	}
	client.Microversion = microversion
	server, err := servers.Get(client, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "server")
	}

	log.Printf("[DEBUG] Retrieved Server %s: %+v", d.Id(), server)

	var mErr *multierror.Error

	mErr = multierror.Append(mErr, d.Set("name", server.Name))

	// Get the instance network and address information
	networks, err := FlattenInstanceNetworks(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Determine the best IPv4 and IPv6 addresses to access the instance with
	hostv4, hostv6 := GetInstanceAccessAddresses(d, networks)

	// AccessIPv4/v6 isn't standard in OpenTelekomCloud, but there have been reports
	// of them being used in some environments.
	if server.AccessIPv4 != "" && hostv4 == "" {
		hostv4 = server.AccessIPv4
	}

	if server.AccessIPv6 != "" && hostv6 == "" {
		hostv6 = server.AccessIPv6
	}

	mErr = multierror.Append(mErr,
		d.Set("network", networks),
		d.Set("access_ip_v4", hostv4),
		d.Set("access_ip_v6", hostv6),
	)

	// Determine the best IP address to use for SSH connectivity.
	// Prefer IPv4 over IPv6.
	var preferredSSHAddress string
	if hostv4 != "" {
		preferredSSHAddress = hostv4
	} else if hostv6 != "" {
		preferredSSHAddress = hostv6
	}

	if preferredSSHAddress != "" {
		// Initialize the connection info
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": preferredSSHAddress,
		})
	}

	if err := d.Set("all_metadata", server.Metadata); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving all_metadata to state for OpenTelekomCloud server (%s): %s", d.Id(), err)
	}

	var secGrpNames []string
	for _, sg := range server.SecurityGroups {
		secGrpNames = append(secGrpNames, sg["name"].(string))
	}
	if err := d.Set("security_groups", secGrpNames); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving security_groups to state for OpenTelekomCloud server (%s): %s", d.Id(), err)
	}

	flavorId, ok := server.Flavor["original_name"].(string)
	if !ok {
		flavorId, ok = server.Flavor["id"].(string)
		if !ok {
			return fmterr.Errorf("error setting OpenTelekomCloud server's flavor: %v", server.Flavor)
		}
	}
	mErr = multierror.Append(mErr,
		d.Set("flavor_id", flavorId),
		d.Set("key_pair", server.KeyName),
		d.Set("description", server.Description),
	)

	flavor, err := flavors.Get(client, flavorId).Extract()
	if err != nil {
		return diag.FromErr(err)
	}
	mErr = multierror.Append(mErr, d.Set("flavor_name", flavor.Name))

	// Create a slice of maps to match the Terraform schema
	var volumesAttached []map[string]interface{}
	for _, volume := range server.VolumesAttached {
		volumesAttached = append(volumesAttached, map[string]interface{}{
			"id": volume.ID,
		})
	}
	// Set instance volume attached information
	mErr = multierror.Append(mErr, d.Set("volume_attached", volumesAttached))

	// Set the instance's image information appropriately
	if err := setImageInformation(client, server, d); err != nil {
		return diag.FromErr(err)
	}

	// Changing `block_device` is possible only on instance creation.
	// At the same time, if instance is created from volume, `setBlockDevice` will set `block_device`
	// to exact attached device, not a source volume triggering instance re-creation
	if _, ok := d.GetOk("block_device"); !ok {
		if err := setBlockDevice(d, meta); err != nil {
			return err
		}
	}

	// Build a custom struct for the availability zone extension
	var serverWithAZ struct {
		servers.Server
		availabilityzones.ServerAvailabilityZoneExt
	}

	// Do another Get so the above work is not disturbed.
	err = servers.Get(client, d.Id()).ExtractInto(&serverWithAZ)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "server")
	}

	mErr = multierror.Append(mErr,
		// Set the availability zone
		d.Set("availability_zone", serverWithAZ.AvailabilityZone),
		// Set the region
		d.Set("region", config.GetRegion(d)),
	)

	// Set the current power_state
	currentStatus := strings.ToLower(server.Status)
	switch currentStatus {
	case "active", "shutoff", "error", "migrating", "shelved_offloaded", "shelved":
		mErr = multierror.Append(mErr, d.Set("power_state", currentStatus))
	default:
		return fmterr.Errorf("invalid power_state for instance %s: %s", d.Id(), server.Status)
	}

	ar, err := resourceECSAutoRecoveryV1Read(ctx, d, meta, d.Id())
	if err != nil && !common.IsResourceNotFound(err) {
		return fmterr.Errorf("error reading auto recovery of instance: %w", err)
	}
	mErr = multierror.Append(mErr, d.Set("auto_recovery", ar))

	computeClient, err := config.ComputeV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud ComputeV1 client: %w", err)
	}
	// save tags
	resourceTags, err := tags.Get(computeClient, "cloudservers", d.Id()).Extract()
	if err != nil {
		return fmterr.Errorf("error fetching OpenTelekomCloud CloudServers tags: %w", err)
	}
	tagMap := common.TagsToMap(resourceTags)
	mErr = multierror.Append(mErr, d.Set("tags", tagMap))

	// Set win instance password
	if v, ok := d.GetOk("ssh_private_key_path"); ok {
		readFile, err := os.ReadFile(v.(string))
		if err != nil {
			return fmterr.Errorf("error reading private key file: %w", err)
		}
		privateKey, err := ssh.ParseRawPrivateKey(readFile)
		if err != nil {
			return fmterr.Errorf("error parsing private key: %w", err)
		}
		pass, err := servers.GetPassword(client, d.Id()).ExtractPassword(privateKey.(*rsa.PrivateKey))
		if err != nil {
			return fmterr.Errorf("error getting password: %w", err)
		}
		mErr = multierror.Append(mErr, d.Set("password", pass))
	} else {
		pass, err := servers.GetPassword(client, d.Id()).ExtractPassword(nil)
		if err != nil {
			return fmterr.Errorf("error getting password: %w", err)
		}
		mErr = multierror.Append(mErr, d.Set("encrypted_password", pass))
	}

	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting opentelekomcloud_compute_instance_v2 values: %w", err)
	}

	return nil
}

func resourceComputeInstanceV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.ComputeV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateV2Client, err)
	}
	client.Microversion = microversion

	var updateOpts servers.UpdateOpts
	if d.HasChanges("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChanges("description") {
		updateOpts.Description = d.Get("description").(string)
	}
	if updateOpts != (servers.UpdateOpts{}) {
		_, err := servers.Update(client, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud server: %w", err)
		}
	}

	if d.HasChange("power_state") {
		powerStateNew := d.Get("power_state").(string)
		if strings.ToLower(powerStateNew) == "shutoff" {
			err = startstop.Stop(client, d.Id()).ExtractErr()
			if err != nil {
				return fmterr.Errorf("error stopping compute instance: %w", err)
			}
			stopStateConf := &resource.StateChangeConf{
				Target:       []string{"SHUTOFF"},
				Refresh:      ServerV2StateRefreshFunc(client, d.Id()),
				Timeout:      d.Timeout(schema.TimeoutUpdate),
				Delay:        10 * time.Second,
				MinTimeout:   3 * time.Second,
				PollInterval: 2 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for instance (%s) to stop", d.Id())
			_, err = stopStateConf.WaitForStateContext(ctx)
			if err != nil {
				return fmterr.Errorf("error waiting for instance (%s) to become inactive(shutoff): %s", d.Id(), err)
			}
		}
		if strings.ToLower(powerStateNew) == "active" {
			err = startstop.Start(client, d.Id()).ExtractErr()
			if err != nil {
				return fmterr.Errorf("error starting compute instance: %w", err)
			}
			startStateConf := &resource.StateChangeConf{
				Target:       []string{"ACTIVE"},
				Refresh:      ServerV2StateRefreshFunc(client, d.Id()),
				Timeout:      d.Timeout(schema.TimeoutUpdate),
				Delay:        10 * time.Second,
				MinTimeout:   3 * time.Second,
				PollInterval: 2 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for instance (%s) to start/unshelve", d.Id())
			_, err = startStateConf.WaitForStateContext(ctx)
			if err != nil {
				return fmterr.Errorf("error waiting for instance (%s) to become active: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("metadata") {
		oldMetadata, newMetadata := d.GetChange("metadata")
		var metadataToDelete []string

		// Determine if any metadata keys were removed from the configuration.
		// Then request those keys to be deleted.
		for oldKey := range oldMetadata.(map[string]interface{}) {
			var found bool
			for newKey := range newMetadata.(map[string]interface{}) {
				if oldKey == newKey {
					found = true
				}
			}

			if !found {
				metadataToDelete = append(metadataToDelete, oldKey)
			}
		}

		for _, key := range metadataToDelete {
			err := servers.DeleteMetadatum(client, d.Id(), key).ExtractErr()
			if err != nil {
				return fmterr.Errorf("error deleting metadata (%s) from server (%s): %s", key, d.Id(), err)
			}
		}

		// Update existing metadata and add any new metadata.
		metadataOpts := make(servers.MetadataOpts)
		for k, v := range newMetadata.(map[string]interface{}) {
			metadataOpts[k] = v.(string)
		}

		_, err := servers.UpdateMetadata(client, d.Id(), metadataOpts).Extract()
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud server (%s) metadata: %s", d.Id(), err)
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

		for _, g := range secGroupsToRemove.List() {
			err := secgroups.RemoveServer(client, d.Id(), g.(string)).ExtractErr()
			if err != nil && err.Error() != "EOF" {
				if _, ok := err.(golangsdk.ErrDefault404); ok {
					continue
				}

				return fmterr.Errorf("error removing security group (%s) from OpenTelekomCloud server (%s): %s", g, d.Id(), err)
			} else {
				log.Printf("[DEBUG] Removed security group (%s) from instance (%s)", g, d.Id())
			}
		}

		for _, g := range secGroupsToAdd.List() {
			err := secgroups.AddServer(client, d.Id(), g.(string)).ExtractErr()
			if err != nil && err.Error() != "EOF" {
				return fmterr.Errorf("error adding security group (%s) to OpenTelekomCloud server (%s): %s", g, d.Id(), err)
			}
			log.Printf("[DEBUG] Added security group (%s) to instance (%s)", g, d.Id())
		}
	}

	if d.HasChange("admin_pass") {
		if newPwd, ok := d.Get("admin_pass").(string); ok {
			err := servers.ChangeAdminPassword(client, d.Id(), newPwd).ExtractErr()
			if err != nil {
				return fmterr.Errorf("error changing admin password of OpenTelekomCloud server (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("flavor_id") || d.HasChange("flavor_name") {
		var newFlavorId string
		var err error
		if d.HasChange("flavor_id") {
			newFlavorId = d.Get("flavor_id").(string)
		} else {
			newFlavorName := d.Get("flavor_name").(string)
			newFlavorId, err = flavors.IDFromName(client, newFlavorName)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		resizeOpts := &servers.ResizeOpts{
			FlavorRef: newFlavorId,
		}
		log.Printf("[DEBUG] Resize configuration: %#v", resizeOpts)
		err = servers.Resize(client, d.Id(), resizeOpts).ExtractErr()
		if err != nil {
			return fmterr.Errorf("error resizing OpenTelekomCloud server: %w", err)
		}

		// Wait for the instance to finish resizing.
		log.Printf("[DEBUG] Waiting for instance (%s) to finish resizing", d.Id())

		stateConf := &resource.StateChangeConf{
			Pending:      []string{"RESIZE"},
			Target:       []string{"VERIFY_RESIZE"},
			Refresh:      ServerV2StateRefreshFunc(client, d.Id()),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			Delay:        10 * time.Second,
			MinTimeout:   3 * time.Second,
			PollInterval: 10 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmterr.Errorf("error waiting for instance (%s) to resize: %s", d.Id(), err)
		}

		// Confirm resize.
		log.Printf("[DEBUG] Confirming resize")
		err = servers.ConfirmResize(client, d.Id()).ExtractErr()
		if err != nil {
			return fmterr.Errorf("error confirming resize of OpenTelekomCloud server: %w", err)
		}

		stateConf = &resource.StateChangeConf{
			Pending:      []string{"VERIFY_RESIZE"},
			Target:       []string{"ACTIVE"},
			Refresh:      ServerV2StateRefreshFunc(client, d.Id()),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			Delay:        10 * time.Second,
			MinTimeout:   3 * time.Second,
			PollInterval: 5 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return fmterr.Errorf("error waiting for instance (%s) to confirm resize: %s", d.Id(), err)
		}
	}

	// update tags
	if d.HasChange("tags") {
		computeClient, err := config.ComputeV1Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating OpenTelekomCloud ComputeV1 client: %w", err)
		}
		if err := common.UpdateResourceTags(computeClient, d, "cloudservers", d.Id()); err != nil {
			return fmterr.Errorf("error updating tags of CloudServer %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("auto_recovery") {
		ar := d.Get("auto_recovery").(bool)
		log.Printf("[DEBUG] Update auto recovery of instance to %t", ar)
		err = setAutoRecoveryForInstance(ctx, d, meta, d.Id(), ar)
		if err != nil {
			return fmterr.Errorf("error updating auto recovery of instance: %w", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceComputeInstanceV2Read(clientCtx, d, meta)
}

func resourceComputeInstanceV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.ComputeV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreateV2Client, err)
	}
	client.Microversion = microversion
	if d.Get("stop_before_destroy").(bool) {
		err = startstop.Stop(client, d.Id()).ExtractErr()
		if err != nil {
			log.Printf("[WARN] Error stopping OpenTelekomCloud instance: %s", err)
		} else {
			stopStateConf := &resource.StateChangeConf{
				Pending:    []string{"ACTIVE"},
				Target:     []string{"SHUTOFF"},
				Refresh:    ServerV2StateRefreshFunc(client, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				Delay:      10 * time.Second,
				MinTimeout: 3 * time.Second,
			}
			log.Printf("[DEBUG] Waiting for instance (%s) to stop", d.Id())
			_, err = stopStateConf.WaitForStateContext(ctx)
			if err != nil {
				log.Printf("[WARN] Error waiting for instance (%s) to stop: %s, proceeding to delete", d.Id(), err)
			}
		}
	}
	log.Printf("[DEBUG] Deleting OpenTelekomCloud Instance %s", d.Id())
	if err := servers.Delete(client, d.Id()).ExtractErr(); err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud server: %w", err)
	}

	// Wait for the instance to delete before moving on.
	log.Printf("[DEBUG] Waiting for instance (%s) to delete", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"ACTIVE", "SHUTOFF"},
		Target:       []string{"DELETED", "SOFT_DELETED"},
		Refresh:      ServerV2StateRefreshFunc(client, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        10 * time.Second,
		MinTimeout:   3 * time.Second,
		PollInterval: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for instance (%s) to delete: %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

// ServerV2StateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an OpenTelekomCloud instance.
func ServerV2StateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		s, err := servers.Get(client, instanceID).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return s, "DELETED", nil
			}
			return nil, "", err
		}

		// get fault message when status is ERROR
		if s.Status == "ERROR" {
			fault := fmt.Errorf("[error code: %d, message: %s]", s.Fault.Code, s.Fault.Message)
			return s, "ERROR", fault
		}
		return s, s.Status, nil
	}
}

func resourceInstanceSecGroupsV2(d *schema.ResourceData) []string {
	secGroupsRaw := d.Get("security_groups").(*schema.Set).List()
	secGroups := make([]string, len(secGroupsRaw))
	for i, secGroup := range secGroupsRaw {
		secGroups[i] = secGroup.(string)
	}
	return secGroups
}

func resourceInstanceMetadataV2(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("metadata").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func ResourceInstanceBlockDevicesV2(_ *schema.ResourceData, bds []interface{}) ([]bootfromvolume.BlockDevice, error) {
	blockDeviceOpts := make([]bootfromvolume.BlockDevice, len(bds))
	for i, bd := range bds {
		bdM := bd.(map[string]interface{})
		blockDeviceOpts[i] = bootfromvolume.BlockDevice{
			UUID:                bdM["uuid"].(string),
			VolumeSize:          bdM["volume_size"].(int),
			BootIndex:           bdM["boot_index"].(int),
			DeleteOnTermination: bdM["delete_on_termination"].(bool),
			GuestFormat:         bdM["guest_format"].(string),
			VolumeType:          bdM["volume_type"].(string),
			DeviceName:          bdM["device_name"].(string),
		}

		sourceType := bdM["source_type"].(string)
		switch sourceType {
		case "blank":
			blockDeviceOpts[i].SourceType = bootfromvolume.SourceBlank
		case "image":
			blockDeviceOpts[i].SourceType = bootfromvolume.SourceImage
		case "snapshot":
			blockDeviceOpts[i].SourceType = bootfromvolume.SourceSnapshot
		case "volume":
			blockDeviceOpts[i].SourceType = bootfromvolume.SourceVolume
		default:
			return blockDeviceOpts, fmt.Errorf("unknown block device source type %s", sourceType)
		}

		destinationType := bdM["destination_type"].(string)
		switch destinationType {
		case "local":
			blockDeviceOpts[i].DestinationType = bootfromvolume.DestinationLocal
		case "volume":
			blockDeviceOpts[i].DestinationType = bootfromvolume.DestinationVolume
		default:
			return blockDeviceOpts, fmt.Errorf("unknown block device destination type %s", destinationType)
		}
	}

	log.Printf("[DEBUG] Block Device Options: %+v", blockDeviceOpts)
	return blockDeviceOpts, nil
}

func resourceInstanceSchedulerHintsV2(_ *schema.ResourceData, schedulerHintsRaw map[string]interface{}) schedulerhints.SchedulerHints {
	var differentHost []string
	if len(schedulerHintsRaw["different_host"].([]interface{})) > 0 {
		for _, dh := range schedulerHintsRaw["different_host"].([]interface{}) {
			differentHost = append(differentHost, dh.(string))
		}
	}

	var sameHost []string
	if len(schedulerHintsRaw["same_host"].([]interface{})) > 0 {
		for _, sh := range schedulerHintsRaw["same_host"].([]interface{}) {
			sameHost = append(sameHost, sh.(string))
		}
	}

	query := make([]interface{}, len(schedulerHintsRaw["query"].([]interface{})))
	if len(schedulerHintsRaw["query"].([]interface{})) > 0 {
		for _, q := range schedulerHintsRaw["query"].([]interface{}) {
			query = append(query, q.(string))
		}
	}

	schedulerHints := schedulerhints.SchedulerHints{
		Group:           schedulerHintsRaw["group"].(string),
		DifferentHost:   differentHost,
		SameHost:        sameHost,
		Query:           query,
		TargetCell:      schedulerHintsRaw["target_cell"].(string),
		BuildNearHostIP: schedulerHintsRaw["build_near_host_ip"].(string),
		Tenancy:         schedulerHintsRaw["tenancy"].(string),
		DedicatedHostID: schedulerHintsRaw["deh_id"].(string),
	}

	return schedulerHints
}

func getImageIDFromConfig(d *schema.ResourceData, config *cfg.Config) (string, error) {
	// If block_device was used, an Image does not need to be specified, unless an image/local
	// combination was used. This emulates normal boot behavior. Otherwise, ignore the image altogether.
	if vL, ok := d.GetOk("block_device"); ok {
		needImage := false
		for _, v := range vL.([]interface{}) {
			vM := v.(map[string]interface{})
			if vM["source_type"] == "image" && vM["destination_type"] == "local" {
				needImage = true
			}
		}
		if !needImage {
			return "", nil
		}
	}

	if imageId := d.Get("image_id").(string); imageId != "" {
		return imageId, nil
	}

	if imageName := d.Get("image_name").(string); imageName != "" {
		return ims.GetImageByName(d, config, imageName)
	}

	return "", errors.New("neither a boot device, image ID, or image name were able to be determined")
}

func setImageInformation(client *golangsdk.ServiceClient, server *servers.Server, d *schema.ResourceData) error {
	imageId := server.Image["id"].(string)
	if imageId != "" {
		if err := d.Set("image_id", imageId); err != nil {
			return err
		}

		if imageName := d.Get("image_name").(string); imageName == "" {
			if image, err := ims2.ListImages(client, ims2.ListImagesOpts{Id: imageId}); err != nil {
				if _, ok := err.(golangsdk.ErrDefault404); ok {
					// If the image name can't be found, don't set name.
					// The most likely scenario is that the image no longer exists in the Image Service
					// but the instance still has a record from when it existed.
					return d.Set("image_name", "Not Found")
				}
				return err
			} else {
				return d.Set("image_name", image[0].Name)
			}
		}
	}

	return nil
}

func getFlavorID(client *golangsdk.ServiceClient, d *schema.ResourceData) (string, error) {
	flavorId := d.Get("flavor_id").(string)

	if flavorId != "" {
		return flavorId, nil
	}

	flavorName := d.Get("flavor_name").(string)
	return flavors.IDFromName(client, flavorName)
}

func resourceComputeSchedulerHintsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if m["group"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["group"].(string)))
	}

	if m["target_cell"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["target_cell"].(string)))
	}

	if m["build_host_near_ip"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["build_host_near_ip"].(string)))
	}

	if m["tenancy"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["tenancy"].(string)))
	}

	if m["deh_id"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["deh_id"].(string)))
	}

	buf.WriteString(fmt.Sprintf("%s-", m["different_host"].([]interface{})))
	buf.WriteString(fmt.Sprintf("%s-", m["same_host"].([]interface{})))
	buf.WriteString(fmt.Sprintf("%s-", m["query"].([]interface{})))

	return hashcode.String(buf.String())
}

func checkBlockDeviceConfig(d *schema.ResourceData) error {
	if vL, ok := d.GetOk("block_device"); ok {
		for _, v := range vL.([]interface{}) {
			vM := v.(map[string]interface{})

			if vM["source_type"] != "blank" && vM["uuid"] == "" {
				return fmt.Errorf("you must specify a UUID for %s block device types", vM["source_type"])
			}

			if vM["source_type"] == "image" && vM["destination_type"] == "volume" {
				if vM["volume_size"] == 0 {
					return fmt.Errorf("you must specify a volume_size when creating a volume from an image")
				}
			}

			if vM["source_type"] == "blank" && vM["destination_type"] == "local" {
				if vM["volume_size"] == 0 {
					return fmt.Errorf("you must specify a volume_size when creating a blank block device")
				}
			}
		}
	}

	return nil
}

// suppressPowerStateDiffs will allow a state of "error" or "migrating" even though we don't
// allow them as a user input.
func suppressPowerStateDiffs(_, old, _ string, _ *schema.ResourceData) bool {
	if old == "error" || old == "migrating" {
		return true
	}
	return false
}

func resourceComputeInstanceV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*cfg.Config)
	client, err := config.ComputeV2Client(config.GetRegion(d))
	if err != nil {
		return nil, fmt.Errorf("error creating ComputeV2 client: %w", err)
	}
	client.Microversion = microversion
	results := make([]*schema.ResourceData, 1)
	if diagRead := resourceComputeInstanceV2Read(ctx, d, meta); diagRead.HasError() {
		return nil, fmt.Errorf("error reading opentelekomcloud_compute_instance_v2 %s: %s", d.Id(), diagRead[0].Summary)
	}

	metadata, err := servers.Metadata(client, d.Id()).Extract()
	if err != nil {
		return nil, fmt.Errorf("unable to read metadata for opentelekomcloud_compute_instance_v2 %s: %s", d.Id(), err)
	}

	if err := d.Set("metadata", metadata); err != nil {
		return nil, fmt.Errorf("error setting metadata")
	}

	results[0] = d

	return results, nil
}

func setBlockDevice(d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ComputeV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating compute v2 client: %w", err)
	}
	client.Microversion = microversion
	raw := servers.Get(client, d.Id())
	if raw.Err != nil {
		return common.CheckDeletedDiag(d, raw.Err, "opentelekomcloud_compute_instance_v2")
	}

	var serverWithAttachments struct {
		VolumesAttached []map[string]interface{} `json:"os-extended-volumes:volumes_attached"`
	}
	if err := raw.ExtractInto(&serverWithAttachments); err != nil {
		log.Printf("[DEBUG] unable to unmarshal raw struct to serverWithAttachments: %s", err)
	}

	log.Printf("[DEBUG] Retrieved opentelekomcloud_compute_instance_v2 %s volume attachments: %#v",
		d.Id(), serverWithAttachments)

	bds := make([]map[string]interface{}, 0)
	if len(serverWithAttachments.VolumesAttached) > 0 {
		blockStorageClient, err := config.BlockStorageV2Client(config.GetRegion(d))
		if err != nil {
			return fmterr.Errorf("error creating volume v2 client: %w", err)
		}

		var volMetaData = struct {
			VolumeImageMetadata map[string]interface{} `json:"volume_image_metadata"`
			ID                  string                 `json:"id"`
			Size                int                    `json:"size"`
			Bootable            string                 `json:"bootable"`
		}{}
		for i, b := range serverWithAttachments.VolumesAttached {
			rawVolume := volumes.Get(blockStorageClient, b["id"].(string))
			if err := rawVolume.ExtractInto(&volMetaData); err != nil {
				log.Printf("[DEBUG] unable to unmarshal raw struct to volume metadata: %s", err)
			}

			log.Printf("[DEBUG] retrieved volume%+v", volMetaData)
			v := map[string]interface{}{
				"delete_on_termination": true,
				"uuid":                  volMetaData.VolumeImageMetadata["image_id"],
				"boot_index":            i,
				"destination_type":      "volume",
				"source_type":           "image",
				"volume_size":           volMetaData.Size,
				"volume_type":           "",
			}

			if volMetaData.Bootable == "true" {
				bds = append(bds, v)
			}
		}

		if err := d.Set("block_device", bds); err != nil {
			return fmterr.Errorf("error setting block_device for compute_instance_v2: %w", err)
		}
	}
	return nil
}
