package crmutil

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

//ControllerData contains cache data for controller
type ControllerData struct {
	platform             platform.Platform
	AppCache             edgeproto.AppCache
	AppInstCache         edgeproto.AppInstCache
	CloudletCache        edgeproto.CloudletCache
	FlavorCache          edgeproto.FlavorCache
	ClusterFlavorCache   edgeproto.ClusterFlavorCache
	ClusterInstCache     edgeproto.ClusterInstCache
	AppInstInfoCache     edgeproto.AppInstInfoCache
	CloudletInfoCache    edgeproto.CloudletInfoCache
	ClusterInstInfoCache edgeproto.ClusterInstInfoCache
	NodeCache            edgeproto.NodeCache
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
	edgeproto.InitClusterFlavorCache(&cd.ClusterFlavorCache)
	edgeproto.InitClusterInstCache(&cd.ClusterInstCache)
	edgeproto.InitNodeCache(&cd.NodeCache)
	// set callbacks to trigger changes
	cd.ClusterInstCache.SetNotifyCb(cd.clusterInstChanged)
	cd.AppInstCache.SetNotifyCb(cd.appInstChanged)
	cd.FlavorCache.SetNotifyCb(cd.flavorChanged)
	cd.ClusterFlavorCache.SetNotifyCb(cd.clusterFlavorChanged)
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
		info.State = edgeproto.CloudletState_CloudletStateErrors
	} else {
		// Is the cloudlet ready at this point?
		info.Errors = nil
		info.State = edgeproto.CloudletState_CloudletStateReady
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
	//   cd.clusterInstInfoState(key, edgeproto.TrackedState_Ready)
	//   for _, app := range MexAppShowAppInst(cluster) {
	//      key := get key from app
	//      cd.appInstInfoState(key, edgeproto.TrackedState_Ready)
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

func (cd *ControllerData) clusterFlavorChanged(key *edgeproto.ClusterFlavorKey, old *edgeproto.ClusterFlavor) {
	flavor := edgeproto.ClusterFlavor{}
	found := cd.ClusterFlavorCache.Get(key, &flavor)
	if found {
		// create (no updates allowed)
		// CRM TODO: register cluster flavor?
	} else {
		// CRM TODO: delete cluster flavor?
	}
}

func (cd *ControllerData) clusterInstChanged(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst) {
	log.DebugLog(log.DebugLevelMexos, "clusterInstChange", "key", key)
	clusterInst := edgeproto.ClusterInst{}
	found := cd.ClusterInstCache.Get(key, &clusterInst)
	if !found {
		return
	}
	log.DebugLog(log.DebugLevelMexos, "clusterInst state", "state", clusterInst.State)

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
		if info.State == edgeproto.TrackedState_Creating || info.State == edgeproto.TrackedState_Updating || info.State == edgeproto.TrackedState_Deleting {
			log.DebugLog(log.DebugLevelMexos, "clusterInst conflicting state", "state", info.State)
			return
		}
	}
	// do request
	if clusterInst.State == edgeproto.TrackedState_CreateRequested {
		// create
		log.DebugLog(log.DebugLevelMexos, "cluster inst create", "clusterInst", clusterInst)
		// create or update k8s cluster on this cloudlet
		flavor := edgeproto.ClusterFlavor{}

		// XXX clusterInstCache has clusterInst but FlavorCache has clusterInst.Flavor.
		flavorFound := cd.ClusterFlavorCache.Get(&clusterInst.Flavor, &flavor)
		if !flavorFound {
			log.DebugLog(log.DebugLevelMexos, "did not find flavor", "flavor", flavor)
			//XXX returning flavor not found error to InstInfoError?
			cd.clusterInstInfoError(key, edgeproto.TrackedState_CreateError, fmt.Sprintf("Did not find flavor %s", clusterInst.Flavor.Name))
			return
		}
		log.DebugLog(log.DebugLevelMexos, "Found flavor", "flavor", flavor)
		cd.clusterInstInfoState(key, edgeproto.TrackedState_Creating)
		go func() {
			var err error
			log.DebugLog(log.DebugLevelMexos, "create cluster inst", "clusterinst", clusterInst)

			err = cd.platform.CreateCluster(&clusterInst, &flavor)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "error cluster create fail", "error", err)
				cd.clusterInstInfoError(key, edgeproto.TrackedState_CreateError, fmt.Sprintf("Create failed: %s", err))
				//XXX seems clusterInstInfoError is overloaded with status for flavor and clustinst.
				return
			}
			log.DebugLog(log.DebugLevelMexos, "adding flavor", "flavor", flavor)
			/*
				Add Flavor never actually did anything.  TODO: implement this or decide against doing it.
				err = mexos.MEXAddFlavorClusterInst(cd.CRMRootLB, &flavor) //Flavor is inside ClusterInst even though it comes from FlavorCache
				if err != nil {
					log.DebugLog(log.DebugLevelMexos, "cannot add flavor", "flavor", flavor)
					cd.clusterInstInfoError(key, edgeproto.TrackedState_CreateError, fmt.Sprintf("Can't add flavor %s, %v", flavor.Key.Name, err))
					return
				}
			*/
			log.DebugLog(log.DebugLevelMexos, "cluster state ready", "clusterinst", clusterInst)
			cd.clusterInstInfoState(key, edgeproto.TrackedState_Ready)
		}()
	} else if clusterInst.State == edgeproto.TrackedState_UpdateRequested {
		// update (TODO)
	} else if clusterInst.State == edgeproto.TrackedState_DeleteRequested {
		log.DebugLog(log.DebugLevelMexos, "cluster inst delete", "clusterinst", clusterInst)
		// clusterInst was deleted
		cd.clusterInstInfoState(key, edgeproto.TrackedState_Deleting)
		go func() {
			var err error
			log.DebugLog(log.DebugLevelMexos, "delete cluster inst", "clusterinst", clusterInst)
			err = cd.platform.DeleteCluster(&clusterInst)
			if err != nil {
				str := fmt.Sprintf("Delete failed: %s", err)
				cd.clusterInstInfoError(key, edgeproto.TrackedState_DeleteError, str)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "set cluster inst deleted", "clusterinst", clusterInst)
			// Deleting local info signals to controller that
			// delete was successful.
			info := edgeproto.ClusterInstInfo{Key: *key}
			cd.ClusterInstInfoCache.Delete(&info, 0)
		}()
	} else if clusterInst.State == edgeproto.TrackedState_Creating {
		cd.clusterInstInfoCheckState(key, edgeproto.TrackedState_Creating,
			edgeproto.TrackedState_Ready,
			edgeproto.TrackedState_CreateError)
	} else if clusterInst.State == edgeproto.TrackedState_Updating {
		cd.clusterInstInfoCheckState(key, edgeproto.TrackedState_Updating,
			edgeproto.TrackedState_Ready,
			edgeproto.TrackedState_UpdateError)
	} else if clusterInst.State == edgeproto.TrackedState_Deleting {
		cd.clusterInstInfoCheckState(key, edgeproto.TrackedState_Deleting,
			edgeproto.TrackedState_NotPresent,
			edgeproto.TrackedState_DeleteError)
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
		if info.State == edgeproto.TrackedState_Creating || info.State == edgeproto.TrackedState_Updating || info.State == edgeproto.TrackedState_Deleting {
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
	if appInst.State == edgeproto.TrackedState_CreateRequested {
		// create
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&appInst.Flavor, &flavor)
		if !flavorFound {
			str := fmt.Sprintf("Flavor %s not found",
				appInst.Flavor.Name)
			cd.appInstInfoError(key, edgeproto.TrackedState_CreateError, str)
			return
		}
		clusterInst := edgeproto.ClusterInst{}
		clusterInstFound := cd.ClusterInstCache.Get(&appInst.ClusterInstKey, &clusterInst)
		if !clusterInstFound {
			str := fmt.Sprintf("Cluster instance %s not found",
				appInst.ClusterInstKey.ClusterKey.Name)
			cd.appInstInfoError(key, edgeproto.TrackedState_CreateError, str)
			return
		}

		cd.appInstInfoState(key, edgeproto.TrackedState_Creating)
		go func() {
			log.DebugLog(log.DebugLevelMexos, "update kube config", "appinst", appInst, "clusterinst", clusterInst)

			names, err := k8smgmt.GetKubeNames(&clusterInst, &app, &appInst)
			if err != nil {
				errstr := fmt.Sprintf("get kube names failed: %s", err)
				cd.appInstInfoError(key, edgeproto.TrackedState_CreateError, errstr)
				log.DebugLog(log.DebugLevelMexos, "can't create app inst", "error", errstr, "key", key)
				return
			}
			err = cd.platform.CreateAppInst(&clusterInst, &app, &appInst, names)
			if err != nil {
				errstr := fmt.Sprintf("Create App Inst failed: %s", err)
				cd.appInstInfoError(key, edgeproto.TrackedState_CreateError, errstr)
				log.DebugLog(log.DebugLevelMexos, "can't create app inst", "error", errstr, "key", key)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "created docker app inst", "appisnt", appInst, "clusterinst", clusterInst)
			cd.appInstInfoState(key, edgeproto.TrackedState_Ready)
		}()
	} else if appInst.State == edgeproto.TrackedState_UpdateRequested {
		// update (TODO)
	} else if appInst.State == edgeproto.TrackedState_DeleteRequested {
		clusterInst := edgeproto.ClusterInst{}
		clusterInstFound := cd.ClusterInstCache.Get(&appInst.ClusterInstKey, &clusterInst)
		if !clusterInstFound {
			str := fmt.Sprintf("Cluster instance %s not found",
				appInst.ClusterInstKey.ClusterKey.Name)
			cd.appInstInfoError(key, edgeproto.TrackedState_DeleteError, str)
			return
		}
		// appInst was deleted
		cd.appInstInfoState(key, edgeproto.TrackedState_Deleting)
		go func() {
			log.DebugLog(log.DebugLevelMexos, "delete app inst", "appinst", appInst, "clusterinst", clusterInst)
			names, err := k8smgmt.GetKubeNames(&clusterInst, &app, &appInst)
			if err != nil {
				errstr := fmt.Sprintf("get kube names failed: %s", err)
				cd.appInstInfoError(key, edgeproto.TrackedState_CreateError, errstr)
				log.DebugLog(log.DebugLevelMexos, "can't delete app inst", "error", errstr, "key", key)
				return
			}
			err = cd.platform.DeleteAppInst(&clusterInst, &app, &appInst, names)
			if err != nil {
				errstr := fmt.Sprintf("Delete App Inst failed: %s", err)
				cd.appInstInfoError(key, edgeproto.TrackedState_DeleteError, errstr)
				log.DebugLog(log.DebugLevelMexos, "can't delete app inst", "error", errstr, "key", key)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "deleted docker app inst", "appisnt", appInst, "clusterinst", clusterInst)
			// Deleting local info signals to controller that
			// delete was successful.
			info := edgeproto.AppInstInfo{Key: *key}
			cd.AppInstInfoCache.Delete(&info, 0)
		}()
	} else if appInst.State == edgeproto.TrackedState_Creating {
		// Controller may send a CRM transitional state after a
		// disconnect or crash. Controller thinks CRM is creating
		// the appInst, and Controller is waiting for a new state from
		// the CRM. If CRM is not creating, or has not just finished
		// creating (ready), set an error state.
		cd.appInstInfoCheckState(key, edgeproto.TrackedState_Creating,
			edgeproto.TrackedState_Ready,
			edgeproto.TrackedState_CreateError)
	} else if appInst.State == edgeproto.TrackedState_Updating {
		cd.appInstInfoCheckState(key, edgeproto.TrackedState_Updating,
			edgeproto.TrackedState_Ready,
			edgeproto.TrackedState_UpdateError)
	} else if appInst.State == edgeproto.TrackedState_Deleting {
		cd.appInstInfoCheckState(key, edgeproto.TrackedState_Deleting,
			edgeproto.TrackedState_NotPresent,
			edgeproto.TrackedState_DeleteError)
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

// CheckState checks that the info is either in the transState or finalState.
// If not, it is an unexpected state, so we set it to the error state.
// This is used when the controller sends CRM a state that implies the
// controller is waiting for the CRM to send back the next state, but the
// CRM does not have any change in progress.
func (cd *ControllerData) clusterInstInfoCheckState(key *edgeproto.ClusterInstKey, transState, finalState, errState edgeproto.TrackedState) {
	cd.ClusterInstInfoCache.UpdateModFunc(key, 0, func(old *edgeproto.ClusterInstInfo) (newObj *edgeproto.ClusterInstInfo, changed bool) {
		if old == nil {
			if transState == edgeproto.TrackedState_NotPresent || finalState == edgeproto.TrackedState_NotPresent {
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
			if transState == edgeproto.TrackedState_NotPresent || finalState == edgeproto.TrackedState_NotPresent {
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
