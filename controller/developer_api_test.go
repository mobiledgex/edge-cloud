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

	api := InitDeveloperApi(&dummy)
	testutil.InternalDeveloperCudTest(t, api, testutil.DevData)

	dummy.Stop()
}
