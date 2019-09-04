package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAppApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)
	tMode := true
	testMode = &tMode
	log.InitTracer("")
	defer log.FinishTracer()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannot create apps without developer
	ctx := log.StartTestSpan(context.Background())
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		require.NotNil(t, err, "Create app without developer")
	}

	// create support data
	testutil.InternalDeveloperCreate(t, &developerApi, testutil.DevData)
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)

	testutil.InternalAppTest(t, "cud", &appApi, testutil.AppData)

	obj := testutil.AppData[3]
	_, err := appApi.DeleteApp(ctx, &obj)
	require.Nil(t, err)

	// vmapp with http should fail
	vmapp := testutil.AppData[3]
	vmapp.AccessPorts = "http:443"
	_, err = appApi.CreateApp(ctx, &vmapp)
	require.NotNil(t, err, "Create vmapp with http port")
	require.Contains(t, err.Error(), "Deployment Type and HTTP access ports are incompatible")
	vmapp.AccessPorts = "HTTP:443"
	_, err = appApi.CreateApp(ctx, &vmapp)
	require.NotNil(t, err, "Create vmapp with http port")
	require.Contains(t, err.Error(), "Deployment Type and HTTP access ports are incompatible")

	dummy.Stop()
}
