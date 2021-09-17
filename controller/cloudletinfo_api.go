package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util/tasks"
)

type CloudletInfoApi struct {
	sync                      *Sync
	store                     edgeproto.CloudletInfoStore
	cache                     edgeproto.CloudletInfoCache
	clearInfraFlavorAlertTask tasks.KeyWorkers
	alertReaperCreated        bool // if an infra flavor is never deleted we never init this worker task
}

var cloudletInfoApi = CloudletInfoApi{}
var cleanupCloudletInfoTimeout = 5 * time.Minute

func InitCloudletInfoApi(sync *Sync) {
	cloudletInfoApi.sync = sync
	cloudletInfoApi.store = edgeproto.NewCloudletInfoStore(sync.store)
	edgeproto.InitCloudletInfoCache(&cloudletInfoApi.cache)
	sync.RegisterCache(&cloudletInfoApi.cache)
}

// Move?
type HandleFlavorAlertWorkerKey struct {
	cloudletKey edgeproto.CloudletKey
	flavor      string
}

// We put CloudletInfo in etcd with a lease, so in case both controller
// and CRM suddenly go away, etcd will remove the stale CloudletInfo data.

func (s *CloudletInfoApi) InjectCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return s.store.Put(ctx, in, s.sync.syncWait)
}

func (s *CloudletInfoApi) EvictCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

func (s *CloudletInfoApi) ShowCloudletInfo(in *edgeproto.CloudletInfo, cb edgeproto.CloudletInfoApi_ShowCloudletInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletInfo) error {
		obj.Status = edgeproto.StatusInfo{}
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletInfoApi) Update(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	var err error
	// for now assume all fields have been specified
	in.Fields = edgeproto.CloudletInfoAllFields
	in.Controller = ControllerId
	changedToOnline := false
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		info := edgeproto.CloudletInfo{}
		if s.store.STMGet(stm, &in.Key, &info) {
			if in.State == dme.CloudletState_CLOUDLET_STATE_READY &&
				info.State != dme.CloudletState_CLOUDLET_STATE_READY {
				changedToOnline = true
			}
		}
		in.Flavors = s.HandleInfraFlavorDeltas(ctx, in, info.Flavors)
		s.store.STMPut(stm, in)
		return nil
	})

	cloudlet := edgeproto.Cloudlet{}
	if !cloudletApi.cache.Get(&in.Key, &cloudlet) {
		return
	}
	if changedToOnline {
		nodeMgr.Event(ctx, "Cloudlet online", in.Key.Organization, in.Key.GetTags(), nil, "state", in.State.String(), "version", in.ContainerVersion)
		features, err := GetCloudletFeatures(ctx, cloudlet.PlatformType)
		if err == nil {
			if features.SupportsMultiTenantCluster && cloudlet.EnableDefaultServerlessCluster {
				go createDefaultMultiTenantCluster(ctx, in.Key)
			}
		}
	}

	newState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
	switch in.State {
	case dme.CloudletState_CLOUDLET_STATE_INIT:
		newState = edgeproto.TrackedState_CRM_INITOK
		if in.ContainerVersion != cloudlet.ContainerVersion {
			nodeMgr.Event(ctx, "Upgrading cloudlet", in.Key.Organization, in.Key.GetTags(), nil, "from-version", cloudlet.ContainerVersion, "to-version", in.ContainerVersion)
		}
	case dme.CloudletState_CLOUDLET_STATE_READY:
		newState = edgeproto.TrackedState_READY
	case dme.CloudletState_CLOUDLET_STATE_UPGRADE:
		newState = edgeproto.TrackedState_UPDATING
	case dme.CloudletState_CLOUDLET_STATE_ERRORS:
		if cloudlet.State == edgeproto.TrackedState_UPDATE_REQUESTED ||
			cloudlet.State == edgeproto.TrackedState_UPDATING {
			newState = edgeproto.TrackedState_UPDATE_ERROR
		} else if cloudlet.State == edgeproto.TrackedState_CREATE_REQUESTED ||
			cloudlet.State == edgeproto.TrackedState_CREATING {
			newState = edgeproto.TrackedState_CREATE_ERROR
		}
	default:
		log.SpanLog(ctx, log.DebugLevelNotify, "Skip cloudletInfo state handling", "key", in.Key, "state", in.State)
		return
	}

	// update only diff of status msgs
	streamKey := edgeproto.GetStreamKeyFromCloudletKey(&in.Key)
	streamObjApi.UpdateStatus(ctx, &in.Status, &streamKey)

	newCloudlet := edgeproto.Cloudlet{}
	key := &in.Key
	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		updateObj := false
		if !cloudletApi.store.STMGet(stm, key, &newCloudlet) {
			return key.NotFoundError()
		}
		if newCloudlet.TrustPolicyState != in.TrustPolicyState && in.TrustPolicyState != edgeproto.TrackedState_TRACKED_STATE_UNKNOWN {
			newCloudlet.TrustPolicyState = in.TrustPolicyState
			updateObj = true
		}
		if newCloudlet.State != newState {
			newCloudlet.State = newState
			if in.Errors != nil {
				newCloudlet.Errors = in.Errors
			}
			if in.State == dme.CloudletState_CLOUDLET_STATE_READY {
				newCloudlet.Errors = nil
			}
			updateObj = true
		}
		if newCloudlet.ContainerVersion != in.ContainerVersion {
			newCloudlet.ContainerVersion = in.ContainerVersion
			updateObj = true
		}
		if updateObj {
			cloudletApi.store.STMPut(stm, &newCloudlet)
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "CloudletInfo state transition error",
			"key", in.Key, "err", err)
	}
	if in.State == dme.CloudletState_CLOUDLET_STATE_READY {
		ClearCloudletAndAppInstDownAlerts(ctx, in)
		// Validate cloudlet resources and generate appropriate warnings
		err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if !cloudletApi.store.STMGet(stm, key, &cloudlet) {
				return key.NotFoundError()
			}
			cloudletRefs := edgeproto.CloudletRefs{}
			cloudletRefsApi.store.STMGet(stm, key, &cloudletRefs)
			return validateResources(ctx, stm, nil, nil, nil, &cloudlet, in, &cloudletRefs, GenResourceAlerts)
		})
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "Failed to validate cloudlet resources",
				"key", in.Key, "err", err)
		}
	}
}

