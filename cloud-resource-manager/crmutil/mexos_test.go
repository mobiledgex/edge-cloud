package crmutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var testRootLB = "mexlb.tdg.mobiledgex.net"
var testExtNet = "external-network-shared"
var testMexDir = os.Getenv("HOME") + "/.mobiledgex"

var FlavorData = []edgeproto.Flavor{
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.medium",
		},
		Ram:   4096,
		Vcpus: 4,
		Disk:  4,
	},
}

var ClusterData = []edgeproto.Cluster{
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Pokemons",
		},
		Flavor: FlavorData[0].Key,
		Nodes:  3,
	},
}

var ClusterInstData = []edgeproto.ClusterInst{
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[0].Key,
		},
		Flavor: ClusterData[0].Flavor,
		Nodes:  ClusterData[0].Nodes,
	},
}

var DevData = []edgeproto.Developer{ //nolint
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Niantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.niantic.com",
	},
}

var AppData = []edgeproto.App{ //nolint
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		ImageType:   edgeproto.ImageType_ImageTypeDocker,
		AccessLayer: edgeproto.AccessLayer_AccessLayerL7,
		Flavor:      FlavorData[0].Key,
		Cluster:     ClusterData[0].Key,
		ImagePath:   "pokemon/go:1.0.0",
	},
}

//XXX AppInstData has CloudletKey inside AppInstKey.
//  But the ClusterInstKey is not part of that. It is part of AppInstData.
//  There is no image repo URI.
//  AccessLayer in AppData.  But not in AppInstData. No way to tell L4 or L7.
var AppInstData = []edgeproto.AppInst{ //nolint
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          1,
		},
		ClusterInstKey: ClusterInstData[0].Key,
		CloudletLoc:    CloudletData[0].Location,
		Uri:            "https://mexlb.tdg.mobiledgex.net/pokemon-go", //XXX
		ImagePath:      AppData[0].ImagePath,
		ImageType:      AppData[0].ImageType,
	},
}

var IsValidMEXOSTest = false

func TestValidateMEXOSEnv(t *testing.T) {
	osUser := os.Getenv("OS_USERNAME")
	osPass := os.Getenv("OS_PASSWORD")
	osTenant := os.Getenv("OS_TENANT")
	osAuthURL := os.Getenv("OS_AUTH_URL")
	osRegion := os.Getenv("OS_REGION_NAME")
	osCACert := os.Getenv("OC_CACERT")

	if osUser != "" && osPass != "" && osTenant != "" &&
		osAuthURL != "" && osRegion != "" && osCACert != "" {
		IsValidMEXOSTest = ValidateMEXOSEnv(true)
	}
}

func TestAddFlavor(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	for _, f := range ValidClusterFlavors {
		err := AddFlavor(f)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestLBAddRoute(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	name := os.Getenv("MEX_TEST_MN")
	err := LBAddRoute(testRootLB, testExtNet, name)
	if err != nil {
		t.Error(err)
	}
}

func TestCopySSHCredential(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	err := CopySSHCredential(testRootLB, testExtNet, "root")
	if err != nil {
		t.Error(err)
	}
}

func TestFindClusterWithKey(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	key := os.Getenv("MEX_TEST_CLUSTER_KEY")
	name, err := FindClusterWithKey(key)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("found", name)
}

func TestAddPathReverseProxy(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	path := "test-docker-nginx"
	origin := "http://localhost:80" //simple case

	errs := AddPathReverseProxy(testRootLB, path, origin)
	if errs != nil {
		t.Error(errs[0])
	}
}

func TestSetKubernetesConfigmapValues(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	err := SetKubernetesConfigmapValues(testRootLB, "mex-k8s", "test-configmap-1", "key1=val1", "key2=val2")

	if err != nil {
		t.Error(err)
	}
}

func TestGetKubernetesConfigmapYAML(t *testing.T) {
	if !IsValidMEXOSTest {
		return
	}
	out, err := GetKubernetesConfigmapYAML(testRootLB, "mex-k8s", "test-configmap-1")

	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(out)
}
