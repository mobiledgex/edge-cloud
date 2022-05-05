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

package crmutil

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/edgexr/edge-cloud/cloud-resource-manager/redundancy"

	"github.com/edgexr/edge-cloud/cloud-resource-manager/platform"
	"github.com/edgexr/edge-cloud/cloudcommon"
	"github.com/edgexr/edge-cloud/cloudcommon/node"
	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/notify"
	"github.com/edgexr/edge-cloud/util/tasks"
	opentracing "github.com/opentracing/opentracing-go"
)

//ControllerData contains cache data for controller
type ControllerData struct {
	platform                             platform.Platform
	cloudletKey                          edgeproto.CloudletKey
	AppCache                             edgeproto.AppCache
	AppInstCache                         edgeproto.AppInstCache
	CloudletCache                        *edgeproto.CloudletCache
	CloudletInternalCache                edgeproto.CloudletInternalCache
	VMPoolCache                          edgeproto.VMPoolCache
	FlavorCache                          edgeproto.FlavorCache
	ClusterInstCache                     edgeproto.ClusterInstCache
	AppInstInfoCache                     edgeproto.AppInstInfoCache
	CloudletInfoCache                    edgeproto.CloudletInfoCache
	VMPoolInfoCache                      edgeproto.VMPoolInfoCache
	ClusterInstInfoCache                 edgeproto.ClusterInstInfoCache
	TrustPolicyCache                     edgeproto.TrustPolicyCache
	TrustPolicyExceptionCache            edgeproto.TrustPolicyExceptionCache
	CloudletPoolCache                    *edgeproto.CloudletPoolCache
	AutoProvPolicyCache                  edgeproto.AutoProvPolicyCache
	AutoScalePolicyCache                 edgeproto.AutoScalePolicyCache
	AlertCache                           edgeproto.AlertCache
	SettingsCache                        edgeproto.SettingsCache
	ResTagTableCache                     edgeproto.ResTagTableCache
	GPUDriverCache                       edgeproto.GPUDriverCache
	AlertPolicyCache                     edgeproto.AlertPolicyCache
	NetworkCache                         edgeproto.NetworkCache
	ExecReqHandler                       *ExecReqHandler
	ExecReqSend                          *notify.ExecRequestSend
	ControllerWait                       chan bool
	ControllerSyncInProgress             bool
	ControllerSyncDone                   chan bool
	WaitPlatformActive                   chan bool
	settings                             edgeproto.Settings
	NodeMgr                              *node.NodeMgr
	VMPool                               edgeproto.VMPool
	VMPoolMux                            sync.Mutex
	VMPoolUpdateMux                      sync.Mutex
	updateVMWorkers                      tasks.KeyWorkers
	updateTrustPolicyKeyworkers          tasks.KeyWorkers
	handleTrustPolicyExceptionKeyWorkers tasks.KeyWorkers
	vmActionRefMux                       sync.Mutex
	vmActionRefAction                    int
	finishInfraResourceThread            chan struct{}
	finishUpdateCloudletInfoHAThread     chan struct{}
	vmActionLastUpdate                   time.Time
	highAvailabilityManager              *redundancy.HighAvailabilityManager
	PlatformCommonInitDone               bool
	UpdateHACompatibilityVersion         bool
}

const CloudletInfoCacheKey = "cloudletInfo"
const InitCompatibilityVersionKey = "initCompatVersion"
const CloudletInfoUpdateExpireMultiple = 20 // relative to PlatformHaInstanceActiveExpireTime how long cloudletInfo cache is valid
const CloudletInfoUpdateRefreshMultiple = 9 // relative to PlatformHaInstanceActiveExpireTime how often to refresh cloudlet info

func (cd *ControllerData) RecvAllEnd(ctx context.Context) {
	if cd.ControllerSyncInProgress {
		cd.ControllerSyncDone <- true
	}
	cd.ControllerSyncInProgress = false
}

func (cd *ControllerData) RecvAllStart() {
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData(pf platform.Platform, key *edgeproto.CloudletKey, nodeMgr *node.NodeMgr, haMgr *redundancy.HighAvailabilityManager) *ControllerData {
	cd := &ControllerData{}
	cd.platform = pf
	cd.cloudletKey = *key
	edgeproto.InitAppCache(&cd.AppCache)
	edgeproto.InitAppInstCache(&cd.AppInstCache)
	edgeproto.InitCloudletInternalCache(&cd.CloudletInternalCache)
	cd.CloudletCache = nodeMgr.CloudletLookup.GetCloudletCache(node.NoRegion)
	cd.CloudletPoolCache = nodeMgr.CloudletPoolLookup.GetCloudletPoolCache(node.NoRegion)
	edgeproto.InitVMPoolCache(&cd.VMPoolCache)
	edgeproto.InitAppInstInfoCache(&cd.AppInstInfoCache)
	edgeproto.InitClusterInstInfoCache(&cd.ClusterInstInfoCache)
	edgeproto.InitCloudletInfoCache(&cd.CloudletInfoCache)
	edgeproto.InitVMPoolInfoCache(&cd.VMPoolInfoCache)
	edgeproto.InitFlavorCache(&cd.FlavorCache)
	edgeproto.InitClusterInstCache(&cd.ClusterInstCache)
	edgeproto.InitAlertCache(&cd.AlertCache)
	edgeproto.InitTrustPolicyCache(&cd.TrustPolicyCache)
	edgeproto.InitTrustPolicyExceptionCache(&cd.TrustPolicyExceptionCache)
	edgeproto.InitAutoProvPolicyCache(&cd.AutoProvPolicyCache)
	edgeproto.InitAutoScalePolicyCache(&cd.AutoScalePolicyCache)
	edgeproto.InitSettingsCache(&cd.SettingsCache)
	edgeproto.InitResTagTableCache(&cd.ResTagTableCache)
	edgeproto.InitGPUDriverCache(&cd.GPUDriverCache)
	edgeproto.InitAlertPolicyCache(&cd.AlertPolicyCache)
	edgeproto.InitNetworkCache(&cd.NetworkCache)
	cd.ExecReqHandler = NewExecReqHandler(cd)
	cd.ExecReqSend = notify.NewExecRequestSend()
	// set callbacks to trigger changes
	cd.ClusterInstCache.SetUpdatedCb(cd.clusterInstChanged)
	cd.ClusterInstCache.SetDeletedCb(cd.clusterInstDeleted)
	cd.AppInstCache.SetUpdatedCb(cd.appInstChanged)
	cd.AppInstCache.SetDeletedCb(cd.appInstDeleted)
	cd.FlavorCache.SetUpdatedCb(cd.flavorChanged)
	cd.CloudletCache.SetUpdatedCb(cd.cloudletChanged)
	cd.CloudletCache.SetDeletedCb(cd.cloudletDeleted)
	cd.VMPoolCache.SetUpdatedCb(cd.VMPoolChanged)
	cd.SettingsCache.SetUpdatedCb(cd.settingsChanged)
	cd.CloudletPoolCache.AddUpdatedCb(cd.cloudletPoolChanged)
	cd.CloudletPoolCache.AddDeletedCb(cd.cloudletPoolDeleted)

	cd.TrustPolicyExceptionCache.SetUpdatedCb(cd.trustPolicyExceptionChanged)
	cd.TrustPolicyExceptionCache.SetDeletedCb(cd.trustPolicyExceptionDeleted)

	cd.ControllerWait = make(chan bool, 1)
	cd.ControllerSyncDone = make(chan bool, 1)
	cd.WaitPlatformActive = make(chan bool, 1)

	cd.NodeMgr = nodeMgr
	cd.highAvailabilityManager = haMgr

	cd.updateVMWorkers.Init("vmpool-updatevm", cd.UpdateVMPool)
	cd.updateTrustPolicyKeyworkers.Init("update-TrustPolicy", cd.UpdateTrustPolicy)
	// create and delete both handled by same function
	cd.handleTrustPolicyExceptionKeyWorkers.Init("handle-TrustPolicyException", cd.HandleTrustPolicyException)

	cd.settings = *edgeproto.GetDefaultSettings()

	// debug functions
	nodeMgr.Debug.AddDebugFunc(GetEnvoyVersionCmd, cd.GetClusterEnvoyVersion)
	nodeMgr.Debug.AddDebugFunc("show-ha-status", haMgr.DumpHAManager)

	return cd
}

func (cd *ControllerData) GetCaches() *platform.Caches {
	return &platform.Caches{
		FlavorCache:               &cd.FlavorCache,
		TrustPolicyCache:          &cd.TrustPolicyCache,
		TrustPolicyExceptionCache: &cd.TrustPolicyExceptionCache,
		CloudletPoolCache:         cd.CloudletPoolCache,
		ClusterInstCache:          &cd.ClusterInstCache,
		ClusterInstInfoCache:      &cd.ClusterInstInfoCache,
		AppCache:                  &cd.AppCache,
		AppInstCache:              &cd.AppInstCache,
		AppInstInfoCache:          &cd.AppInstInfoCache,
		ResTagTableCache:          &cd.ResTagTableCache,
		CloudletCache:             cd.CloudletCache,
		CloudletInternalCache:     &cd.CloudletInternalCache,
		VMPoolCache:               &cd.VMPoolCache,
		VMPoolInfoCache:           &cd.VMPoolInfoCache,
		GPUDriverCache:            &cd.GPUDriverCache,
		NetworkCache:              &cd.NetworkCache,
		CloudletInfoCache:         &cd.CloudletInfoCache,
	}
}

// GatherCloudletInfo gathers all the information about the Cloudlet that
// the controller needs to be able to manage it.
func (cd *ControllerData) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "attempting to gather cloudlet info")
	err := cd.platform.GatherCloudletInfo(ctx, info)
	if err != nil {
		return fmt.Errorf("get limits failed: %s", err)
	}
	return nil
}

// GetInsts queries Openstack/Kubernetes to get all the cluster insts
// and app insts that have been created on the Cloudlet.
// It is called once at startup, and is used to repopulate the cache
// after CRM restart/crash. When the CRM connects to the controller,
// it will send the insts in the cache and the controller will resolve
// any discrepancies between the CRM's current state versus the
// controller's intended state.
//
// The controller does not know about all the steps that are used to
// create/delete a ClusterInst/AppInst, so if the CRM crashed in the
// middle of such a task, it is up to the CRM to clean up any unfinished
// state.
func (cd *ControllerData) GatherInsts() {
	// TODO: Implement me.
	// for _, cluster := range MexClusterShowClustInst() {
	//   key := get key from cluster
	//   cd.clusterInstInfoState(key, edgeproto.TrackedState_READY)
	//   for _, app := range MexAppShowAppInst(cluster) {
	//      key := get key from app
	//      cd.appInstInfoState(key, edgeproto.TrackedState_READY)
	//   }
	// }
}

// Note: these callback functions are called in the context of
// the notify receive thread. If the actions done here not quick,
// they should be done in a separate worker thread.

func (cd *ControllerData) settingsChanged(ctx context.Context, old *edgeproto.Settings, new *edgeproto.Settings) {
	cd.settings = *new
}

func (cd *ControllerData) flavorChanged(ctx context.Context, old *edgeproto.Flavor, new *edgeproto.Flavor) {
	//Do I need to do anything on a flavor change? update existing apps/clusters on this flavor?
	//flavor := edgeproto.Flavor{}
	// found := cd.FlavorCache.Get(&new.Key, &flavor)
	// if found {
	// 	// create (no updates allowed)
	// 	// CRM TODO: register flavor?
	// } else {
	// 	// CRM TODO: delete flavor?
	// }
}

// GetCloudletTrustPolicy finds the policy from the cache.  If a blank policy name is specified, an empty policy is returned
func GetCloudletTrustPolicy(ctx context.Context, name string, cloudletOrg string, privPolCache *edgeproto.TrustPolicyCache) (*edgeproto.TrustPolicy, error) {
	log.SpanLog(ctx, log.DebugLevelInfo, "GetCloudletTrustPolicy")
	if name != "" {
		pp := edgeproto.TrustPolicy{}
		pk := edgeproto.PolicyKey{
			Name:         name,
			Organization: cloudletOrg,
		}
		if !privPolCache.Get(&pk, &pp) {
			log.SpanLog(ctx, log.DebugLevelInfra, "Cannot find Trust Policy from cache", "pk", pk, "pp", pp)
			return nil, fmt.Errorf("fail to find Trust Policy from cache: %s", pk)
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "Found Trust Policy from cache", "pk", pk, "pp", pp)
			return &pp, nil
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "Returning empty trust policy for empty name")
		emptyPol := &edgeproto.TrustPolicy{}
		return emptyPol, nil
	}
}

func GetNetworksForClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, networkCache *edgeproto.NetworkCache) ([]*edgeproto.Network, error) {
	log.SpanLog(ctx, log.DebugLevelInfo, "GetNetworksForClusterInst", "clusterInst", clusterInst)
	networks := []*edgeproto.Network{}
	for _, netName := range clusterInst.Networks {
		net := edgeproto.Network{}
		nk := edgeproto.NetworkKey{
			Name:        netName,
			CloudletKey: clusterInst.Key.CloudletKey,
		}
		if !networkCache.Get(&nk, &net) {
			log.SpanLog(ctx, log.DebugLevelInfra, "Cannot find network from cache", "nk", nk)
			return nil, fmt.Errorf("fail to find network from cache: %s", nk)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Found network from cache", "nk", nk, "net", net)
		networks = append(networks, &net)
	}
	return networks, nil
}

func (cd *ControllerData) vmResourceActionBegin() {
	cd.vmActionRefMux.Lock()
	defer cd.vmActionRefMux.Unlock()
	cd.vmActionRefAction++
}

func (cd *ControllerData) CaptureResourcesSnapshot(ctx context.Context, cloudletKey *edgeproto.CloudletKey) *edgeproto.InfraResourcesSnapshot {
	// Track K8s based AppInstances only for platforms with IPAllocatedPerService
	features := cd.platform.GetFeatures()

	log.SpanLog(ctx, log.DebugLevelInfra, "update cloudlet resources snapshot", "key", cloudletKey)

	// get all the cluster instances deployed on this cloudlet
	deployedClusters := make(map[edgeproto.ClusterInstRefKey]struct{})
	cd.ClusterInstInfoCache.Show(&edgeproto.ClusterInstInfo{State: edgeproto.TrackedState_READY}, func(clusterInstInfo *edgeproto.ClusterInstInfo) error {
		refKey := edgeproto.ClusterInstRefKey{}
		refKey.FromClusterInstKey(&clusterInstInfo.Key)
		deployedClusters[refKey] = struct{}{}
		return nil
	})

	// get all the vm/k8s app instances deployed on this cloudlet
	deployedVMAppInsts := make(map[edgeproto.AppInstRefKey]struct{})
	deployedK8sAppInsts := make(map[edgeproto.AppInstRefKey]struct{})
	cd.AppInstInfoCache.Show(&edgeproto.AppInstInfo{State: edgeproto.TrackedState_READY}, func(appInstInfo *edgeproto.AppInstInfo) error {
		var app edgeproto.App
		if !cd.AppCache.Get(&appInstInfo.Key.AppKey, &app) {
			return nil
		}
		refKey := edgeproto.AppInstRefKey{}
		refKey.FromAppInstKey(&appInstInfo.Key)
		if app.Deployment == cloudcommon.DeploymentTypeVM {
			deployedVMAppInsts[refKey] = struct{}{}
		}
		if platform.TrackK8sAppInst(ctx, &app, features) {
			deployedK8sAppInsts[refKey] = struct{}{}
		}

		return nil
	})

	resources, err := cd.platform.GetCloudletInfraResources(ctx)
	if err != nil {
		errstr := fmt.Sprintf("Cloudlet resource update failed: %v", err)
		log.SpanLog(ctx, log.DebugLevelInfra, "can't fetch cloudlet resources", "error", errstr, "key", cloudletKey)
		cd.NodeMgr.Event(ctx, "Cloudlet infra resource update failure", cloudletKey.Organization, cloudletKey.GetTags(), err)
		return nil
	}
	if resources == nil {
		return nil
	}
	deployedClusterKeys := []edgeproto.ClusterInstRefKey{}
	for k, _ := range deployedClusters {
		deployedClusterKeys = append(deployedClusterKeys, k)
	}
	sort.Slice(deployedClusterKeys, func(ii, jj int) bool {
		return deployedClusterKeys[ii].GetKeyString() < deployedClusterKeys[jj].GetKeyString()
	})
	deployedVMAppKeys := []edgeproto.AppInstRefKey{}
	for k, _ := range deployedVMAppInsts {
		deployedVMAppKeys = append(deployedVMAppKeys, k)
	}
	sort.Slice(deployedVMAppKeys, func(ii, jj int) bool {
		return deployedVMAppKeys[ii].GetKeyString() < deployedVMAppKeys[jj].GetKeyString()
	})

	deployedK8sAppKeys := []edgeproto.AppInstRefKey{}
	for k, _ := range deployedK8sAppInsts {
		deployedK8sAppKeys = append(deployedK8sAppKeys, k)
	}
	sort.Slice(deployedK8sAppKeys, func(ii, jj int) bool {
		return deployedK8sAppKeys[ii].GetKeyString() < deployedK8sAppKeys[jj].GetKeyString()
	})
	resources.ClusterInsts = deployedClusterKeys
	resources.VmAppInsts = deployedVMAppKeys
	resources.K8SAppInsts = deployedK8sAppKeys
	return resources
}
func (cd *ControllerData) vmResourceActionEnd(ctx context.Context, cloudletKey *edgeproto.CloudletKey) {
	cd.vmActionRefMux.Lock()
	defer cd.vmActionRefMux.Unlock()
	cd.vmActionRefAction--
	if cd.vmActionRefAction == 0 {
		resources := cd.CaptureResourcesSnapshot(ctx, cloudletKey)

		cloudletInfo := edgeproto.CloudletInfo{}
		found := cd.CloudletInfoCache.Get(cloudletKey, &cloudletInfo)
		if !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "CloudletInfo not found for cloudlet", "key", cloudletKey)
			return
		}
		if cloudletInfo.State != dme.CloudletState_CLOUDLET_STATE_READY {
			// Cloudlet is not in READY state
			log.SpanLog(ctx, log.DebugLevelInfra, "Cloudlet is not online", "key", cloudletKey)
			return
		}
		// fetch cloudletInfo again, as data might have changed by now
		found = cd.CloudletInfoCache.Get(cloudletKey, &cloudletInfo)
		if !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "CloudletInfo not found for cloudlet", "key", cloudletKey)
			return
		}
		if resources != nil {
			cloudletInfo.ResourcesSnapshot = *resources
		}
		cd.CloudletInfoCache.Update(ctx, &cloudletInfo, 0)
		cd.vmActionLastUpdate = time.Now()
	}
}

func (cd *ControllerData) clusterInstChanged(ctx context.Context, old *edgeproto.ClusterInst, new *edgeproto.ClusterInst) {
	var err error

	if old != nil && old.State == new.State {
		return
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "ClusterInstChange", "key", new.Key, "state", new.State, "old", old)
	// store clusterInstInfo object on CRM bringup, if state is READY
	if old == nil && new.State == edgeproto.TrackedState_READY {
		cd.ClusterInstInfoCache.RefreshObj(ctx, new)
		return
	}
	if !cd.highAvailabilityManager.PlatformInstanceActive {
		log.SpanLog(ctx, log.DebugLevelInfra, "Ignoring cluster change because not active")
		return
	}

	resetStatus := edgeproto.NoResetStatus
	updateClusterCacheCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			cd.ClusterInstInfoCache.SetStatusTask(ctx, &new.Key, value, resetStatus)
		case edgeproto.UpdateStep:
			cd.ClusterInstInfoCache.SetStatusStep(ctx, &new.Key, value, resetStatus)
		}
		resetStatus = edgeproto.NoResetStatus
	}

	// do request
	if new.State == edgeproto.TrackedState_CREATE_REQUESTED {
		// Marks start of clusterinst change and hence increases ref count
		cd.vmResourceActionBegin()
		// create
		log.SpanLog(ctx, log.DebugLevelInfra, "ClusterInst create", "ClusterInst", *new)
		// reset status messages
		resetStatus = edgeproto.ResetStatus
		// create or update k8s cluster on this cloudlet
		err = cd.clusterInstInfoState(ctx, &new.Key, edgeproto.TrackedState_CREATING, updateClusterCacheCallback)
		if err != nil {
			// Marks end of clusterinst change and hence reduces ref count
			cd.vmResourceActionEnd(ctx, &new.Key.CloudletKey)
			return
		}
		go func() {
			var cloudlet edgeproto.Cloudlet
			cspan := log.StartSpan(log.DebugLevelInfra, "crm create ClusterInst", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
			log.SetTags(cspan, new.Key.GetTags())
			defer cspan.Finish()

			// Marks end of clusterinst change and hence reduces ref count
			defer cd.vmResourceActionEnd(ctx, &new.Key.CloudletKey)

			if !cd.CloudletCache.Get(&new.Key.CloudletKey, &cloudlet) {
				log.WarnLog("Could not find cloudlet in cache", "key", new.Key.CloudletKey)
				cd.clusterInstInfoError(ctx, &new.Key, edgeproto.TrackedState_CREATE_ERROR, fmt.Sprintf("Create Failed, Could not find cloudlet in cache %s", new.Key.CloudletKey), updateClusterCacheCallback)
				return
			}
			timeout := cd.settings.CreateClusterInstTimeout.TimeDuration()
			if cloudlet.TimeLimits.CreateClusterInstTimeout != 0 {
				timeout = cloudlet.TimeLimits.CreateClusterInstTimeout.TimeDuration()
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "create cluster inst", "ClusterInst", *new, "timeout", timeout)

			err = cd.platform.CreateClusterInst(ctx, new, updateClusterCacheCallback, timeout)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "error cluster create fail", "error", err)
				cd.clusterInstInfoError(ctx, &new.Key, edgeproto.TrackedState_CREATE_ERROR, fmt.Sprintf("Create failed: %s", err), updateClusterCacheCallback)
				//XXX seems clusterInstInfoError is overloaded with status for flavor and clustinst.
				return
			}
			// Get cluster resources and report to controller.
			updateClusterCacheCallback(edgeproto.UpdateTask, "Getting Cluster Infra Resources")
			resources, err := cd.platform.GetClusterInfraResources(ctx, &new.Key)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "error getting infra resources", "err", err)
			} else {
				err = cd.clusterInstInfoResources(ctx, &new.Key, resources)
				if err != nil {
					// this can happen if the cluster is deleted
					log.SpanLog(ctx, log.DebugLevelInfra, "failed to set cluster inst resources", "err", err)
				}
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "cluster state ready", "ClusterInst", *new)
			cd.clusterInstInfoState(ctx, &new.Key, edgeproto.TrackedState_READY, updateClusterCacheCallback)
		}()
	} else if new.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		// Marks start of clusterinst change and hence increases ref count
		cd.vmResourceActionBegin()
		// Marks end of clusterinst change and hence reduces ref count
		defer cd.vmResourceActionEnd(ctx, &new.Key.CloudletKey)
		log.SpanLog(ctx, log.DebugLevelInfra, "cluster inst update", "ClusterInst", *new)
		// reset status messages
		resetStatus = edgeproto.ResetStatus
		err = cd.clusterInstInfoState(ctx, &new.Key, edgeproto.TrackedState_UPDATING, updateClusterCacheCallback)
		if err != nil {
			return
		}

		log.SpanLog(ctx, log.DebugLevelInfra, "update cluster inst", "ClusterInst", *new)
		err = cd.platform.UpdateClusterInst(ctx, new, updateClusterCacheCallback)
		if err != nil {
			str := fmt.Sprintf("update failed: %s", err)
			cd.clusterInstInfoError(ctx, &new.Key, edgeproto.TrackedState_UPDATE_ERROR, str, updateClusterCacheCallback)
			return
		}
		// Update Runtime info of apps deployed on this Cluster
		log.SpanLog(ctx, log.DebugLevelInfra, "update appinst runtime info", "clusterkey", new.Key)
		var app edgeproto.App
		cd.AppInstCache.Show(&edgeproto.AppInst{}, func(obj *edgeproto.AppInst) error {
			if obj.ClusterInstKey().Matches(&new.Key) && cd.AppCache.Get(&obj.Key.AppKey, &app) {
				if obj.State != edgeproto.TrackedState_READY {
					return nil
				}
				if app.Deployment != cloudcommon.DeploymentTypeKubernetes {
					return nil
				}
				rt, err := cd.platform.GetAppInstRuntime(ctx, new, &app, obj)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "unable to get AppInstRuntime", "key", obj.Key, "err", err)
				} else {
					cd.appInstInfoRuntime(ctx, &obj.Key, edgeproto.TrackedState_READY, rt, updateClusterCacheCallback)
				}
			}
			return nil
		})
		// Get cluster resources and report to controller.
		updateClusterCacheCallback(edgeproto.UpdateTask, "Getting Cluster Infra Resources")
		resources, err := cd.platform.GetClusterInfraResources(ctx, &new.Key)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "error getting infra resources", "err", err)
		} else {
			err = cd.clusterInstInfoResources(ctx, &new.Key, resources)
			if err != nil {
				// this can happen if the cluster is deleted
				log.SpanLog(ctx, log.DebugLevelInfra, "failed to set cluster inst resources", "err", err)
			}
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "cluster state ready", "ClusterInst", *new)
		cd.clusterInstInfoState(ctx, &new.Key, edgeproto.TrackedState_READY, updateClusterCacheCallback)
	} else if new.State == edgeproto.TrackedState_DELETE_REQUESTED {
		// Marks start of clusterinst change and hence increases ref count
		cd.vmResourceActionBegin()
		log.SpanLog(ctx, log.DebugLevelInfra, "cluster inst delete", "ClusterInst", *new)
		// reset status messages
		resetStatus = edgeproto.ResetStatus
		// clusterInst was deleted
		err = cd.clusterInstInfoState(ctx, &new.Key, edgeproto.TrackedState_DELETING, updateClusterCacheCallback)
		if err != nil {
			// Marks end of clusterinst change and hence reduces ref count
			cd.vmResourceActionEnd(ctx, &new.Key.CloudletKey)
			return
		}
		go func() {
			cspan := log.StartSpan(log.DebugLevelInfra, "crm delete ClusterInst", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
			log.SetTags(cspan, new.Key.GetTags())
			defer cspan.Finish()

			// Marks end of clusterinst change and hence reduces ref count
			defer cd.vmResourceActionEnd(ctx, &new.Key.CloudletKey)

			log.SpanLog(ctx, log.DebugLevelInfra, "delete cluster inst", "ClusterInst", *new)
			err = cd.platform.DeleteClusterInst(ctx, new, updateClusterCacheCallback)
			if err != nil {
				str := fmt.Sprintf("Delete failed: %s", err)
				cd.clusterInstInfoError(ctx, &new.Key, edgeproto.TrackedState_DELETE_ERROR, str, updateClusterCacheCallback)
				return
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "set cluster inst deleted", "ClusterInst", *new)

			cd.clusterInstInfoState(ctx, &new.Key, edgeproto.TrackedState_DELETE_DONE, updateClusterCacheCallback)
		}()
	} else if new.State == edgeproto.TrackedState_CREATING {
		transStates := map[edgeproto.TrackedState]struct{}{
			edgeproto.TrackedState_CREATING: struct{}{},
		}
		cd.clusterInstInfoCheckState(ctx, &new.Key, transStates,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_CREATE_ERROR)
	} else if new.State == edgeproto.TrackedState_UPDATING {
		transStates := map[edgeproto.TrackedState]struct{}{
			edgeproto.TrackedState_UPDATING: struct{}{},
		}
		cd.clusterInstInfoCheckState(ctx, &new.Key, transStates,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_UPDATE_ERROR)
	} else if new.State == edgeproto.TrackedState_DELETING {
		transStates := map[edgeproto.TrackedState]struct{}{
			edgeproto.TrackedState_DELETING:    struct{}{},
			edgeproto.TrackedState_DELETE_DONE: struct{}{},
		}
		cd.clusterInstInfoCheckState(ctx, &new.Key, transStates,
			edgeproto.TrackedState_NOT_PRESENT,
			edgeproto.TrackedState_DELETE_ERROR)
	}
}

