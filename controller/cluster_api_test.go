package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
)

func TestClusterApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create support data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalClusterFlavorCreate(t, &clusterFlavorApi, testutil.ClusterFlavorData)

	testutil.InternalClusterTest(t, "cud", &clusterApi, testutil.ClusterData)

	dummy.Stop()
}
