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
	PlatformInitTimeout   = 5 * time.Minute
	CloudletShortWaitTime = 10 * time.Millisecond
	CloudletWaitTime      = 1 * time.Second
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
		in.NotifySrvAddr = "127.0.0.1:51001"
	}

	if in.PhysicalName == "" {
		in.PhysicalName = in.Key.Name
	}

	pfConfig := edgeproto.PlatformConfig{}
	appRoles := VaultRoles{}
	if err := getRolesAndSecrets(&appRoles); err != nil {
		if !*testMode {
			return err
		}
		log.DebugLog(log.DebugLevelApi, "Warning, Failed to get roleIDs - running locally",
			"err", err)
	} else {
		pfConfig.CrmRoleId = appRoles.CRMRoleID
		pfConfig.CrmSecretId = appRoles.CRMSecretID
	}
	pfConfig.PlatformTag = *versionTag
	pfConfig.TlsCertFile = *tlsCertFile
	pfConfig.VaultAddr = *vaultAddr
	pfConfig.RegistryPath = *cloudletRegistryPath
	pfConfig.ImagePath = *cloudletVMImagePath
	pfConfig.TestMode = *testMode
	addrObjs := strings.Split(*notifyAddr, ":")
	if len(addrObjs) != 2 {
		return fmt.Errorf("unable to fetch notify addr of the controller")
	}
	pfConfig.NotifyCtrlAddrs = *publicAddr + ":" + addrObjs[1]

	return s.createCloudletInternal(DefCallContext(), in, &pfConfig, cb)
}

func (s *CloudletApi) createCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	cctx.SetOverride(&in.CrmOverride)
	ctx := cb.Context()

	pfConfig.Span = log.SpanToString(ctx)

	in.TimeLimits.CreateClusterInstTimeout = int64(cloudcommon.CreateClusterInstTimeout)
	in.TimeLimits.UpdateClusterInstTimeout = int64(cloudcommon.UpdateClusterInstTimeout)
	in.TimeLimits.DeleteClusterInstTimeout = int64(cloudcommon.DeleteClusterInstTimeout)
	in.TimeLimits.CreateAppInstTimeout = int64(cloudcommon.CreateAppInstTimeout)
	in.TimeLimits.UpdateAppInstTimeout = int64(cloudcommon.UpdateAppInstTimeout)
	in.TimeLimits.DeleteAppInstTimeout = int64(cloudcommon.DeleteAppInstTimeout)

	pfFlavor := edgeproto.Flavor{}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
		if in.Flavor.Name != "" {
			if !flavorApi.store.STMGet(stm, &in.Flavor, &pfFlavor) {
				return fmt.Errorf("Platform flavor %s not found", in.Flavor.Name)
			}
		} else {
			in.Flavor = DefaultPlatformFlavor.Key
			pfFlavor = DefaultPlatformFlavor
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
		err = WaitForCloudlet(ctx, &in.Key, PlatformInitTimeout, updatecb.cb)
	} else {
		cb.Send(&edgeproto.Result{Message: err.Error()})
	}

	updatedCloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &updatedCloudlet) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if err == nil {
			updatedCloudlet.State = edgeproto.TrackedState_READY
		} else {
			updatedCloudlet.Errors = append(updatedCloudlet.Errors, err.Error())
			updatedCloudlet.State = edgeproto.TrackedState_CREATE_ERROR
		}

		s.store.STMPut(stm, &updatedCloudlet)
		return nil
	})

	if err1 != nil {
		return err1
	}

	if err != nil {
		cb.Send(&edgeproto.Result{Message: "DELETING cloudlet due to failures"})
		undoErr := s.deleteCloudletInternal(cctx.WithUndo(), &updatedCloudlet, pfConfig, cb)
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Undo create cloudlet", "undoErr", undoErr)
		}
	}
	return nil
}

