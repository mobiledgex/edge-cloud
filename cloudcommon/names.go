package cloudcommon

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	yaml "gopkg.in/yaml.v2"
)

// special operator types
var OperatorGCP = "gcp"
var OperatorAzure = "azure"
var OperatorAWS = "aws"

var Organizationplatos = "platos"
var OrganizationMobiledgeX = "MobiledgeX"
var OrganizationEdgeBox = "EdgeBox"

const DefaultCluster string = "DefaultCluster"
const DefaultMultiTenantCluster string = "defaultmtclust"

// platform apps
var PlatosEnablingLayer = "PlatosEnablingLayer"

// cloudlet types
var CloudletKindOpenStack = "openstack"
var CloudletKindAzure = "azure"
var CloudletKindAws = "aws"
var CloudletKindGCP = "gcp"
var CloudletKindDIND = "dind"
var CloudletKindFake = "fake"

var OperatingSystemMac = "mac"
var OperatingSystemLinux = "linux"

// cloudlet vm types
var VMTypeAppVM = "appvm"
var VMTypeRootLB = "rootlb"
var VMTypePlatform = "platform"
var VMTypePlatformClusterMaster = "platform-cluster-master"
var VMTypePlatformClusterNode = "platform-cluster-node"
var VMTypeClusterMaster = "cluster-master"
var VMTypeClusterK8sNode = "cluster-k8s-node"
var VMTypeClusterDockerNode = "cluster-docker-node"

const AutoClusterPrefix = "autocluster"
const ReservableClusterPrefix = "reservable"
const ReserveClusterEvent = "Reserve ClusterInst"
const FreeClusterEvent = "Free ClusterInst reservation"

// network schemes for use by standalone deployments (e.g. DIND)
var NetworkSchemePublicIP = "publicip"
var NetworkSchemePrivateIP = "privateip"

// Metrics common variables - TODO move to edge-cloud-infra after metrics-exporter chagnes
var DeveloperMetricsDbName = "metrics"
var MEXPrometheusAppName = "MEXPrometheusAppName"
var PrometheusPort = int32(9090)
var NFSAutoProvisionAppName = "NFSAutoProvision"
var ProxyMetricsPort = int32(65121)
var ProxyMetricsDefaultListenIP = "127.0.0.1"
var AutoProvMeasurement = "auto-prov-counts"

// AppLabels for the application containers
var MexAppNameLabel = "mexAppName"
var MexAppVersionLabel = "mexAppVersion"

// Instance Lifecycle variables
var EventsDbName = "events"
var CloudletEvent = "cloudlet"
var ClusterInstEvent = "clusterinst"
var ClusterInstCheckpoints = "clusterinst-checkpoints"
var AppInstEvent = "appinst"
var AppInstCheckpoints = "appinst-checkpoints"
var MonthlyInterval = "MONTH"

// Cloudlet resource usage
var CloudletResourceUsageDbName = "cloudlet_resource_usage"
var CloudletFlavorUsageMeasurement = "cloudlet-flavor-usage"

// EdgeEvents Metrics Influx variables
var EdgeEventsMetricsDbName = "edgeevents_metrics"
var LatencyMetric = "latency-metric"
var DeviceMetric = "device-metric"
var CustomMetric = "custom-metric"

// Map used to identify which metrics should go to persistent_metrics db. Value represents the measurement creation status
var EdgeEventsMetrics = map[string]struct{}{
	LatencyMetric: struct{}{},
	DeviceMetric:  struct{}{},
	CustomMetric:  struct{}{},
}

var DownsampledMetricsDbName = "downsampled_metrics"

var IPAddrAllInterfaces = "0.0.0.0"
var IPAddrLocalHost = "127.0.0.1"
var RemoteServerNone = ""

// Client type to access cluster nodes
var ClientTypeRootLB string = "rootlb"
var ClientTypeClusterVM string = "clustervm"

type InstanceEvent string

const (
	CREATED           InstanceEvent = "CREATED"
	UPDATE_START      InstanceEvent = "UPDATE_START"
	UPDATE_ERROR      InstanceEvent = "UPDATE_ERROR"
	UPDATE_COMPLETE   InstanceEvent = "UPDATE_COMPLETE"
	DELETED           InstanceEvent = "DELETED"
	DELETE_ERROR      InstanceEvent = "DELETE_ERROR"
	HEALTH_CHECK_FAIL InstanceEvent = "HEALTH_CHECK_FAIL"
	HEALTH_CHECK_OK   InstanceEvent = "HEALTH_CHECK_OK"
	RESERVED          InstanceEvent = "RESERVED"
	UNRESERVED        InstanceEvent = "UNRESERVED"
)

