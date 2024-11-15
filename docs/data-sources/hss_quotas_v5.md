---
subcategory: "Host Security Service (HSS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_hss_quotas_v5"
sidebar_current: "docs-opentelekomcloud-datasource-hss-quotas-v5"
description: |-
  Use this data source to get the list of HSS quotas within OpenTelekomCloud.
---

Up-to-date reference of API arguments for EIP status you can get at
[documentation portal](https://docs.otc.t-systems.com/host-security-service/api-ref/api_description/quota_management/querying_quota_details.html#listquotasdetail)

# opentelekomcloud_hss_quotas_v5

Use this data source to get the list of HSS quotas within OpenTelekomCloud.

## Example Usage

```hcl
variable resource_id {}

data "opentelekomcloud_hss_quotas_v5" "qt" {
  resource_id = var.resource_id
}
```

## Argument Reference

The following arguments are supported:
* `category` - (Optional, String) Specifies the category of the quotas to be queried.
  The valid values are as follows:
  + `host_resource`: Host protection quota.
  + `container_resource`: Container protection quota.

* `version` - (Optional, String) Specifies the version of the quotas to be queried.
  The valid values are as follows:
  + `hss.version.enterprise`: Enterprise version.
  + `hss.version.premium`: Ultimate version.

* `status` - (Optional, String) Specifies the status of the quotas to be queried.
  The value can be `normal`, `expired`, or `freeze`.

* `used_status` - (Optional, String) Specifies the usage status of the quotas to be queried.
  The value can be `idle` or `used`.

* `host_name` - (Optional, String) Specifies the host name for the quota binding to be queried.

* `resource_id` - (Optional, String) Specifies the resource ID of the HSS quota.

* `charging_mode` - (Optional, String) Specifies the charging mode of the quotas to be queried.
  The valid values are as follows:
  + `on_demand`: The pay-per-use billing mode.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID in UUID format.

* `quotas` - All quotas that match the filter parameters.
  The [quotas](#hss_quotas) structure is documented below.

* `region` - The region in which to query the HSS quotas.

<a name="hss_quotas"></a>
The `quotas` block supports:

* `id` - The ID of quota.

* `version` - The version of quota.

* `status` - The status of quota.

* `used_status` - The usage status of quota.

* `host_id` - The host ID for quota binding.

* `host_name` - The host name for quota binding.

* `charging_mode` - The charging mode of quota.

* `expire_time` - The expiration time of quota, in RFC3339 format. This field is valid when the quota is a trial quota.

* `shared_quota` - Is it a shared quota. The value can be `shared` or `unshared`.

* `tags` - The key/value pairs to associate with the HSS quota.
