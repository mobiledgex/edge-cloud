package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vmspec"
)

type CloudletApi struct {
	sync  *Sync
	store edgeproto.CloudletStore
	cache edgeproto.CloudletCache
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
	cloudletApi           = CloudletApi{}
	DefaultPlatformFlavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "DefaultPlatformFlavor",
		},
		Vcpus: 2,
		Ram:   4096,
		Disk:  20,
	}
)

const (
	PlatformInitTimeout = 20 * time.Minute
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

func InitCloudletApi(sync *Sync) {
	cloudletApi.sync = sync
	cloudletApi.store = edgeproto.NewCloudletStore(sync.store)
	edgeproto.InitCloudletCache(&cloudletApi.cache)
	sync.RegisterCache(&cloudletApi.cache)
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
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

func getCrmEnv(vars map[string]string) {
	// For Vault approle, CRM will have its own role/secret
	if val, ok := os.LookupEnv("VAULT_CRM_ROLE_ID"); ok {
		vars["VAULT_ROLE_ID"] = val
	}
	if val, ok := os.LookupEnv("VAULT_CRM_SECRET_ID"); ok {
		vars["VAULT_SECRET_ID"] = val
	}

	for _, key := range []string{
		"GITHUB_ID",
		"VAULT_TOKEN",
		"JAEGER_ENDPOINT",
	} {
		if val, ok := os.LookupEnv(key); ok {
			vars[key] = val
		}
	}
}

func getPlatformConfig(ctx context.Context, cloudlet *edgeproto.Cloudlet) (*edgeproto.PlatformConfig, error) {
	pfConfig := edgeproto.PlatformConfig{}
	pfConfig.PlatformTag = cloudlet.ContainerVersion
	pfConfig.TlsCertFile = nodeMgr.TlsCertFile
	pfConfig.VaultAddr = nodeMgr.VaultAddr
	pfConfig.UseVaultCas = nodeMgr.InternalPki.UseVaultCAs
	pfConfig.UseVaultCerts = nodeMgr.InternalPki.UseVaultCerts
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
	pfConfig.NotifyCtrlAddrs = *publicAddr + ":" + addrObjs[1]
	pfConfig.Span = log.SpanToString(ctx)
	pfConfig.ChefServerPath = *chefServerPath
	pfConfig.ChefClientInterval = settingsApi.Get().ChefClientInterval
	pfConfig.DeploymentTag = *deploymentTag

	return &pfConfig, nil
}

func isOperatorInfraCloudlet(in *edgeproto.Cloudlet) bool {
	if !in.DeploymentLocal && in.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_OPENSTACK {
		return true
	}
	return false
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

	if in.PhysicalName == "" {
		in.PhysicalName = in.Key.Name
		cb.Send(&edgeproto.Result{Message: "Setting physicalname to match cloudlet name"})
	}

	if in.ContainerVersion == "" {
		in.ContainerVersion = *versionTag
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

	if in.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS &&
		in.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
		if in.InfraConfig.FlavorName == "" {
			return errors.New("Infra flavor name is required for private deployments")
		}
		if in.InfraConfig.ExternalNetworkName == "" {
			return errors.New("Infra external network is required for private deployments")
		}
	}

	if in.VmPool != "" {
		if in.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
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
		if in.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
			return errors.New("VM Pool is mandatory for PlatformTypeVmPool")
		}
	}

	return s.createCloudletInternal(DefCallContext(), in, cb)
}

func (s *CloudletApi) createCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) (reterr error) {
	cctx.SetOverride(&in.CrmOverride)
	ctx := cb.Context()

	defer func() {
		if reterr == nil {
			RecordCloudletEvent(ctx, &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

	pfConfig, err := getPlatformConfig(ctx, in)
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
			if !flavorApi.store.STMGet(stm, &in.Flavor, &pfFlavor) {
				return fmt.Errorf("Platform Flavor %s not found", in.Flavor.Name)
			}
		}
		if in.VmPool != "" {
			vmPoolKey := edgeproto.VMPoolKey{
				Name:         in.VmPool,
				Organization: in.Key.Organization,
			}
			if !vmPoolApi.store.STMGet(stm, &vmPoolKey, &vmPool) {
				return fmt.Errorf("VM Pool %s not found", in.VmPool)
			}
		}
		err := in.Validate(edgeproto.CloudletAllFieldsMap)
		if err != nil {
			return err
		}

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

	var cloudletPlatform pf.Platform
	deleteAccessVars := false
	updatecb := updateCloudletCallback{in, cb}

	if in.DeploymentLocal {
		updatecb.cb(edgeproto.UpdateTask, "Starting CRMServer")
		err = cloudcommon.StartCRMService(ctx, in, pfConfig)
	} else {
		cloudletPlatform, err = pfutils.GetPlatform(ctx, in.PlatformType.String())
		if err == nil {
			if len(accessVars) > 0 {
				err = cloudletPlatform.SaveCloudletAccessVars(ctx, in, accessVars, pfConfig, updatecb.cb)
			}
			if err == nil {
				// Some platform types require caches
				caches := pf.Caches{
					SettingsCache: &settingsApi.cache,
					FlavorCache:   &flavorApi.cache,
				}
				if vmPool.Key.Name != "" {
					var vmPoolMux sync.Mutex
					caches.VMPool = &vmPool
					caches.VMPoolMux = &vmPoolMux
					caches.VMPoolInfoCache = &vmPoolInfoApi.cache
					// This is required to update VMPool object on controller
					caches.VMPoolInfoCache.SetUpdatedCb(func(ctx context.Context, old *edgeproto.VMPoolInfo, new *edgeproto.VMPoolInfo) {
						log.SpanLog(ctx, log.DebugLevelInfo, "VMPoolInfo UpdatedCb", "vmpoolinfo", new)
						vmPoolApi.UpdateFromInfo(ctx, new)
					})

				}
				err = cloudletPlatform.CreateCloudlet(ctx, in, pfConfig, &pfFlavor, &caches, updatecb.cb)
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
		newState := in.State
		if in.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
			newState = edgeproto.TrackedState_READY
		}
		cloudlet := edgeproto.Cloudlet{}
		err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if !s.store.STMGet(stm, &in.Key, &cloudlet) {
				return in.Key.NotFoundError()
			}
			if in.ChefClientKey == nil && cloudlet.State == newState {
				return nil
			}
			cloudlet.ChefClientKey = in.ChefClientKey
			cloudlet.State = newState
			s.store.STMPut(stm, &cloudlet)
			return nil
		})
		if err != nil {
			return err
		}
		if in.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
			if in.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
				cb.Send(&edgeproto.Result{
					Message: "Cloudlet configured successfully. Please run `GetCloudletManifest` to bringup Platform VM(s) for cloudlet services",
				})
			} else {
				cb.Send(&edgeproto.Result{
					Message: "Cloudlet configured successfully. Please bringup cloudlet services manually",
				})
			}
			return nil
		}
		// Wait for CRM to connect to controller
		err = s.WaitForCloudlet(
			ctx, &in.Key,
			edgeproto.TrackedState_CREATE_ERROR, // Set error state
			"Created Cloudlet successfully",     // Set success message
			PlatformInitTimeout, updatecb.cb,
		)
	} else {
		cb.Send(&edgeproto.Result{Message: err.Error()})
	}

	if err != nil {
		cb.Send(&edgeproto.Result{Message: "Deleting Cloudlet due to failures"})
		undoErr := s.deleteCloudletInternal(cctx.WithUndo(), in, pfConfig, cb)
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Undo create Cloudlet", "undoErr", undoErr)
		}
	}
	if deleteAccessVars {
		err1 := cloudletPlatform.DeleteCloudletAccessVars(ctx, in, pfConfig, updatecb.cb)
		if err1 != nil {
			cb.Send(&edgeproto.Result{Message: err1.Error()})
		}
	}
	return err
}

