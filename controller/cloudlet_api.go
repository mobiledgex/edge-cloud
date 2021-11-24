package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/accessapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
)

type CloudletApi struct {
	all             *AllApis
	sync            *Sync
	store           edgeproto.CloudletStore
	cache           *edgeproto.CloudletCache
	accessKeyServer *node.AccessKeyServer
}

// Vault roles for all services
type VaultRoles struct {
	DmeRoleID    string `json:"dmeroleid"`
	DmeSecretID  string `json:"dmesecretid"`
	CRMRoleID    string `json:"crmroleid"`
	CRMSecretID  string `json:"crmsecretid"`
	CtrlRoleID   string `json:"controllerroleid"`
	CtrlSecretID string `json:"controllersecretid"`
}

var (
	DefaultPlatformFlavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "DefaultPlatformFlavor",
		},
		Vcpus: 2,
		Ram:   4096,
		Disk:  20,
	}
)

// Transition states indicate states in which the CRM is still busy.
var CreateCloudletTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_CREATING: struct{}{},
}
var UpdateCloudletTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_UPDATING: struct{}{},
}

const (
	PlatformInitTimeout           = 20 * time.Minute
	DefaultResourceAlertThreshold = 80 // percentage
)

type updateCloudletCallback struct {
	in       *edgeproto.Cloudlet
	callback edgeproto.CloudletApi_CreateCloudletServer
}

func (s *updateCloudletCallback) cb(updateType edgeproto.CacheUpdateType, value string) {
	ctx := s.callback.Context()
	switch updateType {
	case edgeproto.UpdateTask:
		log.SpanLog(ctx, log.DebugLevelApi, "SetStatusTask", "key", s.in.Key, "taskName", value)
		s.in.Status.SetTask(value)
		s.callback.Send(&edgeproto.Result{Message: s.in.Status.ToString()})
	case edgeproto.UpdateStep:
		log.SpanLog(ctx, log.DebugLevelApi, "SetStatusStep", "key", s.in.Key, "stepName", value)
		s.in.Status.SetStep(value)
		s.callback.Send(&edgeproto.Result{Message: s.in.Status.ToString()})
	}
}

func ignoreCRMState(cctx *CallContext) bool {
	if cctx.Override == edgeproto.CRMOverride_IGNORE_CRM ||
		cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE {
		return true
	}
	return false
}

func NewCloudletApi(sync *Sync, all *AllApis) *CloudletApi {
	cloudletApi := CloudletApi{}
	cloudletApi.all = all
	cloudletApi.sync = sync
	cloudletApi.store = edgeproto.NewCloudletStore(sync.store)
	cloudletApi.cache = nodeMgr.CloudletLookup.GetCloudletCache(node.NoRegion)
	sync.RegisterCache(cloudletApi.cache)
	cloudletApi.accessKeyServer = node.NewAccessKeyServer(cloudletApi.cache, nodeMgr.VaultAddr)
	return &cloudletApi
}

func (s *CloudletApi) Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletApi) HasKey(key *edgeproto.CloudletKey) bool {
	return s.cache.HasKey(key)
}

func (s *CloudletApi) ReplaceErrorState(ctx context.Context, in *edgeproto.Cloudlet, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}

		if inst.State != edgeproto.TrackedState_CREATE_ERROR &&
			inst.State != edgeproto.TrackedState_DELETE_ERROR &&
			inst.State != edgeproto.TrackedState_UPDATE_ERROR {
			return nil
		}
		if newState == edgeproto.TrackedState_NOT_PRESENT {
			s.store.STMDel(stm, &in.Key)
			s.all.cloudletRefsApi.store.STMDel(stm, &in.Key)
			s.all.clusterInstApi.deleteCloudletSingularCluster(stm, &in.Key, inst.SingleKubernetesClusterOwner)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

func getCrmEnv(vars map[string]string) {
	for _, key := range []string{
		"JAEGER_ENDPOINT",
		"E2ETEST_TLS",
	} {
		if val, ok := os.LookupEnv(key); ok {
			vars[key] = val
		}
	}
}

func (s *CloudletApi) getPlatformConfig(ctx context.Context, cloudlet *edgeproto.Cloudlet) (*edgeproto.PlatformConfig, error) {
	pfConfig := edgeproto.PlatformConfig{}
	pfConfig.PlatformTag = cloudlet.ContainerVersion
	pfConfig.TlsCertFile = nodeMgr.GetInternalTlsCertFile()
	pfConfig.TlsKeyFile = nodeMgr.GetInternalTlsKeyFile()
	pfConfig.TlsCaFile = nodeMgr.GetInternalTlsCAFile()
	pfConfig.UseVaultPki = nodeMgr.InternalPki.UseVaultPki
	pfConfig.ContainerRegistryPath = *cloudletRegistryPath
	pfConfig.CloudletVmImagePath = *cloudletVMImagePath
	pfConfig.TestMode = *testMode
	pfConfig.EnvVar = make(map[string]string)
	for k, v := range cloudlet.EnvVar {
		pfConfig.EnvVar[k] = v
	}
	pfConfig.Region = *region
	pfConfig.CommercialCerts = *commercialCerts
	pfConfig.AppDnsRoot = *appDNSRoot
	getCrmEnv(pfConfig.EnvVar)
	addrObjs := strings.Split(*notifyAddr, ":")
	if len(addrObjs) != 2 {
		return nil, fmt.Errorf("unable to fetch notify addr of the controller")
	}
	accessAddrObjs := strings.Split(*accessApiAddr, ":")
	if len(accessAddrObjs) != 2 {
		return nil, fmt.Errorf("unable to parse accessApi addr of the controller")
	}
	pfConfig.NotifyCtrlAddrs = *publicAddr + ":" + addrObjs[1]
	pfConfig.AccessApiAddr = *publicAddr + ":" + accessAddrObjs[1]
	pfConfig.Span = log.SpanToString(ctx)
	pfConfig.ChefServerPath = *chefServerPath
	pfConfig.ChefClientInterval = s.all.settingsApi.Get().ChefClientInterval
	pfConfig.DeploymentTag = nodeMgr.DeploymentTag

	return &pfConfig, nil
}

func (s *CloudletApi) startCloudletStream(ctx context.Context, key *edgeproto.CloudletKey, inCb edgeproto.CloudletApi_CreateCloudletServer) (*streamSend, edgeproto.CloudletApi_CreateCloudletServer, error) {
	streamKey := edgeproto.GetStreamKeyFromCloudletKey(key)
	streamSendObj, err := s.all.streamObjApi.startStream(ctx, &streamKey, inCb)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to start Cloudlet stream", "err", err)
		return nil, inCb, err
	}
	return streamSendObj, &CbWrapper{
		streamSendObj: streamSendObj,
		GenericCb:     inCb,
	}, nil
}

func (s *CloudletApi) stopCloudletStream(ctx context.Context, key *edgeproto.CloudletKey, streamSendObj *streamSend, objErr error) {
	streamKey := edgeproto.GetStreamKeyFromCloudletKey(key)
	if err := s.all.streamObjApi.stopStream(ctx, &streamKey, streamSendObj, objErr); err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to stop Cloudlet stream", "err", err)
	}
}

func (s *StreamObjApi) StreamCloudlet(key *edgeproto.CloudletKey, cb edgeproto.StreamObjApi_StreamCloudletServer) error {
	ctx := cb.Context()
	cloudlet := &edgeproto.Cloudlet{}
	// if cloudlet is absent, then stream the deletion status messages
	if !s.all.cloudletApi.cache.Get(key, cloudlet) ||
		cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_DIRECT_ACCESS ||
		(cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS && cloudlet.State != edgeproto.TrackedState_READY) {
		// If restricted scenario, then stream msgs only if either cloudlet obj was not created successfully or it is updating
		return s.StreamMsgs(&edgeproto.AppInstKey{ClusterInstKey: edgeproto.VirtualClusterInstKey{CloudletKey: *key}}, cb)
	}
	cloudletInfo := edgeproto.CloudletInfo{}
	if s.all.cloudletInfoApi.cache.Get(key, &cloudletInfo) {
		if cloudletInfo.State == dme.CloudletState_CLOUDLET_STATE_READY ||
			cloudletInfo.State == dme.CloudletState_CLOUDLET_STATE_ERRORS ||
			cloudletInfo.State == dme.CloudletState_CLOUDLET_STATE_OFFLINE {
			return nil
		}
	}

	// Fetch platform specific status
	pfConfig, err := s.all.cloudletApi.getPlatformConfig(ctx, cloudlet)
	if err != nil {
		return err
	}
	cloudletPlatform, err := pfutils.GetPlatform(ctx, cloudlet.PlatformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return fmt.Errorf("Failed to get platform: %v", err)
	}
	accessApi := accessapi.NewVaultClient(cloudlet, vaultConfig, *region)
	updatecb := updateCloudletCallback{cloudlet, cb}
	err = cloudletPlatform.GetRestrictedCloudletStatus(ctx, cloudlet, pfConfig, accessApi, updatecb.cb)
	if err != nil {
		return fmt.Errorf("Failed to get cloudlet run status: %v", err)
	}

	// Fetch cloudlet info status
	lastMsgId := 0
	done := make(chan bool, 1)
	failed := make(chan bool, 1)

	log.SpanLog(ctx, log.DebugLevelApi, "wait for cloudlet state", "key", key)

	checkState := func(key *edgeproto.CloudletKey) {
		cloudlet := edgeproto.Cloudlet{}
		if !s.all.cloudletApi.cache.Get(key, &cloudlet) {
			return
		}
		cloudletInfo := edgeproto.CloudletInfo{}
		if !s.all.cloudletInfoApi.cache.Get(key, &cloudletInfo) {
			return
		}

		curState := cloudletInfo.State

		if curState == dme.CloudletState_CLOUDLET_STATE_ERRORS ||
			curState == dme.CloudletState_CLOUDLET_STATE_OFFLINE {
			failed <- true
			return
		}

		if curState == dme.CloudletState_CLOUDLET_STATE_READY {
			done <- true
			return
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "watch event for CloudletInfo")
	info := edgeproto.CloudletInfo{}
	cancel := s.all.cloudletInfoApi.cache.WatchKey(key, func(ctx context.Context) {
		if !s.all.cloudletInfoApi.cache.Get(key, &info) {
			return
		}
		for ii := lastMsgId; ii < len(info.Status.Msgs); ii++ {
			cb.Send(&edgeproto.Result{Message: info.Status.Msgs[ii]})
			lastMsgId++
		}
		checkState(key)
	})

	// After setting up watch, check current state,
	// as it may have already changed to target state
	checkState(key)

	select {
	case <-done:
		err = nil
		cb.Send(&edgeproto.Result{Message: "Cloudlet setup successfully"})
	case <-failed:
		if s.all.cloudletInfoApi.cache.Get(key, &info) {
			errs := strings.Join(info.Errors, ", ")
			err = fmt.Errorf("Encountered failures: %s", errs)
		} else {
			err = fmt.Errorf("Unknown failure")
		}
		cb.Send(&edgeproto.Result{Message: err.Error()})
	case <-time.After(PlatformInitTimeout):
		err = fmt.Errorf("Timed out waiting for cloudlet state to be Ready")
		cb.Send(&edgeproto.Result{Message: "platform bringup timed out"})
	}

	cancel()

	return err
}

func (s *CloudletApi) CreateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	if in.IpSupport == edgeproto.IpSupport_IP_SUPPORT_UNKNOWN {
		in.IpSupport = edgeproto.IpSupport_IP_SUPPORT_DYNAMIC
	}
	// TODO: support static IP assignment.
	if in.IpSupport != edgeproto.IpSupport_IP_SUPPORT_DYNAMIC {
		return errors.New("Only dynamic IPs are supported currently")
	}
	if in.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
		// TODO: Validate static ips
	} else {
		// dynamic
		if in.NumDynamicIps < 1 {
			return errors.New("Must specify at least one dynamic public IP available")
		}
	}
	if in.Location.Latitude == 0 && in.Location.Longitude == 0 {
		// user forgot to specify location
		return errors.New("location is missing; 0,0 is not a valid location")
	}

	// If notifysrvaddr is empty, set default value
	if in.NotifySrvAddr == "" {
		in.NotifySrvAddr = "127.0.0.1:0"
	}
	if in.SecondaryNotifySrvAddr == "" {
		in.SecondaryNotifySrvAddr = "127.0.0.1:0"
	}

	if in.ContainerVersion == "" {
		in.ContainerVersion = *versionTag
	}

	if in.DefaultResourceAlertThreshold == 0 {
		in.DefaultResourceAlertThreshold = DefaultResourceAlertThreshold
	}

	if in.DeploymentLocal {
		if in.AccessVars != nil {
			return errors.New("Access vars is not supported for local deployment")
		}
		if in.VmImageVersion != "" {
			return errors.New("VM Image version is not supported for local deployment")
		}
		if in.Deployment != "" {
			return errors.New("Deployment type is not supported for local deployment")
		}
		if in.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
			return errors.New("Infra access type private is not supported for local deployment")
		}
	} else {
		if in.Deployment == "" {
			in.Deployment = cloudcommon.DeploymentTypeDocker
		}
		if !cloudcommon.IsValidDeploymentType(in.Deployment, cloudcommon.ValidCloudletDeployments) {
			return fmt.Errorf("Invalid deployment, must be one of %v", cloudcommon.ValidCloudletDeployments)
		}
	}

	if in.GpuConfig.Driver.Name == "" {
		in.GpuConfig = edgeproto.GPUConfig{}
	}
	return s.createCloudletInternal(DefCallContext(), in, cb)
}

