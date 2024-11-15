---
subcategory: "Host Security Service (HSS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_hss_hosts_v5"
sidebar_current: "docs-opentelekomcloud-datasource-hss-hosts-v5"
description: |-
  Use this data source to get the list of HSS hosts within OpenTelekomCloud.
---

Up-to-date reference of API arguments for EIP status you can get at
[documentation portal](https://docs.otc.t-systems.com/host-security-service/api-ref/api_description/server_management/querying_ecss.html#listhoststatus)

# opentelekomcloud_hss_hosts_v5

Use this data source to get the list of HSS hosts within OpenTelekomCloud.

## Example Usage

```hcl
variable host_id {}

data "opentelekomcloud_hss_hosts_v5" "test" {
  host_id = var.host_id
}
```

## Argument Reference

The following arguments are supported:

* `host_id` - (Optional, String) Specifies the ID of the host to be queried.

* `name` - (Optional, String) Specifies the name of the host to be queried.
  This field will undergo a fuzzy matching query, the query result is for all hosts whose names contain this value.

* `status` - (Optional, String) Specifies the status of the hosts to be queried.
  The valid values are as follows:
  + `ACTIVE`
  + `SHUTOFF`
  + `BUILDING`
  + `ERROR`

* `os_type` - (Optional, String) Specifies the operating system type of the hosts to be queried.
  The valid values are as follows:
  + `Linux`
  + `Windows`

* `agent_status` - (Optional, String) Specifies the agent status of the hosts to be queried.
  The valid values are as follows:
  + `installing`
  + `not_installed`
  + `online`
  + `offline`
  + `install_failed`

* `protect_status` - (Optional, String) Specifies the protection status of the hosts to be queried.
  The valid values are as follows:
  + `closed`
  + `opened`

* `protect_version` - (Optional, String) Specifies the protection version enabled by the hosts to be queried.
  The valid values are as follows:
  + `hss.version.null`
  + `hss.version.enterprise`
  + `hss.version.premium`
  + `hss.version.container.enterprise`

* `protect_charging_mode` - (Optional, String) Specifies the charging mode for the hosts protection quota to be queried.
  The valid values are as follows:
  + `on_demand`

* `detect_result` - (Optional, String) Specifies the security detection result of the hosts to be queried.
  The valid values are as follows:
  + `undetected`
  + `clean`
  + `risk`
  + `scanning`

* `group_id` - (Optional, String) Specifies the host group ID of the hosts to be queried.

* `policy_group_id` - (Optional, String) Specifies the policy group ID of the hosts to be queried.

* `asset_value` - (Optional, String) Specifies the asset importance of the hosts to be queried.
  The valid values are as follows:
  + `important`
  + `common`
  + `test`


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID in UUID format.

* `hosts` - All hosts that match the filter parameters.
  The [hosts](#hss_hosts) structure is documented below.

* `region` - The region in which to query the HSS hosts.

<a name="hss_hosts"></a>
The `hosts` block supports:

* `id` - The ID of the host.

* `name` - The name of the host.

* `status` - The status of the host.

* `os_type` - The operating system type of the host.

* `agent_id` - The agent ID installed on the host.

* `agent_status` - The agent status of the host.

* `protect_status` - The protection status of the host.

* `protect_version` - The protection version enabled by the host.

* `protect_charging_mode` - The charging mode for the host protection quota.

* `resource_id` - The Cloud service resource instance ID.

* `detect_result` - The security detection result of the host.

* `group_id` - The host group ID to which the host belongs.

* `policy_group_id` - The policy group ID to which the host belongs.

* `asset_value` - The asset importance of the host.

* `open_time` - The time to enable host protection.

* `private_ip` - The private IP address of the host.

* `public_ip` - The elastic public IP address of the host.

* `asset_risk_num` - The number of asset risks in the host

* `vulnerability_risk_num` - The number of vulnerability risks in the host.

* `baseline_risk_num` - The number of baseline risks in the host.

* `intrusion_risk_num` - The number of intrusion risks in the host.