func WaitForCloudlet(ctx context.Context, key *edgeproto.CloudletKey, timeout time.Duration, updateCallback edgeproto.CacheUpdateCallback) error {
	curState := edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN
	lastStatusId := uint32(0)
	done := make(chan bool, 1)
	failed := make(chan bool, 1)
	fatal := make(chan bool, 1)

	var err error

	go func() {
		err := cloudcommon.CrmServiceWait(*key)
		if err != nil {
			fatal <- true
		}
	}()

	cancel := cloudletInfoApi.cache.WatchKey(key, func(ctx context.Context) {
		info := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.cache.Get(key, &info) {
			return
		}
		log.SpanLog(ctx, log.DebugLevelApi, "watch event for CloudletInfo")
		curState = info.State
		if info.Status.TaskNumber != 0 &&
			info.Status.TaskNumber != lastStatusId {
			if info.Status.StepName != "" {
				updateCallback(edgeproto.UpdateTask, info.Status.StepName)
			} else {
				updateCallback(edgeproto.UpdateTask, info.Status.TaskName)
			}
			lastStatusId = info.Status.TaskNumber
		}
		if curState == edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN {
			// Complete details is yet to arrive
			return
		}
		if curState == edgeproto.CloudletState_CLOUDLET_STATE_READY {
			done <- true
		} else {
			failed <- true
		}
	})
	// After setting up watch, check current state,
	// as it may have already changed to target state
	info := edgeproto.CloudletInfo{}
	if cloudletInfoApi.cache.Get(key, &info) {
		curState = info.State
		if curState != edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN {
			if curState == edgeproto.CloudletState_CLOUDLET_STATE_READY {
				done <- true
			} else {
				failed <- true
			}
		}
	}

	select {
	case <-done:
		err = nil
		updateCallback(edgeproto.UpdateTask, "Cloudlet created successfully")
	case <-failed:
		err = fmt.Errorf("Encountered failures")
		if cloudletInfoApi.cache.Get(key, &info) {
			updateCallback(edgeproto.UpdateTask, "cloudlet state is not ready: "+info.State.String())
		} else {
			updateCallback(edgeproto.UpdateTask, "unable to fetch cloudlet info")
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
		err = fmt.Errorf("Timedout")
		updateCallback(edgeproto.UpdateTask, "platform bringup timed out")
	}
	cancel()
	// note: do not close done/failed, garbage collector will deal with it.
	return err
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

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := &edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, cur) {
			return objstore.ErrKVStoreKeyNotFound
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
	return err
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	pfConfig := edgeproto.PlatformConfig{}
	pfConfig.VaultAddr = *vaultAddr
	return s.deleteCloudletInternal(DefCallContext(), in, &pfConfig, cb)
}

func (s *CloudletApi) deleteCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	ctx := cb.Context()
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesCloudlet(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return errors.New("Cloudlet in use by static Application Instance")
	}

	clDynInsts := make(map[edgeproto.ClusterInstKey]struct{})
	if clusterInstApi.UsesCloudlet(&in.Key, clDynInsts) {
		return errors.New("Cloudlet in use by static Cluster Instance")
	}

	cctx.SetOverride(&in.CrmOverride)

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if ignoreCRMState(cctx) {
			s.store.STMDel(stm, &in.Key)
			cloudletRefsApi.store.STMDel(stm, &in.Key)
			return nil
		}
		if !cctx.Undo {
			if in.State == edgeproto.TrackedState_DELETE_ERROR {
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

	if ignoreCRMState(cctx) {
		return nil
	}

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

	if err != nil {
		return err
	}

	cb.Send(&edgeproto.Result{Message: "Deleted Cloudlet successfully"})

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		s.store.STMDel(stm, &in.Key)
		cloudletRefsApi.store.STMDel(stm, &in.Key)
		cloudletInfoApi.store.STMDel(stm, &in.Key)
		return nil
	})

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
					"Failed to delete dynamic cluster inst",
					"key", key, "err", derr)
			}
		}
	}
	return err
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
		return err
	})
	return err
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
