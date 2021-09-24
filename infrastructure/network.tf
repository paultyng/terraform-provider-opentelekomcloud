resource "opentelekomcloud_vpc_v1" "this" {
  name   = "default_vpc"
  cidr   = "192.168.0.0/16"
  shared = true
}

resource "opentelekomcloud_vpc_subnet_v1" "this" {
  name        = "default_subnet"
  cidr        = "192.168.0.0/16"
  gateway_ip  = "192.168.0.1"
  vpc_id      = opentelekomcloud_vpc_v1.this.id
  dhcp_enable = true
  dns_list    = ["100.125.4.25", "1.1.1.1"]
}
