package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type ClusterInstApi struct {
	sync  *Sync
	store edgeproto.ClusterInstStore
	cache edgeproto.ClusterInstCache
}

var clusterInstApi = ClusterInstApi{}

// TODO: these timeouts should be adjust based on target platform,
// as some platforms (azure, etc) may take much longer.
// These timeouts should be at least long enough for the controller and
// CRM to exchange an initial set of messages (i.e. 10 sec or so).
var CreateClusterInstTimeout = 30 * time.Minute
var UpdateClusterInstTimeout = 20 * time.Minute
var DeleteClusterInstTimeout = 20 * time.Minute

// Transition states indicate states in which the CRM is still busy.
var CreateClusterInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_Creating: struct{}{},
}
var UpdateClusterInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_Updating: struct{}{},
}
var DeleteClusterInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_Deleting: struct{}{},
}

func InitClusterInstApi(sync *Sync) {
	clusterInstApi.sync = sync
	clusterInstApi.store = edgeproto.NewClusterInstStore(sync.store)
	edgeproto.InitClusterInstCache(&clusterInstApi.cache)
	sync.RegisterCache(&clusterInstApi.cache)
	if *shortTimeouts {
		CreateClusterInstTimeout = 3 * time.Second
		UpdateClusterInstTimeout = 2 * time.Second
		DeleteClusterInstTimeout = 2 * time.Second
	}
}

func (s *ClusterInstApi) HasKey(key *edgeproto.ClusterInstKey) bool {
	return s.cache.HasKey(key)
}

func (s *ClusterInstApi) Get(key *edgeproto.ClusterInstKey, buf *edgeproto.ClusterInst) bool {
	return s.cache.Get(key, buf)
}

