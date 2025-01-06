---
subcategory: "Cloud Container Engine (CCE)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_cce_node_v3"
sidebar_current: "docs-opentelekomcloud-resource-cce-node-v3"
description: |-
  Manages a CCE Cluster Node resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for CCE cluster node you can get at
[documentation portal](https://docs.otc.t-systems.com/cloud-container-engine/api-ref/apis/cluster_management)

# opentelekomcloud_cce_node_v3

Add a node to a container cluster.

## Example Usage

```hcl
variable "cluster_id" {}
variable "ssh_key" {}
variable "availability_zone" {}

resource "opentelekomcloud_cce_node_v3" "node_1" {
  name              = "node1"
  cluster_id        = var.cluster_id
  availability_zone = var.availability_zone

  os          = "EulerOS 2.9"
  flavor_id   = "s2.large.2"
  key_pair    = var.ssh_key
  runtime     = "containerd"
  agency_name = "test-agency"

  bandwidth_size = 100

  root_volume {
    size       = 40
    volumetype = "SATA"
  }

  data_volumes {
    size       = 100
    volumetype = "SATA"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
    extend_params = {
      "useType" = "docker"
    }
  }
}
```

## Argument Reference
The following arguments are supported:

* `cluster_id` - (Required, ForceNew, String) ID of the cluster. Changing this parameter will create a new resource.

* `flavor_id` - (Required, ForceNew, String) Specifies the flavor id. Changing this parameter will create a new resource.

* `availability_zone` - (Required, ForceNew, String) specify the name of the available partition (AZ). Changing this parameter will create a new resource.

* `key_pair` - (Required, ForceNew, String) Key pair name when logging in to select the key pair mode. Changing this parameter will create a new resource.

* `os` - (Optional, ForceNew, String) Node OS. Changing this parameter will create a new resource.

  Supported OS depends on kubernetes version of the cluster.
  | OS           | Kubernetes version |
  | :----------- | :----------------- |
  | HCE OS 2.0   | `v1.29`, `v1.28`, `v1.27`, `v1.25` |
  | Ubuntu 22.04 | `v1.29`, `v1.28`, `v1.27`, `v1.25` |
  | EulerOS release 2.9 | `v1.29`, `v1.28`, `v1.27`, `v1.25`, `v1.23`, `v1.21`, `V1.19` |
  | EulerOS release 2.5 | `v1.25`, `v1.23`, `v1.21`, `V1.19` |

* `billing_mode` - (Optional, ForceNew, Int) Node's billing mode: The value is `0` (on demand). Changing this parameter will create a new resource.

* `name` - (Optional, String) Node Name.

* `subnet_id` - (Optional, ForceNew, String) The ID of the subnet to which the NIC belongs. Changing this parameter will create a new resource.

* `labels` - (Optional, ForceNew, Map) Node tag, key/value pair format. Changing this parameter will create a new resource.

* `tags` - (Optional, Map) The field is alternative to `labels`, key/value pair format.

* `k8s_tags` - (Optional, ForceNew, Map) Tags of a Kubernetes node, key/value pair format.

* `annotations` - (Optional, ForceNew, Map) Node annotation, key/value pair format. Changing this parameter will create a new resource

* `runtime` - (Optional, ForceNew, String) Container runtime. Changing this parameter will create a new resource.
              Use with high-caution, may trigger resource recreation. Options are:
              `docker` - Docker
              `containerd` - Containerd

* `agency_name` - (Optional) IAM agency name. Changing this parameter will create a new resource.

* `taints` - (Optional, ForceNew, List) Taints to created nodes to configure anti-affinity.
  * `key` - (Required, String) A key must contain 1 to 63 characters starting with a letter or digit. Only letters, digits, hyphens (-), underscores (_), and periods (.) are allowed. A DNS subdomain name can be used as the prefix of a key.
  * `value` - (Required, String) A value must start with a letter or digit and can contain a maximum of 63 characters, including letters, digits, hyphens (-), underscores (_), and periods (.).
  * `effect` - (Optional, String) Available options are `NoSchedule`, `PreferNoSchedule`, and `NoExecute`.

* `eip_ids` - (Optional, List) List of existing elastic IP IDs.

-> If the `eip_ids` parameter is configured, you do not need to configure the `eip_count` and `bandwidth` parameters:
`iptype`, `bandwidth_charge_mode`, `bandwidth_size` and `share_type`.

* `eip_count` - (Optional, Int) Number of elastic IPs to be dynamically created.

* `iptype` - (Optional, String) Elastic IP type.

* `bandwidth_size` - (Optional, Int) Bandwidth size.

-> If the `bandwidth_size` parameter is configured, you do not need to configure the
  `eip_count`, `bandwidth_charge_mode`, `sharetype` and `iptype` parameters.

* `bandwidth_charge_mode` - (Optional, String) Bandwidth billing type.

* `sharetype` - (Optional, String) Bandwidth sharing type.

* `extend_param_charging_mode` - (Optional, ForceNew, Int) Node charging mode, 0 is on-demand charging. Changing this parameter will create a new cluster resource.

* `dedicated_host_id` - (Optional, String, ForceNew) Specifies the ID of the DeH to which the node is scheduled.

* `ecs_performance_type` - (Optional, ForceNew, String) Classification of cloud server specifications. Changing this parameter will create a new cluster resource.

* `order_id` - (Optional, ForceNew, String) Order ID, mandatory when the node payment type is the automatic payment package period type.
  Changing this parameter will create a new cluster resource.

* `product_id` - (Optional, ForceNew, String) The Product ID. Changing this parameter will create a new cluster resource.

* `max_pods` - (Optional, ForceNew, Int) The maximum number of instances a node is allowed to create. Changing this parameter will create a new node resource.

* `public_key` - (Optional, ForceNew, String) The Public key. Changing this parameter will create a new cluster resource.

* `private_ip` - (Optional, ForceNew, String) Private IP of the CCE node. Changing this parameter will create a new resource.

* `preinstall` - (Optional, ForceNew, String) Script required before installation. The input value can be a Base64 encoded string or not.
  Changing this parameter will create a new resource.

* `postinstall` - (Optional, ForceNew, String) Script required after installation. The input value can be a Base64 encoded string or not.
  Changing this parameter will create a new resource.

* `docker_base_size` - (Optional, ForceNew, Int) Available disk space of a single Docker container on the node using the device mapper.
  Changing this parameter will create a new node.

* `docker_lvm_config_override` - (Optional, ForceNew, String) `ConfigMap` of the Docker data disk.
  Changing this parameter will create a new node.

  Example:

  `dockerThinpool=vgpaas/90%VG;kubernetesLV=vgpaas/10%VG;diskType=evs;lvType=linear`

  In this example:

  - `userLV`: size of the user space, for example, vgpaas/20%VG.
  - `userPath`: mount path of the user space, for example, /home/wqt-test.
  - `diskType`: disk type. Currently, only the evs, hdd, and ssd are supported.
  - `lvType`: type of a logic volume. Currently, the value can be linear or striped.
  - `dockerThinpool`: Docker space size, for example, vgpaas/60%VG.
  - `kubernetesLV`: kubelet space size, for example, vgpaas/20%VG.

* `root_volume` - (Required, ForceNew, List) It corresponds to the system disk related configuration. Changing this parameter will create a new resource.
  * `size` - (Required, ForceNew, Int) Disk size in GB.
  * `volumetype` - (Required, ForceNew, String) Disk type.
  * `extend_params` - (Optional, ForceNew, Map) Disk expansion parameters. A list of strings which describes additional disk parameters.
  * `extend_param` **DEPRECATED** - (Optional, ForceNew, String) Disk expansion parameters.
  Please use alternative parameter `extend_params`.
  * `kms_id` - (Optional, ForceNew, String) The Encryption KMS ID of the system volume. By default, it tries to get from env by `OS_KMS_ID`.

* `data_volumes` - (Required, ForceNew, List) Represents the data disk to be created. Changing this parameter will create a new resource.
  * `size` - (Required, ForceNew, Int) Disk size in GB.
  * `volumetype` - (Required, ForceNew, String) Disk type.
  * `extend_params` - (Optional, ForceNew, Map) Disk expansion parameters. A list of strings which describes additional disk parameters.
  * `extend_param` **DEPRECATED** - (Optional, ForceNew, String) Disk expansion parameters.
  Please use alternative parameter `extend_params`.
  * `kms_id` - (Optional, ForceNew, String) The Encryption KMS ID of the data volume. By default, it tries to get from env by `OS_KMS_ID`.

-> To enable encryption with the KMS. Firstly, you need to create the agency to grant KMS rights to EVS.
The agency has to be created for a new project first with a user who has security `admin` permissions.
It is created automatically with the first encrypted EVS disk via UI.

## Attributes Reference

All above argument parameters can be exported as attribute parameters along with attribute reference.

* `status` - Node status information.

* `server_id` - ID of the ECS where the node resides.

* `public_ip` - Public IP of the CCE node.

## Timeouts

This resource provides the following timeouts configuration options:

- `create` - Default is 10 minutes.

- `delete` - Default is 10 minutes.
