package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/vmspec"
)

type ClusterInstApi struct {
	sync  *Sync
	store edgeproto.ClusterInstStore
	cache edgeproto.ClusterInstCache
}

var clusterInstApi = ClusterInstApi{}

const ClusterAutoPrefix = "autocluster"

var ClusterAutoPrefixErr = fmt.Sprintf("Cluster name prefix \"%s\" is reserved",
	ClusterAutoPrefix)

// Transition states indicate states in which the CRM is still busy.
var CreateClusterInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_CREATING: struct{}{},
}
var UpdateClusterInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_UPDATING: struct{}{},
}
var DeleteClusterInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_DELETING: struct{}{},
}

func InitClusterInstApi(sync *Sync) {
	clusterInstApi.sync = sync
	clusterInstApi.store = edgeproto.NewClusterInstStore(sync.store)
	edgeproto.InitClusterInstCache(&clusterInstApi.cache)
	sync.RegisterCache(&clusterInstApi.cache)
	if *shortTimeouts {
		cloudcommon.CreateClusterInstTimeout = 3 * time.Second
		cloudcommon.UpdateClusterInstTimeout = 2 * time.Second
		cloudcommon.DeleteClusterInstTimeout = 2 * time.Second
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

func (s *ClusterInstApi) UsesAutoScalePolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, cluster := range s.cache.Objs {
		if cluster.AutoScalePolicy == key.Name {
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
			if val.Liveness == edgeproto.Liveness_LIVENESS_STATIC {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC {
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

func (s *ClusterInstApi) GetCloudletResourceMap(resource string, resmap map[string]*edgeproto.ResTagTableKey) (*edgeproto.ResTagTable, error) {

	var ctx context.Context
	key := resmap[resource]
	tbl, err := resTagTableApi.GetResTagTable(ctx, key)
	if err != nil {
		fmt.Printf("\n\n\tGetCloudletResourceMap failed err: %s\n", err.Error())
		return tbl, err
	}
	return tbl, err
}

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func (s *ClusterInstApi) GetVMSpec(flavorList []*edgeproto.FlavorInfo, nodeflavor edgeproto.Flavor, resmap map[string]*edgeproto.ResTagTableKey) (*vmspec.VMCreationSpec, error) {
	log.InfoLog("GetVMSpec with closest flavor available", "flavorList", flavorList, "nodeflavor", nodeflavor)
	var vmspec vmspec.VMCreationSpec

	sort.Slice(flavorList[:], func(i, j int) bool {
		if flavorList[i].Vcpus < flavorList[j].Vcpus {
			return true
		}
		if flavorList[i].Vcpus > flavorList[j].Vcpus {
			return false
		}
		if flavorList[i].Ram < flavorList[j].Ram {
			return true
		}
		if flavorList[i].Ram > flavorList[j].Ram {
			return false
		}

		return flavorList[i].Disk < flavorList[j].Disk
	})
	for _, flavor := range flavorList {

		if flavor.Vcpus < nodeflavor.Vcpus {
			continue
		}
		if flavor.Ram < nodeflavor.Ram {
			continue
		}
		if flavor.Disk == 0 {
			// flavors of zero disk size mean that the volume is allocated separately
			vmspec.ExternalVolumeSize = nodeflavor.Disk
		} else if flavor.Disk < nodeflavor.Disk {
			continue
		}
		// Good matches for flavor so far, does nodeflavor request a gpu?
		if nodeflavor.Gpus > 0 {
			if !strings.Contains(flavor.Name, "gpu") {
				// now we need to fetch the tag table for the gpu resouce
				//  using the key in this cloudlet map
				//
				fmt.Printf("\n\n\tGetVMSpec flavor %s fetch the gpu res table\n", flavor.Name)
				tbl, err := /*clusterInstApi*/ s.GetCloudletResourceMap("gpu", resmap)

				if err != nil || tbl == nil {
					fmt.Printf("\n\n\t no body home from new fetch rtn...\n")
					// gpu requested, no table, request will fail
					continue
				}
				for _, tag := range tbl.Tags {
					if strings.Contains(flavor.Properties, tag) {
						continue
					}
				}
			}
		} // other optional resources matching e.g. nas_storage etc.
		vmspec.FlavorName = flavor.Name
		log.InfoLog("Found closest flavor", "flavor", flavor, "vmspec", vmspec)

		return &vmspec, nil
	}
	return &vmspec, fmt.Errorf("no suitable platform flavor found for %s, please try a smaller flavor", nodeflavor.Key.Name)
}

func (s *ClusterInstApi) CreateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
	in.Auto = false
	return s.createClusterInstInternal(DefCallContext(), in, cb)
}

// createClusterInstInternal is used to create dynamic cluster insts internally,
// bypassing static assignment. It is also used to create auto-cluster insts.
func (s *ClusterInstApi) createClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	cctx.SetOverride(&in.CrmOverride)
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
	}
	ctx := cb.Context()
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
	} else if in.Deployment == cloudcommon.AppDeploymentTypeDocker {
		if in.NumMasters != 0 || in.NumNodes != 0 {
			return fmt.Errorf("NumMasters and NumNodes not applicable for deployment type %s", cloudcommon.AppDeploymentTypeDocker)
		}
		if in.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			// must be dedicated for docker
			in.IpAccess = edgeproto.IpAccess_IP_ACCESS_DEDICATED
		} else if in.IpAccess != edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			return fmt.Errorf("IpAccess must be dedicated for deployment type %s", cloudcommon.AppDeploymentTypeDocker)
		}
	} else {
		return fmt.Errorf("Invalid deployment type %s for ClusterInst", in.Deployment)
	}

	if in.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
		// default to shared
		in.IpAccess = edgeproto.IpAccess_IP_ACCESS_SHARED
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if clusterInstApi.store.STMGet(stm, &in.Key, in) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
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
		if in.Liveness == edgeproto.Liveness_LIVENESS_UNKNOWN {
			in.Liveness = edgeproto.Liveness_LIVENESS_DYNAMIC
		}
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}
		if cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_AZURE || cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_GCP {
			if in.Deployment != cloudcommon.AppDeploymentTypeKubernetes {
				return errors.New("Only kubernetes clusters can be deployed in Azure or GCP")
			}
			if in.NumNodes == 0 {
				return errors.New("NumNodes cannot be 0 for Azure or GCP")
			}
			if len(in.Key.ClusterKey.Name) > cloudcommon.MaxClusterNameLength {
				return fmt.Errorf("Cluster name limited to %d characters for GCP and Azure", cloudcommon.MaxClusterNameLength)
			}
		}
		if in.AutoScalePolicy != "" {
			policy := edgeproto.AutoScalePolicy{}
			if err := autoScalePolicyApi.STMFind(stm, in.AutoScalePolicy, in.Key.Developer, &policy); err != nil {
				return err
			}
			if in.NumNodes < policy.MinNodes {
				in.NumNodes = policy.MinNodes
			}
			if in.NumNodes > policy.MaxNodes {
				in.NumNodes = policy.MaxNodes
			}
		}
		info := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.store.STMGet(stm, &in.Key.CloudletKey, &info) {
			return fmt.Errorf("No resource information found for Cloudlet %s", in.Key.CloudletKey)
		}
		refs := edgeproto.CloudletRefs{}
		if !cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &refs) {
			initCloudletRefs(&refs, &in.Key.CloudletKey)
		}

		if in.Flavor.Name == "" {
			return errors.New("No Flavor specified")
		}

		nodeFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &in.Flavor, &nodeFlavor) {
			return fmt.Errorf("flavor %s not found", in.Flavor.Name)
		}
		var err error
		vmspec, err := s.GetVMSpec(info.Flavors, nodeFlavor, cloudlet.ResTagMap)
		if err != nil {
			return err
		}
		in.NodeFlavor = vmspec.FlavorName
		in.ExternalVolumeSize = vmspec.ExternalVolumeSize

		log.SpanLog(ctx, log.DebugLevelApi, "Selected Cloudlet Flavor", "vmspec", vmspec)
		// Do we allocate resources based on max nodes (no over-provisioning)?
		refs.UsedRam += nodeFlavor.Ram * uint64(in.NumNodes+in.NumMasters)
		refs.UsedVcores += nodeFlavor.Vcpus * uint64(in.NumNodes+in.NumMasters)
		refs.UsedDisk += (nodeFlavor.Disk + vmspec.ExternalVolumeSize) * uint64(in.NumNodes+in.NumMasters)
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
			in.State = edgeproto.TrackedState_READY
		} else {
			in.State = edgeproto.TrackedState_CREATE_REQUESTED
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
	err = clusterInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_READY, CreateClusterInstTransitions, edgeproto.TrackedState_CREATE_ERROR, cloudcommon.CreateClusterInstTimeout, "Created successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create ClusterInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created ClusterInst successfully"})
		err = nil
	}
	if err != nil {
		// XXX should probably track mod revision ID and only undo
		// if no other changes were made to appInst in the meantime.
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "DELETING ClusterInst due to failures"})
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
	return s.updateClusterInstInternal(DefCallContext(), in, cb)
}

