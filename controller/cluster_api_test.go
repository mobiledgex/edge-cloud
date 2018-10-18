package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestClusterApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	sql, err := EnsureCleanSql()
	require.Nil(t, err, "sql connect")
	defer sql.Close()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync, sql)
	sync.Start()
	defer sync.Done()

	testutil.InternalClusterTest(t, "cud", &clusterApi, testutil.ClusterData)

	dummy.Stop()
}
