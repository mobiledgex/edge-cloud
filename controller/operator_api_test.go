package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
)

func TestOperatorApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)
	InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	api := InitOperatorApi(&dummy)

	testutil.InternalOperatorCudTest(t, api, OperatorData)
	dummy.Stop()
}
