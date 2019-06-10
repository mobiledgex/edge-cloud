package crmutil

import (
	"fmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

//ControllerData contains cache data for controller
type ControllerData struct {
	platform             platform.Platform
	AppCache             edgeproto.AppCache
	AppInstCache         edgeproto.AppInstCache
	CloudletCache        edgeproto.CloudletCache
	FlavorCache          edgeproto.FlavorCache
	ClusterInstCache     edgeproto.ClusterInstCache
	AppInstInfoCache     edgeproto.AppInstInfoCache
	CloudletInfoCache    edgeproto.CloudletInfoCache
	ClusterInstInfoCache edgeproto.ClusterInstInfoCache
	NodeCache            edgeproto.NodeCache
	ExecReqHandler       *ExecReqHandler
	ExecReqSend          *notify.ExecRequestSend
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData(pf platform.Platform) *ControllerData {
	cd := &ControllerData{}
	cd.platform = pf
	edgeproto.InitAppCache(&cd.AppCache)
	edgeproto.InitAppInstCache(&cd.AppInstCache)
	edgeproto.InitCloudletCache(&cd.CloudletCache)
	edgeproto.InitAppInstInfoCache(&cd.AppInstInfoCache)
	edgeproto.InitClusterInstInfoCache(&cd.ClusterInstInfoCache)
	edgeproto.InitCloudletInfoCache(&cd.CloudletInfoCache)
	edgeproto.InitFlavorCache(&cd.FlavorCache)
	edgeproto.InitClusterInstCache(&cd.ClusterInstCache)
	edgeproto.InitNodeCache(&cd.NodeCache)
	cd.ExecReqHandler = NewExecReqHandler(cd)
	cd.ExecReqSend = notify.NewExecRequestSend()
	// set callbacks to trigger changes
	cd.ClusterInstCache.SetNotifyCb(cd.clusterInstChanged)
	cd.AppInstCache.SetNotifyCb(cd.appInstChanged)
	cd.FlavorCache.SetNotifyCb(cd.flavorChanged)
	return cd
}

// GatherCloudletInfo gathers all the information about the Cloudlet that
// the controller needs to be able to manage it.
func (cd *ControllerData) GatherCloudletInfo(info *edgeproto.CloudletInfo) {
	log.DebugLog(log.DebugLevelMexos, "attempting to gather cloudlet info")
	err := cd.platform.GatherCloudletInfo(info)
	if err != nil {
		str := fmt.Sprintf("get limits failed: %s", err)
		info.Errors = append(info.Errors, str)
		info.State = edgeproto.CloudletState_CLOUDLET_STATE_ERRORS
	} else {
		// Is the cloudlet ready at this point?
		info.Errors = nil
		info.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
		log.DebugLog(log.DebugLevelMexos, "cloudlet state ready", "info", info)
	}
}

// GetInsts queries Openstack/Kubernetes to get all the cluster insts
// and app insts that have been created on the Cloudlet.
// It is called once at startup, and is used to repopulate the cache
// after CRM restart/crash. When the CRM connects to the controller,
// it will send the insts in the cache and the controller will resolve
// any discrepancies between the CRM's current state versus the
// controller's intended state.
//
// The controller does not know about all the steps that are used to
// create/delete a ClusterInst/AppInst, so if the CRM crashed in the
// middle of such a task, it is up to the CRM to clean up any unfinished
// state.
func (cd *ControllerData) GatherInsts() {
	// TODO: Implement me.
	// for _, cluster := range MexClusterShowClustInst() {
	//   key := get key from cluster
	//   cd.clusterInstInfoState(key, edgeproto.TrackedState_READY)
	//   for _, app := range MexAppShowAppInst(cluster) {
	//      key := get key from app
	//      cd.appInstInfoState(key, edgeproto.TrackedState_READY)
	//   }
	// }
}

// Note: these callback functions are called in the context of
// the notify receive thread. If the actions done here not quick,
// they should be done in a separate worker thread.

func (cd *ControllerData) flavorChanged(key *edgeproto.FlavorKey, old *edgeproto.Flavor) {
	flavor := edgeproto.Flavor{}
	found := cd.FlavorCache.Get(key, &flavor)
	if found {
		// create (no updates allowed)
		// CRM TODO: register flavor?
	} else {
		// CRM TODO: delete flavor?
	}
}

func (cd *ControllerData) clusterInstChanged(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst) {
	log.DebugLog(log.DebugLevelMexos, "clusterInstChange", "key", key, "old", old)
	clusterInst := edgeproto.ClusterInst{}
	found := cd.ClusterInstCache.Get(key, &clusterInst)
	if !found {
		return
	}
	log.DebugLog(log.DebugLevelMexos, "clusterInst state", "state", clusterInst.State, "new", clusterInst)

	// If CRM crashes or reconnects to controller, controller will resend
	// current state. This is needed to:
	// -restart actions that were lost due to a crash
	// -update cache for dependent objects (AppInst looks up ClusterInst from
	// cache).
	// If it was a disconnect and not a restart, there may alread be a
	// thread in progress. To prevent multiple conflicting threads, check
	// the info state which can tell us if a thread is in progress.
	info := edgeproto.ClusterInstInfo{}
	if infoFound := cd.ClusterInstInfoCache.Get(key, &info); infoFound {
		if info.State == edgeproto.TrackedState_CREATING || info.State == edgeproto.TrackedState_UPDATING || info.State == edgeproto.TrackedState_DELETING {
			log.DebugLog(log.DebugLevelMexos, "clusterInst conflicting state", "state", info.State)
			return
		}
	}

	updateClusterCacheCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			cd.ClusterInstInfoCache.SetStatusTask(key, value)
		case edgeproto.UpdateStep:
			cd.ClusterInstInfoCache.SetStatusStep(key, value)
		}
	}

	// do request
	if clusterInst.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// create
		log.DebugLog(log.DebugLevelMexos, "cluster inst create", "clusterInst", clusterInst)
		// create or update k8s cluster on this cloudlet
		cd.clusterInstInfoState(key, edgeproto.TrackedState_CREATING)
		go func() {
			var err error
			log.DebugLog(log.DebugLevelMexos, "create cluster inst", "clusterinst", clusterInst)
			err = cd.platform.CreateClusterInst(&clusterInst, updateClusterCacheCallback)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "error cluster create fail", "error", err)
				cd.clusterInstInfoError(key, edgeproto.TrackedState_CREATE_ERROR, fmt.Sprintf("Create failed: %s", err))
				//XXX seems clusterInstInfoError is overloaded with status for flavor and clustinst.
				return
			}
			log.DebugLog(log.DebugLevelMexos, "cluster state ready", "clusterinst", clusterInst)
			cd.clusterInstInfoState(key, edgeproto.TrackedState_READY)
		}()
	} else if clusterInst.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		log.DebugLog(log.DebugLevelMexos, "cluster inst update", "clusterinst", clusterInst)
		cd.clusterInstInfoState(key, edgeproto.TrackedState_UPDATING)
		var err error
		log.DebugLog(log.DebugLevelMexos, "update cluster inst", "clusterinst", clusterInst)

		err = cd.platform.UpdateClusterInst(&clusterInst, updateClusterCacheCallback)
		if err != nil {
			str := fmt.Sprintf("update failed: %s", err)
			cd.clusterInstInfoError(key, edgeproto.TrackedState_DELETE_ERROR, str)
			return
		}

		log.DebugLog(log.DebugLevelMexos, "cluster state ready", "clusterinst", clusterInst)
		cd.clusterInstInfoState(key, edgeproto.TrackedState_READY)

	} else if clusterInst.State == edgeproto.TrackedState_DELETE_REQUESTED {
		log.DebugLog(log.DebugLevelMexos, "cluster inst delete", "clusterinst", clusterInst)
		// clusterInst was deleted
		cd.clusterInstInfoState(key, edgeproto.TrackedState_DELETING)
		go func() {
			var err error
			log.DebugLog(log.DebugLevelMexos, "delete cluster inst", "clusterinst", clusterInst)
			err = cd.platform.DeleteClusterInst(&clusterInst)
			if err != nil {
				str := fmt.Sprintf("Delete failed: %s", err)
				cd.clusterInstInfoError(key, edgeproto.TrackedState_DELETE_ERROR, str)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "set cluster inst deleted", "clusterinst", clusterInst)
			// DELETING local info signals to controller that
			// delete was successful.
			info := edgeproto.ClusterInstInfo{Key: *key}
			cd.ClusterInstInfoCache.Delete(&info, 0)
		}()
	} else if clusterInst.State == edgeproto.TrackedState_CREATING {
		cd.clusterInstInfoCheckState(key, edgeproto.TrackedState_CREATING,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_CREATE_ERROR)
	} else if clusterInst.State == edgeproto.TrackedState_UPDATING {
		cd.clusterInstInfoCheckState(key, edgeproto.TrackedState_UPDATING,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_UPDATE_ERROR)
	} else if clusterInst.State == edgeproto.TrackedState_DELETING {
		cd.clusterInstInfoCheckState(key, edgeproto.TrackedState_DELETING,
			edgeproto.TrackedState_NOT_PRESENT,
			edgeproto.TrackedState_DELETE_ERROR)
	}
}

