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
	require.Equal(t, 13, len(newFlavorMap), "testutil.FlavorInfo has changed")
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
	require.Equal(t, 13, len(newFlavorMap), "testutil.FlavorInfo has changed")
	require.Equal(t, 13, len(flavorMap), "testutil.FlavorInfo has changed")

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
	require.Equal(t, 13, len(newFlavorMap), "testutil.FlavorInfo has changed")
	require.Equal(t, 13, len(flavorMap), "testutil.FlavorInfo has changed")

	newFlavorMap[newFlavor1.Name] = &newFlavor1
	newFlavorMap[newFlavor2.Name] = &newFlavor2

	addedFlavors, deletedFlavors, updatedFlavors, _ = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 2, len(addedFlavors), "add 2 flavors")
}

func dumpFlavors(name string, flavors []*edgeproto.FlavorInfo) {
	for i, f := range flavors {
		fmt.Printf("\t\t%s[%d] = %+v\n", name, i, f)
	}
	fmt.Printf("\n")
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
	curFlavors := cloudletInfo.Flavors
	testFlavors := curFlavors
	require.Equal(t, 13, len(curFlavors), "test flavors modified")

	// We always present newFlavors to an Update
	// and read results into curFlavors for sanity
	newFlavors := curFlavors

	/* debug remove */
	fmt.Printf("\n\n===============================> TEST START \n")

	dumpFlavors("existing flavors", cloudletInfo.Flavors)

	// First an empy update, make sure nothing changed
	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\tTEST update 1 empty update\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	curFlavors = cloudletInfo.Flavors
	require.Equal(t, newFlavors, curFlavors, "empty update failed")
	fmt.Printf("\n\n\tTEST update 1 empty update PASS\n")
	// delete the infra-flavor that matches what our target's meta-flavor is using (mapped to)
	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny2" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 12, len(newFlavors), "error removing flavor")
	// the op deleted tiny2 on us while running (crm thread running GatherCloudletInfo())
	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\n\tTEST update 2 delete in use tiny2 update\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)

	time.Sleep(1 * time.Second)
	err = cloudletInfoApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &cloudletInfo.Key, &cloudletInfo) {
			panic("couldnt find cloudlet") // XXX duh
			//return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		return nil
	})

	curFlavors = cloudletInfo.Flavors
	fmt.Printf("TEST--- back from Update #2 num flavors should == 13 len(curFlavors)= %d \n", len(curFlavors))
	dumpFlavors("curFlavs after update #2", curFlavors)
	// check the marking was applied
	for _, flavor := range curFlavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "not deprecated")
			break
		}
	}
	// Since the infra-flavor was being used by our app, rather than being removed
	// during the update, it was left in marked deprecated.
	require.Equal(t, 13, len(curFlavors), "should be present but deprecated")
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
	for k, _ := range alertApi.cache.Objs {

		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			alertFound = true
			alerts++
		}
	}
	if alerts > 1 {
		panic("Found muliple pending alerts should be one")
	}
	require.True(t, alertFound, "failed to find infra flavor deleted alert")
	// I should be able to create an alert key, and use Get to test existence. Didn't work hm

	// Now test flavor re-mapping, with tiny2 deprecated, GetVMSpec will skip this infra flavor
	// GetVMSpec() on our meta flavor x1.tiny that would match our infra flavor.tiny2 if were not deprecated

	fmt.Printf("\n\n\tTEST update 2 delete in use tiny2 deprecated ====> PASS \n")

	dumpFlavors("curFlavors after deprecation", curFlavors)

	fm := edgeproto.FlavorMatch{
		Key:        cloudletInfo.Key,
		FlavorName: testutil.FlavorData[0].Key.Name,
	}
	match, err := cloudletApi.FindFlavorMatch(ctx, &fm)
	require.Nil(t, err, "FindFlavorMatch")
	require.NotEqual(t, match.FlavorName, "flavor.tiny2")
	require.Equal(t, "flavor.small", match.FlavorName, "flavor remapping not as expected test data?")

	// Now delete an infra flavor that is not being used, this should actually remove it.
	// curFlavors = cloudletInfo.Flavors
	// tiny2 is still deprecated, and not on the newFlavors list
	require.Equal(t, 13, len(curFlavors), "should be present but deprecated")
	require.Equal(t, 12, len(newFlavors), "newFlavor count off")
	fmt.Printf("TEST FindFlavorMatch remapping PASS\n")

	// we tested deprecation / remapping and alert generation
	// now delete an unsed flavor, it should go

	for ii, flavor := range newFlavors {
		if flavor.Name == "flavor.doNotUse" {
			newFlavors[ii] = newFlavors[len(newFlavors)-1]
			newFlavors[len(newFlavors)-1] = &edgeproto.FlavorInfo{}
			newFlavors = newFlavors[:len(newFlavors)-1]
		}
	}
	require.Equal(t, 11, len(newFlavors), "newFlavor test count unexpected")
	require.Equal(t, 13, len(cloudletInfo.Flavors), "cloudlet flavors count off did the delete fail?")

	// back from FindFlavorMatch, where did flavor.large-pci go from newFlavors?
	// no, it's there!
	dumpFlavors("TEst after findflavor match newFlavors should have flavor.large-pci", newFlavors)

	dumpFlavors("curCloudletInfo.Flavors:", cloudletInfo.Flavors)
	// Right Here! Fucking cloudletInfo.Flavors has an empty flavor, though the last update shows they were all accounted for.

	// reset newFlavors to testFlavors, and deprecate tiny2 to get back on course...
	reset := false
	for i, f := range cloudletInfo.Flavors {
		if f.Name == "" {
			fmt.Printf("TEST-W-found empty flavor at index %d in cloudletFlavors resetting\n", i)
			reset = true
			break
		}
	}
	if reset {
		cloudletInfo.Flavors = testFlavors
		for _, f := range cloudletInfo.Flavors {
			if f.Name == "flavor.tiny2" {
				f.Deprecated = true
			}
		}
	}

	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	time.Sleep(1 * time.Second)
	err = cloudletInfoApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &cloudletInfo.Key, &cloudletInfo) {
			panic("couldnt find cloudlet") // XXX duh
			//return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		return nil
	})
	dumpFlavors("curCloudletInfo.Flavors:", cloudletInfo.Flavors)
	reset = false

	// now deleted do not use
	fmt.Printf("\n\nTEST Delete flavor.doNotUse newFlavors\n\n")
	dumpFlavors("newFlavors for delete doNotUse Test", newFlavors)
	cloudletInfo.Flavors = newFlavors

	fmt.Printf("\n\n\tTEST update 3 delete not in use doNotUse flavor  update\n\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	time.Sleep(1 * time.Second)
	err = cloudletInfoApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &cloudletInfo.Key, &cloudletInfo) {
			panic("couldnt find cloudlet") // XXX duh
			//return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		return nil
	})
	curFlavors = cloudletInfo.Flavors
	fmt.Printf("\n\n--------------TEST BAck from delete doNotUse Update curFlavors\n\n")
	for _, f := range curFlavors {
		fmt.Printf("\t%+v\n", f)
	}
	require.Equal(t, 12, len(curFlavors), "Update failed to remove flavor.doNotUse")

	found = false
	for _, infraFlavor := range curFlavors {
		if infraFlavor.Name == "flavor.doNotUse" {
			found = true
		}
	}
	require.Equal(t, false, found, "Update failed flavor.doNotUse not deleted but something was!")

	// Present the now deprecated flavor.tiny2 back to newFlavors.
	// Putting back in a flavor that was (mistakenly?) deleted, or
	// simply changing its size etc. The deprecated flavor should be cleared,
	// and if the flavors values were changed, they will be applied,
	// An update event will be issued in this case.

	fmt.Printf("\n\n TEST------------ Adding back flavor %+v should clear via recreated flavor -----------------------------\n\n", testutil.CloudletInfoData[0].Flavors[1])
	newFlavors = append(newFlavors, testutil.CloudletInfoData[0].Flavors[1]) // is this tiny2?
	require.Equal(t, 12, len(newFlavors), "unexpected len of newFlavors")

	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\n\tTEST update 4 recreate tiny2  update\n\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	time.Sleep(1 * time.Second)
	err = cloudletInfoApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletInfoApi.store.STMGet(stm, &cloudletInfo.Key, &cloudletInfo) {
			panic("couldnt find cloudlet") // XXX duh
			//return fmt.Errorf("CloudletInfo %s Not found", key.cloudletKey)
		}
		return nil
	})
	// At this point, we have 3 flavor.tiny1 and one tiny2 that is still deprecated. fuck
	curFlavors = cloudletInfo.Flavors
	if len(curFlavors) != 12 {
		fmt.Printf("\n\nTEST we seem to have an extra flavor after adding back tiny2 is it in there twice:\n")
		for i, f := range curFlavors {
			fmt.Printf("\tcurFlavoars[%d] = %+v\n", i, f)
		}
	}

	require.Equal(t, 12, len(curFlavors), "Update failed")

	for _, flavor := range curFlavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "flavor deprecation mark not cleared")
			break
		}
	}
	for _, flavor := range cloudletInfo.DeprecatedFlavors {
		if flavor.Name == "flavor.tiny2" {
			found = true
			break
		}
	}
	require.Equal(t, true, found, "flavor not removed from cloudletInfo.DeprecatedFlavors list")
	// check the InfraFlavorDeleted alert, it should have been cleared by restoring
	// the deprecated flavor.

	alertFound = false
	for k, _ := range alertApi.cache.Objs {
		fmt.Printf("next keystring: %s\n", k.GetKeyString())
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			alertFound = true
		}
	}
	require.Equal(t, false, alertFound, "Alert not cleared by flavor re-creation")

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
	require.Equal(t, 11, len(newFlavors), "newFlav len off")
	cloudletInfo.Flavors = newFlavors
	fmt.Printf("\n\n\tTEST update 5  remove tiny2 again need another alert  update\n\n")
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	curFlavors = cloudletInfo.Flavors
	require.Equal(t, 12, len(curFlavors), "should be present but deprecated")

	for _, flavor := range newFlavors {
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

	// Ok, with a dangling deprecated flavor, delete the users of it, our task should notice and clear the alert...

	// first delete the clusterInsts, I suppose that will take out any appInsts there

	// 15 = {AppKey:{Organization:MobiledgeX Name:SampleApp Version:1.0.0} ClusterInstKey:{ClusterKey:{Name:autocluster-mt3} CloudletKey:{Organization:AT&T Inc. Name:San Jose Site} Organization:MobiledgeX}}

	// Now we start exercising the worker task called from the using objects as they are deleted
	fmt.Printf("\n\nTEST: ------------------ delete pokeAppInst:------------------\n\n")
	err = appInstApi.DeleteAppInst(&pokeInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Error Delete poke AppInst")

	fmt.Printf("\n\nTEST deleting PokeClust crmOverride: %+v \n\n", pokeClust.CrmOverride)

	pokeClust.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.DeleteClusterInst(&pokeClust, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete poke ClusterInst")
	check = edgeproto.ClusterInst{}
	foundclust = clusterInstApi.cache.Get(&pokeClust.Key, &check)
	require.False(t, foundclust)

	// Ok, this is insuffienct to get the darn DeletedCb() to trigger, only an STMDel on this object will do that, so do that.

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

	for k, _ := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			alertFound = true
		}
	}
	require.Equal(t, false, alertFound, "Alert not cleared by deleting objects")

	_, err = cloudletInfoApi.EvictCloudletInfo(ctx, &cloudletInfo)
	require.Nil(t, err, "Evict")

	dummy.Stop()
}