func (s *CloudletInfoApi) Del(ctx context.Context, key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletInfo{Key: *key}
	s.store.Delete(ctx, &in, wait)
}

func cloudletInfoToAlertLabels(in *edgeproto.CloudletInfo) map[string]string {
	labels := make(map[string]string)
	// Set tags that match cloudlet
	labels["alertname"] = cloudcommon.AlertCloudletDown
	labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeCloudlet
	labels[edgeproto.CloudletKeyTagName] = in.Key.Name
	labels[edgeproto.CloudletKeyTagOrganization] = in.Key.Organization
	return labels
}

func cloudletDownAppInstAlertLabels(appInstKey *edgeproto.AppInstKey) map[string]string {
	labels := appInstKey.GetTags()
	labels["alertname"] = cloudcommon.AlertAppInstDown
	labels[cloudcommon.AlertHealthCheckStatus] = strconv.Itoa(int(dme.HealthCheck_HEALTH_CHECK_CLOUDLET_OFFLINE))
	labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeApp
	return labels
}

// Raise the alarm when the cloudlet goes down
func fireCloudletDownAlert(ctx context.Context, in *edgeproto.CloudletInfo) {
	alert := edgeproto.Alert{}
	alert.State = "firing"
	alert.ActiveAt = dme.Timestamp{}
	ts := time.Now()
	alert.ActiveAt.Seconds = ts.Unix()
	alert.ActiveAt.Nanos = int32(ts.Nanosecond())
	alert.Labels = cloudletInfoToAlertLabels(in)
	alert.Annotations = make(map[string]string)
	alert.Annotations[cloudcommon.AlertAnnotationTitle] = cloudcommon.AlertCloudletDown
	alert.Annotations[cloudcommon.AlertAnnotationDescription] = cloudcommon.AlertCloudletDownDescription
	// send an update
	alertApi.Update(ctx, &alert, 0)
}

func FireCloudletAndAppInstDownAlerts(ctx context.Context, in *edgeproto.CloudletInfo) {
	fireCloudletDownAlert(ctx, in)
	fireCloudletDownAppInstAlerts(ctx, in)
}

func ClearCloudletAndAppInstDownAlerts(ctx context.Context, in *edgeproto.CloudletInfo) {
	clearCloudletDownAlert(ctx, in)
	clearCloudletDownAppInstAlerts(ctx, in)
}

func clearCloudletDownAlert(ctx context.Context, in *edgeproto.CloudletInfo) {
	alert := edgeproto.Alert{}
	alert.Labels = cloudletInfoToAlertLabels(in)
	alertApi.Delete(ctx, &alert, 0)
}

func clearCloudletDownAppInstAlerts(ctx context.Context, in *edgeproto.CloudletInfo) {
	appInstFilter := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{ClusterInstKey: edgeproto.VirtualClusterInstKey{CloudletKey: in.Key}},
	}
	appInstKeys := make([]edgeproto.AppInstKey, 0)
	appInstApi.cache.Show(&appInstFilter, func(obj *edgeproto.AppInst) error {
		appInstKeys = append(appInstKeys, obj.Key)
		return nil
	})
	for _, k := range appInstKeys {
		alert := edgeproto.Alert{}
		alert.Labels = cloudletDownAppInstAlertLabels(&k)
		alertApi.Delete(ctx, &alert, 0)
	}
}

func fireCloudletDownAppInstAlerts(ctx context.Context, in *edgeproto.CloudletInfo) {
	appInstFilter := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{ClusterInstKey: edgeproto.VirtualClusterInstKey{CloudletKey: in.Key}},
	}
	appInstKeys := make([]edgeproto.AppInstKey, 0)
	appInstApi.cache.Show(&appInstFilter, func(obj *edgeproto.AppInst) error {
		appInstKeys = append(appInstKeys, obj.Key)
		return nil
	})
	// exclude SideCar apps which are auto-deployed as part of the cluster
	excludedAppFilter := cloudcommon.GetSideCarAppFilter()
	excludedAppKeys := make(map[edgeproto.AppKey]bool, 0)
	appApi.cache.Show(excludedAppFilter, func(obj *edgeproto.App) error {
		excludedAppKeys[obj.Key] = true
		return nil
	})
	for _, k := range appInstKeys {
		if excluded := excludedAppKeys[k.AppKey]; excluded {
			continue
		}
		alert := edgeproto.Alert{}
		alert.State = "firing"
		alert.ActiveAt = dme.Timestamp{}
		ts := time.Now()
		alert.ActiveAt.Seconds = ts.Unix()
		alert.ActiveAt.Nanos = int32(ts.Nanosecond())
		alert.Labels = cloudletDownAppInstAlertLabels(&k)
		alert.Annotations = make(map[string]string)
		alert.Annotations[cloudcommon.AlertAnnotationTitle] = cloudcommon.AlertAppInstDown
		alert.Annotations[cloudcommon.AlertAnnotationDescription] = "AppInst down due to cloudlet offline"
		alertApi.Update(ctx, &alert, 0)
	}
}