func (cd *ControllerData) clusterInstDeleted(ctx context.Context, old *edgeproto.ClusterInst) {
	log.SpanLog(ctx, log.DebugLevelInfra, "clusterInstDeleted", "ClusterInst", old)
	info := edgeproto.ClusterInstInfo{Key: old.Key}
	cd.ClusterInstInfoCache.Delete(ctx, &info, 0)
}

func (cd *ControllerData) handleTrustPolicyExceptionRuleForCloudletPoolKey(ctx context.Context, appInst *edgeproto.AppInst, cloudletPoolKey edgeproto.CloudletPoolKey, clusterInstKey *edgeproto.ClusterInstKey, action cloudcommon.Action) error {

	// Find TrustPolicyException for this appInst and this cloudletKey
	tpeKey := edgeproto.TrustPolicyExceptionKey{
		AppKey:          appInst.Key.AppKey,
		CloudletPoolKey: cloudletPoolKey,
	}
	var ckey edgeproto.ClusterInstKey
	if clusterInstKey != nil {
		ckey = *clusterInstKey
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionRuleForCloudletPoolKey() TrustPolicyException", "action", action.String(), "tpeKey", tpeKey, "clusterInstKey", clusterInstKey)

	err := cd.TrustPolicyExceptionCache.Show(&edgeproto.TrustPolicyException{Key: tpeKey}, func(tpe *edgeproto.TrustPolicyException) error {
		// The platform-specific implementation will probably involve network calls, which could potentially have long timeouts if stalled/hung,
		// and we don't want to starve the notify thread (which is the context here).
		// Hence we use tasks.KeyWorkers object to spawn a worker thread to do the work
		log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionRuleForCloudletPoolKey() calling dispatchTrustPolicyExceptionKeyWorkers()", "action", action.String(), "tpeKey", tpe.Key, "clusterInstKey", ckey)
		cd.dispatchTrustPolicyExceptionKeyWorkers(ctx, &tpe.Key, &ckey)
		return nil
	})
	return err
}

func (cd *ControllerData) handleTrustPolicyExceptionForCloudlet(ctx context.Context, cloudletKey *edgeproto.CloudletKey, action cloudcommon.Action) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionForCloudlet()", "cloudletKey", cloudletKey, "action", action.String())

	// A cloudlet may be part of many cloudlet pools
	cloudletPoolList, err := cd.CloudletPoolCache.GetPoolsForCloudletKey(cloudletKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionForCloudlet() no cloudletPool", "err", err)
		return err
	}

	cloudlets := make(map[edgeproto.CloudletKey]struct{})
	cloudlets[*cloudletKey] = struct{}{}

	for _, cloudletPoolKey := range cloudletPoolList {
		cd.handleTrustPolicyExceptionForCloudlets(ctx, cloudlets, &cloudletPoolKey, action)
	}

	return nil
}

func (cd *ControllerData) CloudletHasTrustPolicy(ctx context.Context, cloudletKey *edgeproto.CloudletKey) (bool, error) {
	var cloudlet edgeproto.Cloudlet
	if !cd.CloudletCache.Get(cloudletKey, &cloudlet) {
		log.SpanLog(ctx, log.DebugLevelInfra, "CloudletHasTrustPolicy() failed to fetch cloudlet from cache", "cloudletKey", cloudletKey)
		return false, fmt.Errorf("cloudlet %s not found", cloudletKey.String())
	}
	if cloudlet.TrustPolicy == "" {
		// For a TrustPolicy exception, a cloudlet needs to have a TrustPolicy.
		log.SpanLog(ctx, log.DebugLevelInfra, "CloudletHasTrustPolicy() cloudlet does not have a trust policy", "cloudletKey", cloudletKey)
		return false, nil
	}
	return true, nil
}

