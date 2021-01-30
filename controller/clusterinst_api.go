package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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
	if isIPAllocatedPerService(platformType, clusterInst.Key.CloudletKey.Organization) {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			cb.Send(&edgeproto.Result{Message: "Defaulting IpAccess to IpAccessDedicated for platform: " + platName})
			return edgeproto.IpAccess_IP_ACCESS_DEDICATED, nil
		}
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
			return clusterInst.IpAccess, fmt.Errorf("IpAccessShared not supported for platform: %s", platName)
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

func startClusterInstStream(ctx context.Context, key *edgeproto.ClusterInstKey, inCb edgeproto.ClusterInstApi_CreateClusterInstServer) (*streamSend, edgeproto.ClusterInstApi_CreateClusterInstServer, error) {
	streamKey := &edgeproto.AppInstKey{ClusterInstKey: *key}
	streamSendObj, err := streamObjApi.startStream(ctx, streamKey, inCb)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to start ClusterInst stream", "err", err)
		return nil, inCb, err
	}
	return streamSendObj, &CbWrapper{
		streamSendObj: streamSendObj,
		GenericCb:     inCb,
	}, nil
}

func stopClusterInstStream(ctx context.Context, key *edgeproto.ClusterInstKey, streamSendObj *streamSend, objErr error) {
	streamKey := &edgeproto.AppInstKey{ClusterInstKey: *key}
	if err := streamObjApi.stopStream(ctx, streamKey, streamSendObj, objErr); err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to stop ClusterInst stream", "err", err)
	}
}

func (s *StreamObjApi) StreamClusterInst(key *edgeproto.ClusterInstKey, cb edgeproto.StreamObjApi_StreamClusterInstServer) error {
	return s.StreamMsgs(&edgeproto.AppInstKey{ClusterInstKey: *key}, cb)
}

func (s *ClusterInstApi) CreateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	in.Auto = false
	return s.createClusterInstInternal(DefCallContext(), in, cb)
}

