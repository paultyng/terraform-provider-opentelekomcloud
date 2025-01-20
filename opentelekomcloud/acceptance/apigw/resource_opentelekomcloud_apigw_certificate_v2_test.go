package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/apigw/v2/cert"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getCertificateFunc(conf *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.APIGWV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating APIG v2 client: %s", err)
	}
	return cert.Get(client, state.Primary.ID)
}

func TestAccCertificate_basic(t *testing.T) {
	var (
		cer cert.CertificateResp

		rName             = "opentelekomcloud_apigw_certificate_v2.test"
		name              = fmt.Sprintf("apigw_cert_%s", acctest.RandString(5))
		updateName        = fmt.Sprintf("apigw_cert_%s_updated", acctest.RandString(5))
		oldCert, oldPk, _ = openstack.GenerateTestCertKeyPair("www.test.com")
		newCert, newPk, _ = openstack.GenerateTestCertKeyPair("www.test.com")
	)

	rc := common.InitResourceCheck(
		rName,
		&cer,
		getCertificateFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificate_basic_step1(name, oldCert, oldPk),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "type", "global"),
					resource.TestCheckResourceAttr(rName, "instance_id", "common"),
					resource.TestMatchResourceAttr(rName, "effected_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "expires_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "sans.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
				),
			},
			{
				Config: testAccCertificate_basic_step2(updateName, newCert, newPk),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", updateName),
					resource.TestCheckResourceAttr(rName, "type", "global"),
					resource.TestCheckResourceAttr(rName, "instance_id", "common"),
					resource.TestMatchResourceAttr(rName, "effected_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "expires_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "sans.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"content", "private_key",
				},
			},
		},
	})
}

func testAccCertificate_basic_step1(name, content, pk string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_apigw_certificate_v2" "test" {
  name        = "%[1]s"
  content     = <<-EOT
%[2]s
EOT
  private_key = <<-EOT
%[3]s
EOT
}
`, name, content, pk)
}

func testAccCertificate_basic_step2(name, content, pk string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_apigw_certificate_v2" "test" {
  name        = "%[1]s"
  content     = <<-EOT
%[2]s
EOT
  private_key = <<-EOT
%[3]s
EOT
}
`, name, content, pk)
}

func TestAccCertificate_instance(t *testing.T) {
	var (
		cer cert.CertificateResp

		rName             = "opentelekomcloud_apigw_certificate_v2.test"
		name              = fmt.Sprintf("apigw_cert_%s", acctest.RandString(5))
		updateName        = fmt.Sprintf("apigw_cert_%s_updated", acctest.RandString(5))
		oldCert, oldPk, _ = openstack.GenerateTestCertKeyPair("www.test.com")
		newCert, newPk, _ = openstack.GenerateTestCertKeyPair("www.test.com")
	)

	rc := common.InitResourceCheck(
		rName,
		&cer,
		getCertificateFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificate_instance_step1(name, oldCert, oldPk),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttrPair(rName, "instance_id", "opentelekomcloud_apigw_gateway_v2.gateway", "id"),
					resource.TestMatchResourceAttr(rName, "effected_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "expires_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "sans.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"content", "private_key",
				},
			},
			{
				Config: testAccCertificate_instance_step2(updateName, newCert, newPk),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", updateName),
					resource.TestCheckResourceAttrPair(rName, "instance_id", "opentelekomcloud_apigw_gateway_v2.gateway", "id"),
					resource.TestMatchResourceAttr(rName, "effected_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "expires_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "sans.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
				),
			},
		},
	})
}

func testAccCertificate_instance_general(name, content, privateKey string) string {
	return fmt.Sprintf(`
%s

resource "opentelekomcloud_apigw_certificate_v2" "test" {
  instance_id = opentelekomcloud_apigw_gateway_v2.gateway.id
  type        = "instance"
  name        = "%s"
  content     = <<-EOT
%s
EOT
  private_key = <<-EOT
%s
EOT
}
`, testAccAPIGWv2GatewayBasic(name), name, content, privateKey)
}

func testAccCertificate_instance_step1(name, content, pk string) string {
	return testAccCertificate_instance_general(name, content, pk)
}

func testAccCertificate_instance_step2(name, content, pk string) string {
	return testAccCertificate_instance_general(name, content, pk)
}

func TestAccCertificate_instanceWithRootCA(t *testing.T) {
	var (
		cer cert.CertificateResp

		rName             = "opentelekomcloud_apigw_certificate_v2.test"
		name              = fmt.Sprintf("apigw_cert_%s", acctest.RandString(5))
		updateName        = fmt.Sprintf("apigw_cert_%s_updated", acctest.RandString(5))
		oldCert, oldPk, _ = openstack.GenerateTestCertKeyPair("www.test.com")
		newCert, newPk, _ = openstack.GenerateTestCertKeyPair("www.test.com")
		oldRoot, _        = common.GenerateRootCA(oldPk)
		newRoot, _        = common.GenerateRootCA(newPk)
	)

	rc := common.InitResourceCheck(
		rName,
		&cer,
		getCertificateFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificate_instanceWithRootCA_step1(name, oldCert, oldPk, oldRoot),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "type", "instance"),
					resource.TestCheckResourceAttrPair(rName, "instance_id", "opentelekomcloud_apigw_gateway_v2.gateway", "id"),
					resource.TestMatchResourceAttr(rName, "effected_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "expires_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "sans.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
				),
			},
			{
				Config: testAccCertificate_instanceWithRootCA_step2(updateName, newCert, newPk, newRoot),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", updateName),
					resource.TestCheckResourceAttr(rName, "type", "instance"),
					resource.TestCheckResourceAttrPair(rName, "instance_id", "opentelekomcloud_apigw_gateway_v2.gateway", "id"),
					resource.TestMatchResourceAttr(rName, "effected_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "expires_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(rName, "sans.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"content", "private_key", "trusted_root_ca",
				},
			},
		},
	})
}

func testAccCertificate_instanceWithRootCA_general(name, content, privateKey, ca string) string {
	return fmt.Sprintf(`
%[1]s

resource "opentelekomcloud_apigw_certificate_v2" "test" {
  instance_id     = opentelekomcloud_apigw_gateway_v2.gateway.id
  type            = "instance"
  name            = "%[2]s"
  content         = <<-EOT
%[3]s
EOT
  private_key     = <<-EOT
%[4]s
EOT
  trusted_root_ca = <<-EOT
%[5]s
EOT
}
`, testAccAPIGWv2GatewayBasic(name), name, content, privateKey, ca)
}

func testAccCertificate_instanceWithRootCA_step1(name, content, pk, root string) string {
	return testAccCertificate_instanceWithRootCA_general(name, content,
		pk, root)
}

func testAccCertificate_instanceWithRootCA_step2(name, content, pk, root string) string {
	return testAccCertificate_instanceWithRootCA_general(name, content,
		pk, root)
}
