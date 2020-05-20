package edgeproto

import (
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/opentracing/opentracing-go"
	context "golang.org/x/net/context"
)

var testDataKeyPrefix = "_testdatakey"

// Prototype for the upgrade function - takes an objectstore and stm to ensure
// automicity of each upgrade function
type VersionUpgradeFunc func(context.Context, objstore.KVStore) error

// Helper function to run a single upgrade function across all the elements of a KVStore
// fn will be called for each of the entries, and therefore it's up to the
// fn implementation to filter based on the prefix
func RunSingleUpgrade(ctx context.Context, objStore objstore.KVStore, fn VersionUpgradeFunc) error {
	err := fn(ctx, objStore)
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
	span := log.StartSpan(log.DebugLevelInfo, "upgrade")
	span.SetTag("fromVersion", fromVersion)
	span.SetTag("verID", verID)
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	nextVer := verID + 1
	for {
		if fn, ok = VersionHash_UpgradeFuncs[nextVer]; !ok {
			break
		}
		name := VersionHash_UpgradeFuncNames[nextVer]

		uspan := log.StartSpan(log.DebugLevelInfo, name, opentracing.ChildOf(span.Context()))
		uctx := log.ContextWithSpan(context.Background(), uspan)
		if fn != nil {
			// Call the upgrade with an appropriate callback
			if err := RunSingleUpgrade(uctx, objStore, fn); err != nil {
				uspan.Finish()
				return fmt.Errorf("Failed to run %s: %v\n",
					name, err)
			}
			log.SpanLog(uctx, log.DebugLevelUpgrade, "Upgrade complete", "upgradeFunc", name)
		}
		// Write out the new version
		_, err := objStore.ApplySTM(uctx, func(stm concurrency.STM) error {
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
		uspan.Finish()
		if err != nil {
			return fmt.Errorf("Failed to update version for the db: %v\n", err)
		}
		nextVer++
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Upgrade done")
	return nil
}

func TestUpgradeExample(ctx context.Context, objStore objstore.KVStore) error {
	log.DebugLog(log.DebugLevelUpgrade, "TestUpgradeExample - reverse keys and values")
	// Define a prefix for a walk
	keystr := fmt.Sprintf("%s/", testDataKeyPrefix)
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		objStore.Delete(ctx, string(key))
		objStore.Put(ctx, string(val), string(key))
		return nil
	})
	return err
}
