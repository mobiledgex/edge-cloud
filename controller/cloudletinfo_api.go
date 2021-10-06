package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util/tasks"
)

type CloudletInfoApi struct {
	sync                 *Sync
	store                edgeproto.CloudletInfoStore
	cache                edgeproto.CloudletInfoCache
	infraFlavorAlertTask tasks.KeyWorkers
}

var cloudletInfoApi = CloudletInfoApi{}
var cleanupCloudletInfoTimeout = 5 * time.Minute

func InitCloudletInfoApi(sync *Sync) {
	cloudletInfoApi.sync = sync
	cloudletInfoApi.store = edgeproto.NewCloudletInfoStore(sync.store)
	edgeproto.InitCloudletInfoCache(&cloudletInfoApi.cache)
	sync.RegisterCache(&cloudletInfoApi.cache)
	cloudletInfoApi.infraFlavorAlertTask.Init("infraFlavorsAlertReaper", cloudletInfoApi.InfraFlavorAlertTask)
}

type HandleFlavorAlertWorkerKey struct {
	cloudletKey edgeproto.CloudletKey
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
	updateFlavors := in.Flavors
	added := make(map[string]*edgeproto.FlavorInfo)
	deleted := make(map[string]*edgeproto.FlavorInfo)
	updated := make(map[string]*edgeproto.FlavorInfo)
	recreated := make(map[string]*edgeproto.FlavorInfo)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		info := edgeproto.CloudletInfo{}
		if s.store.STMGet(stm, &in.Key, &info) {
			if in.State == dme.CloudletState_CLOUDLET_STATE_READY &&
				info.State != dme.CloudletState_CLOUDLET_STATE_READY {
				changedToOnline = true
			}
			added, deleted, updated, recreated = s.HandleInfraFlavorUpdate(ctx, in, updateFlavors, info.Flavors)
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if len(added)+len(deleted)+len(updated)+len(recreated) != 0 {
		s.fixupFlavorUpdate(ctx, in, added, deleted, updated, recreated)
	}
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

func newDeprecatedFlavorInUseAlert(key *edgeproto.CloudletKey, infraFlavor string) *edgeproto.Alert {
	alert := edgeproto.Alert{}
	alert.Labels = key.GetTags()
	alert.Labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeCloudlet
	alert.Labels["alertname"] = cloudcommon.AlertDeprecatedFlavorInUse
	alert.Labels["infraflavor"] = infraFlavor
	return &alert
}

func ClearDeletedInfraFlavorAlert(ctx context.Context, key *edgeproto.CloudletKey, flavor string) {
	log.SpanLog(ctx, log.DebugLevelInfra, "clear alert for", "flavor", flavor)
	alert := newDeprecatedFlavorInUseAlert(key, flavor)
	alertApi.Delete(ctx, alert, 0)
}

// Raise the alert, pass in the usage info in case someone is interested in what was  using this flavor
func RaiseDeletedInfraFlavorAlert(ctx context.Context, key *edgeproto.CloudletKey, flavor string, reasons string) {
	log.SpanLog(ctx, log.DebugLevelInfra, "raise alert for deleted infra", "flavor", flavor, "reason", reasons)
	alert := newDeprecatedFlavorInUseAlert(key, flavor)
	alert.Annotations = make(map[string]string)
	alert.Annotations["users"] = reasons
	alertApi.Update(ctx, alert, 0)
}

func (s *CloudletInfoApi) findFlavorDeltas(ctx context.Context, flavorMap, newFlavorMap map[string]*edgeproto.FlavorInfo) (added, deleted, updated, recreated map[string]*edgeproto.FlavorInfo) {
	addedFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
	deletedFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
	updatedFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
	recreatedFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
	sameNames := make(map[string]struct{})

	for key, flavor := range flavorMap {
		if _, ok := newFlavorMap[key]; !ok {
			if flavor.Deprecated {
				// flavor was deleted in the past and we're keeping around a deprecated copy
				// deleted flavors are only ones that are deleted by this update
				continue
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: deleted", "flavor", key)
			deletedFlavorsMap[flavor.Name] = flavor
		} else {
			sameNames[key] = struct{}{}
		}
	}
	for key, flavor := range newFlavorMap {
		if _, ok := flavorMap[key]; !ok {
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: added", "flavor", key)
			addedFlavorsMap[flavor.Name] = flavor
		}
	}
	for key, _ := range sameNames {
		oldFlavor := flavorMap[key]
		newFlavor := newFlavorMap[key]
		if !reflect.DeepEqual(oldFlavor, newFlavor) {
			if oldFlavor.Deprecated {
				log.SpanLog(ctx, log.DebugLevelInfra, "deprecated infa flavor recreated", "flavor", key)
				recreatedFlavorsMap[oldFlavor.Name] = oldFlavor
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: updated", "flavor", key)
			updatedFlavorsMap[newFlavor.Name] = newFlavor
		}
	}
	return addedFlavorsMap, deletedFlavorsMap, updatedFlavorsMap, recreatedFlavorsMap
}

func (s *CloudletInfoApi) HandleInfraFlavorUpdate(ctx context.Context, info *edgeproto.CloudletInfo, newFlavors, curFlavors []*edgeproto.FlavorInfo) (added, deleted, updated, recreated map[string]*edgeproto.FlavorInfo) {

	newFlavorsMap := make(map[string]*edgeproto.FlavorInfo) // newFlavors = new  in.Flavors
	curFlavorsMap := make(map[string]*edgeproto.FlavorInfo) // curFlavors = whats was in the db when the update came in
	for _, flavor := range curFlavors {
		curFlavorsMap[flavor.Name] = flavor
	}
	for _, flavor := range newFlavors {
		newFlavorsMap[flavor.Name] = flavor
	}
	addedFlavorsMap, deletedFlavorsMap, updatedFlavorsMap, recreatedFlavorsMap := s.findFlavorDeltas(ctx, curFlavorsMap, newFlavorsMap)
	if len(addedFlavorsMap)+len(deletedFlavorsMap)+len(updatedFlavorsMap)+len(recreatedFlavorsMap) == 0 {
		return addedFlavorsMap, deletedFlavorsMap, updatedFlavorsMap, recreatedFlavorsMap
	}
	// clear old deprecated flavors if they were recreated to avoid overwriting in newFlavorMap
	for _, f := range recreatedFlavorsMap {
		if flavor, found := curFlavorsMap[f.Name]; found {
			flavor.Deprecated = false
		}
	}
	// carry forward deprecated flavors missing from newFlavorMap
	for _, f := range curFlavorsMap {
		if f.Deprecated {
			newFlavorsMap[f.Name] = f
		}
	}
	// add newly deleted flavors as deprecated
	for _, f := range deletedFlavorsMap {
		f.Deprecated = true
		newFlavorsMap[f.Name] = f
	}
	info.Flavors = nil
	for _, flavor := range newFlavorsMap {
		info.Flavors = append(info.Flavors, flavor)
	}
	return addedFlavorsMap, deletedFlavorsMap, updatedFlavorsMap, recreatedFlavorsMap
}

func (s *CloudletInfoApi) InfraFlavorAlertTask(ctx context.Context, k interface{}) {
	key, ok := k.(HandleFlavorAlertWorkerKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorAlert task: key not correct type", "key", key)
		return
	}
	var err error
	clearAlertFlavors := []*string{}
	raiseAlertFlavors := []*string{}
	cloudletInfo := edgeproto.CloudletInfo{}
	refMap := map[string]*InfraFlavorUsage{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &key.cloudletKey, &cloudletInfo) {
			return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		refMap, err = s.getInfraFlavorUsageCounts(ctx, stm, &cloudletInfo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup infra flavor usage failed", "error", err)
			return err
		}
		curFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
		for _, f := range cloudletInfo.Flavors {
			curFlavorsMap[f.Name] = f
		}
		for _, flavor := range cloudletInfo.Flavors {
			if _, found := refMap[flavor.Name]; !found && flavor.Deprecated {
				delete(curFlavorsMap, flavor.Name)
				clearAlertFlavors = append(clearAlertFlavors, &flavor.Name)
			}
			if _, found := refMap[flavor.Name]; found {
				count := refMap[flavor.Name].refCount
				if count != 0 && flavor.Deprecated {
					raiseAlertFlavors = append(raiseAlertFlavors, &flavor.Name)
				}
			}
		}
		cloudletInfo.Flavors = nil
		for _, f := range curFlavorsMap {
			cloudletInfo.Flavors = append(cloudletInfo.Flavors, f)
		}
		cloudletInfoApi.store.STMPut(stm, &cloudletInfo)
		return nil
	})
	for _, f := range clearAlertFlavors {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorAlert task clear deprecated alert, unused", "infra flavor", *f)
		ClearDeletedInfraFlavorAlert(ctx, &key.cloudletKey, *f)
	}
	for _, f := range raiseAlertFlavors {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorAlert task raise deprecated alert", "infra flavor", *f)
		RaiseDeletedInfraFlavorAlert(ctx, &key.cloudletKey, *f, refMap[*f].users)
	}
}

type InfraFlavorUsage struct {
	refCount int
	users    string
}

func (s *CloudletInfoApi) getInfraFlavorUsageCounts(ctx context.Context, stm concurrency.STM, info *edgeproto.CloudletInfo) (map[string]*InfraFlavorUsage, error) {
	// walk clouldetRefs counting number of entities using the cloudlets infra flavors
	// returns a map[flavorName]InfraFlavorUsage with refCount and the discovered obj.Names that contribute to the refCount
	log.SpanLog(ctx, log.DebugLevelInfra, "find usage counts for all flavors of", "cloudlet", info.Key)
	flavorRefs := make(map[string]*InfraFlavorUsage)
	cloudletRefs := edgeproto.CloudletRefs{}

	if !cloudletRefsApi.store.STMGet(stm, &info.Key, &cloudletRefs) {
		return flavorRefs, fmt.Errorf("CloudletRefs not found")
	}
	for _, clust := range cloudletRefs.ClusterInsts {
		clusterInst := edgeproto.ClusterInst{}
		clusterInstKey := edgeproto.ClusterInstKey{
			ClusterKey:   clust.ClusterKey,
			CloudletKey:  info.Key,
			Organization: clust.Organization,
		}
		found := clusterInstApi.cache.Get(&clusterInstKey, &clusterInst)
		if found {
			// If we ever support custom clusters, where each worker node may have its own flavor,
			// this will need revisting
			if _, found := flavorRefs[clusterInst.MasterNodeFlavor]; !found {
				usage := InfraFlavorUsage{}
				flavorRefs[clusterInst.MasterNodeFlavor] = &usage
			}
			usage := flavorRefs[clusterInst.MasterNodeFlavor]
			usage.users = usage.users + clusterInstKey.ClusterKey.Name + " "
			if clusterInst.NumMasters > 0 {
				usage.refCount += int(clusterInst.NumMasters)
			}
			if clusterInst.NumNodes > 0 {
				if _, found := flavorRefs[clusterInst.NodeFlavor]; !found {
					usage := InfraFlavorUsage{}
					flavorRefs[clusterInst.NodeFlavor] = &usage
				}
				usage := flavorRefs[clusterInst.NodeFlavor]
				usage.users = usage.users + clusterInstKey.ClusterKey.Name + " "
				usage.refCount += int(clusterInst.NumNodes)
			}
		}
		for _, appInstRefKey := range cloudletRefs.VmAppInsts {
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
			found := appInstApi.store.STMGet(stm, &appInstKey, &appInst)
			if found {
				if _, found := flavorRefs[appInst.VmFlavor]; !found {
					usage := InfraFlavorUsage{}
					flavorRefs[appInst.VmFlavor] = &usage
				}
				usage := flavorRefs[appInst.VmFlavor]
				usage.users += usage.users + appInst.Key.AppKey.Name + " "
				usage.refCount += 1
			}
		}
	}
	return flavorRefs, nil
}

func genInfraFlavorEvent(ctx context.Context, key *edgeproto.CloudletKey, flavors map[string]*edgeproto.FlavorInfo, reason string) {
	vals := ""
	for _, flavor := range flavors {
		vals = vals + " " + flavor.Name
	}
	nodeMgr.Event(ctx, reason, key.Organization, key.GetTags(), nil, "flavors", vals)
}

func (s *CloudletInfoApi) fixupFlavorUpdate(ctx context.Context, in *edgeproto.CloudletInfo, added, deleted, updated, recreated map[string]*edgeproto.FlavorInfo) {
	// generate events for all added, deleted, or updated
	if len(added) != 0 {
		genInfraFlavorEvent(ctx, &in.Key, added, "flavors added")
	}
	if len(deleted) != 0 {
		genInfraFlavorEvent(ctx, &in.Key, deleted, "flavors deleted")
	}
	if len(updated) != 0 {
		genInfraFlavorEvent(ctx, &in.Key, updated, "flavors updated")
	}
	// Alerts
	for _, flavor := range recreated {
		// clear any deprecated alert since flavor is no longer deprecated (doesn't matter if in use or not)
		ClearDeletedInfraFlavorAlert(ctx, &in.Key, flavor.Name)
	}
	if len(deleted) != 0 {
		workerKey := HandleFlavorAlertWorkerKey{
			cloudletKey: in.Key,
		}
		s.infraFlavorAlertTask.NeedsWork(ctx, workerKey)
	}
}
