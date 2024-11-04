package ddm

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	ddmv1instances "github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/instances"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/schemas"
)

const (
	errCreationV1Client = "error creating OpenTelekomCloud DDMv1 client: %w"
	errCreationV2Client = "error creating OpenTelekomCloud DDMv2 client: %w"
	errCreationV3Client = "error creating OpenTelekomCloud DDMv3 client: %w"
	keyClientV1         = "ddm-v1-client"
	keyClientV2         = "ddm-v2-client"
	keyClientV3         = "ddm-v3-client"
)

func instanceStateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instanceList, err := ddmv1instances.QueryInstances(client, ddmv1instances.QueryInstancesOpts{})
		if err != nil {
			return nil, "Error retrieving DDM v1 Instances", err
		}
		if len(instanceList) == 0 {
			return nil, "DELETED", nil
		}
		for _, instance := range instanceList {
			if instance.Id == instanceID {
				return instance, instance.Status, nil
			}
		}
		return nil, "DELETED", nil
	}
}

func schemaStateRefreshFunc(client *golangsdk.ServiceClient, instanceID string, schemaName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		schemaList, err := schemas.QuerySchemas(client, instanceID, schemas.QuerySchemasOpts{})
		if err != nil {
			return nil, "Error retrieving DDM v1 schemas", err
		}
		if len(schemaList) == 0 {
			return nil, "DELETED", nil
		}
		for _, schema := range schemaList {
			if schema.Name == schemaName {
				return schema, schema.Status, nil
			}
		}
		return nil, "DELETED", nil
	}
}