// Delete from notify just marks the cloudlet offline
func (s *CloudletInfoApi) Delete(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	buf := edgeproto.CloudletInfo{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &buf) {
			return nil
		}
		if buf.NotifyId != in.NotifyId || buf.Controller != ControllerId {
			// updated by another thread or controller
			return nil
		}
		buf.State = dme.CloudletState_CLOUDLET_STATE_OFFLINE
		buf.Fields = []string{edgeproto.CloudletInfoFieldState}
		s.store.STMPut(stm, &buf)
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "notify delete CloudletInfo",
			"key", in.Key, "err", err)
	} else {
		nodeMgr.Event(ctx, "Cloudlet offline", in.Key.Organization, in.Key.GetTags(), nil, "reason", "notify disconnect")
	}
}

func (s *CloudletInfoApi) Flush(ctx context.Context, notifyId int64) {
	// mark all cloudlets from the client as offline
	matches := make([]edgeproto.CloudletKey, 0)
	s.cache.Mux.Lock()
	for _, data := range s.cache.Objs {
		val := data.Obj
		if val.NotifyId != notifyId || val.Controller != ControllerId {
			continue
		}
		matches = append(matches, val.Key)
	}
	s.cache.Mux.Unlock()

	// this creates a new span if there was none - which can happen if this is a cancelled context
	span := log.SpanFromContext(ctx)
	ectx := log.ContextWithSpan(context.Background(), span)

	info := edgeproto.CloudletInfo{}
	for ii, _ := range matches {
		info.Key = matches[ii]
		cloudletReady := false
		err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if s.store.STMGet(stm, &info.Key, &info) {
				if info.NotifyId != notifyId || info.Controller != ControllerId {
					// Updated by another thread or controller
					return nil
				}
			}
			cloudlet := edgeproto.Cloudlet{}
			if cloudletApi.store.STMGet(stm, &info.Key, &cloudlet) {
				cloudletReady = (cloudlet.State == edgeproto.TrackedState_READY)
			}
			info.State = dme.CloudletState_CLOUDLET_STATE_OFFLINE
			log.SpanLog(ctx, log.DebugLevelNotify, "mark cloudlet offline", "key", matches[ii], "notifyid", notifyId)
			s.store.STMPut(stm, &info)
			return nil
		})
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "mark cloudlet offline", "key", matches[ii], "err", err)
		} else {
			nodeMgr.Event(ectx, "Cloudlet offline", info.Key.Organization, info.Key.GetTags(), nil, "reason", "notify disconnect")
			// Send a cloudlet down alert if a cloudlet was ready
			if cloudletReady {
				FireCloudletAndAppInstDownAlerts(ctx, &info)
			}
		}
	}
}

func (s *CloudletInfoApi) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {}

func (s *CloudletInfoApi) getCloudletState(key *edgeproto.CloudletKey) dme.CloudletState {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		obj := data.Obj
		if key.Matches(&obj.Key) {
			return obj.State
		}
	}
	return dme.CloudletState_CLOUDLET_STATE_NOT_PRESENT
}

func checkCloudletReady(cctx *CallContext, stm concurrency.STM, key *edgeproto.CloudletKey, action cloudcommon.Action) error {
	if cctx != nil && ignoreCRM(cctx) {
		return nil
	}
	// Get tracked state, it could be that cloudlet has initiated
	// upgrade but cloudletInfo is yet to transition to it
	cloudlet := edgeproto.Cloudlet{}
	if !cloudletApi.store.STMGet(stm, key, &cloudlet) {
		return key.NotFoundError()
	}
	if action == cloudcommon.Delete && cloudlet.State == edgeproto.TrackedState_DELETE_PREPARE {
		return nil
	}
	if cloudlet.State == edgeproto.TrackedState_UPDATE_REQUESTED ||
		cloudlet.State == edgeproto.TrackedState_UPDATING {
		return fmt.Errorf("Cloudlet %v is upgrading", key)
	}
	if cloudlet.MaintenanceState != dme.MaintenanceState_NORMAL_OPERATION {
		return errors.New("Cloudlet under maintenance, please try again later")
	}

	if cloudlet.State == edgeproto.TrackedState_UPDATE_ERROR {
		return fmt.Errorf("Cloudlet %v is in failed upgrade state, please upgrade it manually", key)
	}
	info := edgeproto.CloudletInfo{}
	if !cloudletInfoApi.store.STMGet(stm, key, &info) {
		return key.NotFoundError()
	}
	if info.State == dme.CloudletState_CLOUDLET_STATE_READY {
		return nil
	}
	return fmt.Errorf("Cloudlet %v not ready, state is %s", key,
		dme.CloudletState_name[int32(info.State)])
}

