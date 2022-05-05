// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	fmt "fmt"
	"path/filepath"
	"regexp"
	"sort"
	strings "strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/edgexr/edge-cloud/cloudcommon"
	distributed_match_engine "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/objstore"
	"github.com/edgexr/edge-cloud/util"
	context "golang.org/x/net/context"
)

func CheckForHttpPorts(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	cbErrs := make([]error, 0)
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInst edgeproto.AppInst
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

func PruneplatosPlatformDevices(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "PruneplatosPlatformDevices")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Device"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var device edgeproto.Device
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
func SetTrusted(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "SetTrusted")

	appKeys, err := getDbObjectKeys(objStore, "App")
	if err != nil {
		return err
	}
	for appKey, _ := range appKeys {
		_, err := objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			appStr := stm.Get(appKey)
			if appStr == "" {
				// deleted in the meantime
				return nil
			}
			var app edgeproto.App
			err2 := json.Unmarshal([]byte(appStr), &app)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", appStr, "err", err2, "app", app)
				return err2
			}
			log.SpanLog(ctx, log.DebugLevelUpgrade, "SetTrusted found app", "appkey", app.Key.String(), "InternalPorts", app.InternalPorts)
			if app.InternalPorts {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Setting PrivacyComplaint to true for internal ports app", "app", app)
				app.Trusted = true
				val, err2 := json.Marshal(app)
				if err2 != nil {
					log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "app", app)
					return err2
				}
				stm.Put(appKey, string(val))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Handles following upgrade:
// * Set default resource alert threshold for cloudlets
// * AddCloudletRefsClusterInstKeys adds ClusterInst keys to cloudlet refs
//   the assumption that Internal-only apps are trusted
func CloudletResourceUpgradeFunc(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "CloudletResourceUpgradeFunc")

	cloudletKeys, err := getDbObjectKeys(objStore, "Cloudlet")
	if err != nil {
		return err
	}
	for key, _ := range cloudletKeys {
		_, err := objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			cloudletStr := stm.Get(key)
			if cloudletStr == "" {
				return nil
			}
			var cloudlet edgeproto.Cloudlet
			err2 := json.Unmarshal([]byte(cloudletStr), &cloudlet)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", cloudletStr, "err", err2, "cloudlet", cloudlet)
				return err2
			}
			log.SpanLog(ctx, log.DebugLevelUpgrade, "set defaultresourcealertthreshold for cloudlet", "cloudletkey", cloudlet.Key.String())
			if cloudlet.DefaultResourceAlertThreshold == 0 {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Setting default alert threshold to 80 for cloudlet", "cloudlet", cloudlet)
				cloudlet.DefaultResourceAlertThreshold = 80
				val, err2 := json.Marshal(cloudlet)
				if err2 != nil {
					log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "cloudlet", cloudlet)
					return err2
				}
				stm.Put(key, string(val))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	clusterMap := make(map[edgeproto.CloudletKey][]edgeproto.ClusterInstRefKey)
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("ClusterInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var clusterInst edgeproto.ClusterInst
		err2 := json.Unmarshal(val, &clusterInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "clusterinst", clusterInst)
			return err2
		}
		clKey := edgeproto.ClusterInstRefKey{}
		clKey.FromClusterInstKey(&clusterInst.Key)
		clusterMap[clusterInst.Key.CloudletKey] = append(clusterMap[clusterInst.Key.CloudletKey], clKey)
		return nil
	})
	if err != nil {
		return err
	}

	vmAppMap := make(map[edgeproto.AppKey]struct{})
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var app edgeproto.App
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

	vmAppInstMap := make(map[edgeproto.CloudletKey][]edgeproto.AppInstRefKey)
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInst edgeproto.AppInst
		err2 := json.Unmarshal(val, &appInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appinst", appInst)
			return err2
		}
		if _, ok := vmAppMap[appInst.Key.AppKey]; !ok {
			return nil
		}
		aiKey := edgeproto.AppInstRefKey{}
		aiKey.FromAppInstKey(&appInst.Key)
		clKey := appInst.Key.ClusterInstKey.CloudletKey
		vmAppInstMap[clKey] = append(vmAppInstMap[clKey], aiKey)
		return nil
	})
	if err != nil {
		return err
	}

	cloudletRefsKeys, err := getDbObjectKeys(objStore, "CloudletRefs")
	if err != nil {
		return err
	}
	for refsKey, _ := range cloudletRefsKeys {
		_, err := objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			refsStr := stm.Get(refsKey)
			if refsStr == "" {
				return nil
			}
			var refs edgeproto.CloudletRefs
			err2 := json.Unmarshal([]byte(refsStr), &refs)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", refsStr, "err", err2, "cloudletrefs", refs)
				return err2
			}
			log.SpanLog(ctx, log.DebugLevelUpgrade, "AddCloudletRefsClusterInstKeys found obj", "cloudletKey", refs.Key.String())
			curClusterInsts := make(map[edgeproto.ClusterInstRefKey]struct{})
			for _, key := range refs.ClusterInsts {
				curClusterInsts[key] = struct{}{}
			}
			curVmAppInsts := make(map[edgeproto.AppInstRefKey]struct{})
			for _, key := range refs.VmAppInsts {
				curVmAppInsts[key] = struct{}{}
			}
			objChanged := false
			clusterKeys, ok := clusterMap[refs.Key]
			if !ok {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "No clusters found for cloudlet", "cloudlet key", refs.Key)
			} else {
				for _, key := range clusterKeys {
					if _, found := curClusterInsts[key]; found {
						continue
					}
					// make sure it still exists
					lookupKey := &edgeproto.ClusterInstKey{}
					lookupKey.FromClusterInstRefKey(&key, &refs.Key)
					str := stm.Get(objstore.DbKeyString("ClusterInst", lookupKey))
					if str == "" {
						continue
					}
					refs.ClusterInsts = append(refs.ClusterInsts, key)
					objChanged = true
				}
			}
			vmAppInstKeys, ok := vmAppInstMap[refs.Key]
			if !ok {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "No vm appinsts found for cloudlet", "cloudlet key", refs.Key)
			} else {
				for _, key := range vmAppInstKeys {
					if _, found := curVmAppInsts[key]; found {
						continue
					}
					// make sure it still exists
					lookupKey := &edgeproto.AppInstKey{}
					lookupKey.FromAppInstRefKey(&key, &refs.Key)
					str := stm.Get(objstore.DbKeyString("AppInst", lookupKey))
					if str == "" {
						continue
					}
					refs.VmAppInsts = append(refs.VmAppInsts, key)
					objChanged = true
				}
			}
			if !objChanged {
				return nil
			}
			val, err2 := json.Marshal(refs)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "cloudletrefs", refs)
				return err2
			}
			stm.Put(refsKey, string(val))
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

