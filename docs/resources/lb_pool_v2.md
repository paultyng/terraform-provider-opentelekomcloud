---
subcategory: "Elastic Load Balancer (ELB)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_lb_pool_v2"
sidebar_current: "docs-opentelekomcloud-resource-lb-pool-v2"
description: |-
  Manages a ELB Pool resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for ELB pool you can get at
[documentation portal](https://docs.otc.t-systems.com/elastic-load-balancing/api-ref/apis_v2.0/backend_server_group)

# opentelekomcloud_lb_pool_v2

Manages an Enhanced LB pool resource within OpenTelekomCloud.

## Example Usage

```hcl
resource "opentelekomcloud_lb_pool_v2" "pool_1" {
  protocol    = "HTTP"
  lb_method   = "ROUND_ROBIN"
  listener_id = "d9415786-5f1a-428b-b35f-2f1523e146d2"

  persistence {
    type        = "APP_COOKIE"
    cookie_name = "testCookie"
  }
}
```

## Argument Reference

The following arguments are supported:

* `tenant_id` - (Optional) Required for admins. The UUID of the tenant who owns
  the pool.  Only administrative users can specify a tenant UUID
  other than their own. Changing this creates a new pool.

* `name` - (Optional) Human-readable name for the pool.

* `description` - (Optional) Human-readable description for the pool.

* `protocol` - (Required) The protocol - can either be TCP, UDP or HTTP.
  Changing this creates a new pool.

-> When a pool is added to a specific listener, the relationships between the load balancer protocol
and the pool protocol are as follows. When the load balancer protocol is `UDP`, the pool protocol must be `UDP`.
When the load balancer protocol is `TCP`, the pool protocol must be `TCP`.
When the load balancer protocol is `HTTP` or `TERMINATED_HTTPS`, the pool protocol must be `HTTP`.

* `loadbalancer_id` - (Optional) The load balancer on which to provision this
  pool. Changing this creates a new pool.

-> One of `loadbalancer_id` or `listener_id` must be provided.

* `listener_id` - (Optional) The Listener on which the members of the pool
  will be associated with. Changing this creates a new pool.

-> One of `loadbalancer_id` or `listener_id` must be provided.

* `lb_method` - (Required) The load balancing algorithm to
  distribute traffic to the pool's members. Must be one of
  `ROUND_ROBIN`, `LEAST_CONNECTIONS`, or `SOURCE_IP`.

* `persistence` - (Optional) Omit this field to prevent session persistence. Indicates
  whether connections in the same session will be processed by the same Pool
  member or not. Changing this creates a new pool.

* `admin_state_up` - (Optional) The administrative state of the pool.
  A valid value is true (UP) or false (DOWN).

The `persistence` argument supports:

* `type` - (Optional; Required if `type != null`) The type of persistence mode. The current specification
  supports `SOURCE_IP`, `HTTP_COOKIE`, and `APP_COOKIE`.

* `cookie_name` - (Optional; Required if `type = APP_COOKIE`) The name of the cookie if persistence mode is set
  appropriately.

## Attributes Reference

The following attributes are exported:

* `id` - The unique ID for the pool.

* `tenant_id` - See Argument Reference above.

* `name` - See Argument Reference above.

* `description` - See Argument Reference above.

* `protocol` - See Argument Reference above.

* `lb_method` - See Argument Reference above.

* `persistence` - See Argument Reference above.

* `admin_state_up` - See Argument Reference above.