func getClusterInstVMRequirements(ctx context.Context, stm concurrency.STM, in *edgeproto.ClusterInst,
	cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo, cloudletRefs *edgeproto.CloudletRefs) ([]edgeproto.VMResource, error) {
	cloudletPlatform, err := pfutils.GetPlatform(ctx, cloudlet.PlatformType.String())
	if err != nil {
		return nil, err
	}
	rootlbFlavor, err := cloudletPlatform.GetRootLBFlavor(ctx)
	if err != nil {
		return nil, err
	}
	lbFlavor := &edgeproto.FlavorInfo{}
	if rootlbFlavor != nil {
		vmspec, err := resTagTableApi.GetVMSpec(ctx, stm, *rootlbFlavor, *cloudlet, *cloudletInfo)
		if err != nil {
			return nil, err
		}
		lbFlavor = vmspec.FlavorInfo
	}
	res, err := cloudcommon.GetClusterInstVMRequirements(ctx, in, cloudletInfo.Flavors, lbFlavor)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func validateCloudletCommonResources(ctx context.Context, stm concurrency.STM, cloudlet *edgeproto.Cloudlet, allClusterResources, reqdVmResources []edgeproto.VMResource) ([]string, error) {
	gpusReqd := uint64(0)
	gpusUsed := uint64(0)
	gpusMax := uint64(0)
	gpuThreshPercent := cloudlet.DefaultResourceAlertThreshold
	for _, resQuota := range cloudlet.ResourceQuotas {
		switch resQuota.Name {
		case cloudcommon.ResourceGpus:
			gpusMax = resQuota.Value
			gpuThreshPercent = resQuota.AlertThreshold
		}
	}
	for _, vmRes := range allClusterResources {
		if vmRes.VmFlavor != nil {
			if resTagTableApi.UsesGpu(ctx, stm, *vmRes.VmFlavor, *cloudlet) {
				gpusUsed += 1
			}
		}
	}
	for _, vmRes := range reqdVmResources {
		if vmRes.VmFlavor != nil {
			if resTagTableApi.UsesGpu(ctx, stm, *vmRes.VmFlavor, *cloudlet) {
				gpusReqd += 1
			}
		}
	}
	var err error
	warnings := []string{}
	if gpusMax > 0 {
		if float64(gpusUsed*100)/float64(gpusMax) > float64(gpuThreshPercent) {
			warnings = append(warnings, fmt.Sprintf("More than %d%% of GPUs are used", gpuThreshPercent))
		}
		gpusAvailable := gpusMax - gpusUsed
		if gpusReqd > gpusAvailable {
			err = fmt.Errorf("Not enough GPUs available, required %d but only %d is available", gpusReqd, gpusAvailable)
		}
	}
	return warnings, err
}

// Validate resource requirements for the VMs on the cloudlet
func validateCloudletInfraResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, infraResources *edgeproto.InfraResourcesSnapshot, allClusterResources, reqdVmResources, diffVmResources []edgeproto.VMResource) ([]string, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "Validate cloudlet resources", "vm resources", reqdVmResources, "cloudlet resources", infraResources)

	defaultAlertThresh := cloudlet.DefaultResourceAlertThreshold

	infraResInfo := make(map[string]*edgeproto.InfraResource)
	for _, resInfo := range infraResources.Info {
		newResInfo := &edgeproto.InfraResource{}
		newResInfo.DeepCopyIn(&resInfo)
		infraResInfo[newResInfo.Name] = newResInfo
	}

	reqdResInfo, err := GetCloudletResourceInfo(ctx, cloudlet, reqdVmResources, infraResInfo)
	if err != nil {
		return nil, err
	}
	allResInfo, err := GetCloudletResourceInfo(ctx, cloudlet, allClusterResources, infraResInfo)
	if err != nil {
		return nil, err
	}
	diffResInfo, err := GetCloudletResourceInfo(ctx, cloudlet, diffVmResources, infraResInfo)
	if err != nil {
		return nil, err
	}
	resQuotasInfo := make(map[string]edgeproto.InfraResource)
	for _, resQuota := range cloudlet.ResourceQuotas {
		resQuotasInfo[resQuota.Name] = edgeproto.InfraResource{
			Name:           resQuota.Name,
			Value:          resQuota.Value,
			AlertThreshold: resQuota.AlertThreshold,
		}
	}

	warnings := []string{}
	quotaResReqd := make(map[string]uint64)

	// Theoretical Validation
	for resName, resInfo := range allResInfo {
		max := resInfo.MaxValue
		thresh := defaultAlertThresh
		// look up quota if any
		if quota, found := resQuotasInfo[resName]; found {
			if quota.Value > 0 {
				// Set max value from resource quota
				max = quota.Value
			}
			if quota.AlertThreshold > 0 {
				// Set threshold value from resource quota
				thresh = quota.AlertThreshold
			}
		}
		if max == 0 {
			// no validation can be done
			continue
		}
		resReqd, ok := reqdResInfo[resName]
		if !ok {
			return nil, fmt.Errorf("Missing resource from required resource info: %s", resName)
		}
		thAvailableResVal := max - resInfo.Value
		if float64(resInfo.Value*100)/float64(max) > float64(thresh) {
			warnings = append(warnings, fmt.Sprintf("More than %d%% of %s is used", thresh, resName))
		}
		if resReqd.Value > thAvailableResVal {
			return warnings, fmt.Errorf("Not enough %s available, required %d%s but only %d%s is available", resName, resReqd.Value, resInfo.Units, thAvailableResVal, resInfo.Units)
		}
		quotaResReqd[resName] = thAvailableResVal
	}

	// Infra based validation
	for resName, _ := range infraResInfo {
		if resInfo, ok := diffResInfo[resName]; ok {
			infraResInfo[resName].Value += resInfo.Value
		}
	}
	for resName, resInfo := range infraResInfo {
		thresh := defaultAlertThresh
		// look up quota if any
		if quota, found := resQuotasInfo[resName]; found {
			if quota.AlertThreshold > 0 {
				// Set threshold value from resource quota
				thresh = quota.AlertThreshold
			}
		}
		if resInfo.MaxValue == 0 {
			// no validation can be done
			continue
		}
		resReqd, ok := reqdResInfo[resName]
		if !ok {
			return nil, fmt.Errorf("Missing resource from required resource info: %s", resName)
		}
		infraAvailableResVal := resInfo.MaxValue - resInfo.Value
		if float64(resInfo.Value*100)/float64(resInfo.MaxValue) > float64(thresh) {
			warnings = append(warnings, fmt.Sprintf("[Infra] More than %d%% of %s is used", thresh, resName))
		}
		if resReqd.Value > infraAvailableResVal {
			return warnings, fmt.Errorf("[Infra] Not enough %s available, required %d%s but only %d%s is available", resName, resReqd.Value, resInfo.Units, infraAvailableResVal, resInfo.Units)
		}
		if resVal, ok := quotaResReqd[resName]; ok {
			// generate alert if expected resource quota is not available
			if infraAvailableResVal < resVal {
				warnings = append(warnings, fmt.Sprintf("[Quota] Expected %s available to be %d%s, but only %d%s is available", resName, resVal, resInfo.Units, infraAvailableResVal, resInfo.Units))
			}
		}
	}

	return warnings, nil
}