func (s *ClusterInstApi) updateClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "updateClusterInstInternal")
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	if in.Fields == nil {
		return fmt.Errorf("nothing specified to update")
	}
	allowedFields := []string{}
	allowedFieldsMap := make(map[string]struct{})
	badFields := []string{}
	for _, field := range in.Fields {
		if field == edgeproto.ClusterInstFieldCrmOverride ||
			field == edgeproto.ClusterInstFieldKey ||
			in.IsKeyField(field) {
			continue
		} else if field == edgeproto.ClusterInstFieldNumNodes {
			allowedFields = append(allowedFields, field)
			allowedFieldsMap[field] = struct{}{}
		} else {
			badFields = append(badFields, field)
		}
	}
	if len(badFields) > 0 {
		// cat all the bad field names and return error
		badstrs := []string{}
		for _, bad := range badFields {
			badstrs = append(badstrs, edgeproto.ClusterInstAllFieldsStringMap[bad])
		}
		return fmt.Errorf("specified field(s) %s cannot be modified", strings.Join(badstrs, ","))
	}
	in.Fields = allowedFields
	if len(allowedFields) == 0 {
		return fmt.Errorf("Nothing specified to modify")
	}

	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
	}

	var inbuf edgeproto.ClusterInst
	var changeCount int
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		changeCount = 0
		if !s.store.STMGet(stm, &in.Key, &inbuf) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if inbuf.NumMasters == 0 {
			return fmt.Errorf("cannot modify single node clusters")
		}

		if !cctx.Undo && inbuf.State != edgeproto.TrackedState_READY && !ignoreTransient(cctx, inbuf.State) {
			if inbuf.State == edgeproto.TrackedState_UPDATE_ERROR {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("previous update failed, %v, trying again", inbuf.Errors)})
			} else {
				return errors.New("ClusterInst busy, cannot update")
			}
		}
		changeCount = inbuf.CopyInFields(in)
		if changeCount == 0 {
			// nothing changed
			return nil
		}
		inbuf.State = edgeproto.TrackedState_UPDATE_REQUESTED
		s.store.STMPut(stm, &inbuf)
		return nil
	})
	if err != nil {
		return err
	}
	if changeCount == 0 {
		return nil
	}
	if ignoreCRM(cctx) {
		return nil
	}
	err = clusterInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_READY, UpdateClusterInstTransitions, edgeproto.TrackedState_UPDATE_ERROR, cloudcommon.UpdateClusterInstTimeout, "Updated successfully", cb.Send)
	return err
}

