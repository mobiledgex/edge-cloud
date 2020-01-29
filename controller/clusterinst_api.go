package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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

func (s *ClusterInstApi) UsesPrivacyPolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, cluster := range s.cache.Objs {
		if cluster.PrivacyPolicy == key.Name && cluster.Key.Developer == key.Developer {
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

// Checks if there is some action in progress by ClusterInst on the cloudlet
func (s *ClusterInstApi) UsingCloudlet(in *edgeproto.CloudletKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, val := range s.cache.Objs {
		if key.CloudletKey.Matches(in) {
			if edgeproto.IsTransientState(val.State) {
				return true
			}
		}
	}
	return false
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
	in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
	in.Auto = false
	return s.createClusterInstInternal(DefCallContext(), in, cb)
}

// createClusterInstInternal is used to create dynamic cluster insts internally,
// bypassing static assignment. It is also used to create auto-cluster insts.
func (s *ClusterInstApi) createClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) (reterr error) {
	cctx.SetOverride(&in.CrmOverride)
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
			return err
		}
	}

	defer func() {
		if reterr == nil {
			RecordClusterInstEvent(cb.Context(), &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

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
	if in.Reservable && in.Key.Developer != cloudcommon.DeveloperMobiledgeX {
		return fmt.Errorf("Only %s ClusterInsts may be reservable", cloudcommon.DeveloperMobiledgeX)
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
		if in.SharedVolumeSize != 0 {
			return fmt.Errorf("SharedVolumeSize not supported for deployment type %s", cloudcommon.AppDeploymentTypeDocker)

		}
	} else {
		return fmt.Errorf("Invalid deployment type %s for ClusterInst", in.Deployment)
	}

	// dedicatedOrShared(2) is removed
	if in.IpAccess == 2 {
		in.IpAccess = edgeproto.IpAccess_IP_ACCESS_UNKNOWN
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if clusterInstApi.store.STMGet(stm, &in.Key, in) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteClusterInst to remove and try again"})
				}
				return in.Key.ExistsError()
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
		isSharedOnly := cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_DIND || cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_EDGEBOX
		platName := edgeproto.PlatformType_name[int32(cloudlet.PlatformType)]
		if isSharedOnly {
			if in.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
				return fmt.Errorf("IpAccess must be shared for cloudlet platform type %s", platName)
			}
			// override unknown
			in.IpAccess = edgeproto.IpAccess_IP_ACCESS_SHARED
		} else if in.Deployment == cloudcommon.AppDeploymentTypeDocker {
			// docker must be dedicated
			if in.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
				return fmt.Errorf("IpAccess must be dedicated for deployment type %s cloudlet platform type %s", cloudcommon.AppDeploymentTypeDocker, platName)
			}
			// override unknown
			in.IpAccess = edgeproto.IpAccess_IP_ACCESS_DEDICATED
		}

		if cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_OPENSTACK && cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_FAKE && in.SharedVolumeSize != 0 {
			return fmt.Errorf("Shared volumes not supported on %s", platName)
		}
		if in.PrivacyPolicy != "" {
			if cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_OPENSTACK && cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_FAKE {
				return fmt.Errorf("Privacy Policy not supported on %s", platName)
			}
			if in.IpAccess != edgeproto.IpAccess_IP_ACCESS_DEDICATED {
				if in.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
					in.IpAccess = edgeproto.IpAccess_IP_ACCESS_DEDICATED
				} else {
					return fmt.Errorf("PrivacyPolicy only supported for IP_ACCESS_DEDICATED")
				}
			}
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
		if in.PrivacyPolicy != "" {
			policy := edgeproto.PrivacyPolicy{}
			if err := privacyPolicyApi.STMFind(stm, in.PrivacyPolicy, in.Key.Developer, &policy); err != nil {
				return err
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
		vmspec, err := resTagTableApi.GetVMSpec(ctx, nodeFlavor, cloudlet, info)
		if err != nil {
			return err
		}
		in.NodeFlavor = vmspec.FlavorName
		in.AvailabilityZone = vmspec.AvailabilityZone
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
	err = clusterInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_READY, CreateClusterInstTransitions, edgeproto.TrackedState_CREATE_ERROR, settingsApi.Get().CreateClusterInstTimeout.TimeDuration(), "Created ClusterInst successfully", cb.Send)
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
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Failed to undo ClusterInst creation: %v", undoErr)})
			log.InfoLog("Undo create ClusterInst", "undoErr", undoErr)
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

func (s *ClusterInstApi) updateClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) (reterr error) {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "updateClusterInstInternal")
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	if in.Fields == nil {
		return fmt.Errorf("nothing specified to update")
	}
	allowedFields := []string{}
	badFields := []string{}
	for _, field := range in.Fields {
		if field == edgeproto.ClusterInstFieldCrmOverride ||
			field == edgeproto.ClusterInstFieldKey ||
			in.IsKeyField(field) {
			continue
		} else if field == edgeproto.ClusterInstFieldNumNodes || field == edgeproto.ClusterInstFieldAutoScalePolicy {
			allowedFields = append(allowedFields, field)
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
			return in.Key.NotFoundError()
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
		if !ignoreCRM(cctx) {
			inbuf.State = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		s.store.STMPut(stm, &inbuf)
		return nil
	})
	if err != nil {
		return err
	}
	if changeCount == 0 {
		return nil
	}

	RecordClusterInstEvent(ctx, &in.Key, cloudcommon.UPDATE_START, cloudcommon.InstanceDown)
	defer func() {
		if reterr == nil {
			RecordClusterInstEvent(ctx, &in.Key, cloudcommon.UPDATE_COMPLETE, cloudcommon.InstanceUp)
		} else {
			RecordClusterInstEvent(ctx, &in.Key, cloudcommon.UPDATE_ERROR, cloudcommon.InstanceDown)
		}
	}()

	if ignoreCRM(cctx) {
		return nil
	}
	err = clusterInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_READY, UpdateClusterInstTransitions, edgeproto.TrackedState_UPDATE_ERROR, settingsApi.Get().UpdateClusterInstTimeout.TimeDuration(), "Updated ClusterInst successfully", cb.Send)
	return err
}