func (s *ClusterInstApi) UsesClusterFlavor(key *edgeproto.ClusterFlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, cluster := range s.cache.Objs {
		if cluster.Flavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *ClusterInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.ClusterInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.CloudletKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LivenessStatic {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LivenessDynamic {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *ClusterInstApi) UsesCluster(key *edgeproto.ClusterKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, val := range s.cache.Objs {
		if val.Key.ClusterKey.Matches(key) {
			return true
		}
	}
	return false
}

func (s *ClusterInstApi) CreateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	in.Liveness = edgeproto.Liveness_LivenessStatic
	in.Auto = false
	return s.createClusterInstInternal(DefCallContext(), in, cb)
}

func GetClosestFlavor(info *edgeproto.CloudletInfo, nodeflavor edgeproto.Flavor) (string, error) {
	log.InfoLog("Get closest flavor available in Cloudlet")
	FlavorList := info.Flavors
	sort.Slice(FlavorList[:], func(i, j int) bool {
		if FlavorList[i].Vcpus < FlavorList[j].Vcpus {
			return true
		}
		if FlavorList[i].Vcpus > FlavorList[j].Vcpus {
			return false
		}
		if FlavorList[i].Ram < FlavorList[j].Ram {
			return true
		}
		if FlavorList[i].Ram > FlavorList[j].Ram {
			return false
		}
		return FlavorList[i].Disk < FlavorList[j].Disk
	})
	for _, flavor := range FlavorList {
		if flavor.Vcpus < nodeflavor.Vcpus {
			continue
		}
		if flavor.Ram < nodeflavor.Ram {
			continue
		}
		if flavor.Disk < nodeflavor.Disk {
			continue
		}
		log.InfoLog("Found closest flavor", "flavor", flavor)
		return flavor.Name, nil
	}
	return "", fmt.Errorf("no suitable platform flavor found for %s, please try a smaller flavor", nodeflavor.Key.Name)
}

// createClusterInstInternal is used to create dynamic cluster insts internally,
// bypassing static assignment. It is also used to create auto-cluster insts.
func (s *ClusterInstApi) createClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	cctx.SetOverride(&in.CrmOverride)
	if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
		return err
	}
	if in.IpAccess == edgeproto.IpAccess_IpAccessUnknown {
		// default to shared
		in.IpAccess = edgeproto.IpAccess_IpAccessShared
	}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if clusterInstApi.store.STMGet(stm, &in.Key, in) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DeleteError && !ignoreTransient(cctx, in.State) {
				if in.State == edgeproto.TrackedState_CreateError {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteClusterInst to remove and try again"})
				}
				return objstore.ErrKVStoreKeyExists
			}
			in.Errors = nil
		} else {
			err := in.Validate(edgeproto.ClusterInstAllFieldsMap)
			if err != nil {
				return err
			}
			if !in.Auto && strings.HasPrefix(in.Key.ClusterKey.Name, ClusterAutoPrefix) {
				return errors.New(ClusterAutoPrefixErr)
			}
		}
		if in.Liveness == edgeproto.Liveness_LivenessUnknown {
			in.Liveness = edgeproto.Liveness_LivenessDynamic
		}
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}
		info := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.store.STMGet(stm, &in.Key.CloudletKey, &info) {
			return fmt.Errorf("No resource information found for Cloudlet %s", in.Key.CloudletKey)
		}
		refs := edgeproto.CloudletRefs{}
		if !cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &refs) {
			initCloudletRefs(&refs, &in.Key.CloudletKey)
		}

		// cluster does not need to exist.
		// cluster will eventually be deprecated and removed.
		var cluster edgeproto.Cluster
		if clusterApi.store.STMGet(stm, &in.Key.ClusterKey, &cluster) {
			if in.Flavor.Name == "" {
				in.Flavor = cluster.DefaultFlavor
			}
		}
		if in.Flavor.Name == "" {
			return errors.New("No ClusterFlavor specified and no default ClusterFlavor for Cluster")
		}
		clusterFlavor := edgeproto.ClusterFlavor{}
		if !clusterFlavorApi.store.STMGet(stm, &in.Flavor, &clusterFlavor) {
			return fmt.Errorf("Cluster flavor %s not found", in.Flavor.Name)
		}
		nodeFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &clusterFlavor.NodeFlavor, &nodeFlavor) {
			return fmt.Errorf("Cluster flavor %s node flavor %s not found",
				in.Flavor.Name, clusterFlavor.NodeFlavor.Name)
		}
		masterFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &clusterFlavor.MasterFlavor, &masterFlavor) {
			return fmt.Errorf("Cluster flavor %s master flavor %s not found",
				in.Flavor.Name, clusterFlavor.MasterFlavor.Name)
		}
		var err error
		in.NodeFlavor, err = GetClosestFlavor(&info, nodeFlavor)
		if err != nil {
			return err
		}
		in.MasterFlavor, err = GetClosestFlavor(&info, masterFlavor)
		if err != nil {
			return err
		}
		log.InfoLog("Selected Cloudlet Flavor", "Node Flavor", in.NodeFlavor, "Master Flavor", in.MasterFlavor)

		// Do we allocate resources based on max nodes (no over-provisioning)?
		refs.UsedRam += nodeFlavor.Ram * uint64(clusterFlavor.MaxNodes)
		refs.UsedVcores += nodeFlavor.Vcpus * uint64(clusterFlavor.MaxNodes)
		refs.UsedDisk += nodeFlavor.Disk * uint64(clusterFlavor.MaxNodes)
		// XXX For now just track, don't enforce.
		if false {
			// XXX what is static overhead?
			var ramOverhead uint64 = 200
			var vcoresOverhead uint64 = 2
			var diskOverhead uint64 = 200
			// check resources
			if refs.UsedRam+ramOverhead > info.OsMaxRam {
				return errors.New("Not enough RAM available")
			}
			if refs.UsedVcores+vcoresOverhead > info.OsMaxVcores {
				return errors.New("Not enough Vcores available")
			}
			if refs.UsedDisk+diskOverhead > info.OsMaxVolGb {
				return errors.New("Not enough Disk available")
			}
		}
		// allocateIP also sets in.IpAccess to either Dedicated or Shared
		err = allocateIP(in, &cloudlet, &refs)
		if err != nil {
			return err
		}
		refs.Clusters = append(refs.Clusters, in.Key.ClusterKey)
		cloudletRefsApi.store.STMPut(stm, &refs)

		if ignoreCRM(cctx) {
			in.State = edgeproto.TrackedState_Ready
		} else {
			in.State = edgeproto.TrackedState_CreateRequested
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
	err = clusterInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_Ready, CreateClusterInstTransitions, edgeproto.TrackedState_CreateError, CreateClusterInstTimeout, "Created successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IgnoreCRMErrors {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create ClusterInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_Ready)
		cb.Send(&edgeproto.Result{Message: "Created ClusterInst successfully"})
		err = nil
	}
	if err != nil {
		// XXX should probably track mod revision ID and only undo
		// if no other changes were made to appInst in the meantime.
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "Deleting ClusterInst due to failures"})
		undoErr := s.deleteClusterInstInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo create clusterinst", "undoErr", undoErr)
		}
	}
	return err
}