func (cd *ControllerData) handleTrustPolicyExceptionRules(ctx context.Context, appInst *edgeproto.AppInst, clusterInstKey *edgeproto.ClusterInstKey, action cloudcommon.Action) error {

	cloudletKey := appInst.Key.ClusterInstKey.CloudletKey

	hasTrustPolicy, tpErr := cd.CloudletHasTrustPolicy(ctx, &cloudletKey)
	if tpErr != nil {
		return tpErr
	}
	if !hasTrustPolicy {
		return nil
	}

	list, err := cd.CloudletPoolCache.GetPoolsForCloudletKey(&cloudletKey)

	if err == nil {
		for _, cloudletPoolKey := range list {
			log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionRules() checking cloudletPoolKey", "cloudletPoolKey", cloudletPoolKey, "action", action.String())
			cd.handleTrustPolicyExceptionRuleForCloudletPoolKey(ctx, appInst, cloudletPoolKey, clusterInstKey, action)
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionRules() no cloudletPoolKeys", "err", err)
	}

	return err
}

func (cd *ControllerData) appInstChanged(ctx context.Context, old *edgeproto.AppInst, new *edgeproto.AppInst) {
	log.SpanLog(ctx, log.DebugLevelInfra, "appInstChanged", "AppInst old", old, "new", new)

	var err error
	if old != nil && old.State == new.State {
		return
	}

	if new.State == edgeproto.TrackedState_READY {
		if old == nil {
			// store appInstInfo object on CRM bringup, if state is READY
			cd.AppInstInfoCache.RefreshObj(ctx, new)
		} else {
			if old.State == edgeproto.TrackedState_READY {
				return
			}
		}
		// Create a TPE if configured
		cd.handleTrustPolicyExceptionForAppInst(ctx, new, cloudcommon.Create)
		return
	}
	if !cd.highAvailabilityManager.PlatformInstanceActive {
		log.SpanLog(ctx, log.DebugLevelInfra, "Ignoring appInst change because not active")
		return
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "app inst changed", "key", new.Key)
	app := edgeproto.App{}
	found := cd.AppCache.Get(&new.Key.AppKey, &app)
	if !found {
		log.SpanLog(ctx, log.DebugLevelInfra, "App not found for AppInst", "key", new.Key)
		return
	}

	trackK8sAppInsts := platform.TrackK8sAppInst(ctx, &app, cd.platform.GetFeatures())

	trackAppResource := false
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		trackAppResource = true
	}
	if trackK8sAppInsts {
		trackAppResource = true
	}

	// do request
	log.SpanLog(ctx, log.DebugLevelInfra, "appInstChanged", "AppInst", new)
	resetStatus := edgeproto.NoResetStatus
	updateAppCacheCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			cd.AppInstInfoCache.SetStatusTask(ctx, &new.Key, value, resetStatus)
		case edgeproto.UpdateStep:
			cd.AppInstInfoCache.SetStatusStep(ctx, &new.Key, value, resetStatus)
		}
		resetStatus = edgeproto.NoResetStatus
	}

	if new.State == edgeproto.TrackedState_CREATE_REQUESTED {
		if trackAppResource {
			// Marks start of appinst change and hence increases ref count
			cd.vmResourceActionBegin()
		}
		// reset status messages
		resetStatus = edgeproto.ResetStatus
		// create
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&new.Flavor, &flavor)
		if !flavorFound {
			str := fmt.Sprintf("Flavor %s not found",
				new.Flavor.Name)
			cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_CREATE_ERROR, str, updateAppCacheCallback)
			// Marks end of appinst change and hence reduces ref count
			if trackAppResource {
				cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
			}
			return
		}
		clusterInst := edgeproto.ClusterInst{}
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInstFound := cd.ClusterInstCache.Get(new.ClusterInstKey(), &clusterInst)
			if !clusterInstFound {
				str := fmt.Sprintf("Cluster instance %s not found",
					new.ClusterInstKey().ClusterKey.Name)
				cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_CREATE_ERROR, str, updateAppCacheCallback)
				// Marks end of appinst change and hence reduces ref count
				if trackAppResource {
					cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
				}
				return
			}
		}

		err = cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_CREATING, updateAppCacheCallback)
		if err != nil {
			// Marks end of appinst change and hence reduces ref count
			if trackAppResource {
				cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
			}
			return
		}
		go func() {
			log.SpanLog(ctx, log.DebugLevelInfra, "update kube config", "AppInst", new, "ClusterInst", clusterInst)
			cspan := log.StartSpan(log.DebugLevelInfra, "crm create AppInst", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
			log.SetTags(cspan, new.Key.GetTags())
			defer cspan.Finish()

			if trackAppResource {
				// Marks end of appinst change and hence reduces ref count
				defer cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
			}
			oldUri := new.Uri
			err = cd.platform.CreateAppInst(ctx, &clusterInst, &app, new, &flavor, updateAppCacheCallback)
			if err != nil {
				errstr := fmt.Sprintf("Create App Inst failed: %s", err)
				cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_CREATE_ERROR, errstr, updateAppCacheCallback)
				log.SpanLog(ctx, log.DebugLevelInfra, "can't create app inst", "error", errstr, "key", new.Key)
				derr := cd.platform.DeleteAppInst(ctx, &clusterInst, &app, new, updateAppCacheCallback)
				if derr != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "can't cleanup app inst", "error", errstr, "key", new.Key)
				}
				return
			}
			if new.Uri != "" && oldUri != new.Uri {
				cd.AppInstInfoCache.SetUri(ctx, &new.Key, new.Uri)
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "created app inst", "appinst", new, "ClusterInst", clusterInst)

			cd.appInstInfoPowerState(ctx, &new.Key, edgeproto.PowerState_POWER_ON)
			rt, err := cd.platform.GetAppInstRuntime(ctx, &clusterInst, &app, new)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "unable to get AppInstRuntime", "key", new.Key, "err", err)
				cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_READY, updateAppCacheCallback)
			} else {
				cd.appInstInfoRuntime(ctx, &new.Key, edgeproto.TrackedState_READY, rt, updateAppCacheCallback)
			}
		}()
	} else if new.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		if trackAppResource {
			// Marks start of appinst change and hence increases ref count
			cd.vmResourceActionBegin()
			// Marks end of appinst change and hence reduces ref count
			defer cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
		}
		// reset status messages
		resetStatus = edgeproto.ResetStatus
		flavor := edgeproto.Flavor{}
		flavorFound := cd.FlavorCache.Get(&new.Flavor, &flavor)
		if !flavorFound {
			str := fmt.Sprintf("Flavor %s not found",
				new.Flavor.Name)
			cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_CREATE_ERROR, str, updateAppCacheCallback)
			return
		}
		err = cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_UPDATING, updateAppCacheCallback)
		if err != nil {
			return
		}
		// Only proceed with power action if current state and it reflecting state is valid
		nextPowerState := edgeproto.GetNextPowerState(new.PowerState, edgeproto.TransientState)
		if nextPowerState != edgeproto.PowerState_POWER_STATE_UNKNOWN {
			cd.appInstInfoPowerState(ctx, &new.Key, nextPowerState)
			log.SpanLog(ctx, log.DebugLevelInfra, "set power state on AppInst", "key", new.Key, "powerState", new.PowerState, "nextPowerState", nextPowerState)
			err = cd.platform.SetPowerState(ctx, &app, new, updateAppCacheCallback)
			if err != nil {
				errstr := fmt.Sprintf("Set AppInst PowerState failed: %s", err)
				cd.appInstInfoPowerState(ctx, &new.Key, edgeproto.PowerState_POWER_STATE_ERROR)
				cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_UPDATE_ERROR, errstr, updateAppCacheCallback)
				log.SpanLog(ctx, log.DebugLevelInfra, "can't set power state on AppInst", "error", err, "key", new.Key)
			} else {
				cd.appInstInfoPowerState(ctx, &new.Key, edgeproto.GetNextPowerState(nextPowerState, edgeproto.FinalState))
				cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_READY, updateAppCacheCallback)
			}
			return
		}
		clusterInst := edgeproto.ClusterInst{}
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInstFound := cd.ClusterInstCache.Get(new.ClusterInstKey(), &clusterInst)
			if !clusterInstFound {
				str := fmt.Sprintf("Cluster instance %s not found",
					new.ClusterInstKey().ClusterKey.Name)
				cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_UPDATE_ERROR, str, updateAppCacheCallback)
				return
			}
		}
		err = cd.platform.UpdateAppInst(ctx, &clusterInst, &app, new, &flavor, updateAppCacheCallback)
		if err != nil {
			errstr := fmt.Sprintf("Update App Inst failed: %s", err)
			cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_UPDATE_ERROR, errstr, updateAppCacheCallback)
			log.SpanLog(ctx, log.DebugLevelInfra, "can't update app inst", "error", errstr, "key", new.Key)
			return
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "updated app inst", "appisnt", new, "ClusterInst", clusterInst)
		rt, err := cd.platform.GetAppInstRuntime(ctx, &clusterInst, &app, new)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to get AppInstRuntime", "key", new.Key, "err", err)
			cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_READY, updateAppCacheCallback)
		} else {
			cd.appInstInfoRuntime(ctx, &new.Key, edgeproto.TrackedState_READY, rt, updateAppCacheCallback)
		}
	} else if new.State == edgeproto.TrackedState_DELETE_REQUESTED {
		if trackAppResource {
			// Marks start of appinst change and hence increases ref count
			cd.vmResourceActionBegin()
		}
		// reset status messages
		resetStatus = edgeproto.ResetStatus
		clusterInst := edgeproto.ClusterInst{}
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInstFound := cd.ClusterInstCache.Get(new.ClusterInstKey(), &clusterInst)
			if !clusterInstFound {
				str := fmt.Sprintf("Cluster instance %s not found",
					new.ClusterInstKey().ClusterKey.Name)
				cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_DELETE_ERROR, str, updateAppCacheCallback)
				if trackAppResource {
					// Marks end of appinst change and hence reduces ref count
					cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
				}
				return
			}
		}
		// appInst was deleted
		err = cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_DELETING, updateAppCacheCallback)
		if err != nil {
			if trackAppResource {
				// Marks end of appinst change and hence reduces ref count
				cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
			}
			return
		}
		go func() {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete app inst", "AppInst", new, "ClusterInst", clusterInst)
			cspan := log.StartSpan(log.DebugLevelInfra, "crm delete AppInst", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
			log.SetTags(cspan, new.Key.GetTags())
			defer cspan.Finish()

			if trackAppResource {
				// Marks end of appinst change and hence reduces ref count
				defer cd.vmResourceActionEnd(ctx, &new.Key.ClusterInstKey.CloudletKey)
			}

			err = cd.platform.DeleteAppInst(ctx, &clusterInst, &app, new, updateAppCacheCallback)
			if err != nil {
				errstr := fmt.Sprintf("Delete App Inst failed: %s", err)
				cd.appInstInfoError(ctx, &new.Key, edgeproto.TrackedState_DELETE_ERROR, errstr, updateAppCacheCallback)
				log.SpanLog(ctx, log.DebugLevelInfra, "can't delete app inst", "error", errstr, "key", new.Key)
				return
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "deleted app inst", "AppInst", new, "ClusterInst", clusterInst)
			cd.appInstInfoState(ctx, &new.Key, edgeproto.TrackedState_DELETE_DONE, updateAppCacheCallback)
		}()
	} else if new.State == edgeproto.TrackedState_CREATING {
		// Controller may send a CRM transitional state after a
		// disconnect or crash. Controller thinks CRM is creating
		// the appInst, and Controller is waiting for a new state from
		// the CRM. If CRM is not creating, or has not just finished
		// creating (ready), set an error state.
		transStates := map[edgeproto.TrackedState]struct{}{
			edgeproto.TrackedState_CREATING: struct{}{},
		}
		cd.appInstInfoCheckState(ctx, &new.Key, transStates,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_CREATE_ERROR)
	} else if new.State == edgeproto.TrackedState_UPDATING {
		transStates := map[edgeproto.TrackedState]struct{}{
			edgeproto.TrackedState_UPDATING: struct{}{},
		}
		cd.appInstInfoCheckState(ctx, &new.Key, transStates,
			edgeproto.TrackedState_READY,
			edgeproto.TrackedState_UPDATE_ERROR)
	} else if new.State == edgeproto.TrackedState_DELETING {
		transStates := map[edgeproto.TrackedState]struct{}{
			edgeproto.TrackedState_DELETING:    struct{}{},
			edgeproto.TrackedState_DELETE_DONE: struct{}{},
		}
		cd.appInstInfoCheckState(ctx, &new.Key, transStates,
			edgeproto.TrackedState_NOT_PRESENT,
			edgeproto.TrackedState_DELETE_ERROR)
	}
}

func (cd *ControllerData) handleTrustPolicyExceptionForAppInst(ctx context.Context, appInst *edgeproto.AppInst, action cloudcommon.Action) {

	app := edgeproto.App{}
	found := cd.AppCache.Get(&appInst.Key.AppKey, &app)
	if !found {
		log.SpanLog(ctx, log.DebugLevelInfra, "App not found for AppInst", "key", appInst.Key)
		return
	}

	if !cloudcommon.IsClusterInstReqd(&app) {
		return
	}

	clusterInst := edgeproto.ClusterInst{}
	clusterInstFound := cd.ClusterInstCache.Get(appInst.ClusterInstKey(), &clusterInst)
	if !clusterInstFound {
		log.SpanLog(ctx, log.DebugLevelInfra, "ClusterInst not found for AppInst", "key", appInst.Key)
		return
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionForAppInst() found", "clusterInst", clusterInst)

	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		// Handle TrustPolicyException rules for a given appInst and target clusterInst
		log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionForAppInst()", "key", appInst.Key, "action", action.String())
		cd.handleTrustPolicyExceptionRules(ctx, appInst, &clusterInst.Key, action)
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionForAppInst() clusterInst not dedicated", "IpAccess", clusterInst.IpAccess.String(), "appInst key", appInst.Key, "action", action.String())
	}
}

func (cd *ControllerData) appInstDeleted(ctx context.Context, old *edgeproto.AppInst) {
	log.SpanLog(ctx, log.DebugLevelInfra, "appInstDeleted", "AppInst", old)
	info := edgeproto.AppInstInfo{Key: old.Key}
	cd.AppInstInfoCache.Delete(ctx, &info, 0)
	cd.handleTrustPolicyExceptionForAppInst(ctx, old, cloudcommon.Delete)
}

func (cd *ControllerData) clusterInstInfoError(ctx context.Context, key *edgeproto.ClusterInstKey, errState edgeproto.TrackedState, err string, updateCallback edgeproto.CacheUpdateCallback) {
	if cd.ClusterInstInfoCache.Get(key, &edgeproto.ClusterInstInfo{}) {
		updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(errState)])
	} else {
		// If info obj is not yet created, send status msg after it is created
		defer updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(errState)])
	}
	cd.ClusterInstInfoCache.SetError(ctx, key, errState, err)
}

func (cd *ControllerData) clusterInstInfoState(ctx context.Context, key *edgeproto.ClusterInstKey, state edgeproto.TrackedState, updateCallback edgeproto.CacheUpdateCallback) error {
	if cd.ClusterInstInfoCache.Get(key, &edgeproto.ClusterInstInfo{}) {
		updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(state)])
	} else {
		// If info obj is not yet created, send status msg after it is created
		defer updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(state)])
	}
	if err := cd.ClusterInstInfoCache.SetState(ctx, key, state); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "ClusterInst set state failed", "err", err)
		return err
	}
	return nil
}

func (cd *ControllerData) clusterInstInfoResources(ctx context.Context, key *edgeproto.ClusterInstKey, resources *edgeproto.InfraResources) error {
	return cd.ClusterInstInfoCache.SetResources(ctx, key, resources)
}

func (cd *ControllerData) appInstInfoError(ctx context.Context, key *edgeproto.AppInstKey, errState edgeproto.TrackedState, err string, updateCallback edgeproto.CacheUpdateCallback) {
	if cd.AppInstInfoCache.Get(key, &edgeproto.AppInstInfo{}) {
		updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(errState)])
	} else {
		// If info obj is not yet created, send status msg after it is created
		defer updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(errState)])
	}
	cd.AppInstInfoCache.SetError(ctx, key, errState, err)
}

func (cd *ControllerData) appInstInfoState(ctx context.Context, key *edgeproto.AppInstKey, state edgeproto.TrackedState, updateCallback edgeproto.CacheUpdateCallback) error {
	if cd.AppInstInfoCache.Get(key, &edgeproto.AppInstInfo{}) {
		updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(state)])
	} else {
		// If info obj is not yet created, send status msg after it is created
		defer updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(state)])
	}

	if err := cd.AppInstInfoCache.SetState(ctx, key, state); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AppInst set state failed", "err", err)
		return err
	}
	return nil
}

func (cd *ControllerData) appInstInfoPowerState(ctx context.Context, key *edgeproto.AppInstKey, state edgeproto.PowerState) error {
	if err := cd.AppInstInfoCache.SetPowerState(ctx, key, state); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AppInst set power state failed", "err", err)
		return err
	}
	return nil
}

func (cd *ControllerData) appInstInfoRuntime(ctx context.Context, key *edgeproto.AppInstKey, state edgeproto.TrackedState, rt *edgeproto.AppInstRuntime, updateCallback edgeproto.CacheUpdateCallback) {
	if cd.AppInstInfoCache.Get(key, &edgeproto.AppInstInfo{}) {
		updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(state)])
	} else {
		// If info obj is not yet created, send status msg after it is created
		defer updateCallback(edgeproto.UpdateTask, edgeproto.TrackedState_CamelName[int32(state)])
	}
	cd.AppInstInfoCache.SetStateRuntime(ctx, key, state, rt)
}

// CheckState checks that the info is either in the transState or finalState.
// If not, it is an unexpected state, so we set it to the error state.
// This is used when the controller sends CRM a state that implies the
// controller is waiting for the CRM to send back the next state, but the
// CRM does not have any change in progress.
func (cd *ControllerData) clusterInstInfoCheckState(ctx context.Context, key *edgeproto.ClusterInstKey, transStates map[edgeproto.TrackedState]struct{}, finalState, errState edgeproto.TrackedState) {
	cd.ClusterInstInfoCache.UpdateModFunc(ctx, key, 0, func(old *edgeproto.ClusterInstInfo) (newObj *edgeproto.ClusterInstInfo, changed bool) {
		if old == nil {
			if _, ok := transStates[edgeproto.TrackedState_NOT_PRESENT]; ok || finalState == edgeproto.TrackedState_NOT_PRESENT {
				return old, false
			}
			old = &edgeproto.ClusterInstInfo{Key: *key}
		}
		if _, ok := transStates[old.State]; !ok && old.State != finalState {
			log.SpanLog(ctx, log.DebugLevelInfra, "inconsistent Controller vs CRM state", "old state", old.State, "transStates", transStates, "final state", finalState)
			new := &edgeproto.ClusterInstInfo{}
			*new = *old
			new.State = errState
			new.Errors = append(new.Errors, "inconsistent Controller vs CRM state")
			return new, true
		}
		return old, false
	})
}

