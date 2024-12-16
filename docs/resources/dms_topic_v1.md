---
subcategory: "Distributed Message Service (DMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_dms_topic_v1"
sidebar_current: "docs-opentelekomcloud-resource-dms-topic-v1"
description: |-
  Manages a DMS Topic resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for DMS topic you can get at
[documentation portal](https://docs.otc.t-systems.com/distributed-message-service/api-ref/out-of-date_apis/api_v1/apis_for_managing_instances)

# opentelekomcloud_dms_topic_v1

Manages a DMS topic in the OpenTelekomCloud DMS Service (Kafka Premium/Platinum).

~>
Deprecated, use `opentelekomcloud_dms_topic_v2` resource instead

## Example Usage: creating dms instance with topic

```hcl
resource "opentelekomcloud_networking_secgroup_v2" "secgroup_1" {
  name        = "secgroup_1"
  description = "secgroup_1"
}

data "opentelekomcloud_dms_az_v1" "az_1" {
  name = "eu-de-01"
}

data "opentelekomcloud_dms_product_v1" "product_1" {
  engine            = "kafka"
  version           = "2.3.0"
  instance_type     = "cluster"
  partition_num     = 300
  storage           = 600
  storage_spec_code = "dms.physical.storage.high"
}

resource "opentelekomcloud_dms_instance_v1" "instance_1" {
  name              = "kafka-test"
  engine            = "kafka"
  product_id        = data.opentelekomcloud_dms_product_v1.product_1.id
  engine_version    = data.opentelekomcloud_dms_product_v1.product_1.version
  specification     = data.opentelekomcloud_dms_product_v1.product_1.bandwidth
  partition_num     = data.opentelekomcloud_dms_product_v1.product_1.partition_num
  storage_spec_code = data.opentelekomcloud_dms_product_v1.product_1.storage_spec_code
  storage_space     = data.opentelekomcloud_dms_product_v1.product_1.storage
  available_zones   = [data.opentelekomcloud_dms_az_v1.az_1.id]
  security_group_id = opentelekomcloud_networking_secgroup_v2.secgroup_1.id
  vpc_id            = var.vpc_id
  subnet_id         = var.subnet_id
  access_user       = var.access_user
  password          = var.password
}

resource "opentelekomcloud_dms_topic_v1" "topic_1" {
  instance_id      = resource.opentelekomcloud_dms_instance_v1.instance_1.id
  name             = "topic-test"
  partition        = 10
  replication      = 2
  sync_replication = true
  retention_time   = 80
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) Indicates the ID of primary DMS instance.

* `name` - (Required) Indicates the name of a topic.

* `partition` - (Optional) Indicates the number of topic partitions,
  which is used to set the number of concurrently consumed messages.
  Value range: `1–20`. Default value: `3`.

* `replication` - (Optional) Indicates the number of replicas,
  which is configured to ensure data reliability.
  Value range: `1–3`. Default value: `3`.

* `sync_replication` - (Optional) Indicates whether to enable synchronous replication.
  After this function is enabled, the `acks` parameter on the producer client must be set to `–1`.
  Otherwise, this parameter does not take effect.

* `retention_time` - (Required) Indicates the retention period of a message. Its default value is `72`.
  Value range: `1–720`. Default value: `72`. Unit: `hour`.

* `sync_message_flush` - (Optional) Indicates whether to enable synchronous flushing.
  Default value: `false`. Synchronous flushing compromises performance.


## Attributes Reference

All above argument parameters can be exported as attribute parameters along with attribute reference.

* `size` - The partition size of the topic.

* `remain_partitions` - Number of remaining partitions.

* `max_partitions` - Total partitions number.
