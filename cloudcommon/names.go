package cloudcommon

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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

const DefaultClust string = "defaultclust"
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
var VMTypePlatformClusterPrimaryNode = "platform-cluster-primary-node"
var VMTypePlatformClusterSecondaryNode = "platform-cluster-secondary-node"

var VMTypeClusterMaster = "cluster-master"
var VMTypeClusterK8sNode = "cluster-k8s-node"
var VMTypeClusterDockerNode = "cluster-docker-node"

// cloudlet node names
var CloudletNodeSharedRootLB = "sharedrootlb"
var CloudletNodeDedicatedRootLB = "dedicatedrootlb"

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
var ProxyMetricsListenUDS = "MetricsUDS" // Unix Domain Socket

var AutoProvMeasurement = "auto-prov-counts"

// AppLabels for the application containers
var MexAppNameLabel = "mexAppName"
var MexAppVersionLabel = "mexAppVersion"
var MexMetricEndpoint = "mexMetricsEndpoint"

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
const EnvoyImageDigest = "sha256:2b07bb8dd35c2a4bb273652b62e85b0bd27d12da94fa11061a9c365d4352e7f9"

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

var RootLBHostname = "shared"

// Fully Qualified Domain Names (FQDNs) primarily come in the
// the following format of 4 "labels" (where domain can actually
// be more than one label itself, i.e. mobiledgex.net):
// cloudletobject.cloudlet.region.domain
// In some cases, another label will be prepended
// (such as for ip-per-k8s-services, the service name is prepended).
// To help avoid the total length limit of 253 when prepending additional
// labels, we restrict the base labels to less than the DNS spec
// per-label restriction of 63, based on how long we expect those
// labels to be in general. For example, we expect most region names to
// be 3-4 characters, while appname+version+org is likely to be much
// longer.
const DnsDomainLabelMaxLen = 40
const DnsRegionLabelMaxLen = 10
const DnsCloudletLabelMaxLen = 50
const DnsCloudletObjectLabelMaxLen = 63

// Values for QOS Priority Session API
const TagPrioritySessionId string = "priority_session_id"
const TagQosProfileName string = "qos_profile_name"
const TagIpUserEquipment string = "ip_user_equipment"

// Wildcard cert for all LBs both shared and dedicated
func GetRootLBFQDNWildcard(cloudlet *edgeproto.Cloudlet) string {
	names := strings.Split(cloudlet.RootLbFqdn, ".")
	names[0] = "*"
	return strings.Join(names, ".")
}

// Old version of getting the shared root lb, does not match wildcard cert.
func GetRootLBFQDNOld(key *edgeproto.CloudletKey, domain string) string {
	loc := util.DNSSanitize(key.Name)
	oper := util.DNSSanitize(key.Organization)
	return fmt.Sprintf("%s.%s.%s", loc, oper, domain)
}

// GetAppInstId returns a string for this AppInst that is likely to be
// unique within the region. It does not guarantee uniqueness.
// The delimiter '.' is removed from the AppInstId so that it can be used
// to append further strings to this ID to build derived unique names.
// Salt can be used by the caller to add an extra field if needed
// to ensure uniqueness. In all cases, any requirements for uniqueness
// must be guaranteed by the caller.
func GetAppInstId(appInst *edgeproto.AppInst, app *edgeproto.App, salt string) string {
	fields := []string{}

	appName := util.DNSSanitize(appInst.Key.AppKey.Name)
	dev := util.DNSSanitize(appInst.Key.AppKey.Organization)
	ver := util.DNSSanitize(appInst.Key.AppKey.Version)
	appId := fmt.Sprintf("%s%s%s", dev, appName, ver)
	fields = append(fields, appId)

	if IsClusterInstReqd(app) {
		cluster := util.DNSSanitize(appInst.Key.ClusterInstKey.ClusterKey.Name)
		fields = append(fields, cluster)
	}

	loc := util.DNSSanitize(appInst.Key.ClusterInstKey.CloudletKey.Name)
	fields = append(fields, loc)

	oper := util.DNSSanitize(appInst.Key.ClusterInstKey.CloudletKey.Organization)
	fields = append(fields, oper)

	if salt != "" {
		salt = util.DNSSanitize(salt)
		fields = append(fields, salt)
	}
	return strings.Join(fields, "-")
}

// FqdnPrefix is used only for IP-per-service platforms that allocate
// an IP for each kubernetes service. Because it adds an extra level of
// DNS label hierarchy and cannot match the wildcard cert, we do not
// support TLS for it.
func FqdnPrefix(svcName string) string {
	return svcName + "."
}

func ServiceFQDN(svcName, baseFQDN string) string {
	return fmt.Sprintf("%s%s", FqdnPrefix(svcName), baseFQDN)
}

// DNS names must have labels <= 63 chars, and the total length
// <= 255 octets (which works out to 253 chars).
func CheckFQDNLengths(prefix, uri string) error {
	fqdn := prefix + uri
	if len(fqdn) > 253 {
		return fmt.Errorf("DNS name %q exceeds 253 chars, please shorten some names", fqdn)
	}
	for _, label := range strings.Split(fqdn, ".") {
		if len(label) > 63 {
			return fmt.Errorf("Label %q of DNS name %q exceeds 63 chars, please shorten it", label, fqdn)
		}
	}
	return nil
}

// For the DME and CRM that require a cloudlet key to be specified
// at startup, this function parses the string argument.
func ParseMyCloudletKey(standalone bool, keystr *string, mykey *edgeproto.CloudletKey) {
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

func GetSideCarAppFilter() *edgeproto.App {
	return &edgeproto.App{
		Key:    edgeproto.AppKey{Organization: OrganizationMobiledgeX},
		DelOpt: edgeproto.DeleteType_AUTO_DELETE,
	}
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
		app.AccessType != edgeproto.AccessType_ACCESS_TYPE_DIRECT {
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
		key.ClusterInstKey.ClusterKey.Name = DefaultClust
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

func GetGPUDriverLicenseCloudletStoragePath(key *edgeproto.GPUDriverKey, cloudletName string) string {
	return fmt.Sprintf("%s/%s/%s", GetGPUDriverStoragePath(key), cloudletName, edgeproto.GPUDriverLicenseConfig)
}

func GetGPUDriverBuildStoragePath(key *edgeproto.GPUDriverKey, buildName, ext string) string {
	return fmt.Sprintf("%s/%s%s", GetGPUDriverStoragePath(key), buildName, ext)
}

func GetGPUDriverURL(key *edgeproto.GPUDriverKey, deploymentTag, buildName, ext string) string {
	return fmt.Sprintf("https://storage.cloud.google.com/%s/%s", GetGPUDriverBucketName(deploymentTag), GetGPUDriverBuildStoragePath(key, buildName, ext))
}

func GetGPUDriverLicenseURL(key *edgeproto.GPUDriverKey, cloudletName, deploymentTag string) string {
	licensePath := GetGPUDriverLicenseStoragePath(key)
	if cloudletName != "" {
		licensePath = GetGPUDriverLicenseCloudletStoragePath(key, cloudletName)
	}
	return fmt.Sprintf("https://storage.cloud.google.com/%s/%s", GetGPUDriverBucketName(deploymentTag), licensePath)
}

func GetGPUDriverBuildPathFromURL(driverURL, deploymentTag string) string {
	return strings.TrimPrefix(driverURL, fmt.Sprintf("https://storage.cloud.google.com/%s/", GetGPUDriverBucketName(deploymentTag)))
}
