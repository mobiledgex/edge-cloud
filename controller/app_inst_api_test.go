package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestAppInstApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	operApi := InitOperatorApi(&dummy)
	cloudletApi := InitCloudletApi(&dummy, operApi)
	devApi := InitDeveloperApi(&dummy)
	appApi := InitAppApi(&dummy, devApi)
	api := InitAppInstApi(&dummy, appApi, cloudletApi)
	operApi.WaitInitDone()
	cloudletApi.WaitInitDone()
	devApi.WaitInitDone()
	appApi.WaitInitDone()
	api.WaitInitDone()

	// cannote create instances without apps and cloudlets
	ctx := context.TODO()
	for _, obj := range testutil.AppInstData {
		_, err := api.CreateAppInst(ctx, &obj)
		assert.NotNil(t, err, "Create app inst without apps/cloudlets")
	}

	// create supporting data
	for _, obj := range testutil.DevData {
		_, err := devApi.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create developer")
	}
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		assert.Nil(t, err, "Create app")
	}
	for _, obj := range testutil.OperatorData {
		_, err := operApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}
	for _, obj := range testutil.CloudletData {
		_, err := cloudletApi.CreateCloudlet(ctx, &obj)
		assert.Nil(t, err, "Create cloudlet")
	}

	testutil.InternalAppInstCudTest(t, api, testutil.AppInstData)
	dummy.Stop()
}
