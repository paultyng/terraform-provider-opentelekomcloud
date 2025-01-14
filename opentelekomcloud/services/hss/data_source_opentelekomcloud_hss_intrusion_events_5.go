package hss

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/event"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceEvents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEventsRead,

		Schema: map[string]*schema.Schema{
			"category": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Event category. Its value can be: host (host security event) or container (container security event).",
			},
			"enterprise_project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Enterprise project ID. The value 0 indicates the default enterprise project. To query all enterprise projects, set this parameter to all_granted_eps.",
			},
			"days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of days to be queried. This parameter is mutually exclusive with begin_time and end_time.",
			},
			"host_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server name.",
			},
			"host_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Host ID.",
			},
			"private_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server IP address.",
			},
			"container_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Container instance name.",
			},
			"event_types": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Intrusion types. Possible values include:\n1001: Malware\n1010: Rootkit\n1011: Ransomware\n1015: Web shell\n1017: Reverse shell\n2001: Common vulnerability exploit\n3002: File privilege escalation\n3003: Process privilege escalation\n3004: Important file change\n3005: File/Directory change\n3007: Abnormal process behavior\n3015: High-risk command execution\n3018: Abnormal shell\n3027: Suspicious crontab tasks\n4002: Brute-force attack\n4004: Abnormal login\n4006: Invalid system account.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"handle_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Status. Possible values: unhandled, handled.",
			},
			"severity": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Threat level. Possible values: Security, Low, Medium, High, Critical.",
			},
			"begin_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Customized start time of a segment. The timestamp is accurate to seconds. The begin_time should be no more than two days earlier than the end_time. This parameter is mutually exclusive with the queried duration.",
			},
			"end_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Customized end time of a segment. The timestamp is accurate to seconds. The begin_time should be no more than two days earlier than the end_time. This parameter is mutually exclusive with the queried duration.",
			},
			"events": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of events returned from the query.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Event ID.",
						},
						"event_class_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Event category identifier.",
						},
						"event_type": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Intrusion type.",
						},
						"event_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Event name.",
						},
						"severity": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Threat level.",
						},
						"container_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Container instance name. Available only for container alarms.",
						},
						"image_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Image name. Available only for container alarms.",
						},
						"host_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Server name.",
						},
						"host_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Host ID.",
						},
						"private_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Server private IP address.",
						},
						"public_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Elastic IP address.",
						},
						"os_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OS type (Linux/Windows).",
						},
						"host_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Server status.",
						},
						"agent_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Agent status.",
						},
						"protect_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Protection status.",
						},
						"asset_value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Asset importance.",
						},
						"attack_phase": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Attack phase.",
						},
						"attack_tag": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Attack tag.",
						},
						"occur_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Occurrence time, accurate to milliseconds.",
						},
						"handle_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Processing status (unhandled/handled).",
						},
						"handle_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Handling time, in milliseconds.",
						},
						"handle_method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Handling method.",
						},
						"handler": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Remarks. Available only for handled alarms.",
						},
						"operate_accept_list": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Supported processing operation.",
						},
						"recommendation": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Handling suggestions.",
						},
						"event_details": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Brief description of the event.",
						},
						"operate_detail_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of operation details associated with the event.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"agent_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Agent ID.",
									},
									"process_pid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Process ID.",
									},
									"is_parent": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indicates whether the process is a parent process.",
									},
									"file_hash": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File hash.",
									},
									"file_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Path to the file.",
									},
									"file_attr": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File attribute.",
									},
									"private_ip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Server private IP address.",
									},
									"login_ip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Login source IP address.",
									},
									"login_user_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Login username.",
									},
									"keyword": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Alarm event keyword, which is used only for the alarm whitelist.",
									},
									"hash": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Alarm event hash, which is used only for the alarm whitelist.",
									},
								},
							},
						},
						"resource_info": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Information about the resource associated with the event.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "User account ID.",
									},
									"project_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Project ID.",
									},
									"enterprise_project_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Enterprise project ID.",
									},
									"region_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Region name.",
									},
									"vpc_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "VPC ID.",
									},
									"ecs_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ECS ID.",
									},
									"vm_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "VM name.",
									},
									"vm_uuid": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Specifies the VM UUID, that is, the server ID.",
									},
									"container_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Container ID.",
									},
									"image_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Image ID.",
									},
									"image_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Image name.",
									},
									"host_attr": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Host attribute.",
									},
									"service": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Service.",
									},
									"microservice": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Microservice.",
									},
									"sys_arch": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "System CPU architecture.",
									},
									"os_bit": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "OS bit version.",
									},
									"os_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "OS type.",
									},
									"os_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "OS name.",
									},
									"os_version": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "OS version.",
									},
								},
							},
						},
						"process_info_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of process information associated with the event.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"process_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Process name.",
									},
									"process_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Process file path.",
									},
									"process_pid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Process ID.",
									},
									"process_uid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Process user ID.",
									},
									"process_username": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Process username.",
									},
									"process_cmdline": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Command line used to start the process.",
									},
									"process_filename": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Process file name.",
									},
									"process_start_time": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Process start time.",
									},
									"process_gid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Process group ID.",
									},
									"process_egid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Valid process group ID.",
									},
									"process_euid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Valid process user ID.",
									},
									"parent_process_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Parent process name.",
									},
									"parent_process_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Parent process file path.",
									},
									"parent_process_pid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Parent process ID.",
									},
									"parent_process_uid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Parent process user ID.",
									},
									"parent_process_cmdline": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Parent process file command line.",
									},
									"parent_process_filename": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Parent process file name.",
									},
									"parent_process_start_time": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Parent process file name.",
									},
									"parent_process_gid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Parent process group ID.",
									},
									"parent_process_egid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Valid parent process group ID.",
									},
									"parent_process_euid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Valid parent process user ID.",
									},
									"child_process_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Subprocess name.",
									},
									"child_process_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Subprocess file path.",
									},
									"child_process_pid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Subprocess ID.",
									},
									"child_process_uid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Subprocess user ID.",
									},
									"child_process_cmdline": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Subprocess file command line.",
									},
									"child_process_filename": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Subprocess file name.",
									},
									"child_process_start_time": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Subprocess start time.",
									},
									"child_process_gid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Subprocess group ID.",
									},
									"child_process_egid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Valid subprocess group ID.",
									},
									"child_process_euid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Valid subprocess user ID.",
									},
									"virt_cmd": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Virtualization command.",
									},
									"virt_process_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Virtualization process name.",
									},
									"escape_mode": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Escape mode.",
									},
									"escape_cmd": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Commands executed after escape.",
									},
									"process_hash": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Process startup file hash.",
									},
								},
							},
						},
						"user_info_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of user information associated with the event.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"user_id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "User UID.",
									},
									"user_gid": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "User GID.",
									},
									"user_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Username.",
									},
									"user_group_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "User group name.",
									},
									"user_home_dir": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "User home directory.",
									},
									"login_ip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "User login IP address.",
									},
									"service_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of service used for login.",
									},
									"service_port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Login service port.",
									},
									"login_mode": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Login mode.",
									},
									"login_last_time": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Last login time.",
									},
									"login_fail_count": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Number of failed login attempts.",
									},
									"pwd_hash": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Password hash.",
									},
									"pwd_with_fuzzing": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Masked password.",
									},
									"pwd_used_days": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Number of days the current password has been in use.",
									},
									"pwd_min_days": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Minimum password validity period.",
									},
									"pwd_max_days": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum password validity period.",
									},
									"pwd_warn_left_days": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Advance warning of password expiration (days).",
									},
								},
							},
						},
						"file_info_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of file information associated with the event.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"file_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Path to the file.",
									},
									"file_alias": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File alias.",
									},
									"file_size": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Size of the file in bytes.",
									},
									"file_mtime": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Time when a file was last modified.",
									},
									"file_atime": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Time when a file was last accessed.",
									},
									"file_ctime": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Time when the status of a file was last changed.",
									},
									"file_hash": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Hash of the file.",
									},
									"file_md5": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File MD5.",
									},
									"file_sha256": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File SHA256.",
									},
									"file_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of the file.",
									},
									"file_content": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File content.",
									},
									"file_attr": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File attribute.",
									},
									"file_operation": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File operation type.",
									},
									"file_action": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Action performed on the file.",
									},
									"file_change_attr": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Old/New attribute.",
									},
									"file_new_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "New file path.",
									},
									"file_desc": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File description.",
									},
									"file_key_word": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File keyword.",
									},
									"is_dir": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Whether it is a directory.",
									},
									"fd_info": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "File handle information.",
									},
									"fd_count": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Number of file handles.",
									},
								},
							},
						},
					},
				},
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Region where event is occurred.",
			},
		},
	}
}

func dataSourceEventsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, hssClientV5, func() (*golangsdk.ServiceClient, error) {
		return config.HssV5Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV5Client, err)
	}

	opts := event.ListOpts{
		Limit:               200,
		HostName:            d.Get("host_name").(string),
		HostID:              d.Get("host_id").(string),
		Category:            d.Get("category").(string),
		EnterpriseProjectId: d.Get("enterprise_project_id").(string),
		Days:                d.Get("days").(int),
		PrivateIP:           d.Get("private_ip").(string),
		ContainerName:       d.Get("container_name").(string),
		EventTypes:          common.ExpandToStringSlice(d.Get("event_types").(*schema.Set).List()),
		HandleStatus:        d.Get("handle_status").(string),
		Severity:            d.Get("severity").(string),
		BeginTime:           d.Get("begin_time").(string),
		EndTime:             d.Get("end_time").(string),
	}
	allEvents, err := event.List(client, opts)
	if err != nil {
		return diag.Errorf("unable to list OpenTelekomCloud HSS intrusion events: %s", err)
	}

	if len(allEvents) == 0 {
		log.Printf("[DEBUG] No intrusion events in OpenTelekomCloud found")
	}

	uuId, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("unable to generate ID: %s", err)
	}

	d.SetId(uuId)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("events", flattenEvents(allEvents)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenEvents(events []event.EventResp) []interface{} {
	if len(events) == 0 {
		return nil
	}

	rst := make([]interface{}, 0, len(events))
	for _, e := range events {
		eventMap := map[string]interface{}{
			"id":                  e.ID,
			"event_class_id":      e.EventClassId,
			"event_type":          e.EventType,
			"event_name":          e.EventName,
			"severity":            e.Severity,
			"container_name":      e.ContainerName,
			"image_name":          e.ImageName,
			"host_name":           e.HostName,
			"host_id":             e.HostId,
			"private_ip":          e.PrivateIP,
			"public_ip":           e.PublicIP,
			"os_type":             e.OsType,
			"host_status":         e.HostStatus,
			"agent_status":        e.AgentStatus,
			"protect_status":      e.ProtectStatus,
			"asset_value":         e.AssetValue,
			"attack_phase":        e.AttackPhase,
			"attack_tag":          e.AttackTag,
			"occur_time":          e.OccurrenceTime,
			"handle_time":         e.HandleTime,
			"handle_status":       e.HandleStatus,
			"handle_method":       e.HandleMethod,
			"handler":             e.Handler,
			"operate_accept_list": e.OperateAcceptList,
			"recommendation":      e.Recommendation,
			"event_details":       e.EventDetails,
		}

		// Flatten nested structs
		var resList []map[string]interface{}
		if e.ResourceInfo != nil {
			resList = append(resList, map[string]interface{}{
				"domain_id":             e.ResourceInfo.DomainId,
				"project_id":            e.ResourceInfo.ProjectId,
				"enterprise_project_id": e.ResourceInfo.EnterpriseProjectId,
				"region_name":           e.ResourceInfo.RegionName,
				"vpc_id":                e.ResourceInfo.VpcId,
				"ecs_id":                e.ResourceInfo.EcsId,
				"vm_name":               e.ResourceInfo.VmName,
				"vm_uuid":               e.ResourceInfo.VmUuid,
				"container_id":          e.ResourceInfo.ContainerId,
				"image_id":              e.ResourceInfo.ImageId,
				"image_name":            e.ResourceInfo.ImageName,
				"host_attr":             e.ResourceInfo.HostAttr,
				"service":               e.ResourceInfo.Service,
				"microservice":          e.ResourceInfo.Microservice,
				"sys_arch":              e.ResourceInfo.Arch,
				"os_bit":                e.ResourceInfo.OsBit,
				"os_type":               e.ResourceInfo.OsType,
				"os_name":               e.ResourceInfo.OsName,
				"os_version":            e.ResourceInfo.OsVersion,
			})
		}
		eventMap["resource_info"] = resList

		// Flatten lists
		var processList []map[string]interface{}
		for _, p := range e.ProcessInfoList {
			processList = append(processList, map[string]interface{}{
				"process_name":              p.ProcessName,
				"process_path":              p.ProcessPath,
				"process_pid":               p.ProcessPid,
				"process_uid":               p.ProcessUid,
				"process_username":          p.ProcessUsername,
				"process_cmdline":           p.ProcessCmdline,
				"process_filename":          p.ProcessFilename,
				"process_start_time":        p.ProcessStartTime,
				"process_gid":               p.ProcessGid,
				"process_egid":              p.ProcessEgid,
				"process_euid":              p.ProcessEuid,
				"parent_process_name":       p.ParentProcessName,
				"parent_process_path":       p.ParentProcessPath,
				"parent_process_pid":        p.ParentProcessPid,
				"parent_process_uid":        p.ParentProcessUid,
				"parent_process_cmdline":    p.ParentProcessCmdline,
				"parent_process_filename":   p.ParentProcessFilename,
				"parent_process_start_time": p.ParentProcessStartTime,
				"child_process_name":        p.ChildProcessName,
				"child_process_path":        p.ChildProcessPath,
				"child_process_pid":         p.ChildProcessPid,
				"child_process_uid":         p.ChildProcessUid,
				"child_process_cmdline":     p.ChildProcessCmdline,
				"child_process_filename":    p.ChildProcessFilename,
				"child_process_start_time":  p.ChildProcessStartTime,
				"virt_cmd":                  p.VirtCmd,
				"virt_process_name":         p.VirtProcessName,
				"escape_mode":               p.EscapeMode,
				"escape_cmd":                p.EscapeCmd,
				"process_hash":              p.ProcessHash,
			})
		}
		eventMap["process_info_list"] = processList

		var userList []map[string]interface{}
		for _, u := range e.UserInfoList {
			userList = append(userList, map[string]interface{}{
				"user_id":            u.UserId,
				"user_gid":           u.UserGid,
				"user_name":          u.UserName,
				"user_group_name":    u.UserGroupName,
				"user_home_dir":      u.UserHomeDir,
				"login_ip":           u.LoginIP,
				"service_type":       u.ServiceType,
				"service_port":       u.ServicePort,
				"login_mode":         u.LoginMode,
				"login_last_time":    u.LoginLastTime,
				"login_fail_count":   u.LoginFailCount,
				"pwd_hash":           u.PwdHash,
				"pwd_with_fuzzing":   u.PwdWithFuzzing,
				"pwd_used_days":      u.PwdUsedDays,
				"pwd_min_days":       u.PwdMinDays,
				"pwd_max_days":       u.PwdMaxDays,
				"pwd_warn_left_days": u.PwdWarnLeftDays,
			})
		}
		eventMap["user_info_list"] = userList

		var fileList []map[string]interface{}
		for _, f := range e.FileInfoList {
			fileList = append(fileList, map[string]interface{}{
				"file_path":        f.FilePath,
				"file_alias":       f.FileAlias,
				"file_size":        f.FileSize,
				"file_mtime":       f.FileMtime,
				"file_atime":       f.FileAtime,
				"file_ctime":       f.FileCtime,
				"file_hash":        f.FileHash,
				"file_md5":         f.FileMd5,
				"file_sha256":      f.FileSha256,
				"file_type":        f.FileType,
				"file_content":     f.FileContent,
				"file_attr":        f.FileAttr,
				"file_operation":   f.FileOperation,
				"file_action":      f.FileAction,
				"file_change_attr": f.FileChangeAttr,
				"file_new_path":    f.FileNewPath,
				"file_desc":        f.FileDesc,
				"file_key_word":    f.FileKeyWord,
				"is_dir":           f.IsDir,
				"fd_info":          f.FdInfo,
				"fd_count":         f.FdCount,
			})
		}
		eventMap["file_info_list"] = fileList

		rst = append(rst, eventMap)
	}

	return rst
}
