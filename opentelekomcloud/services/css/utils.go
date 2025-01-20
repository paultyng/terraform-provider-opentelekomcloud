package css

import "github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/clusters"

const (
	clientError = `error creating CSSv1 client: %w`
	keyClientV1 = "css-v1-client"
)

var defaultDatastore = clusters.Datastore{
	Version: "7.6.2",
	Type:    "elasticsearch",
}
