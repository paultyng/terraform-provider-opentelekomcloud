package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ims/v2/images"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func getImsImageShareResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := cfg.ImageV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud image V2 client: %s", err)
	}
	imgs, err := images.ListImages(c,
		images.ListImagesOpts{
			Id:        state.Primary.ID,
			ImageType: "shared"})
	if err != nil || len(imgs) < 1 {
		return nil, fmt.Errorf("unable to retrieve images: %s", err)
	}
	img := imgs[0]
	return img, nil
}

func TestAccImsImageShare_basic(t *testing.T) {
	privateImageID := os.Getenv("OS_PRIVATE_IMAGE_ID")
	shareProjectID := os.Getenv("OS_PROJECT_ID_2")
	shareProjectID2 := os.Getenv("OS_PROJECT_ID_3")
	if privateImageID == "" || shareProjectID == "" {
		t.Skip("OS_PRIVATE_IMAGE_ID or OS_PROJECT_ID_2 are empty, but test requires")
	}

	var obj interface{}

	rName := "opentelekomcloud_ims_image_share_v1.share_1"

	rc := common.InitResourceCheck(
		rName,
		&obj,
		getImsImageShareResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testImsImageShare_basic(privateImageID, shareProjectID),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
			{
				Config: testImsImageShare_update(privateImageID, shareProjectID, shareProjectID2),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
				),
			},
		},
	})
}

func testImsImageShare_basic(privateImageID, projectToShare string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_ims_image_share_v1" "share_1" {
  source_image_id    = "%[1]s"
  target_project_ids = ["%[2]s"]
}
`, privateImageID, projectToShare)
}

func testImsImageShare_update(privateImageID, projectToShare, projectToShare2 string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_ims_image_share_v1" "share_1" {
  source_image_id    = "%[1]s"
  target_project_ids = ["%[2]s", "%[3]s"]
}
`, privateImageID, projectToShare, projectToShare2)
}
