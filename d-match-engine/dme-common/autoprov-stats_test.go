package dmecommon

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestAutoProvStats(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq | log.DebugLevelMetrics)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	org := "org"
	app := edgeproto.App{
		Key: edgeproto.AppKey{
			Organization: org,
			Name:         "someapp",
			Version:      "1.0",
		},
	}
	carrier := "oper"

	// cloudlets for policies
	apCloudlets := []edgeproto.AutoProvCloudlet{
		edgeproto.AutoProvCloudlet{
			Key: edgeproto.CloudletKey{
				Organization: carrier,
				Name:         "1,1",
			},
			Loc: dme.Loc{
				Latitude:  1,
				Longitude: 1,
			},
		},
		edgeproto.AutoProvCloudlet{
			Key: edgeproto.CloudletKey{
				Organization: carrier,
				Name:         "4,4",
			},
			Loc: dme.Loc{
				Latitude:  4,
				Longitude: 4,
			},
		},
		edgeproto.AutoProvCloudlet{
			Key: edgeproto.CloudletKey{
				Organization: carrier,
				Name:         "8,8",
			},
			Loc: dme.Loc{
				Latitude:  8,
				Longitude: 8,
			},
		},
	}

	// reservable ClusterInsts
	clusterInsts := []edgeproto.ClusterInst{}
	for _, cl := range apCloudlets {
		ci := edgeproto.ClusterInst{
			Key: edgeproto.ClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: "clust",
				},
				CloudletKey:  cl.Key,
				Organization: cloudcommon.OrganizationMobiledgeX,
			},
			Reservable: true,
		}
		clusterInsts = append(clusterInsts, ci)
	}

	locs := []dme.Loc{}
	for _, cl := range apCloudlets {
		locs = append(locs, cl.Loc)
	}

	policies := []edgeproto.AutoProvPolicy{
		edgeproto.AutoProvPolicy{
			Key: edgeproto.PolicyKey{
				Name:         "policy01",
				Organization: org,
			},
			DeployClientCount:   2,
			DeployIntervalCount: 2,
			Cloudlets: []*edgeproto.AutoProvCloudlet{
				&apCloudlets[0],
				&apCloudlets[1],
			},
		},
		edgeproto.AutoProvPolicy{
			Key: edgeproto.PolicyKey{
				Name:         "policy12",
				Organization: org,
			},
			DeployClientCount:   2,
			DeployIntervalCount: 2,
			Cloudlets: []*edgeproto.AutoProvCloudlet{
				&apCloudlets[1],
				&apCloudlets[2],
			},
		},
		edgeproto.AutoProvPolicy{
			Key: edgeproto.PolicyKey{
				Name:         "policy012",
				Organization: org,
			},
			DeployClientCount:   2,
			DeployIntervalCount: 2,
			Cloudlets: []*edgeproto.AutoProvCloudlet{
				&apCloudlets[0],
				&apCloudlets[1],
				&apCloudlets[2],
			},
		},
	}
	// immediate policies (interval count is 1)
	immPolicies := []edgeproto.AutoProvPolicy{
		edgeproto.AutoProvPolicy{
			Key: edgeproto.PolicyKey{
				Name:         "immpolicy01",
				Organization: org,
			},
			DeployClientCount:   2,
			DeployIntervalCount: 1,
			Cloudlets: []*edgeproto.AutoProvCloudlet{
				&apCloudlets[0],
				&apCloudlets[1],
			},
		},
		edgeproto.AutoProvPolicy{
			Key: edgeproto.PolicyKey{
				Name:         "immpolicy12",
				Organization: org,
			},
			DeployClientCount:   2,
			DeployIntervalCount: 1,
			Cloudlets: []*edgeproto.AutoProvCloudlet{
				&apCloudlets[1],
				&apCloudlets[2],
			},
		},
		edgeproto.AutoProvPolicy{
			Key: edgeproto.PolicyKey{
				Name:         "immpolicy012",
				Organization: org,
			},
			DeployClientCount:   2,
			DeployIntervalCount: 1,
			Cloudlets: []*edgeproto.AutoProvCloudlet{
				&apCloudlets[0],
				&apCloudlets[1],
				&apCloudlets[2],
			},
		},
	}

	emptyTest := apStatsTestData{
		app:                app,
		carrier:            carrier,
		expectedCounts:     make(map[edgeproto.CloudletKey]uint64),
		expectedSendCounts: make(map[edgeproto.CloudletKey]uint64),
	}

	// no policies means no stats
	test := emptyTest
	test.clusterInsts = clusterInsts
	test.locs = locs
	test.run(t, ctx)

	// no reservable cluster insts means no stats
	test = emptyTest
	test.policies = append(policies, immPolicies...)
	test.locs = locs
	test.run(t, ctx)

	// single policy01
	test = emptyTest
	test.clusterInsts = clusterInsts
	test.policies = policies[0:1]
	test.locs = locs
	test.expectedCounts = map[edgeproto.CloudletKey]uint64{
		apCloudlets[0].Key: 1, // req loc 1,1
		apCloudlets[1].Key: 2, // req loc 4,4 and 8,8
	}
	test.run(t, ctx)

	// single policy12
	test = emptyTest
	test.clusterInsts = clusterInsts
	test.policies = policies[1:2]
	test.locs = append(locs, locs...) // double requests
	test.expectedCounts = map[edgeproto.CloudletKey]uint64{
		apCloudlets[1].Key: 4, // req loc 1,1 and 4,4,
		apCloudlets[2].Key: 2, // req loc 8,8
	}
	test.run(t, ctx)

	// single policy012
	test = emptyTest
	test.clusterInsts = clusterInsts
	test.policies = policies[2:3]
	test.locs = locs
	test.expectedCounts = map[edgeproto.CloudletKey]uint64{
		apCloudlets[0].Key: 1, // req loc 1,1
		apCloudlets[1].Key: 1, // req loc 4,4
		apCloudlets[2].Key: 1, // req loc 8,8
	}
	test.run(t, ctx)

	// all policies (except immediate)
	test = emptyTest
	test.clusterInsts = clusterInsts
	test.policies = policies
	test.locs = locs
	test.expectedCounts = map[edgeproto.CloudletKey]uint64{
		apCloudlets[0].Key: 1, // req loc 1,1
		apCloudlets[1].Key: 1, // req loc 4,4
		apCloudlets[2].Key: 1, // req loc 8,8
	}

	// all policies (immediate policies should trigger sends)
	test = emptyTest
	test.clusterInsts = clusterInsts
	test.policies = append(policies, immPolicies...)
	test.locs = append(locs, locs...) // double requests
	test.expectedCounts = map[edgeproto.CloudletKey]uint64{
		apCloudlets[0].Key: 2,
		apCloudlets[1].Key: 2,
		apCloudlets[2].Key: 2,
	}
	test.expectedSendCounts = map[edgeproto.CloudletKey]uint64{
		apCloudlets[0].Key: 2,
		apCloudlets[1].Key: 2,
		apCloudlets[2].Key: 2,
	}
	test.run(t, ctx)
}

