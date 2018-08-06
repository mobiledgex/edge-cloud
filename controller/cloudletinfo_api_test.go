package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
)

func TestCloudletInfo(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)

	testutil.InternalCloudletInfoTest(t, "show", &cloudletInfoApi, testutil.CloudletInfoData)
	dummy.Stop()
}

func insertCloudletInfo(data []edgeproto.CloudletInfo) {
	for ii, _ := range data {
		in := &data[ii]
		cloudletInfoApi.Update(in, 0)
	}
}
