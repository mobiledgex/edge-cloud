package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestAppApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)
	InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	devApi := InitDeveloperApi(&dummy)
	api := InitAppApi(&dummy, devApi)

	// cannot create apps without developer
	ctx := context.TODO()
	for _, obj := range AppData {
		_, err := api.CreateApp(ctx, &obj)
		assert.NotNil(t, err, "Create app without developer")
	}

	// create developers
	for _, obj := range DevData {
		_, err := devApi.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create developer")
	}

	testutil.InternalAppCudTest(t, api, AppData)

	dummy.Stop()
}
