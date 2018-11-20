# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [app.proto](#app.proto)
    - [App](#edgeproto.App)
    - [AppKey](#edgeproto.AppKey)
  
    - [ImageType](#edgeproto.ImageType)
  
  
    - [AppApi](#edgeproto.AppApi)
  

- [app_inst.proto](#app_inst.proto)
    - [AppInst](#edgeproto.AppInst)
    - [AppInstInfo](#edgeproto.AppInstInfo)
    - [AppInstKey](#edgeproto.AppInstKey)
    - [AppInstMetrics](#edgeproto.AppInstMetrics)
  
  
  
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
  
  
  
    - [ClusterInstApi](#edgeproto.ClusterInstApi)
    - [ClusterInstInfoApi](#edgeproto.ClusterInstInfoApi)
  

- [common.proto](#common.proto)
  
    - [CRMOverride](#edgeproto.CRMOverride)
    - [IpAccess](#edgeproto.IpAccess)
    - [IpSupport](#edgeproto.IpSupport)
    - [Liveness](#edgeproto.Liveness)
    - [TrackedState](#edgeproto.TrackedState)
  
  
  

- [controller.proto](#controller.proto)
    - [Controller](#edgeproto.Controller)
    - [ControllerKey](#edgeproto.ControllerKey)
  
  
  
    - [ControllerApi](#edgeproto.ControllerApi)
  

- [developer.proto](#developer.proto)
    - [Developer](#edgeproto.Developer)
    - [DeveloperKey](#edgeproto.DeveloperKey)
  
  
  
    - [DeveloperApi](#edgeproto.DeveloperApi)
  

- [flavor.proto](#flavor.proto)
    - [Flavor](#edgeproto.Flavor)
    - [FlavorKey](#edgeproto.FlavorKey)
  
  
  
    - [FlavorApi](#edgeproto.FlavorApi)
  

- [metric.proto](#metric.proto)
    - [Metric](#edgeproto.Metric)
    - [MetricTag](#edgeproto.MetricTag)
    - [MetricVal](#edgeproto.MetricVal)
  
  
  
  

- [node.proto](#node.proto)
    - [Node](#edgeproto.Node)
    - [NodeKey](#edgeproto.NodeKey)
  
    - [NodeType](#edgeproto.NodeType)
  
  
    - [NodeApi](#edgeproto.NodeApi)
  

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



<a name="app.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## app.proto



<a name="edgeproto.App"></a>

### App
Apps are applications that may be instantiated on Cloudlets, providing a back-end service to an application client (using the mobiledgex SDK) running on a user device such as a cell phone, wearable, drone, etc. Applications belong to Developers, and must specify their image and accessibility. Applications are analagous to Pods in Kubernetes, and similarly are tied to a Cluster.
An application in itself is not tied to a Cloudlet, but provides a definition that can be used to instantiate it on a Cloudlet. AppInsts are applications instantiated on a particular Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppKey](#edgeproto.AppKey) |  | Unique identifier key |
| image_path | [string](#string) |  | URI from which to download image |
| image_type | [ImageType](#edgeproto.ImageType) |  | Image type (see ImageType) |
| ip_access | [IpAccess](#edgeproto.IpAccess) |  | IP access type |
| access_ports | [string](#string) |  | For Layer4 access, the ports the app listens on. This is a comma separated list of protocol:port pairs, i.e. tcp:80,http:443,udp:10002. Only tcp, udp, and http protocols are supported; tcp and udp are assumed to be L4, and http is assumed to be L7 access. |
| config | [string](#string) |  | URI of resource to be used to establish config for App. |
| default_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Default flavor for the App, may be overridden by the AppInst |
| cluster | [ClusterKey](#edgeproto.ClusterKey) |  | Cluster on which the App can be instantiated. If not specified during create, a cluster will be automatically created. If specified, it must exist. |
| app_template | [string](#string) |  | Template of kubernetes deployment yaml. Who/What sets this is TDB, but it should not be directly exposed to the user, because we do not want to expose kubernetes to the user. However, because we currently don&#39;t have any other way to set it, for flexibility, for now it is exposed to the user. |
| auth_public_key | [string](#string) |  | public key used for authentication |
| command | [string](#string) |  | Command to start service |
| annotations | [string](#string) |  | Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=&#34;val 3&#34; |
| deployment | [string](#string) |  | Deployment target (kubernetes, docker, kvm, etc) |
| deployment_manifest | [string](#string) |  | Deployment manifest is the deployment specific manifest file/config |
| deployment_generator | [string](#string) |  | Deployment generator target |
| android_package_name | [string](#string) |  | Android package name, optional |
| permits_platform_apps | [bool](#bool) |  | Indicates whether or not platform apps are allowed to perform actions on behalf of this app, such as FindCloudlet |






<a name="edgeproto.AppKey"></a>

### AppKey
AppKey uniquely identifies an Application.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| developer_key | [DeveloperKey](#edgeproto.DeveloperKey) |  | Developer key |
| name | [string](#string) |  | Application name |
| version | [string](#string) |  | Version of the app |





 


<a name="edgeproto.ImageType"></a>

### ImageType
ImageType specifies the image type of the application.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ImageTypeUnknown | 0 | Unknown image type |
| ImageTypeDocker | 1 | Docker container image type compatible with Kubernetes |
| ImageTypeQCOW | 2 | QCOW2 virtual machine image type |


 

 


<a name="edgeproto.AppApi"></a>

### AppApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateApp | [App](#edgeproto.App) | [Result](#edgeproto.Result) | Create an application |
| DeleteApp | [App](#edgeproto.App) | [Result](#edgeproto.Result) | Delete an application |
| UpdateApp | [App](#edgeproto.App) | [Result](#edgeproto.Result) | Update an application |
| ShowApp | [App](#edgeproto.App) | [App](#edgeproto.App) stream | Show applications. Any fields specified will be used to filter results. |

 



<a name="app_inst.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## app_inst.proto



<a name="edgeproto.AppInst"></a>

### AppInst
AppInst is an instance of an App (application) on a Cloudlet. It is defined by an App plus a Cloudlet key. This separation of the definition of the App versus its instantiation is unique to Mobiledgex, and allows the Developer to provide the App defintion, while either the Developer may statically define the instances, or the Mobiledgex platform may dynamically create and destroy instances in response to demand.
When an application is instantiated on a Cloudlet, the user may override the default Flavor of the application. This allows for an instance in one location to be provided more resources than an instance in other locations, in expectation of different demands in different locations.
Many of the fields here are inherited from the App definition. Some are derived, like the mapped ports field, depending upon if the AppInst accessibility is via a shared or dedicated load balancer.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | Unique identifier key |
| cloudlet_loc | [distributed_match_engine.Loc](#distributed_match_engine.Loc) |  | Cached location of the cloudlet |
| uri | [string](#string) |  | Base FQDN (not really URI) for the App. See Service FQDN for endpoint access. |
| cluster_inst_key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Cluster instance on which this is instatiated (not specifiable by user) |
| liveness | [Liveness](#edgeproto.Liveness) |  | Liveness of instance (see Liveness) |
| mapped_ports | [distributed_match_engine.AppPort](#distributed_match_engine.AppPort) | repeated | For instances accessible via a shared load balancer, defines the external ports on the shared load balancer that map to the internal ports External ports should be appended to the Uri for L4 access. |
| flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Flavor defining resource requirements |
| ip_access | [IpAccess](#edgeproto.IpAccess) |  | IP access type. If set to SharedOrDedicated on App, one will be chosen for AppInst |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the AppInst on the Cloudlet |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the AppInst on the Cloudlet |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |
| allocated_ip | [string](#string) |  | allocated IP for dedicated access |






<a name="edgeproto.AppInstInfo"></a>

### AppInstInfo
AppInstInfo provides information from the Cloudlet Resource Manager about the state of the AppInst on the Cloudlet. Whereas the AppInst defines the intent of instantiating an App on a Cloudlet, the AppInstInfo defines the current state of trying to apply that intent on the physical resources of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the AppInst on the Cloudlet |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the AppInst on the Cloudlet |






<a name="edgeproto.AppInstKey"></a>

### AppInstKey
AppInstKey uniquely identifies an Application Instance (AppInst) or Application Instance state (AppInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_key | [AppKey](#edgeproto.AppKey) |  | App key |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which the App is instantiated |
| id | [fixed64](#fixed64) |  | Instance id for defining multiple instances of the same App on the same Cloudlet (not supported yet) |






<a name="edgeproto.AppInstMetrics"></a>

### AppInstMetrics
(TODO) AppInstMetrics provide metrics collected about the application instance on the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| something | [uint64](#uint64) |  | what goes here? Note that metrics for grpc calls can be done by a prometheus interceptor in grpc, so adding call metrics here may be redundant unless they&#39;re needed for billing. |





 

 

 


<a name="edgeproto.AppInstApi"></a>

### AppInstApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Create an application instance |
| DeleteAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Delete an application instance |
| UpdateAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Update an application instance |
| ShowAppInst | [AppInst](#edgeproto.AppInst) | [AppInst](#edgeproto.AppInst) stream | Show application instances. Any fields specified will be used to filter results. |


<a name="edgeproto.AppInstInfoApi"></a>

### AppInstInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAppInstInfo | [AppInstInfo](#edgeproto.AppInstInfo) | [AppInstInfo](#edgeproto.AppInstInfo) stream | Show application instances state. |


<a name="edgeproto.AppInstMetricsApi"></a>

### AppInstMetricsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAppInstMetrics | [AppInstMetrics](#edgeproto.AppInstMetrics) | [AppInstMetrics](#edgeproto.AppInstMetrics) stream | Show application instance metrics. |

 



<a name="cloud-resource-manager.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cloud-resource-manager.proto



<a name="edgeproto.CloudResource"></a>

### CloudResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| category | [CloudResourceCategory](#edgeproto.CloudResourceCategory) |  |  |
| cloudletKey | [CloudletKey](#edgeproto.CloudletKey) |  |  |
| active | [bool](#bool) |  |  |
| id | [int32](#int32) |  |  |
| access_ip | [bytes](#bytes) |  | AccessIp should come from the cloudlet, but for testing it is configurable here. This will need to be removed later. |






<a name="edgeproto.EdgeCloudApp"></a>

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






<a name="edgeproto.EdgeCloudApplication"></a>

### EdgeCloudApplication



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manifest | [string](#string) |  |  |
| kind | [string](#string) |  |  |
| apps | [EdgeCloudApp](#edgeproto.EdgeCloudApp) | repeated |  |





 


<a name="edgeproto.CloudResourceCategory"></a>

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


 

 


<a name="edgeproto.CloudResourceManager"></a>

### CloudResourceManager


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListCloudResource | [CloudResource](#edgeproto.CloudResource) | [CloudResource](#edgeproto.CloudResource) stream |  |
| AddCloudResource | [CloudResource](#edgeproto.CloudResource) | [Result](#edgeproto.Result) |  |
| DeleteCloudResource | [CloudResource](#edgeproto.CloudResource) | [Result](#edgeproto.Result) |  |
| DeployApplication | [EdgeCloudApplication](#edgeproto.EdgeCloudApplication) | [Result](#edgeproto.Result) |  |
| DeleteApplication | [EdgeCloudApplication](#edgeproto.EdgeCloudApplication) | [Result](#edgeproto.Result) |  |

 



<a name="cloudlet.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cloudlet.proto



<a name="edgeproto.Cloudlet"></a>

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






<a name="edgeproto.CloudletInfo"></a>

### CloudletInfo
CloudletInfo provides information from the Cloudlet Resource Manager about the state of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Unique identifier key |
| state | [CloudletState](#edgeproto.CloudletState) |  | State of cloudlet |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| controller | [string](#string) |  | Connected controller unique id |
| os_max_ram | [uint64](#uint64) |  | Maximum Ram in MB on the Cloudlet |
| os_max_vcores | [uint64](#uint64) |  | Maximum number of VCPU cores on the Cloudlet |
| os_max_vol_gb | [uint64](#uint64) |  | Maximum amount of disk in GB on the Cloudlet |
| errors | [string](#string) | repeated | Any errors encountered while making changes to the Cloudlet |






<a name="edgeproto.CloudletKey"></a>

### CloudletKey
CloudletKey uniquely identifies a Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operator_key | [OperatorKey](#edgeproto.OperatorKey) |  | Operator of the cloudlet site |
| name | [string](#string) |  | Name of the cloudlet |






<a name="edgeproto.CloudletMetrics"></a>

### CloudletMetrics
(TODO) CloudletMetrics provide metrics collected about the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| foo | [uint64](#uint64) |  | what goes here? |





 


<a name="edgeproto.CloudletState"></a>

### CloudletState
CloudletState is the state of the Cloudlet.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CloudletStateUnknown | 0 | Unknown |
| CloudletStateErrors | 1 | Create/Delete/Update encountered errors (see Errors field of CloudletInfo) |
| CloudletStateReady | 2 | Cloudlet is created and ready |
| CloudletStateOffline | 3 | Cloudlet is offline (unreachable) |
| CloudletStateNotPresent | 4 | Cloudlet is not present |


 

 


<a name="edgeproto.CloudletApi"></a>

### CloudletApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Result) stream | Create a Cloudlet |
| DeleteCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Result) stream | Delete a Cloudlet |
| UpdateCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Result) stream | Update a Cloudlet |
| ShowCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Cloudlet](#edgeproto.Cloudlet) stream | Show Cloudlets |


<a name="edgeproto.CloudletInfoApi"></a>

### CloudletInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowCloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) | [CloudletInfo](#edgeproto.CloudletInfo) stream | Show CloudletInfos |
| InjectCloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) | [Result](#edgeproto.Result) | Inject (create) a CloudletInfo for regression testing |
| EvictCloudletInfo | [CloudletInfo](#edgeproto.CloudletInfo) | [Result](#edgeproto.Result) | Evict (delete) a CloudletInfo for regression testing |


<a name="edgeproto.CloudletMetricsApi"></a>

### CloudletMetricsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowCloudletMetrics | [CloudletMetrics](#edgeproto.CloudletMetrics) | [CloudletMetrics](#edgeproto.CloudletMetrics) stream | Show Cloudlet metrics |

 



<a name="cluster.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cluster.proto



<a name="edgeproto.Cluster"></a>

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






<a name="edgeproto.ClusterKey"></a>

### ClusterKey
ClusterKey uniquely identifies a Cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Cluster name |





 

 

 


<a name="edgeproto.ClusterApi"></a>

### ClusterApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCluster | [Cluster](#edgeproto.Cluster) | [Result](#edgeproto.Result) | Create a Cluster |
| DeleteCluster | [Cluster](#edgeproto.Cluster) | [Result](#edgeproto.Result) | Delete a Cluster |
| UpdateCluster | [Cluster](#edgeproto.Cluster) | [Result](#edgeproto.Result) | Update a Cluster |
| ShowCluster | [Cluster](#edgeproto.Cluster) | [Cluster](#edgeproto.Cluster) stream | Show Clusters |

 



<a name="clusterflavor.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## clusterflavor.proto



<a name="edgeproto.ClusterFlavor"></a>

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






<a name="edgeproto.ClusterFlavorKey"></a>

### ClusterFlavorKey
ClusterFlavorKey uniquely identifies a Cluster Flavor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |





 

 

 


<a name="edgeproto.ClusterFlavorApi"></a>

### ClusterFlavorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [Result](#edgeproto.Result) | Create a ClusterFlavor |
| DeleteClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [Result](#edgeproto.Result) | Delete a ClusterFlavor |
| UpdateClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [Result](#edgeproto.Result) | Update a ClusterFlavor |
| ShowClusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) | [ClusterFlavor](#edgeproto.ClusterFlavor) stream | Show ClusterFlavors |

 



<a name="clusterinst.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## clusterinst.proto



<a name="edgeproto.ClusterInst"></a>

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
| state | [TrackedState](#edgeproto.TrackedState) |  | State of the cluster instance |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the ClusterInst on the Cloudlet. |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |






<a name="edgeproto.ClusterInstInfo"></a>

### ClusterInstInfo
ClusterInstInfo provides information from the Cloudlet Resource Manager about the state of the ClusterInst on the Cloudlet. Whereas the ClusterInst defines the intent of instantiating a Cluster on a Cloudlet, the ClusterInstInfo defines the current state of trying to apply that intent on the physical resources of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| state | [TrackedState](#edgeproto.TrackedState) |  | State of the cluster instance |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the ClusterInst on the Cloudlet. |






<a name="edgeproto.ClusterInstKey"></a>

### ClusterInstKey
ClusterInstKey uniquely identifies a Cluster Instance (ClusterInst) or Cluster Instance state (ClusterInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_key | [ClusterKey](#edgeproto.ClusterKey) |  | Cluster key |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which the Cluster is instantiated |





 

 

 


<a name="edgeproto.ClusterInstApi"></a>

### ClusterInstApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.Result) stream | Create a Cluster instance |
| DeleteClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.Result) stream | Delete a Cluster instance |
| UpdateClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.Result) stream | Update a Cluster instance |
| ShowClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [ClusterInst](#edgeproto.ClusterInst) stream | Show Cluster instances |


<a name="edgeproto.ClusterInstInfoApi"></a>

### ClusterInstInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowClusterInstInfo | [ClusterInstInfo](#edgeproto.ClusterInstInfo) | [ClusterInstInfo](#edgeproto.ClusterInstInfo) stream | Show Cluster instances state. |

 



<a name="common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## common.proto


 


<a name="edgeproto.CRMOverride"></a>

### CRMOverride
CRMOverride can be applied to commands that issue requests to the CRM.
It should only be used by administrators when bugs have caused the
Controller and CRM to get out of sync. It allows commands from the
Controller to ignore errors from the CRM, or ignore the CRM completely
(messages will not be sent to CRM).

| Name | Number | Description |
| ---- | ------ | ----------- |
| NoOverride | 0 | No override |
| IgnoreCRMErrors | 1 | Ignore errors from CRM |
| IgnoreCRM | 2 | Ignore CRM completely (does not inform CRM of operation) |
| IgnoreTransientState | 3 | Ignore Transient State (only admin should use if CRM crashed) |
| IgnoreCRMandTransientState | 4 | Ignore CRM and Transient State |



<a name="edgeproto.IpAccess"></a>

### IpAccess


| Name | Number | Description |
| ---- | ------ | ----------- |
| IpAccessUnknown | 0 | Unknown IP access |
| IpAccessDedicated | 1 | Dedicated IP access |
| IpAccessDedicatedOrShared | 2 | Dedicated or shared (prefers dedicated) access |
| IpAccessShared | 3 | Shared IP access |



<a name="edgeproto.IpSupport"></a>

### IpSupport
IpSupport indicates the type of public IP support provided by the Cloudlet. Static IP support indicates a set of static public IPs are available for use, and managed by the Controller. Dynamic indicates the Cloudlet uses a DHCP server to provide public IP addresses, and the controller has no control over which IPs are assigned.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IpSupportUnknown | 0 | Unknown IP support |
| IpSupportStatic | 1 | Static IP addresses are provided to and managed by Controller |
| IpSupportDynamic | 2 | IP addresses are dynamically provided by an Operator&#39;s DHCP server |



<a name="edgeproto.Liveness"></a>

### Liveness
Liveness indicates if an object was created statically via an external API call, or dynamically via an internal algorithm.

| Name | Number | Description |
| ---- | ------ | ----------- |
| LivenessUnknown | 0 | Unknown liveness |
| LivenessStatic | 1 | Object managed by external entity |
| LivenessDynamic | 2 | Object managed internally |



<a name="edgeproto.TrackedState"></a>

### TrackedState
TrackedState is used to track the state of an object on a remote node,
i.e. track the state of a ClusterInst object on the CRM (Cloudlet).

| Name | Number | Description |
| ---- | ------ | ----------- |
| TrackedStateUnknown | 0 | Unknown state |
| NotPresent | 1 | Not present (does not exist) |
| CreateRequested | 2 | Create requested |
| Creating | 3 | Creating |
| CreateError | 4 | Create error |
| Ready | 5 | Ready |
| UpdateRequested | 6 | Update requested |
| Updating | 7 | Updating |
| UpdateError | 8 | Update error |
| DeleteRequested | 9 | Delete requested |
| Deleting | 10 | Deleting |
| DeleteError | 11 | Delete error |


 

 

 



<a name="controller.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## controller.proto



<a name="edgeproto.Controller"></a>

### Controller
A Controller is a service that manages the edge-cloud data and controls other edge-cloud micro-services.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ControllerKey](#edgeproto.ControllerKey) |  | Unique identifier key |






<a name="edgeproto.ControllerKey"></a>

### ControllerKey
ControllerKey uniquely defines a Controller


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| addr | [string](#string) |  | external API address |





 

 

 


<a name="edgeproto.ControllerApi"></a>

### ControllerApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowController | [Controller](#edgeproto.Controller) | [Controller](#edgeproto.Controller) stream | Show Controllers |

 



<a name="developer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## developer.proto



<a name="edgeproto.Developer"></a>

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






<a name="edgeproto.DeveloperKey"></a>

### DeveloperKey
DeveloperKey uniquely identifies a Developer (Mobiledgex customer)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Organization or Company Name |





 

 

 


<a name="edgeproto.DeveloperApi"></a>

### DeveloperApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateDeveloper | [Developer](#edgeproto.Developer) | [Result](#edgeproto.Result) | Create a Developer |
| DeleteDeveloper | [Developer](#edgeproto.Developer) | [Result](#edgeproto.Result) | Delete a Developer |
| UpdateDeveloper | [Developer](#edgeproto.Developer) | [Result](#edgeproto.Result) | Update a Developer |
| ShowDeveloper | [Developer](#edgeproto.Developer) | [Developer](#edgeproto.Developer) stream | Show Developers |

 



<a name="flavor.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flavor.proto



<a name="edgeproto.Flavor"></a>

### Flavor
A Flavor identifies the Cpu, Ram, and Disk resources required for either a node in a Cluster, or an application instance. For a node in a cluster, these are the physical resources provided by that node. For an application instance, this defines the resources (per node) that should be allocated to the instance from the Cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [FlavorKey](#edgeproto.FlavorKey) |  | Unique key |
| ram | [uint64](#uint64) |  | RAM in MB |
| vcpus | [uint64](#uint64) |  | VCPU cores |
| disk | [uint64](#uint64) |  | Amount of disk in GB |






<a name="edgeproto.FlavorKey"></a>

### FlavorKey
FlavorKey uniquely identifies a Flavor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |





 

 

 


<a name="edgeproto.FlavorApi"></a>

### FlavorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Create a Flavor |
| DeleteFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Delete a Flavor |
| UpdateFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Update a Flavor |
| ShowFlavor | [Flavor](#edgeproto.Flavor) | [Flavor](#edgeproto.Flavor) stream | Show Flavors |

 



<a name="metric.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## metric.proto



<a name="edgeproto.Metric"></a>

### Metric
Metric is an entry/point in a time series of values for Analytics/Billing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Metric name |
| timestamp | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the metric was captured |
| tags | [MetricTag](#edgeproto.MetricTag) | repeated | Tags associated with the metric for searching/filtering |
| vals | [MetricVal](#edgeproto.MetricVal) | repeated | Values associated with the metric |






<a name="edgeproto.MetricTag"></a>

### MetricTag
MetricTag is used as a tag or label to look up the metric, beyond just the name of the metric.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Metric tag name |
| val | [string](#string) |  | Metric tag value |






<a name="edgeproto.MetricVal"></a>

### MetricVal
MetricVal is a value associated with the metric.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the value |
| dval | [double](#double) |  |  |
| ival | [uint64](#uint64) |  |  |





 

 

 

 



<a name="node.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## node.proto



<a name="edgeproto.Node"></a>

### Node
Node defines a DME (distributed matching engine) or CRM (cloudlet resource manager) instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [NodeKey](#edgeproto.NodeKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |






<a name="edgeproto.NodeKey"></a>

### NodeKey
NodeKey uniquely identifies a DME or CRM node


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name or hostname of node |
| node_type | [NodeType](#edgeproto.NodeType) |  | Node type |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which node is running, or is associated with |





 


<a name="edgeproto.NodeType"></a>

### NodeType
NodeType defines the type of Node

| Name | Number | Description |
| ---- | ------ | ----------- |
| NodeUnknown | 0 | Unknown |
| NodeDME | 1 | Distributed Matching Engine |
| NodeCRM | 2 | Cloudlet Resource Manager |


 

 


<a name="edgeproto.NodeApi"></a>

### NodeApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowNodeLocal | [Node](#edgeproto.Node) | [Node](#edgeproto.Node) stream | Show Nodes connected locally only |
| ShowNode | [Node](#edgeproto.Node) | [Node](#edgeproto.Node) stream | Show all Nodes connected to all Controllers |

 



<a name="notice.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## notice.proto
Notice is the message used by the notify protocol to communicate and coordinate internally between different Mobiledgex services. For details on the notify protocol, see the &#34;MEX Cloud Service Interactions&#34; confluence article.
In general, the protocol is used to synchronize state from one service to another. The protocol is fairly symmetric, with different state being synchronized both from server to client and client to server.


<a name="edgeproto.NoticeReply"></a>

### NoticeReply
NoticyReply is sent from server to client.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [NoticeAction](#edgeproto.NoticeAction) |  | Action to perform |
| version | [uint32](#uint32) |  | Protocol version supported by sender |
| app | [App](#edgeproto.App) |  |  |
| appInst | [AppInst](#edgeproto.AppInst) |  |  |
| cloudlet | [Cloudlet](#edgeproto.Cloudlet) |  |  |
| flavor | [Flavor](#edgeproto.Flavor) |  |  |
| clusterFlavor | [ClusterFlavor](#edgeproto.ClusterFlavor) |  |  |
| clusterInst | [ClusterInst](#edgeproto.ClusterInst) |  |  |






<a name="edgeproto.NoticeRequest"></a>

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
| metric | [Metric](#edgeproto.Metric) |  |  |
| node | [Node](#edgeproto.Node) |  |  |





 


<a name="edgeproto.NoticeAction"></a>

### NoticeAction
NoticeAction denotes what kind of action this notification is for.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 | No action |
| UPDATE | 1 | Update the object |
| DELETE | 2 | Delete the object |
| VERSION | 3 | Version exchange negotitation message |
| SENDALL_END | 4 | Initial send all finished message |



<a name="edgeproto.NoticeRequestor"></a>

### NoticeRequestor
NoticeRequestor indicates which type of service the client is.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NoticeRequestorNone | 0 | Invalid |
| NoticeRequestorDME | 1 | Distributed Matching Engine |
| NoticeRequestorCRM | 2 | Cloudlet Resource Manager |


 

 


<a name="edgeproto.NotifyApi"></a>

### NotifyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| StreamNotice | [NoticeRequest](#edgeproto.NoticeRequest) stream | [NoticeReply](#edgeproto.NoticeReply) stream | Bidrectional stream for exchanging data between controller and DME/CRM |

 



<a name="operator.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## operator.proto



<a name="edgeproto.Operator"></a>

### Operator
An Operator defines a telecommunications provider such as AT&amp;T, T-Mobile, etc. The operators in turn provide Mobiledgex with compute resource Cloudlets that serve as the basis for location-based services.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [OperatorKey](#edgeproto.OperatorKey) |  | Unique identifier key |






<a name="edgeproto.OperatorKey"></a>

### OperatorKey
OperatorKey uniquely identifies an Operator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Company or Organization name of the operator |





 

 

 


<a name="edgeproto.OperatorApi"></a>

### OperatorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateOperator | [Operator](#edgeproto.Operator) | [Result](#edgeproto.Result) | Create an Operator |
| DeleteOperator | [Operator](#edgeproto.Operator) | [Result](#edgeproto.Result) | Delete an Operator |
| UpdateOperator | [Operator](#edgeproto.Operator) | [Result](#edgeproto.Result) | Update an Operator |
| ShowOperator | [Operator](#edgeproto.Operator) | [Operator](#edgeproto.Operator) stream | Show Operators |

 



<a name="refs.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## refs.proto



<a name="edgeproto.CloudletRefs"></a>

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
| used_dynamic_ips | [int32](#int32) |  | Used dynamic IPs |
| used_static_ips | [string](#string) |  | Used static IPs |






<a name="edgeproto.CloudletRefs.RootLbPortsEntry"></a>

### CloudletRefs.RootLbPortsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [int32](#int32) |  |  |
| value | [int32](#int32) |  |  |






<a name="edgeproto.ClusterRefs"></a>

### ClusterRefs
ClusterRefs track used resources within a ClusterInst. Each AppInst specifies a set of required resources (Flavor), so tracking resources used by Apps within a Cluster is necessary to determine if enough resources are available for another AppInst to be instantiated on a ClusterInst.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Cluster Instance key |
| apps | [AppKey](#edgeproto.AppKey) | repeated | Apps instances in the Cluster Instance |
| used_ram | [uint64](#uint64) |  | Used RAM in MB |
| used_vcores | [uint64](#uint64) |  | Used VCPU cores |
| used_disk | [uint64](#uint64) |  | Used disk in GB |





 

 

 


<a name="edgeproto.CloudletRefsApi"></a>

### CloudletRefsApi
This API should be admin-only

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowCloudletRefs | [CloudletRefs](#edgeproto.CloudletRefs) | [CloudletRefs](#edgeproto.CloudletRefs) stream | Show CloudletRefs (debug only) |


<a name="edgeproto.ClusterRefsApi"></a>

### ClusterRefsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowClusterRefs | [ClusterRefs](#edgeproto.ClusterRefs) | [ClusterRefs](#edgeproto.ClusterRefs) stream | Show ClusterRefs (debug only) |

 



<a name="result.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## result.proto



<a name="edgeproto.Result"></a>

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

