package hss

const (
	errCreationV5Client = "error creating OpenTelekomCloud HSS v5 client: %w"
	hssClientV5         = "hss-v5-client"
)

type ProtectStatus string

const (
	ProtectStatusClosed ProtectStatus = "closed"
	ProtectStatusOpened ProtectStatus = "opened"
)

const (
	protectionVersionNull         string = "hss.version.null"
	hostAgentStatusOnline         string = "online"
	getProtectionHostNeedRetryMsg string = "The host cannot be found temporarily, please try again later."
)
