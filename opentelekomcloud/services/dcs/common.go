package dcs

const (
	errCreationClient = "error creating OpenTelekomCloud DCSv1 client: %w"
	dcsClientV2       = "dcs-v2-client"
	floatBitSize      = 64
)

var (
	redisEngineVersion = map[string]bool{
		"4.0": true,
		"5.0": true,
		"6.0": true,
	}

	operateErrorCode = map[string]bool{
		// current state not support
		"DCS.4026": true,
		// instance status is not running
		"DCS.4049": true,
		// backup
		"DCS.4096": true,
		// restore
		"DCS.4097": true,
		// restart
		"DCS.4111": true,
		// resize
		"DCS.4113": true,
		// change config
		"DCS.4114": true,
		// change password
		"DCS.4115": true,
		// upgrade
		"DCS.4116": true,
		// rollback
		"DCS.4117": true,
		// create
		"DCS.4118": true,
		// freeze
		"DCS.4120": true,
		// creating/restarting
		"DCS.4975": true,
	}
)

type ctxType string
