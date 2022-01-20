package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAutoScalePolicyApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testSvcs := testinit(t)
	defer testfinish(testSvcs)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalAutoScalePolicyTest(t, "cud", apis.autoScalePolicyApi, testutil.AutoScalePolicyData)

	policy := edgeproto.AutoScalePolicy{}
	policy.Key.Name = "auto-scale-policy-name"
	policy.Key.Organization = "dev1"
	policy.MinNodes = 1
	policy.MaxNodes = 2
	policy.ScaleUpCpuThresh = 80
	policy.ScaleDownCpuThresh = 20
	policy.TriggerTimeSec = 60

	// test invalid bounds
	p := policy
	p.MaxNodes = 100
	expectBadAutoScaleCreate(t, ctx, apis, &p, "Max nodes cannot exceed")
	p = policy
	p.ScaleUpCpuThresh = 101
	expectBadAutoScaleCreate(t, ctx, apis, &p, "must be between 0 and 100")
	p = policy
	p.ScaleDownCpuThresh = 900
	expectBadAutoScaleCreate(t, ctx, apis, &p, "must be between 0 and 100")
	p = policy
	p.MinNodes = 5
	p.MaxNodes = 5
	expectBadAutoScaleCreate(t, ctx, apis, &p, "Max nodes must be greater than Min")
	p = policy
	p.ScaleUpCpuThresh = 50
	p.ScaleDownCpuThresh = 60
	expectBadAutoScaleCreate(t, ctx, apis, &p, "Scale down cpu threshold must be less than scale up")

	dummy.Stop()
}

func expectBadAutoScaleCreate(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.AutoScalePolicy, msg string) {
	_, err := apis.autoScalePolicyApi.CreateAutoScalePolicy(ctx, in)
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
