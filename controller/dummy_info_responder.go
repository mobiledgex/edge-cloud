// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"sync"
	"time"

	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/notify"
)

// Set up dummy responses for info data expected from CRM.
// Used for unit testing without a real CRM present.
// Functionality here emulates crmutil/controller-data.go.

// Add in a little day to allow intermediate state changes to be seen.
// But don't add too much because it increases the unit test run time.
const DummyInfoDelay = 10 * time.Millisecond

type DummyInfoResponder struct {
	AppInstCache                 *edgeproto.AppInstCache
	AppInstInfoCache             edgeproto.AppInstInfoCache
	ClusterInstCache             *edgeproto.ClusterInstCache
	ClusterInstInfoCache         edgeproto.ClusterInstInfoCache
	RecvAppInstInfo              notify.RecvAppInstInfoHandler
	RecvClusterInstInfo          notify.RecvClusterInstInfoHandler
	VMPoolCache                  *edgeproto.VMPoolCache
	VMPoolInfoCache              edgeproto.VMPoolInfoCache
	RecvVMPoolInfo               notify.RecvVMPoolInfoHandler
	simulateAppCreateFailure     bool
	simulateAppUpdateFailure     bool
	simulateAppDeleteFailure     bool
	simulateClusterCreateFailure bool
	simulateClusterUpdateFailure bool
	simulateClusterDeleteFailure bool
	simulateVMPoolUpdateFailure  bool
	enable                       bool
	pause                        sync.WaitGroup
}

