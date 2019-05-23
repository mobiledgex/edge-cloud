package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	flavor "github.com/mobiledgex/edge-cloud/flavor"

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

func (s *ClusterInstApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
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

// createClusterInstInternal is used to create dynamic cluster insts internally,
// bypassing static assignment. It is also used to create auto-cluster insts.
func (s *ClusterInstApi) createClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
	}
	if in.Key.Developer == "" {
		return fmt.Errorf("Developer cannot be empty")
	}
	if in.Key.CloudletKey.Name == "" {
		return fmt.Errorf("Cloudlet name cannot be empty")
	}
	if in.Key.CloudletKey.OperatorKey.Name == "" {
		return fmt.Errorf("Operator name cannot be empty")
	}
	if in.Key.ClusterKey.Name == "" {
		return fmt.Errorf("Cluster name cannot be empty")
	}
	// validate deployment
	if in.Deployment == "" {
		// assume kubernetes, because that's what we've been doing
		in.Deployment = cloudcommon.AppDeploymentTypeKubernetes
	}
	if in.Deployment == cloudcommon.AppDeploymentTypeHelm {
		// helm runs on kubernetes
		in.Deployment = cloudcommon.AppDeploymentTypeKubernetes
	}
	if in.Deployment == cloudcommon.AppDeploymentTypeVM {
		// friendly error message if they try to specify VM
		return fmt.Errorf("ClusterInst is not needed for deployment type %s, just create an AppInst directly", cloudcommon.AppDeploymentTypeVM)
	}

	// validate other parameters based on deployment type
	if in.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		// must have at least one master, but currently don't support
		// more than one.
		if in.NumMasters == 0 {
			// just set it to 1
			in.NumMasters = 1
		}
		if in.NumMasters > 1 {
			return fmt.Errorf("NumMasters cannot be greater than 1")
		}
		// TODO: support zero nodes
		if in.NumNodes == 0 {
			return fmt.Errorf("Zero NumNodes not supported yet")
		}
	} else if in.Deployment == cloudcommon.AppDeploymentTypeDocker {
		if in.NumMasters != 0 || in.NumNodes != 0 {
			return fmt.Errorf("NumMasters and NumNodes not applicable for deployment type %s", cloudcommon.AppDeploymentTypeDocker)
		}
		if in.IpAccess == edgeproto.IpAccess_IpAccessUnknown {
			// must be dedicated for docker
			in.IpAccess = edgeproto.IpAccess_IpAccessDedicated
		} else if in.IpAccess != edgeproto.IpAccess_IpAccessDedicated {
			return fmt.Errorf("IpAccess must be dedicated for deployment type %s", cloudcommon.AppDeploymentTypeDocker)
		}
	} else {
		return fmt.Errorf("Invalid deployment type %s for ClusterInst", in.Deployment)
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
			return errors.New("No Flavor specified and no default Flavor for Cluster")
		}

		nodeFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &in.Flavor, &nodeFlavor) {
			return fmt.Errorf("Cluster flavor %s node flavor %s not found",
				in.Flavor.Name, nodeFlavor.Key)
		}
		var err error
		in.NodeFlavor, err = flavor.GetClosestFlavor(info.Flavors, nodeFlavor)
		if err != nil {
			return err
		}
		log.InfoLog("Selected Cloudlet Flavor", "Node Flavor", in.NodeFlavor)
		// Do we allocate resources based on max nodes (no over-provisioning)?
		refs.UsedRam += nodeFlavor.Ram * uint64(in.NumNodes+in.NumMasters)
		refs.UsedVcores += nodeFlavor.Vcpus * uint64(in.NumNodes+in.NumMasters)
		refs.UsedDisk += nodeFlavor.Disk * uint64(in.NumNodes+in.NumMasters)
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

func (s *ClusterInstApi) DeleteClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	return s.deleteClusterInstInternal(DefCallContext(), in, cb)
}

func (s *ClusterInstApi) UpdateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_UpdateClusterInstServer) error {
	// Unsupported for now
	return s.updateClusterInstInternal(DefCallContext(), in, cb)
}

func (s *ClusterInstApi) updateClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	log.DebugLog(log.DebugLevelApi, "updateClusterInstInternal", "in", in)

	if in.Fields == nil {
		return fmt.Errorf("nothing specified to update")
	}
	allowedFields := []string{}
	badFields := []string{}
	for _, field := range in.Fields {
		if field == edgeproto.ClusterInstFieldCrmOverride ||
			field == edgeproto.ClusterInstFieldKey ||
			strings.HasPrefix(field, edgeproto.ClusterInstFieldKey+".") {
			// ignore. TODO: generate a func to check if field is a key field.
			continue
		} else if field == edgeproto.ClusterInstFieldNumNodes {
			allowedFields = append(allowedFields, field)
		} else {
			badFields = append(badFields, field)
		}
	}
	if len(badFields) > 0 {
		// TODO: generate func to convert field consts to string names.
		return fmt.Errorf("some specified fields cannot be modified")
	}
	in.Fields = allowedFields

	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
	}

	var inbuf edgeproto.ClusterInst
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &inbuf) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if inbuf.NumMasters == 0 {
			return fmt.Errorf("cannot modify single node clusters")
		}
		if inbuf.State != edgeproto.TrackedState_Ready {
			return fmt.Errorf("cluster is not ready")
		}
		inbuf.CopyInFields(in)
		inbuf.State = edgeproto.TrackedState_UpdateRequested
		s.store.STMPut(stm, &inbuf)
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) {
		return nil
	}
	err = clusterInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_Ready, UpdateClusterInstTransitions, edgeproto.TrackedState_UpdateError, UpdateClusterInstTimeout, "Updated successfully", cb.Send)
	return err
}

func (s *ClusterInstApi) deleteClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	if appInstApi.UsesClusterInst(&in.Key) {
		return errors.New("ClusterInst in use by Application Instance")
	}
	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
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

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if in.State != edgeproto.TrackedState_DeletePrepare {
			return errors.New("ClusterInst expected state DeletePrepare")
		}

		nodeFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &in.Flavor, &nodeFlavor) {
			log.WarnLog("Delete cluster inst: flavor not found",
				"flavor", in.Flavor.Name)
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
			refs.UsedRam -= nodeFlavor.Ram * uint64(in.NumNodes+in.NumMasters)
			refs.UsedVcores -= nodeFlavor.Vcpus * uint64(in.NumNodes+in.NumMasters)
			refs.UsedDisk -= nodeFlavor.Disk * uint64(in.NumNodes+in.NumMasters)
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