func (s *CloudletApi) getCaches(ctx context.Context, vmPool *edgeproto.VMPool) *pf.Caches {
	// Some platform types require caches
	caches := pf.Caches{
		SettingsCache:     &s.all.settingsApi.cache,
		FlavorCache:       &s.all.flavorApi.cache,
		CloudletCache:     s.all.cloudletApi.cache,
		CloudletInfoCache: &s.all.cloudletInfoApi.cache,
	}
	if vmPool != nil && vmPool.Key.Name != "" {
		var vmPoolMux sync.Mutex
		caches.VMPool = vmPool
		caches.VMPoolMux = &vmPoolMux
		caches.VMPoolInfoCache = &s.all.vmPoolInfoApi.cache
		// This is required to update VMPool object on controller
		caches.VMPoolInfoCache.SetUpdatedCb(func(ctx context.Context, old *edgeproto.VMPoolInfo, new *edgeproto.VMPoolInfo) {
			log.SpanLog(ctx, log.DebugLevelInfo, "VMPoolInfo UpdatedCb", "vmpoolinfo", new)
			s.all.vmPoolApi.UpdateFromInfo(ctx, new)
		})

	}
	return &caches
}

func validateResourceQuotaProps(resProps *edgeproto.CloudletResourceQuotaProps, resourceQuotas []edgeproto.ResourceQuota) error {
	resPropsMap := make(map[string]struct{})
	resPropsNames := []string{}
	for _, prop := range resProps.Properties {
		resPropsMap[prop.Name] = struct{}{}
		resPropsNames = append(resPropsNames, prop.Name)
	}
	for _, clRes := range cloudcommon.CloudletResources {
		resPropsMap[clRes.Name] = struct{}{}
		resPropsNames = append(resPropsNames, clRes.Name)
	}
	for _, resQuota := range resourceQuotas {
		if _, ok := resPropsMap[resQuota.Name]; !ok {
			return fmt.Errorf("Invalid quota name: %s, valid names are %s", resQuota.Name, strings.Join(resPropsNames, ","))
		}
	}
	return nil
}

func caseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func caseInsensitiveContainsTimedOut(s string) bool {
	return caseInsensitiveContains(s, "Timed out") || caseInsensitiveContains(s, "timedout")
}

func (s *CloudletApi) createCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, inCb edgeproto.CloudletApi_CreateCloudletServer) (reterr error) {
	cctx.SetOverride(&in.CrmOverride)
	ctx := inCb.Context()

	platName := edgeproto.PlatformType_name[int32(in.PlatformType)]
	features, err := GetCloudletFeatures(ctx, in.PlatformType)
	if err != nil {
		return fmt.Errorf("Failed to get features for platform: %s", err)
	}

	if in.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS &&
		!features.IsVMPool {
		if in.InfraConfig.FlavorName == "" {
			return errors.New("Infra flavor name is required for private deployments")
		}
		if in.InfraConfig.ExternalNetworkName == "" {
			return errors.New("Infra external network is required for private deployments")
		}
	}
	if in.VmPool != "" {
		if !features.IsVMPool {
			return errors.New("VM Pool is only valid for PlatformTypeVmPool")
		}
		vmPoolKey := edgeproto.VMPoolKey{
			Name:         in.VmPool,
			Organization: in.Key.Organization,
		}
		if s.UsesVMPool(&vmPoolKey) {
			return errors.New("VM Pool with this name is already in use by some other Cloudlet")
		}
	} else {
		if features.IsVMPool {
			return errors.New("VM Pool is mandatory for PlatformTypeVmPool")
		}
	}
	if in.EnableDefaultServerlessCluster && !features.SupportsMultiTenantCluster {
		return fmt.Errorf("Serverless cluster not supported on %s", platName)
	}
	if in.TrustPolicy != "" && !features.SupportsTrustPolicy {
		return fmt.Errorf("Trust Policy not supported on %s", platName)
	}
	if in.PlatformHighAvailability && !features.SupportsPlatformHighAvailability {
		return fmt.Errorf("Platform High Availability not supported on %s", platName)
	}
	if err := validateAllianceOrgs(ctx, in); err != nil {
		return err
	}

	cloudletKey := in.Key
	sendObj, cb, err := s.startCloudletStream(ctx, &cloudletKey, inCb)
	if err == nil {
		defer func() {
			s.stopCloudletStream(ctx, &cloudletKey, sendObj, reterr)
		}()
	}

	defer func() {
		if reterr == nil {
			RecordCloudletEvent(ctx, &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

	if in.PhysicalName == "" {
		in.PhysicalName = in.Key.Name
		cb.Send(&edgeproto.Result{Message: "Setting physicalname to match cloudlet name"})
	}

	pfConfig, err := s.getPlatformConfig(ctx, in)
	if err != nil {
		return err
	}

	pfFlavor := edgeproto.Flavor{}
	if in.Flavor.Name == "" {
		in.Flavor = DefaultPlatformFlavor.Key
		pfFlavor = DefaultPlatformFlavor
	}

	accessVars := make(map[string]string)
	if in.AccessVars != nil {
		accessVars = in.AccessVars
		in.AccessVars = nil
	}

	kafkaDetails := node.KafkaCreds{}
	if (in.KafkaUser != "") != (in.KafkaPassword != "") {
		return errors.New("Must specify both kafka username and password, or neither")
	} else if in.KafkaCluster == "" && in.KafkaUser != "" {
		return errors.New("Must specify a kafka cluster endpoint in addition to kafka credentials")
	}
	if in.KafkaCluster != "" {
		kafkaDetails.Endpoint = in.KafkaCluster
		kafkaDetails.Username = in.KafkaUser
		kafkaDetails.Password = in.KafkaPassword
		in.KafkaUser = ""
		in.KafkaPassword = ""
	}

	if in.InfraApiAccess != edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
		accessKey, err := node.GenerateAccessKey()
		if err != nil {
			return err
		}
		in.CrmAccessPublicKey = accessKey.PublicPEM
		in.CrmAccessKeyUpgradeRequired = true
		pfConfig.CrmAccessPrivateKey = accessKey.PrivatePEM

		if in.PlatformHighAvailability {
			secondaryAccessKey, err := node.GenerateAccessKey()
			if err != nil {
				return err
			}
			in.SecondaryCrmAccessPublicKey = secondaryAccessKey.PublicPEM
			in.SecondaryCrmAccessKeyUpgradeRequired = true
			pfConfig.SecondaryCrmAccessPrivateKey = secondaryAccessKey.PrivatePEM
		}
	}

	vmPool := edgeproto.VMPool{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			if !cctx.Undo {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteCloudlet to remove and try again"})
				}
				return in.Key.ExistsError()
			}
			in.Errors = nil
		}
		if in.Flavor.Name != "" && in.Flavor.Name != DefaultPlatformFlavor.Key.Name {
			if !s.all.flavorApi.store.STMGet(stm, &in.Flavor, &pfFlavor) {
				return in.Flavor.NotFoundError()
			}
			if pfFlavor.DeletePrepare {
				return in.Flavor.BeingDeletedError()
			}
		}
		if in.VmPool != "" {
			vmPoolKey := edgeproto.VMPoolKey{
				Name:         in.VmPool,
				Organization: in.Key.Organization,
			}
			if !s.all.vmPoolApi.store.STMGet(stm, &vmPoolKey, &vmPool) {
				return vmPoolKey.NotFoundError()
			}
			if vmPool.DeletePrepare {
				return vmPoolKey.BeingDeletedError()
			}
		}
		if in.GpuConfig.Driver.Name != "" {
			if in.GpuConfig.Driver.Organization != "" && in.GpuConfig.Driver.Organization != in.Key.Organization {
				return fmt.Errorf("Can only use %s or '' org gpu drivers", in.Key.Organization)
			}
			gpuDriver := edgeproto.GPUDriver{}
			if !s.all.gpuDriverApi.store.STMGet(stm, &in.GpuConfig.Driver, &gpuDriver) {
				return in.GpuConfig.Driver.NotFoundError()
			}
			if gpuDriver.DeletePrepare {
				return in.GpuConfig.Driver.BeingDeletedError()
			}
			if gpuDriver.State == ChangeInProgress {
				return fmt.Errorf("GPU driver %s is busy", in.GpuConfig.Driver.String())
			}
			if gpuDriver.LicenseConfig != "" {
				storageClient, err := getGCSStorageClient(ctx)
				if err != nil {
					return err
				}
				defer storageClient.Close()
				url, md5sum, err := setupGPUDriverLicenseConfig(ctx, storageClient, &gpuDriver.Key, gpuDriver.LicenseConfig, in.Key.Name, cb)
				if err != nil {
					return err
				}
				// store the GCS path to license config
				in.GpuConfig.LicenseConfig = url
				in.GpuConfig.LicenseConfigMd5Sum = md5sum
			}
		}
		if in.TrustPolicy != "" {
			policy := edgeproto.TrustPolicy{}
			policy.Key.Name = in.TrustPolicy
			policy.Key.Organization = in.Key.Organization
			if !s.all.trustPolicyApi.store.STMGet(stm, &policy.Key, &policy) {
				return policy.Key.NotFoundError()
			}
			if policy.DeletePrepare {
				return policy.Key.BeingDeletedError()
			}
		}
		for _, rttKey := range in.ResTagMap {
			resTagTable := edgeproto.ResTagTable{}
			if !s.all.resTagTableApi.store.STMGet(stm, rttKey, &resTagTable) {
				return rttKey.NotFoundError()
			}
			if resTagTable.DeletePrepare {
				return rttKey.BeingDeletedError()
			}
		}

		err := in.Validate(edgeproto.CloudletAllFieldsMap)
		if err != nil {
			return err
		}
		if features.IsSingleKubernetesCluster {
			// create ClusterInst representation of Cloudlet
			s.all.clusterInstApi.createCloudletSingularCluster(stm, &in.Key, in.SingleKubernetesClusterOwner)
		}

		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())

		if ignoreCRMState(cctx) {
			in.State = edgeproto.TrackedState_READY
		} else {
			in.State = edgeproto.TrackedState_CREATING
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	if ignoreCRMState(cctx) {
		return nil
	}
	// store kafka details
	if kafkaDetails.Endpoint != "" {
		path := node.GetKafkaVaultPath(*region, in.Key.Name, in.Key.Organization)
		err = vault.PutData(vaultConfig, path, kafkaDetails)
		if err != nil {
			return fmt.Errorf("Unable to store kafka details: %s", err)
		}
	}

	var cloudletPlatform pf.Platform
	deleteAccessVars := false
	updatecb := updateCloudletCallback{in, cb}
	cloudletResourcesCreated := false

	if in.DeploymentLocal {
		updatecb.cb(edgeproto.UpdateTask, "Starting CRMServer")
		err = cloudcommon.StartCRMService(ctx, in, pfConfig, process.HARolePrimary, process.NoRedisAddr)
	} else {
		cloudletPlatform, err = pfutils.GetPlatform(ctx, in.PlatformType.String(), nodeMgr.UpdateNodeProps)
		if err == nil {
			if len(accessVars) > 0 {
				err = cloudletPlatform.SaveCloudletAccessVars(ctx, in, accessVars, pfConfig, nodeMgr.VaultConfig, updatecb.cb)
				if err != nil {
					return err
				}
			}
			var resProps *edgeproto.CloudletResourceQuotaProps
			resProps, err = cloudletPlatform.GetCloudletResourceQuotaProps(ctx)
			if err != nil {
				return err
			}
			err = validateResourceQuotaProps(resProps, in.ResourceQuotas)
			if err == nil {
				// Some platform types require caches
				caches := s.getCaches(ctx, &vmPool)
				accessApi := accessapi.NewVaultClient(in, vaultConfig, *region)
				cloudletResourcesCreated, err = cloudletPlatform.CreateCloudlet(ctx, in, pfConfig, &pfFlavor, caches, accessApi, updatecb.cb)
				if err != nil && len(accessVars) > 0 {
					deleteAccessVars = true
				}
			}
		}
	}

	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create Cloudlet ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created Cloudlet successfully"})
		return nil
	}

	if err == nil {
		cloudlet := edgeproto.Cloudlet{}
		err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			saveCloudlet := false
			if !s.store.STMGet(stm, &in.Key, &cloudlet) {
				return in.Key.NotFoundError()
			}
			if cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
				cloudlet.State = edgeproto.TrackedState_READY
				saveCloudlet = true
			}
			if in.ChefClientKey != nil {
				cloudlet.ChefClientKey = in.ChefClientKey
				saveCloudlet = true
			}
			if in.DeploymentLocal || features.CloudletServicesLocal {
				// Store controller address if crmserver is started locally
				cloudlet.HostController = *externalApiAddr
				saveCloudlet = true
			}
			if saveCloudlet {
				s.store.STMPut(stm, &cloudlet)
			}
			return nil
		})
		if err != nil {
			return err
		}
		if in.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
			cb.Send(&edgeproto.Result{
				Message: "Cloudlet configured successfully. Please run `GetCloudletManifest` to bringup Platform VM(s) for cloudlet services",
			})
			return nil
		}
		// Wait for CRM to connect to controller
		streamKey := edgeproto.GetStreamKeyFromCloudletKey(&in.Key)
		go func() {
			err := cloudcommon.CrmServiceWait(in.Key)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "failed to cleanup crm service", "err", err)
			}
		}()
		err = s.cache.WaitForState(
			ctx, &in.Key,
			edgeproto.TrackedState_READY,
			CreateCloudletTransitions, edgeproto.TrackedState_CREATE_ERROR,
			s.all.settingsApi.Get().CreateCloudletTimeout.TimeDuration(),
			"Created Cloudlet successfully", cb.Send,
			edgeproto.WithStreamObj(redisClient, &streamKey))
	} else {
		cb.Send(&edgeproto.Result{Message: err.Error()})
	}

	if err != nil {
		cb.Send(&edgeproto.Result{Message: "Deleting Cloudlet due to failures"})
		log.SpanLog(ctx, log.DebugLevelInfo, "deleting cloudlet due to failures", "cloudletResourcesCreated", cloudletResourcesCreated)
		undoErr := s.deleteCloudletInternal(cctx.WithUndo(), in, cb, cloudletResourcesCreated)
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Undo create Cloudlet", "undoErr", undoErr)
		}
	}
	if deleteAccessVars {
		err1 := cloudletPlatform.DeleteCloudletAccessVars(ctx, in, pfConfig, nodeMgr.VaultConfig, updatecb.cb)
		if err1 != nil {
			cb.Send(&edgeproto.Result{Message: err1.Error()})
		}
	}
	return err
}

