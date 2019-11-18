package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
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

func (s *CloudletApi) UsesOperator(in *edgeproto.OperatorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, _ := range s.cache.Objs {
		if key.OperatorKey.Matches(in) {
			return true
		}
	}
	return false
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

func getRolesAndSecrets(appRoles *VaultRoles) error {
	// Vault controller level credentials are required to access
	// registry credentials
	roleId := os.Getenv("VAULT_ROLE_ID")
	if roleId == "" {
		return fmt.Errorf("Env variable VAULT_ROLE_ID not set")
	}
	secretId := os.Getenv("VAULT_SECRET_ID")
	if secretId == "" {
		return fmt.Errorf("Env variable VAULT_SECRET_ID not set")
	}

	// Vault CRM level credentials are required to access
	// instantiate crmserver
	crmRoleId := os.Getenv("VAULT_CRM_ROLE_ID")
	if crmRoleId == "" {
		return fmt.Errorf("Env variable VAULT_CRM_ROLE_ID not set")
	}
	crmSecretId := os.Getenv("VAULT_CRM_SECRET_ID")
	if crmSecretId == "" {
		return fmt.Errorf("Env variable VAULT_CRM_SECRET_ID not set")
	}
	appRoles.CtrlRoleID = roleId
	appRoles.CtrlSecretID = secretId
	appRoles.CRMRoleID = crmRoleId
	appRoles.CRMSecretID = crmSecretId
	//Once we integrate DME add dme roles to the same structure
	return nil
}

func getPlatformConfig(ctx context.Context, cloudlet *edgeproto.Cloudlet) (*edgeproto.PlatformConfig, error) {
	pfConfig := edgeproto.PlatformConfig{}
	appRoles := VaultRoles{}
	if err := getRolesAndSecrets(&appRoles); err != nil {
		if !*testMode {
			return nil, err
		}
		log.DebugLog(log.DebugLevelApi, "Warning, Failed to get roleIDs - running locally",
			"err", err)
	} else {
		pfConfig.CrmRoleId = appRoles.CRMRoleID
		pfConfig.CrmSecretId = appRoles.CRMSecretID
	}
	pfConfig.PlatformTag = cloudlet.Version
	pfConfig.TlsCertFile = *tlsCertFile
	pfConfig.VaultAddr = *vaultAddr
	pfConfig.RegistryPath = *cloudletRegistryPath
	pfConfig.ImagePath = *cloudletVMImagePath
	pfConfig.TestMode = *testMode
	addrObjs := strings.Split(*notifyAddr, ":")
	if len(addrObjs) != 2 {
		return nil, fmt.Errorf("unable to fetch notify addr of the controller")
	}
	pfConfig.NotifyCtrlAddrs = *publicAddr + ":" + addrObjs[1]
	pfConfig.Span = log.SpanToString(ctx)

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

	if in.Version == "" {
		in.Version = *versionTag
	}

	return s.createCloudletInternal(DefCallContext(), in, cb)
}

func (s *CloudletApi) createCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	cctx.SetOverride(&in.CrmOverride)
	ctx := cb.Context()

	pfConfig, err := getPlatformConfig(ctx, in)
	if err != nil {
		return err
	}

	in.TimeLimits.CreateClusterInstTimeout = int64(cloudcommon.CreateClusterInstTimeout)
	in.TimeLimits.UpdateClusterInstTimeout = int64(cloudcommon.UpdateClusterInstTimeout)
	in.TimeLimits.DeleteClusterInstTimeout = int64(cloudcommon.DeleteClusterInstTimeout)
	in.TimeLimits.CreateAppInstTimeout = int64(cloudcommon.CreateAppInstTimeout)
	in.TimeLimits.UpdateAppInstTimeout = int64(cloudcommon.UpdateAppInstTimeout)
	in.TimeLimits.DeleteAppInstTimeout = int64(cloudcommon.DeleteAppInstTimeout)

	pfFlavor := edgeproto.Flavor{}
	if in.Flavor.Name == "" {
		in.Flavor = DefaultPlatformFlavor.Key
		pfFlavor = DefaultPlatformFlavor
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			if !cctx.Undo {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteCloudlet to remove and try again"})
				}
				return objstore.ErrKVStoreKeyExists
			}
			in.Errors = nil
		}
		if in.Flavor.Name != "" && in.Flavor.Name != DefaultPlatformFlavor.Key.Name {
			if !flavorApi.store.STMGet(stm, &in.Flavor, &pfFlavor) {
				return fmt.Errorf("Platform Flavor %s not found", in.Flavor.Name)
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

	updatecb := updateCloudletCallback{in, cb}

	if in.DeploymentLocal {
		updatecb.cb(edgeproto.UpdateTask, "Starting CRMServer")
		err = cloudcommon.StartCRMService(ctx, in, pfConfig)
	} else {
		var cloudletPlatform pf.Platform
		cloudletPlatform, err = pfutils.GetPlatform(ctx, in.PlatformType.String())
		if err == nil {
			err = cloudletPlatform.CreateCloudlet(ctx, in, pfConfig, &pfFlavor, updatecb.cb)
		}
	}

	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create Cloudlet ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created Cloudlet successfully"})
		return nil
	}

	if err == nil {
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
			return objstore.ErrKVStoreKeyNotFound
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
		localVersion := cloudlet.Version
		remoteVersion := cloudletInfo.Version

		if curState == edgeproto.CloudletState_CLOUDLET_STATE_ERRORS {
			failed <- true
		}
		if !isVersionConflict(ctx, localVersion, remoteVersion) {
			if curState == edgeproto.CloudletState_CLOUDLET_STATE_READY {
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

	for {
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
		break
	}
	// note: do not close done/failed, garbage collector will deal with it.

	cloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &cloudlet) {
			return objstore.ErrKVStoreKeyNotFound
		}
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

func getCloudletVersion(key *edgeproto.CloudletKey) (string, error) {
	nodeApi.cache.Mux.Lock()
	defer nodeApi.cache.Mux.Unlock()
	for _, obj := range nodeApi.cache.Objs {
		if obj.Key.NodeType != edgeproto.NodeType_NODE_CRM {
			continue
		}
		if obj.Key.CloudletKey != *key {
			continue
		}
		return obj.ImageVersion, nil
	}
	return "", fmt.Errorf("Unable to find Cloudlet node")
}

func isCloudletUpgradeRequired(cloudlet *edgeproto.Cloudlet) error {
	cloudletVersion, err := getCloudletVersion(&cloudlet.Key)
	if err != nil {
		return fmt.Errorf("unable to fetch cloudlet version: %v", err)
	}

	if cloudletVersion == "" {
		return nil
	}

	ctrl_vers, err := util.VersionParse(*versionTag)
	if err != nil {
		return err
	}

	cloudlet_vers, err := util.VersionParse(cloudletVersion)
	if err != nil {
		return err
	}

	new_vers, err := util.VersionParse(cloudlet.Version)
	if err != nil {
		return err
	}

	diff := ctrl_vers.Sub(*new_vers)
	if diff > 0 {
		return fmt.Errorf("cannot upgrade cloudlet to a version below controller version %s", *versionTag)
	}

	diff = cloudlet_vers.Sub(*new_vers)
	if diff > 0 {
		return fmt.Errorf("downgrade from version %s to %s is not supported", cloudlet_vers, new_vers)
	} else if diff < 0 {
		return nil
	} else {
		// Allow users to retry upgrade to same version on an update error
		if cloudlet.State != edgeproto.TrackedState_UPDATE_ERROR {
			return fmt.Errorf("no upgrade required, cloudlet is already of version %s", cloudletVersion)
		}
		return nil
	}
}

func (s *CloudletApi) UpdateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	ctx := cb.Context()
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

	err := in.Validate(fmap)
	if err != nil {
		return err
	}

	cur := &edgeproto.Cloudlet{}
	if !cloudletApi.cache.Get(&in.Key, cur) {
		return objstore.ErrKVStoreKeyNotFound
	}
	cur.CopyInFields(in)

	upgrade := false
	if _, found := fmap[edgeproto.CloudletFieldVersion]; found {
		err = isCloudletUpgradeRequired(cur)
		if err != nil {
			return err
		}
		// verify if image is available in registry
		registry_path := *cloudletRegistryPath + ":" + in.Version
		err = cloudcommon.ValidateDockerRegistryPath(ctx, registry_path, vaultConfig)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to validate cloudlet registry path", "err", err)
			} else {
				return err
			}
		}
		upgrade = true
	}

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	if !ignoreCRMState(cctx) && upgrade {
		if in.DeploymentLocal {
			return fmt.Errorf("upgrade is not supported for local deployments")
		}
		err = s.UpgradeCloudlet(ctx, cur, cb)
		if err != nil {
			return err
		}
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, cur) {
			return objstore.ErrKVStoreKeyNotFound
		}
		cur.CopyInFields(in)
		// In case we need to set TrackedState to ready
		// maybe after a manual upgrade of cloudlet
		if ignoreCRMState(cctx) {
			cur.State = edgeproto.TrackedState_READY
		}
		s.store.STMPut(stm, cur)
		return nil
	})

	if err != nil {
		return err
	}
	cb.Send(&edgeproto.Result{Message: "Updated Cloudlet successfully"})

	// after the cloudlet change is committed, if the location changed,
	// update app insts as well.
	s.UpdateAppInstLocations(ctx, in)

	return nil
}

