package edgeproto

import (
	"errors"
	fmt "fmt"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

var ErrUpgradeRemoveKey = errors.New("Remove object")
var ErrUpgradeReplaceKey = errors.New("Replace object")
var ErrUpgradeAddKey = errors.New("Add object")

// Helper function to run a single upgrade function across all the elements of a KVStore
// fn will be called for each of the entries, and therefore it's up to the
// fn implementation to filter based on the prefix
func runSingleUpgrade(objStore objstore.KVStore, fn VersionUpgradeFunc) error {
	var upgCount uint
	_, err := objStore.ApplySTM(func(stm concurrency.STM) error {
		// Start from the whole region
		keystr := fmt.Sprintf("%d/", objstore.GetRegion())
		err1 := objStore.List(keystr, func(key, val []byte, rev int64) error {

			// run the upgrade function and get an action
			newKey, newVal, err2 := fn(key, val)
			// actions based on the error code returned by the fn callback
			if err2 == nil {
				return nil
			}
			if strings.Contains(err2.Error(), ErrUpgradeRemoveKey.Error()) {
				stm.Del(string(key))
			} else if strings.Contains(err2.Error(), ErrUpgradeAddKey.Error()) {
				stm.Put(string(newKey), string(newVal))
			} else if strings.Contains(err2.Error(), ErrUpgradeReplaceKey.Error()) {
				stm.Del(string(key))
				stm.Put(string(newKey), string(newVal))
			} else {
				log.DebugLog(log.DebugLevelUpgrade, "Upgrade failed", "err", err2)
				return err2
			}
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
		return fmt.Errorf("Could not upgrade objects store entries, err: %v\n", err)
	}
	log.DebugLog(log.DebugLevelUpgrade, "Upgrade complete", "entries upgraded", upgCount)
	return nil
}

// This function walks all upgrade functions from the fromVersion to current
// and upgrades the KVStore using those functions one-by-one
func UpgradeToLatest(fromVersion string, objStore objstore.KVStore) error {
	var f VersionUpgradeFunc
	verID, ok := VersionHash_value["HASH_"+fromVersion]
	if !ok {
		return fmt.Errorf("fromVersion %s doesn't exist\n", fromVersion)
	}
	log.InfoLog("Upgrading", "fromVersion", fromVersion, "verID", verID)
	nextVer := verID + 1
	for {
		if f, ok = VersionHash_UpgradeFuncs[nextVer]; !ok {
			break
		}
		if f != nil {
			// Call the upgrade with an appropriate callback
			if err := runSingleUpgrade(objStore, f); err != nil {
				return fmt.Errorf("Failed to run %v: %v\n", f, err)
			}
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

func UpgradeMexSaltExample(key, val []byte) ([]byte, []byte, error) {
	log.DebugLog(log.DebugLevelUpgrade, "UpgradeMexSaltExample", "key", string(key))
	return key, val, nil
}

func UpgradeFuncExample(key, val []byte) ([]byte, []byte, error) {
	log.DebugLog(log.DebugLevelUpgrade, "UpgradeFuncExample", "key", string(key))
	return key, val, nil
}

func UpgradeFuncReplaceEverything(key, val []byte) ([]byte, []byte, error) {
	return key, val, ErrUpgradeReplaceKey
}