// Handles initializing a new map on existing AppInstRefs objects.
func AppInstRefsDR(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "AppInstRefsDR")

	refsKeys := make(map[string]struct{})
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInstRefs"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		refsKeys[string(key)] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	for key, _ := range refsKeys {
		_, err := objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// get refs
			refsStr := stm.Get(key)
			if refsStr == "" {
				// deleted in the meantime
				return nil
			}
			var refs edgeproto.AppInstRefs
			err2 := json.Unmarshal([]byte(refsStr), &refs)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", refsStr, "err", err2, "appinstrefs", refs)
				return err2
			}
			if refs.DeleteRequestedInsts != nil {
				return nil
			}
			log.SpanLog(ctx, log.DebugLevelUpgrade, "init DeletedRequestedInsts map on AppInstRefs", "refsKey", refs.Key.String())
			refs.DeleteRequestedInsts = make(map[string]uint32)
			val, err2 := json.Marshal(refs)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "AppInstRefs", refs)
				return err2
			}
			stm.Put(key, string(val))
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// TrustPolicyException upgrade func
func TrustPolicyExceptionUpgradeFunc(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
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

	appKeys, err := getDbObjectKeys(objStore, "App")
	if err != nil {
		return err
	}

	for key, _ := range appKeys {
		_, err := objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			appStr := stm.Get(key)
			if appStr == "" {
				// deleted in the meantime
				return nil
			}
			var app edgeproto.App
			err2 := json.Unmarshal([]byte(appStr), &app)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", appStr, "err", err2, "app", app)
				return err2
			}
			var appV0 AppV0RemoteConn
			err2 = json.Unmarshal([]byte(appStr), &appV0)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal app old remote connection", "val", appStr, "err", err2, "app old", appV0)
				return err2
			}
			log.SpanLog(ctx, log.DebugLevelUpgrade, "TrustPolicyExceptionUpgradeFunc found app", "required_outbound", appV0.RequiredOutboundConnections)
			if len(appV0.RequiredOutboundConnections) > 0 && appV0.RequiredOutboundConnections[0].Port != 0 {
				newReqdConns := []edgeproto.SecurityRule{}
				for _, conn := range appV0.RequiredOutboundConnections {
					secRule := edgeproto.SecurityRule{
						Protocol:     conn.Protocol,
						PortRangeMin: conn.Port,
						PortRangeMax: conn.Port,
						RemoteCidr:   conn.RemoteIp + "/32",
					}
					newReqdConns = append(newReqdConns, secRule)
				}
				app.RequiredOutboundConnections = newReqdConns
				val, err2 := json.Marshal(app)
				if err2 != nil {
					log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "app", app)
					return err2
				}
				stm.Put(key, string(val))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Initiate and back-populate cluster refs objects for existing AppInsts