// getAllCloudletResources
// Returns (1) All the VM resources on the cloudlet (2) Diff of VM resources reported by CRM and seen by controller
func getAllCloudletResources(ctx context.Context, stm concurrency.STM, cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo, cloudletRefs *edgeproto.CloudletRefs) ([]edgeproto.VMResource, []edgeproto.VMResource, error) {
	allVmResources := []edgeproto.VMResource{}
	diffVmResources := []edgeproto.VMResource{}
	// get all cloudlet resources (platformVM, sharedRootLB, etc)
	cloudletRes, err := GetCloudletResources(ctx, cloudletInfo)
	if err != nil {
		return nil, nil, err
	}
	allVmResources = append(allVmResources, cloudletRes...)

	snapshotClusters := make(map[edgeproto.ClusterInstRefKey]struct{})
	for _, clKey := range cloudletInfo.ResourcesSnapshot.ClusterInsts {
		snapshotClusters[clKey] = struct{}{}
	}

	snapshotVmAppInsts := make(map[edgeproto.AppInstRefKey]struct{})
	for _, aiKey := range cloudletInfo.ResourcesSnapshot.VmAppInsts {
		snapshotVmAppInsts[aiKey] = struct{}{}
	}

	// get all cluster resources (clusterVM, dedicatedRootLB, etc)
	clusterInstKeys := cloudletRefs.ClusterInsts
	for _, clusterInstRefKey := range clusterInstKeys {
		ci := edgeproto.ClusterInst{}
		clusterInstKey := edgeproto.ClusterInstKey{}
		clusterInstKey.FromClusterInstRefKey(&clusterInstRefKey, &cloudletRefs.Key)
		if !clusterInstApi.store.STMGet(stm, &clusterInstKey, &ci) {
			continue
		}
		if edgeproto.IsDeleteState(ci.State) {
			continue
		}
		allRes, err := getClusterInstVMRequirements(ctx, stm, &ci, cloudlet, cloudletInfo, cloudletRefs)
		if err != nil {
			return nil, nil, err
		}
		allVmResources = append(allVmResources, allRes...)

		// maintain a diff of clusterinsts reported by CRM and what is present in controller,
		// this is done to get accurate resource information
		clRefKey := &edgeproto.ClusterInstRefKey{}
		clRefKey.FromClusterInstKey(&clusterInstKey)
		if _, ok := snapshotClusters[*clRefKey]; ok {
			continue
		}
		diffVmResources = append(diffVmResources, allRes...)
	}
	// get all VM app inst resources
	for _, appInstRefKey := range cloudletRefs.VmAppInsts {
		appInst := edgeproto.AppInst{}
		appInstKey := edgeproto.AppInstKey{}
		appInstKey.FromAppInstRefKey(&appInstRefKey, &cloudlet.Key)
		if !appInstApi.store.STMGet(stm, &appInstKey, &appInst) {
			continue
		}
		if edgeproto.IsDeleteState(appInst.State) {
			continue
		}
		app := edgeproto.App{}
		if !appApi.store.STMGet(stm, &appInstKey.AppKey, &app) {
			return nil, nil, fmt.Errorf("App not found: %v", appInstKey.AppKey)
		}
		vmRes, err := cloudcommon.GetVMAppRequirements(ctx, &app, &appInst, cloudletInfo.Flavors)
		if err != nil {
			return nil, nil, err
		}
		allVmResources = append(allVmResources, vmRes...)

		// maintain a diff of VM appinsts reported by CRM and what is present in controller,
		// this is done to get accurate resource information
		aiRefKey := &edgeproto.AppInstRefKey{}
		aiRefKey.FromAppInstKey(&appInstKey)
		if _, ok := snapshotVmAppInsts[*aiRefKey]; !ok {
			continue
		}
		diffVmResources = append(diffVmResources, vmRes...)
	}
	return allVmResources, diffVmResources, nil
}