func (s *CloudletApi) VerifyTrustPoliciesForAppInsts(ctx context.Context, app *edgeproto.App, appInsts map[edgeproto.AppInstKey]struct{}) error {
	TrustPolicies := make(map[edgeproto.PolicyKey]*edgeproto.TrustPolicy)
	s.all.trustPolicyApi.GetTrustPolicies(TrustPolicies)
	s.cache.Mux.Lock()
	trustedCloudlets := make(map[edgeproto.CloudletKey]*edgeproto.PolicyKey)
	for key, data := range s.cache.Objs {
		val := data.Obj
		if val.TrustPolicy != "" {
			pkey := edgeproto.PolicyKey{
				Organization: val.Key.Organization,
				Name:         val.TrustPolicy,
			}
			trustedCloudlets[key] = &pkey
		}

	}
	s.cache.Mux.Unlock()
	for akey := range appInsts {
		pkey, cloudletFound := trustedCloudlets[akey.ClusterInstKey.CloudletKey]
		if cloudletFound {
			policy, policyFound := TrustPolicies[*pkey]
			if !policyFound {
				return fmt.Errorf("Unable to find trust policy in cache: %s", pkey.String())
			}
			err := s.all.appApi.CheckAppCompatibleWithTrustPolicy(ctx, &akey.ClusterInstKey.CloudletKey, app, policy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// updateTrustPolicyInternal updates the TrustPolicyState to TrackedState_UPDATE_REQUESTED
// and then waits for the update to complete.
func (s *CloudletApi) updateTrustPolicyInternal(ctx context.Context, ckey *edgeproto.CloudletKey, policyName string, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	log.SpanLog(ctx, log.DebugLevelApi, "updateTrustPolicyInternal", "policyName", policyName)

	err := cb.Send(&edgeproto.Result{
		Message: fmt.Sprintf("Doing TrustPolicy: %s Update for Cloudlet: %s", policyName, ckey.String()),
	})
	if err != nil {
		return err
	}
	cloudletInfo := edgeproto.CloudletInfo{}
	cloudlet := &edgeproto.Cloudlet{}
	var updateErr error
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, ckey, cloudlet) {
			return ckey.NotFoundError()
		}
		if !s.all.cloudletInfoApi.cache.Get(ckey, &cloudletInfo) {
			updateErr = fmt.Errorf("CloudletInfo not found for %s", ckey.String())
		} else {
			if cloudletInfo.State != dme.CloudletState_CLOUDLET_STATE_READY {
				updateErr = fmt.Errorf("Cannot modify trust policy for cloudlet in state: %s", cloudletInfo.State)
			}
		}
		if updateErr != nil {
			cloudlet.TrustPolicyState = edgeproto.TrackedState_UPDATE_ERROR
		} else {
			cloudlet.TrustPolicyState = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		cloudlet.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, cloudlet)
		return nil
	})
	if err != nil {
		return err
	}
	if updateErr != nil {
		return updateErr
	}
	targetState := edgeproto.TrackedState_READY
	if policyName == "" {
		targetState = edgeproto.TrackedState_NOT_PRESENT
	}
	err = s.WaitForTrustPolicyState(ctx, ckey, targetState, edgeproto.TrackedState_UPDATE_ERROR, s.all.settingsApi.Get().UpdateTrustPolicyTimeout.TimeDuration())
	if err == nil {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Successful TrustPolicy: %s Update for Cloudlet: %s", policyName, ckey.String())})
	} else if caseInsensitiveContainsTimedOut(err.Error()) {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("In progress TrustPolicy: %s Update for Cloudlet: %s -- %v", policyName, ckey.String(), err.Error())})
	} else {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Failed TrustPolicy: %s Update for Cloudlet: %s -- %v", policyName, ckey.String(), err.Error())})
	}
	return err
}

