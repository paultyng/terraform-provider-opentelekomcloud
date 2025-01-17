---
subcategory: "APIGW"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_apigw_certificate_v2"
sidebar_current: "docs-opentelekomcloud-resource-apigw-certificate-v2"
description: |-
  Manages a APIGW Certificate resource within OpenTelekomCloud.
---

# opentelekomcloud_apigw_certificate_v2

Manages an APIGW SSL certificate resource within OpenTelekomCloud.

## Example Usage

### Manages a global SSL certificate

```hcl
variable "certificate_name" {}
variable "certificate_content" {
  type    = string
  default = "'-----BEGIN CERTIFICATE-----THIS IS YOUR CERT CONTENT-----END CERTIFICATE-----'"
}
variable "certificate_private_key" {
  type    = string
  default = "'-----BEGIN PRIVATE KEY-----THIS IS YOUR PRIVATE KEY-----END PRIVATE KEY-----'"
}

resource "opentelekomcloud_apigw_certificate_v2" "test" {
  name        = var.certificate_name
  content     = var.certificate_content
  private_key = var.certificate_private_key
}
```

### Manages a local SSL certificate in a specified dedicated APIGW instance

```hcl
variable "certificate_name" {}
variable "certificate_content" {
  type    = string
  default = "'-----BEGIN CERTIFICATE-----THIS IS YOUR CERT CONTENT-----END CERTIFICATE-----'"
}
variable "certificate_private_key" {
  type    = string
  default = "'-----BEGIN PRIVATE KEY-----THIS IS YOUR PRIVATE KEY-----END PRIVATE KEY-----'"
}
variable "dedicated_instance_id" {}

resource "opentelekomcloud_apigw_certificate_v2" "test" {
  name        = var.certificate_name
  content     = var.certificate_content
  private_key = var.certificate_private_key
  type        = "instance"
  instance_id = var.dedicated_instance_id
}
```

### Manages a local SSL certificate (with the ROOT CA certificate)

```hcl
variable "certificate_name" {}
variable "certificate_content" {
  type    = string
  default = "'-----BEGIN CERTIFICATE-----THIS IS YOUR CERT CONTENT-----END CERTIFICATE-----'"
}
variable "certificate_private_key" {
  type    = string
  default = "'-----BEGIN PRIVATE KEY-----THIS IS YOUR PRIVATE KEY-----END PRIVATE KEY-----'"
}
variable "root_ca_certificate_content" {
  type    = string
  default = "'-----BEGIN CERTIFICATE-----THIS IS YOUR CERT CONTENT-----END CERTIFICATE-----'"
}
variable "dedicated_instance_id" {}

resource "opentelekomcloud_apigw_certificate_v2" "test" {
  name            = var.certificate_name
  content         = var.certificate_content
  private_key     = var.certificate_private_key
  trusted_root_ca = var.root_ca_certificate_content
  type            = "instance"
  instance_id     = var.dedicated_instance_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) Specifies the certificate name.
  The valid length is limited from `4` to `50`, only Chinese and English letters, digits and underscores (_) are
  allowed. The name must start with an English letter.

* `content` - (Required, String) Specifies the certificate content.

* `private_key` - (Required, String) Specifies the private key of the certificate.

* `type` - (Optional, String, ForceNew) Specifies the certificate type. The valid values are as follows:
  + **instance**
  + **global**

  Defaults to **global**. Changing this will create a new resource.

* `instance_id` - (Optional, String, ForceNew) Specifies the dedicated instance ID to which the certificate belongs.
  Required if `type` is **instance**.
  Changing this will create a new resource.

* `trusted_root_ca` - (Optional, String) Specifies the trusted **ROOT CA** certificate.

-> Currently, the ROOT CA parameter only certificates of type `instance` are support.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The certificate ID.

* `region` - The region where the certificate is located.

* `effected_at` - The effective time of the certificate, in RFC3339 format (YYYY-MM-DDThh:mm:ssZ).

* `expires_at` - The expiration time of the certificate, in RFC3339 format (YYYY-MM-DDThh:mm:ssZ).

* `signature_algorithm` - What signature algorithm the certificate uses.

* `sans` - The SAN (Subject Alternative Names) of the certificate.

## Import

Certificates can be imported using their `id`, e.g.

```bash
$ terraform import opentelekomcloud_apigw_certificate_v2.test <id>
```

Note that the imported state may not be identical to your resource definition, due to some attributes missing from the
API response. The missing attributes include: `content`, `private_key` and `trusted_root_ca`.
It is generally recommended running `terraform plan` after importing a certificate.
You can then decide if changes should be applied to the certificate, or the resource definition should be updated to
align with the certificate. Also, you can ignore changes as below.

```hcl
resource "opentelekomcloud_apigw_certificate_v2" "test" {

  lifecycle {
    ignore_changes = [
      content, private_key, trusted_root_ca,
    ]
  }
}
```
