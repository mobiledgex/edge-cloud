package edgeproto

import (
	"encoding/json"
	fmt "fmt"
	strings "strings"

	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	context "golang.org/x/net/context"
)

func CheckForHttpPorts(ctx context.Context, objStore objstore.KVStore) error {
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	cbErrs := make([]error, 0)
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInst AppInst
		err2 := json.Unmarshal(val, &appInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appInst", appInst)
			cbErrs = append(cbErrs, err2)
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Checking AppInst for invalid legacy http ports", "appInst", appInst)
		for _, mappedPort := range appInst.MappedPorts {
			match := false
			for _, protoVal := range distributed_match_engine.LProto_value {
				if int32(mappedPort.Proto) == protoVal {
					match = true
				}
			}
			if !match {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Invalid protocol found", "appInst", appInst, "AppPort", mappedPort)
				err3 := fmt.Errorf("Invalid protocol found: %d", mappedPort.Proto)
				cbErrs = append(cbErrs, err3)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(cbErrs) == 0 {
		return nil
	}
	return fmt.Errorf("Errors: %v", cbErrs)
}

var samsungEnablingLayer = "SamsungEnablingLayer"

func PruneSamsungPlatformDevices(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "PruneSamsungPlatformDevices")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Device"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var device Device
		err2 := json.Unmarshal(val, &device)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if strings.Contains(strings.ToLower(device.Key.UniqueIdType), strings.ToLower(samsungEnablingLayer)) {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Prune a MEL device", "device", device)
			_, err := objStore.Delete(ctx, string(key))
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to delete platform device key", "key", string(key))
			}
		}
		return nil
	})
	return err
}

// SetTrusted sets the Trusted bit to true for all InternalPorts apps on
// the assumption that Internal-only apps are trusted
func SetTrusted(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "SetTrusted")

	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var app App
		err2 := json.Unmarshal(val, &app)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "app", app)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "SetTrusted found app", "appkey", app.Key.String(), "InternalPorts", app.InternalPorts)
		if app.InternalPorts {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Setting PrivacyComplaint to true for internal ports app", "app", app)
			app.Trusted = true
			val, err2 = json.Marshal(app)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "app", app)
				return err2
			}
			objStore.Put(ctx, string(key), string(val))
		}
		return nil
	})
	return err
}

// AddCloudletRefsClusterInstKeys adds ClusterInst keys to cloudlet refs
// the assumption that Internal-only apps are trusted
func AddCloudletRefsClusterInstKeys(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "AddCloudletRefsClusterInstKeys")

	clusterMap := make(map[CloudletKey][]ClusterInstKey)
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("ClusterInst"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var clusterInst ClusterInst
		err2 := json.Unmarshal(val, &clusterInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "clusterinst", clusterInst)
			return err2
		}
		clKey := clusterInst.Key
		clKey.CloudletKey = CloudletKey{}
		clusterMap[clusterInst.Key.CloudletKey] = append(clusterMap[clusterInst.Key.CloudletKey], clKey)
		return nil
	})
	if err != nil {
		return err
	}

	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("CloudletRefs"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var refs CloudletRefs
		err2 := json.Unmarshal(val, &refs)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cloudletrefs", refs)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "AddCloudletRefsClusterInstKeys found obj", "cloudletKey", refs.Key.String())
		clusterKeys, ok := clusterMap[refs.Key]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "No clusters found for cloudlet", "cloudlet key", refs.Key)
			return nil
		}
		refs.ClusterInsts = clusterKeys
		val, err2 = json.Marshal(refs)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "cloudletrefs", refs)
			return err2
		}
		objStore.Put(ctx, string(key), string(val))
		return nil
	})
	return err
}
