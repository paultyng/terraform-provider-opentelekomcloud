package hss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	group "github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/host"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceHostGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostGroupCreate,
		ReadContext:   resourceHostGroupRead,
		UpdateContext: resourceHostGroupUpdate,
		DeleteContext: resourceHostGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"host_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"risk_host_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"unprotect_host_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"unprotect_host_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceHostGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	groupName := d.Get("name").(string)
	hostIds := common.ExpandToStringListBySet(d.Get("host_ids").(*schema.Set))

	opts := group.CreateOpts{
		Name:    groupName,
		HostIds: hostIds,
	}

	unprotected, err := checkAllHostsAvailable(ctx, client, hostIds, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] All OpenTelekomCloud HSS hosts are availabile.")
	if len(unprotected) > 1 {
		err := d.Set("unprotect_host_ids", unprotected)
		if err != nil {
			log.Printf("[WARN] These OpenTelekomCloud HSS hosts are not protected: %#v", unprotected)
		}
	}
	err = group.Create(client, opts)
	if err != nil {
		return diag.Errorf("error creating OpenTelekomCloud HSS host group: %s", err)
	}

	allHostGroups, err := queryHostGroups(client, groupName)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(allHostGroups) < 1 {
		return common.CheckDeletedDiag(d, err, "OpenTelekomCloud HSS host group")
	}
	d.SetId(allHostGroups[0].ID)

	clientCtx := common.CtxWithClient(ctx, client, hssClientV5)
	return resourceHostGroupRead(clientCtx, d, meta)
}

func checkAllHostsAvailable(ctx context.Context, client *golangsdk.ServiceClient, hostIDs []string, timeout time.Duration) ([]string, error) {
	unprotected := make([]string, 0)
	for _, hostId := range hostIDs {
		log.Printf("[DEBUG] Waiting for the OpenTelekomCloud HSS host (%s) status to become available.", hostId)
		stateConf := &resource.StateChangeConf{
			Pending:      []string{"PENDING"},
			Target:       []string{"COMPLETED"},
			Refresh:      hostStatusRefreshFunc(client, hostId),
			Timeout:      timeout,
			Delay:        30 * time.Second,
			PollInterval: 30 * time.Second,
		}
		unprotectedHostId, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("error waiting for the OpenTelekomCloud HSS host (%s) status to become completed: %s", hostId, err)
		}
		if unprotectedHostId != nil && unprotectedHostId.(string) != "" {
			unprotected = append(unprotected, unprotectedHostId.(string))
		}
	}
	return unprotected, nil
}

func hostStatusRefreshFunc(client *golangsdk.ServiceClient, hostId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var unprotectedHostId string
		hostList, err := group.ListHost(client, group.ListHostOpts{
			HostID: hostId,
		})
		if err != nil {
			return unprotectedHostId, "ERROR", err
		}
		if hostList == nil || len(hostList) < 1 {
			return unprotectedHostId, "PENDING", nil
		}
		if hostList[0].ProtectStatus == string(ProtectStatusClosed) {
			unprotectedHostId = hostList[0].ID
		}
		return unprotectedHostId, "COMPLETED", nil
	}
}

