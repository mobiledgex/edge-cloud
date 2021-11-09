package main

import (
	"context"
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
)

func TestCloudletInfo(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)

	testutil.InternalCloudletInfoTest(t, "show", apis.cloudletInfoApi, testutil.CloudletInfoData)
	dummy.Stop()
}

func insertCloudletInfo(ctx context.Context, apis *AllApis, data []edgeproto.CloudletInfo) {
	for ii := range data {
		in := &data[ii]
		in.State = dme.CloudletState_CLOUDLET_STATE_READY
		apis.cloudletInfoApi.Update(ctx, in, 0)
	}
}

func evictCloudletInfo(ctx context.Context, apis *AllApis, data []edgeproto.CloudletInfo) {
	for ii := range data {
		in := &data[ii]
		apis.cloudletInfoApi.Delete(ctx, in, 0)
	}
}
