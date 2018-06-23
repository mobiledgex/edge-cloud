package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestAppApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)
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

	testutil.InternalAppCudTest(t, &appApi, testutil.AppData)

	dummy.Stop()
}