func (s *ClusterInstApi) deleteClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}
	if appInstApi.UsesClusterInst(&in.Key) {
		return errors.New("ClusterInst in use by Application Instance")
	}
	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
	}
	ctx := cb.Context()

	var prevState edgeproto.TrackedState
	// Set state to prevent other apps from being created on ClusterInst
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if !cctx.Undo && in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR && in.State != edgeproto.TrackedState_DELETE_PREPARE && in.State != edgeproto.TrackedState_UPDATE_ERROR && !ignoreTransient(cctx, in.State) {
			if in.State == edgeproto.TrackedState_DELETE_ERROR {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateClusterInst to rebuild, and try again"})
			}
			return errors.New("ClusterInst busy, cannot delete")
		}
		prevState = in.State
		in.State = edgeproto.TrackedState_DELETE_PREPARE
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
		s.store.Update(ctx, in, s.sync.syncWait)
		return fmt.Errorf("Failed to auto-delete applications from clusterInst %s, %s",
			in.Key.ClusterKey.Name, err.Error())
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if in.State != edgeproto.TrackedState_DELETE_PREPARE {
			return errors.New("ClusterInst expected state DELETE_PREPARE")
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
			in.State = edgeproto.TrackedState_DELETE_REQUESTED
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
	err = clusterInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_NOT_PRESENT, DeleteClusterInstTransitions, edgeproto.TrackedState_DELETE_ERROR, cloudcommon.DeleteClusterInstTimeout, "Deleted ClusterInst successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete ClusterInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_NOT_PRESENT)
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
	case edgeproto.TrackedState_CREATE_REQUESTED:
		if next == edgeproto.TrackedState_CREATING || next == edgeproto.TrackedState_READY || next == edgeproto.TrackedState_CREATE_ERROR {
			return true
		}
	case edgeproto.TrackedState_CREATING:
		if next == edgeproto.TrackedState_READY || next == edgeproto.TrackedState_CREATE_ERROR {
			return true
		}
	case edgeproto.TrackedState_UPDATE_REQUESTED:
		if next == edgeproto.TrackedState_UPDATING || next == edgeproto.TrackedState_READY || next == edgeproto.TrackedState_UPDATE_ERROR {
			return true
		}
	case edgeproto.TrackedState_UPDATING:
		if next == edgeproto.TrackedState_READY || next == edgeproto.TrackedState_UPDATE_ERROR {
			return true
		}
	case edgeproto.TrackedState_DELETE_REQUESTED:
		if next == edgeproto.TrackedState_DELETING || next == edgeproto.TrackedState_NOT_PRESENT || next == edgeproto.TrackedState_DELETE_ERROR {
			return true
		}
	case edgeproto.TrackedState_DELETING:
		if next == edgeproto.TrackedState_NOT_PRESENT || next == edgeproto.TrackedState_DELETE_ERROR {
			return true
		}
	}
	return false
}