func (cd *ControllerData) appInstChanged(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	log.DebugLog(log.DebugLevelMexos, "app inst changed", "key", key)
	appInst := edgeproto.AppInst{}
	found := cd.AppInstCache.Get(key, &appInst)
	if !found {
		return
	}
	// Check current thread state. See comment in clusterInstChanged.
	info := edgeproto.AppInstInfo{}
	if infoFound := cd.AppInstInfoCache.Get(key, &info); infoFound {
		if info.State == edgeproto.TrackedState_CREATING || info.State == edgeproto.TrackedState_UPDATING || info.State == edgeproto.TrackedState_DELETING {
			return
		}
	}
	app := edgeproto.App{}
	found = cd.AppCache.Get(&key.AppKey, &app)
	if !found {
		log.DebugLog(log.DebugLevelMexos, "App not found for AppInst", "key", key)
		return
	}

	// do request
	log.DebugLog(log.DebugLevelMexos, "appInstChanged", "appInst", appInst)
	updateAppCacheCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			cd.AppInstInfoCache.SetStatusTask(&appInst.Key, value)
		case edgeproto.UpdateStep:
			cd.AppInstInfoCache.SetStatusStep(&appInst.Key, value)
		}
	}

	if appInst.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// create
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&appInst.Flavor, &flavor)
		if !flavorFound {
			str := fmt.Sprintf("Flavor %s not found",
				appInst.Flavor.Name)
			cd.appInstInfoError(key, edgeproto.TrackedState_CREATE_ERROR, str)
			return
		}
		clusterInst := edgeproto.ClusterInst{}
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInstFound := cd.ClusterInstCache.Get(&appInst.Key.ClusterInstKey, &clusterInst)
			if !clusterInstFound {
				str := fmt.Sprintf("Cluster instance %s not found",
					appInst.Key.ClusterInstKey.ClusterKey.Name)
				cd.appInstInfoError(key, edgeproto.TrackedState_CREATE_ERROR, str)
				return
			}
		}

		cd.appInstInfoState(key, edgeproto.TrackedState_CREATING)

		go func() {
			log.DebugLog(log.DebugLevelMexos, "update kube config", "appinst", appInst, "clusterinst", clusterInst)

			err := cd.platform.CreateAppInst(&clusterInst, &app, &appInst, &flavor, updateAppCacheCallback)
			if err != nil {
				errstr := fmt.Sprintf("Create App Inst failed: %s", err)
				cd.appInstInfoError(key, edgeproto.TrackedState_CREATE_ERROR, errstr)
				log.InfoLog("can't create app inst", "error", errstr, "key", key)
				log.DebugLog(log.DebugLevelMexos, "cleaning up failed appinst", "key", key)
				derr := cd.platform.DeleteAppInst(&clusterInst, &app, &appInst)
				if derr != nil {
					log.InfoLog("can't cleanup app inst", "error", errstr, "key", key)
				}
				return
			}
			log.DebugLog(log.DebugLevelMexos, "created app inst", "appisnt", appInst, "clusterinst", clusterInst)

			rt, err := cd.platform.GetAppInstRuntime(&clusterInst, &app, &appInst)
			if err != nil {
				log.InfoLog("unable to get AppInstRuntime", "key", key, "err", err)
				cd.appInstInfoState(key, edgeproto.TrackedState_READY)
			} else {
				cd.appInstInfoRuntime(key, edgeproto.TrackedState_READY, rt)
			}
		}()
	} else if appInst.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		// update (TODO)
	} else if appInst.State == edgeproto.TrackedState_DELETE_REQUESTED {
		clusterInst := edgeproto.ClusterInst{}
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInstFound := cd.ClusterInstCache.Get(&appInst.Key.ClusterInstKey, &clusterInst)
			if !clusterInstFound {
				str := fmt.Sprintf("Cluster instance %s not found",
					appInst.Key.ClusterInstKey.ClusterKey.Name)
				cd.appInstInfoError(key, edgeproto.TrackedState_DELETE_ERROR, str)
				return
			}
		}
		// appInst was deleted
		cd.appInstInfoState(key, edgeproto.TrackedState_DELETING)
		go func() {
			log.DebugLog(log.DebugLevelMexos, "delete app inst", "appinst", appInst, "clusterinst", clusterInst)

			err := cd.platform.DeleteAppInst(&clusterInst, &app, &appInst)
			if err != nil {
				errstr := fmt.Sprintf("Delete App Inst failed: %s", err)
				cd.appInstInfoError(key, edgeproto.TrackedState_DELETE_ERROR, errstr)
				log.DebugLog(log.DebugLevelMexos, "can't delete app inst", "error", errstr, "key", key)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "deleted app inst", "appisnt", appInst, "clusterinst", clusterInst)
			// DELETING local info signals to controller that
			// delete was successful.
			info := edgeproto.AppInstInfo{Key: *key}
			cd.AppInstInfoCache.Delete(&info, 0)
		}()
	} else if appInst.State == edgeproto.TrackedState_CREATING {
		// Controller may send a CRM transitional state after a
		// disconnect or crash. Controller thinks CRM is creating
		// the appInst, and Controller is waiting for a new state from
		// the CRM. If CRM is not creating, or has not just finished
		// creating (ready), set an error state.
		cd.appInstInfoCheckState(key, edgeproto.TrackedState_CREATING,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_CREATE_ERROR)
	} else if appInst.State == edgeproto.TrackedState_UPDATING {
		cd.appInstInfoCheckState(key, edgeproto.TrackedState_UPDATING,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_UPDATE_ERROR)
	} else if appInst.State == edgeproto.TrackedState_DELETING {
		cd.appInstInfoCheckState(key, edgeproto.TrackedState_DELETING,
			edgeproto.TrackedState_NOT_PRESENT,
			edgeproto.TrackedState_DELETE_ERROR)
	}
}

