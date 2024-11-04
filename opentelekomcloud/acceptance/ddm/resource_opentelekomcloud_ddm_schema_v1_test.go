package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/schemas"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const ddmSchemaResourceName = "opentelekomcloud_ddm_schema_v1.schema_1"

func getDDMSchemaResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := cfg.DdmV1Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating DDMv1 Client: %s", err)
	}
	schema, err := schemas.QuerySchemaDetails(client, state.Primary.Attributes["instance_id"], state.Primary.Attributes["name"])
	if err != nil {
		return nil, fmt.Errorf("error fetching ddm schema: %s", err)
	}
	return schema.Database, nil
}

func TestAccDdmSchemasV1_basic(t *testing.T) {
	var schema schemas.GetDatabaseResponseBean
	rc := common.InitResourceCheck(
		ddmSchemaResourceName,
		&schema,
		getDDMSchemaResourceFunc,
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDdmSchemaV1Basic,
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(ddmSchemaResourceName, "name", "ddm_schema"),
				),
			},
			{
				ResourceName:      ddmSchemaResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"purge_rds_on_delete",
					"rds",
					"updated_at",
				},
			},
		},
	})
}

var testAccDdmSchemaV1Basic = `
resource "opentelekomcloud_ddm_schema_v1" "schema_1" {
  name         = "ddm_schema"
  instance_id  = "b4cd6aeb0b7445d3bf271457c6941544in09"
  shard_mode   = "cluster"
  shard_number = 8
  shard_unit   = 8
  rds {
    id             = "55d93e249b77461b81f990fa805db3f3in01"
    admin_username = "root"
    admin_password = "test-acc-password-1"
  }
  purge_rds_on_delete = true
}
`
