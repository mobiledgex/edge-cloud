package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type ClusterInstApi struct {
	sync  *Sync
	store edgeproto.ClusterInstStore
	cache edgeproto.ClusterInstCache
}

var clusterInstApi = ClusterInstApi{}

var CreateClusterInstTimeout = 60 * time.Second
var UpdateClusterInstTimeout = 30 * time.Second
var DeleteClusterInstTimeout = 30 * time.Second

func InitClusterInstApi(sync *Sync) {
	clusterInstApi.sync = sync
	clusterInstApi.store = edgeproto.NewClusterInstStore(sync.store)
	edgeproto.InitClusterInstCache(&clusterInstApi.cache)
	clusterInstApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateClusterInst)
	sync.RegisterCache(&clusterInstApi.cache)
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

func (s *ClusterInstApi) GetClusterInstsForCloudlets(cloudlets map[edgeproto.CloudletKey]struct{}, clusterInsts map[edgeproto.ClusterInstKey]struct{}) {
	s.cache.GetClusterInstsForCloudlets(cloudlets, clusterInsts)
}

func (s *ClusterInstApi) CreateNoWait(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	in.Liveness = edgeproto.Liveness_LivenessStatic
	in.Auto = false
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		return s.createClusterInstInternal(stm, in)
	})
	return &edgeproto.Result{}, err
}

func (s *ClusterInstApi) createClusterInstInternal(stm concurrency.STM, in *edgeproto.ClusterInst) error {
	if clusterInstApi.store.STMGet(stm, &in.Key, nil) {
		return objstore.ErrKVStoreKeyExists
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
		return errors.New("No resource information found for Cloudlet")
	}
	refs := edgeproto.CloudletRefs{}
	if !cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &refs) {
		initCloudletRefs(&refs, &in.Key.CloudletKey)
	}

	var cluster edgeproto.Cluster
	if !clusterApi.store.STMGet(stm, &in.Key.ClusterKey, &cluster) {
		return errors.New("Specified Cluster not found")
	}
	if in.Flavor.Name == "" {
		in.Flavor = cluster.DefaultFlavor
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
	if !flavorApi.store.STMGet(stm, &clusterFlavor.MasterFlavor, nil) {
		return fmt.Errorf("Cluster flavor %s master flavor %s not found",
			in.Flavor.Name, clusterFlavor.MasterFlavor.Name)
	}

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
	refs.Clusters = append(refs.Clusters, in.Key.ClusterKey)
	cloudletRefsApi.store.STMPut(stm, &refs)
	s.store.STMPut(stm, in)
	return nil
}

func (s *ClusterInstApi) UpdateNoWait(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update cluster instance not supported")
	// TODO: Set info state to "changing" so update waits for CRM to
	// set state to "ready".
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *ClusterInstApi) DeleteNoWait(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	return s.deleteClusterInstInternal(in)
}

func (s *ClusterInstApi) deleteClusterInstInternal(in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	if appInstApi.UsesClusterInst(&in.Key) {
		return &edgeproto.Result{}, errors.New("ClusterInst in use by Application Instance")
	}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return nil
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
			cloudletRefsApi.store.STMPut(stm, &refs)
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *ClusterInstApi) ShowClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_ShowClusterInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *ClusterInstApi) CreateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	_, err := s.CreateNoWait(cb.Context(), in)
	if err != nil {
		return err
	}
	err = clusterInstInfoApi.cache.WaitForState(&in.Key, edgeproto.ClusterState_ClusterStateReady, CreateClusterInstTimeout, cb.Send)
	if err != nil {
		// XXX should probably track mod revision ID and only undo
		// if no other changes were made to appInst in the meantime.
		// crm failed or some other err, undo
		_, undoErr := s.DeleteNoWait(cb.Context(), in)
		if undoErr != nil {
			log.InfoLog("Undo create clusterinst", "undoErr", undoErr)
		} else {
			undoErr = clusterInstInfoApi.cache.WaitForState(&in.Key, edgeproto.ClusterState_ClusterStateNotPresent, DeleteClusterInstTimeout, nil)
			if undoErr != nil {
				log.InfoLog("Undo create clusterinst", "undoErr", undoErr)
			}
		}
		log.DebugLog(log.DebugLevelApi, "Create cluster inst failed",
			"in", in, "err", err)
	}
	return err
}

func (s *ClusterInstApi) DeleteClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	_, err := s.DeleteNoWait(cb.Context(), in)
	if err != nil {
		return err
	}
	err = clusterInstInfoApi.cache.WaitForState(&in.Key, edgeproto.ClusterState_ClusterStateNotPresent, DeleteClusterInstTimeout, cb.Send)
	if err != nil {
		// crm failed or some other err, undo
		_, undoErr := s.CreateNoWait(cb.Context(), in)
		if undoErr != nil {
			log.InfoLog("Undo delete clusterinst", "undoErr", undoErr)
		} else {
			undoErr = clusterInstInfoApi.cache.WaitForState(&in.Key, edgeproto.ClusterState_ClusterStateReady, CreateClusterInstTimeout, nil)
			if undoErr != nil {
				log.InfoLog("Undo delete clusterinst", "undoErr", undoErr)
			}
		}
		log.DebugLog(log.DebugLevelApi, "Delete cluster inst failed",
			"in", in, "err", err)
	}
	return err
}

func (s *ClusterInstApi) UpdateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_UpdateClusterInstServer) error {
	_, err := s.UpdateNoWait(cb.Context(), in)
	if err != nil {
		return err
	}
	err = clusterInstInfoApi.cache.WaitForState(&in.Key, edgeproto.ClusterState_ClusterStateReady, UpdateClusterInstTimeout, cb.Send)
	if err != nil {
		// TODO: undo.
		// Currently update is not supported. To undo we'd need to get
		// a copy of the data before the change to undo the change.
	}
	return err
}