func (cd *ControllerData) clusterInstInfoError(key *edgeproto.ClusterInstKey, errState edgeproto.TrackedState, err string) {
	cd.ClusterInstInfoCache.SetError(key, errState, err)
}

func (cd *ControllerData) clusterInstInfoState(key *edgeproto.ClusterInstKey, state edgeproto.TrackedState) {
	cd.ClusterInstInfoCache.SetState(key, state)
}

func (cd *ControllerData) appInstInfoError(key *edgeproto.AppInstKey, errState edgeproto.TrackedState, err string) {
	cd.AppInstInfoCache.SetError(key, errState, err)
}

func (cd *ControllerData) appInstInfoState(key *edgeproto.AppInstKey, state edgeproto.TrackedState) {
	cd.AppInstInfoCache.SetState(key, state)
}

func (cd *ControllerData) appInstInfoRuntime(key *edgeproto.AppInstKey, state edgeproto.TrackedState, rt *edgeproto.AppInstRuntime) {
	cd.AppInstInfoCache.SetStateRuntime(key, state, rt)
}

// CheckState checks that the info is either in the transState or finalState.
// If not, it is an unexpected state, so we set it to the error state.
// This is used when the controller sends CRM a state that implies the
// controller is waiting for the CRM to send back the next state, but the
// CRM does not have any change in progress.
func (cd *ControllerData) clusterInstInfoCheckState(key *edgeproto.ClusterInstKey, transState, finalState, errState edgeproto.TrackedState) {
	cd.ClusterInstInfoCache.UpdateModFunc(key, 0, func(old *edgeproto.ClusterInstInfo) (newObj *edgeproto.ClusterInstInfo, changed bool) {
		if old == nil {
			if transState == edgeproto.TrackedState_NOT_PRESENT || finalState == edgeproto.TrackedState_NOT_PRESENT {
				return old, false
			}
			old = &edgeproto.ClusterInstInfo{Key: *key}
		}
		if old.State != transState && old.State != finalState {
			new := &edgeproto.ClusterInstInfo{}
			*new = *old
			new.State = errState
			new.Errors = append(new.Errors, "inconsistent Controller vs CRM state")
			return new, true
		}
		return old, false
	})
}

func (cd *ControllerData) appInstInfoCheckState(key *edgeproto.AppInstKey, transState, finalState, errState edgeproto.TrackedState) {
	cd.AppInstInfoCache.UpdateModFunc(key, 0, func(old *edgeproto.AppInstInfo) (newObj *edgeproto.AppInstInfo, changed bool) {
		if old == nil {
			if transState == edgeproto.TrackedState_NOT_PRESENT || finalState == edgeproto.TrackedState_NOT_PRESENT {
				return old, false
			}
			old = &edgeproto.AppInstInfo{Key: *key}
		}
		if old.State != transState && old.State != finalState {
			new := &edgeproto.AppInstInfo{}
			*new = *old
			new.State = errState
			new.Errors = append(new.Errors, "inconsistent Controller vs CRM state")
			return new, true
		}
		return old, false
	})
}
