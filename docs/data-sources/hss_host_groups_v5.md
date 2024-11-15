---
subcategory: "Host Security Service (HSS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_hss_host_groups_v5"
sidebar_current: "docs-opentelekomcloud-datasource-hss-host-groups-v5"
description: |-
  Use this data source to get the list of HSS host groups within OpenTelekomCloud.
---

Up-to-date reference of API arguments for EIP status you can get at
[documentation portal](https://docs.otc.t-systems.com/host-security-service/api-ref/api_description/server_management/querying_server_groups.html#listhostgroups)

# opentelekomcloud_hss_host_group_v5

Use this data source to get the list of HSS host groups within OpenTelekomCloud.

## Example Usage

```hcl
variable group_id {}

data "opentelekomcloud_hss_host_group_v5" "test" {
  group_id = var.group_id
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional, String) Specifies the ID of the host group to be queried.

* `name` - (Optional, String) Specifies the name of the host group to be queried. This field will undergo a fuzzy
  matching query, the query result is for all host groups whose names contain this value.

* `host_num` - (Optional, String) Specifies the number of hosts in the host groups to be queried.

* `risk_host_num` - (Optional, String) Specifies the number of risky hosts in the host groups to be queried.

* `unprotect_host_num` - (Optional, String) Specifies the number of unprotected hosts in the host groups to be queried.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID in UUID format.

* `groups` - All host groups that match the filter parameters.

* `region` - The region in which to query the HSS host groups.

  The [groups](#hss_groups) structure is documented below.

<a name="hss_groups"></a>
The `groups` block supports:

* `id` - The ID of the host group.

* `name` - The name of the host group.

* `host_num` - The number of hosts in the host group.

* `risk_host_num` - The number of risky hosts in the host group.

* `unprotect_host_num` - The number of unprotected hosts in the host group.

* `host_ids` - The list of host IDs in the host group.
