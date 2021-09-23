package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	//	"github.com/mobiledgex/edge-cloud/cloudcommon"
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

	require.Equal(t, 1, len(addedFlavors), "findFlavorDeltas add")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas add")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas add")
	require.Equal(t, "newFlavor1", addedFlavors[0].Name, "findFlavorDeltas add")

	// case 2 deleteFlavor
	delete(newFlavorMap, "newFlavor1")

	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 0, len(addedFlavors), "findFlavorDeltas delete")
	require.Equal(t, 1, len(deletedFlavors), "findFlavorDeltas delete")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas delete")
	require.Equal(t, "newFlavor1", deletedFlavors[0].Name, "findFlavorDeltas")

	require.Equal(t, len(flavorMap), len(newFlavorMap), "maps unequal")

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
	require.Equal(t, "flavor.tiny1", updatedFlavors[0].Name, "findFlavorDeltas update")

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
	require.Equal(t, "newFlavor1", addedFlavors[0].Name, "findFlavorDeltas multi")
	require.Equal(t, 1, len(deletedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny1", deletedFlavors[0].Name, "findFlavorDeltas multi")
	require.Equal(t, 1, len(updatedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny2", updatedFlavors[0].Name, "findFlavorDeltas multi")

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
	testinit()

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

	// Inject the cloudletInfo test data for Update to fetch
	cldInfo := testutil.CloudletInfoData[0]

	_, err := cloudletInfoApi.InjectCloudletInfo(ctx, &cldInfo)
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

	fmt.Printf("\n\n===============================> TEST START \n")

	for _, f := range newFlavors {
		c, e := cloudletInfoApi.getInfraFlavorUsageCount(ctx, &cloudletInfo, f.Name)
		require.Nil(t, e, "get usage count failed %s", e)
		if f.Name == "flavor.Tiny2" {
			require.Equal(t, 2, c, "use count off for flavor.tiny2")
		}
		if f.Name == "flavor.medium" {
			require.Equal(t, 5, c, "use count off for flavor.medium")
		}
	}
	// First an empy update, make sure nothing changed
	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\tTEST update 1 empty update\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	require.Equal(t, newFlavors, cloudletInfo.Flavors, "empty update failed")
	fmt.Printf("\n\n\tTEST update 1 empty update PASS\n")

	// remove  infra-flavor that matches what our target's meta-flavor is using (mapped to)
	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny2" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 11, len(newFlavors))

	// the op deleted tiny2 on us while running (crm thread running GatherCloudletInfo())
	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\n\tTEST update 2 delete in use tiny2 update\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	fmt.Printf("TEST--- back from Update #2 num flavors should == 13 len(cloudletInfo.Flavors)= %d \n", len(cloudletInfo.Flavors))
	// check the marking was applied
	for _, flavor := range cloudletInfo.Flavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "not deprecated")
			break
		}
	}
	// Since the infra-flavor was being used by our app, rather than being removed
	// during the update, it was left in marked deprecated.
	require.Equal(t, 12, len(cloudletInfo.Flavors), "should be present but deprecated")
	// also maintained on the cloudletInfo: DeprecatedFlavors
	for _, flavor := range cloudletInfo.DeprecatedFlavors {
		if flavor.Name == "flavor.tiny2" {
			found = true
			break
		}
	}
	require.Equal(t, true, found)

	// Verify our deprecated flavor  generated the desired Alert,
	alertFound := false
	alerts := 0
	for k, v := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			fmt.Printf("\n\tTEST found Pending alert as K: %+v V: %+v\n", k, v)
			alertFound = true
			alerts++
		}
	}
	require.Equal(t, 1, alerts, "Alert count wrong")
	require.True(t, alertFound, "failed to find infra flavor deleted alert")
	// we should be able to create an alert key, and use Get to test existence. Didn't work hm
	// Now test flavor re-mapping, with tiny2 deprecated, GetVMSpec will skip this infra flavor
	// GetVMSpec() on our meta flavor x1.tiny that would match our infra flavor.tiny2 if were not deprecated

	fmt.Printf("\n\nTEST update 2 delete in use tiny2 deprecated ====> PASS \n\n")

	fm := edgeproto.FlavorMatch{
		Key:        cloudletInfo.Key,
		FlavorName: testutil.FlavorData[0].Key.Name,
	}
	match, err := cloudletApi.FindFlavorMatch(ctx, &fm)
	require.Nil(t, err, "FindFlavorMatch")
	require.NotEqual(t, match.FlavorName, "flavor.tiny2")
	require.Equal(t, "flavor.small", match.FlavorName, "flavor remapping not as expected test data?")
	require.Equal(t, 12, len(cloudletInfo.Flavors), "should be present but deprecated")

	fmt.Printf("\n\nTEST FindFlavorMatch remapping ====> PASS\n\n")

	// Now delete an infra flavor that is not being used, this should actually remove it.
	// cloudletInfo.Flavors = cloudletInfo.Flavors
	// tiny2 is still deprecated, and not on the newFlavors list, but is in CloudletInfo.Flavors deprecated
	require.Equal(t, 11, len(newFlavors), "newFlavor count off")

	fmt.Printf("\n\nTEST delete tiny1  from newFlavors, should get deleted as not in use\n\n")

	// we tested deprecation / remapping and alert generation
	// now delete an unsed flavor, it should go
	// make sure our flavor.tiny1 is actually not being used currently

	tiny1count, err := cloudletInfoApi.getInfraFlavorUsageCount(ctx, &cloudletInfo, "flavor.tiny1")
	require.Equal(t, 0, tiny1count, "flavor.tiny1 is actually in use")

	found = false
	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny1" {
			found = true
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	// All this crap doesn't happen anymore remove
	if !found {
		fmt.Printf("\n\n Delete tiny1 test: did not find flavor.tiny1 on the current newFlavors List!!!!\n\n")
	}
	for _, f := range newFlavors {
		if f.Name == "flavor.tiny2" {
			fmt.Printf("\n\n\t -----how is tiny2 back on the newFlavors list!!!!!\n\n")
			return
		}
	}

	// delete tiny1 which is not in use from newFlavors
	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny1" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 10, len(newFlavors))
	cloudletInfo.Flavors = newFlavors

	fmt.Printf("\n\n\tTEST update 2 delete not in use tiny1: update\n\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)

	fmt.Printf("\n\n--------------TEST Back from delete tiny1 Update cloudletInfo.Flavors\n\n")

	// if we deleted tiny2 fail it should still be on the list deprecated.
	found = false
	for _, infraFlavor := range cloudletInfo.Flavors {
		if infraFlavor.Name == "flavor.tiny2" {
			found = true
		}
	}
	require.Equal(t, true, found, "Update failed dep flavor.tiny2 was mistakenly deleted!")

	found = false
	for _, infraFlavor := range cloudletInfo.Flavors {
		if infraFlavor.Name == "flavor.tiny1" {
			found = true
		}
	}
	require.Equal(t, false, found, "Update failed flavor.tiny1 not deleted but something was!")

	require.Equal(t, 11, len(cloudletInfo.Flavors), "Update failed to remove flavor.tiny1")

	fmt.Printf("\n\tTEST delete tiny1  ===> PASSED\n\n")

	// Present the now deprecated flavor.tiny2 back to newFlavors.
	// Putting back in a flavor that was (mistakenly?) deleted, or
	// simply changing its size etc. The deprecated flavor should be cleared,
	// and if the flavors values were changed, they will be applied,
	// An update event will be issued in this case.

	fmt.Printf("\n\n TEST #4 ------------ Adding back flavor %+v should clear via recreated flavor -----------------------------\n\n", testutil.CloudletInfoData[0].Flavors[1])
	newFlavors = append(newFlavors, testutil.CloudletInfoData[0].Flavors[1]) // is this tiny2?
	require.Equal(t, 11, len(newFlavors), "unexpected len of newFlavors")

	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\n\tTEST update 4 recreate tiny2  update\n\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	if len(cloudletInfo.Flavors) != 11 {
		fmt.Printf("\n\nTEST we seem to have an extra flavor after adding back tiny2 is it in there twice:\n")
		for i, f := range cloudletInfo.Flavors {
			fmt.Printf("\tcloudletInfo.Flavors[%d] = %+v\n", i, f)
		}
	}

	require.Equal(t, 11, len(cloudletInfo.Flavors), "Update failed")
	// Should not find tiny2 deprecated anymore
	for _, flavor := range cloudletInfo.Flavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, false, flavor.Deprecated, "flavor deprecation mark not cleared")
			break
		}
	}
	// and it should not be present on the deprecated flavors list
	found = false
	for _, flavor := range cloudletInfo.DeprecatedFlavors {
		if flavor.Name == "flavor.tiny2" {
			found = true
			break
		}
	}
	require.Equal(t, false, found, "flavor not removed from cloudletInfo.DeprecatedFlavors list")
	// check the InfraFlavorDeleted alert, it should have been cleared by restoring
	// the deprecated flavor.

	/*   this does not work as an existance check for our alert ? seems they all have unique ids
	alert := edgeproto.Alert{}
	alert.Labels = make(map[string]string)
	alert.Labels[cloudcommon.AlertScopeTypeTag] = cloudcommon.AlertScopeCloudlet
	alert.Labels["alertname"] = cloudcommon.AlertInfraFlavorDeleted
	alert.Labels["ImAllShookUp!"] = cloudletInfo.Key.Name
	alert.Labels["cloudletorg"] = cloudletInfo.Key.Organization
	alert.Labels["infraflavor"] = "flavor.tiny2"
	ak := alert.GetKey()
	aks := ak.GetKeyString()
	fmt.Printf("\n\nTEST aks : %s\n", aks)
	*/
	//	require.Equal(t, "", aks, "Alert exists, should not")

	// Now the other way to auto clear on of these alerts is to remove all refernces to the flavor
	// So we'll remove it again, gen the alert, and delete our test objects which should be noticed
	// by the watcher task in controller-data.go and clear the alert
	// delete the infra-flavor that matches what our targetApp's meta-flavor is using Again

	// Ends basic Update testing. Remaining tests clearing an existing alert by
	// deleting all the using objects.

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

	fmt.Printf("\n\n\tTEST update 5  remove tiny2 again need another alert\n\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	require.Equal(t, 11, len(cloudletInfo.Flavors), "should be present but deprecated")

	for _, flavor := range cloudletInfo.Flavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "not deprecated")
			break
		}
	}
	for _, flavor := range cloudletInfo.DeprecatedFlavors {
		if flavor.Name == "flavor.tiny2" {
			found = true
			break
		}
	}
	require.Equal(t, true, found, "tiny2 not found on deprecated list")
	// and we should have another pending alert for tiny2
	alertFound = false
	alerts = 0
	for k, v := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			fmt.Printf("\n\tTEST the current Pending alert looks like: key: %+v val: %+v\n", k, v)
			alertFound = true
			alerts++
		}
	}
	require.Equal(t, 1, alerts, "Alert count wrong")
	require.True(t, alertFound, "failed to find infra flavor deleted alert")

	fmt.Printf("\n\tTEST: Second Alert ===> PASS  Have new alert, delete all useages of it and test it's cleared\n")

	// Ok, with a dangling deprecated flavor, delete the users of it, our task should notice and clear the alert...
	// first delete the clusterInsts, I suppose that will take out any appInsts there

	// 15 = {AppKey:{Organization:MobiledgeX Name:SampleApp Version:1.0.0} ClusterInstKey:{ClusterKey:{Name:autocluster-mt3} CloudletKey:{Organization:AT&T Inc. Name:San Jose Site} Organization:MobiledgeX}}

	// Now we start exercising the worker task called from the using objects as they are deleted
	err = appInstApi.DeleteAppInst(&pokeInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Error Delete poke AppInst")

	err = clusterInstApi.DeleteClusterInst(&pokeClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete poke ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&pokeClust.Key, &check)
	require.False(t, foundclust)

	fmt.Printf("\n\nTEST deleting everAi  clusterInsts...\n\n")
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

	// Without a yeild, the worker task will fail to find the cloudlet to clear the pending alert
	// as it's already been deleted.
	time.Sleep(1 * time.Second)
	alertFound = false
	for k, _ := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			fmt.Printf("\n\tTEST we find this existing alert still: %+v\n", k)
			alertFound = true
		}
	}
	require.Equal(t, false, alertFound, "Alert not cleared by deleting objects")

	_, err = cloudletInfoApi.EvictCloudletInfo(ctx, &cloudletInfo)
	require.Nil(t, err, "Evict")

	dummy.Stop()
}
