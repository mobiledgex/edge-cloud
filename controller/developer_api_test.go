package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
)

func TestDeveloperApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalDeveloperCudTest(t, &developerApi, testutil.DevData)

	dummy.Stop()
}