func (s *CloudletApi) UpdateCloudlet(in *edgeproto.Cloudlet, inCb edgeproto.CloudletApi_UpdateCloudletServer) (reterr error) {
	ctx := inCb.Context()

	if err := validateAllianceOrgs(ctx, in); err != nil {
		return err
	}

	cloudletKey := in.Key
	sendObj, cb, err := s.startCloudletStream(ctx, &cloudletKey, inCb)
	if err == nil {
		defer func() {
			s.stopCloudletStream(ctx, &cloudletKey, sendObj, reterr)
		}()
	}

	updatecb := updateCloudletCallback{in, cb}

	err = in.ValidateUpdateFields()
	if err != nil {
		return err
	}

	fmap := edgeproto.MakeFieldMap(in.Fields)
	if _, found := fmap[edgeproto.CloudletFieldNumDynamicIps]; found {
		staticSet := false
		if _, staticFound := fmap[edgeproto.CloudletFieldIpSupport]; staticFound {
			if in.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
				staticSet = true
			}
		}
		if in.NumDynamicIps < 1 && !staticSet {
			return errors.New("Cannot specify less than one dynamic IP unless Ip Support Static is specified")
		}
	}

	err = in.Validate(fmap)
	if err != nil {
		return err
	}

	cur := &edgeproto.Cloudlet{}
	if !s.cache.Get(&in.Key, cur) {
		return in.Key.NotFoundError()
	}
	features, err := GetCloudletFeatures(ctx, cur.PlatformType)
	if err != nil {
		return fmt.Errorf("Failed to get features for platform: %s", err)
	}

	accessVars := make(map[string]string)
	if _, found := fmap[edgeproto.CloudletFieldAccessVars]; found {
		if in.DeploymentLocal {
			return errors.New("Access vars is not supported for local deployment")
		}
		accessVars = in.AccessVars
		in.AccessVars = nil
	}

	_, singleKubernetesClusterOwnerSet := fmap[edgeproto.CloudletFieldSingleKubernetesClusterOwner]
	if singleKubernetesClusterOwnerSet {
		// TODO: to support this, we need to use the ClusterRefs
		// to make sure no AppInsts exist, and then we need to delete
		// the current default single cluster and create a new one with
		// the new org.
		return fmt.Errorf("Changing the single kubernetes cluster owner is not supported yet")
	}

	_, kafkaClusterChanged := fmap[edgeproto.CloudletFieldKafkaCluster]
	_, kafkaUserChanged := fmap[edgeproto.CloudletFieldKafkaUser]
	_, kafkaPasswordChanged := fmap[edgeproto.CloudletFieldKafkaPassword]
	if kafkaClusterChanged && in.KafkaCluster == "" {
		in.KafkaCluster = ""
		in.KafkaUser = ""
		in.KafkaPassword = ""
		client, err := vaultConfig.Login()
		if err == nil {
			vault.DeleteKV(client, node.GetKafkaVaultPath(*region, in.Key.Name, in.Key.Organization))
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to login in to vault to delete kafka credentials", "err", err)
		}
	} else if kafkaClusterChanged || kafkaUserChanged || kafkaPasswordChanged {
		// get existing data
		kafkaCreds := node.KafkaCreds{}
		path := node.GetKafkaVaultPath(*region, in.Key.Name, in.Key.Organization)
		err := vault.GetData(vaultConfig, path, 0, &kafkaCreds)
		if kafkaClusterChanged {
			kafkaCreds.Endpoint = in.KafkaCluster
		}
		if kafkaUserChanged {
			kafkaCreds.Username = in.KafkaUser
		}
		if kafkaPasswordChanged {
			kafkaCreds.Password = in.KafkaPassword
		}
		if kafkaUserChanged != kafkaPasswordChanged {
			return errors.New("Must specify both kafka username and password, or neither")
		}
		// must specify either just a new endpoint, or everything
		if !kafkaClusterChanged && kafkaUserChanged {
			return errors.New("Please also specify endpoint when changing username and password")
		}
		// write back changes
		err = vault.PutData(vaultConfig, path, kafkaCreds)
		if err != nil {
			return fmt.Errorf("Unable to store kafka details: %s", err)
		}
	}
	in.KafkaUser = ""
	in.KafkaPassword = ""

	if _, found := fmap[edgeproto.CloudletFieldGpuConfigLicenseConfig]; found {
		storageClient, err := getGCSStorageClient(ctx)
		if err != nil {
			return err
		}
		defer storageClient.Close()
		if in.GpuConfig.LicenseConfig == "" {
			cb.Send(&edgeproto.Result{Message: "Deleting GPU driver license config from secure storage"})
			// Delete license config from GCS
			err = deleteGPUDriverLicenseConfig(ctx, storageClient, &in.GpuConfig.Driver, in.Key.Name)
			if err != nil {
				return err
			}
			in.GpuConfig.LicenseConfigMd5Sum = ""
		} else {
			url, md5sum, err := setupGPUDriverLicenseConfig(ctx, storageClient, &in.GpuConfig.Driver, in.GpuConfig.LicenseConfig, in.Key.Name, cb)
			if err != nil {
				return err
			}
			// store the GCS path to license config
			in.GpuConfig.LicenseConfig = url
			in.GpuConfig.LicenseConfigMd5Sum = md5sum
		}
		in.Fields = append(in.Fields, edgeproto.CloudletFieldGpuConfigLicenseConfigMd5Sum)
	}

	crmUpdateReqd := false
	if _, found := fmap[edgeproto.CloudletFieldEnvVar]; found {
		if _, found := fmap[edgeproto.CloudletFieldMaintenanceState]; found {
			return errors.New("Cannot set envvars if maintenance state is set")
		}
		crmUpdateReqd = true
	}

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	if !ignoreCRMState(cctx) {
		var cloudletPlatform pf.Platform
		cloudletPlatform, err := pfutils.GetPlatform(ctx, cur.PlatformType.String(), nodeMgr.UpdateNodeProps)
		if err != nil {
			return err
		}
		pfConfig, err := s.getPlatformConfig(ctx, in)
		if err != nil {
			return err
		}
		if _, found := fmap[edgeproto.CloudletFieldResourceQuotas]; found {
			resProps, err := cloudletPlatform.GetCloudletResourceQuotaProps(ctx)
			if err != nil {
				return err
			}
			err = validateResourceQuotaProps(resProps, in.ResourceQuotas)
			if err != nil {
				return err
			}
			crmUpdateReqd = true
		}

		if len(accessVars) > 0 {
			err = cloudletPlatform.SaveCloudletAccessVars(ctx, in, accessVars, pfConfig, nodeMgr.VaultConfig, updatecb.cb)
			if err != nil {
				return err
			}
		}
	}

	var newMaintenanceState dme.MaintenanceState
	maintenanceChanged := false
	_, privPolUpdateRequested := fmap[edgeproto.CloudletFieldTrustPolicy]

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur = &edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, cur) {
			return in.Key.NotFoundError()
		}
		cloudletInfo := edgeproto.CloudletInfo{}
		if !s.all.cloudletInfoApi.store.STMGet(stm, &in.Key, &cloudletInfo) {
			return fmt.Errorf("Missing cloudlet info: %v", in.Key)
		}
		cloudletRefs := edgeproto.CloudletRefs{}
		s.all.cloudletRefsApi.store.STMGet(stm, &in.Key, &cloudletRefs)
		if _, found := fmap[edgeproto.CloudletFieldResourceQuotas]; found {
			// get all cloudlet resources (platformVM, sharedRootLB, clusterVms, AppVMs, etc)
			allVmResources, _, _, err := s.all.clusterInstApi.getAllCloudletResources(ctx, stm, cur, &cloudletInfo, &cloudletRefs)
			if err != nil {
				return err
			}
			infraResInfo := make(map[string]edgeproto.InfraResource)
			for _, resInfo := range cloudletInfo.ResourcesSnapshot.Info {
				infraResInfo[resInfo.Name] = resInfo
			}

			allResInfo, err := s.all.cloudletApi.GetCloudletResourceInfo(ctx, stm, cur, allVmResources, infraResInfo)
			if err != nil {
				return err
			}
			err = cloudcommon.ValidateCloudletResourceQuotas(ctx, allResInfo, in.ResourceQuotas)
			if err != nil {
				return err
			}
		}
		if _, found := fmap[edgeproto.CloudletFieldGpuConfig]; found {
			if in.GpuConfig.Driver.Name == "" {
				// clear GPU config
				in.GpuConfig = edgeproto.GPUConfig{}
				in.Fields = append(in.Fields, edgeproto.CloudletFieldGpuConfigDriverName)
				in.Fields = append(in.Fields, edgeproto.CloudletFieldGpuConfigDriverOrganization)
				in.Fields = append(in.Fields, edgeproto.CloudletFieldGpuConfigProperties)
			} else {
				if in.GpuConfig.Driver.Organization != "" && in.GpuConfig.Driver.Organization != in.Key.Organization {
					return fmt.Errorf("Can only use %s or '' org gpu drivers", in.Key.Organization)
				}
				gpuDriver := edgeproto.GPUDriver{}
				if !s.all.gpuDriverApi.store.STMGet(stm, &in.GpuConfig.Driver, &gpuDriver) {
					return fmt.Errorf("GPU driver %s not found", in.GpuConfig.Driver.String())
				}
				if gpuDriver.DeletePrepare {
					return in.GpuConfig.Driver.BeingDeletedError()
				}
				if gpuDriver.State == ChangeInProgress {
					return fmt.Errorf("GPU driver %s is busy", in.GpuConfig.Driver.String())
				}
			}
			crmUpdateReqd = true
		}

		old := edgeproto.Cloudlet{}
		old.DeepCopyIn(cur)
		cur.CopyInFields(in)
		newMaintenanceState = cur.MaintenanceState
		if newMaintenanceState != old.MaintenanceState {
			maintenanceChanged = true
			// don't change maintenance here, we handle it below
			cur.MaintenanceState = old.MaintenanceState
		}
		if newMaintenanceState == dme.MaintenanceState_MAINTENANCE_START || newMaintenanceState == dme.MaintenanceState_MAINTENANCE_START_NO_FAILOVER {
			// return error when trying to put into maintenance but current state is not normal
			if old.MaintenanceState != dme.MaintenanceState_NORMAL_OPERATION {
				return fmt.Errorf("Cloudlet must be in NormalOperation before starting maintenance")
			}
		}
		platName := edgeproto.PlatformType_name[int32(cur.PlatformType)]
		if privPolUpdateRequested {
			if maintenanceChanged {
				return fmt.Errorf("Cannot change both maintenance state and trust policy at the same time")
			}
			if !ignoreCRM(cctx) {
				if cur.State != edgeproto.TrackedState_READY {
					return fmt.Errorf("Trust policy cannot be changed while cloudlet is not ready")
				}
			}
			if in.TrustPolicy != "" {
				if !features.SupportsTrustPolicy {
					return fmt.Errorf("Trust Policy not supported on %s", platName)
				}
				policy := edgeproto.TrustPolicy{}
				policy.Key.Name = in.TrustPolicy
				policy.Key.Organization = in.Key.Organization
				if !s.all.trustPolicyApi.store.STMGet(stm, &policy.Key, &policy) {
					return policy.Key.NotFoundError()
				}
				if policy.DeletePrepare {
					return policy.Key.BeingDeletedError()
				}
				if err := s.all.appInstApi.CheckCloudletAppinstsCompatibleWithTrustPolicy(ctx, &in.Key, &policy); err != nil {
					return err
				}
			}
		}
		if old.EnableDefaultServerlessCluster != cur.EnableDefaultServerlessCluster {
			if maintenanceChanged {
				return fmt.Errorf("Cannot change both enable default serverless cluster and maintenance state")
			}
			if cur.EnableDefaultServerlessCluster {
				if !features.SupportsMultiTenantCluster {
					return fmt.Errorf("Serverless cluster not supported on %s", platName)
				}
				go s.all.clusterInstApi.createDefaultMultiTenantCluster(ctx, cur.Key)
			} else {
				go s.all.clusterInstApi.deleteDefaultMultiTenantCluster(ctx, cur.Key)
			}
		}

		if crmUpdateReqd && !ignoreCRM(cctx) {
			cur.State = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		if privPolUpdateRequested {
			if ignoreCRM(cctx) {
				if cur.TrustPolicy != "" {
					cur.TrustPolicyState = edgeproto.TrackedState_READY
				} else {
					cur.TrustPolicyState = edgeproto.TrackedState_NOT_PRESENT
				}
			}
		}
		cur.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, cur)
		return nil
	})

	if err != nil {
		return err
	}

	// after the cloudlet change is committed, if the location changed,
	// update app insts as well.
	s.UpdateAppInstLocations(ctx, in)

	if crmUpdateReqd && !ignoreCRM(cctx) {
		// Wait for cloudlet to finish upgrading
		streamKey := edgeproto.GetStreamKeyFromCloudletKey(&in.Key)
		err = s.cache.WaitForState(
			ctx, &in.Key,
			edgeproto.TrackedState_READY,
			UpdateCloudletTransitions, edgeproto.TrackedState_UPDATE_ERROR,
			s.all.settingsApi.Get().UpdateCloudletTimeout.TimeDuration(),
			"Cloudlet updated successfully", cb.Send,
			edgeproto.WithStreamObj(redisClient, &streamKey))
		return err
	}
	if privPolUpdateRequested && !ignoreCRM(cctx) {
		// Wait for policy to update
		return s.updateTrustPolicyInternal(ctx, &in.Key, in.TrustPolicy, cb)
	}

	// since default maintenance state is NORMAL_OPERATION, it is better to check
	// if the field is set before handling maintenance state
	if _, found := fmap[edgeproto.CloudletFieldMaintenanceState]; !found || !maintenanceChanged {
		cb.Send(&edgeproto.Result{Message: "Cloudlet updated successfully"})
		return nil
	}
	switch newMaintenanceState {
	case dme.MaintenanceState_NORMAL_OPERATION:
		log.SpanLog(ctx, log.DebugLevelApi, "Stop CRM maintenance")
		if !ignoreCRMState(cctx) {
			timeout := s.all.settingsApi.Get().CloudletMaintenanceTimeout.TimeDuration()
			err = s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_NORMAL_OPERATION_INIT)
			if err != nil {
				return err
			}
			cloudletInfo := edgeproto.CloudletInfo{}
			err = s.all.cloudletInfoApi.waitForMaintenanceState(ctx, &in.Key, dme.MaintenanceState_NORMAL_OPERATION, dme.MaintenanceState_CRM_ERROR, timeout, &cloudletInfo)
			if err != nil {
				return err
			}
			if cloudletInfo.MaintenanceState == dme.MaintenanceState_CRM_ERROR {
				return fmt.Errorf("CRM encountered some errors, aborting")
			}
		}
		err = s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_NORMAL_OPERATION)
		if err != nil {
			return err
		}
		cb.Send(&edgeproto.Result{Message: "Cloudlet is back to normal operation"})
	case dme.MaintenanceState_MAINTENANCE_START:
		// This is a state machine to transition into cloudlet
		// maintenance. Start by triggering AutoProv failovers.
		log.SpanLog(ctx, log.DebugLevelApi, "Start AutoProv failover")
		timeout := s.all.settingsApi.Get().CloudletMaintenanceTimeout.TimeDuration()
		err := cb.Send(&edgeproto.Result{
			Message: "Starting AutoProv failover",
		})
		if err != nil {
			return err
		}
		autoProvInfo := edgeproto.AutoProvInfo{}
		// first reset any old AutoProvInfo
		autoProvInfo = edgeproto.AutoProvInfo{
			Key:              in.Key,
			MaintenanceState: dme.MaintenanceState_NORMAL_OPERATION,
		}
		s.all.autoProvInfoApi.Update(ctx, &autoProvInfo, 0)

		err = s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_FAILOVER_REQUESTED)
		if err != nil {
			return err
		}
		err = s.all.autoProvInfoApi.waitForMaintenanceState(ctx, &in.Key, dme.MaintenanceState_FAILOVER_DONE, dme.MaintenanceState_FAILOVER_ERROR, timeout, &autoProvInfo)
		if err != nil {
			return err
		}
		for _, str := range autoProvInfo.Completed {
			res := edgeproto.Result{
				Message: str,
			}
			if err := cb.Send(&res); err != nil {
				return err
			}
		}
		for _, str := range autoProvInfo.Errors {
			res := edgeproto.Result{
				Message: str,
			}
			if err := cb.Send(&res); err != nil {
				return err
			}
		}
		if len(autoProvInfo.Errors) > 0 {
			undoErr := s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_NORMAL_OPERATION)
			log.SpanLog(ctx, log.DebugLevelApi, "AutoProv maintenance failures", "err", err, "undoErr", undoErr)
			return fmt.Errorf("AutoProv failover encountered some errors, aborting maintenance")
		}
		cb.Send(&edgeproto.Result{
			Message: "AutoProv failover completed",
		})

		log.SpanLog(ctx, log.DebugLevelApi, "AutoProv failover complete")

		// proceed to next state
		fallthrough
	case dme.MaintenanceState_MAINTENANCE_START_NO_FAILOVER:
		log.SpanLog(ctx, log.DebugLevelApi, "Start CRM maintenance")
		cb.Send(&edgeproto.Result{
			Message: "Starting CRM maintenance",
		})
		if !ignoreCRMState(cctx) {
			timeout := s.all.settingsApi.Get().CloudletMaintenanceTimeout.TimeDuration()
			// Tell CRM to go into maintenance mode
			err = s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_CRM_REQUESTED)
			if err != nil {
				return err
			}
			cloudletInfo := edgeproto.CloudletInfo{}
			err = s.all.cloudletInfoApi.waitForMaintenanceState(ctx, &in.Key, dme.MaintenanceState_CRM_UNDER_MAINTENANCE, dme.MaintenanceState_CRM_ERROR, timeout, &cloudletInfo)
			if err != nil {
				return err
			}
			if cloudletInfo.MaintenanceState == dme.MaintenanceState_CRM_ERROR {
				undoErr := s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_NORMAL_OPERATION)
				log.SpanLog(ctx, log.DebugLevelApi, "CRM maintenance failures", "err", err, "undoErr", undoErr)
				return fmt.Errorf("CRM encountered some errors, aborting maintenance")
			}
		}
		cb.Send(&edgeproto.Result{
			Message: "CRM maintenance started",
		})
		log.SpanLog(ctx, log.DebugLevelApi, "CRM maintenance started")
		// transition to maintenance
		err = s.setMaintenanceState(ctx, &in.Key, dme.MaintenanceState_UNDER_MAINTENANCE)
		if err != nil {
			return err
		}
		cb.Send(&edgeproto.Result{
			Message: "Cloudlet is in maintenance",
		})
	}
	return nil
}