func (cd *ControllerData) appInstInfoCheckState(ctx context.Context, key *edgeproto.AppInstKey, transStates map[edgeproto.TrackedState]struct{}, finalState, errState edgeproto.TrackedState) {
	cd.AppInstInfoCache.UpdateModFunc(ctx, key, 0, func(old *edgeproto.AppInstInfo) (newObj *edgeproto.AppInstInfo, changed bool) {
		if old == nil {
			if _, ok := transStates[edgeproto.TrackedState_NOT_PRESENT]; ok || finalState == edgeproto.TrackedState_NOT_PRESENT {
				return old, false
			}
			old = &edgeproto.AppInstInfo{Key: *key}
		}
		if _, ok := transStates[old.State]; !ok && old.State != finalState {
			log.SpanLog(ctx, log.DebugLevelInfra, "inconsistent Controller vs CRM state", "old state", old.State, "transStates", transStates, "final state", finalState)
			new := &edgeproto.AppInstInfo{}
			*new = *old
			new.State = errState
			new.Errors = append(new.Errors, "inconsistent Controller vs CRM state")
			return new, true
		}
		return old, false
	})
}

func (cd *ControllerData) notifyControllerConnect() {
	// Notify controller connect only if:
	// * started manually and not by controller
	// * if started by controller, then notify on INITOK
	select {
	case cd.ControllerWait <- true:
		// Controller - CRM communication started on Notify channel
	default:
	}
}

func (cd *ControllerData) trustPolicyExceptionChanged(ctx context.Context, old *edgeproto.TrustPolicyException, new *edgeproto.TrustPolicyException) {
	log.SpanLog(ctx, log.DebugLevelInfra, "In trustPolicyExceptionChanged()", "new trustPolicyException", new)
	if !cd.highAvailabilityManager.PlatformInstanceActive {
		log.SpanLog(ctx, log.DebugLevelInfra, "Ignoring trust policy change because not active")
		return
	}
	var action cloudcommon.Action
	if old != nil {
		if old.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE &&
			new.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_REJECTED {
			// A case of Delete TPE rules
			action = cloudcommon.Delete
		} else if new.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
			// A case of Add TPE rules
			action = cloudcommon.Create
		} else {
			// No need to handle any other transitions
			log.SpanLog(ctx, log.DebugLevelInfra, "Skipping state transition", "old", old.State.String(), "new", new.State.String())
			return
		}

		log.SpanLog(ctx, log.DebugLevelInfra, "trustPolicyExceptionChanged", "new state", new.State.String(), "action", action.String())
		tpeKey := new.Key
		clusterInsts := cd.GetClusterInstsFromTrustPolicyExceptionKey(ctx, &tpeKey)
		for _, clusterInst := range clusterInsts {
			cloudletKey := clusterInst.Key.CloudletKey
			hasTrustPolicy, tpErr := cd.CloudletHasTrustPolicy(ctx, &cloudletKey)
			if tpErr != nil {
				continue
			}
			if !hasTrustPolicy {
				log.SpanLog(ctx, log.DebugLevelInfra, "trustPolicyExceptionChanged", "no trust policy on cloudlet", cloudletKey)
				continue
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "trustPolicyExceptionChanged() calling dispatchTrustPolicyExceptionKeyWorkers()", "action", action.String(), "tpeKey", tpeKey, "clusterInstKey", clusterInst.Key)
			cd.dispatchTrustPolicyExceptionKeyWorkers(ctx, &tpeKey, &clusterInst.Key)
		}
	}
}

// Common helper function, for cloudletKey either nil (all cloudlets related to a TPE) or a specific cloudletKey
func (cd *ControllerData) getClusterInstsFromTrustPolicyExceptionKeyHelper(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey, cloudletKeyIn *edgeproto.CloudletKey, cloudletPoolKeyIn *edgeproto.CloudletPoolKey) []edgeproto.ClusterInst {

	clusterInsts := []edgeproto.ClusterInst{}
	clusterInst := edgeproto.ClusterInst{}

	appInstFilter := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: tpeKey.AppKey,
		},
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "In getClusterInstsFromTrustPolicyExceptionKeyForCloudletKey()", "tpeKey:", tpeKey)

	appInsts := []edgeproto.AppInst{}
	cd.AppInstCache.Show(&appInstFilter, func(appInst *edgeproto.AppInst) error {
		if appInst.State != edgeproto.TrackedState_READY {
			log.SpanLog(ctx, log.DebugLevelInfra, "found appInst", "appInst", appInst, "skipping because state", appInst.State)
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "found appInst in TrackedState_READY", "appInst", appInst)
		appInsts = append(appInsts, *appInst)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelInfra, "found appInsts", "num:", len(appInsts))

	for _, appInst := range appInsts {
		if !cd.ClusterInstCache.Get(appInst.ClusterInstKey(), &clusterInst) {
			log.SpanLog(ctx, log.DebugLevelInfra, "ClusterInst is missing in cache", "key", appInst.Key)
			continue
		}

		if clusterInst.IpAccess != edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			log.SpanLog(ctx, log.DebugLevelInfra, "clusterInst not dedicated", "IpAccess", clusterInst.IpAccess.String())
			continue
		}

		cloudletKey := appInst.Key.ClusterInstKey.CloudletKey
		if cloudletKeyIn != nil {
			// Check only for specified cloudletKey
			log.SpanLog(ctx, log.DebugLevelInfra, "Checking only for specified cloudletKey", "cloudletKeyIn", cloudletKeyIn)
			if !cloudletKeyIn.Matches(&cloudletKey) {
				log.SpanLog(ctx, log.DebugLevelInfra, "Not matching", "cloudletKey", cloudletKey)
				continue
			}
		}
		list := []edgeproto.CloudletPoolKey{}
		if cloudletPoolKeyIn != nil {
			list = append(list, *cloudletPoolKeyIn)
			log.SpanLog(ctx, log.DebugLevelInfra, "CloudletPool", "cloudletPoolKey", cloudletPoolKeyIn)
		} else {
			var err error
			list, err = cd.CloudletPoolCache.GetPoolsForCloudletKey(&cloudletKey)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "CloudletPool not found", "cloudletKey", cloudletKey)
				continue
			}
		}
		for _, cloudletPoolKey := range list {
			log.SpanLog(ctx, log.DebugLevelInfra, "checking cloudletPoolKey", "cloudletPoolKey", cloudletPoolKey)
			if cloudletPoolKey.Matches(&tpeKey.CloudletPoolKey) {
				log.SpanLog(ctx, log.DebugLevelInfra, "adding clusterInst", "clusterInst", clusterInst)
				clusterInsts = append(clusterInsts, clusterInst)
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "found clusterInsts", "num:", len(clusterInsts))

	return clusterInsts
}

// function : GetClusterInstsFromTrustPolicyExceptionKeyForCloudletKey()
// Input : tpeKey, cloudletKey
// Output : list of clusterInsts where appInsts are deployed in READY state and match the AppKey from tpeKey
// which are deployed on a given cloudlet
func (cd *ControllerData) GetClusterInstsFromTrustPolicyExceptionKeyForCloudletKey(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey, cloudletKeyIn *edgeproto.CloudletKey, cloudletPoolKeyIn *edgeproto.CloudletPoolKey) []edgeproto.ClusterInst {
	return cd.getClusterInstsFromTrustPolicyExceptionKeyHelper(ctx, tpeKey, cloudletKeyIn, cloudletPoolKeyIn)
}

// function : GetClusterInstsFromTrustPolicyExceptionKey()
// Input tpeKey
// Output list of clusterInsts where appInsts are deployed in READY state and match the AppKey from tpeKey
// For all cloudlets in cloudletPool in tpeKey
func (cd *ControllerData) GetClusterInstsFromTrustPolicyExceptionKey(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey) []edgeproto.ClusterInst {
	return cd.getClusterInstsFromTrustPolicyExceptionKeyHelper(ctx, tpeKey, nil, nil)
}

func (cd *ControllerData) trustPolicyExceptionDeleted(ctx context.Context, old *edgeproto.TrustPolicyException) {

	log.SpanLog(ctx, log.DebugLevelInfra, "In trustPolicyExceptionDeleted()", "TrustPolicyException:", old)
	if !cd.highAvailabilityManager.PlatformInstanceActive {
		log.SpanLog(ctx, log.DebugLevelInfra, "Ignoring trust policy deleted because not active")
		return
	}
	if old.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
		tpeKey := old.Key
		log.SpanLog(ctx, log.DebugLevelInfra, "trustPolicyExceptionDeleted", "old.key", tpeKey)
		clusterInsts := cd.GetClusterInstsFromTrustPolicyExceptionKey(ctx, &tpeKey)
		log.SpanLog(ctx, log.DebugLevelInfra, "getClusterInstsFromTrustPolicyExceptionKey() returned", "numclusterInsts", len(clusterInsts))
		for _, clusterInst := range clusterInsts {
			log.SpanLog(ctx, log.DebugLevelInfra, "calling delete dispatchTrustPolicyExceptionKeyWorkers()")
			cd.dispatchTrustPolicyExceptionKeyWorkers(ctx, &tpeKey, &clusterInst.Key)
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "done trustPolicyExceptionDeleted")
}

func (cd *ControllerData) GetTrustPolicyExceptionsForCloudletPoolKey(ctx context.Context, cloudletPoolKey edgeproto.CloudletPoolKey) []*edgeproto.TrustPolicyException {

	tpeArray := []*edgeproto.TrustPolicyException{}

	filter := edgeproto.TrustPolicyException{
		Key: edgeproto.TrustPolicyExceptionKey{
			CloudletPoolKey: cloudletPoolKey,
		},
	}

	cd.TrustPolicyExceptionCache.Show(&filter, func(tpe *edgeproto.TrustPolicyException) error {
		log.SpanLog(ctx, log.DebugLevelInfra, "In GetTrustPolicyExceptionsForCloudletPoolKey()", "adding tpe:", tpe)
		tpeArray = append(tpeArray, tpe)
		return nil
	})

	return tpeArray
}

func (cd *ControllerData) handleTrustPolicyExceptionForCloudlets(ctx context.Context, cloudlets map[edgeproto.CloudletKey]struct{}, cloudletPoolKey *edgeproto.CloudletPoolKey, action cloudcommon.Action) {
	for cloudletKey, _ := range cloudlets {
		log.SpanLog(ctx, log.DebugLevelInfra, "In handleTrustPolicyExceptionForCloudlets()", "cloudletKey", cloudletKey, "action", action.String())

		tpeArray := cd.GetTrustPolicyExceptionsForCloudletPoolKey(ctx, *cloudletPoolKey)
		if len(tpeArray) == 0 {
			log.SpanLog(ctx, log.DebugLevelInfra, "No trust policy exceptions", "cloudletKey", cloudletKey, "action", action.String())
			return
		}

		for _, tpe := range tpeArray {
			clusterInsts := cd.GetClusterInstsFromTrustPolicyExceptionKeyForCloudletKey(ctx, &tpe.Key, &cloudletKey, cloudletPoolKey)
			for _, clusterInst := range clusterInsts {
				log.SpanLog(ctx, log.DebugLevelInfra, "handleTrustPolicyExceptionForCloudlets() calling dispatchTrustPolicyExceptionKeyWorkers()", "action", action.String(), "tpeKey", tpe.Key, "clusterInstKey", clusterInst.Key)
				cd.dispatchTrustPolicyExceptionKeyWorkers(ctx, &tpe.Key, &clusterInst.Key)
			}
		}
	}
}

