# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [alert.proto](#alert.proto)
    - [Alert](#edgeproto.Alert)
    - [Alert.AnnotationsEntry](#edgeproto.Alert.AnnotationsEntry)
    - [Alert.LabelsEntry](#edgeproto.Alert.LabelsEntry)
  
    - [AlertApi](#edgeproto.AlertApi)
  
- [alldata.proto](#alldata.proto)
    - [AllData](#edgeproto.AllData)
  
- [app.proto](#app.proto)
    - [App](#edgeproto.App)
    - [AppAutoProvPolicy](#edgeproto.AppAutoProvPolicy)
    - [AppKey](#edgeproto.AppKey)
    - [ConfigFile](#edgeproto.ConfigFile)
  
    - [AccessType](#edgeproto.AccessType)
    - [DeleteType](#edgeproto.DeleteType)
    - [ImageType](#edgeproto.ImageType)
  
    - [AppApi](#edgeproto.AppApi)
  
- [appinst.proto](#appinst.proto)
    - [AppInst](#edgeproto.AppInst)
    - [AppInstInfo](#edgeproto.AppInstInfo)
    - [AppInstKey](#edgeproto.AppInstKey)
    - [AppInstLookup](#edgeproto.AppInstLookup)
    - [AppInstMetrics](#edgeproto.AppInstMetrics)
    - [AppInstRuntime](#edgeproto.AppInstRuntime)
  
    - [HealthCheck](#dme.HealthCheck)
    - [PowerState](#edgeproto.PowerState)
  
    - [AppInstApi](#edgeproto.AppInstApi)
    - [AppInstInfoApi](#edgeproto.AppInstInfoApi)
    - [AppInstMetricsApi](#edgeproto.AppInstMetricsApi)
  
- [appinstclient.proto](#appinstclient.proto)
    - [AppInstClient](#edgeproto.AppInstClient)
    - [AppInstClientKey](#edgeproto.AppInstClientKey)
  
    - [AppInstClientApi](#edgeproto.AppInstClientApi)
  
- [autoprovpolicy.proto](#autoprovpolicy.proto)
    - [AutoProvCloudlet](#edgeproto.AutoProvCloudlet)
    - [AutoProvCount](#edgeproto.AutoProvCount)
    - [AutoProvCounts](#edgeproto.AutoProvCounts)
    - [AutoProvInfo](#edgeproto.AutoProvInfo)
    - [AutoProvPolicy](#edgeproto.AutoProvPolicy)
    - [AutoProvPolicyCloudlet](#edgeproto.AutoProvPolicyCloudlet)
  
    - [AutoProvPolicyApi](#edgeproto.AutoProvPolicyApi)
  
- [autoscalepolicy.proto](#autoscalepolicy.proto)
    - [AutoScalePolicy](#edgeproto.AutoScalePolicy)
    - [PolicyKey](#edgeproto.PolicyKey)
  
    - [AutoScalePolicyApi](#edgeproto.AutoScalePolicyApi)
  
- [cloudlet.proto](#cloudlet.proto)
    - [Cloudlet](#edgeproto.Cloudlet)
    - [Cloudlet.AccessVarsEntry](#edgeproto.Cloudlet.AccessVarsEntry)
    - [Cloudlet.ChefClientKeyEntry](#edgeproto.Cloudlet.ChefClientKeyEntry)
    - [Cloudlet.EnvVarEntry](#edgeproto.Cloudlet.EnvVarEntry)
    - [Cloudlet.ResTagMapEntry](#edgeproto.Cloudlet.ResTagMapEntry)
    - [CloudletInfo](#edgeproto.CloudletInfo)
    - [CloudletKey](#edgeproto.CloudletKey)
    - [CloudletManifest](#edgeproto.CloudletManifest)
    - [CloudletMetrics](#edgeproto.CloudletMetrics)
    - [CloudletProps](#edgeproto.CloudletProps)
    - [CloudletProps.PropertiesEntry](#edgeproto.CloudletProps.PropertiesEntry)
    - [CloudletResMap](#edgeproto.CloudletResMap)
    - [CloudletResMap.MappingEntry](#edgeproto.CloudletResMap.MappingEntry)
    - [FlavorInfo](#edgeproto.FlavorInfo)
    - [FlavorInfo.PropMapEntry](#edgeproto.FlavorInfo.PropMapEntry)
    - [FlavorMatch](#edgeproto.FlavorMatch)
    - [InfraConfig](#edgeproto.InfraConfig)
    - [OSAZone](#edgeproto.OSAZone)
    - [OSImage](#edgeproto.OSImage)
    - [OperationTimeLimits](#edgeproto.OperationTimeLimits)
    - [PlatformConfig](#edgeproto.PlatformConfig)
    - [PlatformConfig.EnvVarEntry](#edgeproto.PlatformConfig.EnvVarEntry)
    - [PropertyInfo](#edgeproto.PropertyInfo)
  
    - [CloudletState](#dme.CloudletState)
    - [InfraApiAccess](#edgeproto.InfraApiAccess)
    - [PlatformType](#edgeproto.PlatformType)
  
    - [CloudletApi](#edgeproto.CloudletApi)
    - [CloudletInfoApi](#edgeproto.CloudletInfoApi)
    - [CloudletMetricsApi](#edgeproto.CloudletMetricsApi)
  
- [cloudletpool.proto](#cloudletpool.proto)
    - [CloudletPool](#edgeproto.CloudletPool)
    - [CloudletPoolKey](#edgeproto.CloudletPoolKey)
    - [CloudletPoolMember](#edgeproto.CloudletPoolMember)
  
    - [CloudletPoolApi](#edgeproto.CloudletPoolApi)
  
- [cluster.proto](#cluster.proto)
    - [ClusterKey](#edgeproto.ClusterKey)
  
- [clusterinst.proto](#clusterinst.proto)
    - [ClusterInst](#edgeproto.ClusterInst)
    - [ClusterInstInfo](#edgeproto.ClusterInstInfo)
    - [ClusterInstKey](#edgeproto.ClusterInstKey)
  
    - [ClusterInstApi](#edgeproto.ClusterInstApi)
    - [ClusterInstInfoApi](#edgeproto.ClusterInstInfoApi)
  
- [common.proto](#common.proto)
    - [StatusInfo](#edgeproto.StatusInfo)
  
    - [CRMOverride](#edgeproto.CRMOverride)
    - [IpAccess](#edgeproto.IpAccess)
    - [IpSupport](#edgeproto.IpSupport)
    - [Liveness](#edgeproto.Liveness)
    - [MaintenanceState](#dme.MaintenanceState)
    - [TrackedState](#edgeproto.TrackedState)
  
- [controller.proto](#controller.proto)
    - [Controller](#edgeproto.Controller)
    - [ControllerKey](#edgeproto.ControllerKey)
  
    - [ControllerApi](#edgeproto.ControllerApi)
  
- [debug.proto](#debug.proto)
    - [DebugData](#edgeproto.DebugData)
    - [DebugReply](#edgeproto.DebugReply)
    - [DebugRequest](#edgeproto.DebugRequest)
  
    - [DebugApi](#edgeproto.DebugApi)
  
- [device.proto](#device.proto)
    - [Device](#edgeproto.Device)
    - [DeviceData](#edgeproto.DeviceData)
    - [DeviceKey](#edgeproto.DeviceKey)
    - [DeviceReport](#edgeproto.DeviceReport)
  
    - [DeviceApi](#edgeproto.DeviceApi)
  
- [exec.proto](#exec.proto)
    - [CloudletMgmtNode](#edgeproto.CloudletMgmtNode)
    - [ExecRequest](#edgeproto.ExecRequest)
    - [RunCmd](#edgeproto.RunCmd)
    - [RunVMConsole](#edgeproto.RunVMConsole)
    - [ShowLog](#edgeproto.ShowLog)
  
    - [ExecApi](#edgeproto.ExecApi)
  
- [flavor.proto](#flavor.proto)
    - [Flavor](#edgeproto.Flavor)
    - [Flavor.OptResMapEntry](#edgeproto.Flavor.OptResMapEntry)
    - [FlavorKey](#edgeproto.FlavorKey)
  
    - [FlavorApi](#edgeproto.FlavorApi)
  
- [metric.proto](#metric.proto)
    - [Metric](#edgeproto.Metric)
    - [MetricTag](#edgeproto.MetricTag)
    - [MetricVal](#edgeproto.MetricVal)
  
- [node.proto](#node.proto)
    - [Node](#edgeproto.Node)
    - [NodeData](#edgeproto.NodeData)
    - [NodeKey](#edgeproto.NodeKey)
  
    - [NodeApi](#edgeproto.NodeApi)
  
- [notice.proto](#notice.proto)
    - [Notice](#edgeproto.Notice)
    - [Notice.TagsEntry](#edgeproto.Notice.TagsEntry)
  
    - [NoticeAction](#edgeproto.NoticeAction)
  
    - [NotifyApi](#edgeproto.NotifyApi)
  
- [operatorcode.proto](#operatorcode.proto)
    - [OperatorCode](#edgeproto.OperatorCode)
  
    - [OperatorCodeApi](#edgeproto.OperatorCodeApi)
  
- [org.proto](#org.proto)
    - [Organization](#edgeproto.Organization)
    - [OrganizationData](#edgeproto.OrganizationData)
  
    - [OrganizationApi](#edgeproto.OrganizationApi)
  
- [privacypolicy.proto](#privacypolicy.proto)
    - [OutboundSecurityRule](#edgeproto.OutboundSecurityRule)
    - [PrivacyPolicy](#edgeproto.PrivacyPolicy)
  
    - [PrivacyPolicyApi](#edgeproto.PrivacyPolicyApi)
  
- [refs.proto](#refs.proto)
    - [AppInstRefs](#edgeproto.AppInstRefs)
    - [AppInstRefs.InstsEntry](#edgeproto.AppInstRefs.InstsEntry)
    - [CloudletRefs](#edgeproto.CloudletRefs)
    - [CloudletRefs.OptResUsedMapEntry](#edgeproto.CloudletRefs.OptResUsedMapEntry)
    - [CloudletRefs.RootLbPortsEntry](#edgeproto.CloudletRefs.RootLbPortsEntry)
    - [ClusterRefs](#edgeproto.ClusterRefs)
  
    - [AppInstRefsApi](#edgeproto.AppInstRefsApi)
    - [CloudletRefsApi](#edgeproto.CloudletRefsApi)
    - [ClusterRefsApi](#edgeproto.ClusterRefsApi)
  
- [restagtable.proto](#restagtable.proto)
    - [ResTagTable](#edgeproto.ResTagTable)
    - [ResTagTable.TagsEntry](#edgeproto.ResTagTable.TagsEntry)
    - [ResTagTableKey](#edgeproto.ResTagTableKey)
  
    - [OptResNames](#edgeproto.OptResNames)
  
    - [ResTagTableApi](#edgeproto.ResTagTableApi)
  
- [result.proto](#result.proto)
    - [Result](#edgeproto.Result)
  
- [settings.proto](#settings.proto)
    - [Settings](#edgeproto.Settings)
  
    - [SettingsApi](#edgeproto.SettingsApi)
  
- [version.proto](#version.proto)
    - [VersionHash](#edgeproto.VersionHash)
  
- [vmpool.proto](#vmpool.proto)
    - [VM](#edgeproto.VM)
    - [VMNetInfo](#edgeproto.VMNetInfo)
    - [VMPool](#edgeproto.VMPool)
    - [VMPoolInfo](#edgeproto.VMPoolInfo)
    - [VMPoolKey](#edgeproto.VMPoolKey)
    - [VMPoolMember](#edgeproto.VMPoolMember)
    - [VMSpec](#edgeproto.VMSpec)
  
    - [VMAction](#edgeproto.VMAction)
    - [VMState](#edgeproto.VMState)
  
    - [VMPoolApi](#edgeproto.VMPoolApi)
  
- [Scalar Value Types](#scalar-value-types)



<a name="alert.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## alert.proto



<a name="edgeproto.Alert"></a>

### Alert



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [Alert.LabelsEntry](#edgeproto.Alert.LabelsEntry) | repeated | Labels uniquely define the alert |
| annotations | [Alert.AnnotationsEntry](#edgeproto.Alert.AnnotationsEntry) | repeated | Annotations are extra information about the alert |
| state | [string](#string) |  | State of the alert |
| active_at | [distributed_match_engine.Timestamp](#distributed_match_engine.Timestamp) |  | When alert became active |
| value | [double](#double) |  | Any value associated with alert |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| controller | [string](#string) |  | Connected controller unique id |






<a name="edgeproto.Alert.AnnotationsEntry"></a>

### Alert.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.Alert.LabelsEntry"></a>

### Alert.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="edgeproto.AlertApi"></a>

### AlertApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAlert | [Alert](#edgeproto.Alert) | [Alert](#edgeproto.Alert) stream | Show alerts |

 



<a name="alldata.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## alldata.proto



<a name="edgeproto.AllData"></a>

### AllData
AllData contains all data that may be used for declarative
create/delete, or as input for e2e tests.
The order of fields here is important, as objects will be
created in the order they are specified here, and deleted
in the opposite order. The field ID (number) doesn&#39;t matter.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| flavors | [Flavor](#edgeproto.Flavor) | repeated |  |
| settings | [Settings](#edgeproto.Settings) |  |  |
| operator_codes | [OperatorCode](#edgeproto.OperatorCode) | repeated |  |
| res_tag_tables | [ResTagTable](#edgeproto.ResTagTable) | repeated |  |
| cloudlets | [Cloudlet](#edgeproto.Cloudlet) | repeated |  |
| cloudlet_infos | [CloudletInfo](#edgeproto.CloudletInfo) | repeated |  |
| cloudlet_pools | [CloudletPool](#edgeproto.CloudletPool) | repeated |  |
| auto_prov_policies | [AutoProvPolicy](#edgeproto.AutoProvPolicy) | repeated |  |
| auto_prov_policy_cloudlets | [AutoProvPolicyCloudlet](#edgeproto.AutoProvPolicyCloudlet) | repeated |  |
| auto_scale_policies | [AutoScalePolicy](#edgeproto.AutoScalePolicy) | repeated |  |
| privacy_policies | [PrivacyPolicy](#edgeproto.PrivacyPolicy) | repeated |  |
| cluster_insts | [ClusterInst](#edgeproto.ClusterInst) | repeated |  |
| apps | [App](#edgeproto.App) | repeated |  |
| app_instances | [AppInst](#edgeproto.AppInst) | repeated |  |
| app_inst_refs | [AppInstRefs](#edgeproto.AppInstRefs) | repeated |  |
| vm_pools | [VMPool](#edgeproto.VMPool) | repeated |  |





 

 

 

 



<a name="app.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## app.proto



<a name="edgeproto.App"></a>

### App
Application

App belongs to developer organizations and is used to provide information about their application.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppKey](#edgeproto.AppKey) |  | required: true Unique identifier key |
| image_path | [string](#string) |  | URI of where image resides |
| image_type | [ImageType](#edgeproto.ImageType) |  | Image type (see ImageType) |
| access_ports | [string](#string) |  | Comma separated list of protocol:port pairs that the App listens on. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443 |
| default_flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Default flavor for the App, which may be overridden by the AppInst |
| auth_public_key | [string](#string) |  | public key used for authentication |
| command | [string](#string) |  | Command that the container runs to start service |
| annotations | [string](#string) |  | Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=&#34;val 3&#34; |
| deployment | [string](#string) |  | Deployment type (kubernetes, docker, or vm) |
| deployment_manifest | [string](#string) |  | Deployment manifest is the deployment specific manifest file/config For docker deployment, this can be a docker-compose or docker run file For kubernetes deployment, this can be a kubernetes yaml or helm chart file |
| deployment_generator | [string](#string) |  | Deployment generator target to generate a basic deployment manifest |
| android_package_name | [string](#string) |  | Android package name used to match the App name from the Android package |
| del_opt | [DeleteType](#edgeproto.DeleteType) |  | Override actions to Controller |
| configs | [ConfigFile](#edgeproto.ConfigFile) | repeated | Customization files passed through to implementing services |
| scale_with_cluster | [bool](#bool) |  | Option to run App on all nodes of the cluster |
| internal_ports | [bool](#bool) |  | Should this app have access to outside world? |
| revision | [string](#string) |  | Revision can be specified or defaults to current timestamp when app is updated |
| official_fqdn | [string](#string) |  | Official FQDN is the FQDN that the app uses to connect by default |
| md5sum | [string](#string) |  | MD5Sum of the VM-based app image |
| default_shared_volume_size | [uint64](#uint64) |  | shared volume size when creating auto cluster |
| auto_prov_policy | [string](#string) |  | (_deprecated_) Auto provisioning policy name |
| access_type | [AccessType](#edgeproto.AccessType) |  | Access type |
| default_privacy_policy | [string](#string) |  | Privacy policy when creating auto cluster |
| delete_prepare | [bool](#bool) |  | Preparing to be deleted |
| auto_prov_policies | [string](#string) | repeated | Auto provisioning policy names |
| template_delimiter | [string](#string) |  | Delimiter to be used for template parsing, defaults to &#34;[[ ]]&#34; |
| skip_hc_ports | [string](#string) |  | Comma separated list of protocol:port pairs that we should not run health check on Should be configured in case app does not always listen on these ports &#34;all&#34; can be specified if no health check to be run for this app Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443 |






<a name="edgeproto.AppAutoProvPolicy"></a>

### AppAutoProvPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_key | [AppKey](#edgeproto.AppKey) |  | App key |
| auto_prov_policy | [string](#string) |  | Auto provisioning policy name |






<a name="edgeproto.AppKey"></a>

### AppKey
Application unique key

AppKey uniquely identifies an App


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [string](#string) |  | App developer organization |
| name | [string](#string) |  | App name |
| version | [string](#string) |  | App version |






<a name="edgeproto.ConfigFile"></a>

### ConfigFile
ConfigFile


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml |
| config | [string](#string) |  | config file contents or URI reference |





 


<a name="edgeproto.AccessType"></a>

### AccessType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ACCESS_TYPE_DEFAULT_FOR_DEPLOYMENT | 0 | Default load balancer or direct based on deployment |
| ACCESS_TYPE_DIRECT | 1 | Direct access with no load balancer |
| ACCESS_TYPE_LOAD_BALANCER | 2 | Access via a load balancer |



<a name="edgeproto.DeleteType"></a>

### DeleteType
DeleteType

DeleteType specifies if AppInst can be auto deleted or not

| Name | Number | Description |
| ---- | ------ | ----------- |
| NO_AUTO_DELETE | 0 | No autodelete |
| AUTO_DELETE | 1 | Autodelete |



<a name="edgeproto.ImageType"></a>

### ImageType
ImageType

ImageType specifies image type of an App

| Name | Number | Description |
| ---- | ------ | ----------- |
| IMAGE_TYPE_UNKNOWN | 0 | Unknown image type |
| IMAGE_TYPE_DOCKER | 1 | Docker container image type compatible either with Docker or Kubernetes |
| IMAGE_TYPE_QCOW | 2 | QCOW2 virtual machine image type |
| IMAGE_TYPE_HELM | 3 | Helm chart is a separate image type |


 

 


<a name="edgeproto.AppApi"></a>

### AppApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateApp | [App](#edgeproto.App) | [Result](#edgeproto.Result) | Create Application. Creates a definition for an application instance for Cloudlet deployment. |
| DeleteApp | [App](#edgeproto.App) | [Result](#edgeproto.Result) | Delete Application. Deletes a definition of an Application instance. Make sure no other application instances exist with that definition. If they do exist, you must delete those Application instances first. |
| UpdateApp | [App](#edgeproto.App) | [Result](#edgeproto.Result) | Update Application. Updates the definition of an Application instance. |
| ShowApp | [App](#edgeproto.App) | [App](#edgeproto.App) stream | Show Applications. Lists all Application definitions managed from the Edge Controller. Any fields specified will be used to filter results. |
| AddAppAutoProvPolicy | [AppAutoProvPolicy](#edgeproto.AppAutoProvPolicy) | [Result](#edgeproto.Result) |  |
| RemoveAppAutoProvPolicy | [AppAutoProvPolicy](#edgeproto.AppAutoProvPolicy) | [Result](#edgeproto.Result) |  |

 



<a name="appinst.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## appinst.proto



<a name="edgeproto.AppInst"></a>

### AppInst
Application Instance

AppInst is an instance of an App on a Cloudlet where it is defined by an App plus a ClusterInst key. 
Many of the fields here are inherited from the App definition.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | required: true Unique identifier key |
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
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |
| revision | [string](#string) |  | Revision changes each time the App is updated. Refreshing the App Instance will sync the revision with that of the App |
| force_update | [bool](#bool) |  | Force Appinst refresh even if revision number matches App revision number. |
| update_multiple | [bool](#bool) |  | Allow multiple instances to be updated at once |
| configs | [ConfigFile](#edgeproto.ConfigFile) | repeated | Customization files passed through to implementing services |
| shared_volume_size | [uint64](#uint64) |  | shared volume size when creating auto cluster |
| health_check | [HealthCheck](#dme.HealthCheck) |  | Health Check status |
| privacy_policy | [string](#string) |  | Optional privacy policy name |
| power_state | [PowerState](#edgeproto.PowerState) |  | Power State of the AppInst |
| external_volume_size | [uint64](#uint64) |  | Size of external volume to be attached to nodes. This is for the root partition |
| availability_zone | [string](#string) |  | Optional Availability Zone if any |
| vm_flavor | [string](#string) |  | OS node flavor to use |
| opt_res | [string](#string) |  | Optional Resources required by OS flavor if any |






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
| runtime_info | [AppInstRuntime](#edgeproto.AppInstRuntime) |  | AppInst runtime information |
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |
| power_state | [PowerState](#edgeproto.PowerState) |  | Power State of the AppInst |






<a name="edgeproto.AppInstKey"></a>

### AppInstKey
App Instance Unique Key

AppInstKey uniquely identifies an Application Instance (AppInst) or Application Instance state (AppInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_key | [AppKey](#edgeproto.AppKey) |  | App key |
| cluster_inst_key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Cluster instance on which this is instantiated |






<a name="edgeproto.AppInstLookup"></a>

### AppInstLookup
AppInstLookup is used to generate reverse lookup caches


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | unique key |
| policy_key | [PolicyKey](#edgeproto.PolicyKey) |  | lookup by AutoProvPolicy |






<a name="edgeproto.AppInstMetrics"></a>

### AppInstMetrics
(TODO) AppInstMetrics provide metrics collected about the application instance on the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| something | [uint64](#uint64) |  | what goes here? Note that metrics for grpc calls can be done by a prometheus interceptor in grpc, so adding call metrics here may be redundant unless they&#39;re needed for billing. |






<a name="edgeproto.AppInstRuntime"></a>

### AppInstRuntime
AppInst Runtime Info

Runtime information of active AppInsts


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| container_ids | [string](#string) | repeated | List of container names |





 


<a name="dme.HealthCheck"></a>

### HealthCheck
Health check status

Health check status gets set by external, or rootLB health check

| Name | Number | Description |
| ---- | ------ | ----------- |
| HEALTH_CHECK_UNKNOWN | 0 | Health Check is unknown |
| HEALTH_CHECK_FAIL_ROOTLB_OFFLINE | 1 | Health Check failure due to RootLB being offline |
| HEALTH_CHECK_FAIL_SERVER_FAIL | 2 | Health Check failure due to Backend server being unavailable |
| HEALTH_CHECK_OK | 3 | Health Check is ok |



<a name="edgeproto.PowerState"></a>

### PowerState
Power State

Power State of the AppInst

| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_STATE_UNKNOWN | 0 | Unknown |
| POWER_ON_REQUESTED | 1 | Power On Requested |
| POWERING_ON | 2 | Powering On |
| POWER_ON | 3 | Power On |
| POWER_OFF_REQUESTED | 4 | Power Off Requested |
| POWERING_OFF | 5 | Powering Off |
| POWER_OFF | 6 | Power Off |
| REBOOT_REQUESTED | 7 | Reboot Requested |
| REBOOTING | 8 | Rebooting |
| REBOOT | 9 | Reboot |
| POWER_STATE_ERROR | 10 | Error |


 

 


<a name="edgeproto.AppInstApi"></a>

### AppInstApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Create Application Instance. Creates an instance of an App on a Cloudlet where it is defined by an App plus a ClusterInst key. Many of the fields here are inherited from the App definition. |
| DeleteAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Delete Application Instance. Deletes an instance of the App from the Cloudlet. |
| RefreshAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Refresh Application Instance. Restarts an App instance with new App settings or image. |
| UpdateAppInst | [AppInst](#edgeproto.AppInst) | [Result](#edgeproto.Result) stream | Update Application Instance. Updates an Application instance and then refreshes it. |
| ShowAppInst | [AppInst](#edgeproto.AppInst) | [AppInst](#edgeproto.AppInst) stream | Show Application Instances. Lists all the Application instances managed by the Edge Controller. Any fields specified will be used to filter results. |


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

 



<a name="appinstclient.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## appinstclient.proto



<a name="edgeproto.AppInstClient"></a>

### AppInstClient
Client is an AppInst client that called FindCloudlet DME Api


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| client_key | [AppInstClientKey](#edgeproto.AppInstClientKey) |  | Unique identifier key |
| location | [distributed_match_engine.Loc](#distributed_match_engine.Loc) |  | Location of the Client |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |






<a name="edgeproto.AppInstClientKey"></a>

### AppInstClientKey
AppKey uniquely identifies an App


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [AppInstKey](#edgeproto.AppInstKey) |  | AppInst Key |
| unique_id | [string](#string) |  | AppInstClient Unique Id |
| unique_id_type | [string](#string) |  | AppInstClient Unique Id Type |





 

 

 


<a name="edgeproto.AppInstClientApi"></a>

### AppInstClientApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAppInstClient | [AppInstClientKey](#edgeproto.AppInstClientKey) | [AppInstClient](#edgeproto.AppInstClient) stream | Show application instance clients. |
| StreamAppInstClientsLocal | [AppInstClientKey](#edgeproto.AppInstClientKey) | [AppInstClient](#edgeproto.AppInstClient) stream | This is used unternally to forward AppInstClients to other Controllers |

 



<a name="autoprovpolicy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## autoprovpolicy.proto



<a name="edgeproto.AutoProvCloudlet"></a>

### AutoProvCloudlet
AutoProvCloudlet stores the potential cloudlet and location for DME lookup


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet key |
| loc | [distributed_match_engine.Loc](#distributed_match_engine.Loc) |  | Cloudlet location |






<a name="edgeproto.AutoProvCount"></a>

### AutoProvCount
AutoProvCount is used to send potential cloudlet and location counts from DME to Controller


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_key | [AppKey](#edgeproto.AppKey) |  | Target app |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Target cloudlet |
| count | [uint64](#uint64) |  | FindCloudlet client count |
| process_now | [bool](#bool) |  | Process count immediately |
| deploy_now_key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | Immediately deploy to clusterinst |






<a name="edgeproto.AutoProvCounts"></a>

### AutoProvCounts
AutoProvCounts is used to send potential cloudlet and location counts from DME to Controller


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dme_node_name | [string](#string) |  | DME node name |
| timestamp | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the metric was captured |
| counts | [AutoProvCount](#edgeproto.AutoProvCount) | repeated | List of DmeCount from DME |






<a name="edgeproto.AutoProvInfo"></a>

### AutoProvInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet Key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| maintenance_state | [MaintenanceState](#dme.MaintenanceState) |  | failover result state |
| completed | [string](#string) | repeated | Failover actions done if any |
| errors | [string](#string) | repeated | Errors if any |






<a name="edgeproto.AutoProvPolicy"></a>

### AutoProvPolicy
AutoProvPolicy defines the automated provisioning policy


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [PolicyKey](#edgeproto.PolicyKey) |  | Unique identifier key |
| deploy_client_count | [uint32](#uint32) |  | Minimum number of clients within the auto deploy interval to trigger deployment |
| deploy_interval_count | [uint32](#uint32) |  | Number of intervals to check before triggering deployment |
| cloudlets | [AutoProvCloudlet](#edgeproto.AutoProvCloudlet) | repeated | Allowed deployment locations |
| min_active_instances | [uint32](#uint32) |  | Minimum number of active instances for High-Availability |
| max_instances | [uint32](#uint32) |  | Maximum number of instances (active or not) |
| undeploy_client_count | [uint32](#uint32) |  | Number of active clients for the undeploy interval below which trigers undeployment, 0 (default) disables auto undeploy |
| undeploy_interval_count | [uint32](#uint32) |  | Number of intervals to check before triggering undeployment |






<a name="edgeproto.AutoProvPolicyCloudlet"></a>

### AutoProvPolicyCloudlet
AutoProvPolicyCloudlet is used to add and remove Cloudlets from the Auto Provisioning Policy


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [PolicyKey](#edgeproto.PolicyKey) |  | Unique policy identifier key |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet identifier key |





 

 

 


<a name="edgeproto.AutoProvPolicyApi"></a>

### AutoProvPolicyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateAutoProvPolicy | [AutoProvPolicy](#edgeproto.AutoProvPolicy) | [Result](#edgeproto.Result) | Create an Auto Provisioning Policy |
| DeleteAutoProvPolicy | [AutoProvPolicy](#edgeproto.AutoProvPolicy) | [Result](#edgeproto.Result) | Delete an Auto Provisioning Policy |
| UpdateAutoProvPolicy | [AutoProvPolicy](#edgeproto.AutoProvPolicy) | [Result](#edgeproto.Result) | Update an Auto Provisioning Policy |
| ShowAutoProvPolicy | [AutoProvPolicy](#edgeproto.AutoProvPolicy) | [AutoProvPolicy](#edgeproto.AutoProvPolicy) stream | Show Auto Provisioning Policies. Any fields specified will be used to filter results. |
| AddAutoProvPolicyCloudlet | [AutoProvPolicyCloudlet](#edgeproto.AutoProvPolicyCloudlet) | [Result](#edgeproto.Result) | Add a Cloudlet to the Auto Provisioning Policy |
| RemoveAutoProvPolicyCloudlet | [AutoProvPolicyCloudlet](#edgeproto.AutoProvPolicyCloudlet) | [Result](#edgeproto.Result) | Remove a Cloudlet from the Auto Provisioning Policy |

 



<a name="autoscalepolicy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## autoscalepolicy.proto



<a name="edgeproto.AutoScalePolicy"></a>

### AutoScalePolicy
AutoScalePolicy defines when and how ClusterInsts will have their
nodes scaled up or down.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [PolicyKey](#edgeproto.PolicyKey) |  | Unique identifier key |
| min_nodes | [uint32](#uint32) |  | Minimum number of cluster nodes |
| max_nodes | [uint32](#uint32) |  | Maximum number of cluster nodes |
| scale_up_cpu_thresh | [uint32](#uint32) |  | Scale up cpu threshold (percentage 1 to 100) |
| scale_down_cpu_thresh | [uint32](#uint32) |  | Scale down cpu threshold (percentage 1 to 100) |
| trigger_time_sec | [uint32](#uint32) |  | Trigger time defines how long trigger threshold must be satified in seconds before acting upon it. |






<a name="edgeproto.PolicyKey"></a>

### PolicyKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [string](#string) |  | Name of the organization for the cluster that this policy will apply to |
| name | [string](#string) |  | Policy name |





 

 

 


<a name="edgeproto.AutoScalePolicyApi"></a>

### AutoScalePolicyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateAutoScalePolicy | [AutoScalePolicy](#edgeproto.AutoScalePolicy) | [Result](#edgeproto.Result) | Create an Auto Scale Policy |
| DeleteAutoScalePolicy | [AutoScalePolicy](#edgeproto.AutoScalePolicy) | [Result](#edgeproto.Result) | Delete an Auto Scale Policy |
| UpdateAutoScalePolicy | [AutoScalePolicy](#edgeproto.AutoScalePolicy) | [Result](#edgeproto.Result) | Update an Auto Scale Policy |
| ShowAutoScalePolicy | [AutoScalePolicy](#edgeproto.AutoScalePolicy) | [AutoScalePolicy](#edgeproto.AutoScalePolicy) stream | Show Auto Scale Policies. Any fields specified will be used to filter results. |

 



<a name="cloudlet.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cloudlet.proto



<a name="edgeproto.Cloudlet"></a>

### Cloudlet
Cloudlet

A Cloudlet is a set of compute resources at a particular location, provided by an Operator.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | required: true Unique identifier key |
| location | [distributed_match_engine.Loc](#distributed_match_engine.Loc) |  | Location of the Cloudlet site |
| ip_support | [IpSupport](#edgeproto.IpSupport) |  | Type of IP support provided by Cloudlet (see IpSupport) |
| static_ips | [string](#string) |  | List of static IPs for static IP support |
| num_dynamic_ips | [int32](#int32) |  | Number of dynamic IPs available for dynamic IP support |
| time_limits | [OperationTimeLimits](#edgeproto.OperationTimeLimits) |  | time limits which override global settings if non-zero |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the Cloudlet. |
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the cloudlet |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |
| deployment_local | [bool](#bool) |  | Deploy cloudlet services locally |
| platform_type | [PlatformType](#edgeproto.PlatformType) |  | Platform type |
| notify_srv_addr | [string](#string) |  | Address for the CRM notify listener to run on |
| flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Min system resource requirements for platform |
| physical_name | [string](#string) |  | Physical infrastructure cloudlet name |
| env_var | [Cloudlet.EnvVarEntry](#edgeproto.Cloudlet.EnvVarEntry) | repeated | Single Key-Value pair of env var to be passed to CRM |
| container_version | [string](#string) |  | Cloudlet container version |
| config | [PlatformConfig](#edgeproto.PlatformConfig) |  | Platform Config Info |
| res_tag_map | [Cloudlet.ResTagMapEntry](#edgeproto.Cloudlet.ResTagMapEntry) | repeated | Optional resource to restagtbl key map key values = [gpu, nas, nic] |
| access_vars | [Cloudlet.AccessVarsEntry](#edgeproto.Cloudlet.AccessVarsEntry) | repeated | Variables required to access cloudlet |
| vm_image_version | [string](#string) |  | MobiledgeX baseimage version where CRM services reside |
| deployment | [string](#string) |  | Deployment type to bring up CRM services (docker, kubernetes) |
| infra_api_access | [InfraApiAccess](#edgeproto.InfraApiAccess) |  | Infra Access Type is the type of access available to Infra API Endpoint |
| infra_config | [InfraConfig](#edgeproto.InfraConfig) |  | Infra specific config |
| chef_client_key | [Cloudlet.ChefClientKeyEntry](#edgeproto.Cloudlet.ChefClientKeyEntry) | repeated | Chef client key |
| maintenance_state | [MaintenanceState](#dme.MaintenanceState) |  | State for maintenance |
| override_policy_container_version | [bool](#bool) |  | Override container version from policy file |
| vm_pool | [string](#string) |  | VM Pool |






<a name="edgeproto.Cloudlet.AccessVarsEntry"></a>

### Cloudlet.AccessVarsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.Cloudlet.ChefClientKeyEntry"></a>

### Cloudlet.ChefClientKeyEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.Cloudlet.EnvVarEntry"></a>

### Cloudlet.EnvVarEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.Cloudlet.ResTagMapEntry"></a>

### Cloudlet.ResTagMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ResTagTableKey](#edgeproto.ResTagTableKey) |  |  |






<a name="edgeproto.CloudletInfo"></a>

### CloudletInfo
CloudletInfo provides information from the Cloudlet Resource Manager about the state of the Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Unique identifier key |
| state | [CloudletState](#dme.CloudletState) |  | State of cloudlet |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| controller | [string](#string) |  | Connected controller unique id |
| os_max_ram | [uint64](#uint64) |  | Maximum Ram in MB on the Cloudlet |
| os_max_vcores | [uint64](#uint64) |  | Maximum number of VCPU cores on the Cloudlet |
| os_max_vol_gb | [uint64](#uint64) |  | Maximum amount of disk in GB on the Cloudlet |
| errors | [string](#string) | repeated | Any errors encountered while making changes to the Cloudlet |
| flavors | [FlavorInfo](#edgeproto.FlavorInfo) | repeated | Supported flavors by the Cloudlet |
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |
| container_version | [string](#string) |  | Cloudlet container version |
| availability_zones | [OSAZone](#edgeproto.OSAZone) | repeated | Availability Zones if any |
| os_images | [OSImage](#edgeproto.OSImage) | repeated | Local Images availble to cloudlet |
| controller_cache_received | [bool](#bool) |  | Indicates all controller data has been sent to CRM |
| maintenance_state | [MaintenanceState](#dme.MaintenanceState) |  | State for maintenance |






<a name="edgeproto.CloudletKey"></a>

### CloudletKey
Cloudlet unique key

CloudletKey uniquely identifies a Cloudlet.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [string](#string) |  | Organization of the cloudlet site |
| name | [string](#string) |  | Name of the cloudlet |






<a name="edgeproto.CloudletManifest"></a>

### CloudletManifest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manifest | [string](#string) |  | Manifest to bringup cloudlet VM and services. |






<a name="edgeproto.CloudletMetrics"></a>

### CloudletMetrics
(TODO) CloudletMetrics provide metrics collected about the Cloudlet. They are sent to a metrics collector for analytics. They are not stored in the persistent distributed database, but are stored as a time series in some other database or files.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| foo | [uint64](#uint64) |  | what goes here? |






<a name="edgeproto.CloudletProps"></a>

### CloudletProps



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platform_type | [PlatformType](#edgeproto.PlatformType) |  | Platform type |
| properties | [CloudletProps.PropertiesEntry](#edgeproto.CloudletProps.PropertiesEntry) | repeated | Single Key-Value pair of env var to be passed to CRM |






<a name="edgeproto.CloudletProps.PropertiesEntry"></a>

### CloudletProps.PropertiesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [PropertyInfo](#edgeproto.PropertyInfo) |  |  |






<a name="edgeproto.CloudletResMap"></a>

### CloudletResMap
optional resource input consists of a resource specifier and clouldkey name


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Resource cloudlet key |
| mapping | [CloudletResMap.MappingEntry](#edgeproto.CloudletResMap.MappingEntry) | repeated | Resource mapping info |






<a name="edgeproto.CloudletResMap.MappingEntry"></a>

### CloudletResMap.MappingEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.FlavorInfo"></a>

### FlavorInfo
Flavor details from the Cloudlet


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the flavor on the Cloudlet |
| vcpus | [uint64](#uint64) |  | Number of VCPU cores on the Cloudlet |
| ram | [uint64](#uint64) |  | Ram in MB on the Cloudlet |
| disk | [uint64](#uint64) |  | Amount of disk in GB on the Cloudlet |
| prop_map | [FlavorInfo.PropMapEntry](#edgeproto.FlavorInfo.PropMapEntry) | repeated | OS Flavor Properties, if any |






<a name="edgeproto.FlavorInfo.PropMapEntry"></a>

### FlavorInfo.PropMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.FlavorMatch"></a>

### FlavorMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet ctx |
| flavor_name | [string](#string) |  |  |
| availability_zone | [string](#string) |  |  |






<a name="edgeproto.InfraConfig"></a>

### InfraConfig
Infra specific configuration used for Cloudlet deployments


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| external_network_name | [string](#string) |  | Infra specific external network name |
| flavor_name | [string](#string) |  | Infra specific flavor name |






<a name="edgeproto.OSAZone"></a>

### OSAZone



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| status | [string](#string) |  |  |






<a name="edgeproto.OSImage"></a>

### OSImage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | image name |
| tags | [string](#string) |  | optional tags present on image |
| properties | [string](#string) |  | image properties/metadata |
| disk_format | [string](#string) |  | format qcow2, img, etc |






<a name="edgeproto.OperationTimeLimits"></a>

### OperationTimeLimits
Operation time limits

Time limits for cloudlet create, update and delete operations


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_cluster_inst_timeout | [int64](#int64) |  | override default max time to create a cluster instance (duration) |
| update_cluster_inst_timeout | [int64](#int64) |  | override default max time to update a cluster instance (duration) |
| delete_cluster_inst_timeout | [int64](#int64) |  | override default max time to delete a cluster instance (duration) |
| create_app_inst_timeout | [int64](#int64) |  | override default max time to create an app instance (duration) |
| update_app_inst_timeout | [int64](#int64) |  | override default max time to update an app instance (duration) |
| delete_app_inst_timeout | [int64](#int64) |  | override default max time to delete an app instance (duration) |






<a name="edgeproto.PlatformConfig"></a>

### PlatformConfig
Platform specific configuration required for Cloudlet management


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| container_registry_path | [string](#string) |  | Path to Docker registry holding edge-cloud image |
| cloudlet_vm_image_path | [string](#string) |  | Path to platform base image |
| notify_ctrl_addrs | [string](#string) |  | Address of controller notify port (can be multiple of these) |
| vault_addr | [string](#string) |  | Vault address |
| tls_cert_file | [string](#string) |  | TLS cert file |
| tls_key_file | [string](#string) |  | TLS key file |
| tls_ca_file | [string](#string) |  | TLS ca file |
| env_var | [PlatformConfig.EnvVarEntry](#edgeproto.PlatformConfig.EnvVarEntry) | repeated | Environment variables |
| platform_tag | [string](#string) |  | Tag of edge-cloud image |
| test_mode | [bool](#bool) |  | Internal Test flag |
| span | [string](#string) |  | Span string |
| cleanup_mode | [bool](#bool) |  | Internal cleanup flag |
| region | [string](#string) |  | Region |
| commercial_certs | [bool](#bool) |  | Get certs from vault or generate your own for the root load balancer |
| use_vault_certs | [bool](#bool) |  | Use Vault certs for internal TLS communication |
| use_vault_cas | [bool](#bool) |  | Use Vault CAs to authenticate TLS communication |
| app_dns_root | [string](#string) |  | App domain name root |
| chef_server_path | [string](#string) |  | Path to Chef Server |
| chef_client_interval | [int32](#int32) |  | Chef client interval |
| deployment_tag | [string](#string) |  | Deployment Tag |






<a name="edgeproto.PlatformConfig.EnvVarEntry"></a>

### PlatformConfig.EnvVarEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.PropertyInfo"></a>

### PropertyInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the property |
| description | [string](#string) |  | Description of the property |
| value | [string](#string) |  | Default value of the property |
| secret | [bool](#bool) |  | Is the property a secret value, will be hidden |
| mandatory | [bool](#bool) |  | Is the property mandatory |
| internal | [bool](#bool) |  | Is the property internal, not to be set by Operator |





 


<a name="dme.CloudletState"></a>

### CloudletState
CloudletState is the state of the Cloudlet.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CLOUDLET_STATE_UNKNOWN | 0 | Unknown |
| CLOUDLET_STATE_ERRORS | 1 | Create/Delete/Update encountered errors (see Errors field of CloudletInfo) |
| CLOUDLET_STATE_READY | 2 | Cloudlet is created and ready |
| CLOUDLET_STATE_OFFLINE | 3 | Cloudlet is offline (unreachable) |
| CLOUDLET_STATE_NOT_PRESENT | 4 | Cloudlet is not present |
| CLOUDLET_STATE_INIT | 5 | Cloudlet is initializing |
| CLOUDLET_STATE_UPGRADE | 6 | Cloudlet is upgrading |
| CLOUDLET_STATE_NEED_SYNC | 7 | Cloudlet needs data to synchronize |



<a name="edgeproto.InfraApiAccess"></a>

### InfraApiAccess
Infra API Access

InfraApiAccess is the type of access available to Infra API endpoint

| Name | Number | Description |
| ---- | ------ | ----------- |
| DIRECT_ACCESS | 0 | Infra API endpoint is accessible from public network |
| RESTRICTED_ACCESS | 1 | Infra API endpoint is not accessible from public network |



<a name="edgeproto.PlatformType"></a>

### PlatformType
Platform Type

PlatformType is the supported list of cloudlet types

| Name | Number | Description |
| ---- | ------ | ----------- |
| PLATFORM_TYPE_FAKE | 0 | Fake Cloudlet |
| PLATFORM_TYPE_DIND | 1 | DIND Cloudlet |
| PLATFORM_TYPE_OPENSTACK | 2 | Openstack Cloudlet |
| PLATFORM_TYPE_AZURE | 3 | Azure Cloudlet |
| PLATFORM_TYPE_GCP | 4 | GCP Cloudlet |
| PLATFORM_TYPE_EDGEBOX | 5 | Edgebox Cloudlet |
| PLATFORM_TYPE_FAKEINFRA | 6 | Fake Infra Cloudlet |
| PLATFORM_TYPE_VSPHERE | 7 | VMWare VSphere (ESXi) |
| PLATFORM_TYPE_AWS | 8 | AWS Cloudlet |
| PLATFORM_TYPE_VM_POOL | 9 | VM Pool Cloudlet |


 

 


<a name="edgeproto.CloudletApi"></a>

### CloudletApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Result) stream | Create Cloudlet. Sets up Cloudlet services on the Operator&#39;s compute resources, and integrated as part of MobiledgeX edge resource portfolio. These resources are managed from the Edge Controller. |
| DeleteCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Result) stream | Delete Cloudlet. Removes the Cloudlet services where they are no longer managed from the Edge Controller. |
| UpdateCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Result](#edgeproto.Result) stream | Update Cloudlet. Updates the Cloudlet configuration and manages the upgrade of Cloudlet services. |
| ShowCloudlet | [Cloudlet](#edgeproto.Cloudlet) | [Cloudlet](#edgeproto.Cloudlet) stream | Show Cloudlets. Lists all the cloudlets managed from Edge Controller. |
| GetCloudletManifest | [Cloudlet](#edgeproto.Cloudlet) | [CloudletManifest](#edgeproto.CloudletManifest) | Get Cloudlet Manifest. Shows deployment manifest required to setup cloudlet |
| GetCloudletProps | [CloudletProps](#edgeproto.CloudletProps) | [CloudletProps](#edgeproto.CloudletProps) | Get Cloudlet Properties. Shows all the infra properties used to setup cloudlet |
| AddCloudletResMapping | [CloudletResMap](#edgeproto.CloudletResMap) | [Result](#edgeproto.Result) | Add Optional Resource tag table |
| RemoveCloudletResMapping | [CloudletResMap](#edgeproto.CloudletResMap) | [Result](#edgeproto.Result) | Add Optional Resource tag table |
| FindFlavorMatch | [FlavorMatch](#edgeproto.FlavorMatch) | [FlavorMatch](#edgeproto.FlavorMatch) | Discover if flavor produces a matching platform flavor |


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

 



<a name="cloudletpool.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cloudletpool.proto



<a name="edgeproto.CloudletPool"></a>

### CloudletPool
CloudletPool defines a pool of Cloudlets that have restricted access.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [CloudletPoolKey](#edgeproto.CloudletPoolKey) |  | CloudletPool key |
| cloudlets | [string](#string) | repeated | Cloudlets part of the pool |






<a name="edgeproto.CloudletPoolKey"></a>

### CloudletPoolKey
CloudletPool unique key

CloudletPoolKey uniquely identifies a CloudletPool.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [string](#string) |  | Name of the organization this pool belongs to |
| name | [string](#string) |  | CloudletPool Name |






<a name="edgeproto.CloudletPoolMember"></a>

### CloudletPoolMember
CloudletPoolMember is used to add and remove a Cloudlet from a CloudletPool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [CloudletPoolKey](#edgeproto.CloudletPoolKey) |  | CloudletPool key |
| cloudlet_name | [string](#string) |  | Cloudlet key |





 

 

 


<a name="edgeproto.CloudletPoolApi"></a>

### CloudletPoolApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCloudletPool | [CloudletPool](#edgeproto.CloudletPool) | [Result](#edgeproto.Result) | Create a CloudletPool |
| DeleteCloudletPool | [CloudletPool](#edgeproto.CloudletPool) | [Result](#edgeproto.Result) | Delete a CloudletPool |
| UpdateCloudletPool | [CloudletPool](#edgeproto.CloudletPool) | [Result](#edgeproto.Result) | Update a CloudletPool |
| ShowCloudletPool | [CloudletPool](#edgeproto.CloudletPool) | [CloudletPool](#edgeproto.CloudletPool) stream | Show CloudletPools |
| AddCloudletPoolMember | [CloudletPoolMember](#edgeproto.CloudletPoolMember) | [Result](#edgeproto.Result) | Add a Cloudlet to a CloudletPool |
| RemoveCloudletPoolMember | [CloudletPoolMember](#edgeproto.CloudletPoolMember) | [Result](#edgeproto.Result) | Remove a Cloudlet from a CloudletPool |

 



<a name="cluster.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cluster.proto



<a name="edgeproto.ClusterKey"></a>

### ClusterKey
ClusterKey uniquely identifies a Cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Cluster name |





 

 

 

 



<a name="clusterinst.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## clusterinst.proto



<a name="edgeproto.ClusterInst"></a>

### ClusterInst
Cluster Instance

ClusterInst is an instance of a Cluster on a Cloudlet. 
It is defined by a Cluster, Cloudlet, and Developer key.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [ClusterInstKey](#edgeproto.ClusterInstKey) |  | required: true Unique key |
| flavor | [FlavorKey](#edgeproto.FlavorKey) |  | Flavor of the k8s node |
| liveness | [Liveness](#edgeproto.Liveness) |  | Liveness of instance (see Liveness) |
| auto | [bool](#bool) |  | Auto is set to true when automatically created by back-end (internal use only) |
| state | [TrackedState](#edgeproto.TrackedState) |  | State of the cluster instance |
| errors | [string](#string) | repeated | Any errors trying to create, update, or delete the ClusterInst on the Cloudlet. |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |
| ip_access | [IpAccess](#edgeproto.IpAccess) |  | IP access type (RootLB Type) |
| allocated_ip | [string](#string) |  | Allocated IP for dedicated access |
| node_flavor | [string](#string) |  | Cloudlet specific node flavor |
| deployment | [string](#string) |  | Deployment type (kubernetes or docker) |
| num_masters | [uint32](#uint32) |  | Number of k8s masters (In case of docker deployment, this field is not required) |
| num_nodes | [uint32](#uint32) |  | Number of k8s nodes (In case of docker deployment, this field is not required) |
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |
| external_volume_size | [uint64](#uint64) |  | Size of external volume to be attached to nodes. This is for the root partition |
| auto_scale_policy | [string](#string) |  | Auto scale policy name |
| availability_zone | [string](#string) |  | Optional Resource AZ if any |
| image_name | [string](#string) |  | Optional resource specific image to launch |
| reservable | [bool](#bool) |  | If ClusterInst is reservable |
| reserved_by | [string](#string) |  | For reservable MobiledgeX ClusterInsts, the current developer tenant |
| shared_volume_size | [uint64](#uint64) |  | Size of an optional shared volume to be mounted on the master |
| privacy_policy | [string](#string) |  | Optional privacy policy name |
| master_node_flavor | [string](#string) |  | Generic flavor for k8s master VM when worker nodes &gt; 0 |
| skip_crm_cleanup_on_failure | [bool](#bool) |  | Prevents cleanup of resources on failure within CRM, used for diagnostic purposes |
| opt_res | [string](#string) |  | Optional Resources required by OS flavor if any |






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
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |






<a name="edgeproto.ClusterInstKey"></a>

### ClusterInstKey
Cluster Instance unique key

ClusterInstKey uniquely identifies a Cluster Instance (ClusterInst) or Cluster Instance state (ClusterInstInfo).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_key | [ClusterKey](#edgeproto.ClusterKey) |  | Name of Cluster |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Name of Cloudlet on which the Cluster is instantiated |
| organization | [string](#string) |  | Name of Developer organization that this cluster belongs to |





 

 

 


<a name="edgeproto.ClusterInstApi"></a>

### ClusterInstApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.Result) stream | Create Cluster Instance. Creates an instance of a Cluster on a Cloudlet, defined by a Cluster Key and a Cloudlet Key. ClusterInst is a collection of compute resources on a Cloudlet on which AppInsts are deployed. |
| DeleteClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.Result) stream | Delete Cluster Instance. Deletes an instance of a Cluster deployed on a Cloudlet. |
| UpdateClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [Result](#edgeproto.Result) stream | Update Cluster Instance. Updates an instance of a Cluster deployed on a Cloudlet. |
| ShowClusterInst | [ClusterInst](#edgeproto.ClusterInst) | [ClusterInst](#edgeproto.ClusterInst) stream | Show Cluster Instances. Lists all the cluster instances managed by Edge Controller. |


<a name="edgeproto.ClusterInstInfoApi"></a>

### ClusterInstInfoApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowClusterInstInfo | [ClusterInstInfo](#edgeproto.ClusterInstInfo) | [ClusterInstInfo](#edgeproto.ClusterInstInfo) stream | Show Cluster instances state. |

 



<a name="common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## common.proto



<a name="edgeproto.StatusInfo"></a>

### StatusInfo
Status Information

Used to track status of create/delete/update for resources that are being modified 
by the controller via the CRM.  Tasks are the high level jobs that are to be completed.
Steps are work items within a task. Within the clusterinst and appinst objects this
is converted to a string


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_number | [uint32](#uint32) |  |  |
| max_tasks | [uint32](#uint32) |  |  |
| task_name | [string](#string) |  |  |
| step_name | [string](#string) |  |  |





 


<a name="edgeproto.CRMOverride"></a>

### CRMOverride
Overrides default CRM behaviour

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



<a name="edgeproto.IpAccess"></a>

### IpAccess
IpAccess Options

IpAccess indicates the type of RootLB that Developer requires for their App

| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_ACCESS_UNKNOWN | 0 | Unknown IP access |
| IP_ACCESS_DEDICATED | 1 | Dedicated RootLB |
| IP_ACCESS_SHARED | 3 | Shared RootLB |



<a name="edgeproto.IpSupport"></a>

### IpSupport
Type of public IP support

Static IP support indicates a set of static public IPs are available for use, and managed by the Controller. Dynamic indicates the Cloudlet uses a DHCP server to provide public IP addresses, and the controller has no control over which IPs are assigned.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_SUPPORT_UNKNOWN | 0 | Unknown IP support |
| IP_SUPPORT_STATIC | 1 | Static IP addresses are provided to and managed by Controller |
| IP_SUPPORT_DYNAMIC | 2 | IP addresses are dynamically provided by an Operator&#39;s DHCP server |



<a name="edgeproto.Liveness"></a>

### Liveness
Liveness Options

Liveness indicates if an object was created statically via an external API call, or dynamically via an internal algorithm.

| Name | Number | Description |
| ---- | ------ | ----------- |
| LIVENESS_UNKNOWN | 0 | Unknown liveness |
| LIVENESS_STATIC | 1 | Object managed by external entity |
| LIVENESS_DYNAMIC | 2 | Object managed internally |
| LIVENESS_AUTOPROV | 3 | Object created by Auto Provisioning, treated like Static except when deleting App |



<a name="dme.MaintenanceState"></a>

### MaintenanceState
Cloudlet Maintenance States

Maintenance allows for planned downtimes of Cloudlets.
These states involve message exchanges between the Controller,
the AutoProv service, and the CRM. Certain states are only set
by certain actors.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NORMAL_OPERATION | 0 | Normal operational state |
| MAINTENANCE_START | 1 | Request start of maintenance |
| FAILOVER_REQUESTED | 2 | Trigger failover for any HA AppInsts |
| FAILOVER_DONE | 3 | Failover done |
| FAILOVER_ERROR | 4 | Some errors encountered during maintenance failover |
| MAINTENANCE_START_NO_FAILOVER | 5 | Request start of maintenance without AutoProv failover |
| CRM_REQUESTED | 6 | Request CRM to transition to maintenance |
| CRM_UNDER_MAINTENANCE | 7 | CRM request done and under maintenance |
| CRM_ERROR | 8 | CRM failed to go into maintenance |
| UNDER_MAINTENANCE | 31 | Under maintenance |



<a name="edgeproto.TrackedState"></a>

### TrackedState
Tracked States

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
| CRM_INITOK | 13 | CRM INIT OK |
| CREATING_DEPENDENCIES | 14 | Creating dependencies (state used to tracked dependent object change progress) |


 

 

 



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
| build_master | [string](#string) |  | Build Master Version |
| build_head | [string](#string) |  | Build Head Version |
| build_author | [string](#string) |  | Build Author |
| hostname | [string](#string) |  | Hostname |






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

 



<a name="debug.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## debug.proto



<a name="edgeproto.DebugData"></a>

### DebugData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requests | [DebugRequest](#edgeproto.DebugRequest) | repeated |  |






<a name="edgeproto.DebugReply"></a>

### DebugReply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node | [NodeKey](#edgeproto.NodeKey) |  | Service node identifier (see NodeShow) |
| output | [string](#string) |  | Debug output, if any |
| id | [uint64](#uint64) |  | Id used internally |






<a name="edgeproto.DebugRequest"></a>

### DebugRequest
DebugRequest. Keep everything in one struct to make it easy to send commands without having to change the code.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node | [NodeKey](#edgeproto.NodeKey) |  | Service node identifier (see NodeShow) |
| levels | [string](#string) |  | Comma separated list of debug level names: etcd,api,notify,dmereq,locapi,infra,metrics,upgrade,info,sampled |
| cmd | [string](#string) |  | Debug command (use &#34;help&#34; to see available commands) |
| pretty | [bool](#bool) |  | if possible, make output pretty |
| id | [uint64](#uint64) |  | Id used internally |
| args | [string](#string) |  | Additional arguments for cmd |
| timeout | [int64](#int64) |  | custom timeout (duration, defaults to 10s) |





 

 

 


<a name="edgeproto.DebugApi"></a>

### DebugApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| EnableDebugLevels | [DebugRequest](#edgeproto.DebugRequest) | [DebugReply](#edgeproto.DebugReply) stream |  |
| DisableDebugLevels | [DebugRequest](#edgeproto.DebugRequest) | [DebugReply](#edgeproto.DebugReply) stream |  |
| ShowDebugLevels | [DebugRequest](#edgeproto.DebugRequest) | [DebugReply](#edgeproto.DebugReply) stream |  |
| RunDebug | [DebugRequest](#edgeproto.DebugRequest) | [DebugReply](#edgeproto.DebugReply) stream |  |

 



<a name="device.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## device.proto



<a name="edgeproto.Device"></a>

### Device
Device represents a device on the MobiledgeX platform
We record when this device first showed up on our platform


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated |  |
| key | [DeviceKey](#edgeproto.DeviceKey) |  | Key |
| first_seen | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the device was registered |
| last_seen | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the device was last seen(Future use) |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |






<a name="edgeproto.DeviceData"></a>

### DeviceData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| devices | [Device](#edgeproto.Device) | repeated |  |






<a name="edgeproto.DeviceKey"></a>

### DeviceKey
DeviceKey is an identifier for a given device on the MobiledgeX platform
It is defined by a unique id and unique id type
And example of such a device is a MEL device that hosts several applications


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| unique_id_type | [string](#string) |  | Type of unique ID provided by the client |
| unique_id | [string](#string) |  | Unique identification of the client device or user. May be overridden by the server. |






<a name="edgeproto.DeviceReport"></a>

### DeviceReport
DeviceReport is a reporting message. It takes a begining and end time
for the report


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [DeviceKey](#edgeproto.DeviceKey) |  | Device Key |
| begin | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp of the beginning of the report |
| end | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp of the beginning of the report |





 

 

 


<a name="edgeproto.DeviceApi"></a>

### DeviceApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| InjectDevice | [Device](#edgeproto.Device) | [Result](#edgeproto.Result) |  |
| ShowDevice | [Device](#edgeproto.Device) | [Device](#edgeproto.Device) stream |  |
| EvictDevice | [Device](#edgeproto.Device) | [Result](#edgeproto.Result) |  |
| ShowDeviceReport | [DeviceReport](#edgeproto.DeviceReport) | [Device](#edgeproto.Device) stream | Device Reports API. |

 



<a name="exec.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## exec.proto



<a name="edgeproto.CloudletMgmtNode"></a>

### CloudletMgmtNode



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | Type of Cloudlet Mgmt Node |
| name | [string](#string) |  | Name of Cloudlet Mgmt Node |






<a name="edgeproto.ExecRequest"></a>

### ExecRequest
ExecRequest is a common struct for enabling a connection to execute some work on a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_inst_key | [AppInstKey](#edgeproto.AppInstKey) |  | Target AppInst |
| container_id | [string](#string) |  | ContainerId is the name or ID of the target container, if applicable |
| offer | [string](#string) |  | Offer |
| answer | [string](#string) |  | Answer |
| err | [string](#string) |  | Any error message |
| cmd | [RunCmd](#edgeproto.RunCmd) |  | Command to run (one of) |
| log | [ShowLog](#edgeproto.ShowLog) |  | Show log (one of) |
| console | [RunVMConsole](#edgeproto.RunVMConsole) |  | Console (one of) |
| timeout | [int64](#int64) |  | Timeout |
| access_url | [string](#string) |  | Access URL |
| edge_turn_addr | [string](#string) |  | EdgeTurn Server Address |






<a name="edgeproto.RunCmd"></a>

### RunCmd



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| command | [string](#string) |  | Command or Shell |
| cloudlet_mgmt_node | [CloudletMgmtNode](#edgeproto.CloudletMgmtNode) |  | Cloudlet Mgmt Node |






<a name="edgeproto.RunVMConsole"></a>

### RunVMConsole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  | VM Console URL |






<a name="edgeproto.ShowLog"></a>

### ShowLog



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| since | [string](#string) |  | Show logs since either a duration ago (5s, 2m, 3h) or a timestamp (RFC3339) |
| tail | [int32](#int32) |  | Show only a recent number of lines |
| timestamps | [bool](#bool) |  | Show timestamps |
| follow | [bool](#bool) |  | Stream data |





 

 

 


<a name="edgeproto.ExecApi"></a>

### ExecApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| RunCommand | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | Run a Command or Shell on a container |
| RunConsole | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | Run console on a VM |
| ShowLogs | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | View logs for AppInst |
| AccessCloudlet | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | Access Cloudlet VM |
| SendLocalRequest | [ExecRequest](#edgeproto.ExecRequest) | [ExecRequest](#edgeproto.ExecRequest) | This is used internally to forward requests to other Controllers.e |

 



<a name="flavor.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flavor.proto



<a name="edgeproto.Flavor"></a>

### Flavor
Flavors define the compute, memory, and storage capacity of computing instances. 
To put it simply, a flavor is an available hardware configuration for a server. 
It defines the size of a virtual server that can be launched.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [FlavorKey](#edgeproto.FlavorKey) |  | Unique key for the new flavor. |
| ram | [uint64](#uint64) |  | RAM in megabytes |
| vcpus | [uint64](#uint64) |  | Number of virtual CPUs |
| disk | [uint64](#uint64) |  | Amount of disk space in gigabytes |
| opt_res_map | [Flavor.OptResMapEntry](#edgeproto.Flavor.OptResMapEntry) | repeated | Optional Resources request, key = [gpu, nas, nic] gpu kinds: [gpu, vgpu, pci] form: $resource=$kind:[$alias]$count ex: optresmap=gpu=vgpus:nvidia-63:1 |






<a name="edgeproto.Flavor.OptResMapEntry"></a>

### Flavor.OptResMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.FlavorKey"></a>

### FlavorKey
Flavor

FlavorKey uniquely identifies a Flavor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Flavor name |





 

 

 


<a name="edgeproto.FlavorApi"></a>

### FlavorApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Create a Flavor |
| DeleteFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Delete a Flavor |
| UpdateFlavor | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Update a Flavor |
| ShowFlavor | [Flavor](#edgeproto.Flavor) | [Flavor](#edgeproto.Flavor) stream | Show Flavors |
| AddFlavorRes | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Add Optional Resource |
| RemoveFlavorRes | [Flavor](#edgeproto.Flavor) | [Result](#edgeproto.Result) | Remove Optional Resource |

 



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
| bval | [bool](#bool) |  |  |
| sval | [string](#string) |  |  |





 

 

 

 



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
| build_master | [string](#string) |  | Build Master Version |
| build_head | [string](#string) |  | Build Head Version |
| build_author | [string](#string) |  | Build Author |
| hostname | [string](#string) |  | Hostname |
| container_version | [string](#string) |  | Docker edge-cloud container version which node instance use |
| internal_pki | [string](#string) |  | Internal PKI Config |






<a name="edgeproto.NodeData"></a>

### NodeData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nodes | [Node](#edgeproto.Node) | repeated |  |






<a name="edgeproto.NodeKey"></a>

### NodeKey
NodeKey uniquely identifies a DME or CRM node


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name or hostname of node |
| type | [string](#string) |  | Node type |
| cloudlet_key | [CloudletKey](#edgeproto.CloudletKey) |  | Cloudlet on which node is running, or is associated with |
| region | [string](#string) |  | Region the node is in |





 

 

 


<a name="edgeproto.NodeApi"></a>

### NodeApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowNode | [Node](#edgeproto.Node) | [Node](#edgeproto.Node) stream | Show all Nodes connected to all Controllers |

 



<a name="notice.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## notice.proto
Notice is the message used by the notify protocol to communicate and coordinate internally between different Mobiledgex services. For details on the notify protocol, see the &#34;MEX Cloud Service Interactions&#34; confluence article.
In general, the protocol is used to synchronize state from one service to another. The protocol is fairly symmetric, with different state being synchronized both from server to client and client to server.


<a name="edgeproto.Notice"></a>

### Notice



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [NoticeAction](#edgeproto.NoticeAction) |  | Action to perform |
| version | [uint32](#uint32) |  | Protocol version supported by sender |
| any | [google.protobuf.Any](#google.protobuf.Any) |  | Data |
| want_objs | [string](#string) | repeated | Wanted Objects |
| filter_cloudlet_key | [bool](#bool) |  | Filter by cloudlet key |
| span | [string](#string) |  | Opentracing span |
| mod_rev | [int64](#int64) |  | Database revision for which object was last modified |
| tags | [Notice.TagsEntry](#edgeproto.Notice.TagsEntry) | repeated | Extra tags |






<a name="edgeproto.Notice.TagsEntry"></a>

### Notice.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


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


 

 


<a name="edgeproto.NotifyApi"></a>

### NotifyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| StreamNotice | [Notice](#edgeproto.Notice) stream | [Notice](#edgeproto.Notice) stream | Bidrectional stream for exchanging data between controller and DME/CRM |

 



<a name="operatorcode.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## operatorcode.proto



<a name="edgeproto.OperatorCode"></a>

### OperatorCode
OperatorCode maps a carrier code to an Operator organization name


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | MCC plus MNC code, or custom carrier code designation. |
| organization | [string](#string) |  | Operator Organization name |





 

 

 


<a name="edgeproto.OperatorCodeApi"></a>

### OperatorCodeApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateOperatorCode | [OperatorCode](#edgeproto.OperatorCode) | [Result](#edgeproto.Result) | Create Operator Code. Create a code for an Operator. |
| DeleteOperatorCode | [OperatorCode](#edgeproto.OperatorCode) | [Result](#edgeproto.Result) | Delete Operator Code. Delete a code for an Operator. |
| ShowOperatorCode | [OperatorCode](#edgeproto.OperatorCode) | [OperatorCode](#edgeproto.OperatorCode) stream | Show Operator Code. Show Codes for an Operator. |

 



<a name="org.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## org.proto



<a name="edgeproto.Organization"></a>

### Organization



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Organization name |






<a name="edgeproto.OrganizationData"></a>

### OrganizationData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| orgs | [Organization](#edgeproto.Organization) | repeated |  |





 

 

 


<a name="edgeproto.OrganizationApi"></a>

### OrganizationApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| OrganizationInUse | [Organization](#edgeproto.Organization) | [Result](#edgeproto.Result) | Check if an Organization is in use. |

 



<a name="privacypolicy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## privacypolicy.proto



<a name="edgeproto.OutboundSecurityRule"></a>

### OutboundSecurityRule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| protocol | [string](#string) |  | tcp, udp, icmp |
| port_range_min | [uint32](#uint32) |  | TCP or UDP port range start |
| port_range_max | [uint32](#uint32) |  | TCP or UDP port range end |
| remote_cidr | [string](#string) |  | remote CIDR X.X.X.X/X |






<a name="edgeproto.PrivacyPolicy"></a>

### PrivacyPolicy
PrivacyPolicy defines security restrictions for cluster instances
nodes scaled up or down.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [PolicyKey](#edgeproto.PolicyKey) |  | Unique identifier key |
| outbound_security_rules | [OutboundSecurityRule](#edgeproto.OutboundSecurityRule) | repeated | list of outbound security rules for whitelisting traffic |





 

 

 


<a name="edgeproto.PrivacyPolicyApi"></a>

### PrivacyPolicyApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreatePrivacyPolicy | [PrivacyPolicy](#edgeproto.PrivacyPolicy) | [Result](#edgeproto.Result) | Create a Privacy Policy |
| DeletePrivacyPolicy | [PrivacyPolicy](#edgeproto.PrivacyPolicy) | [Result](#edgeproto.Result) | Delete a Privacy policy |
| UpdatePrivacyPolicy | [PrivacyPolicy](#edgeproto.PrivacyPolicy) | [Result](#edgeproto.Result) | Update a Privacy policy |
| ShowPrivacyPolicy | [PrivacyPolicy](#edgeproto.PrivacyPolicy) | [PrivacyPolicy](#edgeproto.PrivacyPolicy) stream | Show Privacy Policies. Any fields specified will be used to filter results. |

 



<a name="refs.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## refs.proto



<a name="edgeproto.AppInstRefs"></a>

### AppInstRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [AppKey](#edgeproto.AppKey) |  | App key |
| insts | [AppInstRefs.InstsEntry](#edgeproto.AppInstRefs.InstsEntry) | repeated | AppInsts for App (key is JSON of AppInst Key) |






<a name="edgeproto.AppInstRefs.InstsEntry"></a>

### AppInstRefs.InstsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [uint32](#uint32) |  |  |






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
| root_lb_ports | [CloudletRefs.RootLbPortsEntry](#edgeproto.CloudletRefs.RootLbPortsEntry) | repeated | Used ports on root load balancer. Map key is public port, value is a bitmap for the protocol bitmap: bit 0: tcp, bit 1: udp |
| used_dynamic_ips | [int32](#int32) |  | Used dynamic IPs |
| used_static_ips | [string](#string) |  | Used static IPs |
| opt_res_used_map | [CloudletRefs.OptResUsedMapEntry](#edgeproto.CloudletRefs.OptResUsedMapEntry) | repeated | Used Optional Resources |






<a name="edgeproto.CloudletRefs.OptResUsedMapEntry"></a>

### CloudletRefs.OptResUsedMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [uint32](#uint32) |  |  |






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





 

 

 


<a name="edgeproto.AppInstRefsApi"></a>

### AppInstRefsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ShowAppInstRefs | [AppInstRefs](#edgeproto.AppInstRefs) | [AppInstRefs](#edgeproto.AppInstRefs) stream | Show AppInstRefs (debug only) |


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

 



<a name="restagtable.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## restagtable.proto



<a name="edgeproto.ResTagTable"></a>

### ResTagTable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated |  |
| key | [ResTagTableKey](#edgeproto.ResTagTableKey) |  |  |
| tags | [ResTagTable.TagsEntry](#edgeproto.ResTagTable.TagsEntry) | repeated | one or more string tags |
| azone | [string](#string) |  | availability zone(s) of resource if required |






<a name="edgeproto.ResTagTable.TagsEntry"></a>

### ResTagTable.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="edgeproto.ResTagTableKey"></a>

### ResTagTableKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Resource Table Name |
| organization | [string](#string) |  | Operator organization of the cloudlet site. |





 


<a name="edgeproto.OptResNames"></a>

### OptResNames


| Name | Number | Description |
| ---- | ------ | ----------- |
| GPU | 0 |  |
| NAS | 1 |  |
| NIC | 2 |  |


 

 


<a name="edgeproto.ResTagTableApi"></a>

### ResTagTableApi
This API should be admin-only

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateResTagTable | [ResTagTable](#edgeproto.ResTagTable) | [Result](#edgeproto.Result) | Create TagTable |
| DeleteResTagTable | [ResTagTable](#edgeproto.ResTagTable) | [Result](#edgeproto.Result) | Delete TagTable |
| UpdateResTagTable | [ResTagTable](#edgeproto.ResTagTable) | [Result](#edgeproto.Result) |  |
| ShowResTagTable | [ResTagTable](#edgeproto.ResTagTable) | [ResTagTable](#edgeproto.ResTagTable) stream | show TagTable |
| AddResTag | [ResTagTable](#edgeproto.ResTagTable) | [Result](#edgeproto.Result) | add new tag(s) to TagTable |
| RemoveResTag | [ResTagTable](#edgeproto.ResTagTable) | [Result](#edgeproto.Result) | remove existing tag(s) from TagTable |
| GetResTagTable | [ResTagTableKey](#edgeproto.ResTagTableKey) | [ResTagTable](#edgeproto.ResTagTable) | Fetch a copy of the TagTable |

 



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





 

 

 

 



<a name="settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto



<a name="edgeproto.Settings"></a>

### Settings
Global settings


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| shepherd_metrics_collection_interval | [int64](#int64) |  | Shepherd metrics collection interval for k8s and docker appInstances (duration) |
| shepherd_alert_evaluation_interval | [int64](#int64) |  | Shepherd alert evaluation interval for k8s and docker appInstances (duration) |
| shepherd_health_check_retries | [int32](#int32) |  | Number of times Shepherd Health Check fails before we mark appInst down |
| shepherd_health_check_interval | [int64](#int64) |  | Health Checking probing frequency (duration) |
| auto_deploy_interval_sec | [double](#double) |  | Auto Provisioning Stats push and analysis interval (seconds) |
| auto_deploy_offset_sec | [double](#double) |  | Auto Provisioning analysis offset from interval (seconds) |
| auto_deploy_max_intervals | [uint32](#uint32) |  | Auto Provisioning Policy max allowed intervals |
| create_app_inst_timeout | [int64](#int64) |  | Create AppInst timeout (duration) |
| update_app_inst_timeout | [int64](#int64) |  | Update AppInst timeout (duration) |
| delete_app_inst_timeout | [int64](#int64) |  | Delete AppInst timeout (duration) |
| create_cluster_inst_timeout | [int64](#int64) |  | Create ClusterInst timeout (duration) |
| update_cluster_inst_timeout | [int64](#int64) |  | Update ClusterInst timeout (duration) |
| delete_cluster_inst_timeout | [int64](#int64) |  | Delete ClusterInst timeout (duration) |
| master_node_flavor | [string](#string) |  | Default flavor for k8s master VM and &gt; 0 workers |
| load_balancer_max_port_range | [int32](#int32) |  | Max IP Port range when using a load balancer |
| max_tracked_dme_clients | [int32](#int32) |  | Max DME clients to be tracked at the same time. |
| chef_client_interval | [int32](#int32) |  | Default chef client interval (duration) |
| influx_db_metrics_retention | [int64](#int64) |  | Default influxDB metrics retention policy (duration) |
| cloudlet_maintenance_timeout | [int32](#int32) |  | Default Cloudlet Maintenance timeout (used twice for AutoProv and Cloudlet) |
| update_vm_pool_timeout | [int64](#int64) |  | Update VM pool timeout (duration) |





 

 

 


<a name="edgeproto.SettingsApi"></a>

### SettingsApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| UpdateSettings | [Settings](#edgeproto.Settings) | [Result](#edgeproto.Result) | Update settings |
| ResetSettings | [Settings](#edgeproto.Settings) | [Result](#edgeproto.Result) | Reset all settings to their defaults |
| ShowSettings | [Settings](#edgeproto.Settings) | [Settings](#edgeproto.Settings) | Show settings |

 



<a name="version.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## version.proto


 


<a name="edgeproto.VersionHash"></a>

### VersionHash
Below enum lists hashes as well as corresponding versions

| Name | Number | Description |
| ---- | ------ | ----------- |
| HASH_d41d8cd98f00b204e9800998ecf8427e | 0 |  |
| HASH_d4ca5418a77d22d968ce7a2afc549dfe | 9 | interim versions deleted |
| HASH_7848d42e3a2eaf36e53bbd3af581b13a | 10 |  |
| HASH_f31b7a9d7e06f72107e0ab13c708704e | 11 |  |
| HASH_03fad51f0343d41f617329151f474d2b | 12 |  |
| HASH_7d32a983fafc3da768e045b1dc4d5f50 | 13 |  |
| HASH_747c14bdfe2043f09d251568e4a722c6 | 14 |  |
| HASH_c7fb20f545a5bc9869b00bb770753c31 | 15 |  |
| HASH_83cd5c44b5c7387ebf7d055e7345ab42 | 16 |  |
| HASH_d8a4e697d0d693479cfd9c1c523d7e06 | 17 |  |
| HASH_e8360aa30f234ecefdfdb9fb2dc79c20 | 18 |  |
| HASH_c53c7840d242efc7209549a36fcf9e04 | 19 |  |
| HASH_1a57396698c4ade15f0579c9f5714cd6 | 20 |  |


 

 

 



<a name="vmpool.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vmpool.proto



<a name="edgeproto.VM"></a>

### VM



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | VM Name |
| net_info | [VMNetInfo](#edgeproto.VMNetInfo) |  | VM IP |
| group_name | [string](#string) |  | VM Group Name |
| state | [VMState](#edgeproto.VMState) |  | VM State |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Last updated time |
| internal_name | [string](#string) |  | VM Internal Name |
| flavor | [FlavorInfo](#edgeproto.FlavorInfo) |  | VM Flavor |






<a name="edgeproto.VMNetInfo"></a>

### VMNetInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| external_ip | [string](#string) |  | External IP |
| internal_ip | [string](#string) |  | Internal IP |






<a name="edgeproto.VMPool"></a>

### VMPool
VMPool defines a pool of VMs to be part of a Cloudlet


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [VMPoolKey](#edgeproto.VMPoolKey) |  | VMPool Key |
| vms | [VM](#edgeproto.VM) | repeated | list of VMs to be part of VM pool |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the VM pool |
| errors | [string](#string) | repeated | Any errors trying to add/remove VM to/from VM Pool |
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |






<a name="edgeproto.VMPoolInfo"></a>

### VMPoolInfo
VMPoolInfo is used to manage VM pool from Cloudlet


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [string](#string) | repeated | Fields are used for the Update API to specify which fields to apply |
| key | [VMPoolKey](#edgeproto.VMPoolKey) |  | Unique identifier key |
| notify_id | [int64](#int64) |  | Id of client assigned by server (internal use only) |
| vms | [VM](#edgeproto.VM) | repeated | list of VMs |
| state | [TrackedState](#edgeproto.TrackedState) |  | Current state of the VM pool on the Cloudlet |
| errors | [string](#string) | repeated | Any errors trying to add/remove VM to/from VM Pool |
| status | [StatusInfo](#edgeproto.StatusInfo) |  | status is used to reflect progress of creation or other events |






<a name="edgeproto.VMPoolKey"></a>

### VMPoolKey
VMPool unique key

VMPoolKey uniquely identifies a VMPool.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [string](#string) |  | Organization of the vmpool |
| name | [string](#string) |  | Name of the vmpool |






<a name="edgeproto.VMPoolMember"></a>

### VMPoolMember
VMPoolMember is used to add and remove VM from VM Pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [VMPoolKey](#edgeproto.VMPoolKey) |  | VMPool key |
| vm | [VM](#edgeproto.VM) |  | VM part of VM Pool |
| crm_override | [CRMOverride](#edgeproto.CRMOverride) |  | Override actions to CRM |






<a name="edgeproto.VMSpec"></a>

### VMSpec
VMSpec defines the specification of VM required by CRM


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| internal_name | [string](#string) |  | VM internal name |
| external_network | [bool](#bool) |  | VM has external network defined or not |
| internal_network | [bool](#bool) |  | VM has internal network defined or not |
| flavor | [Flavor](#edgeproto.Flavor) |  | VM flavor |





 


<a name="edgeproto.VMAction"></a>

### VMAction
VM Action

VMAction is the action to be performed on VM Pool

| Name | Number | Description |
| ---- | ------ | ----------- |
| VM_ACTION_DONE | 0 | Done performing action |
| VM_ACTION_ALLOCATE | 1 | Allocate VMs from VM Pool |
| VM_ACTION_RELEASE | 2 | Release VMs from VM Pool |



<a name="edgeproto.VMState"></a>

### VMState
VM State

VMState is the state of the VM

| Name | Number | Description |
| ---- | ------ | ----------- |
| VM_FREE | 0 | VM is free to use |
| VM_IN_PROGRESS | 1 | VM is in progress |
| VM_IN_USE | 2 | VM is in use |
| VM_ADD | 3 | Add VM |
| VM_REMOVE | 4 | Remove VM |
| VM_UPDATE | 5 | Update VM |
| VM_FORCE_FREE | 6 | Forcefully free a VM, to be used at user&#39;s discretion |


 

 


<a name="edgeproto.VMPoolApi"></a>

### VMPoolApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateVMPool | [VMPool](#edgeproto.VMPool) | [Result](#edgeproto.Result) | Create VMPool. Creates VM pool which will have VMs defined. |
| DeleteVMPool | [VMPool](#edgeproto.VMPool) | [Result](#edgeproto.Result) | Delete VMPool. Deletes VM pool given that none of VMs part of this pool is used. |
| UpdateVMPool | [VMPool](#edgeproto.VMPool) | [Result](#edgeproto.Result) | Update VMPool. Updates a VM pool&#39;s VMs. |
| ShowVMPool | [VMPool](#edgeproto.VMPool) | [VMPool](#edgeproto.VMPool) stream | Show VMPools. Lists all the VMs part of the VM pool. |
| AddVMPoolMember | [VMPoolMember](#edgeproto.VMPoolMember) | [Result](#edgeproto.Result) | Add VMPoolMember. Adds a VM to existing VM Pool. |
| RemoveVMPoolMember | [VMPoolMember](#edgeproto.VMPoolMember) | [Result](#edgeproto.Result) | Remove VMPoolMember. Removes a VM from existing VM Pool. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

