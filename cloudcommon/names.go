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

// Cloudlet Platform nodes -- update IsPlatformNode if adding to this list

type NodeType int

const (
	NodeTypeAppVM NodeType = iota
	NodeTypeSharedRootLB
	NodeTypeDedicatedRootLB
	NodeTypePlatformVM
	NodeTypePlatformHost
	NodeTypePlatformK8sClusterMaster
	NodeTypePlatformK8sClusterPrimaryNode
	NodeTypePlatformK8sClusterSecondaryNode
	// Cloudlet Compute nodes
	NodeTypeK8sClusterMaster
	NodeTypeK8sClusterNode
	NodeTypeDockerClusterNode
)

func (n NodeType) String() string {
	switch n {
	case NodeTypeAppVM:
		return "appvm"
	case NodeTypeSharedRootLB:
		return "sharedrootlb"
	case NodeTypeDedicatedRootLB:
		return "dedicatedrootlb"
	case NodeTypePlatformVM:
		return "platformvm"
	case NodeTypePlatformHost:
		return "platformhost"
	case NodeTypePlatformK8sClusterMaster:
		return "platform-k8s-cluster-master"
	case NodeTypePlatformK8sClusterPrimaryNode:
		return "platform-k8s-cluster-primary-node"
	case NodeTypePlatformK8sClusterSecondaryNode:
		return "platform-k8s-cluster-secondary-node"
	case NodeTypeK8sClusterMaster:
		return "k8s-cluster-master"
	case NodeTypeK8sClusterNode:
		return "k8s-cluster-node"
	case NodeTypeDockerClusterNode:
		return "docker-cluster-node"
	}
	return "unknown node type"
}

// resource types
var ResourceTypeK8sLBSvc = "k8s-lb-svc"

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

// GCS Storage Bucket Name: used to store GPU driver packages
func GetGPUDriverBucketName(deploymentTag string) string {
	if deploymentTag == "" {
		deploymentTag = "local"
	}
	return fmt.Sprintf("mobiledgex-%s-gpu-drivers", deploymentTag)
}

func GetGPUDriverStoragePath(key *edgeproto.GPUDriverKey, region string) (string, error) {
	orgName := key.Organization
	if key.Organization == "" {
		orgName = OrganizationMobiledgeX
	}
	sPath := StoragePath{}
	err := sPath.AppendPaths(region, orgName, key.Name)
	if err != nil {
		return "", err
	}
	return sPath.String(), nil
}

func GetGPUDriverLicenseStoragePath(key *edgeproto.GPUDriverKey, region string) (string, error) {
	driverStoragePath, err := GetGPUDriverStoragePath(key, region)
	if err != nil {
		return "", err
	}
	sPath := StoragePath{}
	err = sPath.AppendPaths("licenseconfig", edgeproto.GPUDriverLicenseConfig)
	if err != nil {
		return "", err
	}
	return driverStoragePath + "/" + sPath.String(), nil
}

func GetGPUDriverLicenseCloudletStoragePath(key *edgeproto.GPUDriverKey, region string, cloudletKey *edgeproto.CloudletKey) (string, error) {
	if key.Organization != "" && key.Organization != cloudletKey.Organization {
		return "", fmt.Errorf("Can only use %s or '' org gpu drivers", key.Organization)
	}
	driverStoragePath, err := GetGPUDriverStoragePath(key, region)
	if err != nil {
		return "", err
	}
	// If GPU driver org is empty i.e. it is owned by MobiledgeX org, then add cloudletOrg to storage path
	cloudletOrg := ""
	if key.Organization == "" {
		cloudletOrg = cloudletKey.Organization
	}
	sPath := StoragePath{}
	err = sPath.AppendPaths("cloudlet", "licenseconfig", cloudletOrg, cloudletKey.Name, edgeproto.GPUDriverLicenseConfig)
	if err != nil {
		return "", err
	}
	return driverStoragePath + "/" + sPath.String(), nil
}

func GetGPUDriverBuildStoragePath(key *edgeproto.GPUDriverKey, region, buildName, ext string) (string, error) {
	driverStoragePath, err := GetGPUDriverStoragePath(key, region)
	if err != nil {
		return "", err
	}
	sPath := StoragePath{}
	err = sPath.AppendPaths("build", buildName, ext)
	if err != nil {
		return "", err
	}
	return driverStoragePath + "/" + sPath.String(), nil
}

func GetGPUDriverURL(bucketName, buildStoragePath string) string {
	return fmt.Sprintf("https://storage.cloud.google.com/%s/%s", bucketName, buildStoragePath)
}

func GetGPUDriverLicenseURL(bucketName, licenseConfigStoragePath string) string {
	return fmt.Sprintf("https://storage.cloud.google.com/%s/%s", bucketName, licenseConfigStoragePath)
}

func IsPlatformNode(nodeTypeStr string) bool {
	switch nodeTypeStr {
	case NodeTypePlatformVM.String():
		fallthrough
	case NodeTypePlatformHost.String():
		fallthrough
	case NodeTypePlatformK8sClusterMaster.String():
		fallthrough
	case NodeTypePlatformK8sClusterPrimaryNode.String():
		fallthrough
	case NodeTypePlatformK8sClusterSecondaryNode.String():
		return true
	}
	return false
}

func IsLBNode(nodeTypeStr string) bool {
	return nodeTypeStr == NodeTypeDedicatedRootLB.String() || nodeTypeStr == NodeTypeSharedRootLB.String()
}
