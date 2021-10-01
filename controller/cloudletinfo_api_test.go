package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudletInfo(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, testutil.CloudletInfoData)

	testutil.InternalCloudletInfoTest(t, "show", &cloudletInfoApi, testutil.CloudletInfoData)
	dummy.Stop()
}

func insertCloudletInfo(ctx context.Context, data []edgeproto.CloudletInfo) {
	for ii, _ := range data {
		in := &data[ii]
		in.State = dme.CloudletState_CLOUDLET_STATE_READY
		cloudletInfoApi.Update(ctx, in, 0)
	}
}

func TestFlavorInfoDelta(t *testing.T) {
	// Focus on the flavor delta work routine used to find added, deleted, and updated flavors.
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify | log.DebugLevelEvents)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	flavorMap := make(map[string]*edgeproto.FlavorInfo)
	newFlavorMap := make(map[string]*edgeproto.FlavorInfo)

	for _, flavor := range testutil.CloudletInfoData[0].Flavors {
		flavorMap[flavor.Name] = flavor
		newFlavorMap[flavor.Name] = flavor
	}
	require.Equal(t, 12, len(newFlavorMap), "testutil.FlavorInfo has changed")
	// If we ask right now, we should see no changes, most common case.
	addedFlavors, deletedFlavors, updatedFlavors, _ := cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)

	require.Equal(t, 0, len(addedFlavors), "findFlavorDeltas")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas")

	newFlavor1 := edgeproto.FlavorInfo{
		Name:  "newFlavor1",
		Vcpus: uint64(10),
		Ram:   uint64(4096),
		Disk:  uint64(80),
	}
	newFlavor2 := edgeproto.FlavorInfo{
		Name:  "newFlavor2",
		Vcpus: uint64(6),
		Ram:   uint64(8192),
		Disk:  uint64(60),
	}

	// case 1 addFlavor
	newFlavorMap[newFlavor1.Name] = &newFlavor1
	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 13, len(newFlavorMap), "newFlavor count off")

	require.Equal(t, 1, len(addedFlavors), "findFlavorDeltas add")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas add")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas add")
	require.Equal(t, "newFlavor1", addedFlavors["newFlavor1"].Name, "findFlavorDeltas add")

	flavorMap[newFlavor1.Name] = &newFlavor1
	// case 2 deleteFlavor
	delete(newFlavorMap, "newFlavor1")
	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 0, len(addedFlavors), "findFlavorDeltas delete")
	require.Equal(t, 1, len(deletedFlavors), "findFlavorDeltas delete")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas delete")
	require.Equal(t, "newFlavor1", deletedFlavors["newFlavor1"].Name, "findFlavorDeltas")

	delete(flavorMap, "newFlavor1")
	require.Equal(t, len(flavorMap), len(newFlavorMap), "test maps unequal len")
	// case 3 updateFlavor
	// Make  flavor.tiny1 have 2 Vcpus
	newTiny1 := edgeproto.FlavorInfo{
		Name:  "flavor.tiny1",
		Vcpus: uint64(2),
		Ram:   uint64(512),
		Disk:  uint64(10),
	}

	newFlavorMap["flavor.tiny1"] = &newTiny1

	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 0, len(addedFlavors), "findFlavorDeltas update")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas update")
	require.Equal(t, 1, len(updatedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny1", updatedFlavors["flavor.tiny1"].Name, "findFlavorDeltas update")

	delete(newFlavorMap, "newFlavor1")
	// case 4 one change of each type simultaniously
	// Delete flavor.tiny1, add newFlavor1, and update tiny2 to use 1024 ram and Vcpus to 2
	// start with them equal

	for _, flavor := range testutil.CloudletInfoData[0].Flavors {
		flavorMap[flavor.Name] = flavor
		newFlavorMap[flavor.Name] = flavor
	}
	require.Equal(t, 12, len(newFlavorMap), "testutil.FlavorInfo has changed")
	require.Equal(t, 12, len(flavorMap), "testutil.FlavorInfo has changed")

	// Delete tiny1
	delete(newFlavorMap, "flavor.tiny1")
	// add newFlavor1
	newFlavorMap[newFlavor1.Name] = &newFlavor1

	newTiny2 := edgeproto.FlavorInfo{
		Name:  "flavor.tiny2",
		Vcpus: uint64(2),
		Ram:   uint64(1024),
		Disk:  uint64(10),
	}
	// update tiny2
	newFlavorMap[newTiny2.Name] = &newTiny2

	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 1, len(addedFlavors), "findFlavorDeltas update")
	require.Equal(t, "newFlavor1", addedFlavors["newFlavor1"].Name, "findFlavorDeltas multi")
	require.Equal(t, 1, len(deletedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny1", deletedFlavors["flavor.tiny1"].Name, "findFlavorDeltas multi")
	require.Equal(t, 1, len(updatedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny2", updatedFlavors["flavor.tiny2"].Name, "findFlavorDeltas multi")

	// case 5 two adds at a time
	newFlavorMap = nil
	flavorMap = nil
	flavorMap = make(map[string]*edgeproto.FlavorInfo)
	newFlavorMap = make(map[string]*edgeproto.FlavorInfo)
	for _, flavor := range testutil.CloudletInfoData[0].Flavors {
		flavorMap[flavor.Name] = flavor
		newFlavorMap[flavor.Name] = flavor
	}
	require.Equal(t, 12, len(newFlavorMap), "testutil.FlavorInfo has changed")
	require.Equal(t, 12, len(flavorMap), "testutil.FlavorInfo has changed")

	newFlavorMap[newFlavor1.Name] = &newFlavor1
	newFlavorMap[newFlavor2.Name] = &newFlavor2

	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 2, len(addedFlavors), "add 2 flavors")
}

func TestFlavorInfoUpdate(t *testing.T) {
	// Focus on the CloudletInfo.Update rtn deleting an infra flavor that is in use.
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	*localEtcd = true
	*initLocalEtcd = true

	testinit()
	defer testfinish()
	services = Services{}

	dummy := dummyEtcd{}
	dummy.Start()

	fmt.Printf("\n\tTEST startServices\n")
	err := startServices()
	defer stopServices()
	require.Nil(t, err, "start crm")
	fmt.Printf("\tTEST startServices done, proceeding with test\n")

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData())

	// Inject the cloudletInfo test data for Update to fetch
	cldInfo := testutil.CloudletInfoData[0]

	_, err = cloudletInfoApi.InjectCloudletInfo(ctx, &cldInfo)
	require.Nil(t, err)

	responder := NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)
	responder.SetSimulateClusterCreateFailure(false)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)

	pokeClust := testutil.ClusterInstData[0]
	pokeClust.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.CreateClusterInst(&pokeClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create poke ClusterInst")
	check := edgeproto.ClusterInst{}
	foundclust := clusterInstApi.cache.Get(&pokeClust.Key, &check)
	require.True(t, foundclust)
	everAiClust := testutil.ClusterInstData[3]
	everAiClust.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.CreateClusterInst(&everAiClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create everAi ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&everAiClust.Key, &check)
	require.True(t, foundclust)

	reserveClust := testutil.ClusterInstData[7]
	reserveClust.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.CreateClusterInst(&reserveClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create reservable  ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&reserveClust.Key, &check)
	require.True(t, foundclust)

	var found bool

	// we have 3 clusterInsts, so create some apps on them
	// PokeCluster I guess has pokemon
	pokeInst := testutil.AppInstData[0]
	pokeInst.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS // let STMDEL happen to trigger our deleteCbs
	err = appInstApi.CreateAppInst(&pokeInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "CreateAppInst failed")

	cldRefs := testutil.CloudletRefsData[0]
	sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudletRefsApi.store.STMPut(stm, &cldRefs)
		return nil
	})

	cloudletInfo := testutil.CloudletInfoData[0]
	// We always present newFlavors for an Update
	// and always inspect cloudletInfo.Flavors after the update
	newFlavors := make([]*edgeproto.FlavorInfo, len(testutil.CloudletInfoData[0].Flavors))
	copy(newFlavors, testutil.CloudletInfoData[0].Flavors)
	require.Equal(t, 12, len(newFlavors), "test_data  flavors count unexpected")

	countRefs, e := cloudletInfoApi.getInfraFlavorUsageCounts(ctx, &cloudletInfo)
	require.Nil(t, e, "get usage count failed %s", e)
	count := countRefs["flavor.tiny2"]
	require.Equal(t, 8, count, "use count off for flavor.tiny2")
	count = countRefs["flavor.medium"]
	require.Equal(t, 5, count, "use count off for flavor.medium")

	// First an empy update, make sure nothing changed
	cloudletInfo.Flavors = newFlavors
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	require.Equal(t, newFlavors, cloudletInfo.Flavors, "empty update failed")
	// remove  infra-flavor that matches what our target's meta-flavor is using (mapped to)
	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny2" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 11, len(newFlavors))
	fmt.Printf("\n\tTEST empty update PASS\n")

	// the op deleted tiny2 on us while running (crm thread running GatherCloudletInfo())
	cloudletInfo.Flavors = newFlavors
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	// check the marking was applied

	for i, flavor := range cloudletInfo.Flavors {
		fmt.Printf("\n\tTEST cloudletInfo.Flavors[%d] = %+v\n", i, flavor)
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "not deprecated")
			break
		}
	}

	// Since the infra-flavor was being used by our app, rather than being removed
	// during the update, it was left in marked deprecated.
	require.Equal(t, 12, len(cloudletInfo.Flavors), "should be present but deprecated")

	// Verify our deprecated flavor  generated the desired Alert,
	// we should be able to create an alert with the expected labels and get it key and look it up
	// to test existence. Didn't work.
	alertFound := false
	alerts := 0
	for k, _ := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			if strings.Contains(k.GetKeyString(), "flavor.tiny2") {
				alertFound = true
				alerts++
				break
			}
		}
	}
	require.Equal(t, 1, alerts, "Alert count wrong")
	require.True(t, alertFound, "failed to find infra flavor deleted alert")

	fmt.Printf("\n\tTEST deprecated flavor PASS\n")

	fmt.Printf("\n\tTEST flavor remapping  start \n")
	// Now test flavor re-mapping, with tiny2 deprecated, GetVMSpec will skip this infra flavor
	// GetVMSpec() on our meta flavor x1.tiny that would match our infra flavor.tiny2 if were not deprecated
	fm := edgeproto.FlavorMatch{
		Key:        cloudletInfo.Key,
		FlavorName: testutil.FlavorData[0].Key.Name,
	}
	match, err := cloudletApi.FindFlavorMatch(ctx, &fm)
	require.Nil(t, err, "FindFlavorMatch")
	require.NotEqual(t, match.FlavorName, "flavor.tiny2")
	require.Equal(t, "flavor.small", match.FlavorName, "flavor remapping not as expected test data?")
	require.Equal(t, 12, len(cloudletInfo.Flavors), "should be present but deprecated")
	// Now delete an infra flavor that is not being used, this should actually remove it.
	// cloudletInfo.Flavors = cloudletInfo.Flavors
	// tiny2 is still deprecated, and not on the newFlavors list, but is in CloudletInfo.Flavors deprecated
	require.Equal(t, 11, len(newFlavors), "newFlavor count off")
	// we tested deprecation / remapping and alert generation

	fmt.Printf("\n\tTEST flavor remapping  PASS\n")

	// now delete an unsed flavor, it should go quietly
	// make sure our flavor.tiny1 is still not being used.
	countRefs, err = cloudletInfoApi.getInfraFlavorUsageCounts(ctx, &cloudletInfo)
	require.Nil(t, err, "getInforaFlavorUsageCounts failed")
	tiny1count := countRefs["flavor.tiny1"]
	require.Equal(t, 0, tiny1count, "flavor.tiny1 is unexpectedly in use!")

	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny1" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 10, len(newFlavors))

	fmt.Printf("\n\tTEST start delete tiny1 test\n")
	cloudletInfo.Flavors = newFlavors
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		found = cloudletInfoApi.store.STMGet(stm, &cloudletInfo.Key, &cloudletInfo)
		require.Equal(t, true, found, "CloudletInfo not found")
		return nil
	})

	require.Equal(t, 11, len(cloudletInfo.Flavors), "tiny2 should exist dep, tiny1 should be deleted")
	// if we deleted tiny2 it should still be on the list deprecated.
	found = false
	for _, infraFlavor := range cloudletInfo.Flavors {
		if infraFlavor.Name == "flavor.tiny2" {
			require.Equal(t, true, infraFlavor.Deprecated, "tiny2 no longer deprecated")
			found = true
			break
		}
	}
	require.Equal(t, true, found, "Update failed dep flavor.tiny2 was mistakenly deleted!")
	found = false
	for _, infraFlavor := range cloudletInfo.Flavors {
		if infraFlavor.Name == "flavor.tiny1" {
			found = true
			break
		}
	}
	require.Equal(t, false, found, "Update failed flavor.tiny1 not deleted but something was")
	require.Equal(t, 11, len(cloudletInfo.Flavors), "Update failed to remove flavor.tiny1")

	fmt.Printf("\n\tTEST delete tiny1 test   PASS\n")

	// Present the now deprecated flavor.tiny2 back to newFlavors.
	// Putting back in a flavor that was previsouly deprecated should clear
	// the dep mark restoring the flavor. (It will again be considered for matching meta flavors)
	// and if the flavors values were changed, they will be applied,
	// An update event will be issued in that  case

	fmt.Printf("\n\tTEST recreate flavor.tiny2 \n")
	require.Equal(t, "flavor.tiny2", testutil.CloudletInfoData[0].Flavors[1].Name, "test_data changed")
	newFlavors = append(newFlavors, testutil.CloudletInfoData[0].Flavors[1]) // is this tiny2
	require.Equal(t, 11, len(newFlavors), "unexpected len of newFlavors")

	cloudletInfo.Flavors = newFlavors
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	require.Equal(t, 11, len(cloudletInfo.Flavors), "Update failed")
	// Should not find tiny2 deprecated anymore
	for _, flavor := range cloudletInfo.Flavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, false, flavor.Deprecated, "flavor deprecation mark not cleared")
			break
		}
	}

	fmt.Printf("\n\tTEST recreate flavor.tiny2    PASS\n")
	// check the InfraFlavorDeleted alert, it should have been cleared by restoring
	// the deprecated flavor.

	// Ends basic Update testing
	// Now the other way to auto clear one of these alerts is to remove all refernces to the flavor
	// So we'll remove it again, gen the alert, and delete our test objects, which have calls to
	// our worker task to check for pending alerts containing their flavor(s). Once all objects using
	// tiny2 have been deleted, the task should clear the alert and pull out the flavor.

	fmt.Printf("\n\tTEST delete tiny2 again to generated another alert\n")
	// pull tiny2 back out to gen the alert
	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny2" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 10, len(newFlavors), "newFlav len off")
	cloudletInfo.Flavors = newFlavors

	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	require.Equal(t, 11, len(cloudletInfo.Flavors), "should be present but deprecated")

	for _, flavor := range cloudletInfo.Flavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "not deprecated")
			break
		}
	}
	// and we should have another pending alert for tiny2
	alertFound = false
	alerts = 0
	for k, _ := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			alertFound = true
			alerts++
			break
		}
	}
	require.Equal(t, 1, alerts, "Alert count wrong")
	require.True(t, alertFound, "failed to find infra flavor deleted alert")
	// Ok, with an existing deprecated flavor, and pending Alert  delete the users of it, our task should notice and clear the alert
	// first delete the clusterInsts, I suppose that will take out any appInsts there

	fmt.Printf("\n\tTEST delete tiny2 again to generated another alert PASS\n")

	fmt.Printf("\n\tTEST clear pending alert via delete clusters/apps\n")

	// Now we start exercising the worker task called from the using objects as they are deleted
	err = appInstApi.DeleteAppInst(&pokeInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Error Delete poke AppInst")

	err = clusterInstApi.DeleteClusterInst(&pokeClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete poke ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&pokeClust.Key, &check)
	require.False(t, foundclust)
	everAiClust = testutil.ClusterInstData[3]
	err = clusterInstApi.DeleteClusterInst(&everAiClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete everAi ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&everAiClust.Key, &check)
	require.False(t, foundclust)
	reserveClust = testutil.ClusterInstData[7]
	err = clusterInstApi.DeleteClusterInst(&reserveClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete reservable  ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&reserveClust.Key, &check)
	require.False(t, foundclust)
	// wait for the task to complete
	cloudletInfoApi.infraFlavorAlertTask.WaitIdle()

	alertFound = false
	for k, _ := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			alertFound = true
			break
		}
	}
	require.Equal(t, false, alertFound, "Alert not cleared by deleting objects")
	sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		found = cloudletInfoApi.store.STMGet(stm, &cloudletInfo.Key, &cloudletInfo)
		require.Equal(t, true, found, "CloudletInfo not found")
		return nil
	})
	// We started wtih 12 infra flavors, deleted tiny1, and now tiny2,
	require.Equal(t, 10, len(cloudletInfo.Flavors), "cloudletInfoFlavors count")
	// finsh clean up

	fmt.Printf("\n\tTEST clear pending alert via delete clusters/apps  PASS\n")

	// should we add tiny1 and tiny2 back into cloueltlet?

	_, err = cloudletInfoApi.EvictCloudletInfo(ctx, &cloudletInfo)
	// More to cleanup? XXX
	require.Nil(t, err, "Evict")
	dummy.Stop()
}
