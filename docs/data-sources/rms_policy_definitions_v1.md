---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_policy_definitions_v1"
sidebar_current: "docs-opentelekomcloud-datasource-rms-policy-definitions-v1"
description: |-
  Manages an RMS Policy Definitions data source within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RMS Policy Definitions you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/compliance/querying_all_built-in_policies.html#rms-04-0501)


# opentelekomcloud_rms_policy_definitions_v1

Use this data source to query policy definition list.

## Example Usage

```hcl
variable "trigger_type" {}

data "opentelekomcloud_rms_policy_definitions_v1" "test" {
  trigger_type = var.trigger_type
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, String) Specifies the name of the policy definitions used to query definition list.

* `policy_type` - (Optional, String) Specifies the policy type used to query definition list.
  The valid value is **builtin**.

* `policy_rule_type` - (Optional, String) Specifies the policy rule type used to query definition list.

* `trigger_type` - (Optional, String) Specifies the trigger type used to query definition list.
  The valid values are **resource** and **period**.

* `keywords` - (Optional, List) Specifies the keyword list used to query definition list.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The data source ID.

* `definitions` - The policy definition list.
  The [object](#policy_definitions) structure is documented below.

<a name="policy_definitions"></a>
The `definitions` block supports:

* `id` - The ID of the policy definition.

* `name` - The name of the policy definition.

* `policy_type` - The policy type of the policy definition.

* `description` - The description of the policy definition.

* `policy_rule_type` - The policy rule type of the policy definition.

* `policy_rule` - The policy rule of the policy definition.

* `trigger_type` - The trigger type of the policy definition.

* `keywords` - The keyword list of the policy definition.

* `parameters` - The parameter reference map of the policy definition.
