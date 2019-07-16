package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
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

var cloudletApi = CloudletApi{}

func InitCloudletApi(sync *Sync) {
	cloudletApi.sync = sync
	cloudletApi.store = edgeproto.NewCloudletStore(sync.store)
	edgeproto.InitCloudletCache(&cloudletApi.cache)
	sync.RegisterCache(&cloudletApi.cache)
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.cache.GetAllKeys(keys)
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

func (s *CloudletApi) UsesPlatform(in *edgeproto.PlatformKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, _ := range s.cache.Objs {
		if s.cache.Objs[key].Platform.Matches(in) {
			return true
		}
	}
	return false
}

func (s *CloudletApi) ReplaceErrorState(in *edgeproto.Cloudlet, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
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

	return s.createCloudletInternal(DefCallContext(), in, cb)
}
func (s *CloudletApi) createCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	cctx.SetOverride(&in.CrmOverride)

	in.TimeLimits.CreateClusterInstTimeout = int64(cloudcommon.CreateClusterInstTimeout)
	in.TimeLimits.UpdateClusterInstTimeout = int64(cloudcommon.UpdateClusterInstTimeout)
	in.TimeLimits.DeleteClusterInstTimeout = int64(cloudcommon.DeleteClusterInstTimeout)
	in.TimeLimits.CreateAppInstTimeout = int64(cloudcommon.CreateAppInstTimeout)
	in.TimeLimits.UpdateAppInstTimeout = int64(cloudcommon.UpdateAppInstTimeout)
	in.TimeLimits.DeleteAppInstTimeout = int64(cloudcommon.DeleteAppInstTimeout)

	pf := edgeproto.Platform{}
	pfFlavor := edgeproto.Flavor{}

	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteCloudlet to remove and try again"})
				}
				return objstore.ErrKVStoreKeyExists
			}
			in.Errors = nil
		}
		if !platformApi.store.STMGet(stm, &in.Platform, &pf) {
			return fmt.Errorf("Platform %s not found", in.Platform.Name)
		}
		if !flavorApi.store.STMGet(stm, pf.Flavor, &pfFlavor) {
			return fmt.Errorf("Platform flavor %s not found", pf.Flavor.Name)
		}
		if ignoreCRM(cctx) {
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

	if ignoreCRM(cctx) {
		return nil
	}

	updateCloudletCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			log.DebugLog(log.DebugLevelApi, "SetStatusTask", "key", in.Key, "taskName", value)
			in.Status.SetTask(value)
			cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
		case edgeproto.UpdateStep:
			log.DebugLog(log.DebugLevelApi, "SetStatusStep", "key", in.Key, "stepName", value)
			in.Status.SetStep(value)
			cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
		}
	}

	cloudletPlatform, ok := cloudletPlatforms[pf.PlatformType]
	if !ok {
		return fmt.Errorf("Platform plugin %s not found", pf.PlatformType.String())
	}

	if in.DeploymentLocal {
		updateCloudletCallback(edgeproto.UpdateTask, "Starting CRMServer")
		err = cloudcommon.StartCRMService(in, &pf)
	} else {
		if !IsPlatformInternal(&pf) {
			if pf.ImagePath == "" {
				return fmt.Errorf("Platform must have imagepath specified")
			}
			if pf.RegistryPath == "" {
				return fmt.Errorf("Platform must have registrypath specified")
			}
		}
		err = cloudletPlatform.CreateCloudlet(in, &pf, &pfFlavor, updateCloudletCallback)
	}

	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create Cloudlet ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created Cloudlet successfully"})
		return nil
	}
	if err != nil {
		in.State = edgeproto.TrackedState_CREATE_ERROR
		in.Errors = append(in.Errors, err.Error())
		s.store.Put(in, s.sync.syncWait)

		cb.Send(&edgeproto.Result{Message: err.Error()})
		cb.Send(&edgeproto.Result{Message: "DELETING cloudlet due to failures"})

		undoErr := s.deleteCloudletInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo create cloudlet", "undoErr", undoErr)
		}
		return nil
	}

	// Wait for CRM to connect to controller
	var cloudletInfo edgeproto.CloudletInfo
	start := time.Now()
	timedout := false
	for {
		err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			if !cloudletInfoApi.store.STMGet(stm, &in.Key, &cloudletInfo) {
				return objstore.ErrKVStoreKeyNotFound
			}
			return nil
		})
		if err == nil {
			break
		}
		elapsed := time.Since(start)
		if elapsed >= (PlatformInitTimeout) {
			timedout = true
			break
		}
		// Wait till timeout
		time.Sleep(10 * time.Second)
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		updatedCloudlet := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, &updatedCloudlet) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if !cloudletInfoApi.store.STMGet(stm, &in.Key, &cloudletInfo) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if timedout {
			updatedCloudlet.State = edgeproto.TrackedState_CREATE_ERROR
			updatedCloudlet.Errors = append(updatedCloudlet.Errors, "platform bringup timed out")
		} else {
			if cloudletInfo.State == edgeproto.CloudletState_CLOUDLET_STATE_READY {
				updatedCloudlet.State = edgeproto.TrackedState_READY
				updateCloudletCallback(edgeproto.UpdateTask, "Cloudlet created successfully")
			} else {
				updatedCloudlet.State = edgeproto.TrackedState_CREATE_ERROR
				updatedCloudlet.Errors = append(updatedCloudlet.Errors, "cloudlet state is not ready: "+cloudletInfo.State.String())
			}
		}

		s.store.STMPut(stm, &updatedCloudlet)
		return nil
	})

	return err
}

