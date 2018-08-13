# Protocol Documentation
<a name="top"/>

## Table of Contents

- [app.proto](#app.proto)
    - [App](#edgeproto.App)
    - [AppKey](#edgeproto.AppKey)
  
    - [AccessLayer](#edgeproto.AccessLayer)
    - [ImageType](#edgeproto.ImageType)
  
  
    - [AppApi](#edgeproto.AppApi)
  

- [app_inst.proto](#app_inst.proto)
    - [AppInst](#edgeproto.AppInst)
    - [AppInstInfo](#edgeproto.AppInstInfo)
    - [AppInstKey](#edgeproto.AppInstKey)
    - [AppInstMetrics](#edgeproto.AppInstMetrics)
    - [AppPort](#edgeproto.AppPort)
  
    - [AppState](#edgeproto.AppState)
  
  
    - [AppInstApi](#edgeproto.AppInstApi)
    - [AppInstInfoApi](#edgeproto.AppInstInfoApi)
    - [AppInstMetricsApi](#edgeproto.AppInstMetricsApi)
  

- [cloud-resource-manager.proto](#cloud-resource-manager.proto)
    - [CloudResource](#edgeproto.CloudResource)
    - [EdgeCloudApp](#edgeproto.EdgeCloudApp)
    - [EdgeCloudApplication](#edgeproto.EdgeCloudApplication)
  
    - [CloudResourceCategory](#edgeproto.CloudResourceCategory)
  
  
    - [CloudResourceManager](#edgeproto.CloudResourceManager)
  

- [cloudlet.proto](#cloudlet.proto)
    - [Cloudlet](#edgeproto.Cloudlet)
    - [CloudletInfo](#edgeproto.CloudletInfo)
    - [CloudletKey](#edgeproto.CloudletKey)
    - [CloudletMetrics](#edgeproto.CloudletMetrics)
  
    - [CloudletState](#edgeproto.CloudletState)
  
  
    - [CloudletApi](#edgeproto.CloudletApi)
    - [CloudletInfoApi](#edgeproto.CloudletInfoApi)
    - [CloudletMetricsApi](#edgeproto.CloudletMetricsApi)
  

- [cluster.proto](#cluster.proto)
    - [Cluster](#edgeproto.Cluster)
    - [ClusterKey](#edgeproto.ClusterKey)
  
  
  
    - [ClusterApi](#edgeproto.ClusterApi)
  

- [clusterflavor.proto](#clusterflavor.proto)
    - [ClusterFlavor](#edgeproto.ClusterFlavor)
    - [ClusterFlavorKey](#edgeproto.ClusterFlavorKey)
  
  
  
    - [ClusterFlavorApi](#edgeproto.ClusterFlavorApi)
  

- [clusterinst.proto](#clusterinst.proto)
    - [ClusterInst](#edgeproto.ClusterInst)
    - [ClusterInstInfo](#edgeproto.ClusterInstInfo)
    - [ClusterInstKey](#edgeproto.ClusterInstKey)
  
    - [ClusterState](#edgeproto.ClusterState)
  
  
    - [ClusterInstApi](#edgeproto.ClusterInstApi)
    - [ClusterInstInfoApi](#edgeproto.ClusterInstInfoApi)
  

- [common.proto](#common.proto)
  
    - [IpSupport](#edgeproto.IpSupport)
    - [L4Proto](#edgeproto.L4Proto)
    - [Liveness](#edgeproto.Liveness)
  
  
  

- [developer.proto](#developer.proto)
    - [Developer](#edgeproto.Developer)
    - [DeveloperKey](#edgeproto.DeveloperKey)
  
  
  
    - [DeveloperApi](#edgeproto.DeveloperApi)
  

- [flavor.proto](#flavor.proto)
    - [Flavor](#edgeproto.Flavor)
    - [FlavorKey](#edgeproto.FlavorKey)
  
  
  
    - [FlavorApi](#edgeproto.FlavorApi)
  

- [notice.proto](#notice.proto)
    - [NoticeReply](#edgeproto.NoticeReply)
    - [NoticeRequest](#edgeproto.NoticeRequest)
  
    - [NoticeAction](#edgeproto.NoticeAction)
    - [NoticeRequestor](#edgeproto.NoticeRequestor)
  
  
    - [NotifyApi](#edgeproto.NotifyApi)
  

- [operator.proto](#operator.proto)
    - [Operator](#edgeproto.Operator)
    - [OperatorKey](#edgeproto.OperatorKey)
  
  
  
    - [OperatorApi](#edgeproto.OperatorApi)
  

- [refs.proto](#refs.proto)
    - [CloudletRefs](#edgeproto.CloudletRefs)
    - [CloudletRefs.RootLbPortsEntry](#edgeproto.CloudletRefs.RootLbPortsEntry)
    - [ClusterRefs](#edgeproto.ClusterRefs)
  
  
  
    - [CloudletRefsApi](#edgeproto.CloudletRefsApi)
    - [ClusterRefsApi](#edgeproto.ClusterRefsApi)
  

- [result.proto](#result.proto)
    - [Result](#edgeproto.Result)
  
  
  
  

- [Scalar Value Types](#scalar-value-types)



<a name="app.proto"/>
<p align="right"><a href="#top">Top</a></p>

## app.proto



<a name="edgeproto.App"/>

### App
Apps are applications that may be instantiated on Cloudlets, providing a back-end service to an application client (using the mobiledgex SDK) running on a user device such as a cell phone, wearable, drone, etc. Applications belong to Developers, and must specify their image and accessibility. Applications are analagous to Pods in Kubernetes, and similarly are tied to a Cluster.
An application in itself is not tied to a Cloudlet, but provides a definition that can be used to instantiate it on a Cloudlet. AppInsts are applications instantiated on a particular Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppKey](#edgeproto.AppKey) |  | Unique identifier key |
| image_path | [string](#string) |  | Path to the application container or VM on image repo |
| image_type | [ImageType](#edgeproto.ImageType) |  | Image type (see ImageType) |
| access_layer | [AccessLayer](#edgeproto.AccessLayer) |  | Access layer(s) (see AccessLayer) |
| access_ports | [string](#string) |  | For Layer4 access, the ports the app listens on. This is a comma separated list of protocol:port pairs, i.e. tcp:80,tcp:443,udp:10002. Only tcp and udp protocols are supported. |
| config_map | [string](#string) |  | Initial config passed to image (docker only?). is this a string format of the file or a pointer to the file stored elsewhere? |
| default_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Default flavor for the App, may be overridden by the AppInst |
| cluster | [ClusterKey](#edgeproto.ClusterKey) |  | Cluster on which the App can be instantiated. If not specified during create, a cluster will be automatically created. If specified, it must exist. |






<a name="edgeproto.AppKey"/>

### AppKey
AppKey uniquely identifies an Application.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| developer_key | [DeveloperKey](#edgeproto.DeveloperKey) |  | Developer key |
| name | [string](#string) |  | Application name |
| version | [string](#string) |  | Version of the app |





 


<a name="edgeproto.AccessLayer"/>

### AccessLayer
AccessLayer defines what layers are exposed for the application.

| Name | Number | Description |
| ---- | ------ | ----------- |
| AccessLayerUnknown | 0 | No external access for the application |
| AccessLayerL4 | 1 | Layer 4 (tcp/udp) access |
| AccessLayerL7 | 2 | Layer 7 (https path) access |
| AccessLayerL4L7 | 3 | Both layer 4 and layer 4 access |



<a name="edgeproto.ImageType"/>

### ImageType
ImageType specifies the image type of the application.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ImageTypeUnknown | 0 | Unknown image type |
| ImageTypeDocker | 1 | Docker container image type |
| ImageTypeQCOW | 2 | QCOW2 virtual machine image type |


 

 


<a name="edgeproto.AppApi"/>

### AppApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateApp | [App](#edgeproto.App) | [Result](#edgeproto.App) | Create an application |
| DeleteApp | [App](#edgeproto.App) | [Result](#edgeproto.App) | Delete an application |
| UpdateApp | [App](#edgeproto.App) | [Result](#edgeproto.App) | Update an application |
| ShowApp | [App](#edgeproto.App) | [App](#edgeproto.App) | Show applications. Any fields specified will be used to filter results. |

 



<a name="app_inst.proto"/>
<p align="right"><a href="#top">Top</a></p>

## app_inst.proto



<a name="edgeproto.AppInst"/>

### AppInst
AppInst is an instance of an App (application) on a Cloudlet. It is defined by an App plus a Cloudlet key. This separation of the definition of the App versus its instantiation is unique to Mobiledgex, and allows the Developer to provide the App defintion, while either the Developer may statically define the instances, or the Mobiledgex platform may dynamically create and destroy instances in response to demand.
When an application is instantiated on a Cloudlet, the user may override the default Flavor of the application. This allows for an instance in one location to be provided more resources than an instance in other locations, in expectation of different demands in different locations.
Many of the fields here are inherited from the App definition. Some are derived, like the mapped ports field, depending upon if the AppInst accessibility is via a shared or dedicated load balancer.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | Unique identifier key |
| cloudlet_loc | [distributed_match_engine.Loc](#distributed_match_engine.Loc) |  | Cached location of the cloudlet |
| uri | [string](#string) |  | URI to connect to this instance |
| cluster_inst_key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Cluster instance on which this is instatiated (not specifiable by user) |
| liveness | [Liveness](#edgeproto.Liveness) |  | Liveness of instance (see Liveness) |
| image_path | [string](#string) |  | Path to image to be able to download image |
| image_type | [ImageType](#edgeproto.ImageType) |  | Image type (see ImageType) |
| mapped_ports | [AppPort](#edgeproto.AppPort) | repeated | For instances accessible via a shared load balancer, defines the external ports on the shared load balancer that map to the internal ports External ports should be appended to the Uri for L4 access. |
| mapped_path | [string](#string) |  | Mapped path to append to Uri for public access. Only valid for L7 access types. |
| config_map | [string](#string) |  | Initial config passed to docker |
| flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Flavor defining resource requirements |
| access_layer | [AccessLayer](#edgeproto.AccessLayer) |  | Access layer(s) |






<a name="edgeproto.AppInstInfo"/>

### AppInstInfo
AppInstInfo provides information from the Cloudlet Resource Manager about the state of the AppInst on the Cloudlet. Whereas the AppInst defines the intent of instantiating an App on a Cloudlet, the AppInstInfo defines the current state of trying to apply that intent on the physical resources of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| state | [AppState](#edgeproto.AppState) |  | Current state of the AppInst on the Cloudlet |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the AppInst on the Cloudlet |






<a name="edgeproto.AppInstKey"/>

### AppInstKey
AppInstKey uniquely identifies an Application Instance (AppInst) or Application Instance state (AppInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_key | [AppKey](#edgeproto.AppKey) |  | App key |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which the App is instantiated |
| id | [fixed64](#fixed64) |  | Instance id for defining multiple instances of the same App on the same Cloudlet (not supported yet) |






<a name="edgeproto.AppInstMetrics"/>

### AppInstMetrics
(TODO) AppInstMetrics provide metrics collected about the application instance on the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| something | [uint64](#uint64) |  | what goes here? Note that metrics for grpc calls can be done by a prometheus interceptor in grpc, so adding call metrics here may be redundant unless they&#39;re needed for billing. |






<a name="edgeproto.AppPort"/>

### AppPort
AppPort describes an L4 public access port mapping. This is used to track external to internal mappings for access via a shared load balancer or reverse proxy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proto | [L4Proto](#edgeproto.L4Proto) |  | TCP or UDP protocol |
| internal_port | [int32](#int32) |  | Container port |
| public_port | [int32](#int32) |  | Public facing port (may be mapped on shared LB reverse proxy) |





 


<a name="edgeproto.AppState"/>

### AppState
AppState defines the state of the AppInst. This state is defined both by the state on the Controller, and the state on the Cloudlet where the AppInst is instantiated. Some of the states are intermediate states to denote a change in progress.

| Name | Number | Description |
| ---- | ------ | ----------- |
| AppStateUnknown | 0 | AppInst state unknown |
| AppStateBuilding | 1 | AppInst state in the process of being created |
| AppStateReady | 2 | AppInst state created and ready |
| AppStateErrors | 3 | AppInst change encountered errors, see Errors field of AppInstInfo |
| AppStateDeleting | 4 | AppInst in the process of being deleted |
| AppStateDeleted | 5 | AppInst was deleted |
| AppStateChanging | 6 | AppInst in the process of being updated |
| AppStateNotPresent | 7 | AppInst is not present |


 

 


<a name="edgeproto.AppInstApi"/>

### AppInstApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.AppInst) | Create an application instance |
| DeleteAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.AppInst) | Delete an application instance |
| UpdateAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.AppInst) | Update an application instance |
| ShowAppInst | [AppInst](#edgeproto.AppInst) | [AppInst](#edgeproto.AppInst) | Show application instances. Any fields specified will be used to filter results. |


<a name="edgeproto.AppInstInfoApi"/>

### AppInstInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAppInstInfo | [AppInstInfo](#edgeproto.AppInstInfo) | [AppInstInfo](#edgeproto.AppInstInfo) | Show application instances state. |


<a name="edgeproto.AppInstMetricsApi"/>

### AppInstMetricsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAppInstMetrics | [AppInstMetrics](#edgeproto.AppInstMetrics) | [AppInstMetrics](#edgeproto.AppInstMetrics) | Show application instance metrics. |

 



<a name="cloud-resource-manager.proto"/>
<p align="right"><a href="#top">Top</a></p>

## cloud-resource-manager.proto



<a name="edgeproto.CloudResource"/>

### CloudResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| category | [CloudResourceCategory](#edgeproto.CloudResourceCategory) |  |  |
| cloudletKey | [CloudletKey](#edgeproto.CloudletKey) |  |  |
| active | [bool](#bool) |  |  |
| id | [int32](#int32) |  |  |
| access_ip | [bytes](#bytes) |  | AccessIp should come from the cloudlet, but for testing it is configurable here. This will need to be removed later. |






<a name="edgeproto.EdgeCloudApp"/>

### EdgeCloudApp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| repository | [string](#string) |  |  |
| image | [string](#string) |  |  |
| cpu | [string](#string) |  |  |
| memory | [string](#string) |  |  |
| limitfactor | [int32](#int32) |  |  |
| exposure | [string](#string) |  |  |
| replicas | [int32](#int32) |  |  |
| context | [string](#string) |  |  |
| namespace | [string](#string) |  |  |
| region | [string](#string) |  |  |
| flavor | [string](#string) |  |  |
| network | [string](#string) |  |  |
| appInstKey | [AppInstKey](#edgeproto.AppInstKey) |  |  |






<a name="edgeproto.EdgeCloudApplication"/>

### EdgeCloudApplication



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manifest | [string](#string) |  |  |
| kind | [string](#string) |  |  |
| apps | [EdgeCloudApp](#edgeproto.EdgeCloudApp) | repeated |  |





 


<a name="edgeproto.CloudResourceCategory"/>

### CloudResourceCategory


| Name | Number | Description |
| ---- | ------ | ----------- |
| AllCloudResources | 0 |  |
| Kubernetes | 200 |  |
| k8s | 200 |  |
| Mesos | 201 |  |
| AWS | 202 |  |
| GCP | 203 |  |
| Azure | 204 |  |
| DigitalOcean | 205 |  |
| PacketNet | 206 |  |
| OpenStack | 300 |  |
| Docker | 301 |  |
| EKS | 400 |  |
| AKS | 402 |  |
| GKS | 403 |  |


 

 


<a name="edgeproto.CloudResourceManager"/>

### CloudResourceManager


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListCloudResource | [CloudResource](#edgeproto.CloudResource) | [CloudResource](#edgeproto.CloudResource) |  |
| AddCloudResource | [CloudResource](#edgeproto.CloudResource) | [Result](#edgeproto.CloudResource) |  |
| DeleteCloudResource | [CloudResource](#edgeproto.CloudResource) | [Result](#edgeproto.CloudResource) |  |
| DeployApplication | [EdgeCloudApplication](#edgeproto.EdgeCloudApplication) | [Result](#edgeproto.EdgeCloudApplication) |  |
| DeleteApplication | [EdgeCloudApplication](#edgeproto.EdgeCloudApplication) | [Result](#edgeproto.EdgeCloudApplication) |  |

 



<a name="cloudlet.proto"/>
<p align="right"><a href="#top">Top</a></p>

## cloudlet.proto



<a name="edgeproto.Cloudlet"/>

### Cloudlet
A Cloudlet is a set of compute resources at a particular location, typically an Operator&#39;s regional data center, or a cell tower. The Cloudlet is managed by a Cloudlet Resource Manager, which communicates with the Mobiledgex Controller and allows AppInsts (application instances) to be instantiated on the Cloudlet.
A Cloudlet will be created by either a Mobiledgex admin or an Operator that provides the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Unique identifier key |
| access_uri | [string](#string) |  | URI to use to connect to and create and administer the Cloudlet. This is not the URI for applications clients to access their back-end instances. |
| location | [distributed_match_engine.Loc](#distributed_match_engine.Loc) |  | Location of the Cloudlet site |
| ip_support | [IpSupport](#edgeproto.IpSupport) |  | Type of IP support provided by Cloudlet (see IpSupport) |
| static_ips | [string](#string) |  | List of static IPs for static IP support |
| num_dynamic_ips | [int32](#int32) |  | Number of dynamic IPs available for dynamic IP support

Certs for accessing cloudlet site |






<a name="edgeproto.CloudletInfo"/>

### CloudletInfo
CloudletInfo provides information from the Cloudlet Resource Manager about the state of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Unique identifier key |
| state | [CloudletState](#edgeproto.CloudletState) |  | State of cloudlet |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| os_max_ram | [uint64](#uint64) |  | Maximum Ram in MB on the Cloudlet |
| os_max_vcores | [uint64](#uint64) |  | Maximum number of VCPU cores on the Cloudlet |
| os_max_vol_gb | [uint64](#uint64) |  | Maximum amount of disk in GB on the Cloudlet |
| errors | [string](#string) | repeated | Any errors encountered while making changes to the Cloudlet |






<a name="edgeproto.CloudletKey"/>

### CloudletKey
CloudletKey uniquely identifies a Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operator_key | [OperatorKey](#edgeproto.OperatorKey) |  | Operator of the cloudlet site |
| name | [string](#string) |  | Name of the cloudlet |






<a name="edgeproto.CloudletMetrics"/>

### CloudletMetrics
(TODO) CloudletMetrics provide metrics collected about the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| foo | [uint64](#uint64) |  | what goes here? |





 


<a name="edgeproto.CloudletState"/>

### CloudletState
CloudletState is the state of the Cloudlet.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CloudletStateUnknown | 0 | Unknown |
| CloudletStateErrors | 1 | Create/Delete/Update encountered errors (see Errors field of CloudletInfo) |
| CloudletStateReady | 2 | Cloudlet is created and ready |
| CloudletStateOffline | 3 | Cloudlet is offline (unreachable) |
| CloudletStateNotPresent | 4 | Cloudlet is not present |


 

 


<a name="edgeproto.CloudletApi"/>

### CloudletApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Cloudlet) | Create a Cloudlet |
| DeleteCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Cloudlet) | Delete a Cloudlet |
| UpdateCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Cloudlet) | Update a Cloudlet |
| ShowCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Cloudlet](#edgeproto.Cloudlet) | Show Cloudlets |


<a name="edgeproto.CloudletInfoApi"/>

### CloudletInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowCloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) | [CloudletInfo](#edgeproto.CloudletInfo) | Show CloudletInfos |
| InjectCloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) | [Result](#edgeproto.CloudletInfo) | Inject (create) a CloudletInfo for regression testing |
| EvictCloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) | [Result](#edgeproto.CloudletInfo) | Evict (delete) a CloudletInfo for regression testing |


<a name="edgeproto.CloudletMetricsApi"/>

### CloudletMetricsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowCloudletMetrics | [CloudletMetrics](#edgeproto.CloudletMetrics) | [CloudletMetrics](#edgeproto.CloudletMetrics) | Show Cloudlet metrics |

 



<a name="cluster.proto"/>
<p align="right"><a href="#top">Top</a></p>

## cluster.proto



<a name="edgeproto.Cluster"/>

### Cluster
Clusters define a set of resources that are provided to one or more Apps tied to the cluster. The set of resources is defined by the Cluster flavor. The Cluster definition here is analogous to a Kubernetes cluster.
Like Apps, a Cluster is merely a definition, but is not instantiated on any Cloudlets. ClusterInsts are Clusters instantiated on a particular Cloudlet.
In comparison to ClusterFlavors which are fairly static and controller by administrators, Clusters are much more dynamic and created and deleted by the user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterKey](#edgeproto.ClusterKey) |  | Unique key |
| default_flavor | [ClusterFlavorKey](#edgeproto.ClusterFlavorKey) |  | Default flavor of the Cluster, may be overridden on the ClusterInst |
| auto | [bool](#bool) |  | Auto is set to true when automatically created by back-end (internal use only) |






<a name="edgeproto.ClusterKey"/>

### ClusterKey
ClusterKey uniquely identifies a Cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Cluster name |





 

 

 


<a name="edgeproto.ClusterApi"/>

### ClusterApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCluster | [Cluster](#edgeproto.Cluster) | [Result](#edgeproto.Cluster) | Create a Cluster |
| DeleteCluster | [Cluster](#edgeproto.Cluster) | [Result](#edgeproto.Cluster) | Delete a Cluster |
| UpdateCluster | [Cluster](#edgeproto.Cluster) | [Result](#edgeproto.Cluster) | Update a Cluster |
| ShowCluster | [Cluster](#edgeproto.Cluster) | [Cluster](#edgeproto.Cluster) | Show Clusters |

 



<a name="clusterflavor.proto"/>
<p align="right"><a href="#top">Top</a></p>

## clusterflavor.proto



<a name="edgeproto.ClusterFlavor"/>

### ClusterFlavor
ClusterFlavor defines a set of resources for a Cluster. ClusterFlavors should be fairly static objects that are almost never changed, and are only modified by Mobiledgex administrators.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterFlavorKey](#edgeproto.ClusterFlavorKey) |  | Unique key |
| node_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Flavor of each node in the Cluster |
| master_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Flavor of each master node in the Cluster |
| num_nodes | [uint32](#uint32) |  | Initial number of nodes in the Cluster |
| max_nodes | [uint32](#uint32) |  | Maximum number of nodes allowed in the Cluster (for auto-scaling) |
| num_masters | [uint32](#uint32) |  | Number of master nodes in the Cluster |






<a name="edgeproto.ClusterFlavorKey"/>

### ClusterFlavorKey
ClusterFlavorKey uniquely identifies a Cluster Flavor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |





 

 

 


<a name="edgeproto.ClusterFlavorApi"/>

### ClusterFlavorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [Result](#edgeproto.ClusterFlavor) | Create a ClusterFlavor |
| DeleteClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [Result](#edgeproto.ClusterFlavor) | Delete a ClusterFlavor |
| UpdateClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [Result](#edgeproto.ClusterFlavor) | Update a ClusterFlavor |
| ShowClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [ClusterFlavor](#edgeproto.ClusterFlavor) | Show ClusterFlavors |

 



<a name="clusterinst.proto"/>
<p align="right"><a href="#top">Top</a></p>

## clusterinst.proto



<a name="edgeproto.ClusterInst"/>

### ClusterInst
ClusterInst is an instance of a Cluster on a Cloudlet. It is defined by a Cluster plus a Cloudlet key. This separation of the definition of the Cluster versus its instance is unique to Mobiledgex, and allows the Developer to provide the Cluster definition, while either the Developer may statically define the instances, or the Mobiledgex platform may dynamically create and destroy instances in response to demand.
When a Cluster is instantiated on a Cloudlet, the user may override the default ClusterFlavor of the Cluster. This allows for an instance in one location to be provided more resources than an instance in other locations, in expectation of different demands in different locations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Unique key |
| flavor | [ClusterFlavorKey](#edgeproto.ClusterFlavorKey) |  | ClusterFlavor of the Cluster |
| liveness | [Liveness](#edgeproto.Liveness) |  | Liveness of instance (see Liveness) |
| auto | [bool](#bool) |  | Auto is set to true when automatically created by back-end (internal use only) |






<a name="edgeproto.ClusterInstInfo"/>

### ClusterInstInfo
ClusterInstInfo provides information from the Cloudlet Resource Manager about the state of the ClusterInst on the Cloudlet. Whereas the ClusterInst defines the intent of instantiating a Cluster on a Cloudlet, the ClusterInstInfo defines the current state of trying to apply that intent on the physical resources of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| state | [ClusterState](#edgeproto.ClusterState) |  | State of the cluster |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the ClusterInst on the Cloudlet. |






<a name="edgeproto.ClusterInstKey"/>

### ClusterInstKey
ClusterInstKey uniquely identifies a Cluster Instance (ClusterInst) or Cluster Instance state (ClusterInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_key | [ClusterKey](#edgeproto.ClusterKey) |  | Cluster key |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which the Cluster is instantiated |





 


<a name="edgeproto.ClusterState"/>

### ClusterState
ClusterState defines the state of the ClusterInst. This state is defined both by the state on the Controller, and the state on the Cloudlet where the ClusterInst is instantiated. Some of the states are intermediate states to denote a change in progress.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ClusterStateUnknown | 0 | ClusterInst state unknown |
| ClusterStateBuilding | 1 | ClusterInst state in the process of being created |
| ClusterStateReady | 2 | ClusterInst state created and ready |
| ClusterStateErrors | 3 | ClusterInst change encountered errors, see Errors field of ClusterInstInfo |
| ClusterStateDeleting | 4 | ClusterInst in the process of being deleted |
| ClusterStateDeleted | 5 | ClusterInst was deleted |
| ClusterStateChanging | 6 | ClusterInst in the process of being updated |
| ClusterStateNotPresent | 7 | ClusterInst is not present |


 

 


<a name="edgeproto.ClusterInstApi"/>

### ClusterInstApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.ClusterInst) | Create a Cluster instance |
| DeleteClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.ClusterInst) | Delete a Cluster instance |
| UpdateClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.ClusterInst) | Update a Cluster instance |
| ShowClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [ClusterInst](#edgeproto.ClusterInst) | Show Cluster instances |


<a name="edgeproto.ClusterInstInfoApi"/>

### ClusterInstInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowClusterInstInfo | [ClusterInstInfo](#edgeproto.ClusterInstInfo) | [ClusterInstInfo](#edgeproto.ClusterInstInfo) | Show Cluster instances state. |

 



<a name="common.proto"/>
<p align="right"><a href="#top">Top</a></p>

## common.proto


 


<a name="edgeproto.IpSupport"/>

### IpSupport
IpSupport indicates the type of public IP support provided by the Cloudlet. Static IP support indicates a set of static public IPs are available for use, and managed by the Controller. Dynamic indicates the Cloudlet uses a DHCP server to provide public IP addresses, and the controller has no control over which IPs are assigned.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IpSupportUnknown | 0 | Unknown IP support |
| IpSupportStatic | 1 | Static IP addresses are provided to and managed by Controller |
| IpSupportDynamic | 2 | IP addresses are dynamically provided by an Operator&#39;s DHCP server |



<a name="edgeproto.L4Proto"/>

### L4Proto
L4Proto indicates which L4 protocol to use for accessing an application on a particular port. This is required by Kubernetes for port mapping.

| Name | Number | Description |
| ---- | ------ | ----------- |
| L4ProtoUnknown | 0 | Unknown protocol |
| L4ProtoTCP | 1 | TCP protocol |
| L4ProtoUDP | 2 | UDP protocol |



<a name="edgeproto.Liveness"/>

### Liveness
Liveness indicates if an object was created statically via an external API call, or dynamically via an internal algorithm.

| Name | Number | Description |
| ---- | ------ | ----------- |
| LivenessUnknown | 0 | Unknown liveness |
| LivenessStatic | 1 | Object managed by external entity |
| LivenessDynamic | 2 | Object managed internally |


 

 

 



<a name="developer.proto"/>
<p align="right"><a href="#top">Top</a></p>

## developer.proto



<a name="edgeproto.Developer"/>

### Developer
A Developer defines a Mobiledgex customer that can create and manage applications, clusters, instances, etc. Applications and other objects created by one Developer cannot be seen or managed by other Developers. Billing will likely be done on a per-developer basis.
Creating a developer identity is likely the first step of (self-)registering a new customer.
TODO: user management, auth, etc is not implemented yet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [DeveloperKey](#edgeproto.DeveloperKey) |  | Unique identifier key |
| username | [string](#string) |  | Login name (TODO) |
| passhash | [string](#string) |  | Encrypted password (TODO) |
| address | [string](#string) |  | Physical address |
| email | [string](#string) |  | Contact email |






<a name="edgeproto.DeveloperKey"/>

### DeveloperKey
DeveloperKey uniquely identifies a Developer (Mobiledgex customer)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Organization or Company Name |





 

 

 


<a name="edgeproto.DeveloperApi"/>

### DeveloperApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateDeveloper | [Developer](#edgeproto.Developer) | [Result](#edgeproto.Developer) | Create a Developer |
| DeleteDeveloper | [Developer](#edgeproto.Developer) | [Result](#edgeproto.Developer) | Delete a Developer |
| UpdateDeveloper | [Developer](#edgeproto.Developer) | [Result](#edgeproto.Developer) | Update a Developer |
| ShowDeveloper | [Developer](#edgeproto.Developer) | [Developer](#edgeproto.Developer) | Show Developers |

 



<a name="flavor.proto"/>
<p align="right"><a href="#top">Top</a></p>

## flavor.proto



<a name="edgeproto.Flavor"/>

### Flavor
A Flavor identifies the Cpu, Ram, and Disk resources required for either a node in a Cluster, or an application instance. For a node in a cluster, these are the physical resources provided by that node. For an application instance, this defines the resources (per node) that should be allocated to the instance from the Cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [FlavorKey](#edgeproto.FlavorKey) |  | Unique key |
| ram | [uint64](#uint64) |  | RAM in MB |
| vcpus | [uint64](#uint64) |  | VCPU cores |
| disk | [uint64](#uint64) |  | Amount of disk in GB |






<a name="edgeproto.FlavorKey"/>

### FlavorKey
FlavorKey uniquely identifies a Flavor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |





 

 

 


<a name="edgeproto.FlavorApi"/>

### FlavorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Flavor) | Create a Flavor |
| DeleteFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Flavor) | Delete a Flavor |
| UpdateFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Flavor) | Update a Flavor |
| ShowFlavor | [Flavor](#edgeproto.Flavor) | [Flavor](#edgeproto.Flavor) | Show Flavors |

 



<a name="notice.proto"/>
<p align="right"><a href="#top">Top</a></p>

## notice.proto
Notice is the message used by the notify protocol to communicate and coordinate internally between different Mobiledgex services. For details on the notify protocol, see the &#34;MEX Cloud Service Interactions&#34; confluence article.
In general, the protocol is used to synchronize state from one service to another. The protocol is fairly symmetric, with different state being synchronized both from server to client and client to server.


<a name="edgeproto.NoticeReply"/>

### NoticeReply
NoticyReply is sent from server to client.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [NoticeAction](#edgeproto.NoticeAction) |  | Action to perform |
| version | [uint32](#uint32) |  | Protocol version supported by sender |
| appInst | [AppInst](#edgeproto.AppInst) |  |  |
| cloudlet | [Cloudlet](#edgeproto.Cloudlet) |  |  |
| flavor | [Flavor](#edgeproto.Flavor) |  |  |
| clusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) |  |  |
| clusterInst | [ClusterInst](#edgeproto.ClusterInst) |  |  |






<a name="edgeproto.NoticeRequest"/>

### NoticeRequest
NoticeRequest is sent from client to server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [NoticeAction](#edgeproto.NoticeAction) |  | Action to perform |
| version | [uint32](#uint32) |  | Protocol version supported by receiver |
| requestor | [NoticeRequestor](#edgeproto.NoticeRequestor) |  | Client requestor type |
| revision | [uint64](#uint64) |  | Revision of database |
| cloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) |  |  |
| appInstInfo | [AppInstInfo](#edgeproto.AppInstInfo) |  |  |
| clusterInstInfo | [ClusterInstInfo](#edgeproto.ClusterInstInfo) |  |  |





 


<a name="edgeproto.NoticeAction"/>

### NoticeAction
NoticeAction denotes what kind of action this notification is for.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 | No action |
| UPDATE | 1 | Update the object |
| DELETE | 2 | Delete the object |
| VERSION | 3 | Version exchange negotitation message |
| SENDALL_END | 4 | Initial send all finished message |



<a name="edgeproto.NoticeRequestor"/>

### NoticeRequestor
NoticeRequestor indicates which type of service the client is.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NoticeRequestorNone | 0 | Invalid |
| NoticeRequestorDME | 1 | Distributed Matching Engine |
| NoticeRequestorCRM | 2 | Cloudlet Resource Manager |


 

 


<a name="edgeproto.NotifyApi"/>

### NotifyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| StreamNotice | [NoticeRequest](#edgeproto.NoticeRequest) | [NoticeReply](#edgeproto.NoticeRequest) | Bidrectional stream for exchanging data between controller and DME/CRM |

 



<a name="operator.proto"/>
<p align="right"><a href="#top">Top</a></p>

## operator.proto



<a name="edgeproto.Operator"/>

### Operator
An Operator defines a telecommunications provider such as AT&amp;T, xmobx, etc. The operators in turn provide Mobiledgex with compute resource Cloudlets that serve as the basis for location-based services.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [OperatorKey](#edgeproto.OperatorKey) |  | Unique identifier key |






<a name="edgeproto.OperatorKey"/>

### OperatorKey
OperatorKey uniquely identifies an Operator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Company or Organization name of the operator |





 

 

 


<a name="edgeproto.OperatorApi"/>

### OperatorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateOperator | [Operator](#edgeproto.Operator) | [Result](#edgeproto.Operator) | Create an Operator |
| DeleteOperator | [Operator](#edgeproto.Operator) | [Result](#edgeproto.Operator) | Delete an Operator |
| UpdateOperator | [Operator](#edgeproto.Operator) | [Result](#edgeproto.Operator) | Update an Operator |
| ShowOperator | [Operator](#edgeproto.Operator) | [Operator](#edgeproto.Operator) | Show Operators |

 



<a name="refs.proto"/>
<p align="right"><a href="#top">Top</a></p>

## refs.proto



<a name="edgeproto.CloudletRefs"/>

### CloudletRefs
CloudletRefs track used resources and Clusters instantiated on a Cloudlet. Used resources are compared against max resources for a Cloudlet to determine if resources are available for a new Cluster to be instantiated on the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet key |
| clusters | [ClusterKey](#edgeproto.ClusterKey) | repeated | Clusters instantiated on the Cloudlet |
| used_ram | [uint64](#uint64) |  | Used RAM in MB |
| used_vcores | [uint64](#uint64) |  | Used VCPU cores |
| used_disk | [uint64](#uint64) |  | Used disk in GB |
| root_lb_ports | [CloudletRefs.RootLbPortsEntry](#edgeproto.CloudletRefs.RootLbPortsEntry) | repeated | Used ports on root load balancer. Map key is public port, value is unused. |






<a name="edgeproto.CloudletRefs.RootLbPortsEntry"/>

### CloudletRefs.RootLbPortsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [int32](#int32) |  |  |
| value | [int32](#int32) |  |  |






<a name="edgeproto.ClusterRefs"/>

### ClusterRefs
ClusterRefs track used resources within a ClusterInst. Each AppInst specifies a set of required resources (Flavor), so tracking resources used by Apps within a Cluster is necessary to determine if enough resources are available for another AppInst to be instantiated on a ClusterInst.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Cluster Instance key |
| apps | [AppKey](#edgeproto.AppKey) | repeated | Apps instances in the Cluster Instance |
| used_ram | [uint64](#uint64) |  | Used RAM in MB |
| used_vcores | [uint64](#uint64) |  | Used VCPU cores |
| used_disk | [uint64](#uint64) |  | Used disk in GB |





 

 

 


<a name="edgeproto.CloudletRefsApi"/>

### CloudletRefsApi
This API should be admin-only

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowCloudletRefs | [CloudletRefs](#edgeproto.CloudletRefs) | [CloudletRefs](#edgeproto.CloudletRefs) | Show CloudletRefs (debug only) |


<a name="edgeproto.ClusterRefsApi"/>

### ClusterRefsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowClusterRefs | [ClusterRefs](#edgeproto.ClusterRefs) | [ClusterRefs](#edgeproto.ClusterRefs) | Show ClusterRefs (debug only) |

 



<a name="result.proto"/>
<p align="right"><a href="#top">Top</a></p>

## result.proto



<a name="edgeproto.Result"/>

### Result
Result is a generic object for returning the result of an API call. In general, result is not used. The error value returned by the GRPC API call is used instead.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  | Message, may be success or failure message |
| code | [int32](#int32) |  | Error code, 0 indicates success, non-zero indicates failure (not implemented) |





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <a name="double" /> double |  | double | double | float |
| <a name="float" /> float |  | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |

