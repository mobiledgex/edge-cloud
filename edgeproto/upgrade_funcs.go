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

var PlatosEnablingLayer = "PlatosEnablingLayer"

func PruneplatosPlatformDevices(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "PruneplatosPlatformDevices")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Device"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var device Device
		err2 := json.Unmarshal(val, &device)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if strings.Contains(strings.ToLower(device.Key.UniqueIdType), strings.ToLower(PlatosEnablingLayer)) {
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
			if _, perr := objStore.Put(ctx, string(key), string(val)); perr != nil {
				return perr
			}
		}
		return nil
	})
	return err
}

// Handles following upgrade:
// * Set default resource alert threshold for cloudlets
// * AddCloudletRefsClusterInstKeys adds ClusterInst keys to cloudlet refs
//   the assumption that Internal-only apps are trusted
func CloudletResourceUpgradeFunc(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "CloudletResourceUpgradeFunc")

	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Cloudlet"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var cloudlet Cloudlet
		err2 := json.Unmarshal(val, &cloudlet)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cloudlet", cloudlet)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "set defaultresourcealertthreshold for cloudlet", "cloudletkey", cloudlet.Key.String())
		if cloudlet.DefaultResourceAlertThreshold == 0 {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Setting default alert threshold to 80 for cloudlet", "cloudlet", cloudlet)
			cloudlet.DefaultResourceAlertThreshold = 80
			val, err2 = json.Marshal(cloudlet)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "cloudlet", cloudlet)
				return err2
			}
			if _, perr := objStore.Put(ctx, string(key), string(val)); perr != nil {
				return perr
			}
		}
		return nil
	})

	clusterMap := make(map[CloudletKey][]ClusterInstRefKey)
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("ClusterInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var clusterInst ClusterInst
		err2 := json.Unmarshal(val, &clusterInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "clusterinst", clusterInst)
			return err2
		}
		clKey := ClusterInstRefKey{}
		clKey.FromClusterInstKey(&clusterInst.Key)
		clusterMap[clusterInst.Key.CloudletKey] = append(clusterMap[clusterInst.Key.CloudletKey], clKey)
		return nil
	})
	if err != nil {
		return err
	}

	vmAppMap := make(map[AppKey]struct{})
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var app App
		err2 := json.Unmarshal(val, &app)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "app", app)
			return err2
		}
		if app.Deployment == "vm" {
			vmAppMap[app.Key] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return err
	}

	vmAppInstMap := make(map[CloudletKey][]AppInstRefKey)
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInst AppInst
		err2 := json.Unmarshal(val, &appInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appinst", appInst)
			return err2
		}
		if _, ok := vmAppMap[appInst.Key.AppKey]; !ok {
			return nil
		}
		aiKey := AppInstRefKey{}
		aiKey.FromAppInstKey(&appInst.Key)
		clKey := appInst.Key.ClusterInstKey.CloudletKey
		vmAppInstMap[clKey] = append(vmAppInstMap[clKey], aiKey)
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
		objChanged := false
		clusterKeys, ok := clusterMap[refs.Key]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "No clusters found for cloudlet", "cloudlet key", refs.Key)
		} else {
			refs.ClusterInsts = clusterKeys
			objChanged = true
		}
		vmAppInstKeys, ok := vmAppInstMap[refs.Key]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "No vm appinsts found for cloudlet", "cloudlet key", refs.Key)
		} else {
			refs.VmAppInsts = vmAppInstKeys
			objChanged = true
		}
		if objChanged {
			val, err2 = json.Marshal(refs)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "cloudletrefs", refs)
				return err2
			}
			if _, perr := objStore.Put(ctx, string(key), string(val)); perr != nil {
				return perr
			}
		}
		return nil
	})
	return err
}

// Handles initializing a new map on existing AppInstRefs objects.
func AppInstRefsDR(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "AppInstRefsDR")

	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInstRefs"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var refs AppInstRefs
		err2 := json.Unmarshal(val, &refs)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appinstrefs", refs)
			return err2
		}
		if refs.DeleteRequestedInsts != nil {
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "init DeletedRequestedInsts map on AppInstRefs", "refsKey", refs.Key.String())
		refs.DeleteRequestedInsts = make(map[string]uint32)
		val, err2 = json.Marshal(refs)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "AppInstRefs", refs)
			return err2
		}
		if _, perr := objStore.Put(ctx, string(key), string(val)); perr != nil {
			return perr
		}
		return nil
	})
	return err
}

// TrustPolicyException upgrade func
func TrustPolicyExceptionUpgradeFunc(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "TrustPolicyExceptionUpgradeFunc")

	type RemoteConnection struct {
		// tcp, udp, icmp
		Protocol string `json:"protocol,omitempty"`
		// port
		Port uint32 `json:"port,omitempty"`
		// remote IP
		RemoteIp string `json:"remote_ip,omitempty"`
	}

	type AppV0RemoteConn struct {
		RequiredOutboundConnections []*RemoteConnection `json:"required_outbound_connections,omitempty"`
	}

	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var app App
		err2 := json.Unmarshal(val, &app)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "app", app)
			return err2
		}
		var appV0 AppV0RemoteConn
		err2 = json.Unmarshal(val, &appV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal app old remote connection", "val", string(val), "err", err2, "app old", appV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "TrustPolicyExceptionUpgradeFunc found app", "required_outbound", appV0.RequiredOutboundConnections)
		if len(appV0.RequiredOutboundConnections) > 0 {
			newReqdConns := []SecurityRule{}
			for _, conn := range appV0.RequiredOutboundConnections {
				secRule := SecurityRule{
					Protocol:     conn.Protocol,
					PortRangeMin: conn.Port,
					PortRangeMax: conn.Port,
					RemoteCidr:   conn.RemoteIp + "/32",
				}
				newReqdConns = append(newReqdConns, secRule)
			}
			app.RequiredOutboundConnections = newReqdConns
			val, err2 = json.Marshal(app)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "app", app)
				return err2
			}
			if _, perr := objStore.Put(ctx, string(key), string(val)); perr != nil {
				return perr
			}
		}
		return nil
	})
	return err
}
