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

func getImsImageShareAcceptResourceFunc(cfg *cfg.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := cfg.ImageV2Client(env.OS_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud image V2 client: %s", err)
	}
	imgs, err := images.ListImages(c,
		images.ListImagesOpts{
			Id:        state.Primary.Attributes["image_id"],
			ImageType: "shared"})
	if err != nil || len(imgs) < 1 {
		return nil, fmt.Errorf("unable to retrieve images: %s", err)
	}
	img := imgs[0]
	return img, nil
}

func TestAccImsImageShareAccept_basic(t *testing.T) {
	sharedImageID := os.Getenv("OS_SHARED_IMAGE_ID")
	if sharedImageID == "" {
		t.Skip("OS_SHARED_IMAGE_ID is empty, but test requires")
	}
	var obj interface{}
	rName := "opentelekomcloud_ims_image_share_accept_v1.act"
	rc := common.InitResourceCheck(
		rName,
		&obj,
		getImsImageShareAcceptResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testImsImageShareAccept_basic(sharedImageID),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "image_id", sharedImageID),
				),
			},
		},
	})
}

func testImsImageShareAccept_basic(sharedImageID string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_ims_image_share_accept_v1" "act" {
 image_id = "%s"
}
`, sharedImageID)
}
