package upgrade

import "github.com/mobiledgex/edge-cloud/edgeproto"
import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"

// AppInstKey V0
type AppInstKeyV0 struct {
	// App key
	AppKey edgeproto.AppKey `protobuf:"bytes,1,opt,name=app_key,json=appKey" json:"app_key"`
	// Cloudlet on which the App is instantiated
	CloudletKey edgeproto.CloudletKey `protobuf:"bytes,2,opt,name=cloudlet_key,json=cloudletKey" json:"cloudlet_key"`
	// Instance id for defining multiple instances of the same App on the same Cloudlet (not supported yet)
	Id uint64 `protobuf:"fixed64,3,opt,name=id,proto3" json:"id,omitempty"`
}

// V0 of AppInst
type AppInstV0 struct {
	// Fields are used for the Update API to specify which fields to apply
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique identifier key
	Key AppInstKeyV0 `protobuf:"bytes,2,opt,name=key" json:"key"`
	// Cached location of the cloudlet
	CloudletLoc distributed_match_engine.Loc `protobuf:"bytes,3,opt,name=cloudlet_loc,json=cloudletLoc" json:"cloudlet_loc"`
	// Base FQDN (not really URI) for the App. See Service FQDN for endpoint access.
	Uri string `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	// Cluster instance on which this is instantiated
	ClusterInstKey edgeproto.ClusterInstKey `protobuf:"bytes,5,opt,name=cluster_inst_key,json=clusterInstKey" json:"cluster_inst_key"`
	// Liveness of instance (see Liveness)
	Liveness edgeproto.Liveness `protobuf:"varint,6,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
	// For instances accessible via a shared load balancer, defines the external
	// ports on the shared load balancer that map to the internal ports
	// External ports should be appended to the Uri for L4 access.
	MappedPorts []distributed_match_engine.AppPort `protobuf:"bytes,9,rep,name=mapped_ports,json=mappedPorts" json:"mapped_ports"`
	// Flavor defining resource requirements
	Flavor edgeproto.FlavorKey `protobuf:"bytes,12,opt,name=flavor" json:"flavor"`
	// Current state of the AppInst on the Cloudlet
	State edgeproto.TrackedState `protobuf:"varint,14,opt,name=state,proto3,enum=edgeproto.TrackedState" json:"state,omitempty"`
	// Any errors trying to create, update, or delete the AppInst on the Cloudlet
	Errors []string `protobuf:"bytes,15,rep,name=errors" json:"errors,omitempty"`
	// Override actions to CRM
	CrmOverride edgeproto.CRMOverride `protobuf:"varint,16,opt,name=crm_override,json=crmOverride,proto3,enum=edgeproto.CRMOverride" json:"crm_override,omitempty"`
	// Created at time
	CreatedAt distributed_match_engine.Timestamp `protobuf:"bytes,21,opt,name=created_at,json=createdAt" json:"created_at"`
	// Version of this object
	Version uint32 `protobuf:"varint,99,opt,name=version,proto3" json:"version,omitempty"`
}