func (s *ClusterInstApi) UpdateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_UpdateClusterInstServer) error {
	// Unsupported for now
	return errors.New("Update cluster instance not supported yet")
}

func (s *ClusterInstApi) DeleteClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	return s.deleteClusterInstInternal(DefCallContext(), in, cb)
}

func (s *ClusterInstApi) deleteClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	if appInstApi.UsesClusterInst(&in.Key) {
		return errors.New("ClusterInst in use by Application Instance")
	}
	if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
		return err
	}

	var prevState edgeproto.TrackedState
	// Set state to prevent other apps from being created on ClusterInst
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if !cctx.Undo && in.State != edgeproto.TrackedState_Ready && in.State != edgeproto.TrackedState_CreateError && in.State != edgeproto.TrackedState_DeletePrepare && !ignoreTransient(cctx, in.State) {
			if in.State == edgeproto.TrackedState_DeleteError {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateClusterInst to rebuild, and try again"})
			}
			return errors.New("ClusterInst busy, cannot delete")
		}
		prevState = in.State
		in.State = edgeproto.TrackedState_DeletePrepare
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	// Delete appInsts that are set for autodelete
	if err := appInstApi.AutoDeleteAppInsts(&in.Key, cb); err != nil {
		// restore previous state since we failed pre-delete actions
		in.State = prevState
		s.store.Update(in, s.sync.syncWait)
		return fmt.Errorf("Failed to auto-delete applications from clusterInst %s, %s",
			in.Key.ClusterKey.Name, err.Error())
	}

	cctx.SetOverride(&in.CrmOverride)
	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if in.State != edgeproto.TrackedState_DeletePrepare {
			return errors.New("ClusterInst expected state DeletePrepare")
		}

		clusterFlavor := edgeproto.ClusterFlavor{}
		nodeFlavor := edgeproto.Flavor{}
		if !clusterFlavorApi.store.STMGet(stm, &in.Flavor, &clusterFlavor) {
			log.WarnLog("Delete cluster inst: cluster flavor not found",
				"clusterflavor", in.Flavor.Name)
		} else {
			if !flavorApi.store.STMGet(stm, &clusterFlavor.NodeFlavor, &nodeFlavor) {
				log.WarnLog("Delete cluster inst: node flavor not found",
					"flavor", clusterFlavor.NodeFlavor.Name)
			}
		}
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
			log.WarnLog("Delete cluster inst: cloudlet not found",
				"cloudlet", in.Key.CloudletKey)
		}
		refs := edgeproto.CloudletRefs{}
		if cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &refs) {
			ii := 0
			for ; ii < len(refs.Clusters); ii++ {
				if refs.Clusters[ii].Matches(&in.Key.ClusterKey) {
					break
				}
			}
			if ii < len(refs.Clusters) {
				// explicity zero out deleted item to
				// prevent memory leak
				a := refs.Clusters
				copy(a[ii:], a[ii+1:])
				a[len(a)-1] = edgeproto.ClusterKey{}
				refs.Clusters = a[:len(a)-1]
			}
			// remove used resources
			refs.UsedRam -= nodeFlavor.Ram * uint64(clusterFlavor.MaxNodes)
			refs.UsedVcores -= nodeFlavor.Vcpus * uint64(clusterFlavor.MaxNodes)
			refs.UsedDisk -= nodeFlavor.Disk * uint64(clusterFlavor.MaxNodes)
			freeIP(in, &cloudlet, &refs)

			cloudletRefsApi.store.STMPut(stm, &refs)
		}
		if ignoreCRM(cctx) {
			// CRM state should be the same as before the
			// operation failed, so just need to clean up
			// controller state.
			s.store.STMDel(stm, &in.Key)
		} else {
			in.State = edgeproto.TrackedState_DeleteRequested
			s.store.STMPut(stm, in)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) {
		return nil
	}
	err = clusterInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_NotPresent, DeleteClusterInstTransitions, edgeproto.TrackedState_DeleteError, DeleteClusterInstTimeout, "Deleted ClusterInst successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IgnoreCRMErrors {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete ClusterInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_NotPresent)
		cb.Send(&edgeproto.Result{Message: "Deleted ClusterInst successfully"})
		err = nil
	}
	if err != nil {
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "Recreating ClusterInst due to failure"})
		undoErr := s.createClusterInstInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo delete clusterinst", "undoErr", undoErr)
		}
	}
	return err
}

