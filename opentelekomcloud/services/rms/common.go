package rms

import (
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const (
	errCreationRMSV1Client = "error creating OpenTelekomCloud RMS v1 client: %w"
	rmsClientV1            = "rms-v1-client"
)

func GetRmsDomainId(client *golangsdk.ServiceClient, config *cfg.Config) (domainID string) {
	domainID = client.DomainID

	if domainID == "" {
		domainID = config.DomainClient.AKSKAuthOptions.DomainID
	}

	return
}