func (s *CloudletApi) UpgradeCloudlet(ctx context.Context, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	updatecb := updateCloudletCallback{in, cb}

	if appInstApi.UsingCloudlet(&in.Key) {
		return fmt.Errorf("AppInst is busy on the cloudlet")
	}
	if clusterInstApi.UsingCloudlet(&in.Key) {
		return fmt.Errorf("ClusterInst is busy on the cloudlet")
	}

	if err := cloudletInfoApi.checkCloudletReady(&in.Key); err != nil {
		if in.State == edgeproto.TrackedState_UPDATE_ERROR {
			info := edgeproto.CloudletInfo{}
			if cloudletInfoApi.cache.Get(&in.Key, &info) &&
				info.State == edgeproto.CloudletState_CLOUDLET_STATE_READY {
				// Allow upgrade even on UPDATE_ERROR, so that end-users can retry upgrade
			} else {
				return err
			}
		} else {
			return err
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "fetch platform config")
	pfConfig, err := getPlatformConfig(ctx, in)
	if err != nil {
		return err
	}
	// cleanup old crms post upgrade
	pfConfig.CleanupMode = true
	cloudlet := edgeproto.Cloudlet{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cloudlet) {
			return objstore.ErrKVStoreKeyNotFound
		}
		cloudlet.Config = *pfConfig
		cloudlet.Version = in.Version
		cloudlet.State = edgeproto.TrackedState_UPDATE_REQUESTED

		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	if err != nil {
		return err
	}

	updatecb.cb(edgeproto.UpdateTask, "Upgrading Cloudlet")

	// Wait for cloudlet to finish upgrading
	err = s.WaitForCloudlet(
		ctx, &in.Key,
		edgeproto.TrackedState_UPDATE_ERROR, // Set error state
		"Upgraded Cloudlet successfully",    // Set success message
		PlatformInitTimeout, updatecb.cb,
	)

	return err
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	pfConfig, err := getPlatformConfig(cb.Context(), in)
	if err != nil {
		return err
	}
	return s.deleteCloudletInternal(DefCallContext(), in, pfConfig, cb)
}

func (s *CloudletApi) deleteCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	ctx := cb.Context()
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
			return objstore.ErrKVStoreKeyNotFound
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
			return objstore.ErrKVStoreKeyNotFound
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
	return err
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
			return objstore.ErrKVStoreKeyNotFound
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
			return objstore.ErrKVStoreKeyNotFound
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
		key.OperatorKey = in.Key.OperatorKey
		tbl, err := resTagTableApi.GetResTagTable(ctx, &key)
		if err == objstore.ErrKVStoreKeyNotFound {
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
			return objstore.ErrKVStoreKeyNotFound
		}
		for resource, tblname := range in.Mapping {
			key := edgeproto.ResTagTableKey{
				Name:        tblname,
				OperatorKey: in.Key.OperatorKey,
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
	for _, inst := range appInstApi.cache.Objs {
		if inst.Key.ClusterInstKey.CloudletKey.Matches(&in.Key) {
			keys = append(keys, inst.Key)
		}
	}
	appInstApi.cache.Mux.Unlock()

	inst := edgeproto.AppInst{}
	for ii, _ := range keys {
		inst = *appInstApi.cache.Objs[keys[ii]]
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

	for key, obj := range s.cache.Objs {
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