func isVersionConflict(ctx context.Context, localVersion, remoteVersion string) bool {
	if localVersion == "" {
		log.SpanLog(ctx, log.DebugLevelApi, "Ignoring cloudlet validity check as local cloudlet version is missing")
	} else if remoteVersion == "" {
		log.SpanLog(ctx, log.DebugLevelApi, "Ignoring cloudlet validity check as remote cloudlet version is missing")
	} else if localVersion != remoteVersion {
		log.SpanLog(ctx, log.DebugLevelApi, "Ignoring cloudlet info from old cloudlet", "localVersion", localVersion, "remoteVersion", remoteVersion)
		return true
	}
	return false
}

func (s *CloudletApi) UpdateCloudletState(ctx context.Context, key *edgeproto.CloudletKey, newState edgeproto.TrackedState) error {
	cloudlet := edgeproto.Cloudlet{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &cloudlet) {
			return key.NotFoundError()
		}
		cloudlet.State = newState
		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	return err
}

func (s *CloudletApi) WaitForCloudlet(ctx context.Context, key *edgeproto.CloudletKey, errorState edgeproto.TrackedState, successMsg string, timeout time.Duration, updateCallback edgeproto.CacheUpdateCallback) error {
	lastStatusId := uint32(0)
	done := make(chan bool, 1)
	failed := make(chan bool, 1)
	fatal := make(chan bool, 1)

	var err error

	go func() {
		err := cloudcommon.CrmServiceWait(*key)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "failed to cleanup crm service", "err", err)
			fatal <- true
		}
	}()

	log.SpanLog(ctx, log.DebugLevelApi, "wait for cloudlet state", "key", key, "errorState", errorState)

	info := edgeproto.CloudletInfo{}
	if cloudletInfoApi.cache.Get(key, &info) {
		lastStatusId = info.Status.TaskNumber
	}

	checkState := func(key *edgeproto.CloudletKey) {
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.cache.Get(key, &cloudlet) {
			return
		}
		cloudletInfo := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.cache.Get(key, &cloudletInfo) {
			return
		}

		curState := cloudletInfo.State
		localVersion := cloudlet.ContainerVersion
		remoteVersion := cloudletInfo.ContainerVersion

		if curState == edgeproto.CloudletState_CLOUDLET_STATE_ERRORS {
			failed <- true
			return
		}

		if !isVersionConflict(ctx, localVersion, remoteVersion) {
			if curState == edgeproto.CloudletState_CLOUDLET_STATE_READY &&
				(cloudlet.State != edgeproto.TrackedState_UPDATE_REQUESTED) {
				done <- true
			}
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "watch event for CloudletInfo")
	cancel := cloudletInfoApi.cache.WatchKey(key, func(ctx context.Context) {
		if !cloudletInfoApi.cache.Get(key, &info) {
			return
		}

		if info.Status.TaskNumber == 0 {
			lastStatusId = 0
		} else if lastStatusId == info.Status.TaskNumber {
			// got repeat message
			lastStatusId = info.Status.TaskNumber

		} else if info.Status.TaskNumber > lastStatusId {
			if info.Status.StepName != "" {
				updateCallback(edgeproto.UpdateTask, info.Status.StepName)
			} else {
				updateCallback(edgeproto.UpdateTask, info.Status.TaskName)
			}
			lastStatusId = info.Status.TaskNumber
		}
		checkState(key)
	})

	// After setting up watch, check current state,
	// as it may have already changed to target state
	checkState(key)

	select {
	case <-done:
		err = nil
		if successMsg != "" {
			updateCallback(edgeproto.UpdateTask, successMsg)
		}
	case <-failed:
		if cloudletInfoApi.cache.Get(key, &info) {
			errs := strings.Join(info.Errors, ", ")
			err = fmt.Errorf("Encountered failures: %s", errs)
		} else {
			err = fmt.Errorf("Unknown failure")
		}
	case <-fatal:
		out := ""
		out, err = cloudcommon.GetCloudletLog(ctx, key)
		if err != nil || out == "" {
			out = fmt.Sprintf("Please look at %s for more details", cloudcommon.GetCloudletLogFile(key.Name))
		} else {
			out = fmt.Sprintf("Failure: %s", out)
		}
		updateCallback(edgeproto.UpdateTask, out)
		err = errors.New(out)
	case <-time.After(timeout):
		err = fmt.Errorf("Timed out waiting for cloudlet state to be Ready")
		updateCallback(edgeproto.UpdateTask, "platform bringup timed out")
	}

	cancel()
	// note: do not close done/failed, garbage collector will deal with it.

	cloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &cloudlet) {
			return key.NotFoundError()
		}
		cloudlet.AccessVars = nil
		if err == nil {
			cloudlet.Errors = nil
			cloudlet.State = edgeproto.TrackedState_READY
		} else {
			cloudlet.Errors = []string{err.Error()}
			cloudlet.State = errorState
		}

		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	if err1 != nil {
		return err1
	}

	return err
}

