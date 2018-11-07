package crmutil

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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
		ImagePath:      AppData[0].ImagePath,
		ImageType:      AppData[0].ImageType,
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

func TestGenDeployment(t *testing.T) {
	rootLB := &MEXRootLB{
		Name: "rootLB",
	}
	clusterInst := &edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: "cluster1",
			},
			CloudletKey: edgeproto.CloudletKey{
				OperatorKey: edgeproto.OperatorKey{
					Name: "op1",
				},
				Name: "cloudlet1",
			},
		},
		Flavor: edgeproto.ClusterFlavorKey{
			Name: "c1.medium",
		},
	}
	appInst := &edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				DeveloperKey: edgeproto.DeveloperKey{
					Name: "dev1",
				},
				Name:    "app1",
				Version: "1.0.0",
			},
			CloudletKey: edgeproto.CloudletKey{
				OperatorKey: edgeproto.OperatorKey{
					Name: "op1",
				},
				Name: "cloudlet1",
			},
		},
		ClusterInstKey: clusterInst.Key,
		Uri:            "cloudlet1.op1.mobiledgex.net",
		ImagePath:      "registry.mobiledgex.net:5000/dev1/app1/image",
		ImageType:      edgeproto.ImageType_ImageTypeDocker,
		MappedPorts: []dme.AppPort{
			dme.AppPort{
				Proto:        dme.LProto_LProtoUDP,
				InternalPort: 10001,
				PublicPort:   3001,
			},
			dme.AppPort{
				Proto:        dme.LProto_LProtoTCP,
				InternalPort: 55,
				PublicPort:   3002,
			},
			dme.AppPort{
				Proto:        dme.LProto_LProtoHTTP,
				InternalPort: 8080,
				PublicPort:   443,
				PublicPath:   "dev1/app11000/http8080",
			},
		},
		Config:   "app1 -d -v",
		IpAccess: edgeproto.IpAccess_IpAccessShared,
	}
	mf, err := fillAppTemplate(rootLB, appInst, clusterInst)
	fmt.Printf("mf: %v\n", mf)
	if err != nil {
		t.Error(err)
		return
	}
	kubeManifest, err := genKubeManifest(mf)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(kubeManifest)

	// check template read from file
	file, err := os.Create("template")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = file.WriteString(kubeManifestSimpleShared)
	if err != nil {
		t.Error(err)
		return
	}
	file.Sync()
	file.Close()
	defer os.Remove("template")
	mf.Spec.KubeManifestTemplate = "file://template"
	kubeManifest2, err := genKubeManifest(mf)
	if err != nil {
		t.Error(err)
		return
	}
	if strings.Compare(kubeManifest, kubeManifest2) != 0 {
		t.Error("file template not equal")
		fmt.Println(kubeManifest2)
		return
	}

	// check template read from http
	go func() {
		http.HandleFunc("/template", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(kubeManifestSimpleShared))
		})
		if err := http.ListenAndServe(":12345", nil); err != nil {
			t.Error(err)
		}
	}()
	mf.Spec.KubeManifestTemplate = "http://localhost:12345/template"
	kubeManifest3, err := genKubeManifest(mf)
	if err != nil {
		t.Error(err)
		return
	}
	if strings.Compare(kubeManifest, kubeManifest3) != 0 {
		t.Error("http template not equal")
		fmt.Println(kubeManifest3)
		return
	}
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
