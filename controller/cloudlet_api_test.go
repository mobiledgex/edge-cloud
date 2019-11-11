package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

type stateTransition struct {
	triggerState  edgeproto.CloudletState
	expectedState edgeproto.TrackedState
}

func TestCloudletApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create operators
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	// create flavors
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)

	testutil.InternalCloudletTest(t, "cud", &cloudletApi, testutil.CloudletData)

	// test invalid location values
	clbad := testutil.CloudletData[0]
	clbad.Key.Name = "bad loc"
	testBadLat(t, ctx, &clbad, []float64{0, 90.1, -90.1, -1323213, 1232334}, "create")
	testBadLong(t, ctx, &clbad, []float64{0, 180.1, -180.1, -1323213, 1232334}, "create")

	clbad = testutil.CloudletData[0]
	clbad.Key.Name = "test num dyn ips"
	err := cloudletApi.CreateCloudlet(&clbad, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	clbad.NumDynamicIps = 0
	clbad.Fields = []string{edgeproto.CloudletFieldNumDynamicIps}
	err = cloudletApi.UpdateCloudlet(&clbad, testutil.NewCudStreamoutCloudlet(ctx))
	require.NotNil(t, err)

	cl := testutil.CloudletData[1]
	cl.Key.Name = "test invalid lat-long"
	err = cloudletApi.CreateCloudlet(&cl, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	testBadLat(t, ctx, &cl, []float64{0, 90.1, -90.1, -1323213, 1232334}, "update")
	testBadLong(t, ctx, &cl, []float64{0, 180.1, -180.1, -1323213, 1232334}, "update")

	testCloudletUpgrade(t, ctx)

	dummy.Stop()
}

func testBadLat(t *testing.T, ctx context.Context, clbad *edgeproto.Cloudlet, lats []float64, action string) {
	for _, lat := range lats {
		clbad.Location.Latitude = lat
		clbad.Fields = []string{edgeproto.CloudletFieldLocationLatitude}
		switch action {
		case "create":
			err := cloudletApi.CreateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "create cloudlet bad latitude")
		case "update":
			err := cloudletApi.UpdateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "update cloudlet bad latitude")
		}
	}
}

func testBadLong(t *testing.T, ctx context.Context, clbad *edgeproto.Cloudlet, longs []float64, action string) {
	for _, long := range longs {
		clbad.Location.Longitude = long
		clbad.Fields = []string{edgeproto.CloudletFieldLocationLongitude}
		switch action {
		case "create":
			err := cloudletApi.CreateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "create cloudlet bad longitude")
		case "update":
			err := cloudletApi.CreateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "update cloudlet bad longitude")
		}
	}
}

func WaitForState(key *edgeproto.CloudletKey, state edgeproto.TrackedState, count int) error {
	lastState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
	for i := 0; i < count; i++ {
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.cache.Get(key, &cloudlet) {
			return fmt.Errorf("unable to find cloudlet")
		}
		if cloudlet.State == state {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
		lastState = cloudlet.State
	}

	return fmt.Errorf("Unable to get desired cloudlet state, actual state %s, desired state %s", lastState, state)
}

func forceCloudletInfoState(ctx context.Context, key *edgeproto.CloudletKey, state edgeproto.CloudletState) {
	info := edgeproto.CloudletInfo{}
	info.Key = *key
	info.State = state
	cloudletInfoApi.cache.Update(ctx, &info, 0)
}

func testCloudletUpgrade(t *testing.T, ctx context.Context) {
	var stateTransitions []stateTransition

	stateTransitions = []stateTransition{
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_UPGRADE,
			expectedState: edgeproto.TrackedState_UPDATING,
		},
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_INIT,
			expectedState: edgeproto.TrackedState_CRM_INITOK,
		},
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_READY,
			expectedState: edgeproto.TrackedState_READY,
		},
	}
	testUpgradeScenario(t, ctx, &stateTransitions, "success")

	stateTransitions = []stateTransition{
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_INIT,
			expectedState: edgeproto.TrackedState_CRM_INITOK,
		},
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_READY,
			expectedState: edgeproto.TrackedState_READY,
		},
	}
	testUpgradeScenario(t, ctx, &stateTransitions, "success")

	stateTransitions = []stateTransition{
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_UPGRADE,
			expectedState: edgeproto.TrackedState_UPDATING,
		},
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_INIT,
			expectedState: edgeproto.TrackedState_CRM_INITOK,
		},
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_ERRORS,
			expectedState: edgeproto.TrackedState_UPDATE_ERROR,
		},
	}
	testUpgradeScenario(t, ctx, &stateTransitions, "fail")

	stateTransitions = []stateTransition{
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_UPGRADE,
			expectedState: edgeproto.TrackedState_UPDATING,
		},
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_ERRORS,
			expectedState: edgeproto.TrackedState_UPDATE_ERROR,
		},
	}
	testUpgradeScenario(t, ctx, &stateTransitions, "fail")

	stateTransitions = []stateTransition{
		stateTransition{
			triggerState:  edgeproto.CloudletState_CLOUDLET_STATE_ERRORS,
			expectedState: edgeproto.TrackedState_UPDATE_ERROR,
		},
	}
	testUpgradeScenario(t, ctx, &stateTransitions, "fail")
}

func testUpgradeScenario(t *testing.T, ctx context.Context, transitions *[]stateTransition, scenario string) {
	var err error
	cloudlet := testutil.CloudletData[2]
	err = cloudletApi.CreateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)

	go func() {
		forceCloudletInfoState(ctx, &cloudlet.Key, edgeproto.CloudletState_CLOUDLET_STATE_READY)
		err := cloudletApi.UpgradeCloudlet(ctx, &cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
		if scenario == "fail" {
			require.NotNil(t, err, "upgrade cloudlet should fail")
		} else {
			require.Nil(t, err, "upgrade cloudlet should succeed")
		}
	}()

	err = WaitForState(&cloudlet.Key, edgeproto.TrackedState_UPDATE_REQUESTED, 10)
	require.Nil(t, err, "cloudlet state transtions")

	for _, transition := range *transitions {
		forceCloudletInfoState(ctx, &cloudlet.Key, transition.triggerState)
		err = WaitForState(&cloudlet.Key, transition.expectedState, 10)
		require.Nil(t, err, fmt.Sprintf("cloudlet state transtions for %s scenario", scenario))
	}

	err = cloudletApi.DeleteCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
}