func validateResources(ctx context.Context, stm concurrency.STM, clusterInst *edgeproto.ClusterInst, vmAppInst *edgeproto.AppInst, cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo, cloudletRefs *edgeproto.CloudletRefs) error {
	reqdVmResources := []edgeproto.VMResource{}
	if clusterInst != nil {
		ciResources, err := getClusterInstVMRequirements(ctx, stm, clusterInst, cloudlet, cloudletInfo, cloudletRefs)
		if err != nil {
			return err
		}
		reqdVmResources = append(reqdVmResources, ciResources...)
	}
	if vmAppInst != nil {
		app := edgeproto.App{}
		if !appApi.store.STMGet(stm, &vmAppInst.Key.AppKey, &app) {
			return fmt.Errorf("App not found: %v", vmAppInst.Key.AppKey)
		}
		vmAppResources, err := cloudcommon.GetVMAppRequirements(ctx, &app, vmAppInst, cloudletInfo.Flavors)
		if err != nil {
			return err
		}
		reqdVmResources = append(reqdVmResources, vmAppResources...)
	}

	// get all cloudlet resources (platformVM, sharedRootLB, clusterVms, AppVMs, etc)
	allVmResources, diffVmResources, err := getAllCloudletResources(ctx, stm, cloudlet, cloudletInfo, cloudletRefs)
	if err != nil {
		return err
	}

	warnings, err := validateCloudletCommonResources(ctx, stm, cloudlet, allVmResources, reqdVmResources)
	if err != nil {
		return err
	}
	infraWarnings, err := validateCloudletInfraResources(ctx, cloudlet, &cloudletInfo.ResourcesSnapshot, allVmResources, reqdVmResources, diffVmResources)
	if err != nil {
		return err
	}

	// generate alerts for these warnings
	// clear off those alerts which are no longer firing
	warnings = append(warnings, infraWarnings...)
	handleResourceUsageAlerts(ctx, stm, &cloudlet.Key, warnings)
	return nil
}

func getCloudletResourceMetric(ctx context.Context, stm concurrency.STM, key *edgeproto.CloudletKey) ([]*edgeproto.Metric, error) {
	cloudlet := edgeproto.Cloudlet{}
	if !cloudletApi.store.STMGet(stm, key, &cloudlet) {
		return nil, fmt.Errorf("Cloudlet not found: %v", key)
	}
	cloudletInfo := edgeproto.CloudletInfo{}
	if !cloudletInfoApi.store.STMGet(stm, key, &cloudletInfo) {
		return nil, fmt.Errorf("CloudletInfo not found: %v", key)
	}
	cloudletRefs := edgeproto.CloudletRefs{}
	if !cloudletRefsApi.store.STMGet(stm, key, &cloudletRefs) {
		return nil, fmt.Errorf("CloudletRefs not found: %v", key)
	}
	cloudletPlatform, err := pfutils.GetPlatform(ctx, cloudlet.PlatformType.String())
	if err != nil {
		return nil, err
	}
	pfType := cloudletPlatform.GetType()

	// get all cloudlet resources (platformVM, sharedRootLB, clusterVms, AppVMs, etc)
	allResources, _, err := getAllCloudletResources(ctx, stm, &cloudlet, &cloudletInfo, &cloudletRefs)
	if err != nil {
		return nil, err
	}

	ramUsed := uint64(0)
	vcpusUsed := uint64(0)
	diskUsed := uint64(0)
	for _, vmRes := range allResources {
		if vmRes.VmFlavor != nil {
			ramUsed += vmRes.VmFlavor.Ram
			vcpusUsed += vmRes.VmFlavor.Vcpus
			diskUsed += vmRes.VmFlavor.Disk
		}
	}

	resMetric := edgeproto.Metric{}
	ts, _ := types.TimestampProto(time.Now())
	resMetric.Timestamp = *ts
	resMetric.Name = fmt.Sprintf("%s-resource-usage", pfType)
	resMetric.AddTag("cloudletorg", key.Organization)
	resMetric.AddTag("cloudlet", key.Name)
	resMetric.AddIntVal("ramUsed", ramUsed)
	resMetric.AddIntVal("vcpusUsed", vcpusUsed)
	resMetric.AddIntVal("diskUsed", diskUsed)

	// get additional infra specific metric
	err = cloudletPlatform.GetClusterAdditionalResourceMetric(ctx, &cloudlet, &resMetric, allResources)
	if err != nil {
		return nil, err
	}

	// get cloudlet metric
	gpusUsed := uint64(0)
	flavorCount := make(map[string]uint64)
	for _, vmRes := range allResources {
		if vmRes.VmFlavor != nil {
			if resTagTableApi.UsesGpu(ctx, stm, *vmRes.VmFlavor, cloudlet) {
				gpusUsed += 1
			}
			if _, ok := flavorCount[vmRes.VmFlavor.Name]; ok {
				flavorCount[vmRes.VmFlavor.Name] += 1
			} else {
				flavorCount[vmRes.VmFlavor.Name] = 1
			}
		}
	}
	resMetric.AddIntVal("gpusUsed", gpusUsed)
	metrics := []*edgeproto.Metric{}
	metrics = append(metrics, &resMetric)

	for fName, fCount := range flavorCount {
		flavorMetric := edgeproto.Metric{}
		flavorMetric.Name = "cloudlet-flavor-usage"
		flavorMetric.Timestamp = *ts
		flavorMetric.AddTag("cloudletorg", cloudlet.Key.Organization)
		flavorMetric.AddTag("cloudlet", cloudlet.Key.Name)
		flavorMetric.AddTag("flavor", fName)
		flavorMetric.AddIntVal("count", fCount)
		metrics = append(metrics, &flavorMetric)
	}
	return metrics, nil
}