func (cd *ControllerData) cloudletPoolChanged(ctx context.Context, old *edgeproto.CloudletPool, new *edgeproto.CloudletPool) {
	log.SpanLog(ctx, log.DebugLevelInfra, "In cloudletPoolChanged()", "new CloudletPool:", new, "old CloudletPool:", old)

	if old == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "In cloudletPoolChanged() no old cloudletpool")
		return
	}
	cloudletsOld := make(map[edgeproto.CloudletKey]struct{})
	cloudletsNew := make(map[edgeproto.CloudletKey]struct{})

	cloudletsAdded := make(map[edgeproto.CloudletKey]struct{})
	cloudletsRemoved := make(map[edgeproto.CloudletKey]struct{})

	for _, oldCloudlet := range old.Cloudlets {
		cloudletsOld[oldCloudlet] = struct{}{}
	}
	for _, newCloudlet := range new.Cloudlets {
		cloudletsNew[newCloudlet] = struct{}{}
	}

	for _, oldCloudlet := range old.Cloudlets {
		if oldCloudlet.Name == "" {
			continue
		}
		_, ok := cloudletsNew[oldCloudlet]
		if !ok {
			cloudletsRemoved[oldCloudlet] = struct{}{}
			log.SpanLog(ctx, log.DebugLevelInfra, "cloudletsRemoved", "old", oldCloudlet)
		}
	}

	for _, newCloudlet := range new.Cloudlets {
		if newCloudlet.Name == "" {
			continue
		}
		_, ok := cloudletsOld[newCloudlet]
		if !ok {
			cloudletsAdded[newCloudlet] = struct{}{}
			log.SpanLog(ctx, log.DebugLevelInfra, "cloudletsAdded", "new", newCloudlet)
		}
	}
	if len(cloudletsRemoved) > 0 {
		cd.handleTrustPolicyExceptionForCloudlets(ctx, cloudletsRemoved, &old.Key, cloudcommon.Delete)
	}
	if len(cloudletsAdded) > 0 {
		cd.handleTrustPolicyExceptionForCloudlets(ctx, cloudletsAdded, &new.Key, cloudcommon.Create)
	}
}

func (cd *ControllerData) cloudletPoolDeleted(ctx context.Context, old *edgeproto.CloudletPool) {
	log.SpanLog(ctx, log.DebugLevelInfra, "In cloudletPoolDeleted()", "old CloudletPool:", old)
}

func (cd *ControllerData) cloudletChanged(ctx context.Context, old *edgeproto.Cloudlet, new *edgeproto.Cloudlet) {
	// do request
	log.SpanLog(ctx, log.DebugLevelInfra, "cloudletChanged", "cloudlet", new)

	cloudletInfo := edgeproto.CloudletInfo{}
	// for federated cloudlet, set cloudletinfo object if it is empty
	if old == nil && new.Key.FederatedOrganization != "" {
		found := cd.CloudletInfoCache.Get(&new.Key, &cloudletInfo)
		if !found {
			cloudletInfo.Key = new.Key
		}
		cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
		cloudletInfo.CompatibilityVersion = cloudcommon.GetCRMCompatibilityVersion()
		cd.CloudletInfoCache.Update(ctx, &cloudletInfo, 0)
	}

	found := cd.CloudletInfoCache.Get(&new.Key, &cloudletInfo)
	if !found {
		log.SpanLog(ctx, log.DebugLevelInfra, "CloudletInfo not found for cloudlet", "key", new.Key)
		return
	}
	if !cd.highAvailabilityManager.PlatformInstanceActive {
		// the cloudlet state can be anything if this is an inactive CRM
		log.SpanLog(ctx, log.DebugLevelInfra, "doing notify controller connect cloudlet changed because not currently active", "newstate", new.State, "cloudinfo state", cloudletInfo.State)
		cd.notifyControllerConnect()
		return
	} else if new.State == edgeproto.TrackedState_CRM_INITOK {
		if cloudletInfo.State == dme.CloudletState_CLOUDLET_STATE_INIT {
			cd.notifyControllerConnect()
		}
	}

	updateInfo := false
	if old != nil && old.MaintenanceState != new.MaintenanceState {
		switch new.MaintenanceState {
		case dme.MaintenanceState_CRM_REQUESTED:
			// TODO: perhaps trigger LBs to reset tcp connections
			// to gracefully force clients to move to another
			// cloudlets - but we may need to add another phase
			// in here to allow DMEs to register that Cloudlet
			// is unavailable before doing so, otherwise clients
			// will just redirected back here.

			// Acknowledge controller that CRM is in maintenance
			cloudletInfo.MaintenanceState = dme.MaintenanceState_CRM_UNDER_MAINTENANCE
			updateInfo = true
		case dme.MaintenanceState_NORMAL_OPERATION_INIT:
			// Set state back to normal so DME will allow clients
			// for this Cloudlet.
			cloudletInfo.MaintenanceState = dme.MaintenanceState_NORMAL_OPERATION
			updateInfo = true
		}
	}
	if updateInfo {
		cd.UpdateCloudletInfo(ctx, &cloudletInfo)
	}
	if old != nil && old.TrustPolicyState != new.TrustPolicyState {
		switch new.TrustPolicyState {
		case edgeproto.TrackedState_UPDATE_REQUESTED:
			log.SpanLog(ctx, log.DebugLevelInfra, "Updating Trust Policy", "new state", new.TrustPolicyState)
			if new.State != edgeproto.TrackedState_READY {
				log.SpanLog(ctx, log.DebugLevelInfra, "Update policy cannot be done until cloudlet is ready")
				cloudletInfo.TrustPolicyState = edgeproto.TrackedState_UPDATE_ERROR
				cd.UpdateCloudletInfo(ctx, &cloudletInfo)
			} else {
				cloudletInfo.TrustPolicyState = edgeproto.TrackedState_UPDATING
				cd.UpdateCloudletInfo(ctx, &cloudletInfo)
				cd.updateTrustPolicyKeyworkers.NeedsWork(ctx, new.Key)
			}
		}
	}

	updateCloudletCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			cd.CloudletInfoCache.SetStatusTask(ctx, &new.Key, value)
		case edgeproto.UpdateStep:
			cd.CloudletInfoCache.SetStatusStep(ctx, &new.Key, value)
		}
	}

	if new.State == edgeproto.TrackedState_UPDATE_REQUESTED {
		cloudlet := edgeproto.Cloudlet{}
		// Check current thread state
		if cloudletFound := cd.CloudletCache.Get(&new.Key, &cloudlet); cloudletFound {
			if cloudlet.State == edgeproto.TrackedState_UPDATING {
				return
			}
		}
		cloudletInfo := edgeproto.CloudletInfo{}
		found := cd.CloudletInfoCache.Get(&new.Key, &cloudletInfo)
		if !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "CloudletInfo not found for cloudlet", "key", new.Key)
			return
		}
		if cloudletInfo.State == dme.CloudletState_CLOUDLET_STATE_UPGRADE {
			// Cloudlet is already upgrading
			log.SpanLog(ctx, log.DebugLevelInfra, "Cloudlet update already in progress", "key", new.Key)
			return
		}
		// Reset old Status
		cloudletInfo.Status.StatusReset()
		cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_UPGRADE
		cd.UpdateCloudletInfo(ctx, &cloudletInfo)

		err := cd.platform.UpdateCloudlet(ctx, new, updateCloudletCallback)
		if err != nil {
			errstr := fmt.Sprintf("Update Cloudlet failed: %v", err)
			log.InfoLog("can't update cloudlet", "error", errstr, "key", new.Key)

			cloudletInfo.Errors = append(cloudletInfo.Errors, errstr)
			cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_ERRORS
			cd.UpdateCloudletInfo(ctx, &cloudletInfo)
			return
		}
		resources, err := cd.platform.GetCloudletInfraResources(ctx)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Cloudlet resources not found for cloudlet", "key", new.Key, "err", err)
		}
		// fetch cloudletInfo again, as status might have changed as part of UpdateCloudlet
		found = cd.CloudletInfoCache.Get(&new.Key, &cloudletInfo)
		if !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "CloudletInfo not found for cloudlet", "key", new.Key)
			return
		}
		cloudletInfo.ResourcesSnapshot.PlatformVms = resources.PlatformVms
		cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
		cloudletInfo.Status.StatusReset()
		cd.UpdateCloudletInfo(ctx, &cloudletInfo)
	}
}

func (cd *ControllerData) cloudletDeleted(ctx context.Context, old *edgeproto.Cloudlet) {
	log.SpanLog(ctx, log.DebugLevelInfra, "cloudletDeleted", "Cloudlet", old)
	if old.Key.FederatedOrganization != "" {
		// cloudlet info
		info := edgeproto.CloudletInfo{Key: old.Key}
		cd.CloudletInfoCache.Delete(ctx, &info, 0)
	}
}

// This func must be called with cd.VMPoolMux lock held
func (cd *ControllerData) UpdateVMPoolInfo(ctx context.Context, state edgeproto.TrackedState, errStr string) {
	info := edgeproto.VMPoolInfo{}
	info.Key = cd.VMPool.Key
	info.Vms = cd.VMPool.Vms
	info.State = state
	// note that if there are no errors, this should clear any existing errors state at the Controller
	if errStr != "" {
		info.Errors = []string{errStr}
	}
	cd.VMPoolInfoCache.Update(ctx, &info, 0)
}

func (cd *ControllerData) VMPoolChanged(ctx context.Context, old *edgeproto.VMPool, new *edgeproto.VMPool) {
	log.SpanLog(ctx, log.DebugLevelInfra, "VMPoolChanged", "newvmpool", new, "oldvmpool", old)
	if !cd.highAvailabilityManager.PlatformInstanceActive {
		log.SpanLog(ctx, log.DebugLevelInfra, "Ignoring VM Pool changed because not active")
		return
	}
	if old == nil || old.State == new.State {
		return
	}
	if new.State != edgeproto.TrackedState_UPDATE_REQUESTED {
		return
	}

	cd.updateVMWorkers.NeedsWork(ctx, new.Key)
}

func isVMChanged(old *edgeproto.VM, new *edgeproto.VM) bool {
	if old == nil {
		return true
	}
	if new == nil {
		return false
	}
	if new.NetInfo.ExternalIp != old.NetInfo.ExternalIp ||
		new.NetInfo.InternalIp != old.NetInfo.InternalIp {
		return true
	}
	return false
}

func (cd *ControllerData) markUpdateVMs(ctx context.Context, vmPool *edgeproto.VMPool) (bool, map[string]edgeproto.VM, []edgeproto.VM) {
	log.SpanLog(ctx, log.DebugLevelInfra, "markUpdateVMs", "vmpool", vmPool)
	cd.VMPoolMux.Lock()
	defer cd.VMPoolMux.Unlock()

	changeVMs := make(map[string]edgeproto.VM)
	for _, vm := range vmPool.Vms {
		changeVMs[vm.Name] = vm
	}

	// Incoming VM pool will have one of four cases:
	//  - All VMs in pool, with some VMs with VMState ADD, to add new VMs
	//  - All VMs in pool, with some VMs with VMState REMOVE, to remove some VMs
	//  - All VMs in pool, with all VMs with VMState UPDATE, to replace existing set of VMs with new set
	//  - All VMs in pool, with some VMs with VMState FORCE_FREE, to forcefully free some VMs
	//  - All VMs in pool with no ADD/REMOVE/UPDATE states, this happens on notify reconnect. We treat it as UPDATE above
	//  - It should never be the case that VMs will have more than one of ADD/REMOVE/UPDATE set on them in a single update

	changed := false
	newVMs := []edgeproto.VM{}
	validateVMs := []edgeproto.VM{}
	oldVMs := make(map[string]edgeproto.VM)
	updateVMs := make(map[string]edgeproto.VM)
	for _, vm := range cd.VMPool.Vms {
		cVM, ok := changeVMs[vm.Name]
		if !ok {
			// Ignored for UPDATE.
			// For ADD/REMOVE, this really shouldn't happen,
			// but preserve the VM we have locally.
			newVMs = append(newVMs, vm)
			continue
		}
		delete(changeVMs, vm.Name)
		switch cVM.State {
		case edgeproto.VMState_VM_ADD:
			cd.UpdateVMPoolInfo(
				ctx,
				edgeproto.TrackedState_UPDATE_ERROR,
				fmt.Sprintf("VM %s already exists", vm.Name),
			)
			return false, nil, nil
		case edgeproto.VMState_VM_REMOVE:
			if vm.State != edgeproto.VMState_VM_FREE {
				log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMPool, conflicting state", "vm", vm.Name, "state", vm.State)
				cd.UpdateVMPoolInfo(
					ctx,
					edgeproto.TrackedState_UPDATE_ERROR,
					fmt.Sprintf("Unable to delete VM %s, as it is in use", vm.Name),
				)
				return false, nil, nil
			}
			changed = true
			vm.State = edgeproto.VMState_VM_REMOVE
			newVMs = append(newVMs, vm)
		case edgeproto.VMState_VM_UPDATE:
			if isVMChanged(&vm, &cVM) {
				if vm.State != edgeproto.VMState_VM_FREE {
					log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMPool, conflicting state", "vm", vm.Name, "state", vm.State)
					cd.UpdateVMPoolInfo(
						ctx,
						edgeproto.TrackedState_UPDATE_ERROR,
						fmt.Sprintf("Unable to update VM %s, as it is in use", vm.Name),
					)
					return false, nil, nil
				}
				oldVMs[vm.Name] = vm
				validateVMs = append(validateVMs, cVM)
				updateVMs[vm.Name] = cVM
			} else {
				updateVMs[vm.Name] = vm
			}
			changed = true
		case edgeproto.VMState_VM_FORCE_FREE:
			log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMPool, forcefully free vm", "vm", vm.Name, "current state", vm.State)
			vm.State = edgeproto.VMState_VM_FREE
			vm.InternalName = ""
			vm.GroupName = ""
			updateVMs[vm.Name] = vm
			changed = true
		default:
			newVMs = append(newVMs, vm)
		}
	}
	for _, vm := range changeVMs {
		validateVMs = append(validateVMs, vm)
		if vm.State == edgeproto.VMState_VM_ADD {
			newVMs = append(newVMs, vm)
		} else if vm.State == edgeproto.VMState_VM_UPDATE {
			updateVMs[vm.Name] = vm
		} else if vm.State == edgeproto.VMState_VM_FORCE_FREE {
			vm.State = edgeproto.VMState_VM_FREE
			vm.InternalName = ""
			vm.GroupName = ""
			updateVMs[vm.Name] = vm
		}
		changed = true
	}

	// As part of update, vms can also be removed,
	// hence verify those vms as well
	if len(updateVMs) > 0 {
		newVMs = []edgeproto.VM{}
		for _, vm := range cd.VMPool.Vms {
			if uVM, ok := updateVMs[vm.Name]; ok {
				newVMs = append(newVMs, uVM)
				changed = true
				delete(updateVMs, vm.Name)
			} else {
				if vm.State != edgeproto.VMState_VM_FREE {
					log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMPool, conflicting state", "vm", vm.Name, "state", vm.State)
					cd.UpdateVMPoolInfo(
						ctx,
						edgeproto.TrackedState_UPDATE_ERROR,
						fmt.Sprintf("Unable to delete VM %s, as it is in use", vm.Name),
					)
					return false, nil, nil
				}
			}
		}
		for _, vm := range updateVMs {
			newVMs = append(newVMs, vm)
			changed = true
		}
	}

	if changed {
		cd.VMPool.Vms = newVMs
	} else {
		// notify controller, nothing to update
		log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMPool, nothing to update", "vmpoolkey", vmPool.Key)
		cd.UpdateVMPoolInfo(ctx, edgeproto.TrackedState_READY, "")
	}
	return changed, oldVMs, validateVMs
}

