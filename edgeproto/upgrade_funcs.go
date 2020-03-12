package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	context "golang.org/x/net/context"
)

func SetDefaultLoadBalancerMaxPortRange(ctx context.Context, objStore objstore.KVStore) error {
	log.DebugLog(log.DebugLevelUpgrade, "SetDefaultLoadBalancerMaxPortRange - default to 50")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Settings"))
	err := objStore.List(keystr, func(key, val []byte, rev int64) error {
		var settings Settings
		err2 := json.Unmarshal(val, &settings)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if settings.LoadBalancerMaxPortRange == 0 {
			log.DebugLog(log.DebugLevelUpgrade, "Defaulting LoadBalancerMaxPortRange")
			settings.LoadBalancerMaxPortRange = 50
			val, err2 = json.Marshal(settings)
			if err2 != nil {
				log.DebugLog(log.DebugLevelUpgrade, "Failed to marshal obj", "key", key, "obj", settings, "err", err2)
				return err2
			}
			objStore.Put(ctx, string(key), string(val))
		}
		return nil
	})
	return err
}

func SetDefaultMaxTrackedDmeClients(ctx context.Context, objStore objstore.KVStore) error {
	log.DebugLog(log.DebugLevelUpgrade, "SetDefaultMaxTrackedDmeClients - default to 100")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Settings"))
	err := objStore.List(keystr, func(key, val []byte, rev int64) error {
		var settings Settings
		err2 := json.Unmarshal(val, &settings)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if settings.MaxTrackedDmeClients == 0 {
			log.DebugLog(log.DebugLevelUpgrade, "Defaulting MaxTrackedDmeClients")
			settings.MaxTrackedDmeClients = 100
			val, err2 = json.Marshal(settings)
			if err2 != nil {
				log.DebugLog(log.DebugLevelUpgrade, "Failed to marshal obj", "key", key, "obj", settings, "err", err2)
				return err2
			}
			objStore.Put(ctx, string(key), string(val))
		}
		return nil
	})
	return err
}