func AddClusterRefs(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "ClusterRefs")

	// Get all AppInsts
	appInstKeys, err := getDbObjectKeys(objStore, "AppInst")
	if err != nil {
		return err
	}

	// Use an STM to update refs to avoid conflicts with multiple
	// controllers and to keep it idempotent
	for aiKey, _ := range appInstKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// get AppInst
			appInstStr := stm.Get(aiKey)
			if appInstStr == "" {
				// must have been deleted in the meantime
				return nil
			}
			appInst := edgeproto.AppInst{}
			err := json.Unmarshal([]byte(appInstStr), &appInst)
			if err != nil {
				return fmt.Errorf("Unmarshal AppInst %s failed: %s", aiKey, err)
			}
			// get App, so we can skip VM AppInsts
			appKey := objstore.DbKeyString("App", &appInst.Key.AppKey)
			appStr := stm.Get(appKey)
			if appStr == "" {
				return fmt.Errorf("No App found for AppInst %s", aiKey)
			}
			app := edgeproto.App{}
			err = json.Unmarshal([]byte(appStr), &app)
			if err != nil {
				return fmt.Errorf("Unmarshal App %s failed: %s", appKey, err)
			}
			if app.Deployment == "vm" {
				// no ClusterInst so no refs
				return nil
			}
			// add AppInst ref to ClusterRefs
			clusterInstKey := appInst.ClusterInstKey()
			refsKey := objstore.DbKeyString("ClusterRefs", clusterInstKey)
			refsStr := stm.Get(refsKey)
			refs := edgeproto.ClusterRefs{}
			if refsStr != "" {
				err = json.Unmarshal([]byte(refsStr), &refs)
				if err != nil {
					return fmt.Errorf("Unmarshal ClusterRefs %s failed: %s", refsKey, err)
				}
			} else {
				refs.Key = *clusterInstKey
			}
			appRefKey := appInst.Key.ClusterRefsAppInstKey()
			found := false
			for _, k := range refs.Apps {
				if k.Matches(appRefKey) {
					found = true
					break
				}
			}
			if found {
				// already there, no change needed
				return nil
			}
			refs.Apps = append(refs.Apps, *appRefKey)
			// sort for determinism for unit-test comparison
			sort.Slice(refs.Apps, func(i, j int) bool {
				return refs.Apps[i].GetKeyString() < refs.Apps[j].GetKeyString()
			})
			refsData, err := json.Marshal(refs)
			if err != nil {
				return fmt.Errorf("Marshal ClusterRefs %s failed: %s", refsKey, err)
			}
			stm.Put(refsKey, string(refsData))
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// This is the old cloudcommon.GetAppFQN() function which was used by
// the vmlayer to generate the name for heat stacks, etc.
func oldGetAppFQN(key *edgeproto.AppKey) string {
	app := util.DNSSanitize(key.Name)
	dev := util.DNSSanitize(key.Organization)
	ver := util.DNSSanitize(key.Version)
	return fmt.Sprintf("%s%s%s", dev, app, ver)
}

func AddAppInstUniqueId(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	// Get all AppInsts
	appInstKeys, err := getDbObjectKeys(objStore, "AppInst")
	if err != nil {
		return err
	}

	// Use an STM to avoid conflicts with multiple
	// controllers and to keep it idempotent
	for aiKey, _ := range appInstKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// get AppInst
			appInstStr := stm.Get(aiKey)
			if appInstStr == "" {
				// must have been deleted in the meantime
				return nil
			}
			appInst := edgeproto.AppInst{}
			err := json.Unmarshal([]byte(appInstStr), &appInst)
			if err != nil {
				return fmt.Errorf("Unmarshal AppInst %s failed: %s", aiKey, err)
			}
			if appInst.UniqueId != "" {
				// already set
				return nil
			}
			appInst.UniqueId = oldGetAppFQN(&appInst.Key.AppKey)
			aiData, err := json.Marshal(appInst)
			if err != nil {
				return fmt.Errorf("Marshal AppInst %s failed: %s", appInst.Key.GetKeyString(), err)
			}
			stm.Put(aiKey, string(aiData))
			// store unique id - these may conflict but
			// there's not much we can do about the old ones.
			idKey := edgeproto.AppInstIdDbKey(appInst.UniqueId)
			stm.Put(idKey, appInst.UniqueId)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getDbObjectKeys(objStore objstore.KVStore, dbPrefix string) (map[string]struct{}, error) {
	keys := make(map[string]struct{})
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString(dbPrefix))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		keys[string(key)] = struct{}{}
		return nil
	})
	return keys, err
}