func (s *CloudletApi) setMaintenanceState(ctx context.Context, key *edgeproto.CloudletKey, state dme.MaintenanceState) error {
	changedState := false
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := &edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, cur) {
			return key.NotFoundError()
		}
		if cur.MaintenanceState == state {
			return nil
		}
		changedState = true
		cur.MaintenanceState = state
		s.store.STMPut(stm, cur)
		return nil
	})

	msg := ""
	switch state {
	case dme.MaintenanceState_UNDER_MAINTENANCE:
		msg = "Cloudlet maintenance start"
	case dme.MaintenanceState_NORMAL_OPERATION:
		msg = "Cloudlet maintenance done"
	}
	if msg != "" && changedState {
		nodeMgr.Event(ctx, msg, key.Organization, key.GetTags(), nil, "maintenance-state", state.String())
	}
	return err
}

func (s *CloudletApi) PlatformDeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_PlatformDeleteCloudletServer) error {
	ctx := cb.Context()
	updatecb := updateCloudletCallback{in, cb}
	var cloudletPlatform pf.Platform
	cloudletPlatform, err := pfutils.GetPlatform(ctx, in.PlatformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return err
	}

	var pfConfig *edgeproto.PlatformConfig
	vmPool := edgeproto.VMPool{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		var err error
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		pfConfig, err = s.getPlatformConfig(cb.Context(), in)
		if err != nil {
			return err
		}
		if in.VmPool != "" {
			vmPoolKey := edgeproto.VMPoolKey{
				Name:         in.VmPool,
				Organization: in.Key.Organization,
			}
			if !s.all.vmPoolApi.store.STMGet(stm, &vmPoolKey, &vmPool) {
				return fmt.Errorf("VM Pool %s not found", in.VmPool)
			}
		}
		in.State = edgeproto.TrackedState_DELETE_REQUESTED
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	if in.DeploymentLocal {
		updatecb.cb(edgeproto.UpdateTask, "Stopping CRMServer")
		return cloudcommon.StopCRMService(ctx, in, process.HARoleAll)
	}

	// Some platform types require caches
	caches := s.getCaches(ctx, &vmPool)
	accessApi := accessapi.NewVaultClient(in, vaultConfig, *region)
	return cloudletPlatform.DeleteCloudlet(ctx, in, pfConfig, caches, accessApi, updatecb.cb)
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	return s.deleteCloudletInternal(DefCallContext(), in, cb, true)
}

func (s *CloudletApi) deleteCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, inCb edgeproto.CloudletApi_DeleteCloudletServer, cloudletResourcesCreated bool) (reterr error) {
	ctx := inCb.Context()

	cloudletKey := in.Key
	sendObj, cb, err := s.startCloudletStream(ctx, &cloudletKey, inCb)
	if err == nil {
		defer func() {
			s.stopCloudletStream(ctx, &cloudletKey, sendObj, reterr)
		}()
	}

	defer func() {
		if reterr == nil {
			RecordCloudletEvent(ctx, &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
		}
	}()

	var dynInsts map[edgeproto.AppInstKey]struct{}
	var clDynInsts map[edgeproto.ClusterInstKey]struct{}

	cctx.SetOverride(&in.CrmOverride)

	var features *platform.Features
	var prevState edgeproto.TrackedState
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		dynInsts = make(map[edgeproto.AppInstKey]struct{})
		clDynInsts = make(map[edgeproto.ClusterInstKey]struct{})
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		if in.DeletePrepare {
			return in.Key.BeingDeletedError()
		}
		features, err = GetCloudletFeatures(ctx, in.PlatformType)
		if err != nil {
			return fmt.Errorf("Failed to get features for platform: %s", err)
		}
		var defaultClustKey *edgeproto.ClusterInstKey
		if features.IsSingleKubernetesCluster {
			defaultClustKey = getDefaultClustKey(in.Key, in.SingleKubernetesClusterOwner)
		}
		refs := edgeproto.CloudletRefs{}
		if s.all.cloudletRefsApi.store.STMGet(stm, &in.Key, &refs) {
			err = s.all.clusterInstApi.deleteCloudletOk(stm, &refs, clDynInsts)
			if err != nil {
				return err
			}
		}
		err = s.all.appInstApi.deleteCloudletOk(stm, &refs, defaultClustKey, dynInsts)
		if err != nil {
			return err
		}
		if err := validateDeleteState(cctx, "Cloudlet", in.State, in.Errors, cb.Send); err != nil {
			return err
		}
		prevState = in.State
		in.DeletePrepare = true
		// TODO: remove redundant DELETE_PREPARE state
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		if reterr == nil {
			return
		}
		undoErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if !s.store.STMGet(stm, &in.Key, in) {
				return in.Key.NotFoundError()
			}
			changed := false
			if in.State == edgeproto.TrackedState_DELETE_PREPARE {
				// restore previous state since we failed pre-delete actions
				in.State = prevState
				changed = true
			}
			if in.DeletePrepare {
				in.DeletePrepare = false
				changed = true
			}
			if changed {
				s.store.STMPut(stm, in)
			}
			return nil
		})
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to undo delete prepare", "key", in.Key, "err", undoErr)
		}
	}()

	// for cascaded deletes of ClusterInst/AppInst, skip cloudlet
	// ready check because it's not ready - it's being deleted.
	cctx.SkipCloudletReadyCheck = true

	autoProvPolicies := s.all.autoProvPolicyApi.UsesCloudlet(&in.Key)
	if len(autoProvPolicies) > 0 {
		strs := []string{}
		for _, key := range autoProvPolicies {
			strs = append(strs, key.GetKeyString())
		}
		return fmt.Errorf("Cloudlet in use by AutoProvPolicy %s", strings.Join(strs, ", "))
	}
	cloudletPoolKeys := s.all.cloudletPoolApi.UsesCloudlet(&in.Key)
	if len(cloudletPoolKeys) > 0 {
		strs := []string{}
		for _, key := range cloudletPoolKeys {
			strs = append(strs, key.GetKeyString())
		}
		return fmt.Errorf("Cloudlet in use by CloudletPool %s", strings.Join(strs, ", "))
	}
	if networkKey := s.all.networkApi.UsesCloudlet(&in.Key); networkKey != nil {
		return fmt.Errorf("Cloudlet in use by Network %s", networkKey.GetKeyString())
	}
	// Delete dynamic instances while Cloudlet is still in database
	// and CRM is still up.
	err = s.all.appInstApi.AutoDeleteAppInsts(ctx, dynInsts, cctx.Override, cb)
	if err != nil {
		return err
	}
	if len(clDynInsts) > 0 {
		for key, _ := range clDynInsts {
			clInst := edgeproto.ClusterInst{Key: key}
			derr := s.all.clusterInstApi.deleteClusterInstInternal(cctx.Clone(), &clInst, cb)
			if derr != nil {
				log.SpanLog(ctx, log.DebugLevelApi,
					"Failed to delete dynamic ClusterInst",
					"key", key, "err", derr)
				return derr
			}
		}
	}

	if !ignoreCRMState(cctx) {
		if in.HostController != "" && in.HostController != *externalApiAddr {
			// connect to Controller where Cloudlet is running and do delete
			conn, cErr := ControllerConnect(ctx, in.HostController)
			if cErr != nil {
				return cErr
			}
			cmd := edgeproto.NewCloudletApiClient(conn)
			stream, sErr := cmd.PlatformDeleteCloudlet(ctx, in)
			if sErr != nil {
				return sErr
			}
			var sMsg *edgeproto.Result
			for {
				sMsg, sErr = stream.Recv()
				if sErr == io.EOF {
					sErr = nil
					break
				}
				if sErr != nil {
					break
				}
				cb.Send(sMsg)
			}
			if sErr != nil {
				return sErr
			}
		} else {
			// run delete on this Controller
			err = s.PlatformDeleteCloudlet(in, cb)
		}
		if err != nil {
			// if we are ignoring CRM errors, or if there were no resources created, proceed with deletion
			if cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS || !cloudletResourcesCreated {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete Cloudlet ignoring CRM failure: %s", err.Error())})
				s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_NOT_PRESENT)
				err = nil
			} else {
				return err
			}
		}
	}

	// Delete cloudlet from database
	updateCloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &updateCloudlet) {
			return in.Key.NotFoundError()
		}
		if !cctx.Undo && err != nil {
			updateCloudlet.State = edgeproto.TrackedState_DELETE_ERROR
			s.store.STMPut(stm, &updateCloudlet)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		s.all.cloudletRefsApi.store.STMDel(stm, &in.Key)
		if features.IsSingleKubernetesCluster {
			s.all.clusterInstApi.deleteCloudletSingularCluster(stm, &in.Key, updateCloudlet.SingleKubernetesClusterOwner)
		}
		cb.Send(&edgeproto.Result{Message: "Deleted Cloudlet successfully"})
		return nil
	})
	if err1 != nil {
		return err1
	}

	if err != nil {
		return err
	}

	// also delete associated info
	// Note: don't delete cloudletinfo, that will get deleted once CRM
	// disconnects. Otherwise if admin deletes/recreates Cloudlet with
	// CRM connected the whole time, we will end up without cloudletInfo.
	// also delete dynamic instances
	if in.KafkaCluster != "" {
		client, err := vaultConfig.Login()
		if err == nil {
			vault.DeleteKV(client, node.GetKafkaVaultPath(*region, in.Key.Name, in.Key.Organization))
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to login in to vault to delete kafka credentials", "key", in.Key, "err", err)
		}
	}
	if updateCloudlet.GpuConfig.LicenseConfig != "" {
		storageClient, err := getGCSStorageClient(ctx)
		if err != nil {
			return err
		}
		defer storageClient.Close()
		// Delete license config from GCS
		err = deleteGPUDriverLicenseConfig(ctx, storageClient, &updateCloudlet.GpuConfig.Driver, in.Key.Name)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to delete GPU driver license config from secure storage", "key", in.Key, "err", err)
		}
	}

	s.all.cloudletInfoApi.cleanupCloudletInfo(ctx, &in.Key)
	s.all.autoProvInfoApi.Delete(ctx, &edgeproto.AutoProvInfo{Key: in.Key}, 0)
	s.all.alertApi.CleanupCloudletAlerts(ctx, &in.Key)
	return nil
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		copy := *obj
		copy.Status = edgeproto.StatusInfo{}
		copy.ChefClientKey = make(map[string]string)
		err := cb.Send(&copy)
		return err
	})
	return err
}

