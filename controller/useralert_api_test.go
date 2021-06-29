package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestUserAlertApi(t *testing.T) {
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

	testutil.InternalUserAlertTest(t, "cud", &userAlertApi, testutil.UserAlertData)

	// invalid user alert
	userAlert := testutil.UserAlertData[0]
	userAlert.ActiveConnLimit = 10

	_, err := userAlertApi.CreateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Both active connections and cpu cannot be set for a user alert")
}
