package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestDeveloperApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	sql, err := EnsureCleanSql()
	require.Nil(t, err, "sql")
	defer sql.Close()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync, sql)
	sync.Start()
	defer sync.Done()

	testutil.InternalDeveloperTest(t, "cud", &developerApi, testutil.DevData)

	dummy.Stop()
}
