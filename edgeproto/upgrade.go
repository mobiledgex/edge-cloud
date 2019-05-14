package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

var testDataKeyPrefix = "_testdatakey"

// Prototype for the upgrade function - takes an objectstore and stm to ensure
// automicity of each upgrade function
type VersionUpgradeFunc func(objstore.KVStore, concurrency.STM) error

// Helper function to run a single upgrade function across all the elements of a KVStore
// fn will be called for each of the entries, and therefore it's up to the
// fn implementation to filter based on the prefix
func RunSingleUpgrade(objStore objstore.KVStore, fn VersionUpgradeFunc) error {
	_, err := objStore.ApplySTM(func(stm concurrency.STM) error {
		return fn(objStore, stm)
	})
	if err != nil {
		return fmt.Errorf("Could not upgrade objects store entries, err: %v\n", err)
	}
	return nil
}

// This function walks all upgrade functions from the fromVersion to current
// and upgrades the KVStore using those functions one-by-one
func UpgradeToLatest(fromVersion string, objStore objstore.KVStore) error {
	var fn VersionUpgradeFunc
	verID, ok := VersionHash_value["HASH_"+fromVersion]
	if !ok {
		return fmt.Errorf("fromVersion %s doesn't exist\n", fromVersion)
	}
	log.InfoLog("Upgrading", "fromVersion", fromVersion, "verID", verID)
	nextVer := verID + 1
	for {
		if fn, ok = VersionHash_UpgradeFuncs[nextVer]; !ok {
			break
		}
		if fn != nil {
			// Call the upgrade with an appropriate callback
			if err := RunSingleUpgrade(objStore, fn); err != nil {
				return fmt.Errorf("Failed to run %s: %v\n",
					VersionHash_UpgradeFuncNames[nextVer], err)
			}
			log.DebugLog(log.DebugLevelUpgrade, "Upgrade complete", "upgradeFunc",
				VersionHash_UpgradeFuncNames[nextVer])
		}
		// Write out the new version
		_, err := objStore.ApplySTM(func(stm concurrency.STM) error {
			// Start from the whole region
			key := objstore.DbKeyPrefixString("Version")
			versionStr, ok := VersionHash_name[nextVer]
			if !ok {
				return fmt.Errorf("No hash string for version")
			}
			versionStr = versionStr[5:]
			stm.Put(string(key), versionStr)
			return nil
		})
		if err != nil {
			return fmt.Errorf("Failed to update version for the db: %v\n", err)
		}
		nextVer++
	}
	return nil
}

func TestUpgradeExample(objStore objstore.KVStore, stm concurrency.STM) error {
	log.DebugLog(log.DebugLevelUpgrade, "TestUpgradeExample - reverse keys and values")
	// Define a prefix for a walk
	keystr := fmt.Sprintf("%s/", testDataKeyPrefix)
	err := objStore.List(keystr, func(key, val []byte, rev int64) error {
		stm.Del(string(key))
		stm.Put(string(val), string(key))
		return nil
	})
	return err
}

// Below is an implementation of AddClusterInstKeyToAppInstKey
// This upgrade modifies AppInstKey to include ClusterInstKey rather than CloudletKey+Id
// This allows multiple instances of an app on the same cloudlet, but different cluster instances
func AddClusterInstKeyToAppInstKey(objStore objstore.KVStore, stm concurrency.STM) error {
	// Below are the data-structures for the older version of AppInstKey and AppInst
	type AppInstKeyV0_AddClusterInstKeyToAppInstKey struct {
		AppKey      AppKey      `protobuf:"bytes,1,opt,name=app_key,json=appKey" json:"app_key"`
		CloudletKey CloudletKey `protobuf:"bytes,2,opt,name=cloudlet_key,json=cloudletKey" json:"cloudlet_key"`
		Id          uint64      `protobuf:"fixed64,3,opt,name=id,proto3" json:"id,omitempty"`
	}
	type AppInstV0_AddClusterInstKeyToAppInstKey struct {
		Fields         []string                                   `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
		Key            AppInstKeyV0_AddClusterInstKeyToAppInstKey `protobuf:"bytes,2,opt,name=key" json:"key"`
		CloudletLoc    distributed_match_engine.Loc               `protobuf:"bytes,3,opt,name=cloudlet_loc,json=cloudletLoc" json:"cloudlet_loc"`
		Uri            string                                     `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
		ClusterInstKey ClusterInstKey                             `protobuf:"bytes,5,opt,name=cluster_inst_key,json=clusterInstKey" json:"cluster_inst_key"`
		Liveness       Liveness                                   `protobuf:"varint,6,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
		MappedPorts    []distributed_match_engine.AppPort         `protobuf:"bytes,9,rep,name=mapped_ports,json=mappedPorts" json:"mapped_ports"`
		Flavor         FlavorKey                                  `protobuf:"bytes,12,opt,name=flavor" json:"flavor"`
		State          TrackedState                               `protobuf:"varint,14,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
		Errors         []string                                   `protobuf:"bytes,15,rep,name=errors" json:"errors,omitempty"`
		CrmOverride    CRMOverride                                `protobuf:"varint,16,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
		CreatedAt      distributed_match_engine.Timestamp         `protobuf:"bytes,21,opt,name=created_at,json=createdAt" json:"created_at"`
	}

	var upgCount uint
	log.DebugLog(log.DebugLevelUpgrade, "AddClusterInstKeyToAppInstKey - change AppInstKey to contain ClusterInstKey instead of CloudletKey")
	// Define a prefix for a walk
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err := objStore.List(keystr, func(key, val []byte, rev int64) error {
		var appV0 AppInstV0_AddClusterInstKeyToAppInstKey
		err2 := json.Unmarshal(val, &appV0)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		log.DebugLog(log.DebugLevelUpgrade, "Upgrading AppInst from V0 to V1", "AppInstV0", appV0)
		appV1 := AppInst{}
		appV1.Fields = appV0.Fields
		appV1.Key.AppKey = appV0.Key.AppKey
		appV1.Key.ClusterInstKey = appV0.ClusterInstKey
		// There is a case in a yml file conversion that autocluster clusterInst doesn't specify cloudletkey
		if appV0.Key.CloudletKey != appV0.ClusterInstKey.CloudletKey {
			appV1.Key.ClusterInstKey.CloudletKey = appV0.Key.CloudletKey
		}
		appV1.CloudletLoc = appV0.CloudletLoc
		appV1.Uri = appV0.Uri
		appV1.Liveness = appV0.Liveness
		appV1.MappedPorts = appV0.MappedPorts
		appV1.Flavor = appV0.Flavor
		appV1.State = appV0.State
		appV1.Errors = appV0.Errors
		appV1.CrmOverride = appV0.CrmOverride
		appV1.CreatedAt = appV0.CreatedAt
		log.DebugLog(log.DebugLevelUpgrade, "Upgraded AppInstV1", "AppInstV1", appV1)
		stm.Del(string(key))
		keystr := objstore.DbKeyString("AppInst", appV1.GetKey())
		val, _ = json.Marshal(appV1)
		stm.Put(keystr, string(val))
		upgCount++
		return nil
	})
	log.DebugLog(log.DebugLevelUpgrade, "Upgrade object count", "Upgrade Count", upgCount)
	return err
}
