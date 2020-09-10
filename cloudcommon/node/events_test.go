package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelEvents)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// elasticsearch docker takes a while to start up (~20s),
	// so make sure to include all unit-testing against it here.
	esProc := process.ElasticSearch{}
	esProc.Common.Name = "elasticsearch-unit-test"
	err := esProc.StartLocal("")
	require.Nil(t, err)
	defer esProc.StopLocal()

	// events rely on nodeMgr
	nodeMgr := NodeMgr{}
	err = nodeMgr.Init(ctx, NodeTypeController, "", WithRegion("unit-test"),
		WithESUrls("http://localhost:9200"))
	require.Nil(t, err)

	starttime := time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
	ts := starttime

	keyTags := map[string]string{
		"app":         "myapp",
		"apporg":      "devOrg",
		"appver":      "1.0",
		"cloudlet":    "cloudlet1",
		"cloudletorg": "operOrg",
		"cluster":     "testclust",
		"clusterorg":  "MobiledgeX",
	}
	keyTags2 := map[string]string{
		"cloudlet":    "cloudlet1",
		"cloudletorg": "operOrg",
	}
	// create events
	org := "devOrg"
	org2 := "operOrg"
	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "test start", "", "event", nil, nil, ts)

	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "cloudlet online", org2, "event", keyTags2, nil, ts)

	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "create AppInst", org, "event", keyTags, nil, ts)

	ts = ts.Add(time.Minute)
	keyTags["cloudlet"] = "cloudlet2"
	nodeMgr.EventAtTime(ctx, "create AppInst", org, "event", keyTags, fmt.Errorf("failed, unknown failure"), ts, "the reason", "AutoProv")

	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "delete AppInst", org, "event", keyTags, fmt.Errorf("failed, random failure"), ts, "the reason", "just because")

	// for some reason ES is not ready immediately for searching
	time.Sleep(time.Second)

	//
	// ---------------------------------------------------
	// Tests for term aggregations
	// ---------------------------------------------------
	//

	// check terms aggregations over all events
	search := EventSearch{
		TimeRange: util.TimeRange{
			StartTime: starttime,
			EndTime:   time.Now(),
		},
		Limit: 100,
	}
	terms, err := nodeMgr.EventTerms(ctx, &search)
	require.Nil(t, err)
	expectedTerms := EventTerms{
		Names: []string{
			"cloudlet online",
			"controller start",
			"create AppInst",
			"delete AppInst",
			"test start",
		},
		Orgs:    []string{"", "devOrg", "operOrg"},
		Types:   []string{"event"},
		Regions: []string{"unit-test"},
		TagKeys: []string{
			"app",
			"apporg",
			"appver",
			"cloudlet",
			"cloudletorg",
			"cluster",
			"clusterorg",
			"hostname",
			"lineno",
			"node",
			"noderegion",
			"nodetype",
			"spanid",
			"the reason",
			"traceid",
		},
	}
	require.Equal(t, expectedTerms, *terms)

	// check terms aggregations filtered by allowed org
	es := search
	es.AllowedOrgs = []string{"devOrg"}
	terms, err = nodeMgr.EventTerms(ctx, &es)
	require.Nil(t, err)
	expectedTerms = EventTerms{
		Names: []string{
			"create AppInst",
			"delete AppInst",
		},
		Orgs:    []string{"devOrg"},
		Types:   []string{"event"},
		Regions: []string{"unit-test"},
		TagKeys: []string{
			"app",
			"apporg",
			"appver",
			"cloudlet",
			"cloudletorg",
			"cluster",
			"clusterorg",
			"hostname",
			"lineno",
			"spanid",
			"the reason",
			"traceid",
		},
	}
	require.Equal(t, expectedTerms, *terms)

	//
	// ---------------------------------------------------
	// Tests for filter searches
	// ---------------------------------------------------
	//

	// limit time range to just our test events.
	// this avoids the startup event added by nodeMgr.Init().
	search = EventSearch{
		TimeRange: util.TimeRange{
			StartTime: starttime,
			EndTime:   starttime.Add(time.Hour),
		},
		Limit: 100,
	}

	// find all events
	events, err := nodeMgr.ShowEvents(ctx, &search)
	require.Nil(t, err)
	require.Equal(t, 5, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "cloudlet online", events[3].Name)
	require.Equal(t, "test start", events[4].Name)

	// find all create AppInst events
	es = search
	es.Match.Names = []string{"create AppInst"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)

	// find by multiple names
	es = search
	es.Match.Names = []string{"create AppInst", "delete AppInst"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)

	// find all create events
	es = search
	es.Match.Names = []string{"create*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)

	// search text by words - name is a keyword so must be exact or wildcard
	es = search
	es.Match.Names = []string{"create"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))

	// support wildcard matching
	es = search
	es.Match.Names = []string{"*App*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)

	// support wildcard matching
	es = search
	es.Match.Names = []string{"create App*", "delete App*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)

	// search for all that failed, regardless of error
	es = search
	es.Match.Failed = true
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)

	// search for particular error message
	es = search
	es.Match.Error = "random failure"
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	// note that order of words doesn't matter, nor does capitalization
	es = search
	es.Match.Error = "Failure Random"
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)

	// search by org
	// for security, org is a keyword so requires an exact string match
	es = search
	es.Match.Orgs = []string{"devOrg"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	// search by org does not allow partial match
	es = search
	es.Match.Orgs = []string{"dev"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	// search by org does not allow case insensitivity
	es = search
	es.Match.Orgs = []string{"devorg"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	// search by org supports wildcard, but should probably be filtered
	// by MC for RBAC.
	es = search
	es.Match.Orgs = []string{"dev*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)

	// search by tag
	es = search
	es.Match.Tags = map[string]string{
		"app": "myapp",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	// search by tag key must be exact match
	es = search
	es.Match.Tags = map[string]string{
		"reason": "AutoProv",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	// search by tag key must be exact match
	es = search
	es.Match.Tags = map[string]string{
		"the reason": "AutoProv",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	// search by multiple tags must include all
	es = search
	es.Match.Tags = map[string]string{
		"app":      "myapp",
		"cloudlet": "cloudlet1",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	// search by multiple tags must include all
	es = search
	es.Match.Tags = map[string]string{
		"app":      "myapp",
		"cloudlet": "cloudlet2",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	// search by tag value can be wildcard
	es = search
	es.Match.Tags = map[string]string{
		"app":      "myapp",
		"cloudlet": "cloudlet*",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	// verify lineno tag is set correctly
	es = search
	es.Match.Tags = map[string]string{
		"lineno": "*events_test.go*",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 5, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "cloudlet online", events[3].Name)
	require.Equal(t, "test start", events[4].Name)

	// verify allowedOrgs enforcement
	es = search
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	es = search
	es.Match.Orgs = []string{"devOrg"}
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))

	// find all events for multiple allowed orgs
	es = search
	es.AllowedOrgs = []string{org, org2}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 4, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "cloudlet online", events[3].Name)

	// search by time range
	es = search
	es.StartTime = starttime
	es.EndTime = starttime.Add(2*time.Minute + 200*time.Millisecond)
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "cloudlet online", events[0].Name)
	require.Equal(t, "test start", events[1].Name)

	//
	// ---------------------------------------------------
	// Tests for relevance searches
	// ---------------------------------------------------
	//

	// search looking for error message
	es = search
	es.Match.Orgs = []string{"devOrg", "operOrg"}
	es.Match.Error = "failed"
	es.Match.Tags = map[string]string{
		"app":        "myapp",
		"the reason": "because",
	}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 5, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "", events[2].Error) // should be empty
	require.Equal(t, "cloudlet online", events[3].Name)
	require.Equal(t, "test start", events[4].Name)

	// search looking for failed autoprov
	es = search
	es.Match.Orgs = []string{"devOrg", "operOrg"}
	es.Match.Failed = true
	es.Match.Tags = map[string]string{
		"app":        "myapp",
		"the reason": "autoprov",
	}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 5, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "", events[2].Error) // should be empty
	require.Equal(t, "cloudlet online", events[3].Name)
	require.Equal(t, "test start", events[4].Name)

	// search for autoprov creates
	es = search
	es.Match.Orgs = []string{"devOrg", "operOrg"}
	es.Match.Names = []string{"*create*"}
	es.Match.Tags = map[string]string{
		"app":      "myapp",
		"cloudlet": "cloudlet1",
	}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 5, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	require.Equal(t, "", events[0].Error) // should be empty
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "delete AppInst", events[2].Name)
	require.Equal(t, "cloudlet online", events[3].Name)
	require.Equal(t, "test start", events[4].Name)

	// verify allowedOrgs enforcement
	es = search
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	es = search
	es.Match.Orgs = []string{"devOrg"}
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))

	//
	// ---------------------------------------------------
	// Test not matching searches
	// ---------------------------------------------------
	//

	// not names
	es = search
	es.NotMatch.Names = []string{"create AppInst"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "cloudlet online", events[1].Name)
	require.Equal(t, "test start", events[2].Name)

	// tags plus not failed
	es = search
	es.Match.Tags = map[string]string{
		"app": "myapp",
	}
	es.NotMatch.Failed = true
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	require.Equal(t, "create AppInst", events[0].Name)

	// not tags
	es = search
	es.NotMatch.Tags = map[string]string{
		"cloudlet": "cloudlet2",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	require.Equal(t, "cloudlet online", events[1].Name)
	require.Equal(t, "test start", events[2].Name)

	es = search
	es.NotMatch.Names = []string{"create App*", "delete App*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "cloudlet online", events[0].Name)
	require.Equal(t, "test start", events[1].Name)
}