// createClusterInstInternal is used to create dynamic cluster insts internally,
// bypassing static assignment. It is also used to create auto-cluster insts.
func (s *ClusterInstApi) createClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, inCb edgeproto.ClusterInstApi_CreateClusterInstServer) (reterr error) {
	cctx.SetOverride(&in.CrmOverride)
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	ctx := inCb.Context()

	clusterInstKey := in.Key
	sendObj, cb, err := startClusterInstStream(ctx, &clusterInstKey, inCb)
	if err == nil {
		defer func() {
			stopClusterInstStream(ctx, &clusterInstKey, sendObj, reterr)
		}()
	}

	defer func() {
		if reterr == nil {
			RecordClusterInstEvent(cb.Context(), &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

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

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
			in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
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
			cloudlet.PlatformType != edgeproto.PlatformType_PLATFORM_TYPE_VCD &&
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
		if resTagTableApi.UsesGpu(ctx, stm, *vmspec.FlavorInfo, cloudlet) {
			in.OptRes = "gpu"
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

		err = validateResources(ctx, stm, in, nil, &cloudlet, &info, &refs)
		if err != nil {
			return err
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
		err = allocateIP(in, &cloudlet, cloudlet.PlatformType, &refs)
		if err != nil {
			return err
		}
		cRefKey := edgeproto.ClusterInstRefKey{}
		cRefKey.FromClusterInstKey(&in.Key)
		refs.ClusterInsts = append(refs.ClusterInsts, cRefKey)
		cloudletRefsApi.store.STMPut(stm, &refs)

		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())

		if ignoreCRM(cctx) {
			in.State = edgeproto.TrackedState_READY
		} else {
			in.State = edgeproto.TrackedState_CREATE_REQUESTED
		}
		in.Status = edgeproto.StatusInfo{}
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
	if err == nil {
		metrics := []*edgeproto.Metric{}
		resErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			metrics, err = getCloudletResourceMetric(ctx, stm, &in.Key.CloudletKey)
			return err
		})
		if resErr == nil {
			services.cloudletResourcesInfluxQ.AddMetric(metrics...)
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to generate cloudlet resource usage metric", "clusterInstKey", in.Key, "err", resErr)
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

func (s *ClusterInstApi) updateClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, inCb edgeproto.ClusterInstApi_DeleteClusterInstServer) (reterr error) {
	ctx := inCb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "updateClusterInstInternal")

	err := in.ValidateUpdateFields()
	if err != nil {
		return err
	}
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	cctx.SetOverride(&in.CrmOverride)

	clusterInstKey := in.Key
	sendObj, cb, err := startClusterInstStream(ctx, &clusterInstKey, inCb)
	if err == nil {
		defer func() {
			stopClusterInstStream(ctx, &clusterInstKey, sendObj, reterr)
		}()
	}

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

		// Get diff of nodes to validate and track cloudlet resources
		resChanged := false
		resClusterInst := &edgeproto.ClusterInst{}
		resClusterInst.DeepCopyIn(&inbuf)
		if resClusterInst.NumNodes > in.NumNodes {
			// update diff
			resClusterInst.NumNodes = in.NumNodes - resClusterInst.NumNodes
			resChanged = true
		}
		if resClusterInst.NumMasters > in.NumMasters {
			// update diff
			resClusterInst.NumMasters = in.NumMasters - resClusterInst.NumMasters
			resChanged = true
		}
		if resChanged {
			cloudlet := edgeproto.Cloudlet{}
			if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
				return errors.New("Specified Cloudlet not found")
			}
			info := edgeproto.CloudletInfo{}
			if !cloudletInfoApi.store.STMGet(stm, &in.Key.CloudletKey, &info) {
				return fmt.Errorf("No resource information found for Cloudlet %s", in.Key.CloudletKey)
			}
			cloudletRefs := edgeproto.CloudletRefs{}
			cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudletRefs)
			err = validateResources(ctx, stm, resClusterInst, nil, &cloudlet, &info, &cloudletRefs)
			if err != nil {
				return err
			}
		}

		changeCount = inbuf.CopyInFields(in)
		if changeCount == 0 && !retry {
			// nothing changed
			return nil
		}
		if err := validateClusterInstUpdates(ctx, stm, &inbuf); err != nil {
			return err
		}

		if !ignoreCRM(cctx) {
			inbuf.State = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		inbuf.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		inbuf.Status = edgeproto.StatusInfo{}
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
	if err == nil {
		metrics := []*edgeproto.Metric{}
		resErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			metrics, err = getCloudletResourceMetric(ctx, stm, &in.Key.CloudletKey)
			return err
		})
		if resErr == nil {
			services.cloudletResourcesInfluxQ.AddMetric(metrics...)
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to generate cloudlet resource usage metric", "clusterInstKey", in.Key, "err", resErr)
		}
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to generate cloudlet resource usage metric", "clusterInstKey", in.Key, "err", resErr)
	}
	return err
}

