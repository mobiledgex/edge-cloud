# State Transitions for ClusterInst/AppInst

When a ClusterInst/AppInst is created/updated/deleted, there is both a transaction on the Controller, and a set of work that must be done by the CRM (to Openstack/K8s/Azure/etc). We need to coordinate these two changes and make sure they are in sync.

The requirements we want to meet are:
1. CRM does not run conflicting changes in parallel, i.e. create ClusterInst A at the same time as delete ClusterInst A.
2. CRM does not run dependent changes in parallel, i.e. create ClusterInst A at the same as create AppInst AA (which is instantiated in ClusterInst A).
3. Controller can implement an algorithm to dynamically controller ClusterInst/AppInst create/delete to match customer demand.

Due to the requirements, it is important that the controller know what the state of the CRM to as much accuracy as possible. The state in question is whether each ClusterInst/AppInst object is present on the Cloudlet or not, or if it is in a transitional state of being created/modified/deleted. With knowledge of the state, the controller can ensure it does not ask the CRM to run conflicting changes in parallel, and can also make appropriate and timely decisions about where to provision ClusterInst/AppInsts.

# Two Actors and a Coordinator

There are really two actors at play here. The first is a user (or algorithm/AI impersonating the user) which tells the Controller to create/update/delete ClusterInsts/AppInsts. The second is the CRM, which tells the Controller when it's starting work, and when it's done with work.

The coordinator is the Controller which is getting told by both actors what the intended state is (from user) and if that state has been implemented (from the CRM). Because there is no synchronization via transactions or distributed lock, these messages can shift around in time and arrive at the Controller in a different order than the wall-clock time at when they were issued.

To avoid race conditions from both actors trying to set the state at the same time, we define a state machine that restricts which actor can set the state based on the current state. At each state, only one actor can transition to the next state. The state machine is defined below, divided into sections based on create, delete, and update. Note that the state of NotPresent means the object does not exist.

```
Create:
Current State      Actor            Next State
=============      =====            ========== 
NotPresent         User create      CreateRequested
CreateRequested    CRM response     Creating
                   CRM response     Ready (if Creating message was missed)
                   CRM error        CreateError
Creating           CRM response     Ready
                   CRM error        CreateError
CreateError        Controller       NotPresent (undo)

Update:
Current State      Actor            Next State
=============      =====            ========== 
Ready              User update      UpdateRequested
UpdateRequested    CRM response     Updating
                   CRM response     Ready (if Updating message was missed)
                   CRM error        UpdateError
Updating           CRM response     Ready
                   CRM error        UpdateError
UpdateError        Controller       Ready (undo, commits previous state)

Delete:
Current State      Actor            Next State
=============      =====            ========== 
Ready              User delete      DeletePrepare
DeletePrepare      Controller       DeleteRequested
DeleteRequested    CRM response     Deleting
                   CRM response     Deleted (if Deleting message was missed)
                   CRM error        DeleteError
Deleting           CRM response     NotPresent
                   CRM error        DeleteError
DeleteError        Controller       Ready (undo)
```

Additionally, the state machine needs to handle Controller crashing, CRM crashing, or network disconnect. In all cases, the CRM will reconnect to the Controller and the CRM will first send the state of all ClusterInsts/AppInsts it has.

On Controller crash or network disconnect, the CRM will still be running. Any threads still running will indicate transitional states that will be consistent with whatever Controller state was last committed, i.e. for Creating, Controller state must be CreateRequested. Therefore there is nothing special to be done for transitional states as they fall under the normal state machine transitions.

On CRM crash, there will no longer be any threads running, thus after reading the current state of Openstack/k8s, the states will either be Ready, NotPresent, or possible some Errror state which is consistent with the Controller's current state. Therefore we only need to handle Ready/NotPresent in resolving the inconsistency.

# State Transitions for Cloudlet Upgrade

An upgrade is initiated by user from controller. During which, modifications to appInst/clusterInst is not allowed.

There are three actors at play here. First is controller, via which upgrade is initiated, second is old CRM service which will be upgraded and third is new CRM service.
Upgrade is intiated from controller, old CRM then starts new CRM service. Once new CRM service is up, it kills old CRM service.

Since this needs coordination between controller and cloudlet states, there are two states at play here.
One is TrackedState, used by controller to track actions. And other one is CloudletState, used by CRM to describe it actions.
Because a TrackedState can trigger a change in CloudletState and vice-versa, state transitions must be in-sync for both the states.
Also, the only trusted state is that of controller (TrackedState), as old/new CRM can crash and come up.

Upgrade error is left as it is, so that it can be fixed manually by admin.
Here we also ensures that, in error state, modifications to appInst/clusterInst is disallowed, unless it is fixed manually.
```
+-----------------+------------------------------------+------------------+-----------------------------------------------+
|                 |                                    |                  |         Cloudlet State Transition             |
|    Controller   |               Actor                |    Controller    +-----------------------+-----------------------+
|  Current State  |                                    |    Next State    |         CRMv1         |    CRMv2              |
+-------------------------------------------------------------------------------------------------------------------------+
| Ready           | User initiated UpgradeCloudlet     | UpdateRequested  | Ready                 |                       |
+-------------------------------------------------------------------------------------------------------------------------+
| UpdateRequested | CRMv1 response                     |                  | Ready -> UpgradeInit  |                       |
|                 | Controller response                | Updating         |                       |                       |
|                 +-------------------------------------------------------------------------------------------------------+
|                 | CRMv1 error                        |                  | Ready -> Error        |                       |
|                 | Controller response                | UpdateError      |                       |                       |
+-------------------------------------------------------------------------------------------------------------------------+
| Updating        | CRMv1 brings up CRMv2 (success)    |                  | UpgradeInit           | Init                  |
|                 | CRMv2 response                     |                  |                       | Init -> UpgradeDone   |
|                 | Controller response                | UpdateDone       |                       |                       |
|                 +-------------------------------------------------------------------------------------------------------+
|                 | CRMv1 brings up CRMv2 (success)    |                  | UpgradeInit           | Init                  |
|                 | CRMv2 error                        |                  |                       | Init -> Error         |
|                 | Controller response                | UpdateError      |                       |                       |
|                 +-------------------------------------------------------------------------------------------------------+
|                 | CRMv1 brings up CRMv2 (failure)    |                  | UpgradeInit -> Error  |                       |
|                 | Controller response                | UpdateError      |                       |                       |
+-------------------------------------------------------------------------------------------------------------------------+
| UpdateDone      | CRMv2 response                     |                  |                       | UpgradeDone -> Ready  |
|                 | Controller response                | Ready            |                       |                       |
+-----------------+------------------------------------+------------------+-----------------------+-----------------------+
```
