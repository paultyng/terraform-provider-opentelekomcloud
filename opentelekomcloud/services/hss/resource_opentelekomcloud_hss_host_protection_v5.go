package hss

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/host"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceHostProtection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostProtectionCreate,
		ReadContext:   resourceHostProtectionRead,
		UpdateContext: resourceHostProtectionUpdate,
		DeleteContext: resourceHostProtectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHostProtectionImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"host_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"charging_mode": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_wait_host_available": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"host_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detect_result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHostProtectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}
	hostId := d.Get("host_id").(string)
	if d.Get("is_wait_host_available").(bool) {
		if err := waitingForHostAvailable(ctx, client, hostId, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("error waiting for OpenTelekomCloud HSS host (%s) agent status to become online: %s", hostId, err)
		}
	} else {
		checkHostAvailableErr := checkHostAvailable(client, hostId)
		if checkHostAvailableErr != nil {
			return diag.FromErr(checkHostAvailableErr)
		}
	}

	// Due to API limitations, when switching host protection for the first time, protection needs to be closed first.
	err = closeHostProtection(client, hostId)
	if err != nil {
		return diag.Errorf("error closing host protection before opening OpenTelekomCloud HSS host (%s) protection: %s",
			hostId, err)
	}

	err = switchHostsProtectStatus(client, hostId, d)
	if err != nil {
		return diag.Errorf("error opening OpenTelekomCloud HSS host (%s) protection: %s", hostId, err)
	}

	d.SetId(hostId)

	clientCtx := common.CtxWithClient(ctx, client, hssClientV5)
	return resourceHostProtectionRead(clientCtx, d, meta)
}

func getProtectionHost(client *golangsdk.ServiceClient, hostId string) (*host.HostResp, error) {
	hostList, err := host.ListHost(client, host.ListHostOpts{HostID: hostId})
	if err != nil {
		return nil, fmt.Errorf("error querying OpenTelekomCloud HSS hosts: %s", err)
	}
	if len(hostList) == 0 {
		return nil, fmt.Errorf("%s", getProtectionHostNeedRetryMsg)
	}

	return &hostList[0], nil
}

func checkHostAvailable(client *golangsdk.ServiceClient, hostId string) error {
	hostList, err := host.ListHost(client, host.ListHostOpts{HostID: hostId})
	if err != nil {
		return fmt.Errorf("error querying OpenTelekomCloud HSS hosts: %s", err)
	}
	if len(hostList) == 0 {
		return fmt.Errorf("the host (%s) does not exist", hostId)
	}

	agentStatus := hostList[0].AgentStatus
	if agentStatus != hostAgentStatusOnline {
		return fmt.Errorf("the host anget status for OpenTelekomCloud HSS protection must be: %s,"+
			" but the current host (%s) agent status is: %s ", hostAgentStatusOnline, hostId, agentStatus)
	}

	return nil
}

func closeHostProtection(client *golangsdk.ServiceClient, hostId string) error {
	_, err := host.ChangeProtectionStatus(client,
		host.ProtectionOpts{
			Version: protectionVersionNull,
			HostIds: []string{hostId},
		},
	)
	if err != nil {
		return err
	}
	return err
}

func switchHostsProtectStatus(client *golangsdk.ServiceClient, hostId string, d *schema.ResourceData) error {
	_, err := host.ChangeProtectionStatus(client,
		host.ProtectionOpts{
			Version:      d.Get("version").(string),
			ChargingMode: d.Get("charging_mode").(string),
			ResourceId:   d.Get("resource_id").(string),
			HostIds:      []string{hostId},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func waitingForHostAvailable(ctx context.Context, client *golangsdk.ServiceClient, hostId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PENDING"},
		Target:  []string{"COMPLETED"},
		Refresh: func() (interface{}, string, error) {
			h, err := getProtectionHost(client, hostId)
			if err != nil {
				if err.Error() == getProtectionHostNeedRetryMsg {
					return nil, "PENDING", nil
				}

				return nil, "ERROR", err
			}

			if h.AgentStatus == hostAgentStatusOnline {
				return h, "COMPLETED", nil
			}

			return h, "PENDING", nil
		},
		Timeout:      timeout,
		Delay:        20 * time.Second,
		PollInterval: 20 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func resourceHostProtectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	hostList, err := host.ListHost(client, host.ListHostOpts{HostID: d.Id()})
	if err != nil {
		return diag.Errorf("error querying OpenTelekomCloud HSS hosts: %s", err)
	}

	h := hostList[0]
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("host_id", h.ID),
		d.Set("version", h.Version),
		d.Set("charging_mode", h.ChargingMode),
		d.Set("host_name", h.Name),
		d.Set("host_status", h.HostStatus),
		d.Set("private_ip", h.PrivateIp),
		d.Set("agent_id", h.AgentId),
		d.Set("agent_status", h.AgentStatus),
		d.Set("os_type", h.OsType),
		d.Set("status", h.ProtectStatus),
		d.Set("detect_result", h.DetectResult),
		d.Set("asset_value", h.AssetValue),
	)
	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceHostProtectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	checkHostAvailableErr := checkHostAvailable(client, d.Id())
	if checkHostAvailableErr != nil {
		return diag.FromErr(checkHostAvailableErr)
	}

	if d.HasChanges("version", "charging_mode", "resource_id") {
		// Due to API limitations, when switching host protection for the first time, protection needs to be closed first.
		err = closeHostProtection(client, d.Id())
		if err != nil {
			return diag.Errorf("error closing host protection before updating OpenTelekomCloud HSS host (%s) protection: %s", d.Id(), err)
		}

		err = switchHostsProtectStatus(client, d.Id(), d)
		if err != nil {
			return diag.Errorf("error updating OpenTelekomCloud HSS host (%s) protection: %s", d.Id(), err)
		}
	}

	return resourceHostProtectionRead(ctx, d, meta)
}

func resourceHostProtectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	err = closeHostProtection(client, d.Id())
	if err != nil {
		return diag.Errorf("error closing OpenTelekomCloud HSS host (%s) protection: %s", d.Id(), err)
	}

	return nil
}

func resourceHostProtectionImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return nil, fmt.Errorf(errCreationV5Client, err)
	}

	checkHostAvailableErr := checkHostAvailable(client, d.Id())

	return []*schema.ResourceData{d}, checkHostAvailableErr
}
