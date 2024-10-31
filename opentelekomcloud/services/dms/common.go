package dms

import (
	"encoding/json"
	"log"
	"strings"
)

const (
	errCreationClient   = "error creating OpenTelekomCloud DMSv1 client: %w"
	errCreationClientV2 = "error creating OpenTelekomCloud DMSv2 client: %w"
	dmsClientV2         = "dms-v2-client"
)

func MarshalValue(i interface{}) string {
	if i == nil {
		return ""
	}

	jsonRaw, err := json.Marshal(i)
	if err != nil {
		log.Printf("[WARN] failed to marshal %#v: %s", i, err)
		return ""
	}

	return strings.Trim(string(jsonRaw), `"`)
}