func (cd *ControllerData) UpdateVMPool(ctx context.Context, k interface{}) {
	key, ok := k.(edgeproto.VMPoolKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "Unexpected failure, key not VMPoolKey", "key", key)
		return
	}
	log.SetContextTags(ctx, key.GetTags())
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMPool", "vmpoolkey", key)

	var vmPool edgeproto.VMPool
	if !cd.VMPoolCache.Get(&key, &vmPool) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to fetch vm pool cache from controller")
		return
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "found vmpool", "vmpool", vmPool)

	cd.UpdateVMPoolInfo(ctx, edgeproto.TrackedState_UPDATING, "")

	changed, oldVMs, validateVMs := cd.markUpdateVMs(ctx, &vmPool)
	if !changed {
		return
	}

	// verify if new/updated VM is reachable
	var err error
	if len(validateVMs) > 0 {
		err = cd.platform.VerifyVMs(ctx, validateVMs)
	}

	// Update lock to update VMPool & gather new flavor list (cloudletinfo)
	cd.VMPoolUpdateMux.Lock()
	defer cd.VMPoolUpdateMux.Unlock()

	// New function block so that we can call defer on VMPoolMux Unlock
	fErr := func() error {
		cd.VMPoolMux.Lock()
		defer cd.VMPoolMux.Unlock()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to verify VMs", "vms", validateVMs, "err", err)
			// revert intermediate states
			revertVMs := []edgeproto.VM{}
			for _, vm := range cd.VMPool.Vms {
				switch vm.State {
				case edgeproto.VMState_VM_ADD:
					continue
				case edgeproto.VMState_VM_REMOVE:
					vm.State = edgeproto.VMState_VM_FREE
				case edgeproto.VMState_VM_UPDATE:
					if oVM, ok := oldVMs[vm.Name]; ok {
						vm = oVM
					}
					vm.State = edgeproto.VMState_VM_FREE
				}
				revertVMs = append(revertVMs, vm)
			}
			cd.VMPool.Vms = revertVMs
			cd.UpdateVMPoolInfo(
				ctx,
				edgeproto.TrackedState_UPDATE_ERROR,
				fmt.Sprintf("%v", err))
			return err
		}

		newVMs := []edgeproto.VM{}
		for _, vm := range cd.VMPool.Vms {
			switch vm.State {
			case edgeproto.VMState_VM_ADD:
				vm.State = edgeproto.VMState_VM_FREE
				newVMs = append(newVMs, vm)
			case edgeproto.VMState_VM_REMOVE:
				continue
			case edgeproto.VMState_VM_UPDATE:
				vm.State = edgeproto.VMState_VM_FREE
				newVMs = append(newVMs, vm)
			default:
				newVMs = append(newVMs, vm)
			}
		}
		// save VM to VM pool
		cd.VMPool.Vms = newVMs
		return nil
	}()
	if fErr != nil {
		return
	}

	// calculate Flavor info and send CloudletInfo again
	log.SpanLog(ctx, log.DebugLevelInfra, "gather vmpool flavors", "vmpool", key, "cloudlet", cd.cloudletKey)
	var cloudletInfo edgeproto.CloudletInfo
	if !cd.CloudletInfoCache.Get(&cd.cloudletKey, &cloudletInfo) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to update vmpool flavors, missing cloudletinfo", "vmpool", key, "cloudlet", cd.cloudletKey)
	} else {
		err = cd.platform.GatherCloudletInfo(ctx, &cloudletInfo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to gather vmpool flavors", "vmpool", key, "cloudlet", cd.cloudletKey, "err", err)
		} else {
			cd.UpdateCloudletInfo(ctx, &cloudletInfo)
		}
	}

	// notify controller
	cd.UpdateVMPoolInfo(ctx, edgeproto.TrackedState_READY, "")
}

func (cd *ControllerData) UpdateTrustPolicy(ctx context.Context, k interface{}) {
	cloudletKey, ok := k.(edgeproto.CloudletKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "Unexpected failure, key not CloudletKey", "cloudletKey", cloudletKey)
		return
	}
	log.SetContextTags(ctx, cloudletKey.GetTags())
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateTrustPolicy", "cloudletKey", cloudletKey)

	var cloudlet edgeproto.Cloudlet
	if !cd.CloudletCache.Get(&cloudletKey, &cloudlet) {
		log.FatalLog("failed to fetch cloudlet from cache")
	}
	var TrustPolicy edgeproto.TrustPolicy
	if cloudlet.TrustPolicy != "" {
		pkey := edgeproto.PolicyKey{
			Organization: cloudlet.Key.Organization,
			Name:         cloudlet.TrustPolicy,
		}
		if !cd.TrustPolicyCache.Get(&pkey, &TrustPolicy) {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to fetch trust policy from cache")
			return
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateTrustPolicy", "cloudlet.TrustPolicy", cloudlet.TrustPolicy, "TrustPolicyCache TrustPolicy", TrustPolicy)
	err := cd.platform.UpdateTrustPolicy(ctx, &TrustPolicy)
	log.SpanLog(ctx, log.DebugLevelInfra, "Update Privacy Done", "err", err)

	// Trust Policy is added/removed from a cloudlet.
	// If there are any trust policy exceptions programmed, they should also be added/removed.
	if cloudlet.TrustPolicy == "" {
		log.SpanLog(ctx, log.DebugLevelInfra, "UpdateTrustPolicy removed trust policy from cloudlet", "cloudletKey", cloudletKey)
		cd.handleTrustPolicyExceptionForCloudlet(ctx, &cloudletKey, cloudcommon.Delete)
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "UpdateTrustPolicy added trust policy on cloudlet", "cloudletKey", cloudletKey)
		cd.handleTrustPolicyExceptionForCloudlet(ctx, &cloudletKey, cloudcommon.Create)
	}
	cloudletInfo := edgeproto.CloudletInfo{}
	found := cd.CloudletInfoCache.Get(&cloudletKey, &cloudletInfo)
	if !found {
		log.SpanLog(ctx, log.DebugLevelInfra, "CloudletInfo not found for cloudlet", "key", cloudletKey)
		return
	}
	if err != nil {
		cloudletInfo.TrustPolicyState = edgeproto.TrackedState_UPDATE_ERROR
	} else {
		if cloudlet.TrustPolicy == "" {
			cloudletInfo.TrustPolicyState = edgeproto.TrackedState_NOT_PRESENT
		} else {
			cloudletInfo.TrustPolicyState = edgeproto.TrackedState_READY
		}
	}
	cd.UpdateCloudletInfo(ctx, &cloudletInfo)
}

// Common code to handle add and removal of trust policy exception rules
// There should only be one key worker for both create and delete of TPE.
// This guarantees ordering is preserved for the same TPE key.
func (cd *ControllerData) HandleTrustPolicyException(ctx context.Context, k interface{}) {
	tpeKeyClustInstKey, ok := k.(cloudcommon.TrustPolicyExceptionKeyClusterInstKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "HandleTrustPolicyException Unexpected failure, key not TrustPolicyExceptionKeyClusterInstKey", "key", tpeKeyClustInstKey)
		return
	}
	tpeKey := tpeKeyClustInstKey.TpeKey
	clusterInstKey := tpeKeyClustInstKey.ClusterInstKey
	log.SetContextTags(ctx, tpeKey.GetTags())
	log.SpanLog(ctx, log.DebugLevelInfra, "HandleTrustPolicyException", "TrustPolicyExceptionKey", tpeKey, "clusterInstKey", clusterInstKey)

	tpeArray := cd.GetTrustPolicyExceptionsFromKey(ctx, &tpeKey)

	if len(tpeArray) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "TrustPolicyExceptionCache.Get , key not found, calling DeleteTrustPolicyException", "key", tpeKey)
		err := cd.platform.DeleteTrustPolicyException(ctx, &tpeKey, &clusterInstKey)
		log.SpanLog(ctx, log.DebugLevelInfra, "platform DeleteTrustPolicyException Done", "err", err)
		return
	}

	for _, TrustPolicyException := range tpeArray {
		if cd.tpeNeeded(ctx, &tpeKeyClustInstKey) {
			log.SpanLog(ctx, log.DebugLevelInfra, "tpeNeeded, calling platform.UpdateTrustPolicyException()")
			err := cd.platform.UpdateTrustPolicyException(ctx, TrustPolicyException, &clusterInstKey)
			log.SpanLog(ctx, log.DebugLevelInfra, "platform UpdateTrustPolicyException Done", "err", err)
			continue
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "tpeNeeded false, calling platform.DeleteTrustPolicyException()")
		err := cd.platform.DeleteTrustPolicyException(ctx, &tpeKey, &clusterInstKey)
		log.SpanLog(ctx, log.DebugLevelInfra, "platform DeleteTrustPolicyException Done", "err", err)
	}
}

func (cd *ControllerData) GetTrustPolicyExceptionsFromKey(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey) []*edgeproto.TrustPolicyException {

	tpeArray := []*edgeproto.TrustPolicyException{}
	// Find tpe based on Key, irrespective of the tpe name.
	filter := edgeproto.TrustPolicyException{
		Key: *tpeKey,
	}

	cd.TrustPolicyExceptionCache.Show(&filter, func(tpe *edgeproto.TrustPolicyException) error {
		log.SpanLog(ctx, log.DebugLevelInfra, "In GetTrustPolicyExceptionsFromKey()", "adding tpe:", tpe)
		// The tpe passed in from Show is what's in memory in the cache.
		// To access it outside the cache lock,  need to make a copy.
		tpeCopy := edgeproto.TrustPolicyException{}
		tpeCopy.DeepCopyIn(tpe)
		tpeArray = append(tpeArray, &tpeCopy)
		return nil
	})

	return tpeArray
}

func (cd *ControllerData) RefreshAppInstRuntime(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Refresh appinst runtime info")
	var app edgeproto.App
	var clusterInst edgeproto.ClusterInst
	cd.AppInstCache.Show(&edgeproto.AppInst{}, func(obj *edgeproto.AppInst) error {
		if obj.State != edgeproto.TrackedState_READY {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to get runtime info, as AppInst is not in Ready state", "key", obj.Key, "state", obj.State)
			return nil
		}
		if !cd.AppCache.Get(&obj.Key.AppKey, &app) {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to get runtime info, as app definition is missing", "key", obj.Key)
			return nil
		}
		if app.Deployment == cloudcommon.DeploymentTypeVM {
			return nil
		}
		if !cd.ClusterInstCache.Get(obj.ClusterInstKey(), &clusterInst) {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to get runtime info, as ClusterInst is missing", "key", obj.Key)
			return nil
		}
		rt, err := cd.platform.GetAppInstRuntime(ctx, &clusterInst, &app, obj)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to get AppInstRuntime", "key", obj.Key, "err", err)
		} else {
			cd.AppInstInfoCache.SetStateRuntime(ctx, &obj.Key, obj.State, rt)
		}
		return nil
	})
}

