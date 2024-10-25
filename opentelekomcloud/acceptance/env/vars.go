package env

import (
	"os"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

var (
	OsFlavorID           = flavorID()
	OsImageName          = imageName()
	OsExtNetworkName     = extNetworkName()
	OS_REGION_NAME       string
	OS_ACCESS_KEY        = os.Getenv("OS_ACCESS_KEY")
	OS_SECRET_KEY        = os.Getenv("OS_SECRET_KEY")
	OS_AVAILABILITY_ZONE = os.Getenv("OS_AVAILABILITY_ZONE")
	OsSubnetName         = os.Getenv("OS_SUBNET_NAME")
	OS_KEYPAIR_NAME      = os.Getenv("OS_KEYPAIR_NAME")
	OS_KMS_ID            = os.Getenv("OS_KMS_ID")
	OsKmsName            = os.Getenv("OS_KMS_NAME")
	OS_BMS_FLAVOR_NAME   = os.Getenv("OS_BMS_FLAVOR_NAME")
	OS_TO_TENANT_ID      = os.Getenv("OS_TO_TENANT_ID")
	OS_TENANT_NAME       = GetTenantName()
	OS_PROJECT_ID        = os.Getenv("OS_PROJECT_ID")
	OS_DEST_REGION       = os.Getenv("OS_DEST_REGION")
	OS_DEST_PROJECT_ID   = os.Getenv("OS_DEST_PROJECT_ID")
	OS_DC_HOSTING_ID     = os.Getenv("OS_DC_HOSTING_ID")
)

func flavorID() string {
	if f := os.Getenv("OS_FLAVOR_ID"); f != "" {
		return f
	}
	return "s2.xlarge.4" // 4 vCPUs + 16GB RAM
}

func extNetworkName() string {
	if nw := os.Getenv("OS_EXT_NETWORK_NAME"); nw != "" {
		return nw
	}
	return "admin_external_net" // value valid for OTC PROD, both eu-de and eu-nl
}

func imageName() string {
	if image := os.Getenv("OS_IMAGE_NAME"); image != "" {
		return image
	}
	return "Standard_Debian_11_latest" // value valid for OTC PROD, both eu-de and eu-nl
}

func GetTenantName() cfg.ProjectName {
	tn := os.Getenv("OS_TENANT_NAME")
	if tn == "" {
		tn = os.Getenv("OS_PROJECT_NAME")
	}
	return cfg.ProjectName(tn)
}
