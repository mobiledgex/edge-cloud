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

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalOperatorCudTest(t, &operatorApi, testutil.OperatorData)
	dummy.Stop()
}
