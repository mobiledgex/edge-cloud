package main

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

// Set up dummy responses for info data expected from CRM.
// Used for unit testing without a real CRM present.
// Functionality here emulates crmutil/controller-data.go.

// Add in a little day to allow intermediate state changes to be seen.
// But don't add too much because it increases the unit test run time.
const DummyInfoDelay = 10 * time.Millisecond

func NewDummyInfoResponder(appInstCache *edgeproto.AppInstCache, clusterInstCache *edgeproto.ClusterInstCache, recvAppInstInfo notify.RecvAppInstInfoHandler, recvClusterInstInfo notify.RecvClusterInstInfoHandler) *DummyInfoResponder {
	d := DummyInfoResponder{
		AppInstCache:        appInstCache,
		ClusterInstCache:    clusterInstCache,
		RecvAppInstInfo:     recvAppInstInfo,
		RecvClusterInstInfo: recvClusterInstInfo,
	}
	d.AppInstCache.SetNotifyCb(d.runAppInstChanged)
	d.ClusterInstCache.SetNotifyCb(d.runClusterInstChanged)
	edgeproto.InitClusterInstInfoCache(&d.ClusterInstInfoCache)
	edgeproto.InitAppInstInfoCache(&d.AppInstInfoCache)
	d.AppInstInfoCache.SetNotifyCb(d.appInstInfoCb)
	d.ClusterInstInfoCache.SetNotifyCb(d.clusterInstInfoCb)
	return &d
}

type DummyInfoResponder struct {
	AppInstCache                 *edgeproto.AppInstCache
	AppInstInfoCache             edgeproto.AppInstInfoCache
	ClusterInstCache             *edgeproto.ClusterInstCache
	ClusterInstInfoCache         edgeproto.ClusterInstInfoCache
	RecvAppInstInfo              notify.RecvAppInstInfoHandler
	RecvClusterInstInfo          notify.RecvClusterInstInfoHandler
	simulateAppCreateFailure     bool
	simulateAppUpdateFailure     bool
	simulateAppDeleteFailure     bool
	simulateClusterCreateFailure bool
	simulateClusterUpdateFailure bool
	simulateClusterDeleteFailure bool
}

func (d *DummyInfoResponder) SetSimulateAppCreateFailure(state bool) {
	d.simulateAppCreateFailure = state
}

func (d *DummyInfoResponder) SetSimulateAppDeleteFailure(state bool) {
	d.simulateAppDeleteFailure = state
}

func (d *DummyInfoResponder) SetSimulateClusterCreateFailure(state bool) {
	d.simulateClusterCreateFailure = state
}

func (d *DummyInfoResponder) SetSimulateClusterDeleteFailure(state bool) {
	d.simulateClusterDeleteFailure = state
}

func (d *DummyInfoResponder) runClusterInstChanged(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst) {
	go d.clusterInstChanged(key)
}

func (d *DummyInfoResponder) runAppInstChanged(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	go d.appInstChanged(key)
}

func (d *DummyInfoResponder) clusterInstChanged(key *edgeproto.ClusterInstKey) {
	inst := edgeproto.ClusterInst{}
	found := d.ClusterInstCache.Get(key, &inst)
	if !found {
		return
	}
	if inst.State == edgeproto.TrackedState_UpdateRequested {
		// update
		log.DebugLog(log.DebugLevelApi, "Update ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(key, edgeproto.TrackedState_Updating)
		log.DebugLog(log.DebugLevelApi, "ClusterInst ready", "key", key)
		if d.simulateClusterUpdateFailure {
			d.ClusterInstInfoCache.SetError(key, edgeproto.TrackedState_UpdateError, "crm update ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(key, edgeproto.TrackedState_Ready)
		}
	} else if inst.State == edgeproto.TrackedState_CreateRequested {
		// create
		log.DebugLog(log.DebugLevelApi, "Create ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(key, edgeproto.TrackedState_Creating)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "ClusterInst ready", "key", key)
		if d.simulateClusterCreateFailure {
			d.ClusterInstInfoCache.SetError(key, edgeproto.TrackedState_CreateError, "crm create ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(key, edgeproto.TrackedState_Ready)
		}
	} else if inst.State == edgeproto.TrackedState_DeleteRequested {
		// delete
		log.DebugLog(log.DebugLevelApi, "Delete ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(key, edgeproto.TrackedState_Deleting)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "ClusterInst deleted", "key", key)
		if d.simulateClusterDeleteFailure {
			d.ClusterInstInfoCache.SetError(key, edgeproto.TrackedState_DeleteError, "crm delete ClusterInst failed")
		} else {
			info := edgeproto.ClusterInstInfo{Key: *key}
			d.ClusterInstInfoCache.Delete(&info, 0)
		}
	}
}

func (d *DummyInfoResponder) appInstChanged(key *edgeproto.AppInstKey) {
	inst := edgeproto.AppInst{}
	found := d.AppInstCache.Get(key, &inst)
	if !found {
		return
	}
	if inst.State == edgeproto.TrackedState_UpdateRequested {
		// update
		log.DebugLog(log.DebugLevelApi, "Update app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(key, edgeproto.TrackedState_Updating)
		log.DebugLog(log.DebugLevelApi, "app inst ready", "key", key)
		if d.simulateAppUpdateFailure {
			d.AppInstInfoCache.SetError(key, edgeproto.TrackedState_UpdateError, "crm update app inst failed")
		} else {
			d.AppInstInfoCache.SetState(key, edgeproto.TrackedState_Ready)
		}
	} else if inst.State == edgeproto.TrackedState_CreateRequested {
		// create
		log.DebugLog(log.DebugLevelApi, "Create app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(key, edgeproto.TrackedState_Creating)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "app inst ready", "key", key)
		if d.simulateAppCreateFailure {
			d.AppInstInfoCache.SetError(key, edgeproto.TrackedState_CreateError, "crm create app inst failed")
		} else {
			d.AppInstInfoCache.SetState(key, edgeproto.TrackedState_Ready)
		}
	} else if inst.State == edgeproto.TrackedState_DeleteRequested {
		// delete
		log.DebugLog(log.DebugLevelApi, "Delete app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(key, edgeproto.TrackedState_Deleting)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "app inst deleted", "key", key)
		if d.simulateAppDeleteFailure {
			d.AppInstInfoCache.SetError(key, edgeproto.TrackedState_DeleteError, "crm delete app inst failed")
		} else {
			info := edgeproto.AppInstInfo{Key: *key}
			d.AppInstInfoCache.Delete(&info, 0)
		}
	}
}

func (d *DummyInfoResponder) clusterInstInfoCb(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInstInfo) {
	inst := edgeproto.ClusterInstInfo{}
	if d.ClusterInstInfoCache.Get(key, &inst) {
		d.RecvClusterInstInfo.Update(&inst, 0)
	} else {
		inst.Key = *key
		d.RecvClusterInstInfo.Delete(&inst, 0)
	}
}

func (d *DummyInfoResponder) appInstInfoCb(key *edgeproto.AppInstKey, old *edgeproto.AppInstInfo) {
	inst := edgeproto.AppInstInfo{}
	if d.AppInstInfoCache.Get(key, &inst) {
		d.RecvAppInstInfo.Update(&inst, 0)
	} else {
		inst.Key = *key
		d.RecvAppInstInfo.Delete(&inst, 0)
	}
}
