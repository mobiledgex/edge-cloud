package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	etcd_client "github.com/mobiledgex/edge-cloud/controller/etcd_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var ErrUpgradNotSupported = errors.New("Unsupported upgrade path")

// Map of the names to upgrade functions
var UpgradeFuncs = map[string]interface{}{
	"UpgradeAppInstV0toV1": UpgradeEtcdAppInstV0toV1,
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

func UpgradeYamlAppInstV0toV1(yaml string) (*edgeproto.ApplicationData, uint, error) {
	var appData ApplicationData_AppInstV0
	var appDataNew *edgeproto.ApplicationData = &edgeproto.ApplicationData{}
	var upgCount uint
	// Read in the yaml file
	if err := util.ReadYamlFile(yaml, &appData, "", false); err != nil {
		return nil, 0, fmt.Errorf("Could not parse the input yaml %s, err: %v\n", yaml, err)
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
			return nil, 0, fmt.Errorf("Failed to upgrade appInst %v, err: %v\n", app, err)
		}
		log.DebugLog(log.DebugLevelUpgrade, "Upgraded app to V1", "old", app, "new", newApp)
		appDataNew.AppInstances = append(appDataNew.AppInstances, *newApp)
		upgCount++
	}
	appDataNew.CloudletInfos = appData.CloudletInfos
	appDataNew.AppInstInfos = appData.AppInstInfos
	appDataNew.ClusterInstInfos = appData.ClusterInstInfos
	appDataNew.Nodes = appData.Nodes
	return appDataNew, upgCount, nil
}

func UpgradeEtcdAppInstV0toV1(etcdUrls string, region uint) (uint, error) {
	var upgCount uint
	objstore.InitRegion(uint32(region))
	objStore, err := etcd_client.GetEtcdClientBasic(etcdUrls)
	if err != nil {
		return upgCount, fmt.Errorf("Cannot init etcd at url: %s, err: %v\n", etcdUrls, err)
	}
	_, err = objStore.ApplySTM(func(stm concurrency.STM) error {
		keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
		err1 := objStore.List(keystr, func(key, val []byte, rev int64) error {
			var appV0 AppInstV0
			err2 := json.Unmarshal(val, &appV0)
			if err2 != nil {
				log.DebugLog(log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
				return err2
			}
			log.DebugLog(log.DebugLevelUpgrade, "Upgrading AppInst from V0 to V1", "AppInstV0", appV0)
			appV1, err2 := UpgradeAppInstV0toV1(&appV0)
			if err2 != nil {
				log.DebugLog(log.DebugLevelUpgrade, "Upgrade failed", "err", err2)
				return err2
			}
			log.DebugLog(log.DebugLevelUpgrade, "Upgraded AppInstV1", "AppInstV1", appV1)
			stm.Del(string(key))
			keystr := objstore.DbKeyString("AppInst", appV1.GetKey())
			val, _ = json.Marshal(appV1)
			stm.Put(keystr, string(val))
			upgCount++
			return nil
		})
		if err1 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Failed to get keys for objstore", "err", err1)
			return err1
		}
		return nil
	})
	if err != nil {
		return upgCount, fmt.Errorf("Could not upgrade objects store entries, err: %v\n", err)
	}
	return upgCount, nil
}
