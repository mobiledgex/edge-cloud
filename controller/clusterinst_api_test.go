package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClusterInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannot create insts without cluster/cloudlet
	ctx := context.TODO()
	for _, obj := range testutil.ClusterInstData {
		_, err := clusterInstApi.CreateClusterInst(ctx, &obj)
		assert.NotNil(t, err, "Create cluster inst without cloudlet")
	}

	// create support data
	for _, obj := range testutil.OperatorData {
		_, err := operatorApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}
	for _, obj := range testutil.CloudletData {
		_, err := cloudletApi.CreateCloudlet(ctx, &obj)
		assert.Nil(t, err, "Create cloudlet")
	}
	for _, obj := range testutil.FlavorData {
		_, err := flavorApi.CreateFlavor(ctx, &obj)
		assert.Nil(t, err, "Create flavor")
	}
	for _, obj := range testutil.ClusterData {
		_, err := clusterApi.CreateCluster(ctx, &obj)
		assert.Nil(t, err, "Create cluster")
	}

	testutil.InternalClusterInstCudTest(t, &clusterInstApi, testutil.ClusterInstData)

	dummy.Stop()
}
