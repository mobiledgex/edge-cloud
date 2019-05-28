package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AppPortV0_LintFixAppInstFields struct {
	Proto        distributed_match_engine.LProto `protobuf:"varint,1,opt,name=proto,proto3,enum=distributed_match_engine.LProto" json:"proto,omitempty"`
	InternalPort int32                           `protobuf:"varint,2,opt,name=internal_port,json=internalPort,proto3" json:"internal_port,omitempty"`
	PublicPort   int32                           `protobuf:"varint,3,opt,name=public_port,json=publicPort,proto3" json:"public_port,omitempty"`
	PathPrefix   string                          `protobuf:"bytes,4,opt,name=path_prefix,json=pathPrefix,proto3" json:"path_prefix,omitempty"`
	FQDNPrefix   string                          `protobuf:"bytes,5,opt,name=FQDN_prefix,json=FQDNPrefix,proto3" json:"FQDN_prefix,omitempty"`
}

type AppInstV1_LintFixAppInstFields struct {
	Fields              []string                           `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                 AppInstKey                         `protobuf:"bytes,2,opt,name=key" json:"key"`
	CloudletLoc         distributed_match_engine.Loc       `protobuf:"bytes,3,opt,name=cloudlet_loc,json=cloudletLoc" json:"cloudlet_loc"`
	Uri                 string                             `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	Liveness            Liveness                           `protobuf:"varint,6,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
	MappedPorts         []AppPortV0_LintFixAppInstFields   `protobuf:"bytes,9,rep,name=mapped_ports,json=mappedPorts" json:"mapped_ports"`
	Flavor              FlavorKey                          `protobuf:"bytes,12,opt,name=flavor" json:"flavor"`
	State               TrackedState                       `protobuf:"varint,14,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
	Errors              []string                           `protobuf:"bytes,15,rep,name=errors" json:"errors,omitempty"`
	CrmOverride         CRMOverride                        `protobuf:"varint,16,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
	RuntimeInfo         AppInstRuntime                     `protobuf:"bytes,17,opt,name=runtime_info,json=runtimeInfo" json:"runtime_info"`
	CreatedAt           distributed_match_engine.Timestamp `protobuf:"bytes,21,opt,name=created_at,json=createdAt" json:"created_at"`
	AutoClusterIpAccess IpAccess                           `protobuf:"varint,22,opt,name=auto_cluster_ip_access,json=autoClusterIpAccess,proto3,enum=edgeproto.IpAccess" json:"auto_cluster_ip_access,omitempty"`
}

// Below is an implementation of AddClusterInstKeyToAppInstKey
// This upgrade modifies AppInstKey to include ClusterInstKey rather than CloudletKey+Id
// This allows multiple instances of an app on the same cloudlet, but different cluster instances
func AddClusterInstKeyToAppInstKey(objStore objstore.KVStore) error {
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
		MappedPorts    []AppPortV0_LintFixAppInstFields           `protobuf:"bytes,9,rep,name=mapped_ports,json=mappedPorts" json:"mapped_ports"`
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
		appV1 := AppInstV1_LintFixAppInstFields{}
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
		objStore.Delete(string(key))
		newkey := objstore.DbKeyString("AppInst", &appV1.Key)
		val, err2 = json.Marshal(appV1)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", appV1, "err", err2)
			return err2
		}
		objStore.Put(newkey, string(val))
		upgCount++
		return nil
	})
	log.DebugLog(log.DebugLevelUpgrade, "Upgrade object count", "Upgrade Count", upgCount)
	return err
}

// Below is an implementation of LintFixAppInstFields
// This upgrade modifies takes care of lint fix for AppPort JSON field
// The change is in effect after marshalling AppV2
func LintFixAppInstFields(objStore objstore.KVStore) error {
	var upgCount uint
	log.DebugLog(log.DebugLevelUpgrade, "LintFixAppInstFields - Fix lint issue on AppInst Fields")
	// Define a prefix for a walk
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err := objStore.List(keystr, func(key, val []byte, rev int64) error {
		var appV1 AppInstV1_LintFixAppInstFields
		err2 := json.Unmarshal(val, &appV1)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		log.DebugLog(log.DebugLevelUpgrade, "Upgrading AppInst from V1 to V2", "AppInstV1", appV1)
		appV2 := AppInst{}
		appV2.Fields = appV1.Fields
		appV2.Key = appV1.Key
		appV2.CloudletLoc = appV1.CloudletLoc
		appV2.Uri = appV1.Uri
		appV2.Liveness = appV1.Liveness
		for _, mappedPortV1 := range appV1.MappedPorts {
			mappedPortV2 := distributed_match_engine.AppPort{}
			mappedPortV2.Proto = mappedPortV1.Proto
			mappedPortV2.InternalPort = mappedPortV1.InternalPort
			mappedPortV2.PublicPort = mappedPortV1.PublicPort
			mappedPortV2.PathPrefix = mappedPortV1.PathPrefix
			mappedPortV2.FqdnPrefix = mappedPortV1.FQDNPrefix
			appV2.MappedPorts = append(appV2.MappedPorts, mappedPortV2)
		}
		appV2.Flavor = appV1.Flavor
		appV2.State = appV1.State
		appV2.Errors = appV1.Errors
		appV2.CrmOverride = appV1.CrmOverride
		appV2.RuntimeInfo = appV1.RuntimeInfo
		appV2.CreatedAt = appV1.CreatedAt
		appV2.AutoClusterIpAccess = appV1.AutoClusterIpAccess
		log.DebugLog(log.DebugLevelUpgrade, "Upgrade lint: fix fqdn name", "AppInstV2", appV2)
		val, err2 = json.Marshal(appV2)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Failed to marshal obj", "key", key, "obj", appV2, "err", err2)
			return err2
		}
		objStore.Put(string(key), string(val))
		upgCount++
		return nil
	})
	log.DebugLog(log.DebugLevelUpgrade, "Upgrade object count", "Upgrade Count", upgCount)
	return err
}

func AddClusterInstDeploymentField(objStore objstore.KVStore) error {
	var upgCount uint
	log.DebugLog(log.DebugLevelUpgrade, "AddClusterInstDeploymentField - defaults to kubernetes")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("ClusterInst"))
	err := objStore.List(keystr, func(key, val []byte, rev int64) error {
		dat := make(map[string]interface{})
		err2 := json.Unmarshal(val, &dat)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if _, found := dat["deployment"]; found {
			return nil
		}
		dat["deployment"] = "kubernetes"
		log.DebugLog(log.DebugLevelUpgrade, "Added ClusterInst deployment field", "key", key)
		val, err2 = json.Marshal(dat)
		if err2 != nil {
			log.DebugLog(log.DebugLevelUpgrade, "Failed to marshal obj", "key", key, "obj", dat, "err", err2)
			return err2
		}
		objStore.Put(string(key), string(val))
		upgCount++
		return nil
	})
	log.DebugLog(log.DebugLevelUpgrade, "Upgrade object count", "Upgrade Count", upgCount)
	return err
}