// Clean up CloudletInfo after Cloudlet delete.
// Only delete if state is Offline.
func (s *CloudletInfoApi) cleanupCloudletInfo(ctx context.Context, key *edgeproto.CloudletKey) {
	done := make(chan bool, 1)
	info := edgeproto.CloudletInfo{}
	checkState := func() {
		if !s.cache.Get(key, &info) {
			done <- true
			return
		}
		if info.State == dme.CloudletState_CLOUDLET_STATE_OFFLINE {
			done <- true
		}
	}
	cancel := s.cache.WatchKey(key, func(ctx context.Context) {
		checkState()
	})
	defer cancel()
	// after setting up watch, check current state,
	// as it may have already changed to target state
	checkState()

	select {
	case <-done:
	case <-time.After(cleanupCloudletInfoTimeout):
		log.SpanLog(ctx, log.DebugLevelApi, "timed out waiting for CloudletInfo to go Offline", "key", key)
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		info := edgeproto.CloudletInfo{}
		if !s.store.STMGet(stm, key, &info) {
			return nil
		}
		if info.State != dme.CloudletState_CLOUDLET_STATE_OFFLINE {
			return fmt.Errorf("could not delete CloudletInfo, state is %s instead of offline", info.State.String())
		}
		s.store.STMDel(stm, key)
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "cleanup CloudletInfo failed", "err", err)
	}
	// clean up any associated alerts with this cloudlet
	ClearCloudletAndAppInstDownAlerts(ctx, &info)
}

func (s *CloudletInfoApi) waitForMaintenanceState(ctx context.Context, key *edgeproto.CloudletKey, targetState, errorState dme.MaintenanceState, timeout time.Duration, result *edgeproto.CloudletInfo) error {
	done := make(chan bool, 1)
	check := func(ctx context.Context) {
		if !s.cache.Get(key, result) {
			log.SpanLog(ctx, log.DebugLevelApi, "wait for CloudletInfo state info not found", "key", key)
			return
		}
		if result.MaintenanceState == targetState || result.MaintenanceState == errorState {
			done <- true
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "wait for CloudletInfo state", "target", targetState)

	cancel := s.cache.WatchKey(key, check)

	// after setting up watch, check current state,
	// as it may have already changed to the target state
	check(ctx)

	var err error
	select {
	case <-done:
	case <-time.After(timeout):
		err = fmt.Errorf("timed out waiting for CloudletInfo maintenance state")
	}
	cancel()

	return err
}

func getCloudletPropertyBool(info *edgeproto.CloudletInfo, prop string, def bool) bool {
	if info.Properties == nil {
		return def
	}
	str, found := info.Properties[prop]
	if !found {
		return def
	}
	val, err := strconv.ParseBool(str)
	if err != nil {
		return def
	}
	return val
}

var pendingAlerts = 0

func ClearDeletedInfraFlavorAlert(ctx context.Context, info *edgeproto.CloudletInfo, flavor string /* *edgeproto.FlavorInfo*/) {

	log.SpanLog(ctx, log.DebugLevelInfra, "clear alert for recreated", "flavor", flavor)
	fmt.Printf("\n\nClearDeletedInfraFlavorAlert-I-Begin flavor %s cloudlentInfo %s", flavor, info.Key.Name)
	alert := edgeproto.Alert{}
	alert.Labels = make(map[string]string)
	alert.Labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeCloudlet
	alert.Labels["alertname"] = cloudcommon.AlertInfraFlavorDeleted
	alert.Labels["cloudlet"] = info.Key.Name
	alert.Labels["cloudletorg"] = info.Key.Organization
	alert.Labels["infraflavor"] = flavor
	alertApi.Delete(ctx, &alert, 0)
	pendingAlerts--
	fmt.Printf("\n\n[pending %d] ClearDeletedInfraFlavorAlert-I-Done flavor %s cloudlet %s\n\n", pendingAlerts, flavor, info.Key.Name)
}

func RaiseDeletedInfraFlavorAlert(ctx context.Context, info *edgeproto.CloudletInfo, flavor string) {
	log.SpanLog(ctx, log.DebugLevelInfra, "raise alert for deleted infra", "flavor", flavor)
	alert := edgeproto.Alert{}
	alert.State = "firing"
	alert.ActiveAt = dme.Timestamp{}
	ts := time.Now()
	alert.ActiveAt.Seconds = ts.Unix()
	alert.ActiveAt.Nanos = int32(ts.Nanosecond())
	alert.Labels = make(map[string]string)
	alert.Labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeCloudlet
	alert.Labels["alertname"] = cloudcommon.AlertInfraFlavorDeleted
	alert.Labels["cloudlet"] = info.Key.Name
	alert.Labels["cloudletorg"] = info.Key.Organization
	alert.Labels["infraflavor"] = flavor
	// Probably don't need these annotations anymore xxx
	alert.Annotations = make(map[string]string)
	alert.Annotations["infraflavor"] = flavor
	alert.Annotations[cloudcommon.AlertAnnotationTitle] = cloudcommon.AlertInfraFlavorDeleted
	alert.Annotations[cloudcommon.AlertAnnotationDescription] = cloudcommon.AlertInfraFlavorDeletedDescription
	alertApi.Update(ctx, &alert, 0)
	pendingAlerts++
	fmt.Printf("\n\n[pendingAlerts %d] RaiseDeletedInfraFlavorAlert-I-Done flavor %s cloudlet %s\n\n", pendingAlerts, flavor, info.Key.Name)
}

// Given all the flavors in the in.Flavors from the cloudletInfo Update, return a list of missing flavors.
// That is, flavors in use that are not found on info.Flavor list.
// With the current signature, we can only return a list of flavor names. These names should be found
// on the deletedFlavors list, and are the ones that can't be allowed to be simply deleted, rather marked
// deprecated and added in to the in.Flavor list as such. A re-created flavor will naturally clear its deprecated flag.

// new 9/16 make getMissingFlavors return a map[string]int flavor name and the number of objects using the flavor
// doesn't work finding keys in a map

func (s *CloudletInfoApi) getMissingFlavors(ctx context.Context, info *edgeproto.CloudletInfo) ([]string, map[string]int, error) {

	fmt.Printf("\n----------------getMissingFlavors-I-Begin\n")

	var missingFlavors []string
	missingFlavorsMap := make(map[string]struct{})
	flavRefMap := make(map[string]int)

	inFlavorMap := make(map[string]*edgeproto.FlavorInfo)
	for _, f := range info.Flavors {
		inFlavorMap[f.Name] = f
	}
	cloudRefs := edgeproto.CloudletRefs{}
	//	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
	if !cloudletRefsApi.cache.Get(&info.Key, &cloudRefs) {
		fmt.Printf("\n\n-getMissingFlavors-E-couldn't find cloudletRefs for %+v\n\n", info.Key)
		return missingFlavors, flavRefMap, fmt.Errorf("CloudletRefs not found")
	}

	for _, clust := range cloudRefs.ClusterInsts {
		clusterInst := edgeproto.ClusterInst{}
		clusterInstKey := edgeproto.ClusterInstKey{
			ClusterKey:   clust.ClusterKey,
			CloudletKey:  info.Key,
			Organization: clust.Organization,
		}
		found := clusterInstApi.cache.Get(&clusterInstKey, &clusterInst)
		if found {
			count := 0
			fmt.Printf("\tgetMissingFlavors-I-found clusterInst %+v\n", clusterInst.Key)
			if _, found := inFlavorMap[clusterInst.NodeFlavor]; !found {
				count++
				missingFlavorsMap[clusterInst.NodeFlavor] = struct{}{}
				flavRefMap[clusterInst.NodeFlavor] = count
			}
			if _, found := inFlavorMap[clusterInst.MasterNodeFlavor]; !found {
				missingFlavorsMap[clusterInst.MasterNodeFlavor] = struct{}{}
				flavRefMap[clusterInst.MasterNodeFlavor] = count
			}
		}
		for _, appInstRefKey := range cloudRefs.VmAppInsts {
			appInst := edgeproto.AppInst{}
			appInstKey := edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Organization: appInstRefKey.AppKey.Organization,
					Name:         appInstRefKey.AppKey.Name,
					Version:      appInstRefKey.AppKey.Version,
				},
				ClusterInstKey: edgeproto.VirtualClusterInstKey{
					ClusterKey:   clust.ClusterKey,
					CloudletKey:  info.Key,
					Organization: clusterInst.Key.Organization,
				},
			}
			fmt.Printf("\tgetMissingFlavors-I-found appInst %+v\n", appInst.Key)
			found := appInstApi.cache.Get(&appInstKey, &appInst)
			if found {
				count := 0
				if _, found := inFlavorMap[appInst.Flavor.Name]; !found {
					count++
					missingFlavorsMap[appInst.Flavor.Name] = struct{}{}
					flavRefMap[appInst.Flavor.Name] = count
				}
			}
		}
	}
	for f, i := range flavRefMap {
		fmt.Printf("\tnext missing flavor : %s use count: %d\n", f, i)
		missingFlavors = append(missingFlavors, f)
	}
	fmt.Printf("----------------getMissingFlavors-I-Done\n")
	return missingFlavors, flavRefMap, nil
}