func (s *CloudletApi) RemoveCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	var err error
	cl := edgeproto.Cloudlet{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}

		for resource, _ := range in.Mapping {
			delete(cl.ResTagMap, resource)
		}
		s.store.STMPut(stm, &cl)
		return err
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletApi) AddCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {

	var err error
	cl := edgeproto.Cloudlet{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		if cl.DeletePrepare {
			return in.Key.BeingDeletedError()
		}
		if cl.ResTagMap == nil {
			cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
		}
		for resource, tblname := range in.Mapping {
			if valerr, ok := s.all.resTagTableApi.ValidateResName(ctx, resource); !ok {
				return valerr
			}
			resource = strings.ToLower(resource)
			var key edgeproto.ResTagTableKey
			key.Name = tblname
			key.Organization = in.Key.Organization
			tbl := edgeproto.ResTagTable{}
			if !s.all.resTagTableApi.store.STMGet(stm, &key, &tbl) {
				return key.NotFoundError()
			}
			if tbl.DeletePrepare {
				return key.BeingDeletedError()
			}
			cl.ResTagMap[resource] = &key
		}
		s.store.STMPut(stm, &cl)
		return err
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletApi) UpdateAppInstLocations(ctx context.Context, in *edgeproto.Cloudlet) {
	fmap := edgeproto.MakeFieldMap(in.Fields)
	if _, found := fmap[edgeproto.CloudletFieldLocation]; !found {
		// no location fields updated
		return
	}

	// find all appinsts associated with the cloudlet
	keys := make([]edgeproto.AppInstKey, 0)
	s.all.appInstApi.cache.Mux.Lock()
	for _, data := range s.all.appInstApi.cache.Objs {
		inst := data.Obj
		if inst.Key.ClusterInstKey.CloudletKey.Matches(&in.Key) {
			keys = append(keys, inst.Key)
		}
	}
	s.all.appInstApi.cache.Mux.Unlock()

	inst := edgeproto.AppInst{}
	for ii, _ := range keys {
		inst = *s.all.appInstApi.cache.Objs[keys[ii]].Obj
		inst.Fields = make([]string, 0)
		if _, found := fmap[edgeproto.CloudletFieldLocationLatitude]; found {
			inst.CloudletLoc.Latitude = in.Location.Latitude
			inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLatitude)
		}
		if _, found := fmap[edgeproto.CloudletFieldLocationLongitude]; found {
			inst.CloudletLoc.Longitude = in.Location.Longitude
			inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLongitude)
		}
		if len(inst.Fields) == 0 {
			break
		}

		err := s.all.appInstApi.updateAppInstStore(ctx, &inst)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Update AppInst Location",
				"inst", inst, "err", err)
		}
	}
}

