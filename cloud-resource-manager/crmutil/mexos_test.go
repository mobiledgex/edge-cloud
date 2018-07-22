package crmutil

import (
	"fmt"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"os"
	"testing"
)

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

var DevData = []edgeproto.Developer{
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Niantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.niantic.com",
	},
}

var AppData = []edgeproto.App{
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
var AppInstData = []edgeproto.AppInst{
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          1,
		},
		ClusterInstKey: ClusterInstData[0].Key,
		CloudletLoc:    CloudletData[0].Location,
		Uri:            "https://mex-lb-1.mobiledgex.net/pokemon-go", //XXX
		ImagePath:      AppData[0].ImagePath,
		ImageType:      AppData[0].ImageType,
	},
}

func TestAddFlavor(t *testing.T) {
	for _, f := range ValidClusterFlavors {
		err := AddFlavor(f)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCreateClusterFromClusterInstData(t *testing.T) {
	TestAddFlavor(t)

	for _, c := range ClusterInstData {
		err := CreateClusterFromClusterInstData(eRootLBName, &c)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestDeleteClusterByName(t *testing.T) {
	for _, c := range ClusterInstData {
		err := DeleteClusterByName(eRootLBName, c.Key.ClusterKey.Name)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestEnableRootLB(t *testing.T) {
	err := EnableRootLB(eRootLBName)
	if err != nil {
		t.Error(err)
	}
}

func TestWaitForRootLB(t *testing.T) {
	err := WaitForRootLB(eRootLBName)

	if err != nil {
		t.Error(err)
	}
}

func TestRunMEXAgent(t *testing.T) {
	err := RunMEXAgent(eRootLBName, false)
	if err != nil {
		t.Error(err)
	}
}

func TestRemoveMexAgent(t *testing.T) {
	err := RemoveMEXAgent(eRootLBName)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateMexAgent(t *testing.T) {
	err := UpdateMEXAgent(eRootLBName)
	if err != nil {
		t.Error(err)
	}
}

func TestLBAddRoute(t *testing.T) {
	name := os.Getenv("MEX_TEST_MN")
	err := LBAddRoute(eRootLBName, name)
	if err != nil {
		t.Error(err)
	}
}

func TestCopySSHCredential(t *testing.T) {
	err := CopySSHCredential(eRootLBName, eMEXExternalNetwork, "root")
	if err != nil {
		t.Error(err)
	}
}

func TestIsClusterReady(t *testing.T) {
	name := os.Getenv("MEX_TEST_MN")
	ready, err := IsClusterReady(eRootLBName, name, "x1.medium")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ready", ready)
}

func TestCopyKubeConfig(t *testing.T) {
	name := os.Getenv("MEX_TEST_MN")
	err := CopyKubeConfig(eRootLBName, name)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFindClusterWithKey(t *testing.T) {
	key := os.Getenv("MEX_TEST_CLUSTER_KEY")
	name, err := FindClusterWithKey(key)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("found", name)
}

func TestCreateDockerApp(t *testing.T) {
	err := CreateDockerApp(eRootLBName, "test-docker-nginx", os.Getenv("MEX_TEST_MN"), "x1.medium",
		"docker.io", eRootLBName+"/test-docker-nginx", "nginx", "", "", "L7")
	if err != nil {
		t.Error(err)
	}
}

func TestAddPathReverseProxy(t *testing.T) {
	path := "test-docker-nginx"
	origin := "http://localhost:80" //simple case

	errs := AddPathReverseProxy(eRootLBName, path, origin)
	if errs != nil {
		t.Error(errs[0])
	}
}

func TestCreateKubernetesApp(t *testing.T) {
	err := CreateKubernetesApp(eRootLBName, "mex-k8s", "nginx", "https://k8s.io/examples/application/deployment.yaml")
	if err != nil {
		t.Error(err)
	}
}

func TestCreateKubernetesNamespace(t *testing.T) {
	err := CreateKubernetesNamespace(eRootLBName, "mex-k8s", "https://k8s.io/examples/admin/namespace-prod.json")

	if err != nil {
		t.Error(err)
	}
}

func TestSetKubernetesConfigmapValues(t *testing.T) {
	err := SetKubernetesConfigmapValues(eRootLBName, "mex-k8s", "test-configmap-1", "key1=val1", "key2=val2")

	if err != nil {
		t.Error(err)
	}
}

func TestGetKubernetesConfigmapYAML(t *testing.T) {
	out, err := GetKubernetesConfigmapYAML(eRootLBName, "mex-k8s", "test-configmap-1")

	if err != nil {
		t.Error(err)
	}
	fmt.Println(out)
}
