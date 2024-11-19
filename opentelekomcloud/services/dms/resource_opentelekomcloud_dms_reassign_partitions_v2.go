package dms

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/instances/management"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDmsReassignPartitionsV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsReassignPartitionsV2Create,
		ReadContext:   resourceReassignPartitionsV2Read,
		DeleteContext: resourceReassignPartitionsV2Delete,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"reassignments": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topic": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"brokers": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"replication_factor": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"assignments": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"partition": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"partition_brokers": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
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
			"throttle": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"is_schedule": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"execute_at": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"time_estimate": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reassignment_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceDmsReassignPartitionsV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, dmsClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	instanceId := d.Get("instance_id").(string)

	createReassignPartitionOpts := management.InitPartitionReassigningOpts{
		Reassignments: getPartitionReassignments(d),
		Throttle:      d.Get("throttle").(int),
		IsSchedule:    d.Get("is_schedule").(bool),
		TimeEstimate:  d.Get("time_estimate").(bool),
		ExecuteAt:     int64(d.Get("execute_at").(int)),
	}

	initResp, err := management.InitPartitionReassigning(client, instanceId, &createReassignPartitionOpts)
	if err != nil {
		return diag.Errorf("error submitting partition reassignment for Kafka instance: %s", err)
	}

	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("instance_id", instanceId),
		d.Set("reassignment_time", initResp.ReassignmentTime),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceReassignPartitionsV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	instanceId := d.Get("instance_id").(string)
	mErr := multierror.Append(
		nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("instance_id", instanceId),
		d.Set("reassignment_time", d.Get("reassignment_time").(string)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceReassignPartitionsV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	errorMsg := "Deleting resource is not supported. The resource is only removed from the state, the task remains in the cloud."
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  errorMsg,
		},
	}
}

func getPartitionReassignments(d *schema.ResourceData) []management.PartitionReassign {
	var reassignments []management.PartitionReassign

	reassignmentsRaw := d.Get("reassignments").([]interface{})
	for i := range reassignmentsRaw {
		reassignmentRaw := reassignmentsRaw[i].(map[string]interface{})
		reassignment := management.PartitionReassign{}
		reassignment.Topic = reassignmentRaw["topic"].(string)
		reassignment.Brokers = fetchIntegerLists(reassignmentRaw["brokers"].([]interface{}))
		reassignment.ReplicationFactor = reassignmentRaw["replication_factor"].(int)
		var assignments []*management.TopicAssignment
		assignmentsRaw := reassignmentRaw["assignments"].([]interface{})
		for j := range assignmentsRaw {
			assignmentRaw := assignmentsRaw[j].(map[string]interface{})
			assignment := management.TopicAssignment{
				Partition:        assignmentRaw["partition"].(int),
				PartitionBrokers: fetchIntegerLists(assignmentRaw["partition_brokers"].([]interface{})),
			}
			assignments = append(assignments, &assignment)
		}
		reassignment.Assignment = assignments
		reassignments = append(reassignments, reassignment)
	}
	return reassignments
}

func fetchIntegerLists(listRaw []interface{}) []int {
	processedList := make([]int, 0)
	for _, v := range listRaw {
		processedList = append(processedList, v.(int))
	}
	return processedList
}
