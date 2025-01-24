---
subcategory: "Cloud Container Engine (CCE)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_cce_cluster_v3"
sidebar_current: "docs-opentelekomcloud-resource-cce-cluster-v3"
description: |-
  Manages a CCE Cluster resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for CCE cluster you can get at
[documentation portal](https://docs.otc.t-systems.com/cloud-container-engine/api-ref/apis/cluster_management)

# opentelekomcloud_cce_cluster_v3

Provides a cluster resource management.

~>
  Before starting working with CCE, you need to authorize it via _console_ or [creating agency](#creating-agency).
  Otherwise, you will face the following error during cluster creation:
  `CCE is not authorized, see `cce_cluster_v3` documentation for details`.

~>
  You need to authorize CCE for the default (`eu-de`) project for CCE to be able to pull SWR images.

## Example Usage

### Simple cluster

```hcl
variable "flavor_id" {}
variable "vpc_id" {}
variable "subnet_id" {}

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name        = "cluster"
  description = "Create cluster"

  cluster_type           = "VirtualMachine"
  flavor_id              = var.flavor_id
  vpc_id                 = var.vpc_id
  subnet_id              = var.subnet_id
  container_network_type = "overlay_l2"
  authentication_mode    = "rbac"
  kube_proxy_mode        = "ipvs"
}
```

### Turbo cluster

```hcl
data "opentelekomcloud_vpc_subnet_v1" "shared_subnet" {
  name = "shared_test"
}

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                    = "turbo"
  cluster_type            = "VirtualMachine"
  flavor_id               = "cce.s1.small"
  vpc_id                  = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.vpc_id
  subnet_id               = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  container_network_type  = "eni"
  kubernetes_svc_ip_range = "10.247.0.0/16"
  ignore_addons           = true
  eni_subnet_id           = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.subnet_id
  eni_subnet_cidr         = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.cidr
}
```

### Installing ICAgent on Cluster creation

~>
When creating a cluster in the OTC UI, ICAgent is deployed automatically. This does not apply if a cluster is created via Terraform/API.

To make AOM work in conjunction with CCE, the ICAgent needs to be deployed on the cluster. You can do this automatically by adding the appropriate annotation to the cluster resource.

```hcl
variable "flavor_id" {}
variable "vpc_id" {}
variable "subnet_id" {}
resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name                   = "cluster"
  description            = "Create cluster"
  cluster_type           = "VirtualMachine"
  flavor_id              = var.flavor_id
  vpc_id                 = var.vpc_id
  subnet_id              = var.subnet_id
  container_network_type = "overlay_l2"
  authentication_mode    = "rbac"
  kube_proxy_mode        = "ipvs"
  annotations            = { "cluster.install.addons.external/install" = "[{\"addonTemplateName\":\"icagent\"}]" }
}
```

### Creating agency

You can create agency for CCE authorization using `opentelekomcloud_identity_agency_v3` resource.
For agency creation your user need to have corresponding permissions, which are not required for authorizing CCE via console

```hcl
resource "opentelekomcloud_identity_agency_v3" "enable_cce_auto_creation" {
  name                  = "cce_admin_trust"
  description           = "Created by Terraform to auto create cce"
  delegated_domain_name = "op_svc_cce"
  dynamic "project_role" {
    for_each = var.projects
    content {
      project = project_role.value
      roles = [
        "Tenant Administrator"
      ]
    }
  }
}

resource "opentelekomcloud_cce_cluster_v3" "cluster_1" {
  name        = "cluster"
  description = "Create cluster"

  cluster_type           = "VirtualMachine"
  flavor_id              = var.flavor_id
  vpc_id                 = var.vpc_id
  subnet_id              = var.subnet_id
  container_network_type = "overlay_l2"
  authentication_mode    = "rbac"

  depends_on = [opentelekomcloud_identity_agency_v3.enable_cce_auto_creation]
}
```

### CCE HA cluster

```hcl
variable "vpc_id" {}
variable "subnet_id" {}

resource "opentelekomcloud_cce_cluster_v3" "cluster" {
  name                   = "cluster"
  flavor_id              = "cce.s2.small"
  vpc_id                 = var.vpc_id
  subnet_id              = var.subnet_id
  container_network_type = "overlay_l2"

  masters {
    availability_zone = "eu-de-01"
  }
  masters {
    availability_zone = "eu-de-02"
  }
  masters {
    availability_zone = "eu-de-03"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Cluster name. Changing this parameter will create a new cluster resource.

* `labels` - (Optional) Cluster tag, key/value pair format. Changing this parameter will create a new cluster resource.

* `annotations` - (Optional) Cluster annotation, key/value pair format. Changing this parameter will create a new cluster resource.

* `flavor_id` - (Required) Cluster specifications. Changing this parameter will create a new cluster resource.
  * `cce.s1.small` - small-scale single cluster (up to 50 nodes).
  * `cce.s1.medium` - medium-scale single cluster (up to 200 nodes).
  * `cce.s2.small` - small-scale HA cluster (up to 50 nodes).
  * `cce.s2.medium` - medium-scale HA cluster (up to 200 nodes).
  * `cce.s2.large` - large-scale HA cluster (up to 1000 nodes).
  * `cce.s2.xlarge` - ultra-large-scale, high availability cluster (<= 2,000 nodes).

* `cluster_version` - (Optional) For the cluster version, possible values are `v1.27`, `v1.25`, `v1.23`, `v1.21`.
  If this parameter is not set, the cluster of the latest version is created by default.
  Changing this parameter will create a new cluster resource. [OTC-API](https://docs.otc.t-systems.com/en-us/api2/cce/cce_02_0236.html)

* `cluster_type` - (Required) Cluster Type, possible values are `VirtualMachine` and `BareMetal`. Changing this parameter will create a new cluster resource.

* `description` - (Optional) Cluster description.

* `billing_mode` - (Optional) Charging mode of the cluster, which is 0 (on demand). Changing this parameter will create a new cluster resource.

* `extend_param` - (Optional) Extended parameter. Changing this parameter will create a new cluster resource.
  [List of cluster extended params.](https://docs.otc.t-systems.com/cloud-container-engine/api-ref/apis/cluster_management/creating_a_cluster.html#cce-02-0236-table17575013586)

* `enable_volume_encryption` - (Optional) System and data disks encryption of master nodes. Changing this parameter will create a new cluster resource.

* `vpc_id` - (Required) The ID of the VPC used to create the node. Changing this parameter will create a new cluster resource.

* `subnet_id` - (Required) The Network ID of the subnet used to create the node. Changing this parameter will create a new cluster resource.

* `security_group_id` - (Optional) Default worker node security group ID of the cluster. If specified, the cluster will be bound to the target security group.
  Otherwise, the system will automatically create a default worker node security group for you.
  The default worker node security group needs to allow access from certain ports to ensure normal communications.
  Changing this parameter will create a new cluster resource.

* `highway_subnet_id` - (Optional) The ID of the high speed network used to create bare metal nodes. Changing this parameter will create a new cluster resource.

* `container_network_type` - (Required) Container network type.
  * `overlay_l2` - An overlay_l2 network built for containers by using Open vSwitch(OVS)
  * `underlay_ipvlan` - An underlay_ipvlan network built for bare metal servers by using ipvlan.
  * `vpc-router` - An vpc-router network built for containers by using ipvlan and custom VPC routes.

* `container_network_cidr` - (Optional) Container network segment. Changing this parameter will create a new cluster resource.

* `eni_subnet_id` -  - (Optional) Specifies the ENI subnet ID. Specified when creating a CCE Turbo cluster. Changing this parameter will create a new cluster resource.

* `eni_subnet_cidr` - (Optional) Specifies the ENI network segment. Specified when creating a CCE Turbo cluster. Changing this parameter will create a new cluster resource.

* `authentication_mode` - (Optional) Cluster authentication mode.
  * Clusters of Kubernetes v1.11 and earlier
    Possible values: `x509`, `rbac`, and `authenticating_proxy`
  * Clusters of Kubernetes v1.13 and later
    Possible values: `rbac` and `authenticating_proxy`

  Default value: `x509`
  Changing this parameter will create a new cluster resource.

* `authenticating_proxy_ca` - (Optional) CA root certificate provided in the `authenticating_proxy` mode.
  Deprecated, use `authenticating_proxy` instead.

* `authenticating_proxy` - (Optional) Authenticating proxy configuration. Required if `authentication_mode` is set to `authenticating_proxy`.
  * `ca` - X509 CA certificate configured in `authenticating_proxy` mode. The maximum size of the certificate is 1 MB.
  * `cert` - Client certificate issued by the X509 CA certificate configured in `authenticating_proxy` mode.
  This certificate is used for authentication from kube-apiserver to the extended API server.
  * `private_key` - Private key of the client certificate issued by the X509 CA certificate configured in `authenticating_proxy` mode.
  This key is used for authentication from kube-apiserver to the extended API server.

~>
  The private key used by the Kubernetes cluster does not support password encryption. Use an unencrypted private key.

* `multi_az` - (Optional) Enable multiple AZs for the cluster, only when using HA flavors. Changing this parameter will create a new cluster resource.
  This parameter and `masters` are alternative.

* `masters` - (Optional, List, ForceNew) Specifies the advanced configuration of master nodes.
  The [object](#cce_cluster_masters) structure is documented below.
  This parameter and `multi_az` are alternative. Changing this parameter will create a new cluster resource.

* `eip` - (Optional) EIP address of the cluster.

* `kubernetes_svc_ip_range` - (Optional) Service CIDR block, or the IP address range which the kubernetes
  clusterIp must fall within. This parameter is available only for clusters of v1.11.7 and later.

* `no_addons` - (Optional) Remove addons installed by the default after the cluster creation.

* `ignore_addons` - (Optional) Skip all cluster addons operations.

* `ignore_certificate_users_data` - (Optional) Skip sensitive user data.

* `ignore_certificate_clusters_data` - (Optional) Skip sensitive cluster data.

* `kube_proxy_mode` - Service forwarding mode. Two modes are available:
  * `iptables`: Traditional kube-proxy uses iptables rules to implement service load balancing.
    In this mode, too many iptables rules will be generated when many services are deployed.
    In addition, non-incremental updates will cause a latency and even obvious performance issues
    in the case of heavy service traffic.
  * `ipvs`: Optimized kube-proxy mode with higher throughput and faster speed.
    This mode supports incremental updates and can keep connections uninterrupted during service updates.
    It is suitable for large-sized clusters.

* `delete_evs` - (Optional) Specified whether to delete associated EVS disks when deleting the CCE cluster.
  valid values are **true**, **try** and **false**. Default is **false**.

* `delete_obs` - (Optional) Specified whether to delete associated OBS buckets when deleting the CCE cluster.
  valid values are **true**, **try** and **false**. Default is **false**.

* `delete_sfs` - (Optional) Specified whether to delete associated SFS file systems when deleting the CCE
  cluster. valid values are **true**, **try** and **false**. Default is **false**.

* `delete_efs` - (Optional) Specified whether to unbind associated SFS Turbo file systems when deleting the CCE
  cluster. valid values are **true**, **try** and **false**. Default is **false**.

* `delete_eni` - (Optional) Specified whether to delete ENI ports when deleting the CCE
  cluster. valid values are **true**, **try** and **false**. Default is **false**.

* `delete_net` - (Optional) Specified whether to delete cluster Service/ingress-related resources, such as ELB when deleting the CCE
  cluster. valid values are **true**, **try** and **false**. Default is **false**.

* `delete_all_storage` - (Optional) Specified whether to delete all associated storage resources when deleting the CCE
  cluster. valid values are **true**, **try** and **false**. Default is **false**.

* `delete_all_network` - (Optional) Specified whether to delete all associated network resources when deleting the CCE
  cluster. valid values are **true**, **try** and **false**. Default is **false**.

<a name="cce_cluster_masters"></a>
The `masters` block supports:

* `availability_zone` - (Optional, String, ForceNew) Specifies the availability zone of the master node.
  Changing this parameter will create a new cluster resource.

-> Note: Cluster custom deletion info and properties can be checked here:
  [Deleting a Specified Cluster.](https://docs.otc.t-systems.com/cloud-container-engine/api-ref/apis/cluster_management/deleting_a_specified_cluster.html)

## Attributes Reference

All above argument parameters can be exported as attribute parameters along with attribute reference.

* `id` - ID of the cluster resource.

* `status` - Cluster status information.

* `internal` - The internal network address.

* `external` - The external network address.

* `external_otc` - The endpoint of the cluster to be accessed through API Gateway.

* `certificate_clusters/name` - The cluster name.

* `certificate_clusters/server` - The server IP address.

* `certificate_clusters/certificate_authority_data` - The certificate data.

* `certificate_users/name` - The user name.

* `certificate_users/client_certificate_data` - The client certificate data.

* `certificate_users/client_key_data` - The client key data.

* `installed_addons` - List of installed addon IDs. Empty if `ignore_addons` is `true`.

* `security_group_control` - ID of the autogenerated security group for the CCE master port.

* `security_group_node` - ID of the autogenerated security group for the CCE nodes.

## Timeouts

This resource provides the following timeouts configuration options:

- `create` - Default is 30 minutes.

- `delete` - Default is 30 minutes.

## Import

Cluster can be imported using the cluster id, e.g.

```shell
terraform import opentelekomcloud_cce_cluster_v3.cluster_1 4779ab1c-7c1a-44b1-a02e-93dfc361b32d
```