func ignoreTransient(cctx *CallContext, state edgeproto.TrackedState) bool {
	if cctx.Override == edgeproto.CRMOverride_IGNORE_TRANSIENT_STATE ||
		cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE {
		if state == edgeproto.TrackedState_CREATING ||
			state == edgeproto.TrackedState_CREATE_REQUESTED ||
			state == edgeproto.TrackedState_UPDATE_REQUESTED ||
			state == edgeproto.TrackedState_DELETE_REQUESTED ||
			state == edgeproto.TrackedState_UPDATING ||
			state == edgeproto.TrackedState_DELETING ||
			state == edgeproto.TrackedState_DELETE_PREPARE {
			return true
		}
	}
	return false
}

func ignoreCRM(cctx *CallContext) bool {
	if cctx.Undo || cctx.Override == edgeproto.CRMOverride_IGNORE_CRM ||
		cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE {
		return true
	}
	return false
}

func (s *ClusterInstApi) UpdateFromInfo(ctx context.Context, in *edgeproto.ClusterInstInfo) {
	log.SpanLog(ctx, log.DebugLevelApi, "update ClusterInst", "state", in.State, "status", in.Status)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State == in.State {
			if inst.Status == in.Status {
				return nil
			} else {
				log.SpanLog(ctx, log.DebugLevelApi, "status change only")
				inst.Status = in.Status
				s.store.STMPut(stm, &inst)
				return nil
			}
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.SpanLog(ctx, log.DebugLevelApi, "invalid state transition", "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		inst.Status = in.Status
		if in.State == edgeproto.TrackedState_CREATE_ERROR || in.State == edgeproto.TrackedState_DELETE_ERROR || in.State == edgeproto.TrackedState_UPDATE_ERROR {
			inst.Errors = in.Errors
		} else {
			inst.Errors = nil
		}
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *ClusterInstApi) DeleteFromInfo(ctx context.Context, in *edgeproto.ClusterInstInfo) {
	log.SpanLog(ctx, log.DebugLevelApi, "delete ClusterInst from info", "state", in.State)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_DELETING && inst.State != edgeproto.TrackedState_DELETE_REQUESTED {
			log.SpanLog(ctx, log.DebugLevelApi, "invalid state transition", "cur", inst.State, "next", edgeproto.TrackedState_NOT_PRESENT)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
}

func (s *ClusterInstApi) ReplaceErrorState(ctx context.Context, in *edgeproto.ClusterInst, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
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