func validateClusterInstUpdates(ctx context.Context, stm concurrency.STM, in *edgeproto.ClusterInst) error {
	cloudlet := edgeproto.Cloudlet{}
	if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
		return errors.New("Specified Cloudlet not found")
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

func (s *ClusterInstApi) deleteClusterInstInternal(cctx *CallContext, in *edgeproto.ClusterInst, inCb edgeproto.ClusterInstApi_DeleteClusterInstServer) (reterr error) {
	log.SpanLog(inCb.Context(), log.DebugLevelApi, "delete ClusterInst internal", "key", in.Key)
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
	ctx := inCb.Context()

	clusterInstKey := in.Key
	sendObj, cb, err := startClusterInstStream(ctx, &clusterInstKey, inCb)
	if err == nil {
		defer func() {
			stopClusterInstStream(ctx, &clusterInstKey, sendObj, reterr)
		}()
	}

	var prevState edgeproto.TrackedState
	// Set state to prevent other apps from being created on ClusterInst
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
			return fmt.Errorf("ClusterInst busy (%s), cannot delete", in.State.String())
		}
		prevState = in.State
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		in.Status = edgeproto.StatusInfo{}
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
		if cloudlet.MaintenanceState != dme.MaintenanceState_NORMAL_OPERATION {
			return errors.New("Cloudlet under maintenance, please try again later")
		}
		refs := edgeproto.CloudletRefs{}
		if cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &refs) {
			ii := 0
			for ; ii < len(refs.ClusterInsts); ii++ {
				cKey := edgeproto.ClusterInstKey{}
				cKey.FromClusterInstRefKey(&refs.ClusterInsts[ii], &in.Key.CloudletKey)
				if cKey.Matches(&in.Key) {
					break
				}
			}
			if ii < len(refs.ClusterInsts) {
				// explicity zero out deleted item to
				// prevent memory leak
				a := refs.ClusterInsts
				copy(a[ii:], a[ii+1:])
				a[len(a)-1] = edgeproto.ClusterInstRefKey{}
				refs.ClusterInsts = a[:len(a)-1]
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
	if err == nil {
		metrics := []*edgeproto.Metric{}
		resErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			metrics, err = getCloudletResourceMetric(ctx, stm, &in.Key.CloudletKey)
			return err
		})
		if resErr == nil {
			services.cloudletResourcesInfluxQ.AddMetric(metrics...)
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to generate cloudlet resource usage metric", "clusterInstKey", in.Key, "err", resErr)
		}
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to generate cloudlet resource usage metric", "clusterInstKey", in.Key, "err", resErr)
	}
	return err
}