func (s *CloudletApi) showCloudletsByKeys(keys map[edgeproto.CloudletKey]struct{}, cb func(obj *edgeproto.Cloudlet) error) error {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for key, data := range s.cache.Objs {
		obj := data.Obj
		if _, found := keys[key]; !found {
			continue
		}
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateAllianceOrgs(ctx context.Context, in *edgeproto.Cloudlet) error {
	// check for duplicate orgs
	// make sure can't add your own org
	orgs := make(map[string]struct{})
	for _, org := range in.AllianceOrgs {
		if org == in.Key.Organization {
			return fmt.Errorf("Cannot add cloudlet's own org %q as alliance org", org)
		}
		if _, ok := orgs[org]; ok {
			return fmt.Errorf("Duplicate alliance org %q specified", org)
		}
		orgs[org] = struct{}{}
	}
	return nil
}

func (s *CloudletApi) AddCloudletAllianceOrg(ctx context.Context, in *edgeproto.CloudletAllianceOrg) (*edgeproto.Result, error) {
	if in.Organization == "" {
		return &edgeproto.Result{}, fmt.Errorf("No alliance organization specified")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cl := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		cl.AllianceOrgs = append(cl.AllianceOrgs, in.Organization)
		if err := validateAllianceOrgs(ctx, &cl); err != nil {
			return err
		}
		s.store.STMPut(stm, &cl)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletApi) RemoveCloudletAllianceOrg(ctx context.Context, in *edgeproto.CloudletAllianceOrg) (*edgeproto.Result, error) {
	if in.Organization == "" {
		return &edgeproto.Result{}, fmt.Errorf("No alliance organization specified")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cl := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		changed := false
		for ii, org := range cl.AllianceOrgs {
			if org != in.Organization {
				continue
			}
			cl.AllianceOrgs = append(cl.AllianceOrgs[:ii], cl.AllianceOrgs[ii+1:]...)
			changed = true
			break
		}
		if !changed {
			return nil
		}
		s.store.STMPut(stm, &cl)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletApi) FindFlavorMatch(ctx context.Context, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, error) {

	cl := edgeproto.Cloudlet{}
	var spec *vmspec.VMCreationSpec
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {

		if !s.all.cloudletApi.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		cli := edgeproto.CloudletInfo{}
		if !s.all.cloudletInfoApi.store.STMGet(stm, &in.Key, &cli) {
			return in.Key.NotFoundError()
		}
		mexFlavor := edgeproto.Flavor{}
		mexFlavor.Key.Name = in.FlavorName
		if !s.all.flavorApi.store.STMGet(stm, &mexFlavor.Key, &mexFlavor) {
			return in.Key.NotFoundError()
		}
		var verr error
		spec, verr = s.all.resTagTableApi.GetVMSpec(ctx, stm, mexFlavor, cl, cli)
		if verr != nil {
			return verr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	in.FlavorName = spec.FlavorName
	in.AvailabilityZone = spec.AvailabilityZone
	return in, nil
}

func RecordCloudletEvent(ctx context.Context, cloudletKey *edgeproto.CloudletKey, event cloudcommon.InstanceEvent, serverStatus string) {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.CloudletEvent
	ts, _ := types.TimestampProto(time.Now())
	metric.Timestamp = *ts
	metric.AddStringVal("cloudletorg", cloudletKey.Organization)
	metric.AddTag("cloudlet", cloudletKey.Name)
	metric.AddStringVal("event", string(event))
	metric.AddStringVal("status", serverStatus)

	services.events.AddMetric(&metric)
}

func (s *CloudletApi) GetCloudletManifest(ctx context.Context, key *edgeproto.CloudletKey) (*edgeproto.CloudletManifest, error) {
	cloudlet := &edgeproto.Cloudlet{}
	if !s.all.cloudletApi.cache.Get(key, cloudlet) {
		return nil, key.NotFoundError()
	}

	features, err := GetCloudletFeatures(ctx, cloudlet.PlatformType)
	if err != nil {
		return nil, fmt.Errorf("Failed to get features for platform: %s", err)
	}

	pfFlavor := edgeproto.Flavor{}
	if !features.IsVMPool {
		if cloudlet.Flavor.Name == "" || cloudlet.Flavor.Name == DefaultPlatformFlavor.Key.Name {
			cloudlet.Flavor = DefaultPlatformFlavor.Key
			pfFlavor = DefaultPlatformFlavor
		} else {
			if !s.all.flavorApi.cache.Get(&cloudlet.Flavor, &pfFlavor) {
				return nil, cloudlet.Flavor.NotFoundError()
			}
		}
	}

	pfConfig, err := s.getPlatformConfig(ctx, cloudlet)
	if err != nil {
		return nil, err
	}
	accessApi := accessapi.NewVaultClient(cloudlet, vaultConfig, *region)
	cloudletPlatform, err := pfutils.GetPlatform(ctx, cloudlet.PlatformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return nil, err
	}
	accessKey, err := node.GenerateAccessKey()
	if err != nil {
		return nil, err
	}
	pfConfig.CrmAccessPrivateKey = accessKey.PrivatePEM
	var secondaryAccessKeyPublicPEM string
	if cloudlet.PlatformHighAvailability {
		secondaryAccessKey, err := node.GenerateAccessKey()
		if err != nil {
			return nil, err
		}
		secondaryAccessKeyPublicPEM = secondaryAccessKey.PublicPEM
		pfConfig.SecondaryCrmAccessPrivateKey = secondaryAccessKey.PrivatePEM
	}
	vmPool := edgeproto.VMPool{}
	caches := s.getCaches(ctx, &vmPool)
	manifest, err := cloudletPlatform.GetCloudletManifest(ctx, cloudlet, pfConfig, accessApi, &pfFlavor, caches)
	if err != nil {
		return nil, err
	}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudlet := &edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, cloudlet) {
			return key.NotFoundError()
		}
		if cloudlet.CrmAccessPublicKey != "" || cloudlet.SecondaryCrmAccessPublicKey != "" {
			return fmt.Errorf("Cloudlet has access key registered, please revoke the current access key first so a new one can be generated for the manifest")
		}
		cloudlet.CrmAccessPublicKey = accessKey.PublicPEM
		cloudlet.CrmAccessKeyUpgradeRequired = true
		if cloudlet.PlatformHighAvailability {
			cloudlet.SecondaryCrmAccessPublicKey = secondaryAccessKeyPublicPEM
			cloudlet.SecondaryCrmAccessKeyUpgradeRequired = true
		}
		s.store.STMPut(stm, cloudlet)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func (s *CloudletApi) UsesVMPool(vmPoolKey *edgeproto.VMPoolKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, data := range s.cache.Objs {
		val := data.Obj
		cVMPoolKey := edgeproto.VMPoolKey{
			Organization: key.Organization,
			Name:         val.VmPool,
		}
		if vmPoolKey.Matches(&cVMPoolKey) {
			return true
		}
	}
	return false
}

func (s *CloudletApi) UsesGPUDriver(driverKey *edgeproto.GPUDriverKey) (bool, []string) {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	cloudlets := []string{}
	inUse := false
	for _, data := range s.cache.Objs {
		val := data.Obj
		if driverKey.Matches(&val.GpuConfig.Driver) {
			cloudlets = append(cloudlets, val.Key.Name)
			inUse = true
		}
	}
	return inUse, cloudlets
}

func (s *CloudletApi) UsesFlavor(key *edgeproto.FlavorKey) *edgeproto.CloudletKey {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for k, data := range s.cache.Objs {
		val := data.Obj
		if val.Flavor.Matches(key) {
			return &k
		}
	}
	return nil
}

func (s *CloudletApi) UsesResTagTable(key *edgeproto.ResTagTableKey) *edgeproto.CloudletKey {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for ck, data := range s.cache.Objs {
		val := data.Obj
		for _, k := range val.ResTagMap {
			if k.Matches(key) {
				return &ck
			}
		}
	}
	return nil
}

func (s *CloudletApi) GetCloudletProps(ctx context.Context, in *edgeproto.CloudletProps) (*edgeproto.CloudletProps, error) {

	cloudletPlatform, err := pfutils.GetPlatform(ctx, in.PlatformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return nil, err
	}

	return cloudletPlatform.GetCloudletProps(ctx)
}

func GetCloudletFeatures(ctx context.Context, platformType edgeproto.PlatformType) (*platform.Features, error) {
	cloudletPlatform, err := pfutils.GetPlatform(ctx, platformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return nil, err
	}
	return cloudletPlatform.GetFeatures(), nil
}

func (s *CloudletApi) RevokeAccessKey(ctx context.Context, key *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudlet := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, &cloudlet) {
			return key.NotFoundError()
		}
		cloudlet.CrmAccessPublicKey = ""
		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelApi, "revoked crm access key", "CloudletKey", *key, "err", err)
	return &edgeproto.Result{}, err
}

func (s *CloudletApi) GenerateAccessKey(ctx context.Context, key *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	res := edgeproto.Result{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		res.Message = ""
		cloudlet := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, &cloudlet) {
			return key.NotFoundError()
		}
		keyPair, err := node.GenerateAccessKey()
		if err != nil {
			return err
		}
		cloudlet.CrmAccessPublicKey = keyPair.PublicPEM
		res.Message = keyPair.PrivatePEM
		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelApi, "generated crm access key", "CloudletKey", *key, "err", err)
	return &res, err
}

func (s *CloudletApi) UsesTrustPolicy(key *edgeproto.PolicyKey, stateMatch edgeproto.TrackedState) *edgeproto.CloudletKey {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for k, data := range s.cache.Objs {
		cloudlet := data.Obj
		if cloudlet.TrustPolicy == key.Name && cloudlet.Key.Organization == key.Organization {
			if stateMatch == edgeproto.TrackedState_TRACKED_STATE_UNKNOWN || stateMatch == cloudlet.State {
				return &k
			}
		}
	}
	return nil
}

func (s *CloudletApi) ValidateCloudletsUsingTrustPolicy(ctx context.Context, trustPolicy *edgeproto.TrustPolicy) error {
	log.SpanLog(ctx, log.DebugLevelApi, "ValidateCloudletsUsingTrustPolicy", "policy", trustPolicy)
	cloudletKeys := make(map[*edgeproto.CloudletKey]struct{})
	s.cache.Mux.Lock()
	for ck, data := range s.cache.Objs {
		val := data.Obj
		if ck.Organization != trustPolicy.Key.Organization || val.TrustPolicy != trustPolicy.Key.Name {
			continue
		}
		copyKey := edgeproto.CloudletKey{
			Organization: ck.Organization,
			Name:         ck.Name,
		}
		cloudletKeys[&copyKey] = struct{}{}
	}
	s.cache.Mux.Unlock()
	for k := range cloudletKeys {
		err := s.all.appInstApi.CheckCloudletAppinstsCompatibleWithTrustPolicy(ctx, k, trustPolicy)
		if err != nil {
			return fmt.Errorf("AppInst on cloudlet %s not compatible with trust policy - %s", strings.TrimSpace(k.String()), err.Error())
		}
	}
	return nil
}

func (s *CloudletApi) UpdateCloudletsUsingTrustPolicy(ctx context.Context, trustPolicy *edgeproto.TrustPolicy, cb edgeproto.TrustPolicyApi_CreateTrustPolicyServer) error {
	s.cache.Mux.Lock()
	type updateResult struct {
		errString string
	}

	updateResults := make(map[edgeproto.CloudletKey]chan updateResult)
	for k, data := range s.cache.Objs {
		val := data.Obj
		if k.Organization != trustPolicy.Key.Organization || val.TrustPolicy != trustPolicy.Key.Name {
			continue
		}

		updateResults[k] = make(chan updateResult)
		go func(k edgeproto.CloudletKey) {
			log.SpanLog(ctx, log.DebugLevelApi, "updating trust policy for cloudlet", "key", k)
			err := s.updateTrustPolicyInternal(ctx, &k, trustPolicy.Key.Name, cb)
			if err == nil {
				updateResults[k] <- updateResult{errString: ""}
			} else {
				updateResults[k] <- updateResult{errString: err.Error()}
			}
		}(k)
	}
	s.cache.Mux.Unlock()
	if len(updateResults) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "no cloudlets matched", "key", trustPolicy.Key)
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Trust policy updated, no cloudlets affected")})
		return nil
	}

	numPassed := 0
	numFailed := 0
	numTotal := 0
	numInProgress := 0
	for k, r := range updateResults {
		numTotal++
		result := <-r
		log.SpanLog(ctx, log.DebugLevelApi, "cloudletUpdateResult ", "key", k, "error", result.errString)
		if result.errString == "" {
			numPassed++
		} else if caseInsensitiveContainsTimedOut(result.errString) {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Update cloudlet is in progress: %s - %s Please use 'cloudlet show' to check current status", k, result.errString)})
			numInProgress++
		} else {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Failed to update cloudlet: %s - %s", k, result.errString)})
			numFailed++
		}
	}
	cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Processed: %d Cloudlets.  Passed: %d InProgress: %d Failed: %d", numTotal, numPassed, numInProgress, numFailed)})
	if numPassed == 0 {
		if numInProgress == 0 {
			return fmt.Errorf("Failed to update trust policy on any cloudlets")
		}
		// If numInProgress is nonzero, there is still at least one cloudlet still doing the update which may eventually succeed.
		// If we return an error here, the UpdateTrustPolicy API itself will fail, and the trust policy in etcd will be reverted to the pre-update state.
		// This could cause an inconsistency, and so better to return nil error in this case. Fall through.
	}
	return nil
}

func (s *CloudletApi) WaitForTrustPolicyState(ctx context.Context, key *edgeproto.CloudletKey, targetState edgeproto.TrackedState, errorState edgeproto.TrackedState, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelApi, "WaitForTrustPolicyState", "target", targetState, "timeout", timeout)
	done := make(chan bool, 1)
	failed := make(chan bool, 1)
	cloudlet := edgeproto.Cloudlet{}
	check := func(ctx context.Context) {
		if !s.cache.Get(key, &cloudlet) {
			log.SpanLog(ctx, log.DebugLevelApi, "Error: WaitForTrustPolicyState cloudlet not found", "key", key)
			failed <- true
		}
		log.SpanLog(ctx, log.DebugLevelApi, "WaitForTrustPolicyState initial get from cache", "curState", cloudlet.TrustPolicyState, "targetState", targetState)
		if cloudlet.TrustPolicyState == targetState {
			done <- true
		} else if cloudlet.TrustPolicyState == errorState {
			failed <- true
		}
	}
	cancel := s.cache.WatchKey(key, check)
	check(ctx)
	var err error
	select {
	case <-done:
	case <-failed:
		err = fmt.Errorf("Error in updating Trust Policy")
	case <-time.After(timeout):
		err = fmt.Errorf("Timed out waiting for Trust Policy")
	}
	cancel()
	log.SpanLog(ctx, log.DebugLevelApi, "WaitForTrustPolicyState state done", "target", targetState, "curState", cloudlet.TrustPolicyState)
	return err
}

