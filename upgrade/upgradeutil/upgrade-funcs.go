package upgrade

import (
	"errors"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var ErrUpgradNotSupported = errors.New("Unsupported upgrade path")

// Map of the names to upgrade functions
var UpgradeFuncs = map[string]interface{}{
	"UpgradeAppInstV0toV1":   UpgradeAppInstV0toV1,
	"DowngradeAppInstV1toV0": DowngradeAppInstV1toV0,
}

// Map of the names to yaml upgrade functions
var UpgradeYamlFuncs = map[string]interface{}{
	"UpgradeAppInstV0toV1": UpgradeYamlAppInstV0toV1,
}

func UpgradeAppInstV0toV1(appInst *AppInstV0) (*edgeproto.AppInst, error) {
	if appInst.Version != 0 {
		log.DebugLog(log.DebugLevelUpgrade, "Already at a non zero version", "key", appInst.Key,
			"version", appInst.Version)
		return nil, ErrUpgradNotSupported
	}
	newAppInst := edgeproto.AppInst{}
	newAppInst.Fields = appInst.Fields
	newAppInst.Key.AppKey = appInst.Key.AppKey
	newAppInst.Key.ClusterInstKey = appInst.ClusterInstKey
	// There is a case in a yml file conversion that autocluster clusterInst doesn't specify cloudletkey
	if appInst.Key.CloudletKey != appInst.ClusterInstKey.CloudletKey {
		newAppInst.Key.ClusterInstKey.CloudletKey = appInst.Key.CloudletKey
	}
	newAppInst.CloudletLoc = appInst.CloudletLoc
	newAppInst.Uri = appInst.Uri
	newAppInst.Liveness = appInst.Liveness
	newAppInst.MappedPorts = appInst.MappedPorts
	newAppInst.Flavor = appInst.Flavor
	newAppInst.State = appInst.State
	newAppInst.Errors = appInst.Errors
	newAppInst.CrmOverride = appInst.CrmOverride
	newAppInst.CreatedAt = appInst.CreatedAt
	newAppInst.Version = 1
	return &newAppInst, nil
}

func DowngradeAppInstV1toV0(appInst *edgeproto.AppInst) (*AppInstV0, error) {
	if appInst.Version != 1 {
		log.DebugLog(log.DebugLevelUpgrade, "Downgrading an incorrect version", "key", appInst.Key,
			"version", appInst.Version)
		return nil, ErrUpgradNotSupported
	}
	oldAppInst := AppInstV0{}
	oldAppInst.Fields = appInst.Fields
	oldAppInst.Key.AppKey = appInst.Key.AppKey
	oldAppInst.Key.CloudletKey = appInst.Key.ClusterInstKey.CloudletKey
	oldAppInst.Key.Id = 1 // Assume only a single appInst per cluster
	oldAppInst.CloudletLoc = appInst.CloudletLoc
	oldAppInst.Uri = appInst.Uri
	oldAppInst.ClusterInstKey = appInst.Key.ClusterInstKey
	oldAppInst.Liveness = appInst.Liveness
	oldAppInst.MappedPorts = appInst.MappedPorts
	oldAppInst.Flavor = appInst.Flavor
	oldAppInst.State = appInst.State
	oldAppInst.Errors = appInst.Errors
	oldAppInst.CrmOverride = appInst.CrmOverride
	oldAppInst.CreatedAt = appInst.CreatedAt
	oldAppInst.Version = 0
	return &oldAppInst, nil
}

func UpgradeYamlAppInstV0toV1(yaml string) (*edgeproto.ApplicationData, error) {
	var appData ApplicationData_AppInstV0
	var appDataNew *edgeproto.ApplicationData = &edgeproto.ApplicationData{}
	// Read in the yaml file
	if err := util.ReadYamlFile(yaml, &appData, "", false); err != nil {
		return nil, fmt.Errorf("Could not parse the input yaml %s, err: %v\n", yaml, err)
	}
	appDataNew.Operators = appData.Operators
	appDataNew.Cloudlets = appData.Cloudlets
	appDataNew.Flavors = appData.Flavors
	appDataNew.ClusterFlavors = appData.ClusterFlavors
	appDataNew.Clusters = appData.Clusters
	appDataNew.ClusterInsts = appData.ClusterInsts
	appDataNew.Developers = appData.Developers
	appDataNew.Applications = appData.Applications

	// Convert AppInstaces V0 to V1
	for _, app := range appData.AppInstances {
		newApp, err := UpgradeAppInstV0toV1(&app)
		if err != nil {
			return nil, fmt.Errorf("Failed to upgrade appInst %v, err: %v\n", app, err)
		}
		log.DebugLog(log.DebugLevelUpgrade, "Upgraded app to V1", "old", app, "new", newApp)
		appDataNew.AppInstances = append(appDataNew.AppInstances, *newApp)
	}
	appDataNew.CloudletInfos = appData.CloudletInfos
	appDataNew.AppInstInfos = appData.AppInstInfos
	appDataNew.ClusterInstInfos = appData.ClusterInstInfos
	appDataNew.Nodes = appData.Nodes
	return appDataNew, nil
}

func UpgradeEtcdAppInstV0toV1(etcdUrls string) error {
	//TODO
	return nil
}
