---
subcategory: "Config"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_rms_policy_assignment_evaluate_v1"
sidebar_current: "docs-opentelekomcloud-rms-policy-assignment-evaluate-v1"
description: |-
  Manages an RMS Policy Assignment Evaluate resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for RMS Policy Assignment Evaluate you can get at
[documentation portal](https://docs.otc.t-systems.com/config/api-ref/apis/compliance/running_a_resource_compliance_evaluation.html#rms-04-0510)

# opentelekomcloud_rms_policy_assignment_evaluate_v1

Manages a RMS policy assignment evaluate resource within OpenTelekomCloud resources.

## Example Usage

```hcl
variable "policy_assignment_id" {}

resource "opentelekomcloud_rms_policy_assignment_evaluate_v1" "test" {
  policy_assignment_id = var.policy_assignment_id
}
```

## Argument Reference

The following arguments are supported:

* `policy_assignment_id` - (Required, String, ForceNew) Specifies the ID of the policy assignment to evaluate.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the policy assignment evaluate.
