package crmutil

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

//ControllerData contains cache data for controller
type ControllerData struct {
	CRMRootLB            *MEXRootLB
	AppInstCache         edgeproto.AppInstCache
	CloudletCache        edgeproto.CloudletCache
	FlavorCache          edgeproto.FlavorCache
	ClusterFlavorCache   edgeproto.ClusterFlavorCache
	ClusterInstCache     edgeproto.ClusterInstCache
	AppInstInfoCache     edgeproto.AppInstInfoCache
	CloudletInfoCache    edgeproto.CloudletInfoCache
	ClusterInstInfoCache edgeproto.ClusterInstInfoCache
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData() *ControllerData {
	cd := &ControllerData{}
	edgeproto.InitAppInstCache(&cd.AppInstCache)
	edgeproto.InitCloudletCache(&cd.CloudletCache)
	edgeproto.InitAppInstInfoCache(&cd.AppInstInfoCache)
	edgeproto.InitClusterInstInfoCache(&cd.ClusterInstInfoCache)
	edgeproto.InitCloudletInfoCache(&cd.CloudletInfoCache)
	edgeproto.InitFlavorCache(&cd.FlavorCache)
	edgeproto.InitClusterFlavorCache(&cd.ClusterFlavorCache)
	edgeproto.InitClusterInstCache(&cd.ClusterInstCache)
	// set callbacks to trigger changes
	cd.ClusterInstCache.SetNotifyCb(cd.clusterInstChanged)
	cd.AppInstCache.SetNotifyCb(cd.appInstChanged)
	cd.FlavorCache.SetNotifyCb(cd.flavorChanged)
	cd.ClusterFlavorCache.SetNotifyCb(cd.clusterFlavorChanged)
	return cd
}

// GatherCloudletInfo gathers all the information about the Cloudlet that
// the controller needs to be able to manage it.
func GatherCloudletInfo(info *edgeproto.CloudletInfo) {
	limits, err := oscli.GetLimits()
	if err != nil {
		str := fmt.Sprintf("Openstack get limits failed: %s", err)
		info.Errors = append(info.Errors, str)
		info.State = edgeproto.CloudletState_CloudletStateErrors
		return
	}

	//XXX only return a subset and only max vals
	for _, l := range limits {
		if l.Name == "MaxTotalCores" {
			info.OsMaxVcores = uint64(l.Value)
		} else if l.Name == "MaxTotalRamSize" {
			info.OsMaxRam = uint64(l.Value)
		} else if l.Name == "MaxTotalVolumeGigabytes" {
			info.OsMaxVolGb = uint64(l.Value)
		}
	}
	// Is the cloudlet ready at this point?
	info.Errors = nil
	info.State = edgeproto.CloudletState_CloudletStateReady
	log.DebugLog(log.DebugLevelMexos, "update limits", "info", info, "limits", limits)
}

// Note: these callback functions are called in the context of
// the notify receive thread. If the actions done here not quick,
// they should be done in a separate worker thread.

func (cd *ControllerData) flavorChanged(key *edgeproto.FlavorKey) {
	flavor := edgeproto.Flavor{}
	found := cd.FlavorCache.Get(key, &flavor)
	if found {
		// create (no updates allowed)
		// CRM TODO: register flavor?
	} else {
		// CRM TODO: delete flavor?
	}
}

func (cd *ControllerData) clusterFlavorChanged(key *edgeproto.ClusterFlavorKey) {
	flavor := edgeproto.ClusterFlavor{}
	found := cd.ClusterFlavorCache.Get(key, &flavor)
	if found {
		// create (no updates allowed)
		// CRM TODO: register cluster flavor?
	} else {
		// CRM TODO: delete cluster flavor?
	}
}

func (cd *ControllerData) clusterInstChanged(key *edgeproto.ClusterInstKey) {
	log.DebugLog(log.DebugLevelMexos, "clusterInstChange", "key", key)
	clusterInst := edgeproto.ClusterInst{}
	found := cd.ClusterInstCache.Get(key, &clusterInst)
	if found {
		log.DebugLog(log.DebugLevelMexos, "cluster inst changed", "clusterInst", clusterInst)
		// create or update k8s cluster on this cloudlet
		cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateBuilding)
		flavor := edgeproto.ClusterFlavor{}

		// XXX clusterInstCache has clusterInst but FlavorCache has clusterInst.Flavor.
		flavorFound := cd.ClusterFlavorCache.Get(&clusterInst.Flavor, &flavor)
		if !flavorFound {
			log.DebugLog(log.DebugLevelMexos, "did not find flavor", "flavor", flavor)
			//XXX returning flavor not found error to InstInfoError?
			cd.clusterInstInfoError(key, fmt.Sprintf("Did not find flavor %s", clusterInst.Flavor.Name))
			return
		}
		log.DebugLog(log.DebugLevelMexos, "Found flavor", "flavor", flavor)
		go func() {
			var err error
			log.DebugLog(log.DebugLevelMexos, "cluster inst changed")
			if !IsValidMEXOSEnv {
				log.DebugLog(log.DebugLevelMexos, "not valid mexos env, fake cluster ready")
				cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateReady)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "create cluster inst", "clusterinst", clusterInst)
			err = MEXClusterCreateClustInst(cd.CRMRootLB, &clusterInst)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "error cluster create fail", "error", err)
				cd.clusterInstInfoError(key, fmt.Sprintf("Create failed: %s", err))
				//XXX seems clusterInstInfoError is overloaded with status for flavor and clustinst.
				return
			}
			log.DebugLog(log.DebugLevelMexos, "adding flavor", "flavor", flavor)
			err = MEXAddFlavorClusterInst(&flavor) //Flavor is inside ClusterInst even though it comes from FlavorCache
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "cannot add flavor", "flavor", flavor)
				cd.clusterInstInfoError(key, fmt.Sprintf("Can't add flavor %s, %v", flavor.Key.Name, err))
				return
			}
			log.DebugLog(log.DebugLevelMexos, "cluster state ready", "clusterinst", clusterInst)
			cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateReady)
		}()
	} else {
		log.DebugLog(log.DebugLevelMexos, "cluster inst deleted", "clusterinst", clusterInst)
		// clusterInst was deleted
		go func() {
			var err error
			log.DebugLog(log.DebugLevelMexos, "cluster inst changed, deleted")
			if !IsValidMEXOSEnv {
				log.DebugLog(log.DebugLevelMexos, "invalid mexos env, fake cluster state deleted")
				cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateDeleted)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "remove cluster inst", "clusterinst", clusterInst)
			err = MEXClusterRemoveClustInst(cd.CRMRootLB, &clusterInst)
			if err != nil {
				str := fmt.Sprintf("Delete failed: %s", err)
				cd.clusterInstInfoError(key, str)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "set cluster inst deleted", "clusterinst", clusterInst)
			cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateDeleted)
		}()
	}
}

