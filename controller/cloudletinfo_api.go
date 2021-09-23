package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	//	"strings"
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
	updateFlavors := in.Flavors
	var preUpdateFlavors []*edgeproto.FlavorInfo
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		info := edgeproto.CloudletInfo{}
		if s.store.STMGet(stm, &in.Key, &info) {
			if in.State == dme.CloudletState_CLOUDLET_STATE_READY &&
				info.State != dme.CloudletState_CLOUDLET_STATE_READY {
				changedToOnline = true
			}
			preUpdateFlavors = info.Flavors
		}

		s.store.STMPut(stm, in)
		return nil
	})
	// crm running GetCldudletInfo in a periodic thread may trigger an update, fix it up if needed.
	s.HandleInfraFlavorUpdate(ctx, in, updateFlavors, preUpdateFlavors)
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

func ClearDeletedInfraFlavorAlert(ctx context.Context, info *edgeproto.CloudletInfo, flavor string /* *edgeproto.FlavorInfo*/) {

	fmt.Printf("\n\tClearDeletedInfraFlavors-I-delete alert for %s\n\n", flavor)
	log.SpanLog(ctx, log.DebugLevelInfra, "clear alert for recreated", "flavor", flavor)
	alert := edgeproto.Alert{}
	alert.Labels = make(map[string]string)
	alert.Labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeCloudlet
	alert.Labels["alertname"] = cloudcommon.AlertInfraFlavorDeleted
	alert.Labels["cloudlet"] = info.Key.Name
	alert.Labels["cloudletorg"] = info.Key.Organization
	alert.Labels["infraflavor"] = flavor
	alertApi.Delete(ctx, &alert, 0)
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
}

// Given all the flavors in the in.Flavors from the cloudletInfo Update, return a list of missing flavors.
// That is, flavors in use that are not found on info.Flavor list.

func (s *CloudletInfoApi) getMissingFlavors(ctx context.Context, info *edgeproto.CloudletInfo, inFlavorMap map[string]*edgeproto.FlavorInfo) ([]string, error) {
	var missingFlavors []string
	missingFlavorsMap := make(map[string]struct{})
	cloudRefs := edgeproto.CloudletRefs{}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletRefsApi.cache.Get(&info.Key, &cloudRefs) {
			log.SpanLog(ctx, log.DebugLevelInfra, "missing flavors: no refs for", "cloudlet", info.Key)
			return fmt.Errorf("CloudletRefs not found")
		}
		return nil
	})
	if err != nil {
		return missingFlavors, err
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
			if _, found := inFlavorMap[clusterInst.NodeFlavor]; !found {
				missingFlavorsMap[clusterInst.NodeFlavor] = struct{}{}
			}
			if _, found := inFlavorMap[clusterInst.MasterNodeFlavor]; !found {
				missingFlavorsMap[clusterInst.MasterNodeFlavor] = struct{}{}
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
			found := appInstApi.cache.Get(&appInstKey, &appInst)
			if found {
				if _, found := inFlavorMap[appInst.VmFlavor]; !found {
					missingFlavorsMap[appInst.Flavor.Name] = struct{}{}
				}
			}
		}
	}
	for k, _ := range missingFlavorsMap {
		missingFlavors = append(missingFlavors, k)
	}
	return missingFlavors, nil
}