func (s *CloudletApi) UpdateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	ctx := cb.Context()

	err := in.ValidateUpdateFields()
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
	if !cloudletApi.cache.Get(&in.Key, cur) {
		return in.Key.NotFoundError()
	}

	accessVars := make(map[string]string)
	if _, found := fmap[edgeproto.CloudletFieldAccessVars]; found {
		if in.DeploymentLocal {
			return errors.New("Access vars is not supported for local deployment")
		}
		accessVars = in.AccessVars
		in.AccessVars = nil
	}

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	if !ignoreCRMState(cctx) {
		updatecb := updateCloudletCallback{in, cb}
		var cloudletPlatform pf.Platform
		cloudletPlatform, err := pfutils.GetPlatform(ctx, in.PlatformType.String())
		if err != nil {
			return err
		}
		pfConfig, err := getPlatformConfig(ctx, in)
		if err != nil {
			return err
		}
		if len(accessVars) > 0 {
			err = cloudletPlatform.SaveCloudletAccessVars(ctx, in, accessVars, pfConfig, updatecb.cb)
			if err != nil {
				return err
			}
		}
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := &edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, cur) {
			return in.Key.NotFoundError()
		}
		cur.CopyInFields(in)
		s.store.STMPut(stm, cur)
		return nil
	})

	if err != nil {
		return err
	}

	// after the cloudlet change is committed, if the location changed,
	// update app insts as well.
	s.UpdateAppInstLocations(ctx, in)

	maintenance := edgeproto.MaintenanceState_NORMAL_OPERATION
	if _, found := fmap[edgeproto.CloudletFieldMaintenanceState]; found {
		maintenance = in.MaintenanceState
	}
	if maintenance == edgeproto.MaintenanceState_MAINTENANCE_START {
		// This is a state machine to transition into cloudlet
		// maintenance. Start by triggering AutoProv failovers.
		log.SpanLog(ctx, log.DebugLevelApi, "Start AutoProv failover")
		timeout := settingsApi.Get().CloudletMaintenanceTimeout.TimeDuration()
		err := cb.Send(&edgeproto.Result{
			Message: "Starting AutoProv failover",
		})
		if err != nil {
			return err
		}
		// first reset any old AutoProvInfo
		autoProvInfo := edgeproto.AutoProvInfo{
			Key:              in.Key,
			MaintenanceState: edgeproto.MaintenanceState_NORMAL_OPERATION,
		}
		autoProvInfoApi.Update(ctx, &autoProvInfo, 0)

		err = s.setMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_FAILOVER_REQUESTED)
		if err != nil {
			return err
		}
		err = autoProvInfoApi.waitForMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_FAILOVER_DONE, edgeproto.MaintenanceState_FAILOVER_ERROR, timeout, &autoProvInfo)
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
			undoErr := s.setMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_NORMAL_OPERATION)
			log.SpanLog(ctx, log.DebugLevelApi, "AutoProv maintenance failures", "err", err, "undoErr", undoErr)
			return fmt.Errorf("AutoProv failover encountered some errors, aborting maintenance")
		}
		cb.Send(&edgeproto.Result{
			Message: "AutoProv failover completed",
		})

		log.SpanLog(ctx, log.DebugLevelApi, "AutoProv failover complete")
		// proceed to next state
		maintenance = edgeproto.MaintenanceState_MAINTENANCE_START_NO_FAILOVER
	}
	if maintenance == edgeproto.MaintenanceState_MAINTENANCE_START_NO_FAILOVER {
		log.SpanLog(ctx, log.DebugLevelApi, "Start CRM failover")
		cb.Send(&edgeproto.Result{
			Message: "Starting CRM failover",
		})
		if !ignoreCRMState(cctx) {
			timeout := settingsApi.Get().CloudletMaintenanceTimeout.TimeDuration()
			// Tell CRM to go into maintenance mode
			err = s.setMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_CRM_REQUESTED)
			if err != nil {
				return err
			}
			cloudletInfo := edgeproto.CloudletInfo{}
			err = cloudletInfoApi.waitForMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_CRM_UNDER_MAINTENANCE, edgeproto.MaintenanceState_CRM_ERROR, timeout, &cloudletInfo)
			if err != nil {
				return err
			}
			if cloudletInfo.MaintenanceState == edgeproto.MaintenanceState_CRM_ERROR {
				undoErr := s.setMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_NORMAL_OPERATION)
				log.SpanLog(ctx, log.DebugLevelApi, "AutoProv maintenance failures", "err", err, "undoErr", undoErr)
				return fmt.Errorf("CRM encountered some errors, aborting maintenance")
			}
		}
		cb.Send(&edgeproto.Result{
			Message: "CRM failover complete",
		})
		log.SpanLog(ctx, log.DebugLevelApi, "CRM failover complete")
		// transition to maintenance
		err = s.setMaintenanceState(ctx, &in.Key, edgeproto.MaintenanceState_UNDER_MAINTENANCE)
		if err != nil {
			return err
		}
		cb.Send(&edgeproto.Result{
			Message: "Cloudlet is in maintenance",
		})
	}
	cb.Send(&edgeproto.Result{
		Message: "Cloudlet updated successfully",
	})
	return nil
}