func (d *DummyInfoResponder) InitDummyInfoResponder() {
	d.enable = true
	// init appinst obj
	if d.AppInstCache != nil {
		d.AppInstCache.SetNotifyCb(d.runAppInstChanged)
		d.AppInstCache.SetDeletedCb(d.runAppInstDeleted)
	}
	edgeproto.InitAppInstInfoCache(&d.AppInstInfoCache)
	d.AppInstInfoCache.SetNotifyCb(d.appInstInfoCb)
	// init clusterinst obj
	if d.ClusterInstCache != nil {
		d.ClusterInstCache.SetNotifyCb(d.runClusterInstChanged)
		d.ClusterInstCache.SetDeletedCb(d.runClusterInstDeleted)
	}
	edgeproto.InitClusterInstInfoCache(&d.ClusterInstInfoCache)
	d.ClusterInstInfoCache.SetNotifyCb(d.clusterInstInfoCb)
	// init vmpool obj
	if d.VMPoolCache != nil {
		d.VMPoolCache.SetNotifyCb(d.runVMPoolChanged)
		d.VMPoolCache.SetDeletedCb(d.runVMPoolDeleted)
	}
	edgeproto.InitVMPoolInfoCache(&d.VMPoolInfoCache)
	d.VMPoolInfoCache.SetNotifyCb(d.vmPoolInfoCb)
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

func (d *DummyInfoResponder) SetSimulateVMPoolUpdateFailure(state bool) {
	d.simulateVMPoolUpdateFailure = state
}

// Pauses responder until unpaused.
// Warning: don't double-pause or double-unpause.
func (d *DummyInfoResponder) SetPause(enable bool) {
	if enable {
		d.pause.Add(1)
	} else {
		d.pause.Done()
	}
}

func (d *DummyInfoResponder) runClusterInstChanged(ctx context.Context, key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst, modRev int64) {
	if !d.enable {
		return
	}
	// copy out from cache since data may change while thread runs
	inst := edgeproto.ClusterInst{}
	found := d.ClusterInstCache.Get(key, &inst)
	if !found {
		return
	}
	go d.clusterInstChanged(ctx, &inst)
}

func (d *DummyInfoResponder) runClusterInstDeleted(ctx context.Context, old *edgeproto.ClusterInst) {
	if !d.enable {
		return
	}
	copy := &edgeproto.ClusterInst{}
	copy.DeepCopyIn(old)
	go d.clusterInstDeleted(ctx, copy)
}

func (d *DummyInfoResponder) runAppInstChanged(ctx context.Context, key *edgeproto.AppInstKey, old *edgeproto.AppInst, modRev int64) {
	if !d.enable {
		return
	}
	// copy out from cache since data may change while thread runs
	inst := edgeproto.AppInst{}
	found := d.AppInstCache.Get(key, &inst)
	if !found {
		return
	}
	go d.appInstChanged(ctx, &inst)
}

func (d *DummyInfoResponder) runAppInstDeleted(ctx context.Context, old *edgeproto.AppInst) {
	if !d.enable {
		return
	}
	copy := &edgeproto.AppInst{}
	copy.DeepCopyIn(old)
	go d.appInstDeleted(ctx, copy)
}

func (d *DummyInfoResponder) runVMPoolChanged(ctx context.Context, key *edgeproto.VMPoolKey, old *edgeproto.VMPool, modRev int64) {
	if !d.enable {
		return
	}
	// copy out from cache since data may change while thread runs
	inst := edgeproto.VMPool{}
	found := d.VMPoolCache.Get(key, &inst)
	if !found {
		return
	}
	go d.vmPoolChanged(ctx, &inst)
}

func (d *DummyInfoResponder) runVMPoolDeleted(ctx context.Context, old *edgeproto.VMPool) {
	if !d.enable {
		return
	}
	copy := &edgeproto.VMPool{}
	copy.DeepCopyIn(old)
	go d.vmPoolDeleted(ctx, copy)
}

func (d *DummyInfoResponder) clusterInstChanged(ctx context.Context, inst *edgeproto.ClusterInst) {
	key := &inst.Key
	if inst.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		// update
		log.SpanLog(ctx, log.DebugLevelApi, "Update ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_UPDATING)
		time.Sleep(DummyInfoDelay)
		d.pause.Wait()
		log.SpanLog(ctx, log.DebugLevelApi, "ClusterInst ready", "key", key)
		if d.simulateClusterUpdateFailure {
			d.ClusterInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_UPDATE_ERROR, "crm update ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// create
		log.SpanLog(ctx, log.DebugLevelApi, "Create ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_CREATING)
		time.Sleep(DummyInfoDelay)
		d.pause.Wait()
		log.SpanLog(ctx, log.DebugLevelApi, "ClusterInst ready", "key", key)
		if d.simulateClusterCreateFailure {
			d.ClusterInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_CREATE_ERROR, "crm create ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_DELETE_REQUESTED {
		// delete
		log.SpanLog(ctx, log.DebugLevelApi, "Delete ClusterInst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_DELETING)
		time.Sleep(DummyInfoDelay)
		d.pause.Wait()
		log.SpanLog(ctx, log.DebugLevelApi, "ClusterInst deleted", "key", key)
		if d.simulateClusterDeleteFailure {
			d.ClusterInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_DELETE_ERROR, "crm delete ClusterInst failed")
		} else {
			d.ClusterInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_DELETE_DONE)
		}
	}
}

func (d *DummyInfoResponder) clusterInstDeleted(ctx context.Context, old *edgeproto.ClusterInst) {
	info := edgeproto.ClusterInstInfo{Key: old.Key}
	d.ClusterInstInfoCache.Delete(ctx, &info, 0)
}

func (d *DummyInfoResponder) appInstChanged(ctx context.Context, inst *edgeproto.AppInst) {
	key := &inst.Key
	if inst.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		// update
		log.SpanLog(ctx, log.DebugLevelApi, "Update app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_UPDATING)
		time.Sleep(DummyInfoDelay)
		d.pause.Wait()
		log.SpanLog(ctx, log.DebugLevelApi, "app inst ready", "key", key)
		if d.simulateAppUpdateFailure {
			d.AppInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_UPDATE_ERROR, "crm update app inst failed")
		} else {
			d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// create
		log.SpanLog(ctx, log.DebugLevelApi, "Create app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_CREATING)
		time.Sleep(DummyInfoDelay)
		d.pause.Wait()
		log.SpanLog(ctx, log.DebugLevelApi, "app inst ready", "key", key)
		if d.simulateAppCreateFailure {
			d.AppInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_CREATE_ERROR, "crm create app inst failed")
		} else {
			d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
		}
	} else if inst.State == edgeproto.TrackedState_DELETE_REQUESTED {
		// delete
		log.SpanLog(ctx, log.DebugLevelApi, "Delete app inst", "key", key)
		time.Sleep(DummyInfoDelay)
		d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_DELETING)
		time.Sleep(DummyInfoDelay)
		d.pause.Wait()
		log.SpanLog(ctx, log.DebugLevelApi, "app inst deleted", "key", key)
		if d.simulateAppDeleteFailure {
			d.AppInstInfoCache.SetError(ctx, key, edgeproto.TrackedState_DELETE_ERROR, "crm delete app inst failed")
		} else {
			d.AppInstInfoCache.SetState(ctx, key, edgeproto.TrackedState_DELETE_DONE)
		}
	}
}

func (d *DummyInfoResponder) appInstDeleted(ctx context.Context, old *edgeproto.AppInst) {
	info := edgeproto.AppInstInfo{Key: old.Key}
	d.AppInstInfoCache.Delete(ctx, &info, 0)
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

func (d *DummyInfoResponder) vmPoolInfoCb(ctx context.Context, key *edgeproto.VMPoolKey, old *edgeproto.VMPoolInfo, modRev int64) {
	info := edgeproto.VMPoolInfo{}
	if d.VMPoolInfoCache.Get(key, &info) {
		d.RecvVMPoolInfo.Update(ctx, &info, modRev)
	} else {
		info.Key = *key
		d.RecvVMPoolInfo.Delete(ctx, &info, modRev)
	}
}

func (d *DummyInfoResponder) vmPoolChanged(ctx context.Context, inst *edgeproto.VMPool) {
	if inst.State != edgeproto.TrackedState_UPDATE_REQUESTED {
		return
	}
	key := &inst.Key
	// update
	log.SpanLog(ctx, log.DebugLevelApi, "Update VMPool", "key", key)
	time.Sleep(DummyInfoDelay)
	d.VMPoolInfoCache.SetState(ctx, key, edgeproto.TrackedState_UPDATING)
	if d.simulateVMPoolUpdateFailure {
		d.VMPoolInfoCache.SetError(ctx, key, edgeproto.TrackedState_UPDATE_ERROR, "crm update VMPool failed")
	} else {
		newVMs := []edgeproto.VM{}
		for _, vm := range inst.Vms {
			switch vm.State {
			case edgeproto.VMState_VM_ADD:
				vm.State = edgeproto.VMState_VM_FREE
				newVMs = append(newVMs, vm)
			case edgeproto.VMState_VM_REMOVE:
				continue
			case edgeproto.VMState_VM_UPDATE:
				vm.State = edgeproto.VMState_VM_FREE
				newVMs = append(newVMs, vm)
			default:
				newVMs = append(newVMs, vm)
			}
		}
		// save VM to VM pool
		info := edgeproto.VMPoolInfo{}
		if !d.VMPoolInfoCache.Get(&inst.Key, &info) {
			info.Key = *key
		}
		info.Vms = newVMs
		d.VMPoolInfoCache.Update(ctx, &info, 0)

		log.SpanLog(ctx, log.DebugLevelApi, "VMPool ready", "key", key)
		d.VMPoolInfoCache.SetState(ctx, key, edgeproto.TrackedState_READY)
	}
}

func (d *DummyInfoResponder) vmPoolDeleted(ctx context.Context, old *edgeproto.VMPool) {
	info := edgeproto.VMPoolInfo{Key: old.Key}
	d.VMPoolInfoCache.Delete(ctx, &info, 0)
}
