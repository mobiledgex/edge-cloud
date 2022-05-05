// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package platform

import (
	"context"
	"strings"
	"sync"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/chefmgmt"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/redundancy"
	"github.com/edgexr/edge-cloud/cloudcommon"
	"github.com/edgexr/edge-cloud/cloudcommon/node"
	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type PlatformConfig struct {
	CloudletKey         *edgeproto.CloudletKey
	PhysicalName        string
	Region              string
	TestMode            bool
	CloudletVMImagePath string
	VMImageVersion      string
	PackageVersion      string
	EnvVars             map[string]string
	NodeMgr             *node.NodeMgr
	AppDNSRoot          string
	RootLBFQDN          string
	ChefServerPath      string
	DeploymentTag       string
	Upgrade             bool
	AccessApi           AccessApi
	TrustPolicy         string
	CacheDir            string
	GPUConfig           *edgeproto.GPUConfig
}

type Caches struct {
	SettingsCache             *edgeproto.SettingsCache
	FlavorCache               *edgeproto.FlavorCache
	TrustPolicyCache          *edgeproto.TrustPolicyCache
	TrustPolicyExceptionCache *edgeproto.TrustPolicyExceptionCache
	CloudletPoolCache         *edgeproto.CloudletPoolCache
	ClusterInstCache          *edgeproto.ClusterInstCache
	ClusterInstInfoCache      *edgeproto.ClusterInstInfoCache
	AppInstCache              *edgeproto.AppInstCache
	AppInstInfoCache          *edgeproto.AppInstInfoCache
	AppCache                  *edgeproto.AppCache
	ResTagTableCache          *edgeproto.ResTagTableCache
	CloudletCache             *edgeproto.CloudletCache
	CloudletInternalCache     *edgeproto.CloudletInternalCache
	VMPoolCache               *edgeproto.VMPoolCache
	VMPoolInfoCache           *edgeproto.VMPoolInfoCache
	GPUDriverCache            *edgeproto.GPUDriverCache
	NetworkCache              *edgeproto.NetworkCache
	CloudletInfoCache         *edgeproto.CloudletInfoCache
	// VMPool object managed by CRM
	VMPool    *edgeproto.VMPool
	VMPoolMux *sync.Mutex
}

// Features that the platform supports or enables
type Features struct {
	SupportsMultiTenantCluster               bool
	SupportsSharedVolume                     bool
	SupportsTrustPolicy                      bool
	SupportsKubernetesOnly                   bool // does not support docker/VM
	KubernetesRequiresWorkerNodes            bool // k8s cluster cannot be master only
	CloudletServicesLocal                    bool // cloudlet services running locally to controller
	IPAllocatedPerService                    bool // Every k8s service gets a public IP (GCP/etc)
	SupportsImageTypeOVF                     bool // Supports OVF images for VM deployments
	IsVMPool                                 bool // cloudlet is just a pool of pre-existing VMs
	IsFake                                   bool // Just for unit-testing/e2e-testing
	SupportsAdditionalNetworks               bool // Additional networks can be added
	IsSingleKubernetesCluster                bool // Entire platform is just a single K8S cluster
	SupportsAppInstDedicatedIP               bool // Supports per AppInst dedicated IPs
	SupportsPlatformHighAvailabilityOnK8s    bool // Supports High Availablity with 2 CRMs on K8s
	SupportsPlatformHighAvailabilityOnDocker bool // Supports HA on docker

	NoKubernetesClusterAutoScale bool // No support for k8s cluster auto-scale
}

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetVersionProperties returns properties related to the platform version
	GetVersionProperties() map[string]string
	// Get platform features
	GetFeatures() *Features
	// InitCommon is called once during CRM startup to do steps needed for both active or standby. If the platform does not support
	// H/A and does not need separate steps for the active unit, then just this func can be implemented and InitHAConditional can be left empty
	InitCommon(ctx context.Context, platformConfig *PlatformConfig, caches *Caches, haMgr *redundancy.HighAvailabilityManager, updateCallback edgeproto.CacheUpdateCallback) error
	// InitHAConditional is only needed for platforms which support H/A. It is called in the following cases: 1) when platform initially starts in a non-switchover case
	// 2) in a switchover case if the previouly-active unit is running a different version as specified by GetInitHAConditionalCompatibilityVersion
	InitHAConditional(ctx context.Context, platformConfig *PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error
	// GetInitializationCompatibilityVersion returns a version as a string. When doing switchovers, if the new version matches the previous version, then InitHAConditional
	// is not called again. If there is a mismatch, then InitHAConditional will be called again.
	GetInitHAConditionalCompatibilityVersion(ctx context.Context) string
	// Gather information about the cloudlet platform.
	// This includes available resources, flavors, etc.
	// Returns true if sync with controller is required
	GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error
	// Create a Kubernetes Cluster on the cloudlet.
	CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error
	// Delete a Kuberentes Cluster on the cloudlet.
	DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Update the cluster
	UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Get resources used by the cloudlet
	GetCloudletInfraResources(ctx context.Context) (*edgeproto.InfraResourcesSnapshot, error)
	// Get cluster additional resources used by the vms specific to the platform
	GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource
	// Get Cloudlet Resource Properties
	GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error)
	// Get cluster additional resource metric
	GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error
	// Get resources used by the cluster
	GetClusterInfraResources(ctx context.Context, clusterKey *edgeproto.ClusterInstKey) (*edgeproto.InfraResources, error)
	// Create an AppInst on a Cluster
	CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Update an AppInst
	UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Get AppInst runtime information
	GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error)
	// Get the client to manage the ClusterInst
	GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error)
	// Get the client to manage the specified platform management node
	GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode, ops ...pc.SSHClientOp) (ssh.Client, error)
	// List the cloudlet management nodes used by this platform
	ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst, vmAppInsts []edgeproto.AppInst) ([]edgeproto.CloudletMgmtNode, error)
	// Get the command to pass to PlatformClient for the container command
	GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error)
	// Get the console URL of the VM appInst
	GetConsoleUrl(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst) (string, error)
	// Set power state of the AppInst
	SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Create Cloudlet returns cloudletResourcesCreated, error
	CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *Caches, accessApi AccessApi, updateCallback edgeproto.CacheUpdateCallback) (bool, error)
	UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet
	DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *Caches, accessApi AccessApi, updateCallback edgeproto.CacheUpdateCallback) error
	// Save Cloudlet AccessVars
	SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet AccessVars
	DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error
	// Performs Upgrades for things like k8s config
	PerformUpgrades(ctx context.Context, caches *Caches, cloudletState dme.CloudletState) error
	// Get Cloudlet Manifest Config
	GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi AccessApi, flavor *edgeproto.Flavor, caches *Caches) (*edgeproto.CloudletManifest, error)
	// Verify VM
	VerifyVMs(ctx context.Context, vms []edgeproto.VM) error
	// Get Cloudlet Properties
	GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error)
	// Platform-sepcific access data lookup (only called from Controller context)
	GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error)
	// Update the cloudlet's Trust Policy
	UpdateTrustPolicy(ctx context.Context, TrustPolicy *edgeproto.TrustPolicy) error
	//  Create and Update TrustPolicyException
	UpdateTrustPolicyException(ctx context.Context, TrustPolicyException *edgeproto.TrustPolicyException, clusterInstKey *edgeproto.ClusterInstKey) error
	// Delete TrustPolicyException
	DeleteTrustPolicyException(ctx context.Context, TrustPolicyExceptionKey *edgeproto.TrustPolicyExceptionKey, clusterInstKey *edgeproto.ClusterInstKey) error
	// Get restricted cloudlet create status
	GetRestrictedCloudletStatus(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi AccessApi, updateCallback edgeproto.CacheUpdateCallback) error
	// Get ssh clients of all root LBs
	GetRootLBClients(ctx context.Context) (map[string]ssh.Client, error)
	// Get RootLB Flavor
	GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error)
	// Called when the platform instance switches activity. Currently only transition from Standby to Active is allowed.
	ActiveChanged(ctx context.Context, platformActive bool) error
	// Sanitizes the name to make it conform to platform requirements.
	NameSanitize(name string) string
}

