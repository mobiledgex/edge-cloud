package crmutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var testRootLB = "mexlb.gddt.mobiledgex.net"
var testExtNet = "external-network-shared"
var testMexDir = os.Getenv("HOME") + "/.mobiledgex"

var FlavorData = []edgeproto.Flavor{
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "m4.medium",
		},
		Ram:   4096,
		Vcpus: 4,
		Disk:  4,
	},
}

var ClusterFlavorData = []edgeproto.ClusterFlavor{
	edgeproto.ClusterFlavor{
		Key: edgeproto.ClusterFlavorKey{
			Name: "x1.medium",
		},
		NodeFlavor:   FlavorData[0].Key,
		MasterFlavor: FlavorData[0].Key,
		NumNodes:     2,
		MaxNodes:     2,
		NumMasters:   1,
	},
}

var ClusterData = []edgeproto.Cluster{
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Pillimos",
		},
		DefaultFlavor: ClusterFlavorData[0].Key,
	},
}

var ClusterInstData = []edgeproto.ClusterInst{
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[0].Key,
		},
		Flavor: ClusterData[0].DefaultFlavor,
	},
}

var DevData = []edgeproto.Developer{ //nolint
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Atlantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.atlantic.com",
	},
}

var AppData = []edgeproto.App{ //nolint
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pillimo Go!",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_ImageTypeDocker,
		DefaultFlavor: FlavorData[0].Key,
		Cluster:       ClusterData[0].Key,
		ImagePath:     "pillimo/go:1.0.0",
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
		Uri:            "https://mexlb.gddt.mobiledgex.net/pillimo-go", //XXX
		Flavor:         FlavorData[0].Key,
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

func TestFormNginxProxyRequest(t *testing.T) {
	ports := []PortDetail{
		{MexProto: "LProtoHTTP", PublicPort: 8888, InternalPort: 8888, PublicPath: "test1"},
		{MexProto: "LProtoTCP", PublicPort: 8889, InternalPort: 8889, PublicPath: "test2"},
		{MexProto: "LProtoUDP", PublicPort: 8887, InternalPort: 8887, PublicPath: "test3"},
	}

	pl, err := FormNginxProxyRequest(ports, "10.0.0.1", "test123")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(*pl)
}

func TestGetSvcExternalIP(t *testing.T) {
	home := os.Getenv("HOME")
	path := home + "/.mobiledgex/aks-testcluster.kubeconfig"

	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("skip getSvcExternalIP test")
		return
	}
	eip, err := getSvcExternalIP("testapp", path)
	if err != nil {
		t.Errorf("error %v", err)
		return
	}
	fmt.Println("external IP", eip)
}

func TestLookupDNS(t *testing.T) {
	ipa, err := LookupDNS("google.com")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(ipa)
}

func TestWaitforDNSRegistration(t *testing.T) {
	err := WaitforDNSRegistration("google.com")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("google.com ok")
	dnsRegisterRetryDelay = time.Millisecond
	err = WaitforDNSRegistration("asdfasdfasdfasdfasdfasdfasdfadsfadf.com")
	if err != nil {
		fmt.Println("timeout as expected")
		fmt.Println("returned err", err)
	} else {
		t.Errorf("unexpected success")
	}
}

func TestURI2FQDN(t *testing.T) {
	fqdn := uri2fqdn("http://google.com")
	if fqdn != "google.com" {
		t.Errorf("expected google.com, got %s", fqdn)
	}
	fqdn = uri2fqdn("https://google.com")
	if fqdn != "google.com" {
		t.Errorf("expected google.com,got %s", fqdn)
	}
}

func TestValidateURI(t *testing.T) {
	err := validateURI("http://google.com")
	if err != nil {
		t.Error(err)
	}
	err = validateURI("google.com")
	if err != nil {
		t.Error(err)
	}
}
