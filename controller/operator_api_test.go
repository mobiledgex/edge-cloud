package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
)

func TestOperatorApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	api := InitOperatorApi(&dummy)
	api.WaitInitDone()
	testutil.InternalOperatorCudTest(t, api, testutil.OperatorData)
	dummy.Stop()
}
