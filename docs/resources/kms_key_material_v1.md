---
subcategory: "Key Management Service (KMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_kms_key_material_v1"
sidebar_current: "docs-opentelekomcloud-resource-kms-key-material-v1"
description: |-
  Manages a KMS Key Material resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for KMS Key Material you can get at
[documentation portal](https://docs.otc.t-systems.com/key-management-service/api-ref/apis/cmk_management/importing_cmk_material.html)

# opentelekomcloud_kms_key_material_v1

Manages a KMS key material resource within OpenTelekomCloud.

-> NOTE: Please confirm that the state of the imported key is pending import.

## Example Usage

### Basic usage

variable "key_id" {}
variable "import_token" {}
variable "encrypted_key_material" {}

```hcl
resource "opentelekomcloud_kms_key_material_v1" "test" {
  key_id                 = var.key_id
  import_token           = var.import_token
  encrypted_key_material = var.encrypted_key_material
}
```

### Complete key material import workflow

```hcl
locals {
  encrypt_script = <<-EOF
   #!/bin/bash
   INPUT=$(cat)
   PUBLIC_KEY=$(echo "$INPUT" | jq -r '.public_key')
   KEY_MATERIAL=$(echo "$INPUT" | jq -r '.input')
   TRIMMED_INPUT=$(echo -n "$KEY_MATERIAL" | head -c 32)
   echo "$PUBLIC_KEY" > public_key.pem
   ENCRYPTED=$(echo -n "$TRIMMED_INPUT" | openssl rsautl -encrypt -pubin -inkey public_key.pem -pkcs | base64 -w 0)
   rm public_key.pem
   printf '{"output":"%s"}' "$ENCRYPTED"
 EOF

  public_key = <<-EOF
   -----BEGIN PUBLIC KEY-----
   ${data.opentelekomcloud_kms_key_material_parameters_v1.params.public_key}
   -----END PUBLIC KEY-----
 EOF
}

resource "random_password" "key_material" {
  length           = 32
  special          = false
  override_special = "!#$%"
  lifecycle {
    ignore_changes = all
  }
}

data "external" "encrypt_key_material" {
  program = ["bash", "-c", local.encrypt_script]
  query = {
    input      = random_password.key_material.result
    public_key = local.public_key
  }
}

data "opentelekomcloud_kms_key_material_parameters_v1" "params" {
  key_id             = opentelekomcloud_kms_key_v1.key_1.id
  wrapping_algorithm = "RSAES_PKCS1_V1_5"
}

resource "opentelekomcloud_kms_key_material_v1" "test" {
  depends_on             = [data.external.encrypt_key_material]
  key_id                 = opentelekomcloud_kms_key_v1.key_1.id
  import_token           = data.opentelekomcloud_kms_key_material_parameters_v1.params.import_token
  encrypted_key_material = data.external.encrypt_key_material.result.output
  lifecycle {
    ignore_changes = all
  }
}

resource "opentelekomcloud_kms_key_v1" "key_1" {
  key_alias = "key_test"
  origin    = "external"
}
```

## Argument Reference

The following arguments are supported:

* `key_id` - (Required, String, ForceNew) Specifies the ID of the KMS key.
  Changing this creates a new resource.

* `import_token` - (Required, String, ForceNew) Specifies the key import token in Base64 format.
  The value contains `200` to `6144` characters, including letters, digits, slashes(/) and equals(=). This value is
  obtained through the interface [Obtaining Key Import Parameters](https://docs.otc.t-systems.com/key-management-service/api-ref/apis/cmk_management/obtaining_cmk_import_parameters.html)
  or by using `data_source/kms_key_material_parameters_v1`.

* `encrypted_key_material` - (Required, String, ForceNew) Specifies the encrypted symmetric key material in Base64 format.
  The value contains `344` to `360` characters, including letters, digits, slashes(/) and equals(=).
  If an asymmetric key is imported, this parameter is a temporary intermediate key used to encrypt the private key.
  This value is obtained refer to
  [documentation](https://docs.otc.t-systems.com/key-management-service/umn/user_guide/key_management/creating_cmks_using_imported_key_material/importing_a_key_material.html).

* `expiration_time` - (Optional, String, ForceNew) Specifies the expiration time of the key material.
  This field is only valid for symmetric keys. The time is in the format of timestamp, that is, the
  offset seconds from 1970-01-01 00:00:00 UTC to the specified time.
  The time must be greater than the current time.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID which equals the `key_id`.

* `key_state` - The status of the kms key. The valid values are as follows:
  **1**: To be activated
  **2**: Enabled.
  **3**: Disabled.
  **4**: Pending deletion.
  **5**: Pending import.

* `region` - The region in which KMS key is created.

## Import

The KMS key material can be imported using `id`, e.g.

```bash
$ terraform import opentelekomcloud_kms_key_material_v1.test 7056d636-ac60-4663-8a6c-82d3c32c1c64
```

Note that the imported state may not be identical to your resource definition,
due to `import_token`, `encrypted_key_material` and `encrypted_privatekey` are missing from the API response.
It is generally recommended running `terraform plan` after importing a KMS key material.
You can then decide if changes should be applied to the KMS key material, or the resource
definition should be updated to align with the KMS key material. Also you can ignore changes as below.

```hcl
resource "opentelekomcloud_kms_key_material_v1" "test" {
  lifecycle {
    ignore_changes = [import_token, encrypted_key_material, encrypted_privatekey]
  }
}
```
