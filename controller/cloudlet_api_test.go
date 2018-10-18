package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudletApi(t *testing.T) {
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

	// cannot create cloudlets without apps
	for _, obj := range testutil.CloudletData {
		err := cloudletApi.CreateCloudlet(&obj, &testutil.CudStreamoutCloudlet{})
		assert.NotNil(t, err, "Create cloudlet without operator")
	}

	// create operators
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)

	testutil.InternalCloudletTest(t, "cud", &cloudletApi, testutil.CloudletData)
	dummy.Stop()
}