// deprecated fqdn functions

const oldRootLBHostname = "shared"

// GetCloudletBaseFQDN gets the base 3-label FQDN for the cloudlet.
// For TLS termination, we should only require a single cert that
// wildcard matches *.<cloudlet-base-fqdn>, as all other DNS names
// should only add one more label on top of the base fqdn.
func oldGetCloudletBaseFQDN(key *edgeproto.CloudletKey, domain string) string {
	loc := util.DNSSanitize(key.Name)
	oper := util.DNSSanitize(key.Organization)
	return fmt.Sprintf("%s.%s.%s", loc, oper, domain)
}

// GetRootLBFQDN gets the global Load Balancer's Fully Qualified Domain Name
// for apps using "shared" IP access.
func oldGetRootLBFQDN(key *edgeproto.CloudletKey, domain string) string {
	return fmt.Sprintf("%s.%s", oldRootLBHostname, oldGetCloudletBaseFQDN(key, domain))
}

// GetDedicatedLBFQDN gets the cluster-specific Load Balancer's Fully Qualified Domain Name
// for clusters using "dedicated" IP access.
func oldGetDedicatedLBFQDN(cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey, domain string) string {
	clust := util.DNSSanitize(clusterKey.Name)
	return fmt.Sprintf("%s.%s", clust, oldGetCloudletBaseFQDN(cloudletKey, domain))
}

// GetAppFQDN gets the app-specific Load Balancer's Fully Qualified Domain Name
// for apps using "dedicated" IP access. This will not allow TLS, but will
// ensure uniqueness when an IP is assigned per k8s-service per AppInst per cluster.
func oldGetAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey, clusterKey *edgeproto.ClusterKey, domain string) string {
	clusterBase := oldGetDedicatedLBFQDN(cloudletKey, clusterKey, domain)
	appFQN := oldGetAppFQN(&key.AppKey)
	return fmt.Sprintf("%s.%s", appFQN, clusterBase)
}

// GetVMAppFQDN gets the app-specific Fully Qualified Domain Name
// for VM based apps
func oldGetVMAppFQDN(key *edgeproto.AppInstKey, cloudletKey *edgeproto.CloudletKey, domain string) string {
	appFQN := oldGetAppFQN(&key.AppKey)
	return fmt.Sprintf("%s.%s", appFQN, oldGetCloudletBaseFQDN(cloudletKey, domain))
}