var InstanceUp = "UP"
var InstanceDown = "DOWN"

// DIND script to pull from kubeadm-dind-cluster
var DindScriptName = "dind-cluster-v1.14.sh"

var MexNodePrefix = "mex-k8s-node-"

// GCP limits to 40, Azure has issues above 54.  For consistency go with the lower limit
const MaxClusterNameLength = 40

// Common cert name. Cannot use common name as filename since envoy doesn't know if the app is dedicated or not
const CertName = "envoyTlsCerts"
const EnvoyImageDigest = "sha256:9bc06553ad6add6bfef1d8a1b04f09721415975e2507da0a2d5b914c066474df"

// CloudletInfo properties
const (
	// Cloudlet supports multi-tenant k8s cluster
	CloudletSupportsMT = "supports-mt"
)

// PlatformApps is the set of all special "platform" developers.   Key
// is DeveloperName:AppName.  Currently only platos's Enabling layer is included.
var platformApps = map[string]bool{
	Organizationplatos + ":" + PlatosEnablingLayer: true,
}

// IsPlatformApp true if the developer/app combo is a platform app
func IsPlatformApp(devname string, appname string) bool {
	_, ok := platformApps[devname+":"+appname]
	return ok
}

var AllocatedIpDynamic = "dynamic"

// GetRootLBFQDN gets the global Load Balancer's Fully Qualified Domain Name
// for apps using "shared" IP access.
func GetRootLBFQDN(key *edgeproto.CloudletKey, domain string) string {
	loc := util.DNSSanitize(key.Name)
	oper := util.DNSSanitize(key.Organization)
	return fmt.Sprintf("%s.%s.%s", loc, oper, domain)
}

// GetDedicatedLBFQDN gets the cluster-specific Load Balancer's Fully Qualified Domain Name
// for clusters using "dedicated" IP access.
func GetDedicatedLBFQDN(cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey, domain string) string {
	clust := util.DNSSanitize(clusterKey.Name)
	loc := util.DNSSanitize(cloudletKey.Name)
	oper := util.DNSSanitize(cloudletKey.Organization)
	return fmt.Sprintf("%s.%s.%s.%s", clust, loc, oper, domain)
}

// Get Fully Qualified Name for the App i.e. with developer & version info
func GetAppFQN(key *edgeproto.AppKey) string {
	app := util.DNSSanitize(key.Name)
	dev := util.DNSSanitize(key.Organization)
	ver := util.DNSSanitize(key.Version)
	return fmt.Sprintf("%s%s%s", dev, app, ver)
}

// GetAppFQDN gets the app-specific Load Balancer's Fully Qualified Domain Name
// for apps using "dedicated" IP access.
func GetAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey, domain string) string {
	lb := GetDedicatedLBFQDN(cloudletKey, clusterKey, domain)
	appFQN := GetAppFQN(&key.AppKey)
	return fmt.Sprintf("%s.%s", appFQN, lb)
}

// GetVMAppFQDN gets the app-specific Fully Qualified Domain Name
// for VM based apps
func GetVMAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey, domain string) string {
	lb := GetRootLBFQDN(cloudletKey, domain)
	appFQN := GetAppFQN(&key.AppKey)
	return fmt.Sprintf("%s.%s", appFQN, lb)
}

func FqdnPrefix(svcName string) string {
	return svcName + "."
}

func ServiceFQDN(svcName, baseFQDN string) string {
	return fmt.Sprintf("%s%s", FqdnPrefix(svcName), baseFQDN)
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
		log.FatalLog("Invalid cloudletKey", "key", mykey, "err", err)
	}
}

func IsClusterInstReqd(app *edgeproto.App) bool {
	if app.Deployment == DeploymentTypeVM {
		return false
	}
	return true
}

func IsSideCarApp(app *edgeproto.App) bool {
	if app.Key.Organization == OrganizationMobiledgeX && app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
		return true
	}
	return false
}

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "nohostname"
	}
	return hostname
}

func GetAppClientType(app *edgeproto.App) string {
	clientType := ClientTypeRootLB
	if app.Deployment == DeploymentTypeDocker &&
		app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
		// docker commands can be run on either the rootlb or on the docker
		// vm. The default is to run on the rootlb client
		// If using a load balancer access, a separate VM is always used for
		// docker vs the LB, and we always use host networking mode
		clientType = ClientTypeClusterVM
	}
	return clientType
}

