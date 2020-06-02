package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	distributed_match_engine1 "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	context "golang.org/x/net/context"
)

func SetDefaultLoadBalancerMaxPortRange(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "SetDefaultLoadBalancerMaxPortRange - default to 50")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Settings"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var settings Settings
		err2 := json.Unmarshal(val, &settings)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if settings.LoadBalancerMaxPortRange == 0 {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Defaulting LoadBalancerMaxPortRange")
			settings.LoadBalancerMaxPortRange = 50
			val, err2 = json.Marshal(settings)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", key, "obj", settings, "err", err2)
				return err2
			}
			objStore.Put(ctx, string(key), string(val))
		}
		return nil
	})
	return err
}

func SetDefaultMaxTrackedDmeClients(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "SetDefaultMaxTrackedDmeClients - default to 100")
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Settings"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var settings Settings
		err2 := json.Unmarshal(val, &settings)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		if settings.MaxTrackedDmeClients == 0 {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Defaulting MaxTrackedDmeClients")
			settings.MaxTrackedDmeClients = 100
			val, err2 = json.Marshal(settings)
			if err2 != nil {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", key, "obj", settings, "err", err2)
				return err2
			}
			objStore.Put(ctx, string(key), string(val))
		}
		return nil
	})
	return err
}

type OperatorCodeV0_OrgRestructure struct {
	Code         string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	OperatorName string `protobuf:"bytes,2,opt,name=operator_name,json=operatorName,proto3" json:"operator_name,omitempty"`
}

type OperatorKeyV0_OrgRestructure struct {
	// Company or Organization name of the operator
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

type CloudletKeyV0_OrgRestructure struct {
	OperatorKey OperatorKeyV0_OrgRestructure `protobuf:"bytes,1,opt,name=operator_key,json=operatorKey" json:"operator_key"`
	Name        string                       `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

type CloudletV0_OrgRestructure struct {
	Fields           []string                                    `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key              CloudletKeyV0_OrgRestructure                `protobuf:"bytes,2,opt,name=key" json:"key"`
	Location         dme.Loc                                     `protobuf:"bytes,5,opt,name=location" json:"location"`
	IpSupport        IpSupport                                   `protobuf:"varint,6,opt,name=ip_support,json=ipSupport,proto3,enum=edgeproto.IpSupport" json:"ip_support,omitempty"`
	StaticIps        string                                      `protobuf:"bytes,7,opt,name=static_ips,json=staticIps,proto3" json:"static_ips,omitempty"`
	NumDynamicIps    int32                                       `protobuf:"varint,8,opt,name=num_dynamic_ips,json=numDynamicIps,proto3" json:"num_dynamic_ips,omitempty"`
	TimeLimits       OperationTimeLimits                         `protobuf:"bytes,9,opt,name=time_limits,json=timeLimits" json:"time_limits"`
	Errors           []string                                    `protobuf:"bytes,10,rep,name=errors" json:"errors,omitempty"`
	Status           StatusInfo                                  `protobuf:"bytes,11,opt,name=status" json:"status"`
	State            TrackedState                                `protobuf:"varint,12,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
	CrmOverride      CRMOverride                                 `protobuf:"varint,13,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
	DeploymentLocal  bool                                        `protobuf:"varint,14,opt,name=deployment_local,json=deploymentLocal,proto3" json:"deployment_local,omitempty"`
	PlatformType     PlatformType                                `protobuf:"varint,15,opt,name=platform_type,json=platformType,proto3,enum=edgeproto.PlatformType" json:"platform_type,omitempty"`
	NotifySrvAddr    string                                      `protobuf:"bytes,16,opt,name=notify_srv_addr,json=notifySrvAddr,proto3" json:"notify_srv_addr,omitempty"`
	Flavor           FlavorKey                                   `protobuf:"bytes,17,opt,name=flavor" json:"flavor"`
	PhysicalName     string                                      `protobuf:"bytes,18,opt,name=physical_name,json=physicalName,proto3" json:"physical_name,omitempty"`
	EnvVar           map[string]string                           `protobuf:"bytes,19,rep,name=env_var,json=envVar" json:"env_var,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ContainerVersion string                                      `protobuf:"bytes,20,opt,name=container_version,json=containerVersion,proto3" json:"container_version,omitempty"`
	Config           PlatformConfig                              `protobuf:"bytes,21,opt,name=config" json:"config"`
	ResTagMap        map[string]*ResTagTableKeyV0_OrgRestructure `protobuf:"bytes,22,rep,name=res_tag_map,json=resTagMap" json:"res_tag_map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value"`
	AccessVars       map[string]string                           `protobuf:"bytes,23,rep,name=access_vars,json=accessVars" json:"access_vars,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	VmImageVersion   string                                      `protobuf:"bytes,24,opt,name=vm_image_version,json=vmImageVersion,proto3" json:"vm_image_version,omitempty"`
	PackageVersion   string                                      `protobuf:"bytes,25,opt,name=package_version,json=packageVersion,proto3" json:"package_version,omitempty"`
}

type CloudletPoolKeyV0_OrgRestructure struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

type CloudletPoolMemberV0_OrgRestructure struct {
	PoolKey     CloudletPoolKey              `protobuf:"bytes,1,opt,name=pool_key,json=poolKey" json:"pool_key"`
	CloudletKey CloudletKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=cloudlet_key,json=cloudletKey" json:"cloudlet_key"`
}

type CloudletInfoV0_OrgRestructure struct {
	Fields            []string                     `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key               CloudletKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=key" json:"key"`
	State             CloudletState                `protobuf:"varint,3,opt,name=state,proto3,enum=edgeproto.CloudletState" json:"state,omitempty"`
	NotifyId          int64                        `protobuf:"varint,4,opt,name=notify_id,json=notifyId,proto3" json:"notify_id,omitempty"`
	Controller        string                       `protobuf:"bytes,5,opt,name=controller,proto3" json:"controller,omitempty"`
	OsMaxRam          uint64                       `protobuf:"varint,6,opt,name=os_max_ram,json=osMaxRam,proto3" json:"os_max_ram,omitempty"`
	OsMaxVcores       uint64                       `protobuf:"varint,7,opt,name=os_max_vcores,json=osMaxVcores,proto3" json:"os_max_vcores,omitempty"`
	OsMaxVolGb        uint64                       `protobuf:"varint,8,opt,name=os_max_vol_gb,json=osMaxVolGb,proto3" json:"os_max_vol_gb,omitempty"`
	Errors            []string                     `protobuf:"bytes,9,rep,name=errors" json:"errors,omitempty"`
	Flavors           []*FlavorInfo                `protobuf:"bytes,10,rep,name=flavors" json:"flavors,omitempty"`
	Status            StatusInfo                   `protobuf:"bytes,11,opt,name=status" json:"status"`
	ContainerVersion  string                       `protobuf:"bytes,12,opt,name=container_version,json=containerVersion,proto3" json:"container_version,omitempty"`
	AvailabilityZones []*OSAZone                   `protobuf:"bytes,13,rep,name=availability_zones,json=availabilityZones" json:"availability_zones,omitempty"`
	OsImages          []*OSImage                   `protobuf:"bytes,14,rep,name=os_images,json=osImages" json:"os_images,omitempty"`
}

type CloudletRefsV0_OrgRestructure struct {
	Key            CloudletKeyV0_OrgRestructure `protobuf:"bytes,1,opt,name=key" json:"key"`
	Clusters       []ClusterKey                 `protobuf:"bytes,2,rep,name=clusters" json:"clusters"`
	UsedRam        uint64                       `protobuf:"varint,4,opt,name=used_ram,json=usedRam,proto3" json:"used_ram,omitempty"`
	UsedVcores     uint64                       `protobuf:"varint,5,opt,name=used_vcores,json=usedVcores,proto3" json:"used_vcores,omitempty"`
	UsedDisk       uint64                       `protobuf:"varint,6,opt,name=used_disk,json=usedDisk,proto3" json:"used_disk,omitempty"`
	RootLbPorts    map[int32]int32              `protobuf:"bytes,8,rep,name=root_lb_ports,json=rootLbPorts" json:"root_lb_ports,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	UsedDynamicIps int32                        `protobuf:"varint,9,opt,name=used_dynamic_ips,json=usedDynamicIps,proto3" json:"used_dynamic_ips,omitempty"`
	UsedStaticIps  string                       `protobuf:"bytes,10,opt,name=used_static_ips,json=usedStaticIps,proto3" json:"used_static_ips,omitempty"`
	OptResUsedMap  map[string]uint32            `protobuf:"bytes,11,rep,name=opt_res_used_map,json=optResUsedMap" json:"opt_res_used_map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

type AppKeyV0_OrgRestructure struct {
	DeveloperKey DeveloperKeyV0_OrgRestructure `protobuf:"bytes,1,opt,name=developer_key,json=developerKey" json:"developer_key"`
	Name         string                        `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Version      string                        `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
}

type AppV0_OrgRestructure struct {
	Fields                  []string                `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                     AppKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=key" json:"key"`
	ImagePath               string                  `protobuf:"bytes,4,opt,name=image_path,json=imagePath,proto3" json:"image_path,omitempty"`
	ImageType               ImageType               `protobuf:"varint,5,opt,name=image_type,json=imageType,proto3,enum=edgeproto.ImageType" json:"image_type,omitempty"`
	AccessPorts             string                  `protobuf:"bytes,7,opt,name=access_ports,json=accessPorts,proto3" json:"access_ports,omitempty"`
	DefaultFlavor           FlavorKey               `protobuf:"bytes,9,opt,name=default_flavor,json=defaultFlavor" json:"default_flavor"`
	AuthPublicKey           string                  `protobuf:"bytes,12,opt,name=auth_public_key,json=authPublicKey,proto3" json:"auth_public_key,omitempty"`
	Command                 string                  `protobuf:"bytes,13,opt,name=command,proto3" json:"command,omitempty"`
	Annotations             string                  `protobuf:"bytes,14,opt,name=annotations,proto3" json:"annotations,omitempty"`
	Deployment              string                  `protobuf:"bytes,15,opt,name=deployment,proto3" json:"deployment,omitempty"`
	DeploymentManifest      string                  `protobuf:"bytes,16,opt,name=deployment_manifest,json=deploymentManifest,proto3" json:"deployment_manifest,omitempty"`
	DeploymentGenerator     string                  `protobuf:"bytes,17,opt,name=deployment_generator,json=deploymentGenerator,proto3" json:"deployment_generator,omitempty"`
	AndroidPackageName      string                  `protobuf:"bytes,18,opt,name=android_package_name,json=androidPackageName,proto3" json:"android_package_name,omitempty"`
	DelOpt                  DeleteType              `protobuf:"varint,20,opt,name=del_opt,json=delOpt,proto3,enum=edgeproto.DeleteType" json:"del_opt,omitempty"`
	Configs                 []*ConfigFile           `protobuf:"bytes,21,rep,name=configs" json:"configs,omitempty"`
	ScaleWithCluster        bool                    `protobuf:"varint,22,opt,name=scale_with_cluster,json=scaleWithCluster,proto3" json:"scale_with_cluster,omitempty"`
	InternalPorts           bool                    `protobuf:"varint,23,opt,name=internal_ports,json=internalPorts,proto3" json:"internal_ports,omitempty"`
	Revision                int32                   `protobuf:"varint,24,opt,name=revision,proto3" json:"revision,omitempty"`
	OfficialFqdn            string                  `protobuf:"bytes,25,opt,name=official_fqdn,json=officialFqdn,proto3" json:"official_fqdn,omitempty"`
	Md5Sum                  string                  `protobuf:"bytes,26,opt,name=md5sum,proto3" json:"md5sum,omitempty"`
	DefaultSharedVolumeSize uint64                  `protobuf:"varint,27,opt,name=default_shared_volume_size,json=defaultSharedVolumeSize,proto3" json:"default_shared_volume_size,omitempty"`
	AutoProvPolicy          string                  `protobuf:"bytes,28,opt,name=auto_prov_policy,json=autoProvPolicy,proto3" json:"auto_prov_policy,omitempty"`
	AccessType              AccessType              `protobuf:"varint,29,opt,name=access_type,json=accessType,proto3,enum=edgeproto.AccessType" json:"access_type,omitempty"`
	DefaultPrivacyPolicy    string                  `protobuf:"bytes,30,opt,name=default_privacy_policy,json=defaultPrivacyPolicy,proto3" json:"default_privacy_policy,omitempty"`
	DeletePrepare           bool                    `protobuf:"varint,31,opt,name=delete_prepare,json=deletePrepare,proto3" json:"delete_prepare,omitempty"`
}

type ClusterInstKeyV0_OrgRestructure struct {
	ClusterKey  ClusterKey                   `protobuf:"bytes,1,opt,name=cluster_key,json=clusterKey" json:"cluster_key"`
	CloudletKey CloudletKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=cloudlet_key,json=cloudletKey" json:"cloudlet_key"`
	Developer   string                       `protobuf:"bytes,3,opt,name=developer,proto3" json:"developer,omitempty"`
}

type DeveloperKeyV0_OrgRestructure struct {
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}
type ClusterInstV0_OrgRestructure struct {
	Fields             []string                        `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                ClusterInstKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=key" json:"key"`
	Flavor             FlavorKey                       `protobuf:"bytes,3,opt,name=flavor" json:"flavor"`
	Liveness           Liveness                        `protobuf:"varint,9,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
	Auto               bool                            `protobuf:"varint,10,opt,name=auto,proto3" json:"auto,omitempty"`
	State              TrackedState                    `protobuf:"varint,4,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
	Errors             []string                        `protobuf:"bytes,5,rep,name=errors" json:"errors,omitempty"`
	CrmOverride        CRMOverride                     `protobuf:"varint,6,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
	IpAccess           IpAccess                        `protobuf:"varint,7,opt,name=ip_access,json=ipAccess,proto3,enum=edgeproto.IpAccess" json:"ip_access,omitempty"`
	AllocatedIp        string                          `protobuf:"bytes,8,opt,name=allocated_ip,json=allocatedIp,proto3" json:"allocated_ip,omitempty"`
	NodeFlavor         string                          `protobuf:"bytes,11,opt,name=node_flavor,json=nodeFlavor,proto3" json:"node_flavor,omitempty"`
	Deployment         string                          `protobuf:"bytes,15,opt,name=deployment,proto3" json:"deployment,omitempty"`
	NumMasters         uint32                          `protobuf:"varint,13,opt,name=num_masters,json=numMasters,proto3" json:"num_masters,omitempty"`
	NumNodes           uint32                          `protobuf:"varint,14,opt,name=num_nodes,json=numNodes,proto3" json:"num_nodes,omitempty"`
	Status             StatusInfo                      `protobuf:"bytes,16,opt,name=status" json:"status"`
	ExternalVolumeSize uint64                          `protobuf:"varint,17,opt,name=external_volume_size,json=externalVolumeSize,proto3" json:"external_volume_size,omitempty"`
	AutoScalePolicy    string                          `protobuf:"bytes,18,opt,name=auto_scale_policy,json=autoScalePolicy,proto3" json:"auto_scale_policy,omitempty"`
	AvailabilityZone   string                          `protobuf:"bytes,19,opt,name=availability_zone,json=availabilityZone,proto3" json:"availability_zone,omitempty"`
	ImageName          string                          `protobuf:"bytes,20,opt,name=image_name,json=imageName,proto3" json:"image_name,omitempty"`
	Reservable         bool                            `protobuf:"varint,21,opt,name=reservable,proto3" json:"reservable,omitempty"`
	ReservedBy         string                          `protobuf:"bytes,22,opt,name=reserved_by,json=reservedBy,proto3" json:"reserved_by,omitempty"`
	SharedVolumeSize   uint64                          `protobuf:"varint,23,opt,name=shared_volume_size,json=sharedVolumeSize,proto3" json:"shared_volume_size,omitempty"`
	PrivacyPolicy      string                          `protobuf:"bytes,24,opt,name=privacy_policy,json=privacyPolicy,proto3" json:"privacy_policy,omitempty"`
	MasterNodeFlavor   string                          `protobuf:"bytes,25,opt,name=master_node_flavor,json=masterNodeFlavor,proto3" json:"master_node_flavor,omitempty"`
}
type AppInstKeyV0_OrgRestructure struct {
	AppKey         AppKeyV0_OrgRestructure         `protobuf:"bytes,1,opt,name=app_key,json=appKey" json:"app_key"`
	ClusterInstKey ClusterInstKeyV0_OrgRestructure `protobuf:"bytes,4,opt,name=cluster_inst_key,json=clusterInstKey" json:"cluster_inst_key"`
}

type AppInstV0_OrgRestructure struct {
	Fields              []string                            `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                 AppInstKeyV0_OrgRestructure         `protobuf:"bytes,2,opt,name=key" json:"key"`
	CloudletLoc         distributed_match_engine.Loc        `protobuf:"bytes,3,opt,name=cloudlet_loc,json=cloudletLoc" json:"cloudlet_loc"`
	Uri                 string                              `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	Liveness            Liveness                            `protobuf:"varint,6,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
	MappedPorts         []distributed_match_engine1.AppPort `protobuf:"bytes,9,rep,name=mapped_ports,json=mappedPorts" json:"mapped_ports"`
	Flavor              FlavorKey                           `protobuf:"bytes,12,opt,name=flavor" json:"flavor"`
	State               TrackedState                        `protobuf:"varint,14,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
	Errors              []string                            `protobuf:"bytes,15,rep,name=errors" json:"errors,omitempty"`
	CrmOverride         CRMOverride                         `protobuf:"varint,16,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
	RuntimeInfo         AppInstRuntime                      `protobuf:"bytes,17,opt,name=runtime_info,json=runtimeInfo" json:"runtime_info"`
	CreatedAt           distributed_match_engine.Timestamp  `protobuf:"bytes,21,opt,name=created_at,json=createdAt" json:"created_at"`
	AutoClusterIpAccess IpAccess                            `protobuf:"varint,22,opt,name=auto_cluster_ip_access,json=autoClusterIpAccess,proto3,enum=edgeproto.IpAccess" json:"auto_cluster_ip_access,omitempty"`
	Status              StatusInfo                          `protobuf:"bytes,23,opt,name=status" json:"status"`
	Revision            int32                               `protobuf:"varint,24,opt,name=revision,proto3" json:"revision,omitempty"`
	ForceUpdate         bool                                `protobuf:"varint,25,opt,name=force_update,json=forceUpdate,proto3" json:"force_update,omitempty"`
	UpdateMultiple      bool                                `protobuf:"varint,26,opt,name=update_multiple,json=updateMultiple,proto3" json:"update_multiple,omitempty"`
	Configs             []*ConfigFile                       `protobuf:"bytes,27,rep,name=configs" json:"configs,omitempty"`
	SharedVolumeSize    uint64                              `protobuf:"varint,28,opt,name=shared_volume_size,json=sharedVolumeSize,proto3" json:"shared_volume_size,omitempty"`
	HealthCheck         HealthCheck                         `protobuf:"varint,29,opt,name=health_check,json=healthCheck,proto3,enum=edgeproto.HealthCheck" json:"health_check,omitempty"`
	PrivacyPolicy       string                              `protobuf:"bytes,30,opt,name=privacy_policy,json=privacyPolicy,proto3" json:"privacy_policy,omitempty"`
	PowerState          PowerState                          `protobuf:"varint,31,opt,name=power_state,json=powerState,proto3,enum=edgeproto.PowerState" json:"power_state,omitempty"`
	ExternalVolumeSize  uint64                              `protobuf:"varint,32,opt,name=external_volume_size,json=externalVolumeSize,proto3" json:"external_volume_size,omitempty"`
	AvailabilityZone    string                              `protobuf:"bytes,33,opt,name=availability_zone,json=availabilityZone,proto3" json:"availability_zone,omitempty"`
	VmFlavor            string                              `protobuf:"bytes,34,opt,name=vm_flavor,json=vmFlavor,proto3" json:"vm_flavor,omitempty"`
}

// this is after org restructure but before revision
type AppInstV1_OrgRestructure struct {
	Fields              []string                            `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                 AppInstKey                          `protobuf:"bytes,2,opt,name=key" json:"key"`
	CloudletLoc         distributed_match_engine.Loc        `protobuf:"bytes,3,opt,name=cloudlet_loc,json=cloudletLoc" json:"cloudlet_loc"`
	Uri                 string                              `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	Liveness            Liveness                            `protobuf:"varint,6,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
	MappedPorts         []distributed_match_engine1.AppPort `protobuf:"bytes,9,rep,name=mapped_ports,json=mappedPorts" json:"mapped_ports"`
	Flavor              FlavorKey                           `protobuf:"bytes,12,opt,name=flavor" json:"flavor"`
	State               TrackedState                        `protobuf:"varint,14,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
	Errors              []string                            `protobuf:"bytes,15,rep,name=errors" json:"errors,omitempty"`
	CrmOverride         CRMOverride                         `protobuf:"varint,16,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
	RuntimeInfo         AppInstRuntime                      `protobuf:"bytes,17,opt,name=runtime_info,json=runtimeInfo" json:"runtime_info"`
	CreatedAt           distributed_match_engine.Timestamp  `protobuf:"bytes,21,opt,name=created_at,json=createdAt" json:"created_at"`
	AutoClusterIpAccess IpAccess                            `protobuf:"varint,22,opt,name=auto_cluster_ip_access,json=autoClusterIpAccess,proto3,enum=edgeproto.IpAccess" json:"auto_cluster_ip_access,omitempty"`
	Status              StatusInfo                          `protobuf:"bytes,23,opt,name=status" json:"status"`
	Revision            int32                               `protobuf:"bytes,24,opt,name=revision,proto3" json:"revision,omitempty"`
	ForceUpdate         bool                                `protobuf:"varint,25,opt,name=force_update,json=forceUpdate,proto3" json:"force_update,omitempty"`
	UpdateMultiple      bool                                `protobuf:"varint,26,opt,name=update_multiple,json=updateMultiple,proto3" json:"update_multiple,omitempty"`
	Configs             []*ConfigFile                       `protobuf:"bytes,27,rep,name=configs" json:"configs,omitempty"`
	SharedVolumeSize    uint64                              `protobuf:"varint,28,opt,name=shared_volume_size,json=sharedVolumeSize,proto3" json:"shared_volume_size,omitempty"`
	HealthCheck         HealthCheck                         `protobuf:"varint,29,opt,name=health_check,json=healthCheck,proto3,enum=edgeproto.HealthCheck" json:"health_check,omitempty"`
	PrivacyPolicy       string                              `protobuf:"bytes,30,opt,name=privacy_policy,json=privacyPolicy,proto3" json:"privacy_policy,omitempty"`
	PowerState          PowerState                          `protobuf:"varint,31,opt,name=power_state,json=powerState,proto3,enum=edgeproto.PowerState" json:"power_state,omitempty"`
	ExternalVolumeSize  uint64                              `protobuf:"varint,32,opt,name=external_volume_size,json=externalVolumeSize,proto3" json:"external_volume_size,omitempty"`
	AvailabilityZone    string                              `protobuf:"bytes,33,opt,name=availability_zone,json=availabilityZone,proto3" json:"availability_zone,omitempty"`
	VmFlavor            string                              `protobuf:"bytes,34,opt,name=vm_flavor,json=vmFlavor,proto3" json:"vm_flavor,omitempty"`
	OptRes              string                              `protobuf:"bytes,35,opt,name=opt_res,json=optRes,proto3" json:"opt_res,omitempty"`
}

type PolicyKeyV0_OrgRestructure struct {
	Developer string `protobuf:"bytes,1,opt,name=developer,proto3" json:"developer,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

type PrivacyPolicyV0_OrgRestructure struct {
	Fields                []string                   `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                   PolicyKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=key" json:"key"`
	OutboundSecurityRules []OutboundSecurityRule     `protobuf:"bytes,3,rep,name=outbound_security_rules,json=outboundSecurityRules" json:"outbound_security_rules"`
}

// this is after org restructure but before revision change
type AppV1_OrgRestructure struct {
	Fields                  []string      `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                     AppKey        `protobuf:"bytes,2,opt,name=key" json:"key"`
	ImagePath               string        `protobuf:"bytes,4,opt,name=image_path,json=imagePath,proto3" json:"image_path,omitempty"`
	ImageType               ImageType     `protobuf:"varint,5,opt,name=image_type,json=imageType,proto3,enum=edgeproto.ImageType" json:"image_type,omitempty"`
	AccessPorts             string        `protobuf:"bytes,7,opt,name=access_ports,json=accessPorts,proto3" json:"access_ports,omitempty"`
	DefaultFlavor           FlavorKey     `protobuf:"bytes,9,opt,name=default_flavor,json=defaultFlavor" json:"default_flavor"`
	AuthPublicKey           string        `protobuf:"bytes,12,opt,name=auth_public_key,json=authPublicKey,proto3" json:"auth_public_key,omitempty"`
	Command                 string        `protobuf:"bytes,13,opt,name=command,proto3" json:"command,omitempty"`
	Annotations             string        `protobuf:"bytes,14,opt,name=annotations,proto3" json:"annotations,omitempty"`
	Deployment              string        `protobuf:"bytes,15,opt,name=deployment,proto3" json:"deployment,omitempty"`
	DeploymentManifest      string        `protobuf:"bytes,16,opt,name=deployment_manifest,json=deploymentManifest,proto3" json:"deployment_manifest,omitempty"`
	DeploymentGenerator     string        `protobuf:"bytes,17,opt,name=deployment_generator,json=deploymentGenerator,proto3" json:"deployment_generator,omitempty"`
	AndroidPackageName      string        `protobuf:"bytes,18,opt,name=android_package_name,json=androidPackageName,proto3" json:"android_package_name,omitempty"`
	DelOpt                  DeleteType    `protobuf:"varint,20,opt,name=del_opt,json=delOpt,proto3,enum=edgeproto.DeleteType" json:"del_opt,omitempty"`
	Configs                 []*ConfigFile `protobuf:"bytes,21,rep,name=configs" json:"configs,omitempty"`
	ScaleWithCluster        bool          `protobuf:"varint,22,opt,name=scale_with_cluster,json=scaleWithCluster,proto3" json:"scale_with_cluster,omitempty"`
	InternalPorts           bool          `protobuf:"varint,23,opt,name=internal_ports,json=internalPorts,proto3" json:"internal_ports,omitempty"`
	Revision                int32         `protobuf:"bytes,24,opt,name=revision,proto3" json:"revision,omitempty"`
	OfficialFqdn            string        `protobuf:"bytes,25,opt,name=official_fqdn,json=officialFqdn,proto3" json:"official_fqdn,omitempty"`
	Md5Sum                  string        `protobuf:"bytes,26,opt,name=md5sum,proto3" json:"md5sum,omitempty"`
	DefaultSharedVolumeSize uint64        `protobuf:"varint,27,opt,name=default_shared_volume_size,json=defaultSharedVolumeSize,proto3" json:"default_shared_volume_size,omitempty"`
	AutoProvPolicy          string        `protobuf:"bytes,28,opt,name=auto_prov_policy,json=autoProvPolicy,proto3" json:"auto_prov_policy,omitempty"`
	AccessType              AccessType    `protobuf:"varint,29,opt,name=access_type,json=accessType,proto3,enum=edgeproto.AccessType" json:"access_type,omitempty"`
	DefaultPrivacyPolicy    string        `protobuf:"bytes,30,opt,name=default_privacy_policy,json=defaultPrivacyPolicy,proto3" json:"default_privacy_policy,omitempty"`
	DeletePrepare           bool          `protobuf:"varint,31,opt,name=delete_prepare,json=deletePrepare,proto3" json:"delete_prepare,omitempty"`
}

type AutoScalePolicyV0_OrgRestructure struct {
	Fields             []string                   `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                PolicyKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=key" json:"key"`
	MinNodes           uint32                     `protobuf:"varint,3,opt,name=min_nodes,json=minNodes,proto3" json:"min_nodes,omitempty"`
	MaxNodes           uint32                     `protobuf:"varint,4,opt,name=max_nodes,json=maxNodes,proto3" json:"max_nodes,omitempty"`
	ScaleUpCpuThresh   uint32                     `protobuf:"varint,5,opt,name=scale_up_cpu_thresh,json=scaleUpCpuThresh,proto3" json:"scale_up_cpu_thresh,omitempty"`
	ScaleDownCpuThresh uint32                     `protobuf:"varint,6,opt,name=scale_down_cpu_thresh,json=scaleDownCpuThresh,proto3" json:"scale_down_cpu_thresh,omitempty"`
	TriggerTimeSec     uint32                     `protobuf:"varint,7,opt,name=trigger_time_sec,json=triggerTimeSec,proto3" json:"trigger_time_sec,omitempty"`
}

type AutoProvPolicyV0_OrgRestructure struct {
	Fields              []string                             `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key                 PolicyKeyV0_OrgRestructure           `protobuf:"bytes,2,opt,name=key" json:"key"`
	DeployClientCount   uint32                               `protobuf:"varint,3,opt,name=deploy_client_count,json=deployClientCount,proto3" json:"deploy_client_count,omitempty"`
	DeployIntervalCount uint32                               `protobuf:"varint,4,opt,name=deploy_interval_count,json=deployIntervalCount,proto3" json:"deploy_interval_count,omitempty"`
	Cloudlets           []*AutoProvCloudletV0_OrgRestructure `protobuf:"bytes,5,rep,name=cloudlets" json:"cloudlets,omitempty"`
}

type AutoProvCloudletV0_OrgRestructure struct {
	Key CloudletKeyV0_OrgRestructure `protobuf:"bytes,1,opt,name=key" json:"key"`
	Loc distributed_match_engine.Loc `protobuf:"bytes,2,opt,name=loc" json:"loc"`
}

type ResTagTableKeyV0_OrgRestructure struct {
	// Resource Table Name
	Name        string                       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	OperatorKey OperatorKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=operator_key,json=operatorKey" json:"operator_key"`
}

type ResTagTableV0_OrgRestructure struct {
	Fields []string                        `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	Key    ResTagTableKeyV0_OrgRestructure `protobuf:"bytes,2,opt,name=key" json:"key"`
	Tags   map[string]string               `protobuf:"bytes,3,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Azone  string                          `protobuf:"bytes,4,opt,name=azone,proto3" json:"azone,omitempty"`
}

func OrgRestructure(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "DevOrgRestructure")
	var operCodeUpgCount uint
	var cloudletUpgCount uint
	var cloudletPoolMemberUpgCount uint
	var cloudletRefsUpgCount uint
	var cloudletInfoUpgCount uint
	var clusterUpgCount uint
	var appUpgCount uint
	var appInstUpgCount uint
	var privacyPolicyUpgCount uint
	var autoScalePolicyUpgCount uint
	var autoProvPolicyUpgCount uint
	var resTagUpgCount uint

	// Operator Codes
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("OperatorCode"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var ocodeV0 OperatorCodeV0_OrgRestructure
		err2 := json.Unmarshal(val, &ocodeV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "ocodeV0", ocodeV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading OperatorCode from V0 to current", "ocodeV0", ocodeV0)
		ocodeV1 := OperatorCode{}
		ocodeV1.Organization = ocodeV0.OperatorName
		ocodeV1.Code = ocodeV0.Code
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded OperatorCode", "ocodeV1", ocodeV1)
		newkey := objstore.DbKeyString("OperatorCode", ocodeV1.GetKey())
		val, err2 = json.Marshal(ocodeV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", ocodeV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		operCodeUpgCount++
		return nil
	})

	// Cloudlets
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("Cloudlet"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var cloudletV0 CloudletV0_OrgRestructure
		err2 := json.Unmarshal(val, &cloudletV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cloudletV0", cloudletV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading Cloudlet from V0 to current", "cloudletV0", cloudletV0)
		cloudletV1 := Cloudlet{}
		cloudletV1.Fields = cloudletV0.Fields
		cloudletV1.Key.Organization = cloudletV0.Key.OperatorKey.Name
		cloudletV1.Key.Name = cloudletV0.Key.Name
		cloudletV1.Location = cloudletV0.Location
		cloudletV1.IpSupport = cloudletV0.IpSupport
		cloudletV1.StaticIps = cloudletV0.StaticIps
		cloudletV1.NumDynamicIps = cloudletV0.NumDynamicIps
		//cloudletV1.TimeLimits = cloudletV0.TimeLimits TODO: investigate this which fails comparison due to duration/float conversion
		cloudletV1.Errors = cloudletV0.Errors
		cloudletV1.Status = cloudletV0.Status
		cloudletV1.State = cloudletV0.State
		cloudletV1.CrmOverride = cloudletV0.CrmOverride
		cloudletV1.DeploymentLocal = cloudletV0.DeploymentLocal
		cloudletV1.PlatformType = cloudletV0.PlatformType
		cloudletV1.NotifySrvAddr = cloudletV0.NotifySrvAddr
		cloudletV1.Flavor = cloudletV0.Flavor
		cloudletV1.PhysicalName = cloudletV0.PhysicalName
		cloudletV1.EnvVar = cloudletV0.EnvVar
		cloudletV1.ContainerVersion = cloudletV0.ContainerVersion
		cloudletV1.Config = cloudletV0.Config
		cloudletV1.ResTagMap = make(map[string]*ResTagTableKey)
		for k, v := range cloudletV0.ResTagMap {
			cloudletV1.ResTagMap[k] = &ResTagTableKey{Name: v.Name, Organization: v.OperatorKey.Name}
		}
		cloudletV1.AccessVars = cloudletV0.AccessVars
		cloudletV1.VmImageVersion = cloudletV0.VmImageVersion
		cloudletV1.PackageVersion = cloudletV0.PackageVersion

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded cloudlet", "cloudletV1", cloudletV1)
		newkey := objstore.DbKeyString("Cloudlet", &cloudletV1.Key)
		val, err2 = json.Marshal(cloudletV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", cloudletV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		cloudletUpgCount++
		return nil
	})

	// Cloudlet Infos
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("CloudletInfo"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var cloudletInfoV0 CloudletInfoV0_OrgRestructure
		err2 := json.Unmarshal(val, &cloudletInfoV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cloudletV0", cloudletInfoV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading Cloudlet from V0 to current", "cloudletInfoV0", cloudletInfoV0)
		cloudletInfoV1 := CloudletInfo{}
		cloudletInfoV1.Fields = cloudletInfoV0.Fields
		cloudletInfoV1.Key.Organization = cloudletInfoV0.Key.OperatorKey.Name
		cloudletInfoV1.Key.Name = cloudletInfoV0.Key.Name
		cloudletInfoV1.State = cloudletInfoV0.State
		cloudletInfoV1.NotifyId = cloudletInfoV0.NotifyId
		cloudletInfoV1.Controller = cloudletInfoV0.Controller
		cloudletInfoV1.OsMaxRam = cloudletInfoV0.OsMaxRam
		cloudletInfoV1.OsMaxVcores = cloudletInfoV0.OsMaxVcores
		cloudletInfoV1.OsMaxVolGb = cloudletInfoV0.OsMaxVolGb
		cloudletInfoV1.Errors = cloudletInfoV0.Errors
		cloudletInfoV1.Flavors = cloudletInfoV0.Flavors
		cloudletInfoV1.Status = cloudletInfoV0.Status
		cloudletInfoV1.ContainerVersion = cloudletInfoV0.ContainerVersion
		cloudletInfoV1.AvailabilityZones = cloudletInfoV0.AvailabilityZones
		cloudletInfoV1.OsImages = cloudletInfoV0.OsImages

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded cloudlet", "cloudletInfoV1", cloudletInfoV1)
		newkey := objstore.DbKeyString("CloudletInfo", &cloudletInfoV1.Key)
		val, err2 = json.Marshal(cloudletInfoV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", cloudletInfoV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		cloudletInfoUpgCount++
		return nil
	})

	// Cloudlet Pool Members
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("CloudletPoolMember"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var cloudletPoolMemberV0 CloudletPoolMemberV0_OrgRestructure
		err2 := json.Unmarshal(val, &cloudletPoolMemberV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cloudletPoolMemberV0", cloudletPoolMemberV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading Cloudlet Pool Member from V0 to current", "cloudletPoolMemberV0", cloudletPoolMemberV0)
		cloudletPoolMemberV1 := CloudletPoolMember{}
		cloudletPoolMemberV1.PoolKey = cloudletPoolMemberV0.PoolKey
		cloudletPoolMemberV1.CloudletKey.Organization = cloudletPoolMemberV0.CloudletKey.OperatorKey.Name
		cloudletPoolMemberV1.CloudletKey.Name = cloudletPoolMemberV0.CloudletKey.Name

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded cloudlet", "cloudletPoolMemberV1", cloudletPoolMemberV1)
		newkey := objstore.DbKeyString("CloudletPoolMember", cloudletPoolMemberV1.GetKey())
		val, err2 = json.Marshal(cloudletPoolMemberV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", cloudletPoolMemberV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		cloudletPoolMemberUpgCount++
		return nil
	})

	// Cloudlet Refs
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("CloudletRefs"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var cloudletRefsV0 CloudletRefsV0_OrgRestructure
		err2 := json.Unmarshal(val, &cloudletRefsV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cloudletRefsV0", cloudletRefsV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading Cloudlet Pool Refs from V0 to current", "cloudletRefsV0", cloudletRefsV0)
		cloudletRefsV1 := CloudletRefs{}
		cloudletRefsV1.Key.Organization = cloudletRefsV0.Key.OperatorKey.Name
		cloudletRefsV1.Key.Name = cloudletRefsV0.Key.Name
		cloudletRefsV1.Clusters = cloudletRefsV0.Clusters
		cloudletRefsV1.UsedRam = cloudletRefsV0.UsedRam
		cloudletRefsV1.UsedVcores = cloudletRefsV0.UsedVcores
		cloudletRefsV1.UsedDisk = cloudletRefsV0.UsedDisk
		cloudletRefsV1.RootLbPorts = cloudletRefsV0.RootLbPorts
		cloudletRefsV1.UsedDynamicIps = cloudletRefsV0.UsedDynamicIps
		cloudletRefsV1.UsedStaticIps = cloudletRefsV0.UsedStaticIps
		cloudletRefsV1.OptResUsedMap = cloudletRefsV0.OptResUsedMap

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded cloudlet refs", "cloudletRefsV1", cloudletRefsV1)
		newkey := objstore.DbKeyString("CloudletRefs", &cloudletRefsV1.Key)
		val, err2 = json.Marshal(cloudletRefsV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", cloudletRefsV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		cloudletRefsUpgCount++
		return nil
	})

	// ClusterInsts
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("ClusterInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var cinstV0 ClusterInstV0_OrgRestructure
		err2 := json.Unmarshal(val, &cinstV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "cinstV0", cinstV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading ClusterInst from V0 to current", "cinstV0", cinstV0)
		cinstV1 := ClusterInst{}
		cinstV1.Fields = cinstV0.Fields
		cinstV1.Key.ClusterKey = cinstV0.Key.ClusterKey
		cinstV1.Key.Organization = cinstV0.Key.Developer
		cinstV1.Key.CloudletKey.Organization = cinstV0.Key.CloudletKey.OperatorKey.Name
		cinstV1.Key.CloudletKey.Name = cinstV0.Key.CloudletKey.Name
		cinstV1.Flavor = cinstV0.Flavor
		cinstV1.Liveness = cinstV0.Liveness
		cinstV1.Auto = cinstV0.Auto
		cinstV1.State = cinstV0.State
		cinstV1.Errors = cinstV0.Errors
		cinstV1.CrmOverride = cinstV0.CrmOverride
		cinstV1.IpAccess = cinstV0.IpAccess
		cinstV1.AllocatedIp = cinstV0.AllocatedIp
		cinstV1.NodeFlavor = cinstV0.NodeFlavor
		cinstV1.Deployment = cinstV0.Deployment
		cinstV1.NumMasters = cinstV0.NumMasters
		cinstV1.NumNodes = cinstV0.NumNodes
		cinstV1.Status = cinstV0.Status
		cinstV1.ExternalVolumeSize = cinstV0.ExternalVolumeSize
		cinstV1.AutoScalePolicy = cinstV0.AutoScalePolicy
		cinstV1.AvailabilityZone = cinstV0.AvailabilityZone
		cinstV1.ImageName = cinstV0.ImageName
		cinstV1.Reservable = cinstV0.Reservable
		cinstV1.ReservedBy = cinstV0.ReservedBy
		cinstV1.SharedVolumeSize = cinstV0.SharedVolumeSize
		cinstV1.PrivacyPolicy = cinstV0.PrivacyPolicy
		cinstV1.MasterNodeFlavor = cinstV0.MasterNodeFlavor

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded ClusterInstance", "cinstV1", cinstV1)
		newkey := objstore.DbKeyString("ClusterInst", &cinstV1.Key)
		val, err2 = json.Marshal(cinstV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", cinstV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		clusterUpgCount++
		return nil
	})
	if err != nil {
		return err
	}

	// Apps
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appV0 AppV0_OrgRestructure
		err2 := json.Unmarshal(val, &appV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appV0", appV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading App from V0 to current", "appV0", appV0)
		appV1 := AppV1_OrgRestructure{}
		appV1.Fields = appV0.Fields
		appV1.Key.Organization = appV0.Key.DeveloperKey.Name
		appV1.Key.Name = appV0.Key.Name
		appV1.Key.Version = appV0.Key.Version
		appV1.ImagePath = appV0.ImagePath
		appV1.ImageType = appV0.ImageType
		appV1.AccessPorts = appV0.AccessPorts
		appV1.DefaultFlavor = appV0.DefaultFlavor
		appV1.AuthPublicKey = appV0.AuthPublicKey
		appV1.Command = appV0.Command

		appV1.Annotations = appV0.Annotations
		appV1.Deployment = appV0.Deployment
		appV1.DeploymentManifest = appV0.DeploymentManifest
		appV1.DeploymentGenerator = appV0.DeploymentGenerator
		appV1.AndroidPackageName = appV0.AndroidPackageName

		appV1.DelOpt = appV0.DelOpt
		appV1.Configs = appV0.Configs
		appV1.InternalPorts = appV0.InternalPorts
		appV1.Revision = appV0.Revision

		appV1.OfficialFqdn = appV0.OfficialFqdn
		appV1.Md5Sum = appV0.Md5Sum
		appV1.DefaultSharedVolumeSize = appV0.DefaultSharedVolumeSize

		appV1.AutoProvPolicy = appV0.AutoProvPolicy
		appV1.AccessType = appV0.AccessType
		appV1.DefaultPrivacyPolicy = appV0.DefaultPrivacyPolicy
		appV1.DeletePrepare = appV0.DeletePrepare

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded app", "appV1", appV1)
		newkey := objstore.DbKeyString("App", &appV1.Key)
		val, err2 = json.Marshal(appV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", appV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		appUpgCount++
		return nil
	})
	if err != nil {
		return err
	}

	// AppInsts
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInstV0 AppInstV0_OrgRestructure
		err2 := json.Unmarshal(val, &appInstV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appInstV0", appInstV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading AppInst from V0 to current", "appInstV0", appInstV0)
		appInstV1 := AppInstV1_OrgRestructure{}
		appInstV1.Fields = appInstV0.Fields
		appInstV1.Key.AppKey.Organization = appInstV0.Key.AppKey.DeveloperKey.Name
		appInstV1.Key.AppKey.Name = appInstV0.Key.AppKey.Name
		appInstV1.Key.AppKey.Version = appInstV0.Key.AppKey.Version
		appInstV1.Key.ClusterInstKey.ClusterKey.Name = appInstV0.Key.ClusterInstKey.ClusterKey.Name
		appInstV1.Key.ClusterInstKey.Organization = appInstV0.Key.ClusterInstKey.Developer
		appInstV1.Key.ClusterInstKey.CloudletKey.Organization = appInstV0.Key.ClusterInstKey.CloudletKey.OperatorKey.Name
		appInstV1.Key.ClusterInstKey.CloudletKey.Name = appInstV0.Key.ClusterInstKey.CloudletKey.Name
		appInstV1.CloudletLoc = appInstV0.CloudletLoc
		appInstV1.Uri = appInstV0.Uri
		appInstV1.Liveness = appInstV0.Liveness
		appInstV1.MappedPorts = appInstV0.MappedPorts
		appInstV1.Flavor = appInstV0.Flavor
		appInstV1.State = appInstV0.State
		appInstV1.Errors = appInstV0.Errors
		appInstV1.CrmOverride = appInstV0.CrmOverride
		appInstV1.RuntimeInfo = appInstV0.RuntimeInfo
		appInstV1.CreatedAt = appInstV0.CreatedAt
		appInstV1.AutoClusterIpAccess = appInstV0.AutoClusterIpAccess
		appInstV1.Status = appInstV0.Status
		appInstV1.Revision = appInstV0.Revision
		appInstV1.ForceUpdate = appInstV0.ForceUpdate
		appInstV1.UpdateMultiple = appInstV0.UpdateMultiple
		appInstV1.Configs = appInstV0.Configs
		appInstV1.SharedVolumeSize = appInstV0.SharedVolumeSize
		appInstV1.HealthCheck = appInstV0.HealthCheck
		appInstV1.PrivacyPolicy = appInstV0.PrivacyPolicy
		appInstV1.PowerState = appInstV0.PowerState
		appInstV1.ExternalVolumeSize = appInstV0.ExternalVolumeSize
		appInstV1.AvailabilityZone = appInstV0.AvailabilityZone
		appInstV1.VmFlavor = appInstV0.VmFlavor

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded appinst", "appInstV1", appInstV1)
		newkey := objstore.DbKeyString("AppInst", &appInstV1.Key)
		val, err2 = json.Marshal(appInstV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", appInstV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		appInstUpgCount++
		return nil
	})
	if err != nil {
		return err
	}

	// PrivacyPolicies
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("PrivacyPolicy"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var privPolV0 PrivacyPolicyV0_OrgRestructure
		err2 := json.Unmarshal(val, &privPolV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "privPolV0", privPolV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading PrivacyPolicy from V0 to current", "privPolV0", privPolV0)
		privPolV1 := PrivacyPolicy{}
		privPolV1.Fields = privPolV0.Fields
		privPolV1.Key.Organization = privPolV0.Key.Developer
		privPolV1.Key.Name = privPolV0.Key.Name
		privPolV1.OutboundSecurityRules = privPolV0.OutboundSecurityRules

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded PrivacyPolicy", "ppv1", privPolV1)
		newkey := objstore.DbKeyString("PrivacyPolicy", &privPolV1.Key)
		val, err2 = json.Marshal(privPolV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", privPolV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		privacyPolicyUpgCount++
		return nil
	})
	if err != nil {
		return err
	}

	// AutoScalePolicies
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AutoScalePolicy"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var autoScalPolV0 AutoScalePolicyV0_OrgRestructure
		err2 := json.Unmarshal(val, &autoScalPolV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "autoScalPolV0", autoScalPolV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading AutoScalePolicy from V0 to current", "autoScalPolV0", autoScalPolV0)
		autoScalPolV1 := AutoScalePolicy{}
		autoScalPolV1.Fields = autoScalPolV0.Fields
		autoScalPolV1.Key.Organization = autoScalPolV0.Key.Developer
		autoScalPolV1.Key.Name = autoScalPolV0.Key.Name
		autoScalPolV1.MinNodes = autoScalPolV0.MinNodes
		autoScalPolV1.MaxNodes = autoScalPolV0.MaxNodes
		autoScalPolV1.ScaleUpCpuThresh = autoScalPolV0.ScaleUpCpuThresh
		autoScalPolV1.ScaleDownCpuThresh = autoScalPolV0.ScaleDownCpuThresh
		autoScalPolV1.TriggerTimeSec = autoScalPolV0.TriggerTimeSec

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded AutoScalePolicy", "autoScalPolV1", autoScalPolV1)
		newkey := objstore.DbKeyString("AutoScalePolicy", &autoScalPolV1.Key)
		val, err2 = json.Marshal(autoScalPolV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", autoScalPolV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		autoScalePolicyUpgCount++
		return nil
	})
	if err != nil {
		return err
	}

	// AutoProvPolicies
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AutoProvPolicy"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var autoProvPolV0 AutoProvPolicyV0_OrgRestructure
		err2 := json.Unmarshal(val, &autoProvPolV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "autoProvPolV0", autoProvPolV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading AutoProvPolicy from V0 to current", "autoProvPolV0", autoProvPolV0)
		autoProvPolV1 := AutoProvPolicy{}
		autoProvPolV1.Fields = autoProvPolV0.Fields
		autoProvPolV1.Key.Organization = autoProvPolV0.Key.Developer
		autoProvPolV1.Key.Name = autoProvPolV0.Key.Name
		autoProvPolV1.DeployClientCount = autoProvPolV0.DeployClientCount
		autoProvPolV1.DeployIntervalCount = autoProvPolV0.DeployIntervalCount
		autoProvPolV1.Cloudlets = []*AutoProvCloudlet{}
		for _, v := range autoProvPolV0.Cloudlets {
			ckey := CloudletKey{Name: v.Key.Name, Organization: v.Key.OperatorKey.Name}
			autoProvPolV1.Cloudlets = append(autoProvPolV1.Cloudlets, &AutoProvCloudlet{Key: ckey, Loc: v.Loc})
		}

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded AutoProvPolicy", "autoProvPolV1", autoProvPolV1)
		newkey := objstore.DbKeyString("AutoProvPolicy", &autoProvPolV1.Key)
		val, err2 = json.Marshal(autoProvPolV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", autoProvPolV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		autoProvPolicyUpgCount++
		return nil
	})

	// ResTagTAbles
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("ResTagTable"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var resTagV0 ResTagTableV0_OrgRestructure
		err2 := json.Unmarshal(val, &resTagV0)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "resTagV0", resTagV0)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading ResTagTable from V0 to current", "resTagV0", resTagV0)
		resTagV1 := ResTagTable{}
		resTagV1.Key.Organization = resTagV0.Key.OperatorKey.Name
		resTagV1.Key.Name = resTagV0.Key.Name
		resTagV1.Fields = resTagV0.Fields
		resTagV1.Tags = resTagV0.Tags
		resTagV1.Azone = resTagV0.Azone

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded AutoProvPolicy", "resTagV1", resTagV1)
		newkey := objstore.DbKeyString("ResTagTable", &resTagV1.Key)
		val, err2 = json.Marshal(resTagV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", resTagV1, "err", err2)
			return err2
		}
		_, err2 = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
			stm.Del(string(key))
			stm.Put(newkey, string(val))
			return nil
		})
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to write new data to etcd", "err", err2)
			return err2
		}
		autoProvPolicyUpgCount++
		return nil
	})

	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrade object counts",
		"operCodeUpgCount", operCodeUpgCount,
		"cloudletUpgCount", cloudletUpgCount,
		"cloudletInfoUpgCount", cloudletInfoUpgCount,
		"cloudletPoolMemberUpgCount", cloudletPoolMemberUpgCount,
		"cloudletRefsUpgCount", cloudletRefsUpgCount,
		"clusterUpgCount", clusterUpgCount,
		"appUpgCount", appUpgCount,
		"appInstUpgCount", appInstUpgCount,
		"privacyPolicyUpgCount", privacyPolicyUpgCount,
		"autoScalePolicyUpgCount", autoScalePolicyUpgCount,
		"autoProvPolicyUpgCount", autoProvPolicyUpgCount,
		"resTagUpgCount", resTagUpgCount)

	return err
}

