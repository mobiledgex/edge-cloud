# Protocol Documentation
<a name="top"/>

## Table of Contents

- [app.proto](#app.proto)
    - [App](#edgeproto.App)
    - [AppKey](#edgeproto.AppKey)
    - [ConfigFile](#edgeproto.ConfigFile)
  
    - [DeleteType](#edgeproto.DeleteType)
    - [ImageType](#edgeproto.ImageType)
  
  
    - [AppApi](#edgeproto.AppApi)
  

- [app_inst.proto](#app_inst.proto)
    - [AppInst](#edgeproto.AppInst)
    - [AppInstInfo](#edgeproto.AppInstInfo)
    - [AppInstKey](#edgeproto.AppInstKey)
    - [AppInstMetrics](#edgeproto.AppInstMetrics)
    - [AppInstRuntime](#edgeproto.AppInstRuntime)
  
  
  
    - [AppInstApi](#edgeproto.AppInstApi)
    - [AppInstInfoApi](#edgeproto.AppInstInfoApi)
    - [AppInstMetricsApi](#edgeproto.AppInstMetricsApi)
  

- [cloudlet.proto](#cloudlet.proto)
    - [AzureProperties](#edgeproto.AzureProperties)
    - [Cloudlet](#edgeproto.Cloudlet)
    - [CloudletInfo](#edgeproto.CloudletInfo)
    - [CloudletInfraCommon](#edgeproto.CloudletInfraCommon)
    - [CloudletInfraProperties](#edgeproto.CloudletInfraProperties)
    - [CloudletKey](#edgeproto.CloudletKey)
    - [CloudletMetrics](#edgeproto.CloudletMetrics)
    - [FlavorInfo](#edgeproto.FlavorInfo)
    - [GcpProperties](#edgeproto.GcpProperties)
    - [OpenStackProperties](#edgeproto.OpenStackProperties)
    - [OpenStackProperties.OpenRcVarsEntry](#edgeproto.OpenStackProperties.OpenRcVarsEntry)
  
    - [CloudletState](#edgeproto.CloudletState)
  
  
    - [CloudletApi](#edgeproto.CloudletApi)
    - [CloudletInfoApi](#edgeproto.CloudletInfoApi)
    - [CloudletMetricsApi](#edgeproto.CloudletMetricsApi)
  

- [cluster.proto](#cluster.proto)
    - [Cluster](#edgeproto.Cluster)
    - [ClusterKey](#edgeproto.ClusterKey)
  
  
  
    - [ClusterApi](#edgeproto.ClusterApi)
  

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
  

- [exec.proto](#exec.proto)
    - [ExecRequest](#edgeproto.ExecRequest)
  
  
  
    - [ExecApi](#edgeproto.ExecApi)
  

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
    - [Notice](#edgeproto.Notice)
  
    - [NoticeAction](#edgeproto.NoticeAction)
  
  
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
  
  
  
  

- [version.proto](#version.proto)
  
    - [VersionHash](#edgeproto.VersionHash)
  
  
  

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
| image_path | [string](#string) |  | URI from which to download image |
| image_type | [ImageType](#edgeproto.ImageType) |  | Image type (see ImageType) |
| access_ports | [string](#string) |  | For Layer4 access, the ports the app listens on. This is a comma separated list of protocol:port pairs, i.e. tcp:80,http:443,udp:10002. Only tcp, udp, and http protocols are supported; tcp and udp are assumed to be L4, and http is assumed to be L7 access. |
| config | [string](#string) |  | URI of resource to be used to establish config for App. |
| default_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Default flavor for the App, may be overridden by the AppInst |
| cluster | [ClusterKey](#edgeproto.ClusterKey) |  | Cluster field is deprecated. Apps no longer have dependencies on Clusters. |
| app_template | [string](#string) |  | Template of kubernetes deployment yaml. Who/What sets this is TDB, but it should not be directly exposed to the user, because we do not want to expose kubernetes to the user. However, because we currently don&#39;t have any other way to set it, for flexibility, for now it is exposed to the user. |
| auth_public_key | [string](#string) |  | public key used for authentication |
| command | [string](#string) |  | Command to start service |
| annotations | [string](#string) |  | Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=&#34;val 3&#34; |
| deployment | [string](#string) |  | Deployment target (kubernetes, docker, kvm, etc) |
| deployment_manifest | [string](#string) |  | Deployment manifest is the deployment specific manifest file/config |
| deployment_generator | [string](#string) |  | Deployment generator target |
| android_package_name | [string](#string) |  | Android package name, optional |
| permits_platform_apps | [bool](#bool) |  | Indicates whether or not platform apps are allowed to perform actions on behalf of this app, such as FindCloudlet |
| del_opt | [DeleteType](#edgeproto.DeleteType) |  | Override actions to Controller |
| configs | [ConfigFile](#edgeproto.ConfigFile) | repeated | Customization files |






<a name="edgeproto.AppKey"/>

### AppKey
AppKey uniquely identifies an Application.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| developer_key | [DeveloperKey](#edgeproto.DeveloperKey) |  | Developer key |
| name | [string](#string) |  | Application name |
| version | [string](#string) |  | Version of the app |






<a name="edgeproto.ConfigFile"/>

### ConfigFile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | kind (type) of config, i.e. k8s-manifest, helm-values, deploygen-config |
| config | [string](#string) |  | config file contents or URI reference |





 


<a name="edgeproto.DeleteType"/>

### DeleteType


| Name | Number | Description |
| ---- | ------ | ----------- |
| NO_AUTO_DELETE | 0 | No autodelete |
| AUTO_DELETE | 1 | Autodelete |



<a name="edgeproto.ImageType"/>

### ImageType
ImageType specifies the image type of the application.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IMAGE_TYPE_UNKNOWN | 0 | Unknown image type |
| IMAGE_TYPE_DOCKER | 1 | Docker container image type compatible with Kubernetes |
| IMAGE_TYPE_QCOW | 2 | QCOW2 virtual machine image type |


 

 


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
| uri | [string](#string) |  | Base FQDN (not really URI) for the App. See Service FQDN for endpoint access. |
| liveness | [Liveness](#edgeproto.Liveness) |  | Liveness of instance (see Liveness) |
| mapped_ports | [distributed_match_engine.AppPort](#distributed_match_engine.AppPort) | repeated | For instances accessible via a shared load balancer, defines the external ports on the shared load balancer that map to the internal ports External ports should be appended to the Uri for L4 access. |
| flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Flavor defining resource requirements |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the AppInst on the Cloudlet |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the AppInst on the Cloudlet |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |
| runtime_info | [AppInstRuntime](#edgeproto.AppInstRuntime) |  | AppInst runtime information |
| created_at | [distributed_match_engine.Timestamp](#distributed_match_engine.Timestamp) |  | Created at time |
| auto_cluster_ip_access | [IpAccess](#edgeproto.IpAccess) |  | IpAccess for auto-clusters. Ignored otherwise. |






<a name="edgeproto.AppInstInfo"/>

### AppInstInfo
AppInstInfo provides information from the Cloudlet Resource Manager about the state of the AppInst on the Cloudlet. Whereas the AppInst defines the intent of instantiating an App on a Cloudlet, the AppInstInfo defines the current state of trying to apply that intent on the physical resources of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the AppInst on the Cloudlet |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the AppInst on the Cloudlet |
| runtime_info | [AppInstRuntime](#edgeproto.AppInstRuntime) |  | AppInst runtime information |






<a name="edgeproto.AppInstKey"/>

### AppInstKey
AppInstKey uniquely identifies an Application Instance (AppInst) or Application Instance state (AppInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_key | [AppKey](#edgeproto.AppKey) |  | App key |
| cluster_inst_key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Cluster instance on which this is instantiated |






<a name="edgeproto.AppInstMetrics"/>

### AppInstMetrics
(TODO) AppInstMetrics provide metrics collected about the application instance on the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| something | [uint64](#uint64) |  | what goes here? Note that metrics for grpc calls can be done by a prometheus interceptor in grpc, so adding call metrics here may be redundant unless they&#39;re needed for billing. |






<a name="edgeproto.AppInstRuntime"/>

### AppInstRuntime



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| container_ids | [string](#string) | repeated | List of container names |





 

 

 


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

 



<a name="cloudlet.proto"/>
<p align="right"><a href="#top">Top</a></p>

## cloudlet.proto



<a name="edgeproto.AzureProperties"/>

### AzureProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| location | [string](#string) |  | azure region e.g. uswest2 |
| resource_group | [string](#string) |  | azure resource group |
| user_name | [string](#string) |  | azure username |
| password | [string](#string) |  | azure password |






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
| controller | [string](#string) |  | Connected controller unique id |
| os_max_ram | [uint64](#uint64) |  | Maximum Ram in MB on the Cloudlet |
| os_max_vcores | [uint64](#uint64) |  | Maximum number of VCPU cores on the Cloudlet |
| os_max_vol_gb | [uint64](#uint64) |  | Maximum amount of disk in GB on the Cloudlet |
| errors | [string](#string) | repeated | Any errors encountered while making changes to the Cloudlet |
| flavors | [FlavorInfo](#edgeproto.FlavorInfo) | repeated | Supported flavors by the Cloudlet |






<a name="edgeproto.CloudletInfraCommon"/>

### CloudletInfraCommon
properites common to all cloudlets


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| docker_registry | [string](#string) |  | the mex docker registry, e.g. registry.mobiledgex.net:5000. |
| dns_zone | [string](#string) |  | DNS Zone |
| registry_file_server | [string](#string) |  | registry file server contains files which get pulled on instantiation such as certs and images |
| cf_key | [string](#string) |  | Cloudflare key

MEX_CF_KEY |
| cf_user | [string](#string) |  | Cloudflare key

MEX_CF_KEY |
| docker_reg_pass | [string](#string) |  | Docker registry password

MEX_DOCKER_REG_PASS |
| network_scheme | [string](#string) |  | network scheme |
| docker_registry_secret | [string](#string) |  | the name of the docker registry secret, e.g. mexgitlabsecret |






<a name="edgeproto.CloudletInfraProperties"/>

### CloudletInfraProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cloudlet_kind | [string](#string) |  | what kind of infrastructure: Azure, GCP, Openstack |
| mexos_container_image_name | [string](#string) |  | name and version of the docker image container image that mexos runs in |
| openstack_properties | [OpenStackProperties](#edgeproto.OpenStackProperties) |  | openstack |
| azure_properties | [AzureProperties](#edgeproto.AzureProperties) |  | azure |
| gcp_properties | [GcpProperties](#edgeproto.GcpProperties) |  | gcp |






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






<a name="edgeproto.FlavorInfo"/>

### FlavorInfo
Flavor details from the Cloudlet


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the flavor on the Cloudlet |
| vcpus | [uint64](#uint64) |  | Number of VCPU cores on the Cloudlet |
| ram | [uint64](#uint64) |  | Ram in MB on the Cloudlet |
| disk | [uint64](#uint64) |  | Amount of disk in GB on the Cloudlet |






<a name="edgeproto.GcpProperties"/>

### GcpProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | gcp project for billing |
| zone | [string](#string) |  | availability zone |
| service_account | [string](#string) |  | service account to login with |
| gcp_auth_key_url | [string](#string) |  | vault credentials link |






<a name="edgeproto.OpenStackProperties"/>

### OpenStackProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| os_external_network_name | [string](#string) |  | name of the external network, e.g. external-network-shared |
| os_image_name | [string](#string) |  | openstack image , e.g. mobiledgex |
| os_external_router_name | [string](#string) |  | openstack router |
| os_mex_network | [string](#string) |  | openstack internal network |
| open_rc_vars | [OpenStackProperties.OpenRcVarsEntry](#edgeproto.OpenStackProperties.OpenRcVarsEntry) | repeated | openrc env vars |






<a name="edgeproto.OpenStackProperties.OpenRcVarsEntry"/>

### OpenStackProperties.OpenRcVarsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="edgeproto.CloudletState"/>

### CloudletState
CloudletState is the state of the Cloudlet.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CLOUDLET_STATE_UNKNOWN | 0 | Unknown |
| CLOUDLET_STATE_ERRORS | 1 | Create/Delete/Update encountered errors (see Errors field of CloudletInfo) |
| CLOUDLET_STATE_READY | 2 | Cloudlet is created and ready |
| CLOUDLET_STATE_OFFLINE | 3 | Cloudlet is offline (unreachable) |
| CLOUDLET_STATE_NOT_PRESENT | 4 | Cloudlet is not present |


 

 


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
| default_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Default flavor of the Cluster, may be overridden on the ClusterInst |
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
| flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Node flavor |
| liveness | [Liveness](#edgeproto.Liveness) |  | Liveness of instance (see Liveness) |
| auto | [bool](#bool) |  | Auto is set to true when automatically created by back-end (internal use only) |
| state | [TrackedState](#edgeproto.TrackedState) |  | State of the cluster instance |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the ClusterInst on the Cloudlet. |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |
| ip_access | [IpAccess](#edgeproto.IpAccess) |  | IP access type |
| allocated_ip | [string](#string) |  | allocated IP for dedicated access |
| node_flavor | [string](#string) |  | Cloudlet specific node flavor |
| deployment | [string](#string) |  | Deployment target (kubernetes, docker, kvm, etc) |
| num_masters | [uint32](#uint32) |  | number of masters |
| num_nodes | [uint32](#uint32) |  | number of nodes |






<a name="edgeproto.ClusterInstInfo"/>

### ClusterInstInfo
ClusterInstInfo provides information from the Cloudlet Resource Manager about the state of the ClusterInst on the Cloudlet. Whereas the ClusterInst defines the intent of instantiating a Cluster on a Cloudlet, the ClusterInstInfo defines the current state of trying to apply that intent on the physical resources of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| state | [TrackedState](#edgeproto.TrackedState) |  | State of the cluster instance |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the ClusterInst on the Cloudlet. |






<a name="edgeproto.ClusterInstKey"/>

### ClusterInstKey
ClusterInstKey uniquely identifies a Cluster Instance (ClusterInst) or Cluster Instance state (ClusterInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_key | [ClusterKey](#edgeproto.ClusterKey) |  | Cluster key |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which the Cluster is instantiated |
| developer | [string](#string) |  | Developer organization this cluster belongs to |





 

 

 


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


 


<a name="edgeproto.CRMOverride"/>

### CRMOverride
CRMOverride can be applied to commands that issue requests to the CRM.
It should only be used by administrators when bugs have caused the
Controller and CRM to get out of sync. It allows commands from the
Controller to ignore errors from the CRM, or ignore the CRM completely
(messages will not be sent to CRM).

| Name | Number | Description |
| ---- | ------ | ----------- |
| NO_OVERRIDE | 0 | No override |
| IGNORE_CRM_ERRORS | 1 | Ignore errors from CRM |
| IGNORE_CRM | 2 | Ignore CRM completely (does not inform CRM of operation) |
| IGNORE_TRANSIENT_STATE | 3 | Ignore Transient State (only admin should use if CRM crashed) |
| IGNORE_CRM_AND_TRANSIENT_STATE | 4 | Ignore CRM and Transient State |



<a name="edgeproto.IpAccess"/>

### IpAccess


| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_ACCESS_UNKNOWN | 0 | Unknown IP access |
| IP_ACCESS_DEDICATED | 1 | Dedicated IP access |
| IP_ACCESS_DEDICATED_OR_SHARED | 2 | Dedicated or shared (prefers dedicated) access |
| IP_ACCESS_SHARED | 3 | Shared IP access |



<a name="edgeproto.IpSupport"/>

### IpSupport
IpSupport indicates the type of public IP support provided by the Cloudlet. Static IP support indicates a set of static public IPs are available for use, and managed by the Controller. Dynamic indicates the Cloudlet uses a DHCP server to provide public IP addresses, and the controller has no control over which IPs are assigned.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_SUPPORT_UNKNOWN | 0 | Unknown IP support |
| IP_SUPPORT_STATIC | 1 | Static IP addresses are provided to and managed by Controller |
| IP_SUPPORT_DYNAMIC | 2 | IP addresses are dynamically provided by an Operator&#39;s DHCP server |



<a name="edgeproto.Liveness"/>

### Liveness
Liveness indicates if an object was created statically via an external API call, or dynamically via an internal algorithm.

| Name | Number | Description |
| ---- | ------ | ----------- |
| LIVENESS_UNKNOWN | 0 | Unknown liveness |
| LIVENESS_STATIC | 1 | Object managed by external entity |
| LIVENESS_DYNAMIC | 2 | Object managed internally |



<a name="edgeproto.TrackedState"/>

### TrackedState
TrackedState is used to track the state of an object on a remote node,
i.e. track the state of a ClusterInst object on the CRM (Cloudlet).

| Name | Number | Description |
| ---- | ------ | ----------- |
| TRACKED_STATE_UNKNOWN | 0 | Unknown state |
| NOT_PRESENT | 1 | Not present (does not exist) |
| CREATE_REQUESTED | 2 | Create requested |
| CREATING | 3 | Creating |
| CREATE_ERROR | 4 | Create error |
| READY | 5 | Ready |
| UPDATE_REQUESTED | 6 | Update requested |
| UPDATING | 7 | Updating |
| UPDATE_ERROR | 8 | Update error |
| DELETE_REQUESTED | 9 | Delete requested |
| DELETING | 10 | Deleting |
| DELETE_ERROR | 11 | Delete error |
| DELETE_PREPARE | 12 | Delete prepare (extra state used by controller to block other changes) |


 

 

 



<a name="controller.proto"/>
<p align="right"><a href="#top">Top</a></p>

## controller.proto



<a name="edgeproto.Controller"/>

### Controller
A Controller is a service that manages the edge-cloud data and controls other edge-cloud micro-services.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ControllerKey](#edgeproto.ControllerKey) |  | Unique identifier key |






<a name="edgeproto.ControllerKey"/>

### ControllerKey
ControllerKey uniquely defines a Controller


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| addr | [string](#string) |  | external API address |





 

 

 


<a name="edgeproto.ControllerApi"/>

### ControllerApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowController | [Controller](#edgeproto.Controller) | [Controller](#edgeproto.Controller) | Show Controllers |

 



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

 



<a name="exec.proto"/>
<p align="right"><a href="#top">Top</a></p>

## exec.proto



<a name="edgeproto.ExecRequest"/>

### ExecRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_inst_key | [AppInstKey](#edgeproto.AppInstKey) |  | Target AppInst |
| command | [string](#string) |  | Command or Shell |
| container_id | [string](#string) |  | ContainerID is the name of the target container, if applicable |
| offer | [string](#string) |  | WebRTC Offer |
| answer | [string](#string) |  | WebRTC Answer |
| err | [string](#string) |  | Any error message |





 

 

 


<a name="edgeproto.ExecApi"/>

### ExecApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| RunCommand | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | Run a Command or Shell on a container or VM |
| SendLocalRequest | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | This is used internally to forward requests to other Controllers. |

 



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

 



<a name="metric.proto"/>
<p align="right"><a href="#top">Top</a></p>

## metric.proto



<a name="edgeproto.Metric"/>

### Metric
Metric is an entry/point in a time series of values for Analytics/Billing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Metric name |
| timestamp | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the metric was captured |
| tags | [MetricTag](#edgeproto.MetricTag) | repeated | Tags associated with the metric for searching/filtering |
| vals | [MetricVal](#edgeproto.MetricVal) | repeated | Values associated with the metric |






<a name="edgeproto.MetricTag"/>

### MetricTag
MetricTag is used as a tag or label to look up the metric, beyond just the name of the metric.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Metric tag name |
| val | [string](#string) |  | Metric tag value |






<a name="edgeproto.MetricVal"/>

### MetricVal
MetricVal is a value associated with the metric.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the value |
| dval | [double](#double) |  |  |
| ival | [uint64](#uint64) |  |  |





 

 

 

 



<a name="node.proto"/>
<p align="right"><a href="#top">Top</a></p>

## node.proto



<a name="edgeproto.Node"/>

### Node
Node defines a DME (distributed matching engine) or CRM (cloudlet resource manager) instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [NodeKey](#edgeproto.NodeKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |






<a name="edgeproto.NodeKey"/>

### NodeKey
NodeKey uniquely identifies a DME or CRM node


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name or hostname of node |
| node_type | [NodeType](#edgeproto.NodeType) |  | Node type |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which node is running, or is associated with |





 


<a name="edgeproto.NodeType"/>

### NodeType
NodeType defines the type of Node

| Name | Number | Description |
| ---- | ------ | ----------- |
| NODE_UNKNOWN | 0 | Unknown |
| NODE_DME | 1 | Distributed Matching Engine |
| NODE_CRM | 2 | Cloudlet Resource Manager |


 

 


<a name="edgeproto.NodeApi"/>

### NodeApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowNodeLocal | [Node](#edgeproto.Node) | [Node](#edgeproto.Node) | Show Nodes connected locally only |
| ShowNode | [Node](#edgeproto.Node) | [Node](#edgeproto.Node) | Show all Nodes connected to all Controllers |

 



<a name="notice.proto"/>
<p align="right"><a href="#top">Top</a></p>

## notice.proto
Notice is the message used by the notify protocol to communicate and coordinate internally between different Mobiledgex services. For details on the notify protocol, see the &#34;MEX Cloud Service Interactions&#34; confluence article.
In general, the protocol is used to synchronize state from one service to another. The protocol is fairly symmetric, with different state being synchronized both from server to client and client to server.


<a name="edgeproto.Notice"/>

### Notice



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [NoticeAction](#edgeproto.NoticeAction) |  | Action to perform |
| version | [uint32](#uint32) |  | Protocol version supported by sender |
| any | [google.protobuf.Any](#google.protobuf.Any) |  | Data |
| want_objs | [string](#string) | repeated | Wanted Objects |
| filter_cloudlet_key | [bool](#bool) |  | Filter by cloudlet key |





 


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


 

 


<a name="edgeproto.NotifyApi"/>

### NotifyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| StreamNotice | [Notice](#edgeproto.Notice) | [Notice](#edgeproto.Notice) | Bidrectional stream for exchanging data between controller and DME/CRM |

 



<a name="operator.proto"/>
<p align="right"><a href="#top">Top</a></p>

## operator.proto



<a name="edgeproto.Operator"/>

### Operator
An Operator defines a telecommunications provider such as AT&amp;T, T-Mobile, etc. The operators in turn provide Mobiledgex with compute resource Cloudlets that serve as the basis for location-based services.


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
| root_lb_ports | [CloudletRefs.RootLbPortsEntry](#edgeproto.CloudletRefs.RootLbPortsEntry) | repeated | Used ports on root load balancer. Map key is public port, value is a bitmap for the protocol bitmap: bit 0: tcp, bit 1: udp |
| used_dynamic_ips | [int32](#int32) |  | Used dynamic IPs |
| used_static_ips | [string](#string) |  | Used static IPs |






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





 

 

 

 



<a name="version.proto"/>
<p align="right"><a href="#top">Top</a></p>

## version.proto


 


<a name="edgeproto.VersionHash"/>

### VersionHash
Below enum lists hashes as well as corresponding versions

| Name | Number | Description |
| ---- | ------ | ----------- |
| HASH_d41d8cd98f00b204e9800998ecf8427e | 0 |  |
| HASH_b35326df0fcd1550b7c0cf6460c4bca2 | 1 |  |
| HASH_52e6980599cd59bbbd0de8d5f4d53d4b | 2 |  |
| HASH_00bdcfa956ca4ee42be87abcd8fcaf1c | 3 |  |
| HASH_0d2d9c0b07ad989e96fb3b3a44924316 | 4 |  |
| HASH_2b79f0b6e402045ee5f68d697b9386ae | 5 |  |


 

 

 



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