func (s *CloudletApi) setMaintenanceState(ctx context.Context, key *edgeproto.CloudletKey, state edgeproto.MaintenanceState) error {
	return s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := &edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, cur) {
			return key.NotFoundError()
		}
		if cur.MaintenanceState == state {
			return nil
		}
		cur.MaintenanceState = state
		s.store.STMPut(stm, cur)
		return nil
	})
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	pfConfig, err := getPlatformConfig(cb.Context(), in)
	if err != nil {
		return err
	}
	return s.deleteCloudletInternal(DefCallContext(), in, pfConfig, cb)
}

func (s *CloudletApi) deleteCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, cb edgeproto.CloudletApi_DeleteCloudletServer) (reterr error) {
	ctx := cb.Context()

	defer func() {
		if reterr == nil {
			RecordCloudletEvent(ctx, &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
		}
	}()

	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesCloudlet(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return errors.New("Cloudlet in use by static AppInst")
	}

	clDynInsts := make(map[edgeproto.ClusterInstKey]struct{})
	if clusterInstApi.UsesCloudlet(&in.Key, clDynInsts) {
		return errors.New("Cloudlet in use by static ClusterInst")
	}

	cctx.SetOverride(&in.CrmOverride)

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		if ignoreCRMState(cctx) {
			// delete happens later, this STM just checks for existence
			return nil
		}
		if !cctx.Undo {
			if in.State == edgeproto.TrackedState_CREATE_REQUESTED ||
				in.State == edgeproto.TrackedState_CREATING ||
				in.State == edgeproto.TrackedState_UPDATE_REQUESTED ||
				in.State == edgeproto.TrackedState_UPDATING {
				return errors.New("Cloudlet busy, cannot be deleted")
			}
			if in.State == edgeproto.TrackedState_DELETE_ERROR &&
				cctx.Override != edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateCloudlet to rebuild, and try again"})
			}
			if in.State == edgeproto.TrackedState_DELETE_REQUESTED ||
				in.State == edgeproto.TrackedState_DELETING ||
				in.State == edgeproto.TrackedState_DELETE_PREPARE {
				return errors.New("Cloudlet busy, already under deletion")
			}
		}
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	if !ignoreCRMState(cctx) {
		updatecb := updateCloudletCallback{in, cb}

		if in.DeploymentLocal {
			updatecb.cb(edgeproto.UpdateTask, "Stopping CRMServer")
			err = cloudcommon.StopCRMService(ctx, in)
		} else {
			var cloudletPlatform pf.Platform
			cloudletPlatform, err = pfutils.GetPlatform(ctx, in.PlatformType.String())
			if err == nil {
				err = cloudletPlatform.DeleteCloudlet(ctx, in, pfConfig, updatecb.cb)
			}
		}
		if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete Cloudlet ignoring CRM failure: %s", err.Error())})
			s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_NOT_PRESENT)
			err = nil
		}
	}

	updateCloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &updateCloudlet) {
			return in.Key.NotFoundError()
		}
		if err != nil {
			updateCloudlet.State = edgeproto.TrackedState_DELETE_ERROR
			s.store.STMPut(stm, &updateCloudlet)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		cloudletRefsApi.store.STMDel(stm, &in.Key)
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
	if len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			derr := appInstApi.deleteAppInstInternal(DefCallContext(), &appInst, cb)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"key", key, "err", derr)
			}
		}
	}
	if len(clDynInsts) > 0 {
		for key, _ := range clDynInsts {
			clInst := edgeproto.ClusterInst{Key: key}
			derr := clusterInstApi.deleteClusterInstInternal(DefCallContext(), &clInst, cb)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic ClusterInst",
					"key", key, "err", derr)
			}
		}
	}
	cloudletPoolMemberApi.cloudletDeleted(ctx, &in.Key)
	cloudletInfoApi.cleanupCloudletInfo(ctx, &in.Key)
	return nil
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
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
		} else {
			if cl.ResTagMap == nil {
				cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
			}
		}

		return err
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	for resource, tblname := range in.Mapping {
		if valerr, ok := resTagTableApi.ValidateResName(resource); !ok {
			return &edgeproto.Result{}, valerr
		}
		resource = strings.ToLower(resource)
		var key edgeproto.ResTagTableKey
		key.Name = tblname
		key.Organization = in.Key.Organization
		tbl, err := resTagTableApi.GetResTagTable(ctx, &key)

		if err != nil && err.Error() == key.NotFoundError().Error() {
			// auto-create empty
			tbl.Key = key
			_, err = resTagTableApi.CreateResTagTable(ctx, tbl)
			if err != nil {
				return &edgeproto.Result{}, err
			}
		}
		cl.ResTagMap[resource] = &key
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		for resource, tblname := range in.Mapping {
			key := edgeproto.ResTagTableKey{
				Name:         tblname,
				Organization: in.Key.Organization,
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
	appInstApi.cache.Mux.Lock()
	for _, data := range appInstApi.cache.Objs {
		inst := data.Obj
		if inst.Key.ClusterInstKey.CloudletKey.Matches(&in.Key) {
			keys = append(keys, inst.Key)
		}
	}
	appInstApi.cache.Mux.Unlock()

	inst := edgeproto.AppInst{}
	for ii, _ := range keys {
		inst = *appInstApi.cache.Objs[keys[ii]].Obj
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

		err := appInstApi.updateAppInstStore(ctx, &inst)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "Update AppInst Location",
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

func (s *CloudletApi) FindFlavorMatch(ctx context.Context, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, error) {

	cl := edgeproto.Cloudlet{}
	var spec *vmspec.VMCreationSpec
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {

		if !cloudletApi.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		cli := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.store.STMGet(stm, &in.Key, &cli) {
			return in.Key.NotFoundError()
		}
		mexFlavor := edgeproto.Flavor{}
		mexFlavor.Key.Name = in.FlavorName
		if !flavorApi.store.STMGet(stm, &mexFlavor.Key, &mexFlavor) {
			return in.Key.NotFoundError()
		}
		var verr error
		spec, verr = resTagTableApi.GetVMSpec(ctx, stm, mexFlavor, cl, cli)
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

func (s *CloudletApi) GetCloudletManifest(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.CloudletManifest, error) {
	cloudlet := &edgeproto.Cloudlet{}
	if !cloudletApi.cache.Get(&in.Key, cloudlet) {
		return nil, in.Key.NotFoundError()
	}

	if cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
		return nil, fmt.Errorf("Not supported for PlatformTypeVmPool")
	}

	pfFlavor := edgeproto.Flavor{}
	if in.Flavor.Name == "" || in.Flavor.Name == DefaultPlatformFlavor.Key.Name {
		in.Flavor = DefaultPlatformFlavor.Key
		pfFlavor = DefaultPlatformFlavor
	} else {
		if !flavorApi.cache.Get(&in.Flavor, &pfFlavor) {
			return nil, in.Flavor.NotFoundError()
		}
	}

	pfConfig, err := getPlatformConfig(ctx, cloudlet)
	if err != nil {
		return nil, err
	}

	cloudletPlatform, err := pfutils.GetPlatform(ctx, cloudlet.PlatformType.String())
	if err != nil {
		return nil, err
	}

	return cloudletPlatform.GetCloudletManifest(ctx, cloudlet, pfConfig, &pfFlavor)
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
