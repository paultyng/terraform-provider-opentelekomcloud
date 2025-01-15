---
subcategory: “Host Security Service (HSS)”
layout: “opentelekomcloud”
page_title: “OpenTelekomCloud: opentelekomcloud_hss_intrusion_events_v5”
sidebar_current: “docs-opentelekomcloud-datasource-hss-intrusion-events-v5”
description: |-
  Use this data source to query HSS events in OpenTelekomCloud, including host and container security events.
---

Up-to-date reference of API arguments for HSS events can be found at the
[documentation portal](https://docs.otc.t-systems.com/host-security-service/api-ref/api_description/intrusion_detection/querying_the_detected_intrusion_list.html#)

# opentelekomcloud_hss_intrusion_events_v5

Use this data source to query HSS events, such as intrusion detections, malware alerts, or suspicious activities, within OpenTelekomCloud.

## Example Usage

```hcl
variable "event_category" {}

data "opentelekomcloud_hss_intrusion_events_v5" "events" {
  category = var.event_category
  days     = 7
}
```

## Argument Reference

The following arguments are supported:

* `category` - (Required, String) Specifies the category of the event. Valid values are:
  *	`host` - Host security events.
  *	`container` - Container security events.

* `enterprise_project_id` - (Optional, String) Specifies the enterprise project ID. Use 0 for the default project or all_granted_eps to query all projects.

* `days` - (Optional, Integer) Specifies the number of days to query events. This parameter is mutually exclusive with begin_time and end_time.

* `host_name` - (Optional, String) Specifies the name of the server to query.

* `host_id` - (Optional, String) Specifies the ID of the host to query.

* `private_ip` - (Optional, String) Specifies the private IP address of the server.

* `container_name` - (Optional, String) Specifies the name of the container instance to query.

* `event_types` - (Optional, Set of Strings) Specifies the types of intrusion events to query. Possible values include but are not limited to:
  *	`1001` - Malware.
  *	`1010` - Rootkit.
  *	`1015` - Web shell.
  *	`3015` - High-risk command execution.
  *	`4002` - Brute-force attack.

* `handle_status` - (Optional, String) Specifies the status of the event. Valid values are:
  *	`unhandled`
  *	`handled`

* `severity` - (Optional, String) Specifies the threat level. Valid values are:
  *	`Security`
  *	`Low`
  *	`Medium`
  *	`High`
  *	`Critical`

* `begin_time` - (Optional, String) Specifies the start time for querying events in ISO 8601 format. This is mutually exclusive with days.

* `end_time` - (Optional, String) Specifies the end time for querying events in ISO 8601 format. This is mutually exclusive with days.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID in UUID format.

* `events` - A list of events that match the query parameters. Each event has the following attributes:
  * `id` - Event ID.
  *	`event_class_id` - Event category identifier.
  *	`event_type` - Intrusion type identifier.
  *	`event_name` - Event name.
  * `severity` - Threat level.
  *	`host_name` - Name of the host associated with the event.
  * `host_id` - Host ID.
  * `private_ip` - Host private IP.
  * `public_ip` - Host public IP (if available).
  * `occur_time` - Time of event occurrence in milliseconds.
  * `handle_status` - Processing status of the event (unhandled or handled).
  *	`handle_time` - Handling time in milliseconds (if applicable).
  *	`recommendation` - Recommended action for the event.
  *	`event_details` - Brief description of the event.
  *	`region` - Region where the event occurred.
  * `operate_detail_list` - List of operation details associated with the event.
    * The [operate_detail_list](#hss_operate_detail_list) structure is documented below.
  * `resource_info` - Information about the resource associated with the event.
    * The [resource_info](#hss_resource_info) structure is documented below.
  * `process_info_list` - List of process information associated with the event.
    * The [process_info_list](#hss_process_info_list) structure is documented below.
  * `user_info_list` - List of user information associated with the event.
    * The [user_info_list](#hss_user_info_list) structure is documented below.
  * `file_info_list` - List of file information associated with the event.
    * The [file_info_list](#hss_file_info_list) structure is documented below.

* `region` - Region where the event occurred.

<a name="hss_operate_detail_list"></a>
The `operate_detail_list` block supports:

* `agent_id` - Agent ID.
* `process_pid` - Process ID.
* `is_parent` - Indicates whether the process is a parent process.
* `file_hash` - File hash.
* `file_path` - Path to the file.
* `file_attr` - File attribute.
* `private_ip` - Server private IP address.
* `login_ip` - Login source IP address.
* `login_user_name` - Login username.
* `keyword` - Alarm event keyword.
* `hash` - Alarm event hash.

<a name="hss_resource_info"></a>
The `resource_info` block supports:
* `domain_id` - User account ID.
* `project_id` - Project ID.
* `enterprise_project_id` - Enterprise project ID.
* `region_name` - Region name.
* `vpc_id` - VPC ID.
* `ecs_id` - ECS ID.
* `vm_name` - VM name.
* `vm_uuid` - VM UUID.
* `container_id` - Container ID.
* `image_id` - Image ID.
* `image_name` - Image name.
* `host_attr` - Host attribute.
* `service` - Service.
* `microservice` - Microservice.
* `sys_arch` - System CPU architecture.
* `os_bit` - OS bit version.
* `os_type` - OS type.
* `os_name` - OS name.
* `os_version` - OS version.

<a name="hss_process_info_list"></a>
The `process_info_list` block supports:
* `process_name` - Process name.
* `process_path` - Process file path.
* `process_pid` - Process ID.
* `process_uid` - Process user ID.
* `process_username` - Process username.
* `process_cmdline` - Command line used to start the process.
* `process_filename` - Process file name.
* `process_start_time` - Process start time.
* `parent_process_name` - Parent process name.
* `parent_process_path` - Parent process file path.
* `parent_process_pid` - Parent process ID.

<a name="hss_user_info_list"></a>
The `user_info_list` block supports:
* `user_id` - User UID.
* `user_gid` - User GID.
* `user_name` - Username.
* `user_group_name` - User group name.
* `user_home_dir` - User home directory.
* `login_ip` - User login IP address.
* `service_type` - Type of service used for login.
* `service_port` - Login service port.
* `login_mode` - Login mode.
* `login_last_time` - Last login time.
* `login_fail_count` - Number of failed login attempts.

<a name="hss_file_info_list"></a>
The `file_info_list` block supports:
* `file_path` - Path to the file.
* `file_alias` - File alias.
* `file_size` - Size of the file in bytes.
* `file_mtime` - Time when a file was last modified.
* `file_atime` - Time when a file was last accessed.
* `file_ctime` - Time when the status of a file was last changed.
* `file_hash` - Hash of the file.
* `file_type` - Type of the file.
* `file_content` - File content.
* `file_attr` - File attribute.
* `file_operation` - File operation type.