func (s *ClusterInstApi) deleteClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) (reterr error) {
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}
	// If it is autoClusterInst and creation had failed, then deletion should proceed
	// even though clusterinst is in use by Application Instance
	//if !(cctx.Undo && strings.HasPrefix(in.Key.ClusterKey.Name, ClusterAutoPrefix)) {
	if appInstApi.UsesClusterInst(&in.Key) {
		return errors.New("ClusterInst in use by Application Instance")
	}
	//}
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
			return in.Key.NotFoundError()
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

	defer func() {
		if reterr == nil {
			RecordClusterInstEvent(ctx, &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
		}
	}()

	// Delete appInsts that are set for autodelete
	if err := appInstApi.AutoDeleteAppInsts(&in.Key, cctx.Override, cb); err != nil {
		// restore previous state since we failed pre-delete actions
		in.State = prevState
		s.store.Update(ctx, in, s.sync.syncWait)
		return fmt.Errorf("Failed to auto-delete applications from ClusterInst %s, %s",
			in.Key.ClusterKey.Name, err.Error())
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		if in.State != edgeproto.TrackedState_DELETE_PREPARE {
			return errors.New("ClusterInst expected state DELETE_PREPARE")
		}

		nodeFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &in.Flavor, &nodeFlavor) {
			log.WarnLog("Delete ClusterInst: flavor not found",
				"flavor", in.Flavor.Name)
		}
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
			log.WarnLog("Delete ClusterInst: cloudlet not found",
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
	err = clusterInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_NOT_PRESENT, DeleteClusterInstTransitions, edgeproto.TrackedState_DELETE_ERROR, settingsApi.Get().DeleteClusterInstTimeout.TimeDuration(), "Deleted ClusterInst successfully", cb.Send)
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
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Failed to undo ClusterInst deletion: %v", undoErr)})
			log.InfoLog("Undo delete ClusterInst", "undoErr", undoErr)
			RecordClusterInstEvent(ctx, &in.Key, cloudcommon.DELETE_ERROR, cloudcommon.InstanceDown)
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
		return edgeproto.IsTransientState(state)
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

func RecordClusterInstEvent(ctx context.Context, clusterInstKey *edgeproto.ClusterInstKey, event cloudcommon.InstanceEvent, serverStatus string) {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.ClusterInstEvent
	ts, _ := types.TimestampProto(time.Now())
	metric.Timestamp = *ts
	metric.AddTag("operator", clusterInstKey.CloudletKey.OperatorKey.Name)
	metric.AddTag("cloudlet", clusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", clusterInstKey.ClusterKey.Name)
	metric.AddTag("dev", clusterInstKey.Developer)
	metric.AddStringVal("event", string(event))
	metric.AddStringVal("status", serverStatus)

	info := edgeproto.ClusterInst{}
	if !clusterInstApi.cache.Get(clusterInstKey, &info) {
		log.SpanLog(ctx, log.DebugLevelApi, "Cannot log event for invalid clusterinst")
		return
	}
	// errors should never happen here since to get to this point the flavor should have already been checked previously, but just in case
	nodeFlavor := edgeproto.Flavor{}
	if !flavorApi.cache.Get(&info.Flavor, &nodeFlavor) {
		log.SpanLog(ctx, log.DebugLevelApi, "flavor not found for recording clusterInst lifecycle", "flavor name", info.Flavor.Name)
	} else {
		metric.AddTag("flavor", info.Flavor.Name)
		metric.AddIntVal("ram", nodeFlavor.Ram)
		metric.AddIntVal("vcpu", nodeFlavor.Vcpus)
		metric.AddIntVal("disk", nodeFlavor.Disk)
		metric.AddIntVal("nodeCount", uint64(info.NumMasters+info.NumNodes))
		metric.AddStringVal("other", fmt.Sprintf("%v", nodeFlavor.OptResMap))
	}

	services.events.AddMetric(&metric)
}
