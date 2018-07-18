package crmutil

import (
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

	// for now, fake it.
	info.OsMaxVcores = 50
	info.OsMaxRam = 500
	info.OsMaxVolGb = 5000
	// Is the cloudlet ready at this point?
	info.State = edgeproto.CloudletState_Ready
}

// Note: these callback functions are called in the context of
// the notify receive thread. If the actions done here not quick,
// they should be done in a separate worker thread.

func (cd *ControllerData) clusterInstChanged(key *edgeproto.ClusterInstKey) {
	clusterInst := edgeproto.ClusterInst{}
	found := cd.ClusterInstCache.Get(key, &clusterInst)
	if found {
		// create or update k8s cluster on this cloudlet
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&clusterInst.Flavor, &flavor)
		if !flavorFound {
			log.InfoLog("Error: did not find flavor for cluster",
				"cluster", clusterInst)
			return
		}
		log.InfoLog("TODO: implement cluster create/update for",
			"cluster", clusterInst)
	} else {
		// clusterInst was deleted
		log.InfoLog("TODO: implement cluster delete for", "key", key)
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
	} else {
		// appInst was deleted
		log.InfoLog("TODO: implement appInst delete for",
			"key", key)
	}
}