func (cd *ControllerData) appInstChanged(key *edgeproto.AppInstKey) {
	log.DebugLog(log.DebugLevelMexos, "app inst changed", "key", key)
	appInst := edgeproto.AppInst{}
	found := cd.AppInstCache.Get(key, &appInst)
	if found {
		// create or update appInst
		cd.appInstInfoState(key, edgeproto.AppState_AppStateBuilding)
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&appInst.Flavor, &flavor)
		if !flavorFound {
			str := fmt.Sprintf("Flavor %s not found",
				appInst.Flavor.Name)
			cd.appInstInfoError(key, str)
			return
		}
		clusterInst := edgeproto.ClusterInst{}
		clusterInstFound := cd.ClusterInstCache.Get(&appInst.ClusterInstKey, &clusterInst)
		if !clusterInstFound {
			str := fmt.Sprintf("Cluster instance %s not found",
				appInst.ClusterInstKey.ClusterKey.Name)
			cd.appInstInfoError(key, str)
			return
		}

		go func() {
			if !IsValidMEXOSEnv {
				log.DebugLog(log.DebugLevelMexos, "not valid mexos env, fake app state ready")
				cd.appInstInfoState(key, edgeproto.AppState_AppStateReady)
				return
			}
			imagetype, err := convertImageType(int(appInst.ImageType))
			if err != nil {
				str := fmt.Sprintf("Invalid image type: %s", err)
				cd.appInstInfoError(key, str)
				log.DebugLog(log.DebugLevelMexos, "can't decode imagetype", "error", str, "imagetype", imagetype, "key", key)
				return
			}
			switch imagetype {
			case "docker":
				log.DebugLog(log.DebugLevelMexos, "docker image type, create app inst", "appinst", appInst, "clusterinst", clusterInst)
				err = MEXCreateAppInst(cd.CRMRootLB, &clusterInst, &appInst)
				if err != nil {
					errstr := fmt.Sprintf("Create failed: %s", err)
					cd.appInstInfoError(key, errstr)
					log.DebugLog(log.DebugLevelMexos, "can't create app inst", "error", errstr, "key", key, "imagetype", imagetype)
					return
				}
				log.DebugLog(log.DebugLevelMexos, "created docker app inst", "appisnt", appInst, "clusterinst", clusterInst)
			case "qcow2":
				log.DebugLog(log.DebugLevelMexos, "qcow2 image type, create app inst", "appinst", appInst, "clusterinst", clusterInst)
				err = MEXCreateAppInst(cd.CRMRootLB, &clusterInst, &appInst)
				if err != nil {
					errstr := fmt.Sprintf("Create failed: %s", err)
					cd.appInstInfoError(key, errstr)
					log.DebugLog(log.DebugLevelMexos, "can't create app inst", "error", errstr, "key", key, "imagetype", imagetype)
					return
				}
				log.DebugLog(log.DebugLevelMexos, "created qcow2 app inst", "appinst", appInst, "clusterinst", clusterInst)
			default:
				cd.appInstInfoError(key, "Unsupported image type")
				return
			}
			log.DebugLog(log.DebugLevelMexos, "update appinst state ready")
			cd.appInstInfoState(key, edgeproto.AppState_AppStateReady)
		}()
	} else {
		// appInst was deleted
		cd.appInstInfoError(key, "Delete not implemented yet")
		// TODO: implement me
	}
}

