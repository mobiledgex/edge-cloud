package cloudcommon

import (
	"encoding/json"
	"fmt"
	"os"
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

//reserved developer types
var OperatorDeveloper = "developer"

var Developerplatos = "platos"
var DeveloperMobiledgeX = "MobiledgeX"

// platform apps
var PlatosEnablingLayer = "PlatosEnablingLayer"

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
var DeveloperMetricsDbName = "clusterstats"
var MEXPrometheusAppName = "MEXPrometheusAppName"

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

// PlatformApps is the set of all special "platform" developers.   Key
// is DeveloperName:AppName.  Currently only platos's Enabling layer is included.
var platformApps = map[string]bool{
	Developerplatos + ":" + PlatosEnablingLayer: true,
}

// IsPlatformApp true if the developer/app combo is a platform app
func IsPlatformApp(devname string, appname string) bool {
	_, ok := platformApps[devname+":"+appname]
	return ok
}

var AllocatedIpDynamic = "dynamic"

var RootLBL7Port int32 = 443

// OperatorDeveloper is a special value used by the public cloud based cloudlet
var operatorDeveloper = edgeproto.OperatorKey{Name: OperatorDeveloper}

//DefaultCloudletKey is a special value for the public cloud based default cloudlet for each app
// is for an appinst deployment maintained by the developer, not Mobiledgex
var DefaultCloudletKey = edgeproto.CloudletKey{OperatorKey: operatorDeveloper, Name: "default"}

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

// GetAppFQDN gets the app-specific Load Balancer's Fully Qualified Domain Name
// for apps using "dedicated" IP access.
func GetAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey) string {
	lb := GetDedicatedLBFQDN(cloudletKey, clusterKey)
	app := util.DNSSanitize(key.AppKey.Name)
	dev := util.DNSSanitize(key.AppKey.DeveloperKey.Name)
	ver := util.DNSSanitize(key.AppKey.Version)
	return fmt.Sprintf("%s%s%s.%s", dev, app, ver, lb)
}

// GetVMAppFQDN gets the app-specific Fully Qualified Domain Name
// for VM based apps
func GetVMAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey) string {
	lb := GetRootLBFQDN(cloudletKey)
	app := util.DNSSanitize(key.AppKey.Name)
	dev := util.DNSSanitize(key.AppKey.DeveloperKey.Name)
	ver := util.DNSSanitize(key.AppKey.Version)
	return fmt.Sprintf("%s%s%s.%s", dev, app, ver, lb)
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

	err = mykey.Validate()
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
