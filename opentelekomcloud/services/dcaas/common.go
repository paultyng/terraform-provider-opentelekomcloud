package dcaas

const (
	errCreateClientV2 = "error creating OpenTelekomCloud DCaaSv2 client: %w"
	errCreateClientV3 = "error creating OpenTelekomCloud DCaaSv3 client: %w"
	keyClientV2       = "dcaas-v2-client"
	keyClientV3       = "dcaas-v3-client"
)

const egDeprecated = "This resource is not longer supported. Please use opentelekomcloud_dc_virtual_gateway_v2 with local_ep_group block instead."

func GetEndpoints(e []interface{}) []string {
	endpoints := make([]string, 0)
	for _, val := range e {
		endpoints = append(endpoints, val.(string))
	}
	return endpoints
}