func (s *ClusterInstApi) ShowClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_ShowClusterInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

// crmTransitionOk checks that the next state received from the CRM is a
// valid transition from the current state.
// See state_transitions.md
func crmTransitionOk(cur edgeproto.TrackedState, next edgeproto.TrackedState) bool {
	switch cur {
	case edgeproto.TrackedState_CreateRequested:
		if next == edgeproto.TrackedState_Creating || next == edgeproto.TrackedState_Ready || next == edgeproto.TrackedState_CreateError {
			return true
		}
	case edgeproto.TrackedState_Creating:
		if next == edgeproto.TrackedState_Ready || next == edgeproto.TrackedState_CreateError {
			return true
		}
	case edgeproto.TrackedState_UpdateRequested:
		if next == edgeproto.TrackedState_Updating || next == edgeproto.TrackedState_Ready || next == edgeproto.TrackedState_UpdateError {
			return true
		}
	case edgeproto.TrackedState_Updating:
		if next == edgeproto.TrackedState_Ready || next == edgeproto.TrackedState_UpdateError {
			return true
		}
	case edgeproto.TrackedState_DeleteRequested:
		if next == edgeproto.TrackedState_Deleting || next == edgeproto.TrackedState_NotPresent || next == edgeproto.TrackedState_DeleteError {
			return true
		}
	case edgeproto.TrackedState_Deleting:
		if next == edgeproto.TrackedState_NotPresent || next == edgeproto.TrackedState_DeleteError {
			return true
		}
	}
	return false
}

func ignoreTransient(cctx *CallContext, state edgeproto.TrackedState) bool {
	if cctx.Override == edgeproto.CRMOverride_IgnoreTransientState ||
		cctx.Override == edgeproto.CRMOverride_IgnoreCRMandTransientState {
		if state == edgeproto.TrackedState_Creating ||
			state == edgeproto.TrackedState_CreateRequested ||
			state == edgeproto.TrackedState_UpdateRequested ||
			state == edgeproto.TrackedState_DeleteRequested ||
			state == edgeproto.TrackedState_Updating ||
			state == edgeproto.TrackedState_Deleting ||
			state == edgeproto.TrackedState_DeletePrepare {
			return true
		}
	}
	return false
}

func ignoreCRM(cctx *CallContext) bool {
	if cctx.Undo || cctx.Override == edgeproto.CRMOverride_IgnoreCRM ||
		cctx.Override == edgeproto.CRMOverride_IgnoreCRMandTransientState {
		return true
	}
	return false
}

func (s *ClusterInstApi) UpdateFromInfo(in *edgeproto.ClusterInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Update ClusterInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State == in.State {
			// already in that state
			return nil
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		if in.State == edgeproto.TrackedState_CreateError || in.State == edgeproto.TrackedState_DeleteError || in.State == edgeproto.TrackedState_UpdateError {
			inst.Errors = in.Errors
		} else {
			inst.Errors = nil
		}
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *ClusterInstApi) DeleteFromInfo(in *edgeproto.ClusterInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Delete ClusterInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_Deleting && inst.State != edgeproto.TrackedState_DeleteRequested {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State,
				"next", edgeproto.TrackedState_NotPresent)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
}

func (s *ClusterInstApi) ReplaceErrorState(in *edgeproto.ClusterInst, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}

		if inst.State != edgeproto.TrackedState_CreateError &&
			inst.State != edgeproto.TrackedState_DeleteError &&
			inst.State != edgeproto.TrackedState_UpdateError {
			return nil
		}
		if newState == edgeproto.TrackedState_NotPresent {
			s.store.STMDel(stm, &in.Key)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}
