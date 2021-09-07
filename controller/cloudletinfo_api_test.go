package main

import (
	"context"
	"strings"
	"testing"

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
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
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
	flavorMap := make(map[string]edgeproto.FlavorInfo)
	newFlavorMap := make(map[string]edgeproto.FlavorInfo)

	for _, flavor := range testutil.CloudletInfoData[0].Flavors {
		flavorMap[flavor.Name] = *flavor
		newFlavorMap[flavor.Name] = *flavor
	}
	require.Equal(t, 13, len(newFlavorMap), "testutil.FlavorInfo has changed")

	// If we ask right now, we should see no changes, most common case.
	addedFlavors, deletedFlavors, updatedFlavors := cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)

	require.Equal(t, 0, len(addedFlavors), "findFlavorDeltas")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas")

	newFlavor1 := edgeproto.FlavorInfo{
		Name:       "newFlavor1",
		Vcpus:      uint64(10),
		Ram:        uint64(4096),
		Disk:       uint64(80),
		Deprecated: false,
	}
	newFlavor2 := edgeproto.FlavorInfo{
		Name:       "newFlavor2",
		Vcpus:      uint64(6),
		Ram:        uint64(8192),
		Disk:       uint64(60),
		Deprecated: false,
	}

	// case 1 addFlavor
	newFlavorMap[newFlavor1.Name] = newFlavor1
	addedFlavors, deletedFlavors, updatedFlavors = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)

	require.Equal(t, 1, len(addedFlavors), "findFlavorDeltas add")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas add")
	require.Equal(t, 0, len(updatedFlavors), "findFlavorDeltas add")
	require.Equal(t, "newFlavor1", addedFlavors[0].Name, "findFlavorDeltas add")

	// case 2 deleteFlavor
	delete(newFlavorMap, "newFlavor1")

	addedFlavors, deletedFlavors, updatedFlavors = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
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

	newFlavorMap["flavor.tiny1"] = newTiny1

	addedFlavors, deletedFlavors, updatedFlavors = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 0, len(addedFlavors), "findFlavorDeltas update")
	require.Equal(t, 0, len(deletedFlavors), "findFlavorDeltas update")
	require.Equal(t, 1, len(updatedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny1", updatedFlavors[0].Name, "findFlavorDeltas update")

	delete(newFlavorMap, "newFlavor1")
	// case 4 one change of each type simultaniously
	// Delete flavor.tiny1, add newFlavor1, and update tiny2 to use 1024 ram and Vcpus to 2
	// start with them equal

	for _, flavor := range testutil.CloudletInfoData[0].Flavors {
		flavorMap[flavor.Name] = *flavor
		newFlavorMap[flavor.Name] = *flavor
	}
	require.Equal(t, 13, len(newFlavorMap), "testutil.FlavorInfo has changed")
	require.Equal(t, 13, len(flavorMap), "testutil.FlavorInfo has changed")

	// Delete tiny1
	delete(newFlavorMap, "flavor.tiny1")
	// add newFlavor1
	newFlavorMap[newFlavor1.Name] = newFlavor1

	newTiny2 := edgeproto.FlavorInfo{
		Name:  "flavor.tiny2",
		Vcpus: uint64(2),
		Ram:   uint64(1024),
		Disk:  uint64(10),
	}
	// update tiny2
	newFlavorMap[newTiny2.Name] = newTiny2

	addedFlavors, deletedFlavors, updatedFlavors = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
	require.Equal(t, 1, len(addedFlavors), "findFlavorDeltas update")
	require.Equal(t, "newFlavor1", addedFlavors[0].Name, "findFlavorDeltas multi")
	require.Equal(t, 1, len(deletedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny1", deletedFlavors[0].Name, "findFlavorDeltas multi")
	require.Equal(t, 1, len(updatedFlavors), "findFlavorDeltas update")
	require.Equal(t, "flavor.tiny2", updatedFlavors[0].Name, "findFlavorDeltas multi")

	// case 5 two adds at a time
	newFlavorMap = nil
	flavorMap = nil
	flavorMap = make(map[string]edgeproto.FlavorInfo)
	newFlavorMap = make(map[string]edgeproto.FlavorInfo)
	for _, flavor := range testutil.CloudletInfoData[0].Flavors {
		flavorMap[flavor.Name] = *flavor
		newFlavorMap[flavor.Name] = *flavor
	}
	require.Equal(t, 13, len(newFlavorMap), "testutil.FlavorInfo has changed")
	require.Equal(t, 13, len(flavorMap), "testutil.FlavorInfo has changed")

	newFlavorMap[newFlavor1.Name] = newFlavor1
	newFlavorMap[newFlavor2.Name] = newFlavor2

	addedFlavors, deletedFlavors, updatedFlavors = cloudletInfoApi.findFlavorDeltas(ctx, flavorMap, newFlavorMap)
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

	//
	for _, alert := range testutil.AlertData {
		alertApi.Update(ctx, &alert, 0)
	}

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)

	// Inject the cloudletInfo test data for Update to fetch
	cldInfo := testutil.CloudletInfoData[0]
	_, err := cloudletInfoApi.InjectCloudletInfo(ctx, &cldInfo)
	require.Nil(t, err)

	// Apps are independent of cloudlet, and is only used to check for meta-flavors in use.
	// Creating an App with the meta-flavor we know will match our target infra-flavor that we will then
	// delete from our CloudletInfo object, and call Update with it. Should result in it not being
	// removed during the update but rather marked as deprecated.
	app := testutil.AppData[2]
	app.DefaultFlavor = testutil.FlavorData[0].Key

	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "CreateApp")

	cloudletInfo := testutil.CloudletInfoData[0]
	curFlavs := cloudletInfo.Flavors

	// delete the infra-flavor that matches what our targetApp's meta-flavor is using
	for ii, flavor := range curFlavs {
		if flavor.Name == "flavor.tiny2" {
			curFlavs[ii] = curFlavs[len(curFlavs)-1]
			curFlavs[len(curFlavs)-1] = &edgeproto.FlavorInfo{}
			curFlavs = curFlavs[:len(curFlavs)-1]
		}
	}
	// assign modified flavors back to our cloudletInfo that we're about to update,
	// simulates what we want out of GatherCloudletInfo() done by crm's infra flavor update thread.
	cloudletInfo.Flavors = curFlavs
	// and call for the update
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	// grab the newly updated CloudletInfo's flavors
	newFlavors := cloudletInfo.Flavors
	// Since the infra-flavor was being used by our app, rather than being removed
	// during the update, it was left in marked deprecated.
	for _, flavor := range newFlavors {
		if flavor.Name == "flavor.tiny2" {
			require.Equal(t, true, flavor.Deprecated, "not deprecated")
		}
	}
	// Verify our test generated the desired Alert
	alertFound := false
	for k, _ := range alertApi.cache.Objs {
		if strings.Contains(k.GetKeyString(), "InfraFlavorDeleted") {
			alertFound = true
		}
	}
	require.True(t, alertFound, "failed to find infra flavor deleted alert")
	// Now delete an infra flavor that is not being used, this should actually remove it
	// An easy taraget would be one of the gpu infra-flavors
	// Not so easy, on't use flavor.large-generic-gpu it messes with clusterInst tests
	// Though it should be ok when I put it back in, but it is not. weird.

	for ii, flavor := range curFlavs {
		if flavor.Name == "flavor.doNotUse" {
			curFlavs[ii] = curFlavs[len(curFlavs)-1]
			curFlavs[len(curFlavs)-1] = &edgeproto.FlavorInfo{}
			curFlavs = curFlavs[:len(curFlavs)-1]
		}
	}
	cloudletInfo.Flavors = curFlavs
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)

	newFlavors = cloudletInfo.Flavors
	require.Equal(t, 11, len(newFlavors), "Update failed")
	var found bool
	for _, infraFlavor := range newFlavors {
		if infraFlavor.Name == "flavor.doNotUse" {
			found = true
		}
	}
	require.Equal(t, false, found, "Update failed")

	// Now put what we've delelted here back, it's used elsewhere in clousterInstApiTest
	//newFlavors = append(newFlavors, testutil.CloudletInfoData[0].Flavors[10])

	// put stuff back the way we found it
	cloudletInfo.Flavors = curFlavs
	cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	_, err = cloudletInfoApi.EvictCloudletInfo(ctx, &cloudletInfo)
	require.Nil(t, err, "Evict")

	dummy.Stop()
	// TODO: continue to create other objs (appInst, clusterInst, vmPool vms)
	// and ensure they also trigger flavor deprecation. See IsCloudletUsingFlavors()
	// Other cases:  multiple in use meta-flavors should all be marked deprecated.
	// And alerts, once we get them on their feet here, we should check the alertCache
	// to ensure they've been published. (And then delete them)
}
