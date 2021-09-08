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
)

type CloudletInfoApi struct {
	sync  *Sync
	store edgeproto.CloudletInfoStore
	cache edgeproto.CloudletInfoCache
}

var cloudletInfoApi = CloudletInfoApi{}
var cleanupCloudletInfoTimeout = 5 * time.Minute

func InitCloudletInfoApi(sync *Sync) {
	cloudletInfoApi.sync = sync
	cloudletInfoApi.store = edgeproto.NewCloudletInfoStore(sync.store)
	edgeproto.InitCloudletInfoCache(&cloudletInfoApi.cache)
	sync.RegisterCache(&cloudletInfoApi.cache)
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

func (cd *CloudletInfoApi) findFlavorDeltas(ctx context.Context, flavorMap, newFlavorMap map[string]edgeproto.FlavorInfo) ([]edgeproto.FlavorInfo, []edgeproto.FlavorInfo, []edgeproto.FlavorInfo) {

	addedFlavors := []edgeproto.FlavorInfo{}
	deletedFlavors := []edgeproto.FlavorInfo{}
	updatedFlavors := []edgeproto.FlavorInfo{}

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
						deletedFlavors = append(deletedFlavors, flavor)
					}
				}
				// deletedFlavors from flavorMap
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
						addedFlavors = append(addedFlavors, flavor)
					}
				}
				// addedflavors to flavorMap
				for _, flavor := range addedFlavors {
					flavorMap[flavor.Name] = flavor
				}
			}
		}
	}
	if len(flavorMap) != len(newFlavorMap) {
		log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh delta internal error")
		return addedFlavors, deletedFlavors, updatedFlavors
	}
	// assert maps len are equal
	// Now check struct updates, deep equal on maps doesn't check beyond len and keys
	for key, val := range newFlavorMap {
		// Updated flavor(s)
		if !reflect.DeepEqual(newFlavorMap[key], flavorMap[key]) {
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: updating")
			// Flavor definition changed update
			updatedFlavors = append(updatedFlavors, val)
		}
	}
	// recap
	log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh:", "old flavs:", oldFlavorCount, "cur flavs", newFlavorCount, "added", len(addedFlavors),
		"deleted", len(deletedFlavors), "updated", len(updatedFlavors))
	return addedFlavors, deletedFlavors, updatedFlavors
}

