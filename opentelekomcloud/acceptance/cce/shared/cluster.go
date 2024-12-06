package shared

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/subnets"
	th "github.com/opentelekomcloud/gophertelekomcloud/testhelper"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/cce"
)

const sharedClusterName = "shared-cluster"

var (
	sharedClusterID     string
	createClusterOnce   sync.Once
	deleteClusterOnce   sync.Once
	sharedClusterUsages int32 = 0
)

func createSharedCluster(t *testing.T) string {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Shared cluster can only be used in acceptance tests")
	}

	t.Helper()
	v := atomic.AddInt32(&sharedClusterUsages, 1)
	t.Logf("Cluster is required by the test. %d test(s) are using cluster.", v)

	createClusterOnce.Do(func() {
		subnet := getSharedSubnet(t)

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.CceV3Client(env.OS_REGION_NAME)
		th.AssertNoErr(t, err)

		th.AssertNoErr(t, quotas.CCEClusterQuota.Acquire())

		// check if cluster already exist

		getCluster, err := clusters.List(client, clusters.ListOpts{
			Name: sharedClusterName,
		})
		th.AssertNoErr(t, err)
		if len(getCluster) > 0 {
			sharedClusterID = getCluster[0].Metadata.Id
		} else {
			sharedClusterID = CreateSharedCluster(t, client, subnet)
		}
	})

	if sharedClusterID == "" {
		t.Fatal("no shared cluster ID is available, cluster creation failed")
	}

	return sharedClusterID
}

func deleteSharedCluster(t *testing.T) {
	t.Helper()
	if v := atomic.AddInt32(&sharedClusterUsages, -1); v > 0 {
		t.Logf("Cluster is released by the test. %d test(s) are still using cluster.", v)
		return
	}
	t.Log("Cluster usage is 0 now, ready to delete the cluster")

	deleteClusterOnce.Do(func() {
		t.Log("starting deleting shared cluster")

		config := common.TestAccProvider.Meta().(*cfg.Config)
		client, err := config.CceV3Client(env.OS_REGION_NAME)
		th.AssertNoErr(t, err)

		th.AssertNoErr(t, clusters.Delete(client, sharedClusterID, clusters.DeleteQueryParams{}))

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"Deleting", "Available", "Unavailable"},
			Target:     []string{"Deleted"},
			Refresh:    cce.WaitForCCEClusterDelete(client, sharedClusterID),
			Timeout:    10 * time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(context.Background())
		th.AssertNoErr(t, err)

		sharedClusterID = ""
		quotas.CCEClusterQuota.Release()
	})
}

// DataSourceCluster - can be used as data.opentelekomcloud_cce_cluster_v3.cluster.id
var DataSourceCluster = fmt.Sprintf(`
data "opentelekomcloud_cce_cluster_v3" "cluster" {
  name = "%s"
}
`, sharedClusterName)

const DataSourceClusterName = "data.opentelekomcloud_cce_cluster_v3.cluster"

func BookCluster(t *testing.T) {
	t.Helper()
	createSharedCluster(t)
	t.Cleanup(func() { deleteSharedCluster(t) })
}

func CreateSharedCluster(t *testing.T, client *golangsdk.ServiceClient, subnet *subnets.Subnet) string {
	t.Log("starting creating shared cluster")

	cluster, err := clusters.Create(client, clusters.CreateOpts{
		Kind:       "Cluster",
		ApiVersion: "v3",
		Metadata: clusters.CreateMetaData{
			Name: sharedClusterName,
		},
		Spec: clusters.Spec{
			Category:    "Turbo",
			Type:        "VirtualMachine",
			Flavor:      "cce.s2.small",
			Description: "Shared cluster for CCE acceptance tests",
			HostNetwork: clusters.HostNetworkSpec{
				VpcId:    subnet.VpcID,
				SubnetId: subnet.ID,
			},
			ContainerNetwork: clusters.ContainerNetworkSpec{
				Mode: "eni",
			},
			EniNetwork: &clusters.EniNetworkSpec{
				SubnetId: subnet.SubnetID,
				Cidr:     subnet.CIDR,
			},
			Authentication: clusters.AuthenticationSpec{
				Mode:                "rbac",
				AuthenticatingProxy: make(map[string]string),
			},
			KubernetesSvcIpRange: "10.247.0.0/16",
		},
	})

	th.AssertNoErr(t, err)
	sharedClusterID = cluster.Metadata.Id

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating"},
		Target:     []string{"Available"},
		Refresh:    cce.WaitForCCEClusterActive(client, sharedClusterID),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(context.Background())
	th.AssertNoErr(t, err)

	return sharedClusterID
}
