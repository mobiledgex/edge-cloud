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
	for _, data := range s.cache.Objs {
		cluster := data.Obj
		if cluster.Flavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *ClusterInstApi) UsesAutoScalePolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		cluster := data.Obj
		if cluster.AutoScalePolicy == key.Name {
			return true
		}
	}
	return false
}

func (s *ClusterInstApi) UsesPrivacyPolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		cluster := data.Obj
		if cluster.PrivacyPolicy == key.Name && cluster.Key.Organization == key.Organization {
			return true
		}
	}
	return false
}

func (s *ClusterInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.ClusterInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, data := range s.cache.Objs {
		val := data.Obj
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
	for key, data := range s.cache.Objs {
		val := data.Obj
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
	for _, data := range s.cache.Objs {
		val := data.Obj
		if val.Key.ClusterKey.Matches(key) {
			return true
		}
	}
	return false
}

// validateAndDefaultIPAccess checks that the IP access type is valid if it is set.  If it is not set
// it returns the new value based on the other parameters
func validateAndDefaultIPAccess(clusterInst *edgeproto.ClusterInst, platformType edgeproto.PlatformType, cb edgeproto.ClusterInstApi_CreateClusterInstServer) (edgeproto.IpAccess, error) {

	platName := edgeproto.PlatformType_name[int32(platformType)]

	// Operators such as GCP and Azure must be dedicated as they allocate a new IP per service
	if isIPAllocatedPerService(clusterInst.Key.CloudletKey.Organization) {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			cb.Send(&edgeproto.Result{Message: "Defaulting IpAccess to IpAccessDedicated for operator: " + clusterInst.Key.CloudletKey.Organization})
			return edgeproto.IpAccess_IP_ACCESS_DEDICATED, nil
		}
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
			return clusterInst.IpAccess, fmt.Errorf("IpAccessShared not supported for operator: %s", clusterInst.Key.CloudletKey.Organization)
		}
		return clusterInst.IpAccess, nil
	}
	if platformType == edgeproto.PlatformType_PLATFORM_TYPE_DIND || platformType == edgeproto.PlatformType_PLATFORM_TYPE_EDGEBOX {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			cb.Send(&edgeproto.Result{Message: "Defaulting IpAccess to IpAccessShared for platform " + platName})
			return edgeproto.IpAccess_IP_ACCESS_SHARED, nil
		}
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			return clusterInst.IpAccess, fmt.Errorf("IpAccessDedicated not supported platform: %s", platformType)
		}
	}
	// Privacy Policy required dedicated
	if clusterInst.PrivacyPolicy != "" {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			cb.Send(&edgeproto.Result{Message: "Defaulting IpAccess to IpAccessDedicated for privacy policy enabled cluster "})
			return edgeproto.IpAccess_IP_ACCESS_DEDICATED, nil
		}
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
			return clusterInst.IpAccess, fmt.Errorf("IpAccessShared not supported for privacy policy enabled cluster")
		}
		return clusterInst.IpAccess, nil
	}
	switch clusterInst.Deployment {
	case cloudcommon.DeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.DeploymentTypeHelm:
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			cb.Send(&edgeproto.Result{Message: "Defaulting IpAccess to IpAccessShared for deployment " + clusterInst.Deployment})
			return edgeproto.IpAccess_IP_ACCESS_SHARED, nil
		}
		return clusterInst.IpAccess, nil
	case cloudcommon.DeploymentTypeDocker:
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			cb.Send(&edgeproto.Result{Message: "Defaulting IpAccess to IpAccessDedicated for deployment " + clusterInst.Deployment})
			return edgeproto.IpAccess_IP_ACCESS_DEDICATED, nil
		}
	}
	return clusterInst.IpAccess, nil
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

	defer func() {
		if reterr == nil {
			RecordClusterInstEvent(cb.Context(), &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

	ctx := cb.Context()
	if in.Key.Organization == "" {
		return fmt.Errorf("ClusterInst Organization cannot be empty")
	}
	if in.Key.CloudletKey.Name == "" {
		return fmt.Errorf("Cloudlet name cannot be empty")
	}
	if in.Key.CloudletKey.Organization == "" {
		return fmt.Errorf("Cloudlet Organization name cannot be empty")
	}
	if in.Key.ClusterKey.Name == "" {
		return fmt.Errorf("Cluster name cannot be empty")
	}
	if in.Reservable && in.Key.Organization != cloudcommon.OrganizationMobiledgeX {
		return fmt.Errorf("Only %s ClusterInsts may be reservable", cloudcommon.OrganizationMobiledgeX)
	}

	// validate deployment
	if in.Deployment == "" {
		// assume kubernetes, because that's what we've been doing
		in.Deployment = cloudcommon.DeploymentTypeKubernetes
	}
	if in.Deployment == cloudcommon.DeploymentTypeHelm {
		// helm runs on kubernetes
		in.Deployment = cloudcommon.DeploymentTypeKubernetes
	}
	if in.Deployment == cloudcommon.DeploymentTypeVM {
		// friendly error message if they try to specify VM
		return fmt.Errorf("ClusterInst is not needed for deployment type %s, just create an AppInst directly", cloudcommon.DeploymentTypeVM)
	}

	// validate other parameters based on deployment type
	if in.Deployment == cloudcommon.DeploymentTypeKubernetes {
		// must have at least one master, but currently don't support
		// more than one.
		if in.NumMasters == 0 {
			// just set it to 1
			in.NumMasters = 1
		}
		if in.NumMasters > 1 {
			return fmt.Errorf("NumMasters cannot be greater than 1")
		}
	} else if in.Deployment == cloudcommon.DeploymentTypeDocker {
		if in.NumMasters != 0 || in.NumNodes != 0 {
			return fmt.Errorf("NumMasters and NumNodes not applicable for deployment type %s", cloudcommon.DeploymentTypeDocker)
		}
		if in.SharedVolumeSize != 0 {
			return fmt.Errorf("SharedVolumeSize not supported for deployment type %s", cloudcommon.DeploymentTypeDocker)

		}
	} else {
		return fmt.Errorf("Invalid deployment type %s for ClusterInst", in.Deployment)
	}

	// dedicatedOrShared(2) is removed
	if in.IpAccess == 2 {
		in.IpAccess = edgeproto.IpAccess_IP_ACCESS_UNKNOWN
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if err := checkCloudletReady(cctx, stm, &in.Key.CloudletKey); err != nil {
			return err
		}
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
		var err error
		platName := edgeproto.PlatformType_name[int32(cloudlet.PlatformType)]
		if cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_OPENSTACK &&
			cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_FAKE &&
			cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_VSPHERE &&
			in.SharedVolumeSize != 0 {
			return fmt.Errorf("Shared volumes not supported on %s", platName)
		}
		if len(in.Key.ClusterKey.Name) > cloudcommon.MaxClusterNameLength {
			return fmt.Errorf("Cluster name limited to %d characters", cloudcommon.MaxClusterNameLength)
		}
		if cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_AZURE || cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_GCP {
			if in.Deployment != cloudcommon.DeploymentTypeKubernetes {
				return errors.New("Only kubernetes clusters can be deployed in Azure or GCP")
			}
			if in.NumNodes == 0 {
				return errors.New("NumNodes cannot be 0 for Azure or GCP")
			}

		}
		if err := validateClusterInstUpdates(ctx, stm, in); err != nil {
			return err
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
		vmspec, err := resTagTableApi.GetVMSpec(ctx, stm, nodeFlavor, cloudlet, info)
		if err != nil {
			return err
		}
		in.NodeFlavor = vmspec.FlavorName
		in.AvailabilityZone = vmspec.AvailabilityZone
		in.ExternalVolumeSize = vmspec.ExternalVolumeSize
		log.SpanLog(ctx, log.DebugLevelApi, "Selected Cloudlet Node Flavor", "vmspec", vmspec, "master flavor", in.MasterNodeFlavor)

		// check if MasterNodeFlavor required
		if in.Deployment == cloudcommon.DeploymentTypeKubernetes && in.NumNodes > 0 {
			masterFlavor := edgeproto.Flavor{}
			masterFlavorKey := edgeproto.FlavorKey{}
			settings := settingsApi.Get()
			masterFlavorKey.Name = settings.MasterNodeFlavor

			if flavorApi.store.STMGet(stm, &masterFlavorKey, &masterFlavor) {
				log.SpanLog(ctx, log.DebugLevelApi, "MasterNodeFlavor found ", "MasterNodeFlavor", settings.MasterNodeFlavor)
				vmspec, err := resTagTableApi.GetVMSpec(ctx, stm, masterFlavor, cloudlet, info)
				if err != nil {
					// Unlikely with reasonably modest settings.MasterNodeFlavor sized flavor
					log.SpanLog(ctx, log.DebugLevelApi, "Error K8s Master Node Flavor matches no eixsting OS flavor", "nodeFlavor", in.NodeFlavor)
					return err
				} else {
					in.MasterNodeFlavor = vmspec.FlavorName
					log.SpanLog(ctx, log.DebugLevelApi, "Selected Cloudlet Master Node Flavor", "vmspec", vmspec, "master flavor", in.MasterNodeFlavor)
				}
			} else {
				// should never be non empty and not found due to validation in update
				// revert to using NodeFlavor (pre EC-1767) and log warning
				in.MasterNodeFlavor = in.NodeFlavor
				log.SpanLog(ctx, log.DebugLevelApi, "Warning : Master Node Flavor does not exist using", "master flavor", in.MasterNodeFlavor)
			}
		}
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
		in.IpAccess, err = validateAndDefaultIPAccess(in, cloudlet.PlatformType, cb)
		if err != nil {
			return err
		}
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

	err := in.ValidateUpdateFields()
	if err != nil {
		return err
	}
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	cctx.SetOverride(&in.CrmOverride)

	var inbuf edgeproto.ClusterInst
	var changeCount int
	retry := false
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		changeCount = 0
		if !s.store.STMGet(stm, &in.Key, &inbuf) {
			return in.Key.NotFoundError()
		}
		if inbuf.NumMasters == 0 {
			return fmt.Errorf("cannot modify single node clusters")
		}
		if err := checkCloudletReady(cctx, stm, &in.Key.CloudletKey); err != nil {
			return err
		}

		if !cctx.Undo && inbuf.State != edgeproto.TrackedState_READY && !ignoreTransient(cctx, inbuf.State) {
			if inbuf.State == edgeproto.TrackedState_UPDATE_ERROR {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("previous update failed, %v, trying again", inbuf.Errors)})
				retry = true
			} else {
				return errors.New("ClusterInst busy, cannot update")
			}
		}
		changeCount = inbuf.CopyInFields(in)
		if changeCount == 0 && !retry {
			// nothing changed
			return nil
		}
		if err := validateClusterInstUpdates(ctx, stm, in); err != nil {
			return err
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
	if changeCount == 0 && !retry {
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

func validateClusterInstUpdates(ctx context.Context, stm concurrency.STM, in *edgeproto.ClusterInst) error {
	cloudlet := edgeproto.Cloudlet{}
	if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
		return errors.New("Specified Cloudlet not found")
	}
	if in.PrivacyPolicy != "" {
		if cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_OPENSTACK && cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_VSPHERE && cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_FAKE {
			platName := edgeproto.PlatformType_name[int32(cloudlet.PlatformType)]
			return fmt.Errorf("Privacy Policy not supported on %s", platName)
		}
		policy := edgeproto.PrivacyPolicy{}
		if err := privacyPolicyApi.STMFind(stm, in.PrivacyPolicy, in.Key.Organization, &policy); err != nil {
			return err
		}
	}
	if in.AutoScalePolicy != "" {
		policy := edgeproto.AutoScalePolicy{}
		if err := autoScalePolicyApi.STMFind(stm, in.AutoScalePolicy, in.Key.Organization, &policy); err != nil {
			return err
		}
		if in.NumNodes < policy.MinNodes {
			in.NumNodes = policy.MinNodes
		}
		if in.NumNodes > policy.MaxNodes {
			in.NumNodes = policy.MaxNodes
		}
	}
	return nil
}

func (s *ClusterInstApi) deleteClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) (reterr error) {
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}
	// If it is autoClusterInst and creation had failed, then deletion should proceed
	// even though clusterinst is in use by Application Instance
	if !(cctx.Undo && strings.HasPrefix(in.Key.ClusterKey.Name, ClusterAutoPrefix)) {
		if appInstApi.UsesClusterInst(&in.Key) {
			return errors.New("ClusterInst in use by Application Instance")
		}
	}
	cctx.SetOverride(&in.CrmOverride)
	ctx := cb.Context()

	var prevState edgeproto.TrackedState
	// Set state to prevent other apps from being created on ClusterInst
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		if err := checkCloudletReady(cctx, stm, &in.Key.CloudletKey); err != nil {
			return err
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
			RecordClusterInstEvent(context.WithValue(ctx, in.Key, *in), &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
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
		if cloudlet.MaintenanceState != edgeproto.MaintenanceState_NORMAL_OPERATION {
			return errors.New("Cloudlet under maintenance, please try again later")
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
	now := time.Now()
	ts, _ := types.TimestampProto(now)
	metric.Timestamp = *ts
	// influx requires that at least one field must be specified when querying so these cant be all tags
	metric.AddStringVal("cloudletorg", clusterInstKey.CloudletKey.Organization)
	metric.AddTag("cloudlet", clusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", clusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", clusterInstKey.Organization)
	metric.AddStringVal("event", string(event))
	metric.AddStringVal("status", serverStatus)

	info, ok := ctx.Value(*clusterInstKey).(edgeproto.ClusterInst)
	if !ok { // if not provided (aka not recording a delete), get the flavorkey and numnodes ourself
		info = edgeproto.ClusterInst{}
		if !clusterInstApi.cache.Get(clusterInstKey, &info) {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Cannot log event for invalid clusterinst")
			return
		}
	}
	// if this is a clusterinst use the org its reserved for instead of MobiledgeX
	metric.AddTag("reservedBy", info.ReservedBy)
	// errors should never happen here since to get to this point the flavor should have already been checked previously, but just in case
	nodeFlavor := edgeproto.Flavor{}
	if !flavorApi.cache.Get(&info.Flavor, &nodeFlavor) {
		log.SpanLog(ctx, log.DebugLevelMetrics, "flavor not found for recording clusterInst lifecycle", "flavor name", info.Flavor.Name)
	} else {
		metric.AddTag("flavor", info.Flavor.Name)
		metric.AddIntVal("ram", nodeFlavor.Ram)
		metric.AddIntVal("vcpu", nodeFlavor.Vcpus)
		metric.AddIntVal("disk", nodeFlavor.Disk)
		metric.AddIntVal("nodeCount", uint64(info.NumMasters+info.NumNodes))
		metric.AddStringVal("other", fmt.Sprintf("%v", nodeFlavor.OptResMap))
	}
	services.events.AddMetric(&metric)

	// if it's a delete, create a usage record of it
	// get all the logs for this clusterinst since the last checkpoint
	go func() {
		if event == cloudcommon.DELETED || event == cloudcommon.UNRESERVED {
			err := CreateClusterUsageRecord(ctx, &info, now, cloudcommon.USAGE_EVENT_END)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelMetrics, "unable to create cluster usage record", "cluster", clusterInstKey, "err", err)
			}
		}
	}()
}