func (s *ClusterInstApi) ShowClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_ShowClusterInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterInst) error {
		obj.Status = edgeproto.StatusInfo{}
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
		if next == edgeproto.TrackedState_DELETING || next == edgeproto.TrackedState_NOT_PRESENT || next == edgeproto.TrackedState_DELETE_ERROR || next == edgeproto.TrackedState_DELETE_DONE {
			return true
		}
	case edgeproto.TrackedState_DELETING:
		if next == edgeproto.TrackedState_NOT_PRESENT || next == edgeproto.TrackedState_DELETE_ERROR || next == edgeproto.TrackedState_DELETE_DONE {
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
	log.SpanLog(ctx, log.DebugLevelApi, "update ClusterInst", "state", in.State, "status", in.Status, "resources", in.ResourcesSnapshot)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		inst.ResourcesSnapshot = in.ResourcesSnapshot
		// update only diff of status msgs
		edgeproto.UpdateStatusDiff(&in.Status, &inst.Status)
		if inst.State == in.State {
			log.SpanLog(ctx, log.DebugLevelApi, "no state change")
			s.store.STMPut(stm, &inst)
			return nil
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.SpanLog(ctx, log.DebugLevelApi, "invalid state transition", "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		if in.State == edgeproto.TrackedState_CREATE_ERROR || in.State == edgeproto.TrackedState_DELETE_ERROR || in.State == edgeproto.TrackedState_UPDATE_ERROR {
			inst.Errors = in.Errors
		} else {
			inst.Errors = nil
		}
		s.store.STMPut(stm, &inst)
		return nil
	})
	if in.State == edgeproto.TrackedState_DELETE_DONE {
		s.DeleteFromInfo(ctx, in)
	}
}

func (s *ClusterInstApi) DeleteFromInfo(ctx context.Context, in *edgeproto.ClusterInstInfo) {
	log.SpanLog(ctx, log.DebugLevelApi, "delete ClusterInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.ClusterInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_DELETING && inst.State != edgeproto.TrackedState_DELETE_REQUESTED && inst.State != edgeproto.TrackedState_DELETE_DONE {
			log.SpanLog(ctx, log.DebugLevelApi, "invalid state transition", "cur", inst.State, "next", edgeproto.TrackedState_DELETE_DONE)
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
	// org field so that influx queries are a lot simpler to retrieve reserved clusters
	if info.ReservedBy != "" {
		metric.AddTag("org", info.ReservedBy)
	} else {
		metric.AddTag("org", clusterInstKey.Organization)
	}
	// errors should never happen here since to get to this point the flavor should have already been checked previously, but just in case
	nodeFlavor := edgeproto.Flavor{}
	if !flavorApi.cache.Get(&info.Flavor, &nodeFlavor) {
		log.SpanLog(ctx, log.DebugLevelMetrics, "flavor not found for recording clusterInst lifecycle", "flavor name", info.Flavor.Name)
	} else {
		metric.AddTag("flavor", info.Flavor.Name)
		metric.AddIntVal("ram", nodeFlavor.Ram)
		metric.AddIntVal("vcpu", nodeFlavor.Vcpus)
		metric.AddIntVal("disk", nodeFlavor.Disk)
		metric.AddIntVal("nodecount", uint64(info.NumMasters+info.NumNodes))
		metric.AddStringVal("other", fmt.Sprintf("%v", nodeFlavor.OptResMap))
	}
	metric.AddStringVal("ipaccess", info.IpAccess.String())

	services.events.AddMetric(&metric)
}