func (s *CloudletApi) UpdateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
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
	if _, found := fmap[edgeproto.CloudletFieldLocationLatitude]; found {
		if in.Location.Latitude == 0 {
			return errors.New("Invalid latitude value of 0")
		}
	}
	if _, found := fmap[edgeproto.CloudletFieldLocationLongitude]; found {
		if in.Location.Longitude == 0 {
			return errors.New("Invalid longitude value of 0")
		}
	}

	_, err := s.store.Update(in, s.sync.syncWait)

	// after the cloudlet change is committed, if the location changed,
	// update app insts as well.
	s.UpdateAppInstLocations(in)
	return err
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	return s.deleteCloudletInternal(DefCallContext(), in, cb)
}

func (s *CloudletApi) deleteCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
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

	pf := edgeproto.Platform{}
	// Set state to prevent other apps from being created on ClusterInst
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if !cctx.Undo && in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR {
			if in.State == edgeproto.TrackedState_DELETE_ERROR {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateCloudlet to rebuild, and try again"})
			}
			return errors.New("Cloudlet busy, cannot delete")
		}
		if !platformApi.store.STMGet(stm, &in.Platform, &pf) {
			return fmt.Errorf("Delete failed, platform %s not found", in.Platform.Name)
		}
		if ignoreCRM(cctx) {
			s.store.STMDel(stm, &in.Key)
		}
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	cloudletPlatform, ok := cloudletPlatforms[pf.PlatformType]
	if !ok {
		return fmt.Errorf("Platform plugin %s not found", pf.PlatformType.String())
	}

	if in.DeploymentLocal {
		err = cloudcommon.StopCRMService(in, &pf)
	} else {
		err = cloudletPlatform.DeleteCloudlet(in, &pf)
	}
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete Cloudlet ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_NOT_PRESENT)
		cb.Send(&edgeproto.Result{Message: "Deleted Cloudlet successfully"})
		err = nil
	}

	if err != nil {
		return err
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		s.store.STMDel(stm, &in.Key)

		cloudletRefsApi.store.STMDel(stm, &in.Key)
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

func (s *CloudletApi) UpdateAppInstLocations(in *edgeproto.Cloudlet) {
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

	first := true
	inst := edgeproto.AppInst{}
	for ii, _ := range keys {
		inst.Key = keys[ii]
		if first {
			inst.Fields = make([]string, 0)
		}
		if _, found := fmap[edgeproto.CloudletFieldLocationLatitude]; found {
			inst.CloudletLoc.Latitude = in.Location.Latitude
			if first {
				inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLatitude)
			}
		}
		if _, found := fmap[edgeproto.CloudletFieldLocationLongitude]; found {
			inst.CloudletLoc.Longitude = in.Location.Longitude
			if first {
				inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLongitude)
			}
		}
		if first {
			if len(inst.Fields) == 0 {
				break
			}
			first = false
		}

		err := appInstApi.updateAppInstInternal(DefCallContext(), &inst, nil)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "Update AppInst Location",
				"inst", inst, "err", err)
		}
	}
}
