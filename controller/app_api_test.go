package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAppApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannot create apps without developer
	ctx := context.TODO()
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		assert.NotNil(t, err, "Create app without developer")
	}

	// create developers
	for _, obj := range testutil.DevData {
		_, err := developerApi.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create developer")
	}
	for _, obj := range testutil.FlavorData {
		_, err := flavorApi.CreateFlavor(ctx, &obj)
		assert.Nil(t, err, "Create flavor")
	}
	for _, obj := range testutil.ClusterData {
		_, err := clusterApi.CreateCluster(ctx, &obj)
		assert.Nil(t, err, "Create cluster")
	}

	testutil.InternalAppCudTest(t, &appApi, testutil.AppData)

	dummy.Stop()
}