type ClusterSvc interface {
	// GetVersionProperties returns properties related to the platform version
	GetVersionProperties() map[string]string
	// Get AppInst Configs
	GetAppInstConfigs(ctx context.Context, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst,
		autoScalePolicy *edgeproto.AutoScalePolicy, settings *edgeproto.Settings,
		userAlerts []edgeproto.AlertPolicy) ([]*edgeproto.ConfigFile, error)
}

// AccessApi handles functions that require secrets access, but
// may be run from either the Controller or CRM context, so may either
// use Vault directly (Controller) or may go indirectly via Controller (CRM).
type AccessApi interface {
	cloudcommon.RegistryAuthApi
	cloudcommon.GetPublicCertApi
	GetCloudletAccessVars(ctx context.Context) (map[string]string, error)
	SignSSHKey(ctx context.Context, publicKey string) (string, error)
	GetSSHPublicKey(ctx context.Context) (string, error)
	GetOldSSHKey(ctx context.Context) (*vault.MEXKey, error)
	GetChefAuthKey(ctx context.Context) (*chefmgmt.ChefAuthKey, error)
	CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error
	GetDNSRecords(ctx context.Context, zone, fqdn string) ([]cloudflare.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, zone, recordID string) error
	GetSessionTokens(ctx context.Context, arg []byte) (map[string]string, error)
	GetKafkaCreds(ctx context.Context) (*node.KafkaCreds, error)
	GetGCSCreds(ctx context.Context) ([]byte, error)
	GetFederationAPIKey(ctx context.Context, fedName string) (string, error)
}

var pfMaps = map[string]string{
	"fakeinfra": "fake",
	"edgebox":   "dind",
	"kindinfra": "kind",
}

func GetType(pfType string) string {
	out := strings.TrimPrefix(pfType, "PLATFORM_TYPE_")
	out = strings.ToLower(out)
	out = strings.Replace(out, "_", "", -1)
	if mapOut, found := pfMaps[out]; found {
		return mapOut
	}
	return out
}

// Track K8s AppInstances for resource management only if platform supports K8s deployments only
func TrackK8sAppInst(ctx context.Context, app *edgeproto.App, features *Features) bool {
	if features.SupportsKubernetesOnly &&
		(app.Deployment == cloudcommon.DeploymentTypeKubernetes ||
			app.Deployment == cloudcommon.DeploymentTypeHelm) {
		return true
	}
	return false
}
