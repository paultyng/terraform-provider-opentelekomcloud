package obs

import "github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"

const (
	errCreationClient = "error creating OBS client: %w"
)

func getDomainID(cfg *cfg.Config) (domainId string) {
	domainId = cfg.DomainID
	if domainId == "" {
		domainId = cfg.DomainClient.DomainID
	}
	return
}
