---
subcategory: "Cloud Search Service (CSS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_css_certificate_v1"
sidebar_current: "docs-opentelekomcloud-datasource-css-certificate-v1"
description: |-
  Is used to obtain the HTTPS certificate of the server from OpenTelekomCloud
---

Up-to-date reference of API arguments for CSS flavor you can get at
[documentation portal](https://docs.otc.t-systems.com/cloud-search-service/api-ref/cluster_management_apis/downloading_the_certificate.html#css-03-0050)

# opentelekomcloud_css_certificate_v1

Use this data source to search matching CSS cluster flavor from OpenTelekomCloud.

## Example Usage

```hcl
data "opentelekomcloud_css_certificate_v1" "cert" {}
```

## Attributes Reference

The following attributes of a single found flavor are exported:

* `id` - Certificate ID.

* `region` - Indicates the region of the certificate.

* `project_id` - Indicates the project id of the certificate.

* `certificate` - String representation of server certificate.