func queryHostGroups(client *golangsdk.ServiceClient, name string) ([]group.HostGroupResp, error) {
	groups, err := group.List(client, group.ListOpts{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching OpenTelekomCloud HSS host group: %s", err)
	}
	return groups, nil
}

func QueryHostGroupById(client *golangsdk.ServiceClient, groupId string) (*group.HostGroupResp, error) {
	allHostGroups, err := queryHostGroups(client, "")
	if err != nil {
		return nil, err
	}
	filter := map[string]interface{}{
		"ID": groupId,
	}
	result, err := common.FilterSliceWithField(allHostGroups, filter)
	if err != nil {
		return nil, fmt.Errorf("error filtering OpenTelekomCloud HSS host groups list: %s", err)
	}
	if len(result) < 1 {
		return nil, golangsdk.ErrDefault404{
			ErrUnexpectedResponseCode: golangsdk.ErrUnexpectedResponseCode{
				Body: []byte(fmt.Sprintf("the OpenTelekomCloud HSS host group (%s) does not exist", groupId)),
			},
		}
	}
	if item, ok := result[0].(group.HostGroupResp); ok {
		return &item, nil
	}
	return nil, fmt.Errorf("invalid OpenTelekomCloud HSS host group list, want 'group.HostGroupResp', but '%T'", result[0])
}

func resourceHostGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	g, err := QueryHostGroupById(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "OpenTelekomCloud HSS host group")
	}
	log.Printf("[DEBUG] The response of OpenTelekomCloud HSS host group is: %#v", g)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("name", g.Name),
		d.Set("host_ids", g.HostIds),
		d.Set("host_num", g.HostNum),
		d.Set("risk_host_num", g.RiskHostNum),
		d.Set("unprotect_host_num", g.UnprotectHostNum),
	)

	if len(d.Get("unprotect_host_ids").([]interface{})) == 0 {
		// The reason for writing an empty array to `unprotect_host_ids` is to avoid unexpected changes
		mErr = multierror.Append(mErr, d.Set("unprotect_host_ids", make([]string, 0)))
	}

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error saving OpenTelekomCloud HSS host group fields: %s", err)
	}
	return nil
}

func resourceHostGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	hostIds := common.ExpandToStringListBySet(d.Get("host_ids").(*schema.Set))

	opts := group.UpdateOpts{
		ID:      d.Id(),
		Name:    d.Get("name").(string),
		HostIds: hostIds,
	}

	unprotected, err := checkAllHostsAvailable(ctx, client, hostIds, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] All OpenTelekomCloud HSS hosts are availabile.")
	if len(unprotected) > 1 {
		err := d.Set("unprotect_host_ids", unprotected)
		if err != nil {
			log.Printf("[WARN] These OpenTelekomCloud HSS hosts are not protected: %#v", unprotected)
		}
	}
	err = group.Update(client, opts)
	if err != nil {
		return diag.Errorf("error updating OpenTelekomCloud HSS host group: %s", err)
	}

	clientCtx := common.CtxWithClient(ctx, client, hssClientV5)
	return resourceHostGroupRead(clientCtx, d, meta)
}

func resourceHostGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}
	err = group.Delete(client, group.DeleteOpts{GroupID: d.Id()})
	if err != nil {
		return common.CheckDeletedDiag(d, parseDeleteHostGroupResponseError(err), "error deleting OpenTelekomCloud HSS host group")
	}

	return nil
}

// When the host group does not exist, the response code for deleting the API is `400`,
// and the response body is as follows:
// {"status_code":400,"request_id":"f17e56c2e92584cfd4614ab467cd6a1b","error_code":"",
// "error_message":"{\"error_code\":\"00100090\",\"error_description\":\"Failed to load server groups.\"}",
// "encoded_authorization_message":""}
func parseDeleteHostGroupResponseError(err error) error {
	var errObj map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(err.Error()), &errObj); jsonErr != nil {
		log.Printf("[WARN] failed to unmarshal error object: %s", jsonErr)
		return err
	}

	statusCode, parseStatusCodeErr := jmespath.Search("status_code", errObj)
	if parseStatusCodeErr != nil || statusCode == nil {
		log.Printf("[WARN] failed to parse status_code from response body: %s", parseStatusCodeErr)
		return err
	}

	if statusCodeFloat, ok := statusCode.(float64); ok && int(statusCodeFloat) == 400 {
		errorMessage, parseErrorMessageErr := jmespath.Search("error_message", errObj)
		if parseErrorMessageErr != nil || errorMessage == nil {
			log.Printf("[WARN] failed to parse error_message: %s", parseErrorMessageErr)
			return err
		}

		var errMsgObj map[string]interface{}
		if errMsgJson := json.Unmarshal([]byte(errorMessage.(string)), &errMsgObj); errMsgJson != nil {
			log.Printf("[WARN] failed to unmarshal error_message: %s", errMsgJson)
			return err
		}

		errorCode, errorCodeErr := jmespath.Search("error_code", errMsgObj)
		if errorCodeErr != nil || errorCode == nil {
			log.Printf("[WARN] failed to extract error_code: %s", errorCodeErr)
			return err
		}

		if errorCode == "00100090" {
			return golangsdk.ErrDefault404{}
		}
	}

	return err
}
