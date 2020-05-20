package main

import (
	"context"
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

func (d *DummyInfoResponder) runClusterInstChanged(ctx context.Context, key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst, modRev int64) {
	go d.clusterInstChanged(ctx, key, modRev)
}

func (d *DummyInfoResponder) runAppInstChanged(ctx context.Context, key *edgeproto.AppInstKey, old *edgeproto.AppInst, modRev int64) {
	go d.appInstChanged(ctx, key, modRev)
}

func (d *DummyInfoResponder) clusterInstChanged(ctx context.Context, key *edgeproto.ClusterInstKey, modRev int64) {
	inst := edgeproto.ClusterInst{}
	found := d.ClusterInstCache.Get(key, &inst)
	if !found {
		return
	}
	if inst.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		// update
		log.DebugLog(log.DebugLevelApi, "Update ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_UPDATING)
		log.DebugLog(log.DebugLevelApi, "ClusterInst ready", "key", key)
		if d.simulateClusterUpdateFailure {
			d.ClusterInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_UPDATE_ERROR, "crm update ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// create
		log.DebugLog(log.DebugLevelApi, "Create ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_CREATING)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "ClusterInst ready", "key", key)
		if d.simulateClusterCreateFailure {
			d.ClusterInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_CREATE_ERROR, "crm create ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_DELETE_REQUESTED {
		// delete
		log.DebugLog(log.DebugLevelApi, "Delete ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_DELETING)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "ClusterInst deleted", "key", key)
		if d.simulateClusterDeleteFailure {
			d.ClusterInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_DELETE_ERROR, "crm delete ClusterInst failed")
		} else {
			info := edgeproto.ClusterInstInfo{Key: *key}
			d.ClusterInstInfoCache.Delete(ctx, &info, 0)
		}
	}
}

func (d *DummyInfoResponder) appInstChanged(ctx context.Context, key *edgeproto.AppInstKey, modRev int64) {
	inst := edgeproto.AppInst{}
	found := d.AppInstCache.Get(key, &inst)
	if !found {
		return
	}
	if inst.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		// update
		log.DebugLog(log.DebugLevelApi, "Update app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_UPDATING)
		log.DebugLog(log.DebugLevelApi, "app inst ready", "key", key)
		if d.simulateAppUpdateFailure {
			d.AppInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_UPDATE_ERROR, "crm update app inst failed")
		} else {
			d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// create
		log.DebugLog(log.DebugLevelApi, "Create app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_CREATING)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "app inst ready", "key", key)
		if d.simulateAppCreateFailure {
			d.AppInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_CREATE_ERROR, "crm create app inst failed")
		} else {
			d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_DELETE_REQUESTED {
		// delete
		log.DebugLog(log.DebugLevelApi, "Delete app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_DELETING)
		time.Sleep(DummyInfoDelay)
		log.DebugLog(log.DebugLevelApi, "app inst deleted", "key", key)
		if d.simulateAppDeleteFailure {
			d.AppInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_DELETE_ERROR, "crm delete app inst failed")
		} else {
			info := edgeproto.AppInstInfo{Key: *key}
			d.AppInstInfoCache.Delete(ctx, &info, 0)
		}
	}
}

func (d *DummyInfoResponder) clusterInstInfoCb(ctx context.Context, key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInstInfo, modRev int64) {
	inst := edgeproto.ClusterInstInfo{}
	if d.ClusterInstInfoCache.Get(key, &inst) {
		d.RecvClusterInstInfo.Update(ctx, &inst, modRev)
	} else {
		inst.Key = *key
		d.RecvClusterInstInfo.Delete(ctx, &inst, modRev)
	}
}

func (d *DummyInfoResponder) appInstInfoCb(ctx context.Context, key *edgeproto.AppInstKey, old *edgeproto.AppInstInfo, modRev int64) {
	inst := edgeproto.AppInstInfo{}
	if d.AppInstInfoCache.Get(key, &inst) {
		d.RecvAppInstInfo.Update(ctx, &inst, modRev)
	} else {
		inst.Key = *key
		d.RecvAppInstInfo.Delete(ctx, &inst, modRev)
	}
}
