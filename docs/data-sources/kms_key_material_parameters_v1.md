---
subcategory: "Key Management Service (KMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_kms_key_material_parameters_v1"
sidebar_current: "docs-opentelekomcloud-datasource-kms-key-material-parameters-v1"
description: |-
  Get the parameters required to import key material into a CMK from OpenTelekomCloud
---

Up-to-date reference of API arguments for Obtaining CMK Import parameters you can get at
[documentation portal](https://docs.otc.t-systems.com/key-management-service/api-ref/apis/cmk_management/obtaining_cmk_import_parameters.html)

# opentelekomcloud_kms_key_material_parameters_v1

Use this data source to get the data required to import key material into a CMK in OpenTelekomCloud KMS.

~> **Warning** This data source returns parameters for a CMK in `Pending_import` state.
  Once the key is successfully imported and the state changes to `Enabled`, the data source will no longer fetch
  new parameters and its computed attributes will be nulled. If other resources utilize fields from this data source, consider
  adding `lifecycle { ignore_changes = [...] }` to handle state transitions properly.

## Example Usage

```hcl
data "opentelekomcloud_kms_key_material_parameters_v1" "params" {
  key_id             = "0d0466b0-e727-4d9c-b35d-f84bb474a37f"
  wrapping_algorithm = "RSAES_PKCS1_V1_5"
}
```

## Argument Reference

* `key_id` - (Required, String) The ID of the CMK to import key material into. Must be 36 bytes and match
  regexp `^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$`.

* `wrapping_algorithm` - (Required, String) The algorithm to be used for wrapping the imported key material.
  Valid values are:
    * `RSAES_PKCS1_V1_5`
    * `RSAES_OAEP_SHA_1`
    * `RSAES_OAEP_SHA_256`

* `sequence` - (Optional, String) 36-byte serial number of the request message.

## Attributes Reference

`id` is set to the date of the retrieved parameters. In addition, the following attributes are exported:

* `import_token` - The import token to use in subsequent ImportKey requests.

* `expiration_time` - The time at which the import token and public key expire.

* `public_key` - The public key to use to encrypt the key material before import (Base64 encoded).
