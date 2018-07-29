package crmutil

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type ControllerData struct {
	AppInstCache      edgeproto.AppInstCache
	CloudletCache     edgeproto.CloudletCache
	FlavorCache       edgeproto.FlavorCache
	ClusterInstCache  edgeproto.ClusterInstCache
	AppInstInfoCache  edgeproto.AppInstInfoCache
	CloudletInfoCache edgeproto.CloudletInfoCache
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData() *ControllerData {
	cd := &ControllerData{}
	edgeproto.InitAppInstCache(&cd.AppInstCache)
	edgeproto.InitCloudletCache(&cd.CloudletCache)
	edgeproto.InitAppInstInfoCache(&cd.AppInstInfoCache)
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
	// limits, err := oscli.GetLimits()
	// for _, limit := range limits {
	// // add info to cloudletInfo

	limits, err := oscli.GetLimits()
	if err != nil {
                //XXX No way to return error
                // And no log.ErrorLog()
		log.InfoLog("Error: can't get openstack limits", "error", err.Error())
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

	// for now, fake it.
	//info.OsMaxVcores = 50
	//info.OsMaxRam = 500
	//info.OsMaxVolGb = 5000
	// Is the cloudlet ready at this point?
	info.State = edgeproto.CloudletState_Ready
}

// Note: these callback functions are called in the context of
// the notify receive thread. If the actions done here not quick,
// they should be done in a separate worker thread.

func (cd *ControllerData) clusterInstChanged(key *edgeproto.ClusterInstKey) {
	//XXX validate CloudletKey
	clusterInst := edgeproto.ClusterInst{}
	found := cd.ClusterInstCache.Get(key, &clusterInst)
	if found {
		// create or update k8s cluster on this cloudlet
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&clusterInst.Flavor, &flavor)
		if !flavorFound {
			log.InfoLog("Error: did not find flavor for cluster",
				"cluster", clusterInst)
			//XXX no way to send error back to controller
			return
		}
		log.InfoLog("TODO: implement cluster create/update for",
			"cluster", clusterInst)

		if IsValidMEXOSEnv {
			go func() {
				err := CreateClusterFromClusterInstData(GetRootLBName(), &clusterInst)
				if err != nil {
					log.InfoLog("Error: failed to create cluster instance", "error", err, "cluster", clusterInst)
					return
				}
				//XXX no way to return results
			}()
		}
	} else {
		// clusterInst was deleted
		log.InfoLog("TODO: implement cluster delete for", "key", key)
		if IsValidMEXOSEnv {
			go func() {
				err := DeleteClusterByName(GetRootLBName(), key.ClusterKey.Name)
				if err != nil {
					log.InfoLog("Error: can't delete cluster %v, %v", key, err)
				}
				//XXX no way to return results
			}()
		}
	}
}

func (cd *ControllerData) appInstChanged(key *edgeproto.AppInstKey) {
	appInst := edgeproto.AppInst{}
	found := cd.AppInstCache.Get(key, &appInst)
	if found {
		// create or update appInst
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&appInst.Flavor, &flavor)
		if !flavorFound {
			log.InfoLog("Error: did not find flavor for appInst",
				"appInst", appInst)
			return
		}
		clusterInst := edgeproto.ClusterInst{}
		clusterInstFound := cd.ClusterInstCache.Get(&appInst.ClusterInstKey, &clusterInst)
		if !clusterInstFound {
			log.InfoLog("Error: did not find clusterInst for appInst",
				"appInst", appInst)
			return
		}
		log.InfoLog("TODO: implement appInst create/update for",
			"appInst", appInst)

		if IsValidMEXOSEnv {
			go func() {
				imagetype, err := convertImageType(int(appInst.ImageType))
				if err != nil {
					log.InfoLog("Error: invalid image type", "imagetype", appInst.ImageType, "error", err)
					return
				}

				//XXX no way to pass Kubernetes deployment, service, yaml, etc.
				//XXX not sure what appInst.Flavor is

				switch imagetype {
				case "docker":
					//Controller missing or not passing information:
					//XXX possibly incorrectly named ImagePath seems to be the only
					//  entry that can be used to specify docker image name.
					//XXX appData has AccessLayer but appInst does not.
					//   al, err := convertAccessLayer(appInst.AccessLayer)
					//XXX no registry specification.
					//XXX no namespace specification.
					//XXX MappedPorts and MappedPath are strings but they can contain
					//     multiple entries. Format is not clear.

					err := CreateDockerApp(GetRootLBName(),
						appInst.Key.AppKey.Name, clusterInst.Key.ClusterKey.Name, appInst.Flavor.Name,
						"docker.io", appInst.Uri, appInst.ImagePath, appInst.MappedPorts, appInst.MappedPath, "unknown")
					if err != nil {
						log.InfoLog("Error: can't create app", "error", err)
						return
					}
				default:
					log.InfoLog("Error: unknown image type")
				}

				//XXX no way to return results
			}()
		}
	} else {
		// appInst was deleted
		log.InfoLog("TODO: implement appInst delete for",
			"key", key)
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