// GetCertsDirAndFiles returns certsDir, certFile, keyFile
func GetCertsDirAndFiles(pwd string) (string, string, string) {
	pwd = strings.TrimSpace(pwd)
	certsDir := pwd + "/envoy/certs"
	certFile := certsDir + "/" + CertName + ".crt"
	keyFile := certsDir + "/" + CertName + ".key"
	return certsDir, certFile, keyFile
}

func GetCloudletResourceUsageMeasurement(pfType string) string {
	return fmt.Sprintf("%s-resource-usage", pfType)
}

// Because of the streamAppInst feature, we cannot change the input
// key based on the App definition. This is because streamAppInst objects
// may exist even if the App definition does not. This is true in two cases,
// 1) If the AppInst/App was deleted, stream object still remains,
// 2) CreateAppInst was called on a missing App, then stream object is created
// with the missing App key, and exists without App/AppInst present.
//
// Consider these use cases where user supplies an AppInst with no ClusterName,
// where the App is a VM app, and we attempt to set the ClusterName based on the
// App type.
// 1) Create App, AppInst, set ClusterName to "VM". Stream object has "VM" set.
// Now delete AppInst, App. Stream object still exists with "VM" set.
// 2) Create same AppInst with missing App (fails). Cannot determine App type,
// so Stream object remains with cluster name blank.
// Now we want to search for the Stream object in both cases above. Neither the
// App nor the AppInst exist. The search supplies the same AppInst key with
// Cluster name blank. The search code cannot determine if the App was a VM app
// or not. If it leaves the cluster name blank, then a lookup in case (1) fails.
// If it assumes the App was a VM App and fills in the cluster name, then the
// lookup in case (2) fails.
// Because of the above, we always set defaults for missing key fields.
// Whether these defaults are valid can be checked later by the code
// (and possibly generate an error), but the important part is that what ends
// up as the key is the database is not dependent on other database objects
// that may or may not exist at the time.
//
// In general we need to be very careful about changing the key that is passed
// in from the user. The key is what the user uses to find the object.
// The user should always be able to pass in the same key to create/find/delete.
// If we modify the key, it must be done in a way that is consistent in ALL
// cases, regardless of the state of the database or which API they call or
// anything else that may change over time. Otherwise, the result is the user
// may not be able to find/delete their object later.
func SetAppInstKeyDefaults(key *edgeproto.AppInstKey) (bool, bool) {
	setClusterOrg := false
	setClusterName := false
	if key.ClusterInstKey.Organization == "" {
		key.ClusterInstKey.Organization = key.AppKey.Organization
		setClusterOrg = true
	}
	if key.ClusterInstKey.ClusterKey.Name == "" {
		key.ClusterInstKey.ClusterKey.Name = DefaultCluster
		setClusterName = true
	}
	return setClusterOrg, setClusterName
}

// GCS Storage Bucket Name: used to store GPU driver packages
func GetGPUDriverBucketName(deploymentTag string) string {
	return fmt.Sprintf("mobiledgex-%s-gpu-drivers", deploymentTag)
}

func GetGPUDriverStoragePath(key *edgeproto.GPUDriverKey) string {
	orgName := key.Organization
	if key.Organization == "" {
		orgName = OrganizationMobiledgeX
	}
	return fmt.Sprintf("%s/%s", orgName, key.Name)
}

func GetGPUDriverLicenseStoragePath(key *edgeproto.GPUDriverKey) string {
	return fmt.Sprintf("%s/%s", GetGPUDriverStoragePath(key), edgeproto.GPUDriverLicenseConfig)
}

func GetGPUDriverBuildStoragePath(key *edgeproto.GPUDriverKey, buildName, ext string) string {
	return fmt.Sprintf("%s/%s%s", GetGPUDriverStoragePath(key), buildName, ext)
}

func GetGPUDriverURL(key *edgeproto.GPUDriverKey, deploymentTag, buildName, ext string) string {
	return fmt.Sprintf("https://storage.cloud.google.com/%s/%s", GetGPUDriverBucketName(deploymentTag), GetGPUDriverBuildStoragePath(key, buildName, ext))
}

func GetGPUDriverLicenseURL(key *edgeproto.GPUDriverKey, deploymentTag string) string {
	return fmt.Sprintf("https://storage.cloud.google.com/%s/%s", GetGPUDriverBucketName(deploymentTag), GetGPUDriverLicenseStoragePath(key))
}

func GetGPUDriverBuildPathFromURL(driverURL, deploymentTag string) string {
	return strings.TrimPrefix(driverURL, fmt.Sprintf("https://storage.cloud.google.com/%s/", GetGPUDriverBucketName(deploymentTag)))
}
