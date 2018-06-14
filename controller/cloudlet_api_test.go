package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestCloudletApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	operApi := InitOperatorApi(&dummy)
	api := InitCloudletApi(&dummy, operApi)
	operApi.WaitInitDone()
	api.WaitInitDone()

	// cannot create cloudlets without apps
	ctx := context.TODO()
	for _, obj := range testutil.CloudletData {
		_, err := api.CreateCloudlet(ctx, &obj)
		assert.NotNil(t, err, "Create cloudlet without operator")
	}

	// create operators
	for _, obj := range testutil.OperatorData {
		_, err := operApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}

	testutil.InternalCloudletCudTest(t, api, testutil.CloudletData)
	dummy.Stop()
}