func AppRevision(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "AppRevision")

	// Apps
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appV1 AppV1_OrgRestructure
		err2 := json.Unmarshal(val, &appV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appV1", appV1)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading App from V1 to current", "appV1", appV1)
		appV2 := App{}
		appV2.Fields = appV1.Fields
		appV2.Key = appV1.Key
		appV2.ImagePath = appV1.ImagePath
		appV2.ImageType = appV1.ImageType
		appV2.AccessPorts = appV1.AccessPorts
		appV2.DefaultFlavor = appV1.DefaultFlavor
		appV2.AuthPublicKey = appV1.AuthPublicKey
		appV2.Command = appV1.Command

		appV2.Annotations = appV1.Annotations
		appV2.Deployment = appV1.Deployment
		appV2.DeploymentManifest = appV1.DeploymentManifest
		appV2.DeploymentGenerator = appV1.DeploymentGenerator
		appV2.AndroidPackageName = appV1.AndroidPackageName

		appV2.DelOpt = appV1.DelOpt
		appV2.Configs = appV1.Configs
		appV2.InternalPorts = appV1.InternalPorts
		if appV1.Revision != 0 {
			appV2.Revision = fmt.Sprintf("%d", appV1.Revision)
		}

		appV2.OfficialFqdn = appV1.OfficialFqdn
		appV2.Md5Sum = appV1.Md5Sum
		appV2.DefaultSharedVolumeSize = appV1.DefaultSharedVolumeSize

		appV2.AutoProvPolicy = appV1.AutoProvPolicy
		appV2.AccessType = appV1.AccessType
		appV2.DefaultPrivacyPolicy = appV1.DefaultPrivacyPolicy
		appV2.DeletePrepare = appV1.DeletePrepare

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded app", "appV2", appV2)
		newkey := objstore.DbKeyString("App", &appV2.Key)
		val, err2 = json.Marshal(appV2)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", appV2, "err", err2)
			return err2
		}
		objStore.Put(ctx, newkey, string(val))
		return nil
	})
	if err != nil {
		return err
	}

	// AppInsts
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInstV1 AppInstV1_OrgRestructure
		err2 := json.Unmarshal(val, &appInstV1)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appInstV1", appInstV1)
			return err2
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgrading AppInst from V1 to current", "appInstV1", appInstV1)
		appInstV2 := AppInst{}
		appInstV2.Fields = appInstV1.Fields
		appInstV2.Key = appInstV1.Key
		appInstV2.CloudletLoc = appInstV1.CloudletLoc
		appInstV2.Uri = appInstV1.Uri
		appInstV2.Liveness = appInstV1.Liveness
		appInstV2.MappedPorts = appInstV1.MappedPorts
		appInstV2.Flavor = appInstV1.Flavor
		appInstV2.State = appInstV1.State
		appInstV2.Errors = appInstV1.Errors
		appInstV2.CrmOverride = appInstV1.CrmOverride
		appInstV2.RuntimeInfo = appInstV1.RuntimeInfo
		appInstV2.CreatedAt = appInstV1.CreatedAt
		appInstV2.AutoClusterIpAccess = appInstV1.AutoClusterIpAccess
		appInstV2.Status = appInstV1.Status
		if appInstV1.Revision != 0 {
			appInstV2.Revision = fmt.Sprintf("%d", appInstV1.Revision)
		}
		appInstV2.ForceUpdate = appInstV1.ForceUpdate
		appInstV2.UpdateMultiple = appInstV1.UpdateMultiple
		appInstV2.Configs = appInstV1.Configs
		appInstV2.SharedVolumeSize = appInstV1.SharedVolumeSize
		appInstV2.HealthCheck = appInstV1.HealthCheck
		appInstV2.PrivacyPolicy = appInstV1.PrivacyPolicy
		appInstV2.PowerState = appInstV1.PowerState
		appInstV2.ExternalVolumeSize = appInstV1.ExternalVolumeSize
		appInstV2.AvailabilityZone = appInstV1.AvailabilityZone
		appInstV2.VmFlavor = appInstV1.VmFlavor
		appInstV2.OptRes = appInstV1.OptRes

		log.SpanLog(ctx, log.DebugLevelUpgrade, "Upgraded appinst", "appInstV2", appInstV2)
		newkey := objstore.DbKeyString("AppInst", &appInstV2.Key)
		val, err2 = json.Marshal(appInstV2)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", newkey, "obj", appInstV2, "err", err2)
			return err2
		}
		objStore.Put(ctx, newkey, string(val))
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func AppInstRefsUpgrade(ctx context.Context, objStore objstore.KVStore) error {
	log.SpanLog(ctx, log.DebugLevelUpgrade, "AppInstRefs")

	// all refs
	allrefs := make(map[AppKey]*AppInstRefs)

	// read Apps
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("App"))
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		app := App{}
		err2 := json.Unmarshal(val, &app)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "app", app)
			return err2
		}
		refs := &AppInstRefs{}
		refs.Key = app.Key
		refs.Insts = make(map[string]uint32)
		allrefs[app.Key] = refs
		return nil
	})
	if err != nil {
		return err
	}

	// read AppInsts
	keystr = fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	err = objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		inst := AppInst{}
		err2 := json.Unmarshal(val, &inst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2)
			return err2
		}
		refs, found := allrefs[inst.Key.AppKey]
		if !found {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "App not found for AppInst", "appinst", inst.Key)
			return inst.Key.AppKey.NotFoundError()
		}
		refs.Insts[inst.Key.GetKeyString()] = 1
		return nil
	})
	if err != nil {
		return err
	}

	// Write refs
	for _, refs := range allrefs {
		log.SpanLog(ctx, log.DebugLevelUpgrade, "New AppInstRefs", "app", refs.Key)
		keystr := objstore.DbKeyString("AppInstRefs", &refs.Key)
		val, err2 := json.Marshal(refs)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Failed to marshal obj", "key", keystr, "obj", refs, "err", err2)
			return err2
		}
		objStore.Put(ctx, keystr, string(val))
	}
	if err != nil {
		return err
	}
	return nil
}