// Return a potentially modified flavor list to store in the db for this cloudlet flavor list during Update.
// It may include flavors from the deletedFlavors list that are currently in use, and those will be marked deprecated.
func (s *CloudletInfoApi) HandleInfraFlavorDeltas(ctx context.Context, in *edgeproto.CloudletInfo, curFlavors []*edgeproto.FlavorInfo) []*edgeproto.FlavorInfo {

	var genAlert = false
	flavorMap := make(map[string]edgeproto.FlavorInfo)
	newFlavorMap := make(map[string]edgeproto.FlavorInfo)
	// recap logging only
	oldFlavorCount := len(curFlavors)
	newFlavorCount := len(in.Flavors)

	// whats in the db currently
	for _, flavor := range curFlavors {
		flavorMap[flavor.Name] = *flavor
	}
	// whats in the CloudletInfo update (from crm: GatherCloudletInfo)
	for _, flavor := range in.Flavors {
		newFlavorMap[flavor.Name] = *flavor
	}
	// determine any changes in flavor list
	addedFlavors, deletedFlavors, updatedFlavors := s.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	if len(addedFlavors)+len(deletedFlavors)+len(updatedFlavors) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "check OS flavor: no changes")
		return in.Flavors
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "check OS flavor: changes", "old flavs:", oldFlavorCount, "cur flavs", newFlavorCount, "added", len(addedFlavors),
		"deleted", len(deletedFlavors), "updated", len(updatedFlavors))

	// Events and Alerts
	if len(addedFlavors) != 0 {
		var vals string = ""
		for _, flavor := range addedFlavors {
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "FLAVOR_ADDED", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)

	}
	if len(updatedFlavors) != 0 {
		var vals string = ""
		// InfraFlavorEvent "UPDATED"
		for _, flavor := range updatedFlavors {
			// attach
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "FLAVOR_UPDATED", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
	}
	if len(deletedFlavors) != 0 {
		var vals string = ""
		// InfraFlavorEvent "DELETED" + Alert
		for _, flavor := range deletedFlavors {
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "FLAVOR_DELETED", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
		deleteFlavors := s.IsCloudletUsingFlavors(ctx, in, deletedFlavors)

		if len(deleteFlavors) != 0 {
			// Add any now deprecated flavors to the updated set of flavors to store with CloudletInfo
			for _, flavor := range deleteFlavors {
				if flavor.Deprecated == true {
					// add this flavor to the update, we're basically adding back in the deleted flavor
					// that is still in use.
					in.Flavors = append(in.Flavors, flavor)
					// If we have at least one deprecated flavor, generate alert
					genAlert = true

				} // if its not deprecatd, its already gone from in.Flavors in the calling Update
			}
		}
		// Only generate an alert if at least one deleteFlavor is marked deprecated, else
		// the delete is ok and will be unnoticed.
		if genAlert {
			alert := edgeproto.Alert{}
			alert.State = "firing"
			alert.ActiveAt = dme.Timestamp{}
			ts := time.Now()
			alert.ActiveAt.Seconds = ts.Unix()
			alert.ActiveAt.Nanos = int32(ts.Nanosecond())
			alert.Labels = make(map[string]string)
			alert.Labels["alertname"] = cloudcommon.AlertInfraFlavorDeleted
			alert.Labels["cloudlet"] = in.Key.Name
			alert.Labels["cloudletorg"] = in.Key.Organization
			alert.Annotations = make(map[string]string)
			for _, flavor := range deletedFlavors {
				alert.Annotations[flavor.Name] = flavor.Name // alt?
			}
			alert.Annotations[cloudcommon.AlertAnnotationTitle] = cloudcommon.AlertInfraFlavorDeleted
			alert.Annotations[cloudcommon.AlertAnnotationDescription] = cloudcommon.AlertInfraFlavorDeletedDescription
			// This call panics from unit tests hm
			alertApi.cache.Update(ctx, &alert, 0)
			alertApi.Update(ctx, &alert, 0)
		}
	}
	return in.Flavors
}

func (s *CloudletInfoApi) IsCloudletUsingFlavors(ctx context.Context, info *edgeproto.CloudletInfo, deletedFlavors []edgeproto.FlavorInfo) []*edgeproto.FlavorInfo {

	log.SpanLog(ctx, log.DebugLevelApi, "IsCloudletUsingFlavors deleted", "flavors", deletedFlavors)

	var deleteFlavors []*edgeproto.FlavorInfo
	// deletedFlavors represent flavors that currently exist in the db, but are not found in the update in.Flavor list
	// If they are in use, we should not allow them to be deleted from the cloudletinfo.Flavor list /db but rather mark them as deprecated.
	var allApps = make(map[edgeproto.AppKey]edgeproto.App)
	var allAppInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	var allClusterInsts = make(map[edgeproto.ClusterInstKey]edgeproto.ClusterInst)
	var allVMPools = make(map[edgeproto.VMPoolKey]edgeproto.VMPool)
	var notFoundError error = fmt.Errorf("NotFound")

	appApi.cache.GetAllKeys(ctx, func(k *edgeproto.AppKey, modRev int64) {
		allApps[*k] = edgeproto.App{}
	})
	appInstApi.cache.GetAllKeys(ctx, func(k *edgeproto.AppInstKey, modRev int64) {
		allAppInsts[*k] = edgeproto.AppInst{}
	})
	clusterInstApi.cache.GetAllKeys(ctx, func(k *edgeproto.ClusterInstKey, modRev int64) {
		allClusterInsts[*k] = edgeproto.ClusterInst{}
	})
	vmPoolApi.cache.GetAllKeys(ctx, func(k *edgeproto.VMPoolKey, modRev int64) {
		allVMPools[*k] = edgeproto.VMPool{}
	})
	// first object found using this to be deleted infra-flavor sets deprecated, stops search, get next
	for _, flavor := range deletedFlavors {
		deleteFlavors = append(deleteFlavors, &flavor)
		// check Apps DefaultFlavor
		for key, _ := range allApps {
			app := edgeproto.App{}
			err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
				if !appApi.store.STMGet(stm, &key, &app) {
					log.SpanLog(ctx, log.DebugLevelApi, "app not found", "app", key)
					return notFoundError
				}
				return nil
			})
			if err != nil {
				continue
			}
			infraFlavorName, err := s.InfraFlavorNameOf(ctx, info.Key, app.DefaultFlavor.Name)
			if err != nil {
				continue
			}
			if infraFlavorName == flavor.Name {
				flavor.Deprecated = true
				break
			}
		}
		if flavor.Deprecated {
			continue
		}
		// check for appInst.Flavor
		for key, _ := range allAppInsts {
			appInst := edgeproto.AppInst{}
			err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
				if !appInstApi.store.STMGet(stm, &key, &appInst) {
					log.SpanLog(ctx, log.DebugLevelApi, "appInst not found", "app", key)
					return notFoundError
				}
				return nil
			})
			if err != nil {
				continue
			}
			infraFlavorName, err := s.InfraFlavorNameOf(ctx, info.Key, appInst.Flavor.Name)
			if err != nil {
				continue
			}
			if infraFlavorName == flavor.Name {
				flavor.Deprecated = true
				break
			}
		}
		if flavor.Deprecated {
			continue
		}
		// Check clustInst.Flavor
		for key, _ := range allClusterInsts {
			clusterInst := edgeproto.ClusterInst{}

			err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
				if !clusterInstApi.store.STMGet(stm, &key, &clusterInst) {
					log.SpanLog(ctx, log.DebugLevelApi, "clusterInst not found", "app", key)
					return notFoundError
				}
				return nil
			})
			if err != nil {
				continue
			}
			infraFlavorName, err := s.InfraFlavorNameOf(ctx, info.Key, clusterInst.Flavor.Name)
			if err != nil {
				continue
			}
			if infraFlavorName == flavor.Name {
				flavor.Deprecated = true
				break
			}
		}
		if flavor.Deprecated {
			continue
		}
		for key, _ := range allVMPools {
			vmPool := edgeproto.VMPool{}
			err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
				if !vmPoolApi.store.STMGet(stm, &key, &vmPool) {
					log.SpanLog(ctx, log.DebugLevelApi, "vmPool not found", "app", key)
					return notFoundError
				}
				return nil
			})
			if err != nil {
				// get next pool
				continue
			}
			for _, vm := range vmPool.Vms {
				infraFlavorName, err := s.InfraFlavorNameOf(ctx, info.Key, vm.Flavor.Name)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelApi, "vm.Flavor.Name not matched", "error", err)
					continue
				}
				if infraFlavorName == flavor.Name {
					flavor.Deprecated = true
					break
				}
			}
		}
	}
	return deleteFlavors
}

func (s *CloudletInfoApi) InfraFlavorNameOf(ctx context.Context, cloudletKey edgeproto.CloudletKey, metaFlavorName string) (string, error) {
	matchspec := edgeproto.FlavorMatch{
		Key:        cloudletKey,
		FlavorName: metaFlavorName,
	}
	vmSpec, err := cloudletApi.FindFlavorMatch(ctx, &matchspec)
	if err != nil {
		return "", err
	}
	return vmSpec.FlavorName, nil
}
