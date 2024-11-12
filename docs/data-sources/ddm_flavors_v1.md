---
subcategory: "Distributed Database Middleware (DDM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_ddm_flavors_v1"
sidebar_current: "docs-opentelekomcloud-datasource-ddm-flavors-v1"
description: |-
  Get DDM flavors from OpenTelekomCloud.
---

Up-to-date reference of API arguments for DDM compute flavors you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-database-middleware/api-ref/apis_recommended/ddm_instances/querying_ddm_node_classes_available_in_an_az.html)

# opentelekomcloud_ddm_flavors_v1

Use this data source to get info of OpenTelekomCloud DDM compute flavors.

## Example Usage
```hcl
data "opentelekomcloud_ddm_engines_v1" "engine_list" {
}

data "opentelekomcloud_ddm_flavors_v1" "flavor_list" {
  engine_id = data.opentelekomcloud_ddm_engines_v1.engine_list.engines.0.id
}
```

## Argument Reference

The following arguments are supported:

* `engine_id` - (Required, String) Specifies the DDM engine ID.

## Attributes Reference

The following attributes are exported:

* `region` - Indicates the region of the DDM compute flavors.
* `engine_id` -  See Argument Reference above.
* `flavor_groups` - Indicates the DDM compute flvaor groups information. Structure is documented below.

The `flavor_groups` block contains:

  - `type` - Indicates the DDM compute flavor group type. The value can be x86 or ARM.
  - `flavors` - Indicates the available compute flavors in the flavor group. Structure is documented below

The `availability_zones` block contains:

  - `id` - Indicates the compute flavor ID.
  - `type_code` - Indicates the resource type code.
  - `code` - Indicates the VM flavor types recorded in DDM.
  - `iaas_code` - Indicates the VM flavor types recorded by the IaaS layer.
  - `cpu` - Indicates the number of CPUs.
  - `memory` - Indicates the memory size, in GB.
  - `max_connections` - Indicates the maximum number of connections.
  - `server_type` - Indicates the compute resource type.
  - `architecture` - Indicates the coompute resource architecture type. The value can be x86 or ARM.
  - `az_status` - Status of the AZ where node classes are available. The key is the AZ ID and the value is the AZ status. The value can be `normal`, `unsupported`, or `sellout`.
