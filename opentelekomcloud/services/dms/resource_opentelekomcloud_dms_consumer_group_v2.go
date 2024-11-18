package dms

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/instances/management"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDmsConsumerGroupV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsConsumerGroupV2Create,
		ReadContext:   resourceDmsConsumerGroupV2Read,
		DeleteContext: resourceDmsConsumerGroupV2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: common.ImportByPath("instance_id", "group_name"),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assignment_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"coordinator_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"member_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assignments": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"topic": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"partitions": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
								},
							},
						},
					},
				},
			},
			"group_message_offsets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"partition": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"lag": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"topic": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message_current_offset": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"message_log_end_offset": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDmsConsumerGroupV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dmsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	instanceId := d.Get("instance_id").(string)

	createConsumerGroupOpts := management.CreateConsumerGroupOpts{
		GroupName:   d.Get("group_name").(string),
		Description: d.Get("description").(string),
	}

	err = management.CreateConsumerGroup(client, instanceId, createConsumerGroupOpts)
	if err != nil {
		return diag.Errorf("error creating consumer group for Kafka instance: %s", err)
	}

	group_id := fmt.Sprintf("%s/%s", instanceId, createConsumerGroupOpts.GroupName)
	d.SetId(group_id)

	clientCtx := common.CtxWithClient(ctx, client, dmsClientV2)
	return resourceDmsConsumerGroupV2Read(clientCtx, d, meta)
}

func resourceDmsConsumerGroupV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dmsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	instanceId := d.Get("instance_id").(string)
	groupName := d.Get("group_name").(string)

	getResp, err := management.GetConsumerGroup(client, instanceId, groupName)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "DMS Kafka instance or consumer group not found")
	}

	consumerGroup := getResp.Group

	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("instance_id", instanceId),
		d.Set("group_name", consumerGroup.GroupId),
		d.Set("state", consumerGroup.State),
		d.Set("coordinator_id", consumerGroup.CoordinatorId),
		d.Set("assignment_strategy", consumerGroup.AssignmentStrategy),
	)

	var memberList []map[string]interface{}
	for _, memberRaw := range consumerGroup.Members {
		member := make(map[string]interface{})
		member["host"] = memberRaw.Host
		member["member_id"] = memberRaw.MemberId
		member["client_id"] = memberRaw.ClientId
		var assignmentList []map[string]interface{}
		for _, assignmentRaw := range memberRaw.Assignment {
			assignment := make(map[string]interface{})
			assignment["topic"] = assignmentRaw.Topic
			assignment["partitions"] = assignmentRaw.Partitions
			assignmentList = append(assignmentList, assignment)
		}
		member["assignments"] = assignmentList
		memberList = append(memberList, member)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("members", memberList),
	)

	var groupMessageOffsets []map[string]interface{}
	for _, groupMessageOffsetRaw := range consumerGroup.GroupMessageOffsets {
		groupMessageOffset := map[string]interface{}{
			"partition":              groupMessageOffsetRaw.Partition,
			"lag":                    groupMessageOffsetRaw.Lag,
			"topic":                  groupMessageOffsetRaw.Topic,
			"message_current_offset": groupMessageOffsetRaw.MessageCurrentOffset,
			"message_log_end_offset": groupMessageOffsetRaw.MessageLogEndOffset,
		}

		groupMessageOffsets = append(groupMessageOffsets, groupMessageOffset)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("group_message_offsets", groupMessageOffsets),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceDmsConsumerGroupV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dmsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	instanceId := d.Get("instance_id").(string)
	groupName := d.Get("group_name").(string)

	err = management.DeleteConsumerGroup(client, instanceId, groupName)
	if err != nil {
		return diag.Errorf("error deleting DMSv2 consumer group: %v", err)
	}

	d.SetId("")
	log.Printf("[DEBUG] DMS Kafka instance consumer '%s' group has been deleted", groupName)
	return nil
}
