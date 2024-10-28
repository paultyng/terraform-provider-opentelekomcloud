package rms

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rms/advanced"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/rms"
)

func getAdvancedQueryResourceFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.RmsV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating RMS V1 client: %s", err)
	}

	domainId := rms.GetRmsDomainId(client, conf)
	return advanced.GetQuery(client, domainId, state.Primary.ID)
}

func TestAccAdvancedQuery_basic(t *testing.T) {
	var obj interface{}

	name := "rms-adv-query-" + acctest.RandString(5)
	rName := "opentelekomcloud_rms_advanced_query_v1.test"

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getAdvancedQueryResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAdvancedQuery_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "expression", "select colume_1 from table_1"),
					resource.TestCheckResourceAttr(rName, "description", "test_description"),
				),
			},
			{
				Config: testAdvancedQuery_basic_update(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "expression", "update table_1 set volume_1 = 5"),
					resource.TestCheckResourceAttr(rName, "description", "test_description_update"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAdvancedQuery_basic(name string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_rms_advanced_query_v1" "test" {
  name        = "%s"
  expression  = "select colume_1 from table_1"
  description = "test_description"
}
`, name)
}

func testAdvancedQuery_basic_update(name string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_rms_advanced_query_v1" "test" {
  name        = "%s"
  expression  = "update table_1 set volume_1 = 5"
  description = "test_description_update"
}
`, name)
}
