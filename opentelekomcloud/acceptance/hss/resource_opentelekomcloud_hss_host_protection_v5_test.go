package hss

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/hss/v5/host"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/hss"
)

func getHostProtectionFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.HssV5Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating HSS v5 client: %s", err)
	}
	hostList, err := host.ListHost(client, host.ListHostOpts{HostID: state.Primary.ID})
	if err != nil {
		return nil, fmt.Errorf("error querying OpenTelekomCloud HSS hosts: %s", err)
	}
	if len(hostList) == 0 || hostList[0].ProtectStatus == string(hss.ProtectStatusClosed) {
		return nil, golangsdk.ErrDefault404{}
	}
	return hostList[0], nil
}

func TestAccHostProtection_basic(t *testing.T) {
	var (
		h     *host.HostGroupResp
		rName = "opentelekomcloud_hss_host_protection_v5.protection"
		name  = fmt.Sprintf("hss-acc-api%s", acctest.RandString(5))
	)

	rc := common.InitResourceCheck(
		rName,
		&h,
		getHostProtectionFunc,
	)

	// Because after closing the protection, the ECS instance will automatically switch to free basic protection,
	// so avoid CheckDestroy here.
	// lintignore:AT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostProtection_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "version", "hss.version.premium"),
					resource.TestCheckResourceAttr(rName, "charging_mode", "on_demand"),
					resource.TestCheckResourceAttrSet(rName, "host_name"),
					resource.TestCheckResourceAttrSet(rName, "private_ip"),
					resource.TestCheckResourceAttrSet(rName, "agent_id"),
					resource.TestCheckResourceAttrSet(rName, "agent_status"),
					resource.TestCheckResourceAttrSet(rName, "os_type"),
					resource.TestCheckResourceAttrSet(rName, "status"),
					resource.TestCheckResourceAttrSet(rName, "detect_result"),
					resource.TestCheckResourceAttrSet(rName, "asset_value"),
				),
			},
			{
				Config: testAccHostProtection_update(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "version", "hss.version.enterprise"),
					resource.TestCheckResourceAttr(rName, "charging_mode", "on_demand"),
					resource.TestCheckResourceAttrSet(rName, "host_name"),
					resource.TestCheckResourceAttrSet(rName, "private_ip"),
					resource.TestCheckResourceAttrSet(rName, "agent_id"),
					resource.TestCheckResourceAttrSet(rName, "agent_status"),
					resource.TestCheckResourceAttrSet(rName, "os_type"),
					resource.TestCheckResourceAttrSet(rName, "status"),
					resource.TestCheckResourceAttrSet(rName, "detect_result"),
					resource.TestCheckResourceAttrSet(rName, "asset_value"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"quota_id", "is_wait_host_available",
				},
			},
		},
	})
}

func testAccHostProtection_base(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_compute_instance_v2" "instance" {
  name              = "%[2]s"
  description       = "my_desc"
  availability_zone = "%[3]s"

  image_name = "Standard_Debian_11_latest"
  flavor_id  = "s3.large.2"

  metadata = {
    foo = "bar"
  }
  network {
    uuid = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  }

  tags = {
    hss = "acc-test"
  }

  user_data = <<-EOF
    #!/bin/bash
    curl -O 'https://hss-agent-podlb.eu-de.otc.t-systems.com:10180/package/agent/linux/x86/hostguard.x86_64.deb'
    echo 'MASTER_IP=hss-agent-podlb.eu-de.otc.t-systems.com:10180' > hostguard_setup_config.conf
    echo 'SLAVE_IP=hss-agent-slave.eu-de.otc-tsi.de:10180' >> hostguard_setup_config.conf
    echo 'ORG_ID=' >> hostguard_setup_config.conf
    dpkg -i hostguard.x86_64.deb
    rm -f hostguard_setup_config.conf
    rm -f hostguard.x86_64.deb
  EOF

  stop_before_destroy = true
}

resource "opentelekomcloud_hss_host_group_v5" "group" {
  name     = "%[2]s"
  host_ids = [opentelekomcloud_compute_instance_v2.instance.id]
}

`, common.DataSourceSubnet, name, env.OS_AVAILABILITY_ZONE)
}

func testAccHostProtection_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_hss_host_protection_v5" "protection" {
  host_id                = opentelekomcloud_compute_instance_v2.instance.id
  version                = "hss.version.premium"
  charging_mode          = "on_demand"
  is_wait_host_available = true
}
`, testAccHostProtection_base(name))
}

func testAccHostProtection_update(name string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_hss_host_protection_v5" "protection" {
  host_id       = opentelekomcloud_compute_instance_v2.instance.id
  version       = "hss.version.enterprise"
  charging_mode = "on_demand"
}
`, testAccHostProtection_base(name))
}