func (cd *ControllerData) StartInfraResourceRefreshThread(cloudletInfo *edgeproto.CloudletInfo) {

	cd.finishInfraResourceThread = make(chan struct{})
	var count int

	go func() {
		done := false
		for !done {
			select {
			case <-time.After(cd.settings.ResourceSnapshotThreadInterval.TimeDuration()):
				span := log.StartSpan(log.DebugLevelApi, "CloudletResourceRefresh thread")
				ctx := log.ContextWithSpan(context.Background(), span)
				if !cd.highAvailabilityManager.PlatformInstanceActive {
					log.SpanLog(ctx, log.DebugLevelInfra, "skipping resource snapshot as platform not active")
					span.Finish()
					continue
				}
				// Cloudlet creates can take many minutes, don't try and interrogate resources before platform is ready.
				if cloudletInfo.State != dme.CloudletState_CLOUDLET_STATE_READY {
					log.SpanLog(ctx, log.DebugLevelInfra, "CloudletResourceRefreshThread", "cloudlet not yet ready", cloudletInfo.Key, "curState", cloudletInfo.State)
					span.Finish()
					continue
				}
				count++
				log.SpanLog(ctx, log.DebugLevelInfra, "CloudletResourceRefreshThread refreshing", "cloudlet", cloudletInfo.Key, "count",
					count, "ThreadIdleTime", cd.settings.ResourceSnapshotThreadInterval)
				cd.vmResourceActionBegin()
				cd.vmResourceActionEnd(ctx, &cloudletInfo.Key)
				span.Finish()
			case <-cd.finishInfraResourceThread:
				done = true
			}
		}
	}()
}

func (cd *ControllerData) StartUpdateCloudletInfoHAThread(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "StartUpdateCloudletInfoHAThread", "interval", cd.settings.PlatformHaInstanceActiveExpireTime.TimeDuration()*CloudletInfoUpdateRefreshMultiple)
	cd.finishUpdateCloudletInfoHAThread = make(chan struct{})

	go func() {
		done := false
		var cloudletInfo edgeproto.CloudletInfo
		for !done {
			select {
			case <-time.After(cd.settings.PlatformHaInstanceActiveExpireTime.TimeDuration() * CloudletInfoUpdateRefreshMultiple):
				if !cd.highAvailabilityManager.PlatformInstanceActive || !cd.PlatformCommonInitDone {
					continue
				}
				span := log.StartSpan(log.DebugLevelSampled, "StartUpdateCloudletInfoHAThread thread")
				ctx := log.ContextWithSpan(context.Background(), span)
				if !cd.CloudletInfoCache.Get(&cd.cloudletKey, &cloudletInfo) {
					log.SpanLog(ctx, log.DebugLevelInfra, "failed to find cloudlet info in cache", "cloudletKey", cd.cloudletKey)
					span.Finish()
					continue
				}
				cd.UpdateCloudletInfoAndVersionHACache(ctx, &cloudletInfo)
				span.Finish()
			case <-cd.finishUpdateCloudletInfoHAThread:
				log.SpanLog(ctx, log.DebugLevelInfra, "StartUpdateCloudletInfoHAThread done")
				done = true
			}
		}
	}()
}

func (cd *ControllerData) FinishInfraResourceRefreshThread() {
	close(cd.finishInfraResourceThread)
}

func (cd *ControllerData) FinishUpdateCloudletInfoHAThread() {
	close(cd.finishUpdateCloudletInfoHAThread)
}

// InitHAManager returns haEnabled, error
func (cd *ControllerData) InitHAManager(ctx context.Context, haMgr *redundancy.HighAvailabilityManager, haKey string) (bool, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitHAManager", "haKey", haKey)
	haEnabled := false
	haCrm := CrmHAProcess{
		controllerData: cd,
	}
	err := haMgr.Init(ctx, haKey, cd.NodeMgr, cd.settings.PlatformHaInstanceActiveExpireTime, cd.settings.PlatformHaInstancePollInterval, &haCrm)
	if err == nil {
		haEnabled = true
	} else if strings.Contains(err.Error(), redundancy.HighAvailabilityManagerDisabled) {
		log.SpanLog(ctx, log.DebugLevelInfo, "high availability disabled", "err", err)
	} else {
		return false, err
	}
	return haEnabled, nil
}

// UpdateCloudletInfo updates the cloudlet info cache while setting the ActiveCrmInstance and StandbyCrm fields based on HA state
func (cd *ControllerData) UpdateCloudletInfo(ctx context.Context, cloudletInfo *edgeproto.CloudletInfo) {
	if cd.highAvailabilityManager.HAEnabled {
		if cd.highAvailabilityManager.PlatformInstanceActive {
			cloudletInfo.ActiveCrmInstance = cd.highAvailabilityManager.HARole
			cloudletInfo.StandbyCrm = false
		} else {
			cloudletInfo.StandbyCrm = true
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateCloudletInfo", "HAEnabled", cd.highAvailabilityManager.HAEnabled, "haActive", cd.highAvailabilityManager.PlatformInstanceActive, "state", cloudletInfo.State, "activeCrm", cloudletInfo.ActiveCrmInstance, "standby", cloudletInfo.StandbyCrm)
	cd.CloudletInfoCache.Update(ctx, cloudletInfo, 0)
	if cd.highAvailabilityManager.HAEnabled && !cloudletInfo.StandbyCrm {
		err := cd.UpdateCloudletInfoAndVersionHACache(ctx, cloudletInfo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "UpdateCloudletInfoCache fail", "err", err)
		}
	}
}

func (cd *ControllerData) GetCloudletInfoFromHACache(ctx context.Context, cloudletInfo *edgeproto.CloudletInfo) error {
	ciVal, err := cd.highAvailabilityManager.GetValue(ctx, CloudletInfoCacheKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "unexpected error getting cloudletinfo from haMgr", "err", err)
		return err
	}
	if ciVal == "" {
		log.SpanLog(ctx, log.DebugLevelInfra, "no existing cloudlet info found")
	} else {
		err = json.Unmarshal([]byte(ciVal), &cloudletInfo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "cloudletInfo unmarshal error", "err", err)
			return err
		}
		cloudletInfo.ActiveCrmInstance = cd.highAvailabilityManager.HARole
		log.SpanLog(ctx, log.DebugLevelInfra, "got cloudletinfo from HA cache", "state", cloudletInfo.State)
	}
	return nil
}

// UpdateCloudletInfoAndVersionHACache updates the value for cloudletInfo and init version that HA Manager has cached in redis
func (cd *ControllerData) UpdateCloudletInfoAndVersionHACache(ctx context.Context, cloudletInfo *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateCloudletInfoAndVersionHACache", "cloudletInfo state", cloudletInfo.State.String())

	expiration := cd.settings.PlatformHaInstanceActiveExpireTime.TimeDuration() * CloudletInfoUpdateExpireMultiple
	ciJson, err := json.Marshal(cloudletInfo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cloudletinfo marshal fail", "cloudletInfo", cloudletInfo, "err", err)
		return fmt.Errorf("cloudletinfo marshal fail - %s", err)
	}
	err = cd.highAvailabilityManager.SetValue(ctx, CloudletInfoCacheKey, string(ciJson), expiration)
	if err != nil {
		return err
	}
	if cd.UpdateHACompatibilityVersion {
		err = cd.highAvailabilityManager.SetValue(ctx, InitCompatibilityVersionKey, cd.platform.GetInitHAConditionalCompatibilityVersion(ctx), expiration)
	}
	return err
}

func (cd *ControllerData) StartHAManagerActiveCheck(ctx context.Context, haMgr *redundancy.HighAvailabilityManager) {
	log.SpanLog(ctx, log.DebugLevelInfra, "StartHAManagerActiveCheck")
	go haMgr.CheckActiveLoop(ctx)
}

func (cd *ControllerData) cloudletInPool(ctx context.Context, cloudletKey *edgeproto.CloudletKey) bool {
	cloudletPoolList, _ := cd.CloudletPoolCache.GetPoolsForCloudletKey(cloudletKey)
	if len(cloudletPoolList) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "cloudletInPool() not found", "cloudletKey", cloudletKey)
		return false
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "cloudletInPool() found", "cloudletKey", cloudletKey)
	return true
}

func (cd *ControllerData) cloudletHasTrustPolicy(ctx context.Context, cloudletKey *edgeproto.CloudletKey) bool {
	hasTrustPolicy, _ := cd.CloudletHasTrustPolicy(ctx, cloudletKey)
	log.SpanLog(ctx, log.DebugLevelInfra, "cloudletHasTrustPolicy()", "cloudletKey", cloudletKey, "hasTrustPolicy", hasTrustPolicy)
	return hasTrustPolicy
}

func (cd *ControllerData) tpeIsActive(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey) bool {
	tpe := edgeproto.TrustPolicyException{}
	if cd.TrustPolicyExceptionCache.Get(tpeKey, &tpe) {
		log.SpanLog(ctx, log.DebugLevelInfra, "tpeIsActive()", "tpe", tpe)
		if tpe.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
			return true
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "tpeIsActive() not found", "tpeKey", tpeKey)

	return false
}

func (cd *ControllerData) clusterInstExists(ctx context.Context, clusterInstKey *edgeproto.ClusterInstKey) bool {
	clusterInst := edgeproto.ClusterInst{}

	clusterInstFound := cd.ClusterInstCache.Get(clusterInstKey, &clusterInst)
	log.SpanLog(ctx, log.DebugLevelInfra, "clusterInstExists()", "clusterInstKey", clusterInstKey, "clusterInstFound", clusterInstFound)

	return clusterInstFound
}

func (cd *ControllerData) appInstExistsForTpe(ctx context.Context, appKey *edgeproto.AppKey, clusterInstKey *edgeproto.ClusterInstKey) bool {

	lookupKey := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: *appKey,
		},
	}
	appInstFound := false
	cd.AppInstCache.Show(&lookupKey, func(appInst *edgeproto.AppInst) error {
		if appInst.ClusterInstKey().Matches(clusterInstKey) && appInst.State == edgeproto.TrackedState_READY {
			appInstFound = true
			log.SpanLog(ctx, log.DebugLevelInfra, "appInstExistsForTpe() matched state/cluster")
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "appInstExistsForTpe() state/cluster not matching", "appInst state", appInst.State.String(), "clusterInstKeyIn", clusterInstKey, "appInst.ClusterInstKey()", appInst.ClusterInstKey())
		}
		return nil
	})

	log.SpanLog(ctx, log.DebugLevelInfra, "appInstExistsForTpe()", "clusterInstKey", clusterInstKey, "appKey", appKey, "appInstFound", appInstFound)
	return appInstFound
}

func (cd *ControllerData) tpeNeeded(ctx context.Context, key *cloudcommon.TrustPolicyExceptionKeyClusterInstKey) bool {
	if !cd.cloudletInPool(ctx, &key.ClusterInstKey.CloudletKey) {
		// cloudlet not in cloudlet pool, so TPE does not apply
		return false
	}
	if !cd.cloudletHasTrustPolicy(ctx, &key.ClusterInstKey.CloudletKey) {
		// cloudlet does not have a trust policy, so no exceptions are needed
		return false
	}
	if !cd.tpeIsActive(ctx, &key.TpeKey) {
		// tpe does not exist, or is not active
		return false
	}
	if !cd.clusterInstExists(ctx, &key.ClusterInstKey) {
		return false
	}
	appInstFound := cd.appInstExistsForTpe(ctx, &key.TpeKey.AppKey, &key.ClusterInstKey)
	if !appInstFound {
		// appInst does not exist (/any more)
		return false
	}
	return true
}

func (cd *ControllerData) dispatchTrustPolicyExceptionKeyWorkers(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey, clusterInstKey *edgeproto.ClusterInstKey) {

	tpeKeyclusterInstKey := cloudcommon.TrustPolicyExceptionKeyClusterInstKey{
		TpeKey:         *tpeKey,
		ClusterInstKey: *clusterInstKey,
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "dispatchTrustPolicyExceptionKeyWorkers() calling handleTrustPolicyExceptionKeyWorkers", "tpeKeyclusterInstKey", tpeKeyclusterInstKey)
	cd.handleTrustPolicyExceptionKeyWorkers.NeedsWork(ctx, tpeKeyclusterInstKey)
}