func AddDnsLabels(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	// Process cloudlets first
	cloudletKeys, err := getDbObjectKeys(objStore, "Cloudlet")
	if err != nil {
		return err
	}
	for key, _ := range cloudletKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// get cloudlet
			cloudletStr := stm.Get(key)
			if cloudletStr == "" {
				return nil // was deleted
			}
			cloudlet := edgeproto.Cloudlet{}
			err := json.Unmarshal([]byte(cloudletStr), &cloudlet)
			if err != nil {
				return fmt.Errorf("Unmarshal Cloudlet %s failed: %s", key, err)
			}
			if cloudlet.DnsLabel != "" {
				return nil // already done
			}
			if err := allApis.cloudletApi.setDnsLabel(stm, &cloudlet); err != nil {
				return fmt.Errorf("Set dns label for cloudlet %s failed, %s", key, err)
			}
			if cloudlet.RootLbFqdn == "" {
				// set old version of rootLBFQDN
				cloudlet.RootLbFqdn = oldGetRootLBFQDN(&cloudlet.Key, *appDNSRoot)
			}
			allApis.cloudletApi.store.STMPut(stm, &cloudlet)
			allApis.cloudletApi.dnsLabelStore.STMPut(stm, cloudlet.DnsLabel)
			return nil
		})
		if err != nil {
			return err
		}
	}

	clusterInstKeys, err := getDbObjectKeys(objStore, "ClusterInst")
	if err != nil {
		return err
	}
	for key, _ := range clusterInstKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			clusterInstStr := stm.Get(key)
			if clusterInstStr == "" {
				return nil // was deleted
			}
			clusterInst := edgeproto.ClusterInst{}
			err := json.Unmarshal([]byte(clusterInstStr), &clusterInst)
			if err != nil {
				return fmt.Errorf("Unmarshal ClusterInst %s failed: %s", key, err)
			}
			if clusterInst.DnsLabel != "" {
				return nil // already done
			}
			if err := allApis.clusterInstApi.setDnsLabel(stm, &clusterInst); err != nil {
				return fmt.Errorf("Set dns label for ClusterInst %s failed, %s", key, err)
			}
			if clusterInst.Fqdn == "" {
				clusterInst.Fqdn = oldGetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey, *appDNSRoot)
			}

			allApis.clusterInstApi.store.STMPut(stm, &clusterInst)
			allApis.clusterInstApi.dnsLabelStore.STMPut(stm, &clusterInst.Key.CloudletKey, clusterInst.DnsLabel)
			return nil
		})
		if err != nil {
			return err
		}
	}

	appInstKeys, err := getDbObjectKeys(objStore, "AppInst")
	if err != nil {
		return err
	}
	for key, _ := range appInstKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			appInstStr := stm.Get(key)
			if appInstStr == "" {
				return nil // was deleted
			}
			appInst := edgeproto.AppInst{}
			err := json.Unmarshal([]byte(appInstStr), &appInst)
			if err != nil {
				return fmt.Errorf("Unmarshal ClusterInst %s failed: %s", key, err)
			}
			if appInst.DnsLabel != "" {
				return nil // already done
			}
			if err := allApis.appInstApi.setDnsLabel(stm, &appInst); err != nil {
				return fmt.Errorf("Set dns label for AppInst %s failed, %s", key, err)
			}
			allApis.appInstApi.store.STMPut(stm, &appInst)
			allApis.appInstApi.dnsLabelStore.STMPut(stm, &appInst.Key.ClusterInstKey.CloudletKey, appInst.DnsLabel)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type OldCloudletPool struct {
	Key       edgeproto.CloudletPoolKey
	Cloudlets []string
}