func (cd *CloudletInfoApi) findFlavorDeltas(ctx context.Context, flavorMap, newFlavorMap map[string]*edgeproto.FlavorInfo) ([]edgeproto.FlavorInfo, []edgeproto.FlavorInfo, []edgeproto.FlavorInfo, []edgeproto.FlavorInfo) {

	fmt.Printf("\n--------------findFlavorDeltas-I-begin\n")

	addedFlavors := []edgeproto.FlavorInfo{}
	deletedFlavors := []edgeproto.FlavorInfo{}
	updatedFlavors := []edgeproto.FlavorInfo{}
	clearAlertFlavors := []edgeproto.FlavorInfo{}

	// logging only
	oldFlavorCount := len(flavorMap)
	newFlavorCount := len(newFlavorMap)

	if !reflect.DeepEqual(flavorMap, newFlavorMap) {
		for key, _ := range flavorMap {
			if _, ok := newFlavorMap[key]; !ok {
				// key has been deleted
				log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: deleting")
				deletedFlavors = []edgeproto.FlavorInfo{}
				// lookup each exsiting flavorMap in new flavors, if not found, it has been deleted
				for oldFlavorName, flavor := range flavorMap {
					if _, ok := newFlavorMap[oldFlavorName]; !ok {

						fmt.Printf("\t-------Delta: adding to deleted flavors: %+v -----------------------------\n", flavor)
						deletedFlavors = append(deletedFlavors, *flavor)

					}
				}
				for _, flavor := range deletedFlavors {
					delete(flavorMap, flavor.Name)
				}
			}
		}
		for key, _ := range newFlavorMap {
			if _, ok := flavorMap[key]; !ok {
				// key has been added
				addedFlavors = []edgeproto.FlavorInfo{}
				log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: adding")
				// lookup all new flavor  names in flavorMap, if its not found, its new
				for newFlavorName, flavor := range newFlavorMap {
					if _, ok := flavorMap[newFlavorName]; !ok {
						addedFlavors = append(addedFlavors, *flavor)
					}
				}
				for _, flavor := range addedFlavors {
					flavorMap[flavor.Name] = &flavor
				}
			}
		}
	}
	if len(flavorMap) != len(newFlavorMap) {
		//		log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh delta internal error")
		//		return addedFlavors, deletedFlavors, updatedFlavors, clearAlertFlavors
		panic("Delta unequal maps")
	}
	// assert maps len are equal
	// Now check struct updates, deep equal on maps doesn't check beyond len and keys
	for key, val := range newFlavorMap {
		fmt.Printf("\t ========================== Delta newFlavorMap[key] = val :  %+v   =====================\n", val)
		if key == "flavor.tiny2" {
			f1 := newFlavorMap[key]
			f2 := flavorMap[key]
			fmt.Printf("\t ======found tiny2 ============= Delta newFlavorMap[key] %+v  flavormap[key] %+v =====================\n", f1, f2)
		}
		// Updated flavor(s)
		if !reflect.DeepEqual(newFlavorMap[key], flavorMap[key]) {
			// special case: If flavorMap[key] is deprecated, we must clear the alert.
			flavor := flavorMap[key]

			if flavor.Deprecated {
				log.SpanLog(ctx, log.DebugLevelInfra, "deprecated infa flavor recreated", "flavor", flavor.Name)
				fmt.Printf("\n\n ######## flavor %s has been RECREATED, clear the existing alert!\n\n", flavor.Name)
				clearAlertFlavors = append(clearAlertFlavors, *val)
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: updating")
			// Flavor definition changed update
			updatedFlavors = append(updatedFlavors, *val)
		}
	}
	// recap
	log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh:", "old flavs:", oldFlavorCount, "cur flavs", newFlavorCount, "added", len(addedFlavors),
		"deleted", len(deletedFlavors), "updated", len(updatedFlavors))
	fmt.Printf("\t---------Delta: Done  len(deletedFlavors) = %d: \n", len(deletedFlavors))
	for _, f := range deletedFlavors {
		fmt.Printf("\t\tnext del:  %+v\n", f)
	}
	return addedFlavors, deletedFlavors, updatedFlavors, clearAlertFlavors
}

