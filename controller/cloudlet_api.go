package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

const (
	DefaultBindPort   = 55091
	CRMBringupTimeout = 5 * time.Minute
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

	roleID := os.Getenv("VAULT_ROLE_ID")
	if roleID == "" {
		return fmt.Errorf("Env variable VAULT_ROLE_ID not set")
	}
	secretID := os.Getenv("VAULT_SECRET_ID")
	if secretID == "" {
		return fmt.Errorf("Env variable VAULT_SECRET_ID not set")
	}

	if in.BindPort < 1 {
		in.BindPort = DefaultBindPort
	}

	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}

		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	// Load platform implementation
	platform, err := pfutils.GetPlatform(in.Platform)
	if err != nil {
		return err
	}

	updateCloudletCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			setStatusTask(in, value, cb)
		case edgeproto.UpdateStep:
			setStatusStep(in, value, cb)
		}
	}

	// Create Cloudlet
	in.State = edgeproto.TrackedState_CREATE_REQUESTED
	err = platform.CreateCloudlet(in, updateCloudletCallback)
	if err != nil {
		in.State = edgeproto.TrackedState_CREATE_ERROR
		in.Errors = append(in.Errors, err.Error())
		err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			s.store.STMPut(stm, in)
			return nil
		})
		return err
	}

	// Wait for CRM to connect to controller
	var cloudletInfo edgeproto.CloudletInfo
	start := time.Now()
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
		if elapsed >= (CRMBringupTimeout) {
			in.State = edgeproto.TrackedState_CREATE_ERROR
			in.Errors = append(in.Errors, "crm bringup timed out")
			break
		}
		// Wait till timeout
		time.Sleep(10 * time.Second)
	}

	if in.State != edgeproto.TrackedState_CREATE_ERROR {
		in.State = edgeproto.TrackedState_READY
		cb.Send(&edgeproto.Result{Message: "Created Cloudlet successfully"})
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		s.store.STMPut(stm, in)
		return nil
	})

	in.TimeLimits.CreateClusterInstTimeout = int64(cloudcommon.CreateClusterInstTimeout)
	in.TimeLimits.UpdateClusterInstTimeout = int64(cloudcommon.UpdateClusterInstTimeout)
	in.TimeLimits.DeleteClusterInstTimeout = int64(cloudcommon.DeleteClusterInstTimeout)
	in.TimeLimits.CreateAppInstTimeout = int64(cloudcommon.CreateAppInstTimeout)
	in.TimeLimits.UpdateAppInstTimeout = int64(cloudcommon.UpdateAppInstTimeout)
	in.TimeLimits.DeleteAppInstTimeout = int64(cloudcommon.DeleteAppInstTimeout)

	return nil
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
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesCloudlet(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return errors.New("Cloudlet in use by static Application Instance")
	}

	clDynInsts := make(map[edgeproto.ClusterInstKey]struct{})
	if clusterInstApi.UsesCloudlet(&in.Key, clDynInsts) {
		return errors.New("Cloudlet in use by static Cluster Instance")
	}

	// Set state to prevent other apps from being created on ClusterInst
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR && in.State != edgeproto.TrackedState_DELETE_PREPARE {
			if in.State == edgeproto.TrackedState_DELETE_ERROR {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateCloudlet to rebuild, and try again"})
			}
			return errors.New("Cloudlet busy, cannot delete")
		}
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		return nil
	})
	if err != nil {
		return err
	}

	// Load platform implementation
	platform, err := pfutils.GetPlatform(in.Platform)
	if err != nil {
		return err
	}

	err = platform.DeleteCloudlet(in)
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

func setStatusTask(in *edgeproto.Cloudlet, taskName string, cb edgeproto.CloudletApi_CreateCloudletServer) {
	log.DebugLog(log.DebugLevelApi, "SetStatusTask", "key", in.Key, "taskName", taskName)
	in.Status.SetTask(taskName)

	cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
}

func setStatusMaxTasks(in *edgeproto.Cloudlet, maxTasks uint32, cb edgeproto.CloudletApi_CreateCloudletServer) {
	log.DebugLog(log.DebugLevelApi, "SetStatusMaxTasks", "key", in.Key, "maxTasks", maxTasks)
	in.Status.SetMaxTasks(maxTasks)

	cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
}

func setStatusStep(in *edgeproto.Cloudlet, stepName string, cb edgeproto.CloudletApi_CreateCloudletServer) {
	log.DebugLog(log.DebugLevelApi, "SetStatusStep", "key", in.Key, "stepName", stepName)
	in.Status.SetStep(stepName)

	cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
}
