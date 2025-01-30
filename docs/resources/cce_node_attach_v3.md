---
subcategory: "Cloud Container Engine (CCE)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_cce_node_attach_v3"
sidebar_current: "docs-opentelekomcloud-resource-cce-node-attach-v3"
description: |-
  Manages a CCE Cluster Node Attach resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for CCE cluster node you can get at
[documentation portal](https://docs.otc.t-systems.com/cloud-container-engine/api-ref/apis/cluster_management)

# opentelekomcloud_cce_node_attach_v3

Add a node from an existing ECS server to a CCE cluster.

-> **Note:** When using the `opentelekomcloud_cce_node_attach_v3` resource with server_id dependent resource, the following
attributes should be included in lifecycle.ignore_changes: `name`, `image_id`, `password`, `key_name`, `tags`, and `nics`.
These attributes are subject to modification during instance lifecycle and should be ignored to prevent undesired resource recreation.

## Basic Usage

```hcl
variable "cluster_id" {}
variable "server_name" {}
variable "server_id" {}
variable "keypair_name" {}
variable "flavor_id" {}
variable "image_id" {}

resource "opentelekomcloud_ecs_instance_v1" "instance_1" {
  name     = var.server_name
  image_id = var.image_id
  flavor   = var.flavor_id
  vpc_id   = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id

  nics {
    network_id = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  }

  availability_zone = "eu-de-01"
  data_disks {
    type = "SSD"
    size = 60
  }

  password                    = "Password@123"
  delete_disks_on_termination = true

  lifecycle {
    ignore_changes = [
      name,
      image_id,
      password,
      key_name,
      tags,
      nics
    ]
  }
}

resource "opentelekomcloud_cce_node_attach_v3" "test" {
  cluster_id = var.cluster_id
  server_id  = opentelekomcloud_ecs_instance_v1.instance_1.id
  key_pair   = var.keypair_name
  os         = "EulerOS 2.5"

  tags = {
    foo = "bar"
    key = "value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required, String, ForceNew) Specifies the ID of the cluster. Changing this parameter will create a new
  resource.

* `name` - (Optional, String) Specifies the Node Name.

* `server_id` - (Required, String, ForceNew) Specifies the ecs server ID. Changing this parameter will create a new
  resource.

* `os` - (Required, String) Specifies the operating System of the node. Changing this parameter will reset the node.
  The value can be **EulerOS 2.5**, **EulerOS 2.9**, **Ubuntu 22.04**, or **HCE OS 2.0**. For more details,
  please see [documentation](https://docs.otc.t-systems.com/cloud-container-engine/umn/nodes/node_oss.html).

* `key_pair` - (Optional, String) Specifies the key pair name when logging in to select the key pair mode.
  This parameter and `password` are alternative.

* `password` - (Optional, String) Specifies the root password when logging in to select the password mode.
  The password consists of 8 to 26 characters and must contain at least three of following: uppercase letters,
  lowercase letters, digits, special characters(!@$%^-_=+[{}]:,./?~#*).
  This parameter can be plain or salted and is alternative to `key_pair`.

* `private_key` - (Optional, String) Specifies the private key of the in used `key_pair`. This parameter is mandatory
  when replacing or unbinding a keypair if the CCE node is in **Active** state.

* `max_pods` - (Optional, Int) Specifies the maximum number of instances a node is allowed to create.
  Changing this parameter will reset the node.

* `system_disk_kms_key_id` - (Optional, String) Specifies the KMS key ID. This is used to encrypt the root volume.
  Changing this parameter will reset the node.

* `runtime` - (Optional, String) Specifies the runtime of the CCE node. Valid values are *docker* and
  *containerd*. Changing this parameter will reset the node.

* `docker_base_size` - (Optional, Int) Specifies the available disk space of a single docker container on the
  node in device mapper mode. Changing this parameter will reset the node.

* `lvm_config` - (Optional, String) Specifies the docker data disk configurations.
  This parameter is alternative to `storage`, and it's recommended to use `storage`.
  The following is an
  example:

```hcl
lvm_config = "dockerThinpool=vgpaas/90%VG;kubernetesLV=vgpaas/10%VG"
```

Changing this parameter will reset the node.

* `storage` - (Optional, List) Specifies the disk initialization management parameter.
  This parameter is alternative to `lvm_config` and supported for clusters of v1.15.11 and later.
  Changing this parameter will reset the node.

  + `selectors` - (Required, List) Specifies the disk selection.
    Matched disks are managed according to match labels and storage type. Structure is documented below.
    Changing this parameter will reset the node.
  + `groups` - (Required, List) Specifies the storage group consists of multiple storage devices.
    This is used to divide storage space. Structure is documented below.
    Changing this parameter will reset the node.

* `preinstall` - (Optional, String) Specifies the script required before installation. The input value can be
  a Base64 encoded string or not. Changing this parameter will reset the node.

* `postinstall` - (Optional, String) Specifies the script required after installation. The input value can be
  a Base64 encoded string or not. Changing this parameter will reset the node.

* `k8s_tags` - (Optional, Map) Specifies the tags of a Kubernetes node, key/value pair format.
  Changing this parameter will reset the node.

* `tags` - (Optional, Map) Specifies the tags of a VM node, key/value pair format.

* `taints` - (Optional, List) Specifies the taints configuration of the nodes to set anti-affinity.
  Changing this parameter will reset the node. Each taint contains the following parameters:

  + `key` - (Required, String) A key must contain 1 to 63 characters starting with a letter or digit.
    Only letters, digits, hyphens (-), underscores (_), and periods (.) are allowed. A DNS subdomain name can be used
    as the prefix of a key. Changing this parameter will reset the node.
  + `value` - (Required, String) A value must start with a letter or digit and can contain a maximum of 63
    characters, including letters, digits, hyphens (-), underscores (_), and periods (.). Changing this parameter will
    reset the node.
  + `effect` - (Required, String) Available options are NoSchedule, PreferNoSchedule, and NoExecute.
    Changing this parameter will reset the node.

The `selectors` block supports:

* `name` - (Required, String) Specifies the selector name, used as the index of `selector_names` in storage group.
  The name of each selector must be unique. Changing this parameter will reset the node.
* `type` - (Optional, String) Specifies the storage type. Currently, only **evs (EVS volumes)** is supported.
  The default value is **evs**. Changing this parameter will reset the node.
* `match_label_size` - (Optional, String) Specifies the matched disk size. If omitted,
  the disk size is not limited. Example: 100. Changing this parameter will reset the node.
* `match_label_volume_type` - (Optional, String) Specifies the EVS disk type. Currently,
  **SSD**, **GPSSD**, and **SAS** are supported. If omitted, the disk type is not limited.
  Changing this parameter will reset the node.
* `match_label_metadata_encrypted` - (Optional, String) Specifies the disk encryption identifier.
  Values can be: **0** indicates that the disk is not encrypted and **1** indicates that the disk is encrypted.
  If omitted, whether the disk is encrypted is not limited. Changing this parameter will reset the node.
* `match_label_metadata_cmkid` - (Optional, String) Specifies the customer master key ID of an encrypted
  disk. Changing this parameter will reset the node.
* `match_label_count` - (Optional, String) Specifies the number of disks to be selected. If omitted,
  all disks of this type are selected. Changing this parameter will reset the node.

The `groups` block supports:

* `name` - (Required, String) Specifies the name of a virtual storage group. Each group name must be unique.
  Changing this parameter will reset the node.
* `cce_managed` - (Optional, Bool) Specifies the whether the storage space is for **kubernetes** and
  **runtime** components. Only one group can be set to true. The default value is **false**.
  Changing this parameter will reset the node.
* `selector_names` - (Required, List) Specifies the list of names of selectors to match.
  This parameter corresponds to name in `selectors`. A group can match multiple selectors,
  but a selector can match only one group. Changing this parameter will reset the node.
* `virtual_spaces` - (Required, List) Specifies the detailed management of space configuration in a group.
  Changing this parameter will reset the node.

  + `name` - (Required, String) Specifies the virtual space name. Currently, only **kubernetes**, **runtime**,
    and **user** are supported. Changing this parameter will reset the node.
  + `size` - (Required, String) Specifies the size of a virtual space. Only an integer percentage is supported.
    Example: 90%. Note that the total percentage of all virtual spaces in a group cannot exceed 100%.
    Changing this parameter will reset the node.
  + `lvm_lv_type` - (Optional, String) Specifies the LVM write mode, values can be **linear** and **striped**.
    This parameter takes effect only in **kubernetes** and **user** configuration. Changing this parameter will create
    a new resource.
  + `lvm_path` - (Optional, String) Specifies the absolute path to which the disk is attached.
    This parameter takes effect only in **user** configuration. Changing this parameter will reset the node.
  + `runtime_lv_type` - (Optional, String) Specifies the LVM write mode, values can be **linear** and **striped**.
    This parameter takes effect only in **runtime** configuration. Changing this parameter will reset the node.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID in UUID format.

* `region` - The region in which nodes is created.

* `status` - Node status information.

* `private_ip` - Private IP of the CCE node.

* `public_ip` - Public IP of the CCE node.

* `flavor_id` - The flavor ID of the CCE node.

* `availability_zone` - The name of the available partition (AZ).

* `subnet_id` - The ID of the subnet to which the NIC belongs.

* `root_volume` - The configuration of the system disk.
  + `size` - The disk size in GB.
  + `volumetype` - The disk type.
  + `extend_params` - The disk expansion parameters.
  + `kms_id` - The ID of a KMS key. This is used to encrypt the volume.

* `data_volumes` - The configurations of the data disk.
  + `size` - The disk size in GB.
  + `volumetype` - The disk type.
  + `extend_params` - The disk expansion parameters.
  + `kms_id` - The ID of a KMS key. This is used to encrypt the volume.
## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 20 minutes.
* `update` - Default is 20 minutes.
* `delete` - Default is 20 minutes.