type apStatsTestData struct {
	app                edgeproto.App
	carrier            string
	clusterInsts       []edgeproto.ClusterInst
	policies           []edgeproto.AutoProvPolicy
	locs               []dme.Loc
	expectedCounts     map[edgeproto.CloudletKey]uint64
	expectedSendCounts map[edgeproto.CloudletKey]uint64
}

func (s *apStatsTestData) run(t *testing.T, ctx context.Context) {
	// reset all data
	actualSendCounts := make(map[edgeproto.CloudletKey]uint64)

	eehandler := &EmptyEdgeEventsHandler{}
	SetupMatchEngine(eehandler)
	InitAutoProvStats(500, 0, 1, &edgeproto.NodeKey{}, func(ctx context.Context, counts *edgeproto.AutoProvCounts) bool {
		require.Equal(t, 1, len(counts.Counts))
		apCount := counts.Counts[0]
		actualSendCounts[apCount.CloudletKey] = apCount.Count
		require.True(t, apCount.ProcessNow)
		return true
	})
	apHandler := AutoProvPolicyHandler{}

	// add policies
	for _, policy := range s.policies {
		apHandler.Update(ctx, &policy, 0)
	}
	// add ClusterInsts
	for _, ci := range s.clusterInsts {
		DmeAppTbl.FreeReservableClusterInsts.Update(ctx, &ci, 0)
	}
	// set policies on App
	app := s.app
	for _, p := range s.policies {
		app.AutoProvPolicies = append(app.AutoProvPolicies, p.Key.Name)
	}
	// add App
	AddApp(ctx, &app)

	// do "find cloudlet" calls
	for _, loc := range s.locs {
		findBestForCarrier(ctx, s.carrier, &app.Key, &loc, 1)
	}

	// get actual stats
	actualCounts := make(map[edgeproto.CloudletKey]uint64)
	for ii, _ := range autoProvStats.shards {
		for key, counts := range autoProvStats.shards[ii].appCloudletCounts {
			actualCounts[key.CloudletKey] = counts.count
		}
	}

	require.Equal(t, s.expectedCounts, actualCounts)
	require.Equal(t, s.expectedSendCounts, actualSendCounts)
}
