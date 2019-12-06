package cloudcommon

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	yaml "gopkg.in/yaml.v2"
)

var AppDNSRoot = "mobiledgex.net"

// special operator types
var OperatorGCP = "gcp"
var OperatorAzure = "azure"

var DeveloperSamsung = "Samsung"
var DeveloperMobiledgeX = "MobiledgeX"

const DefaultVMCluster string = "DefaultVMCluster"

// platform apps
var SamsungEnablingLayer = "SamsungEnablingLayer"

// cloudlet types
var CloudletKindOpenStack = "openstack"
var CloudletKindAzure = "azure"
var CloudletKindGCP = "gcp"
var CloudletKindDIND = "dind"
var CloudletKindFake = "fake"

var OperatingSystemMac = "mac"
var OperatingSystemLinux = "linux"

// network schemes for use by standalone deployments (e.g. DIND)
var NetworkSchemePublicIP = "publicip"
var NetworkSchemePrivateIP = "privateip"

// Metrics common variables - TODO move to edge-cloud-infra after metrics-exporter chagnes
var DeveloperMetricsDbName = "metrics"
var MEXPrometheusAppName = "MEXPrometheusAppName"
var PrometheusPort = int32(9090)
var LBMetricsPort = int32(65121)

// DIND script to pull from kubeadm-dind-cluster
var DindScriptName = "dind-cluster-v1.14.sh"

var MexNodePrefix = "mex-k8s-node-"

// TODO: these timeouts should be adjust based on target platform,
// as some platforms (azure, etc) may take much longer.
// These timeouts should be at least long enough for the controller and
// CRM to exchange an initial set of messages (i.e. 10 sec or so).
var CreateAppInstTimeout = 30 * time.Minute
var UpdateAppInstTimeout = 20 * time.Minute
var DeleteAppInstTimeout = 20 * time.Minute

var CreateClusterInstTimeout = 30 * time.Minute
var UpdateClusterInstTimeout = 20 * time.Minute
var DeleteClusterInstTimeout = 20 * time.Minute

// GCP limits to 40, Azure has issues above 54.  For consistency go with the lower limit
const MaxClusterNameLength = 40

// PlatformApps is the set of all special "platform" developers.   Key
// is DeveloperName:AppName.  Currently only Samsung's Enabling layer is included.
var platformApps = map[string]bool{
	DeveloperSamsung + ":" + SamsungEnablingLayer: true,
}

// Common regular expression for quoted strings parse
var QuotedStringRegex = regexp.MustCompile(`"(.*?)"`)

// IsPlatformApp true if the developer/app combo is a platform app
func IsPlatformApp(devname string, appname string) bool {
	_, ok := platformApps[devname+":"+appname]
	return ok
}

var AllocatedIpDynamic = "dynamic"

var RootLBL7Port int32 = 443

// GetRootLBFQDN gets the global Load Balancer's Fully Qualified Domain Name
// for apps using "shared" IP access.
func GetRootLBFQDN(key *edgeproto.CloudletKey) string {
	loc := util.DNSSanitize(key.Name)
	oper := util.DNSSanitize(key.OperatorKey.Name)
	return fmt.Sprintf("%s.%s.%s", loc, oper, AppDNSRoot)
}

// GetDedicatedLBFQDN gets the cluster-specific Load Balancer's Fully Qualified Domain Name
// for clusters using "dedicated" IP access.
func GetDedicatedLBFQDN(cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey) string {
	clust := util.DNSSanitize(clusterKey.Name)
	loc := util.DNSSanitize(cloudletKey.Name)
	oper := util.DNSSanitize(cloudletKey.OperatorKey.Name)
	return fmt.Sprintf("%s.%s.%s.%s", clust, loc, oper, AppDNSRoot)
}

// Get Fully Qualified Name for the App i.e. with developer & version info
func GetAppFQN(key *edgeproto.AppKey) string {
	app := util.DNSSanitize(key.Name)
	dev := util.DNSSanitize(key.DeveloperKey.Name)
	ver := util.DNSSanitize(key.Version)
	return fmt.Sprintf("%s%s%s", dev, app, ver)
}

// GetAppFQDN gets the app-specific Load Balancer's Fully Qualified Domain Name
// for apps using "dedicated" IP access.
func GetAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey) string {
	lb := GetDedicatedLBFQDN(cloudletKey, clusterKey)
	appFQN := GetAppFQN(&key.AppKey)
	return fmt.Sprintf("%s.%s", appFQN, lb)
}

// GetVMAppFQDN gets the app-specific Fully Qualified Domain Name
// for VM based apps
func GetVMAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey) string {
	lb := GetRootLBFQDN(cloudletKey)
	appFQN := GetAppFQN(&key.AppKey)
	return fmt.Sprintf("%s.%s", appFQN, lb)
}

func FqdnPrefix(svcName string) string {
	return svcName + "."
}

func ServiceFQDN(svcName, baseFQDN string) string {
	return fmt.Sprintf("%s%s", FqdnPrefix(svcName), baseFQDN)
}

// GetL7Path gets the L7 path for L7 access behind the "shared"
// global Load Balancer (reverse proxy). This only the path and
// does not include the Fqdn and port.
func GetL7Path(key *edgeproto.AppInstKey, internalPort int32) string {
	dev := util.DNSSanitize(key.AppKey.DeveloperKey.Name)
	app := util.DNSSanitize(key.AppKey.Name)
	ver := util.DNSSanitize(key.AppKey.Version)
	return fmt.Sprintf("%s/%s%s/p%d", dev, app, ver, internalPort)
}

// For the DME and CRM that require a cloudlet key to be specified
// at startup, this function parses the string argument.
func ParseMyCloudletKey(standalone bool, keystr *string, mykey *edgeproto.CloudletKey) {
	if standalone && *keystr == "" {
		// Use fake cloudlet
		*mykey = testutil.CloudletData[0].Key
		bytes, _ := json.Marshal(mykey)
		*keystr = string(bytes)
		return
	}

	if *keystr == "" {
		log.FatalLog("cloudletKey not specified")
	}

	err := json.Unmarshal([]byte(*keystr), mykey)
	if err != nil {
		err = yaml.Unmarshal([]byte(*keystr), mykey)
	}
	if err != nil {
		log.FatalLog("Failed to parse cloudletKey", "err", err)
	}

	err = mykey.ValidateKey()
	if err != nil {
		log.FatalLog("Invalid cloudletKey", "err", err)
	}
}

func SetNodeKey(hostname *string, nodeType edgeproto.NodeType, cloudletKey *edgeproto.CloudletKey, key *edgeproto.NodeKey) {
	if *hostname == "" {
		*hostname, _ = os.Hostname()
		if *hostname == "" {
			*hostname = "nohostname"
		}
	}
	key.Name = *hostname
	key.NodeType = nodeType
	key.CloudletKey = *cloudletKey
}

func IsClusterInstReqd(app *edgeproto.App) bool {
	if app.Deployment == AppDeploymentTypeVM {
		return false
	}
	return true
}

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "nohostname"
	}
	return hostname
}
