package main

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClusterInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	objstore.InitRegion(1)
	reduceInfoTimeouts()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()
	responder := NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	// cannot create insts without cluster/cloudlet
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
		assert.NotNil(t, err, "Create cluster inst without cloudlet")
	}

	// create support data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalClusterFlavorCreate(t, &clusterFlavorApi, testutil.ClusterFlavorData)
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)
	testutil.InternalClusterCreate(t, &clusterApi, testutil.ClusterData)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the cluster insts will fail.
	responder.SetSimulateCreateFailure(true)
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
		assert.NotNil(t, err, "Create cluster inst responder failures")
		// make sure error matches responder
		assert.Equal(t, "Encountered failures: [crm create cluster inst failed]", err.Error())
	}
	responder.SetSimulateCreateFailure(false)
	assert.Equal(t, 0, len(clusterInstApi.cache.Objs))

	testutil.InternalClusterInstTest(t, "cud", &clusterInstApi, testutil.ClusterInstData)
	// after cluster insts create, check that cloudlet refs data is correct.
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)

	dummy.Stop()
}

func reduceInfoTimeouts() {
	CreateClusterInstTimeout = 1 * time.Second
	UpdateClusterInstTimeout = 1 * time.Second
	DeleteClusterInstTimeout = 1 * time.Second

	CreateAppInstTimeout = 1 * time.Second
	UpdateAppInstTimeout = 1 * time.Second
	DeleteAppInstTimeout = 1 * time.Second
}
