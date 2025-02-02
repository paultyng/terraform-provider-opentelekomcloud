---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_policy_states_v1"
sidebar_current: "docs-opentelekomcloud-rms-resource-policy-states-v1"
description: |-
  Use this data source to get the list of RMS policy states.
---

Up-to-date reference of API arguments for RMS Resource Recorder you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/compliance/index.html)

# opentelekomcloud_rms_policy_states_v1

Use this data source to get the list of RMS policy states.

## Example Usage

```hcl
data "opentelekomcloud_rms_policy_states_v1" "test" {}
```

## Argument Reference

The following arguments are supported:

* `policy_assignment_id` - (Optional, String) Specifies the policy assignment ID.

* `compliance_state` - (Optional, String) Specifies the compliance state.
  The value can be: **Compliant** and **NonCompliant**.

* `resource_name` - (Optional, String) Specifies the resource name.

* `resource_id` - (Optional, String) Specifies the resource ID.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID.

* `states` - The policy states list.

  The [states](#states) structure is documented below.

<a name="states"></a>
The `states` block supports:

* `domain_id` - The domain ID.

* `region_id` - The ID of the region the resource belongs to.

* `resource_id` - The resource ID.

* `resource_name` - The resource name.

* `resource_provider` - The cloud service name.

* `resource_type` - The resource type.

* `trigger_type` - The trigger type. The value can be **resource** or **period**.

* `compliance_state` - The compliance status.

* `policy_assignment_id` - The policy assignment ID.

* `policy_assignment_name` - The policy assignment name.

* `policy_definition_id` - The ID of the policy definition.

* `evaluation_time` - The evaluation time of compliance status.
