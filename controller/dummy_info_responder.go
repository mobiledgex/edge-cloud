package main

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Set up dummy responses for info data expected from CRM.
// Used for unit testing without a real CRM present.
// Functionality here emulates crmutil/controller-data.go.

// Add in a little day to allow intermediate state changes to be seen.
// But don't add too much because it increases the unit test run time.
const DummyInfoDelay = 10 * time.Millisecond

func NewDummyInfoResponder(appInstCache *edgeproto.AppInstCache, clusterInstCache *edgeproto.ClusterInstCache, appInstInfoCache *edgeproto.AppInstInfoCache, clusterInstInfoCache *edgeproto.ClusterInstInfoCache) *DummyInfoResponder {
	d := DummyInfoResponder{
		AppInstCache:         appInstCache,
		AppInstInfoCache:     appInstInfoCache,
		ClusterInstCache:     clusterInstCache,
		ClusterInstInfoCache: clusterInstInfoCache,
	}
	d.AppInstCache.SetNotifyCb(d.runAppInstChanged)
	d.ClusterInstCache.SetNotifyCb(d.runClusterInstChanged)
	return &d
}

type DummyInfoResponder struct {
	AppInstCache         *edgeproto.AppInstCache
	AppInstInfoCache     *edgeproto.AppInstInfoCache
	ClusterInstCache     *edgeproto.ClusterInstCache
	ClusterInstInfoCache *edgeproto.ClusterInstInfoCache
	simulateFailure      bool
}

func (d *DummyInfoResponder) SetSimulateFailure(state bool) {
	d.simulateFailure = state
}

func (d *DummyInfoResponder) runClusterInstChanged(key *edgeproto.ClusterInstKey) {
	go d.clusterInstChanged(key)
}

func (d *DummyInfoResponder) runAppInstChanged(key *edgeproto.AppInstKey) {
	go d.appInstChanged(key)
}

func (d *DummyInfoResponder) clusterInstChanged(key *edgeproto.ClusterInstKey) {
	inst := edgeproto.ClusterInst{}
	found := d.ClusterInstCache.Get(key, &inst)
	if found {
		info := edgeproto.ClusterInstInfo{}
		if d.ClusterInstInfoCache.Get(key, &info) {
			// update
			log.DebugLog(log.DebugLevelApi, "Update cluster inst", "key", key)
			time.Sleep(DummyInfoDelay)
			d.ClusterInstInfoCache.SetState(key, edgeproto.ClusterState_ClusterStateChanging)
			log.DebugLog(log.DebugLevelApi, "cluster inst ready", "key", key)
			if d.simulateFailure {
				d.ClusterInstInfoCache.SetError(key, "crm update cluster inst failed")
			} else {
				d.ClusterInstInfoCache.SetState(key, edgeproto.ClusterState_ClusterStateReady)
			}
		} else {
			// create
			log.DebugLog(log.DebugLevelApi, "Create cluster inst", "key", key)
			time.Sleep(DummyInfoDelay)
			d.ClusterInstInfoCache.SetState(key, edgeproto.ClusterState_ClusterStateBuilding)
			time.Sleep(DummyInfoDelay)
			log.DebugLog(log.DebugLevelApi, "cluster inst ready", "key", key)
			if d.simulateFailure {
				d.ClusterInstInfoCache.SetError(key, "crm create cluster inst failed")
			} else {
				d.ClusterInstInfoCache.SetState(key, edgeproto.ClusterState_ClusterStateReady)
			}
		}
	} else {
		// delete
		log.DebugLog(log.DebugLevelApi, "Delete cluster inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(key, edgeproto.ClusterState_ClusterStateDeleting)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "cluster inst deleted", "key", key)
		if d.simulateFailure {
			d.ClusterInstInfoCache.SetError(key, "crm delete cluster inst failed")
		} else {
			info := edgeproto.ClusterInstInfo{Key: *key}
			d.ClusterInstInfoCache.Delete(&info, 0)
		}
	}
}

func (d *DummyInfoResponder) appInstChanged(key *edgeproto.AppInstKey) {
	inst := edgeproto.AppInst{}
	found := d.AppInstCache.Get(key, &inst)
	if found {
		info := edgeproto.AppInstInfo{}
		if d.AppInstInfoCache.Get(key, &info) {
			// update
			log.DebugLog(log.DebugLevelApi, "Update app inst", "key", key)
			time.Sleep(DummyInfoDelay)
			d.AppInstInfoCache.SetState(key, edgeproto.AppState_AppStateChanging)
			log.DebugLog(log.DebugLevelApi, "app inst ready", "key", key)
			if d.simulateFailure {
				d.AppInstInfoCache.SetError(key, "crm update app inst failed")
			} else {
				d.AppInstInfoCache.SetState(key, edgeproto.AppState_AppStateReady)
			}
		} else {
			// create
			log.DebugLog(log.DebugLevelApi, "Create app inst", "key", key)
			time.Sleep(DummyInfoDelay)
			d.AppInstInfoCache.SetState(key, edgeproto.AppState_AppStateBuilding)
			time.Sleep(DummyInfoDelay)
			log.DebugLog(log.DebugLevelApi, "app inst ready", "key", key)
			if d.simulateFailure {
				d.AppInstInfoCache.SetError(key, "crm create app inst failed")
			} else {
				d.AppInstInfoCache.SetState(key, edgeproto.AppState_AppStateReady)
			}
		}
	} else {
		// delete
		log.DebugLog(log.DebugLevelApi, "Delete app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(key, edgeproto.AppState_AppStateDeleting)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "app inst deleted", "key", key)
		if d.simulateFailure {
			d.AppInstInfoCache.SetError(key, "crm delete app inst failed")
		} else {
			info := edgeproto.AppInstInfo{Key: *key}
			d.AppInstInfoCache.Delete(&info, 0)
		}
	}
}