func (s *CloudletApi) GetCloudletResourceInfo(ctx context.Context, stm concurrency.STM, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) (map[string]edgeproto.InfraResource, error) {
	resQuotasInfo := make(map[string]edgeproto.InfraResource)
	for _, resQuota := range cloudlet.ResourceQuotas {
		resQuotasInfo[resQuota.Name] = edgeproto.InfraResource{
			Name:           resQuota.Name,
			Value:          resQuota.Value,
			AlertThreshold: resQuota.AlertThreshold,
		}
	}

	cloudletPlatform, err := pfutils.GetPlatform(ctx, cloudlet.PlatformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return nil, err
	}
	cloudletRes := map[string]string{
		// Common Cloudlet Resources
		cloudcommon.ResourceRamMb:       cloudcommon.ResourceRamUnits,
		cloudcommon.ResourceVcpus:       "",
		cloudcommon.ResourceGpus:        "",
		cloudcommon.ResourceExternalIPs: "",
	}
	resInfo := make(map[string]edgeproto.InfraResource)
	for resName, resUnits := range cloudletRes {
		infraResMax := uint64(0)
		if infraRes, ok := infraResMap[resName]; ok {
			infraResMax = infraRes.InfraMaxValue
		}
		thresh := cloudlet.DefaultResourceAlertThreshold
		quotaMax := uint64(0)
		// look up quota if any
		if quota, found := resQuotasInfo[resName]; found {
			if quota.Value != 0 {
				quotaMax = quota.Value
			}
			if quota.AlertThreshold > 0 {
				// Set threshold values from Resource quotas
				thresh = quota.AlertThreshold
			}
		}
		resInfo[resName] = edgeproto.InfraResource{
			Name:           resName,
			Units:          resUnits,
			InfraMaxValue:  infraResMax,
			QuotaMaxValue:  quotaMax,
			AlertThreshold: thresh,
		}
	}

	for _, vmRes := range vmResources {
		if vmRes.VmFlavor != nil {
			ramInfo, ok := resInfo[cloudcommon.ResourceRamMb]
			if ok {
				ramInfo.Value += vmRes.VmFlavor.Ram
				resInfo[cloudcommon.ResourceRamMb] = ramInfo
			}
			vcpusInfo, ok := resInfo[cloudcommon.ResourceVcpus]
			if ok {
				vcpusInfo.Value += vmRes.VmFlavor.Vcpus
				resInfo[cloudcommon.ResourceVcpus] = vcpusInfo
			}
			if s.all.resTagTableApi.UsesGpu(ctx, stm, *vmRes.VmFlavor, *cloudlet) {
				gpusInfo, ok := resInfo[cloudcommon.ResourceGpus]
				if ok {
					gpusInfo.Value += 1
					resInfo[cloudcommon.ResourceGpus] = gpusInfo
				}
			}
			if vmRes.Type == cloudcommon.VMTypeRootLB || vmRes.Type == cloudcommon.VMTypePlatform {
				externalIPInfo, ok := resInfo[cloudcommon.ResourceExternalIPs]
				if ok {
					externalIPInfo.Value += 1
					resInfo[cloudcommon.ResourceExternalIPs] = externalIPInfo
				}
			}
		}
	}
	addResInfo := cloudletPlatform.GetClusterAdditionalResources(ctx, cloudlet, vmResources, infraResMap)
	for k, v := range addResInfo {
		thresh := cloudlet.DefaultResourceAlertThreshold
		quotaMax := uint64(0)
		// look up quota if any
		if quota, found := resQuotasInfo[k]; found {
			if quota.Value != 0 {
				quotaMax = quota.Value
			}
			if quota.AlertThreshold > 0 {
				// Set threshold values from Resource quotas
				thresh = quota.AlertThreshold
			}
		}
		v.AlertThreshold = thresh
		v.QuotaMaxValue = quotaMax
		resInfo[k] = v
	}
	return resInfo, nil
}

// Get actual resource info used by the cloudlet
func (s *CloudletApi) GetResourceUsage(ctx context.Context, stm concurrency.STM, cloudlet *edgeproto.Cloudlet, infraResInfo []edgeproto.InfraResource, diffVmResources []edgeproto.VMResource, infraUsage bool) ([]edgeproto.InfraResource, error) {
	resQuotasInfo := make(map[string]edgeproto.InfraResource)
	for _, resQuota := range cloudlet.ResourceQuotas {
		resQuotasInfo[resQuota.Name] = edgeproto.InfraResource{
			Name:           resQuota.Name,
			Value:          resQuota.Value,
			AlertThreshold: resQuota.AlertThreshold,
		}
	}
	defaultAlertThresh := cloudlet.DefaultResourceAlertThreshold
	infraResInfoMap := make(map[string]edgeproto.InfraResource)
	for _, resInfo := range infraResInfo {
		thresh := defaultAlertThresh
		// look up quota if any
		if quota, found := resQuotasInfo[resInfo.Name]; found {
			if quota.Value > 0 {
				// Set max values from Resource quotas
				resInfo.QuotaMaxValue = quota.Value
			}
			if quota.AlertThreshold > 0 {
				// Set threshold values from Resource quotas
				thresh = quota.AlertThreshold
			}
		}
		if !infraUsage {
			resInfo.Value = 0
		}
		resInfo.AlertThreshold = thresh
		infraResInfoMap[resInfo.Name] = resInfo
	}
	diffResInfo, err := s.GetCloudletResourceInfo(ctx, stm, cloudlet, diffVmResources, infraResInfoMap)
	if err != nil {
		return nil, err
	}
	for resName, resInfo := range diffResInfo {
		if infraResInfo, ok := infraResInfoMap[resName]; ok {
			infraResInfo.Value += resInfo.Value
			infraResInfoMap[resName] = infraResInfo
		} else {
			infraResInfoMap[resName] = resInfo
		}
	}
	out := []edgeproto.InfraResource{}
	for _, val := range infraResInfoMap {
		out = append(out, val)
	}
	// sort keys for stable output order
	sort.Slice(out[:], func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out, nil
}

func (s *CloudletApi) GetCloudletResourceUsage(ctx context.Context, usage *edgeproto.CloudletResourceUsage) (*edgeproto.CloudletResourceUsage, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "GetCloudletResourceUsage", "key", usage.Key)
	cloudletResUsage := edgeproto.CloudletResourceUsage{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudlet := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &usage.Key, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}
		cloudletInfo := edgeproto.CloudletInfo{}
		if !s.all.cloudletInfoApi.store.STMGet(stm, &usage.Key, &cloudletInfo) {
			return fmt.Errorf("No resource information found for Cloudlet %s", usage.Key)
		}
		cloudletRefs := edgeproto.CloudletRefs{}
		s.all.cloudletRefsApi.store.STMGet(stm, &usage.Key, &cloudletRefs)
		allVmResources, diffVmResources, _, err := s.all.clusterInstApi.getAllCloudletResources(ctx, stm, &cloudlet, &cloudletInfo, &cloudletRefs)
		if err != nil {
			return err
		}
		cloudletResUsage.Key = usage.Key
		cloudletResUsage.InfraUsage = usage.InfraUsage
		cloudletResUsage.Info = cloudletInfo.ResourcesSnapshot.Info
		resInfo := []edgeproto.InfraResource{}
		if !usage.InfraUsage {
			resInfo, err = s.GetResourceUsage(ctx, stm, &cloudlet, cloudletInfo.ResourcesSnapshot.Info, allVmResources, usage.InfraUsage)
		} else {
			resInfo, err = s.GetResourceUsage(ctx, stm, &cloudlet, cloudletInfo.ResourcesSnapshot.Info, diffVmResources, usage.InfraUsage)
		}
		if err != nil {
			return err
		}
		cloudletResUsage.Info = resInfo
		return nil
	})
	return &cloudletResUsage, err
}

func GetPlatformVMsResources(ctx context.Context, cloudletInfo *edgeproto.CloudletInfo) ([]edgeproto.VMResource, error) {
	resources := []edgeproto.VMResource{}
	for _, vm := range cloudletInfo.ResourcesSnapshot.PlatformVms {
		if vm.InfraFlavor == "" {
			continue
		}
		for _, flavorInfo := range cloudletInfo.Flavors {
			if flavorInfo.Name == vm.InfraFlavor {
				resources = append(resources, edgeproto.VMResource{
					VmFlavor: flavorInfo,
					Type:     vm.Type,
				})
				break
			}
		}
	}
	return resources, nil
}

func (s *CloudletApi) GetCloudletResourceQuotaProps(ctx context.Context, in *edgeproto.CloudletResourceQuotaProps) (*edgeproto.CloudletResourceQuotaProps, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "GetCloudletResourceQuotaProps", "platformtype", in.PlatformType)
	cloudletPlatform, err := pfutils.GetPlatform(ctx, in.PlatformType.String(), nodeMgr.UpdateNodeProps)
	if err != nil {
		return nil, err
	}

	quotaProps := edgeproto.CloudletResourceQuotaProps{}
	quotaProps.Properties = append(quotaProps.Properties, cloudcommon.CloudletResources...)

	pfQuotaProps, err := cloudletPlatform.GetCloudletResourceQuotaProps(ctx)
	if err != nil {
		return nil, err
	}
	quotaProps.Properties = append(quotaProps.Properties, pfQuotaProps.Properties...)

	return &quotaProps, nil
}

func (s *CloudletApi) ShowFlavorsForCloudlet(in *edgeproto.CloudletKey, cb edgeproto.CloudletApi_ShowFlavorsForCloudletServer) error {
	ctx := cb.Context()
	allMetaFlavors := make(map[edgeproto.FlavorKey]struct{})
	flavorCache := &s.all.flavorApi.cache
	flavorCache.GetAllKeys(ctx, func(k *edgeproto.FlavorKey, modRev int64) {
		allMetaFlavors[*k] = struct{}{}
	})
	cloudletKeys := make(map[edgeproto.CloudletKey]struct{})
	if in.ValidateKey() == nil {
		// only one cloudlet specified
		cloudletKeys[*in] = struct{}{}
	} else {
		// find all matching cloudlets
		s.cache.GetAllKeys(ctx, func(k *edgeproto.CloudletKey, modRef int64) {
			if k.Matches(in, edgeproto.MatchFilter()) {
				cloudletKeys[*k] = struct{}{}
			}
		})
	}
	flavors := make(map[edgeproto.FlavorKey]struct{})
	for cloudletKey, _ := range cloudletKeys {
		log.SpanLog(ctx, log.DebugLevelApi, "ShowFlavorsForCloudlet", "cloudletKey", cloudletKey)
		for flavor, _ := range allMetaFlavors {
			fm := edgeproto.FlavorMatch{
				Key:        cloudletKey,
				FlavorName: flavor.Name,
			}
			match, err := s.FindFlavorMatch(ctx, &fm)
			if err != nil {
				continue
			}
			flavors[flavor] = struct{}{}
			log.SpanLog(ctx, log.DebugLevelApi, "ShowFlavorsForCloudlet match", "metaflavor", flavor, "with", match.FlavorName, "on cloudlet", cloudletKey)
		}
	}
	// convert flavors to list so we can sort
	flavorsList := []edgeproto.FlavorKey{}
	for flavorKey, _ := range flavors {
		flavorsList = append(flavorsList, flavorKey)
	}
	sort.Slice(flavorsList, func(i, j int) bool {
		return flavorsList[i].GetKeyString() < flavorsList[j].GetKeyString()
	})
	for _, flavorKey := range flavorsList {
		err := cb.Send(&flavorKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CloudletApi) GetOrganizationsOnCloudlet(in *edgeproto.CloudletKey, cb edgeproto.CloudletApi_GetOrganizationsOnCloudletServer) error {
	orgs := make(map[string]struct{})
	aiFilter := edgeproto.AppInst{}
	aiFilter.Key.ClusterInstKey.CloudletKey = *in
	s.all.appInstApi.cache.Show(&aiFilter, func(appInst *edgeproto.AppInst) error {
		orgs[appInst.Key.AppKey.Organization] = struct{}{}
		orgs[appInst.Key.ClusterInstKey.Organization] = struct{}{}
		return nil
	})
	ciFilter := edgeproto.ClusterInst{}
	ciFilter.Key.CloudletKey = *in
	s.all.clusterInstApi.cache.Show(&ciFilter, func(clusterInst *edgeproto.ClusterInst) error {
		orgs[clusterInst.Key.Organization] = struct{}{}
		return nil
	})
	for name, _ := range orgs {
		org := &edgeproto.Organization{
			Name: name,
		}
		err := cb.Send(org)
		if err != nil {
			return err
		}
	}
	return nil
}
