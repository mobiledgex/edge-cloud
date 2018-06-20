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

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannot create cloudlets without apps
	ctx := context.TODO()
	for _, obj := range testutil.CloudletData {
		_, err := cloudletApi.CreateCloudlet(ctx, &obj)
		assert.NotNil(t, err, "Create cloudlet without operator")
	}

	// create operators
	for _, obj := range testutil.OperatorData {
		_, err := operatorApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}

	testutil.InternalCloudletCudTest(t, &cloudletApi, testutil.CloudletData)
	dummy.Stop()
}
