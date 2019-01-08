package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCloudletApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
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

	// test invalid location values
	clbad := testutil.CloudletData[0]
	clbad.Key.Name = "bad loc"
	testBadLat(t, &clbad, []float64{90.1, -90.1, -1323213, 1232334})
	testBadLong(t, &clbad, []float64{180.1, -180.1, -1323213, 1232334})

	clbad = testutil.CloudletData[0]
	clbad.Key.Name = "test num dyn ips"
	err := cloudletApi.CreateCloudlet(&clbad, &testutil.CudStreamoutCloudlet{})
	assert.Nil(t, err)
	clbad.NumDynamicIps = 0
	clbad.Fields = []string{edgeproto.CloudletFieldNumDynamicIps}
	err = cloudletApi.UpdateCloudlet(&clbad, &testutil.CudStreamoutCloudlet{})
	assert.NotNil(t, err)

	dummy.Stop()
}

func testBadLat(t *testing.T, clbad *edgeproto.Cloudlet, lats []float64) {
	for _, lat := range lats {
		clbad.Location.Latitude = lat
		err := cloudletApi.CreateCloudlet(clbad, &testutil.CudStreamoutCloudlet{})
		assert.NotNil(t, err, "create cloudlet bad latitude")
	}
}

func testBadLong(t *testing.T, clbad *edgeproto.Cloudlet, longs []float64) {
	for _, long := range longs {
		clbad.Location.Longitude = long
		err := cloudletApi.CreateCloudlet(clbad, &testutil.CudStreamoutCloudlet{})
		assert.NotNil(t, err, "create cloudlet bad longitude")
	}
}
