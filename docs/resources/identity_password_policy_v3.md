---
subcategory: "Identity and Access Management (IAM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_identity_password_policy_v3"
sidebar_current: "docs-opentelekomcloud-resource-identity-password-policy-v3"
description: |-
  Manages a IAM Password Policy resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for IAM provider you can get at
[documentation portal](https://docs.otc.t-systems.com/identity-access-management/api-ref/apis/security_settings/modifying_the_password_policy.html)

# opentelekomcloud_identity_password_policy_v3

Manages the IAM account password policy within OpenTelekomCloud.

`Please use it with care!`
-> You _must_ have security admin privileges in your OpenTelekomCloud
cloud to use this resource. Please refer to [User Management Model](https://docs.otc.t-systems.com/en-us/usermanual/iam/iam_01_0034.html).

  During action `terraform destroy` it sets values the same as defaults for this resource.
  Defaults is:
  +  `maximum_consecutive_identical_chars` = 0
  +  `minimum_password_length` = 8
  +  `minimum_password_age` = 0
  +  `number_of_recent_passwords_disallowed` = 1
  +  `password_not_username_or_invert` = true
  +  `password_validity_period` = 0
  +  `password_char_combination` = 2

## Example Usage

```hcl
resource "opentelekomcloud_identity_password_policy_v3" "policy_1" {
  maximum_consecutive_identical_chars   = 0
  minimum_password_length               = 6
  minimum_password_age                  = 0
  number_of_recent_passwords_disallowed = 0
  password_not_username_or_invert       = true
  password_validity_period              = 179
  password_char_combination             = 3
}
```

## Argument Reference

The following arguments are supported:

* `maximum_consecutive_identical_chars` - (Optional, Int) Maximum number of times that a character is allowed to
  consecutively present in a password. Value range: `0-32`.

* `minimum_password_age` - (Optional, Int) Minimum period (minutes) after which users are allowed to make a password change.
  Value range: `0-1440`.

* `minimum_password_length` - (Optional, Int) Minimum number of characters that a password must contain. Value range: `6-32`.
  Default: `8`.

* `number_of_recent_passwords_disallowed` - (Optional, Int) Number of previously used passwords that are not allowed. Value range: `0-10`.
  Default: `1`.

* `password_not_username_or_invert` - (Optional, Bool) Indicates whether the password can be the username or the username spelled backwards.
  Default: `true`.

* `password_validity_period` - (Optional, Int) Password validity period (days).
  Value range: 0-180. Value 0 indicates that this requirement does not apply.

* `password_char_combination` - (Optional, Int) Minimum number of character types that a password must contain.
  Value range: 2-4.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of account password policy, which is the same as the domain ID.

* `maximum_password_length` - The maximum number of characters that a password can contain.

## Import

Identity password policy can be imported using the account ID or domain ID, e.g.

```bash
$ terraform import opentelekomcloud_identity_password_policy_v3.example <ID>
```