func (cd *CloudletInfoApi) findFlavorDeltas(ctx context.Context, flavorMap, newFlavorMap map[string]*edgeproto.FlavorInfo) ([]edgeproto.FlavorInfo, []edgeproto.FlavorInfo, []edgeproto.FlavorInfo, []edgeproto.FlavorInfo) {
	addedFlavors := []edgeproto.FlavorInfo{}
	deletedFlavors := []edgeproto.FlavorInfo{}
	updatedFlavors := []edgeproto.FlavorInfo{}
	clearAlertFlavors := []edgeproto.FlavorInfo{}
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
	// sanity, assert maps are equal len
	if len(flavorMap) != len(newFlavorMap) {
		log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh delta internal error")
		return nil, nil, nil, nil
	}
	// Now check struct updates, deep equal on maps doesn't check beyond len and keys
	for key, val := range newFlavorMap {
		// Updated flavor(s)
		if !reflect.DeepEqual(newFlavorMap[key], flavorMap[key]) {
			// special case: If flavorMap[key] is deprecated, we can clear the alert and remove the mark flavor is recreated
			flavor := flavorMap[key]
			if flavor.Deprecated {
				log.SpanLog(ctx, log.DebugLevelInfra, "deprecated infa flavor recreated", "flavor", flavor.Name)
				clearAlertFlavors = append(clearAlertFlavors, *val)
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh: updating")
			// Flavor definition changed
			updatedFlavors = append(updatedFlavors, *val)
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "flavor refresh:", "old flavs:", oldFlavorCount, "cur flavs", newFlavorCount, "added", len(addedFlavors),
		"deleted", len(deletedFlavors), "updated", len(updatedFlavors))
	return addedFlavors, deletedFlavors, updatedFlavors, clearAlertFlavors
}

func (s *CloudletInfoApi) HandleInfraFlavorUpdate(ctx context.Context, in *edgeproto.CloudletInfo, updateFlavors, preUpdateFlavors []*edgeproto.FlavorInfo) {
	flavorMap := make(map[string]*edgeproto.FlavorInfo)    // preUpdateFlavors
	newFlavorMap := make(map[string]*edgeproto.FlavorInfo) // updateFlvors
	// the STMPut has already happend, here we're fixing up after the fact
	for _, flavor := range preUpdateFlavors {
		flavorMap[flavor.Name] = flavor
	}
	for _, flavor := range in.Flavors {
		newFlavorMap[flavor.Name] = flavor
	}
	// determine changes in flavor lists
	addedFlavors, deletedFlavors, updatedFlavors, clearAlertsFlavors := s.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	if len(addedFlavors)+len(deletedFlavors)+len(updatedFlavors)+len(clearAlertsFlavors) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "check OS flavor: no changes")
		return
	}
	// Events and Alerts

	if len(addedFlavors) != 0 {
		var vals string = ""
		for _, flavor := range addedFlavors {
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "Flavor Added", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
	}
	if len(updatedFlavors) != 0 {
		var vals string = ""
		for _, flavor := range updatedFlavors {
			vals = vals + ", " + flavor.Name
		}
		nodeMgr.Event(ctx, "Flavor Updated", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)
	}
	depFlavorMap := make(map[string]*edgeproto.FlavorInfo)
	for _, f := range in.DeprecatedFlavors {
		depFlavorMap[f.Name] = f
	}
	// clear any pending deleted flavor alerts: flavor recreated case
	for _, flavor := range clearAlertsFlavors {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorUpdate clear pending alert for recreated", "flavor", flavor)
		fmt.Printf("\n\tHandle: ClearDeletedInfraFlavorAlert (recreated) for %+v\n\n", flavor)
		ClearDeletedInfraFlavorAlert(ctx, in, flavor.Name)
	}

	// fix up the deleted flavors, skip if already deprecated, we've handled this one already
	delFlavorMap := make(map[string]*edgeproto.FlavorInfo)
	for _, f := range deletedFlavors {
		fmt.Printf("\n\tHandle consider delflavor %+v\n", f)
		if !f.Deprecated {
			delFlavorMap[f.Name] = &f
		}
	}

	for _, flavor := range delFlavorMap {
		// if deleted flavor not found missing, not in use, ok to leave it deleted, ignore missing if already deprecated.
		var vals string = ""
		vals = vals + ", " + flavor.Name
		nodeMgr.Event(ctx, "Flavor Deleted", in.Key.Organization, in.Key.GetTags(), nil, "flavors", vals)

		missingFlavors, _ := s.getMissingFlavors(ctx, in, flavorMap)
		for _, flavor := range missingFlavors {
			if f, found := delFlavorMap[flavor]; found {
				// both missing + deleted Raise alert, mark deprecated and place on flavorMap to add back to in.Flavors
				log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorUpdate deleted and missing mark deprecated", "flavor", f)
				flavorMap[f.Name] = f
				depFlavor := delFlavorMap[f.Name]
				depFlavor.Deprecated = true
				depFlavorMap[f.Name] = f
				// alert per depFlavor since they don't ness. share a common resolution
				fmt.Printf("\n\tHandle: Deprecate flavor and reaise alert  %+v\n", depFlavor.Name)
				RaiseDeletedInfraFlavorAlert(ctx, in, depFlavor.Name)
				if !s.alertReaperCreated {
					s.clearInfraFlavorAlertTask.Init("infraFlavorsAlertReaper", s.InfraFlavorAlertCleanupTask)
					s.alertReaperCreated = true
				}
			}
		}
		if !s.alertReaperCreated && len(depFlavorMap) > 0 {
			s.clearInfraFlavorAlertTask.Init("infraFlavorsAlertReaper", s.InfraFlavorAlertCleanupTask)
			s.alertReaperCreated = true
		}
	}
	// Finish our fixup. We need to store 1) any cleared deprecation 2) any adding back a deprecated flavor
	in.DeprecatedFlavors = nil
	in.Flavors = nil
	// remove any cleared alerts from the deprecatd flavors list
	for _, f := range clearAlertsFlavors {
		if _, found := depFlavorMap[f.Name]; found {
			delete(depFlavorMap, f.Name)
		}
	}
	for _, f := range depFlavorMap {
		fmt.Printf("\n\tHandle appending flavor %+v for in.DeprecatedFlavors\n", f)
		in.DeprecatedFlavors = append(in.DeprecatedFlavors, f)
		flavorMap[f.Name] = f // ensure it is still part of in.Flavors
	}
	i := 0
	for _, flavor := range flavorMap {
		// check if any alerts were cleared, if so, clear its marking
		for _, f := range clearAlertsFlavors {
			if f.Name == flavor.Name {
				flavor.Deprecated = false
			}
		}
		in.Flavors = append(in.Flavors, flavor)
		i++
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudletInfoApi.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavorUpdate stm failed", "error", err)
	}
	return
}

func (s *CloudletInfoApi) InfraFlavorAlertCleanupTask(ctx context.Context, k interface{}) {

	key, ok := k.(HandleFlavorAlertWorkerKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup Unexpected failure, key not a HandleFlavorAlertWorkerKey", "key", key)
		return
	}
	cloudletInfo := edgeproto.CloudletInfo{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &key.cloudletKey, &cloudletInfo) {
			return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "worker task stm failed", "error", err)
		return
	}
	// no flavor? Then we're fixing up that last cloudletinfo.Update, since something was detected as deleted
	if key.flavor == "" {
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup task invalid args need infra flavor name")
		return
	}
	if len(cloudletInfo.DeprecatedFlavors) == 0 {
		// nothing to do
		log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup task no deprecated flavors for cloudlet, done")
		return
	}

	fmt.Printf("\n-----InfrarFlavorAlertCleaupTask BEGIN  (WORKER TASK) running Deleted Obj for flavor: %s ------------------------------\n\n", key.flavor)
	// xxx we should be able to create a key, and look it up. But it alwayes seems to exist xxx
	for _, val := range alertApi.cache.Objs {
		alert := val.Obj
		if name, found := alert.Labels["alertname"]; found && name == cloudcommon.AlertInfraFlavorDeleted {
			fmt.Printf("\n\tTASK: found alertname and name: %+v\n", name)
		} else {
			continue
		}
		if name, found := alert.Labels["infraflavor"]; found && name != key.flavor {
			continue
		} else {
			fmt.Printf("\n\tTASK: found infraflavor and our flavor name: %+v\n", name)
		}

		fmt.Printf("\n\tTASK found our target alert %+v\n", alert)

		fmt.Printf("\nTASK found alert proceeding to dlete alert and remove flavor: %+v\n", key.flavor)
		curFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
		depFlavorsMap := make(map[string]*edgeproto.FlavorInfo)
		for _, flav := range cloudletInfo.Flavors {
			curFlavorsMap[flav.Name] = flav
		}
		for _, flav := range cloudletInfo.DeprecatedFlavors {
			depFlavorsMap[flav.Name] = flav
		}
		if _, found := depFlavorsMap[key.flavor]; found {
			count, err := s.getInfraFlavorUsageCount(ctx, &cloudletInfo, key.flavor)
			if err != nil {
				fmt.Printf("\n\t: Task getInfraFlavorUsageCount returned err: %s  RETURNING\n", err.Error())
				// log error XXX
				return
			}
			fmt.Printf("\n\tTASK usage count for %s is now %d\n", key.flavor, count)
			if count <= 1 { // must be the last guy standing, clear alert and remove from cloudletInfor.Flavors
				log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup task clear pending Alerts", "infra flavor", key.flavor)
				fmt.Printf("\n\tTASK count %d ClearDeletedInfraFlavorAlert for %s\n", count, key.flavor)
				ClearDeletedInfraFlavorAlert(ctx, &cloudletInfo, key.flavor)

				for _, f := range cloudletInfo.Flavors {
					curFlavorsMap[f.Name] = f
				}
				for _, f := range cloudletInfo.DeprecatedFlavors {
					depFlavorsMap[f.Name] = f
				}

				// remove flavor
				delete(depFlavorsMap, key.flavor)
				delete(curFlavorsMap, key.flavor)
				cloudletInfo.DeprecatedFlavors = nil
				cloudletInfo.Flavors = nil
				for _, f := range depFlavorsMap {
					cloudletInfo.DeprecatedFlavors = append(cloudletInfo.DeprecatedFlavors, f)
				}
				for _, f := range curFlavorsMap {
					cloudletInfo.Flavors = append(cloudletInfo.Flavors, f)
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "InfraFlavor cleanup task delete unused", "infra flavor", key.flavor)
				err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
					cloudletInfoApi.store.STMPut(stm, &cloudletInfo)
					return nil
				})

				fmt.Printf("\n\tTASK flavors stored minus %s\n", key.flavor)
				for i, f := range cloudletInfo.Flavors {
					fmt.Printf("\n\t\tFlavors[%d] = %+v\n", i, f)
				}
				fmt.Printf("\n\tLen of depcreated flavors: %d\n", len(cloudletInfo.DeprecatedFlavors))
				for i, f := range cloudletInfo.DeprecatedFlavors {
					fmt.Printf("\n\t\tDeprecatedFlavors[%d] = %+v\n", i, f)
				}
				if err != nil {
					panic("STM failed\n")
				}
			} else {
				fmt.Printf("\n\tTASK: NOT YET count is %d\n\n", count)
			}
		} else {
			fmt.Printf("\n\tTASK flavor %s not found on cloudlet deprecated flavors list\n", key.flavor)
		}
	}
}

func (s *CloudletInfoApi) getInfraFlavorUsageCount(ctx context.Context, info *edgeproto.CloudletInfo, flavorName string) (int, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "find usage count for ", "infra flavor", flavorName, "cloudlet", info.Key)
	var count int = 0
	cloudRefs := edgeproto.CloudletRefs{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletRefsApi.cache.Get(&info.Key, &cloudRefs) {
			return fmt.Errorf("CloudletRefs not found")
		}
		return nil
	})
	if err != nil {
		return count, err
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
			if clusterInst.NodeFlavor == flavorName {
				count++
				if clusterInst.NumNodes > 0 {
					// revisit if we ever allow different flavors for differnt worker nodes for custom clusters
					count += int(clusterInst.NumNodes)
				}
			}
			if clusterInst.MasterNodeFlavor == flavorName {
				count++
				if clusterInst.NumMasters > 1 {
					count += int(clusterInst.NumMasters - 1)
				}
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
			found := appInstApi.cache.Get(&appInstKey, &appInst)
			if found {
				if appInst.VmFlavor == flavorName {
					count++
				}
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "flavor usage", "count", count)
	return count, nil
}
