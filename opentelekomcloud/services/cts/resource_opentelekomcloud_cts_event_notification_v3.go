package cts

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cts/v3/keyevent"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceCTSEventNotificationV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCTSEventNotificationCreate,
		ReadContext:   resourceCTSEventNotificationRead,
		UpdateContext: resourceCTSEventNotificationUpdate,
		DeleteContext: resourceCTSEventNotificationDelete,

		CustomizeDiff: validateSchema,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"notification_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common.ValidateCTSEventName,
			},
			"operation_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"complete", "customized",
				}, false),
			},
			"operations": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trace_names": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					}},
			},
			"notify_user_list": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_group": {
							Type:     schema.TypeString,
							Required: true,
						},
						"user_list": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					}},
			},
			"topic_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rule": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"notification_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"notification_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func resourceCTSOperationsV3(d *schema.ResourceData) []keyevent.Operations {
	operations := d.Get("operations").([]interface{})
	if len(operations) == 0 {
		return nil
	}
	var keyEvents []keyevent.Operations

	for _, v := range operations {
		operation := v.(map[string]interface{})

		var refinedTraces []string

		for _, s := range operation["trace_names"].([]interface{}) {
			refinedTraces = append(refinedTraces, s.(string))
		}

		keyEvent := keyevent.Operations{
			ServiceType:  operation["service_type"].(string),
			ResourceType: operation["resource_type"].(string),
			TraceNames:   refinedTraces,
		}
		keyEvents = append(keyEvents, keyEvent)
	}

	return keyEvents
}

func resourceCTSUserListV3(d *schema.ResourceData) []keyevent.NotificationUsers {
	userList := d.Get("notify_user_list").([]interface{})
	if len(userList) == 0 {
		return nil
	}

	var usersOpts []keyevent.NotificationUsers

	for _, v := range userList {
		user := v.(map[string]interface{})

		var refinedUserList []string

		for _, s := range user["user_list"].([]interface{}) {
			refinedUserList = append(refinedUserList, s.(string))
		}

		userOpts := keyevent.NotificationUsers{
			UserGroup: user["user_group"].(string),
			UserList:  refinedUserList,
		}
		usersOpts = append(usersOpts, userOpts)
	}

	return usersOpts
}

func buildKeyFilterOpts(d *schema.ResourceData) (*keyevent.Filter, error) {
	rawFilter := d.Get("filter").([]interface{})
	if len(rawFilter) == 0 {
		return nil, nil
	}

	filterData := rawFilter[0].(map[string]interface{})
	filter := keyevent.Filter{
		IsSupportFilter: true,
		Rule:            common.ExpandToStringList(filterData["rule"].([]interface{})),
	}

	conditionStr := filterData["condition"].(string)
	switch conditionStr {
	case "AND":
		filter.Condition = "AND"
	case "OR":
		filter.Condition = "OR"
	default:
		return &filter, fmt.Errorf("invalid condition, want 'AND' or 'OR', but got '%v'", conditionStr)
	}

	log.Printf("[DEBUG] CTS key events notification filter: %#v", filter)
	return &filter, nil
}

func resourceCTSEventNotificationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CtsV3Client(config.GetProjectName(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	createOpts := keyevent.CreateNotificationOpts{
		NotificationName: d.Get("notification_name").(string),
		OperationType:    d.Get("operation_type").(string),
		Operations:       resourceCTSOperationsV3(d),
		NotifyUserList:   resourceCTSUserListV3(d),
		TopicId:          d.Get("topic_id").(string),
	}

	filter, err := buildKeyFilterOpts(d)
	if err != nil {
		return diag.FromErr(err)
	}
	createOpts.Filter = filter

	ctsN, err := keyevent.Create(client, createOpts)
	if err != nil {
		return fmterr.Errorf("error creating CTS event notification: %w", err)
	}

	d.SetId(buildEventNotificationID(ctsN.NotificationId, ctsN.NotificationName))

	if status, ok := d.GetOk("status"); ok && createOpts.TopicId != "" {
		if status.(string) == "disabled" {
			return resourceCTSEventNotificationUpdate(ctx, d, meta)
		}
	}

	return resourceCTSEventNotificationRead(ctx, d, meta)
}

func resourceCTSEventNotificationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CtsV3Client(config.GetProjectName(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	_, eventName := ExtractNotificationID(d.Id())

	ctsNotification, err := keyevent.List(client, keyevent.ListNotificationsOpts{
		NotificationType: "smn",
		NotificationName: eventName,
	})
	if err != nil {
		return fmterr.Errorf("error retrieving cts event notification: %w", err)
	}

	operations := buildOperationsSet(ctsNotification[0].Operations)
	userList := buildUserList(ctsNotification[0].NotifyUserList)

	mErr := multierror.Append(
		d.Set("notification_id", ctsNotification[0].NotificationId),
		d.Set("notification_name", ctsNotification[0].NotificationName),
		d.Set("notification_type", ctsNotification[0].NotificationType),
		d.Set("operation_type", ctsNotification[0].OperationType),
		d.Set("operations", operations),
		d.Set("notify_user_list", userList),
		d.Set("topic_id", ctsNotification[0].TopicId),
		d.Set("project_id", ctsNotification[0].ProjectId),
		d.Set("create_time", ctsNotification[0].CreateTime),
		d.Set("status", ctsNotification[0].Status),
		d.Set("filter", flattenNotificationFilter(&ctsNotification[0].Filter)),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting CTS event notifications fields: %w", err)
	}

	return nil
}

func resourceCTSEventNotificationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CtsV3Client(config.GetProjectName(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	needsUpdate := false
	updateOpts := keyevent.UpdateNotificationOpts{}

	if d.HasChange("operations") {
		updateOpts.Operations = resourceCTSOperationsV3(d)
		needsUpdate = true
	}
	if d.HasChange("notify_user_list") {
		updateOpts.NotifyUserList = resourceCTSUserListV3(d)
		needsUpdate = true
	}
	if d.HasChange("topic_id") {
		updateOpts.TopicId = d.Get("topic_id").(string)
		needsUpdate = true
	}

	eventId, _ := ExtractNotificationID(d.Id())

	if d.HasChange("filter") {
		filter, err := buildKeyFilterOpts(d)
		if err != nil {
			return diag.FromErr(err)
		}
		updateOpts.Filter = filter
		needsUpdate = true
	}

	if needsUpdate || d.HasChange("notification_name") || d.HasChange("operation_type") || d.HasChange("status") {
		updateOpts.NotificationId = eventId
		updateOpts.NotificationName = d.Get("notification_name").(string)
		updateOpts.OperationType = d.Get("operation_type").(string)
		updateOpts.Status = d.Get("status").(string)
		if updateOpts.Status == "enabled" {
			updateOpts.TopicId = d.Get("topic_id").(string)
		}
		if updateOpts.OperationType == "customized" {
			updateOpts.Operations = resourceCTSOperationsV3(d)
			updateOpts.NotifyUserList = resourceCTSUserListV3(d)
		}
	}

	ctsN, err := keyevent.Update(client, updateOpts)
	if err != nil {
		return fmterr.Errorf("error updating CTS event notification: %w", err)
	}
	d.SetId(buildEventNotificationID(ctsN.NotificationId, ctsN.NotificationName))

	return resourceCTSEventNotificationRead(ctx, d, meta)
}

func resourceCTSEventNotificationDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CtsV3Client(config.GetProjectName(d))
	if err != nil {
		return fmterr.Errorf(clientError, err)
	}

	eventID, _ := ExtractNotificationID(d.Id())

	deleteOpts := keyevent.DeleteOpts{
		NotificationId: []string{
			eventID,
		},
	}

	if err := keyevent.Delete(client, deleteOpts); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Successfully deleted cts event notification %s", d.Id())

	d.SetId("")

	return nil
}

func buildOperationsSet(eventResponse []keyevent.Operations) []interface{} {
	var operations []interface{}

	for _, operation := range eventResponse {
		traces := make([]string, len(operation.TraceNames))
		copy(traces, operation.TraceNames)

		refinedOperation := map[string]interface{}{
			"service_type":  operation.ServiceType,
			"resource_type": operation.ResourceType,
			"trace_names":   traces,
		}
		operations = append(operations, refinedOperation)
	}

	return operations
}

func buildUserList(notifyUsersList []keyevent.NotificationUsers) []interface{} {
	var refinedUsers []interface{}

	for _, notifyUser := range notifyUsersList {
		users := make([]string, len(notifyUser.UserList))
		copy(users, notifyUser.UserList)

		refinedUser := map[string]interface{}{
			"user_group": notifyUser.UserGroup,
			"user_list":  users,
		}
		refinedUsers = append(refinedUsers, refinedUser)
	}

	return refinedUsers
}

func buildEventNotificationID(eventID, eventName string) string {
	return fmt.Sprintf("%s/%s", eventID, eventName)
}

func ExtractNotificationID(id string) (string, string) {
	split := strings.Split(id, "/")
	return split[0], split[1]
}

func validateSchema(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	operationType := d.Get("operation_type").(string)

	if _, ok := d.GetOk("operations"); ok && operationType == "complete" {
		return fmt.Errorf("customized operations can't be used with `complete` operation_type")
	}

	return nil
}

func flattenNotificationFilter(filter *keyevent.Filter) []map[string]interface{} {
	if filter == nil {
		return nil
	}
	result := map[string]interface{}{
		"condition": filter.Condition,
		"rule":      filter.Rule,
	}

	return []map[string]interface{}{result}
}