func AddCloudletKeyToCloudletPool(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	cloudletPoolKeys, err := getDbObjectKeys(objStore, "CloudletPool")
	if err != nil {
		return err
	}
	for key, _ := range cloudletPoolKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// Get cloudlet pool
			cloudletPoolStr := stm.Get(key)
			if cloudletPoolStr == "" {
				return nil // was deleted
			}

			// Check if cloudlet pool is already upgraded
			newCloudletPool := edgeproto.CloudletPool{}
			err = json.Unmarshal([]byte(cloudletPoolStr), &newCloudletPool)
			if err == nil {
				return nil // already upgraded
			}

			// Get data in old cloudlet pool format
			oldCloudletPool := OldCloudletPool{}
			err := json.Unmarshal([]byte(cloudletPoolStr), &oldCloudletPool)
			if err != nil {
				return fmt.Errorf("Unmarshal CloudletPool %s failed: %s", key, err)
			}
			if len(oldCloudletPool.Cloudlets) == 0 {
				return nil // nothing to upgrade
			}

			// Fix cloudlet pool object upgrade using regex
			var re = regexp.MustCompile(`"cloudlets":\[.*?\]`)
			emptyCloudletsPoolStr := re.ReplaceAllString(cloudletPoolStr, `"cloudlets":[]`)
			newCloudletPool = edgeproto.CloudletPool{}
			err = json.Unmarshal([]byte(emptyCloudletsPoolStr), &newCloudletPool)
			if err != nil {
				return fmt.Errorf("Unmarshal CloudletPool %s failed: %s", key, err)
			}
			newCloudletPool.Cloudlets = []edgeproto.CloudletKey{}
			for _, clName := range oldCloudletPool.Cloudlets {
				newCloudletPool.Cloudlets = append(newCloudletPool.Cloudlets, edgeproto.CloudletKey{
					Name:         clName,
					Organization: newCloudletPool.Key.Organization,
				})
			}
			allApis.cloudletPoolApi.store.STMPut(stm, &newCloudletPool)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func AddSetupSpecificAppDNSRootForCloudlets(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	cloudletKeys, err := getDbObjectKeys(objStore, "Cloudlet")
	if err != nil {
		return err
	}
	for key, _ := range cloudletKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// Get cloudlet
			cloudletStr := stm.Get(key)
			if cloudletStr == "" {
				return nil // was deleted
			}

			cloudlet := edgeproto.Cloudlet{}
			err = json.Unmarshal([]byte(cloudletStr), &cloudlet)
			if err != nil {
				return fmt.Errorf("Unmarshal Cloudlet %s failed: %s", key, err)
			}

			// this will transition xxx.mobiledgex.net to xxx.mobiledgex-qa.net for example, because
			// our deployment scripts are changing the appDnsRoot parameter that is passed on deploy.
			newRootLbFqdn := getCloudletRootLBFQDN(&cloudlet)
			if cloudlet.RootLbFqdn == newRootLbFqdn {
				return nil // already done
			}
			cloudlet.RootLbFqdn = newRootLbFqdn
			allApis.cloudletApi.store.STMPut(stm, &cloudlet)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func OldGetGPUDriverLicenseStoragePath(key *edgeproto.GPUDriverKey) (string, error) {
	path, err := cloudcommon.GetGPUDriverStoragePath(key, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", path, edgeproto.GPUDriverLicenseConfig), nil
}

func OldGetGPUDriverLicenseCloudletStoragePath(key *edgeproto.GPUDriverKey, cloudletName string) (string, error) {
	path, err := cloudcommon.GetGPUDriverStoragePath(key, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", path, cloudletName, edgeproto.GPUDriverLicenseConfig), nil
}

func OldGetGPUDriverBuildStoragePath(key *edgeproto.GPUDriverKey, buildName, ext string) (string, error) {
	path, err := cloudcommon.GetGPUDriverStoragePath(key, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s%s", path, buildName, ext), nil
}

func AddGPUDriverStoragePaths(ctx context.Context, objStore objstore.KVStore, allApis *AllApis) error {
	// Process GPU drivers
	gpuDriverKeys, err := getDbObjectKeys(objStore, "GPUDriver")
	if err != nil {
		return err
	}
	for key, _ := range gpuDriverKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// get gpuDriver
			gpuDriverStr := stm.Get(key)
			if gpuDriverStr == "" {
				return nil // was deleted
			}
			gpuDriver := edgeproto.GPUDriver{}
			err := json.Unmarshal([]byte(gpuDriverStr), &gpuDriver)
			if err != nil {
				return fmt.Errorf("Unmarshal GPUDriver %s failed: %s", key, err)
			}
			if gpuDriver.StorageBucketName != "" && gpuDriver.LicenseConfigStoragePath != "" {
				return nil // already done
			}
			gpuDriver.StorageBucketName = cloudcommon.GetGPUDriverBucketName(nodeMgr.DeploymentTag)
			gpuDriver.LicenseConfigStoragePath, err = OldGetGPUDriverLicenseStoragePath(&gpuDriver.Key)
			if err != nil {
				return err
			}

			for ii, build := range gpuDriver.Builds {
				driverFileName, err := cloudcommon.GetFileNameWithExt(build.DriverPath)
				if err != nil {
					return err
				}
				ext := filepath.Ext(driverFileName)
				if build.StoragePath != "" {
					continue // already done
				}
				gpuDriver.Builds[ii].StoragePath, err = OldGetGPUDriverBuildStoragePath(&gpuDriver.Key, build.Name, ext)
				if err != nil {
					return err
				}
			}
			allApis.gpuDriverApi.store.STMPut(stm, &gpuDriver)
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Process cloudlets
	cloudletKeys, err := getDbObjectKeys(objStore, "Cloudlet")
	if err != nil {
		return err
	}
	for key, _ := range cloudletKeys {
		_, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			// get cloudlet
			cloudletStr := stm.Get(key)
			if cloudletStr == "" {
				return nil // was deleted
			}
			cloudlet := edgeproto.Cloudlet{}
			err := json.Unmarshal([]byte(cloudletStr), &cloudlet)
			if err != nil {
				return fmt.Errorf("Unmarshal Cloudlet %s failed: %s", key, err)
			}
			if cloudlet.LicenseConfigStoragePath != "" {
				return nil // already done
			}
			cloudlet.LicenseConfigStoragePath, err = OldGetGPUDriverLicenseCloudletStoragePath(&cloudlet.GpuConfig.Driver, cloudlet.Key.Name)
			if err != nil {
				return err
			}
			allApis.cloudletApi.store.STMPut(stm, &cloudlet)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
