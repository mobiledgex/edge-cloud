package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestOperatorCodeApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	testinit()
	defer testfinish()
	log.InitTracer(nil)
	defer log.FinishTracer()

	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	ctx := log.StartTestSpan(context.Background())

	testutil.InternalOperatorCodeTest(t, "cud", &operatorCodeApi, testutil.OperatorCodeData)
	// create duplicate key should fail
	code := testutil.OperatorCodeData[0]
	_, err := operatorCodeApi.CreateOperatorCode(ctx, &code)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "already exists")

	// check not found error on delete
	_, err = operatorCodeApi.DeleteOperatorCode(ctx, &code)
	require.Nil(t, err)
	_, err = operatorCodeApi.DeleteOperatorCode(ctx, &code)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not found")
}
