package crmutil

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

//ControllerData contains cache data for controller
type ControllerData struct {
	CRMRootLB            *MEXRootLB
	AppInstCache         edgeproto.AppInstCache
	CloudletCache        edgeproto.CloudletCache
	FlavorCache          edgeproto.FlavorCache
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
	edgeproto.InitClusterInstCache(&cd.ClusterInstCache)
	// set callbacks to trigger changes
	cd.ClusterInstCache.SetNotifyCb(cd.clusterInstChanged)
	cd.AppInstCache.SetNotifyCb(cd.appInstChanged)
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
}

// Note: these callback functions are called in the context of
// the notify receive thread. If the actions done here not quick,
// they should be done in a separate worker thread.

func (cd *ControllerData) clusterInstChanged(key *edgeproto.ClusterInstKey) {
	clusterInst := edgeproto.ClusterInst{}
	found := cd.ClusterInstCache.Get(key, &clusterInst)
	if found {
		// create or update k8s cluster on this cloudlet
		cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateBuilding)
		flavor := edgeproto.Flavor{}

		// XXX clusterInstCache has clusterInst but FlavorCache has clusterInst.Flavor.
		flavorFound := cd.FlavorCache.Get(&clusterInst.Flavor, &flavor)
		if !flavorFound {
			//XXX returning flavor not found error to InstInfoError?
			cd.clusterInstInfoError(key, fmt.Sprintf("Did not find flavor %s", clusterInst.Flavor.Name))
			return
		}

		go func() {
			var err error

			if IsValidMEXOSEnv {
				err = MEXClusterCreateClustInst(cd.CRMRootLB, &clusterInst)
			}
			if err != nil {
				cd.clusterInstInfoError(key, fmt.Sprintf("Create failed: %s", err))
				//XXX seems clusterInstInfoError is overloaded with status for flavor and clustinst.
			} else {
				cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateReady)
			}
			err = MEXAddFlavorClusterInst(&flavor) //Flavor is inside ClusterInst even though it comes from FlavorCache
			if err != nil {
				cd.clusterInstInfoError(key, fmt.Sprintf("Can't add flavor %s, %v", flavor.Key.Name, err))
			}
		}()
	} else {
		// clusterInst was deleted
		go func() {
			var err error
			if !IsValidMEXOSEnv {
				return
			}
			err = MEXClusterRemoveClustInst(cd.CRMRootLB, &clusterInst)
			if err != nil {
				str := fmt.Sprintf("Delete failed: %s", err)
				cd.clusterInstInfoError(key, str)
				return
			}
			cd.clusterInstInfoState(key, edgeproto.ClusterState_ClusterStateDeleted)
		}()
	}
}

func (cd *ControllerData) appInstChanged(key *edgeproto.AppInstKey) {
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
			imagetype, err := convertImageType(int(appInst.ImageType))
			if err != nil {
				str := fmt.Sprintf("Invalid image type: %s", err)
				cd.appInstInfoError(key, str)
				return
			}

			//XXX not sure what appInst.Flavor is

			switch imagetype {
			case "docker":
				//XXX ImagePath seems to be the only entry that can be used to specify docker image name.
				//XXX no registry & namspace specification.
				//XXX MappedPorts and MappedPath are strings but they can contain multiple entries.

				var err error
				if IsValidMEXOSEnv {
					err = MEXCreateAppInst(cd.CRMRootLB, &clusterInst, &appInst)
				}
				if err != nil {
					cd.appInstInfoError(key, fmt.Sprintf("Create failed: %s", err))
					return
				}
			default:
				cd.appInstInfoError(key, "Unsupported image type")
				return
			}

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