// Return a potentially modified flavor list to store in the db for this cloudlet flavor list during Update.
// It may include flavors from the deletedFlavors list that are currently in use, and those will be marked deprecated.

// This guy is going have to return a liset of missing flavors needing an alert

func (s *CloudletInfoApi) HandleInfraFlavorDeltas(ctx context.Context, in *edgeproto.CloudletInfo, curFlavors []*edgeproto.FlavorInfo) []*edgeproto.FlavorInfo {

	fmt.Printf("\n\n\tHandleInfraFlavorDeltas-I-begin\n\n")

	flavorMap := make(map[string]*edgeproto.FlavorInfo)
	newFlavorMap := make(map[string]*edgeproto.FlavorInfo)
	// recap logging only
	oldFlavorCount := len(curFlavors)
	newFlavorCount := len(in.Flavors)

	// whats in the db currently
	for _, flavor := range curFlavors {
		flavorMap[flavor.Name] = flavor
	}
	// whats in the CloudletInfo update (from crm: GatherCloudletInfo)
	for i, flavor := range in.Flavors {
		fmt.Printf("\t\tHandle-I-next new flavor[%d] in update =  %+v\n", i, flavor)
		newFlavorMap[flavor.Name] = flavor
	}
	// determine changes in flavor list
	addedFlavors, deletedFlavors, updatedFlavors, clearAlertsFlavors := s.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	if len(addedFlavors)+len(deletedFlavors)+len(updatedFlavors)+len(clearAlertsFlavors) == 0 {
		fmt.Printf("\n\n !!!! Handle: No Changes found return noop !!!! \n\n")
		log.SpanLog(ctx, log.DebugLevelInfra, "check OS flavor: no changes")
		return in.Flavors
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "check OS flavor: changes", "old flavs:", oldFlavorCount, "cur flavs", newFlavorCount, "added", len(addedFlavors),
		"deleted", len(deletedFlavors), "updated", len(updatedFlavors), "clearAlerts", len(clearAlertsFlavors))

	fmt.Printf("\tHANDLE back from Delta: len clearAlertsFlavoars %d  pendingAlerts: %d \n", len(clearAlertsFlavors), pendingAlerts)
	// Events and Alerts
	// clear any pending deleted flavor alerts: flavor recreated case)
	if len(clearAlertsFlavors) != 0 {
		for _, flavor := range clearAlertsFlavors {
			fmt.Printf("\tHandle: recreated alert clearing %s\n", flavor.Name)
			ClearDeletedInfraFlavorAlert(ctx, in, flavor.Name)
		}
	}

	if len(addedFlavors) != 0 {
		fmt.Printf("\tHandle  Adding flavors:\n")
		for _, f := range addedFlavors {
			fmt.Printf("\t%+v\n", f)
		}
		var vals string = ""
		for _, flavor := range addedFlavors {
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "Flavor Added", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
	}
	if len(updatedFlavors) != 0 {
		fmt.Printf("\tHandle  Updating  flavors:\n")
		for _, f := range updatedFlavors {
			fmt.Printf("\t%+v\n", f)
		}

		var vals string = ""
		for _, flavor := range updatedFlavors {
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "Flavor Updated", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
	}
	for _, flavor := range clearAlertsFlavors {
		fmt.Printf("\n\nDeprecated flavor %s has been recreated, clear alert\n\n", flavor.Name)
		ClearDeletedInfraFlavorAlert(ctx, in, flavor.Name)
	}
	fmt.Printf("\tHandle: we have %d deletedFlavors\n", len(deletedFlavors))

	if len(deletedFlavors) != 0 {
		delFlavorMap := make(map[string]edgeproto.FlavorInfo)
		for _, f := range in.DeletedFlavors {
			fmt.Printf("\tHandle next in.DeletedFlavors: %+v\n", f)
			delFlavorMap[f.Name] = *f
		}
		// now add in what we've just seen
		for _, f := range deletedFlavors {
			fmt.Printf("\tHandle adding what's just beeen deleted to delFlavaorMap: %+v\n", f)
			if f.Deprecated {
				fmt.Printf("\t\thandle skip deprecated flovor %+v don't include on new deleted flavors\n", f)
				continue
			}
			delFlavorMap[f.Name] = f
		}
		in.DeletedFlavors = nil
		fmt.Printf("\nHandle: wipe out in.DeletedFlavors len now: %d\n", len(in.DeletedFlavors))
		for _, f := range delFlavorMap {
			fmt.Printf("\tHandle adding delFlavorsMap to in.Deleteflavors :  next: %+v\n", f)
			in.DeletedFlavors = append(in.DeletedFlavors, &f)
		}

		fmt.Printf("\tHandle:Update will store these delFlavors from curent update (or should)\n")
		for _, f := range in.DeletedFlavors {
			fmt.Printf("\t\t%+v\n", f)
		}

		// in any case, emit event for delete flavors in use or no
		var vals string = ""
		for _, flavor := range deletedFlavors {
			fmt.Printf("\tHandle raise delete Event for %+v\n", flavor)
			vals = vals + ", " + flavor.Name
			nodeMgr.Event(ctx, "Flavor Deleted", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
		}

		/*
			missingFlavors, _ := s.getMissingFlavors(ctx, in)
			fmt.Printf("\tHandle: missing flavors len: %d\n", len(missingFlavors))
			for _, flavor := range missingFlavors {
				if _, found := delFlavorMap[flavor]; found {
					depFlavor := delFlavorMap[flavor]
					depFlavor.Deprecated = true
					in.Flavors = append(in.Flavors, &depFlavor)
					in.DeprecatedFlavors = append(in.DeprecatedFlavors, &depFlavor)
					// alert per depFlavor since they don't share a common resolution
					RaiseDeletedInfraFlavorAlert(ctx, in, &depFlavor)
					if !s.alertReaperCreated {
						fmt.Printf("\n\nHandle: first time, starting AlertReaper\n\n")
						s.clearInfraFlavorAlertTask.Init("infraFlavorsAlertReaper", s.InfraFlavorAlertCleanupTask)
						s.alertReaperCreated = true
					}
					// we've taken care of the Raise Alert inline, only Clear Alerts need a worker task
					// s.clearInfraFlavorAlertTask.NeedsWork(ctx, in)
				}
			}

		*/
		if !s.alertReaperCreated {
			fmt.Printf("\tHandle: first time seeing delete , starting AlertTAsk\n")
			s.clearInfraFlavorAlertTask.Init("infraFlavorsAlertReaper", s.InfraFlavorAlertCleanupTask)
			s.alertReaperCreated = true
		}

		workerKey := HandleFlavorAlertWorkerKey{
			cloudletKey: in.Key,
		}
		// finish fixing up this update in a worker task
		s.clearInfraFlavorAlertTask.NeedsWork(ctx, workerKey)
	}

	fmt.Printf("\tHandle in.Flavors returned to finish stage1 Update: \n")
	for _, f := range in.Flavors {
		fmt.Printf("\t\t%+v\n", f)
	}

	fmt.Printf("\tHandle in.Flavors returned to finish stage1 Update: \n")
	for _, f := range in.DeletedFlavors {
		fmt.Printf("\t\t%+v\n", f)
	}

	fmt.Printf("\n--------------Handle-I-done\n\n")

	return in.Flavors
}

// Used by objects using infra flavors when they are deleted to check if a DeletedInfraFlavor alert can now be cleared
// Also by the HandleInfraFlavorsDeltas to raise any new alerts. The clear for the recreated flavor case is inline above.
func (s *CloudletInfoApi) InfraFlavorAlertCleanupTask(ctx context.Context, k interface{}) {

	key, ok := k.(HandleFlavorAlertWorkerKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup Unexpected failure, key not CloudletKey", "key", key)
		return
	}
	// fetch cloudletInfo object
	cloudletInfo := edgeproto.CloudletInfo{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &key.cloudletKey, &cloudletInfo) {
			panic("couldnt find cloudlet") // XXX duh
			//return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "worker task stm failed", "error", err)
		return
	}
	// no flavor? Then we're fixing up that last cloudletinfo.Update, since something was detected as deleted
	if key.flavor == "" {
		fmt.Printf("\n------------TASk begin FIXUP Update ------------------\n")
		// we know we have deleted flavor(s) from the lastest cloudletInfo.Update.
		// Check case of raising a new Alert. How? A missing Flavor is found on the deleted  Flavor list.
		// This means the op deleted an infra flavor we are currently using.

		missingFlavors, refMap, _ := s.getMissingFlavors(ctx, &cloudletInfo) // remove refMap we dont use it here
		if len(missingFlavors) == 0 {
			fmt.Printf("\nTASK: No missing Flavors found\n\n")
			log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorAlertCleanupTask no missing flavors found")
		} else {
			fmt.Printf("\tTASK MissingFlavors:\n")
			for _, f := range missingFlavors {
				fmt.Printf("\t missingFlavor: %s refcnt: %d \n", f, refMap[f])
			}
		}
		missingFlavorsMap := make(map[string]*string)
		for _, f := range missingFlavors {
			missingFlavorsMap[f] = &f
		}
		delFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
		for _, f := range cloudletInfo.DeletedFlavors {
			fmt.Printf("tTASK: deletedFlavors on  CloudletInfo: %s\n", f.Name)
			delFlavorsMap[f.Name] = f
		}
		curFlavorMap := make(map[string]*edgeproto.FlavorInfo)
		for i, f := range cloudletInfo.Flavors {
			fmt.Printf("\tTASK: cloudletInfo.Flavors[%d]: %+v\n", i, f)
			curFlavorMap[f.Name] = f
		}

		for _, flavor := range missingFlavors {

			if _, found := delFlavorsMap[flavor]; found {
				fmt.Printf("\tTASK: next missingFlavor that is also on the cloudlets deletedFlavors list len %d  %+v MAKE DEPRECATED \n", len(delFlavorsMap), flavor)
				depFlavor := delFlavorsMap[flavor]
				depFlavor.Deprecated = true
				fmt.Printf("\t\tTASK Marked depFlavor as deprecated: %+v\n", depFlavor)
				// recreating our dummy mock flavor, deprecated based on what was already deleted in the update
				cloudletInfo.DeprecatedFlavors = append(cloudletInfo.DeprecatedFlavors, &edgeproto.FlavorInfo{
					Name:       flavor,
					Vcpus:      depFlavor.Vcpus,
					Ram:        depFlavor.Ram,
					Disk:       depFlavor.Disk,
					PropMap:    depFlavor.PropMap,
					Deprecated: true,
				})

				curFlavorMap[depFlavor.Name] = depFlavor
				// This alert can be cleared by the flavor being added back in a subsequest Update, or
				// when the last using object is deleted.
				fmt.Printf("\tTASK raising alert for %+v +++++++++++++++++++++++++++ \n\n", depFlavor)
				RaiseDeletedInfraFlavorAlert(ctx, &cloudletInfo, flavor)
			}
		} // we should probably take out the cloudletInfo.DeletedFlavorsList now
		// the only reason we have DeletedFlavors is to fix up this single Update, we can clear it now.
		cloudletInfo.DeletedFlavors = nil
		cloudletInfo.Flavors = nil
		fmt.Printf("\tTASK curFlavors back to cloudletInfo.Flavors:\n")
		for _, f := range curFlavorMap {
			fmt.Printf("\t\t%+v\n", f)
			cloudletInfo.Flavors = append(cloudletInfo.Flavors, f)
		}
		fmt.Printf("\n----------TASK Updatefix : Done len(cloudletInfo.DeletedFlavors) = %d\n\n", len(cloudletInfo.DeletedFlavors))
		// end of fixing up last update ?
	} else {
		fmt.Printf("\n-----InfrarFlavorAlertCleaupTask BEGIN  (WORKER TASK) running Deleted Obj for flavor: %s ------------------------------\n\n", key.flavor)

		// Here we've been called from app/cluster inst delete to see if we can now delete a pending alert
		// if ther are no DeletedInfraFlavorAlerts, just return.
		alertFound := false
		// xxx we should be able to create a key, and look it up. But it alwayes seems to exist xxx
		for k, _ := range alertApi.cache.Objs {
			if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
				alertFound = true
			}
		}
		if !alertFound {
			fmt.Printf("\n\nTASK cluster/app delted no alerts, no work return\n\n")
			return
		}
		// now all that's left is to see if the given flavor name is on the deprecated list.
		// If that's the case, we'll need to plow through getMissingFlavors to determine if this is the last reference
		// to that deprecated flavor
		curFlavorsMap := make(map[string]edgeproto.FlavorInfo)
		depFlavorsMap := make(map[string]edgeproto.FlavorInfo)
		for _, flav := range cloudletInfo.Flavors {
			curFlavorsMap[flav.Name] = *flav
		}
		for _, flav := range cloudletInfo.DeprecatedFlavors {
			depFlavorsMap[flav.Name] = *flav
		}
		for _, f := range depFlavorsMap {
			if f.Name == key.flavor {
				// possibly, this is the last useage
				missingFlavors, refMap, _ := s.getMissingFlavors(ctx, &cloudletInfo)
				if len(missingFlavors) == 0 {
					fmt.Printf("\nTASK: No missing Flavors found\n\n")
					log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorAlertCleanupTask no missing flavors found")
				}
				for _, flavor := range missingFlavors {
					// test is missingFlavor in curFlavors? (should not be)
					if _, found := curFlavorsMap[flavor]; !found {
						fmt.Printf("\n\nWorker sees deprecated Flavor %s in depFlavors list, CLEAR it's alert its usage count: %d \n\n", flavor, refMap[flavor])
						if refMap[flavor] <= 1 { // must be the last guy standing
							ClearDeletedInfraFlavorAlert(ctx, &cloudletInfo, flavor)
							// and remove it from the deprecated flavors list
							delete(depFlavorsMap, flavor)
						}
					}
				}

			}
		}
		cloudletInfo.DeprecatedFlavors = nil
		for _, v := range depFlavorsMap {
			cloudletInfo.DeprecatedFlavors = append(cloudletInfo.DeprecatedFlavors, &v)
		}
	}

	fmt.Printf("\nTASK about to store CloudletInfo, flavors:\n")
	for i, f := range cloudletInfo.Flavors {
		fmt.Printf("\t\ttask store Flavors[%d] = %+v\n", i, f)
	}

	fmt.Printf("\tTask: Store updated Cloudlet Info                \n\n")
	// store the results, downstream uses will still find the mock expected flavor marked depreated if they care to look.
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudletInfoApi.store.STMPut(stm, &cloudletInfo)
		return nil
	})
	if err != nil {
		panic("STM failed\n")
	}
	fmt.Printf("\n-----TASK DONE stored %d flavors \n", len(cloudletInfo.Flavors))
}