func convertImageType(imageType int) (string, error) {
	switch imageType {
	case int(edgeproto.ImageType_ImageTypeUnknown):
		return "", fmt.Errorf("unknown image type")
	case int(edgeproto.ImageType_ImageTypeDocker):
		return "docker", nil
	case int(edgeproto.ImageType_ImageTypeQCOW):
		return "qcow2", fmt.Errorf("unsupported qcow2") //XXX not yet
	}
	//XXX no kubernetes types, deployment, rc, rs, svc, po

	return "", fmt.Errorf("unknown")
}

func (cd *ControllerData) clusterInstInfoError(key *edgeproto.ClusterInstKey, err string) {
	info := edgeproto.ClusterInstInfo{}
	if !cd.ClusterInstInfoCache.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = edgeproto.ClusterState_ClusterStateErrors
	cd.ClusterInstInfoCache.Update(&info, 0)
}

func (cd *ControllerData) clusterInstInfoState(key *edgeproto.ClusterInstKey, state edgeproto.ClusterState) {
	info := edgeproto.ClusterInstInfo{}
	if !cd.ClusterInstInfoCache.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	cd.ClusterInstInfoCache.Update(&info, 0)
}

func (cd *ControllerData) appInstInfoError(key *edgeproto.AppInstKey, err string) {
	info := edgeproto.AppInstInfo{}
	if !cd.AppInstInfoCache.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = edgeproto.AppState_AppStateErrors
	cd.AppInstInfoCache.Update(&info, 0)
}

func (cd *ControllerData) appInstInfoState(key *edgeproto.AppInstKey, state edgeproto.AppState) {
	info := edgeproto.AppInstInfo{}
	if !cd.AppInstInfoCache.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	cd.AppInstInfoCache.Update(&info, 0)
}
