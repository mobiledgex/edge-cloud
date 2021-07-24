package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestEvents(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelEvents | log.DebugLevelInfo)

	// elasticsearch docker takes a while to start up (~20s),
	// so make sure to include all unit-testing against it here.
	esProc := process.ElasticSearch{}
	esProc.Common.Name = "elasticsearch-unit-test"
	err := esProc.StartLocal("")
	require.Nil(t, err)
	defer esProc.StopLocal()

	// start Jaeger to test searching spans in elasticsearch
	jaegerProc := process.Jaeger{}
	jaegerProc.Common.Name = "jaeger-unit-test"
	jaegerProc.DockerEnvVars = make(map[string]string)
	jaegerProc.DockerEnvVars["ES_SERVER_URLS"] = "http://elasticsearch:9200"
	jaegerProc.DockerEnvVars["SPAN_STORAGE_TYPE"] = "elasticsearch"
	jaegerProc.Links = []string{"elasticsearch-unit-test:elasticsearch"}
	err = jaegerProc.StartLocalNoTraefik("")
	require.Nil(t, err)
	defer jaegerProc.StopLocal()

	// set true otherwise logger will not log spans for unit-tests
	log.JaegerUnitTest = true

	// events rely on nodeMgr
	nodeMgr := NodeMgr{}
	ctx, _, err := nodeMgr.Init(NodeTypeController, "", WithRegion("unit-test"),
		WithESUrls("http://localhost:9200"))
	require.Nil(t, err)
	defer nodeMgr.Finish()
	nodeMgr.unitTestMode = true

	starttime := time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
	ts := starttime

	org := "devOrg"
	operOrg := "operOrg"
	keyTags := map[string]string{
		edgeproto.AppKeyTagName:                 "myapp",
		edgeproto.AppKeyTagOrganization:         org,
		edgeproto.AppKeyTagVersion:              "1.0",
		edgeproto.CloudletKeyTagName:            "cloudlet1",
		edgeproto.CloudletKeyTagOrganization:    operOrg,
		edgeproto.ClusterKeyTagName:             "testclust",
		edgeproto.ClusterInstKeyTagOrganization: "MobiledgeX",
	}
	keyTags2 := map[string]string{
		edgeproto.CloudletKeyTagName:         "cloudlet1",
		edgeproto.CloudletKeyTagOrganization: operOrg,
	}
	// create events
	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "test start", NoOrg, "event", nil, nil, ts)

	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "cloudlet online", operOrg, "event", keyTags2, nil, ts)

	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "create AppInst", org, "event", keyTags, nil, ts)

	ts = ts.Add(time.Minute)
	keyTags[edgeproto.CloudletKeyTagName] = "cloudlet2"
	nodeMgr.EventAtTime(ctx, "create AppInst", org, "event", keyTags, fmt.Errorf("failed, unknown failure"), ts, "the reason", "AutoProv")

	ts = ts.Add(time.Minute)
	nodeMgr.EventAtTime(ctx, "delete AppInst", org, "event", keyTags, fmt.Errorf("failed, random failure"), ts, "the reason", "just because")

	// add cloudlet to cloudlet pool
	pool := edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Organization: operOrg,
			Name:         "pool1",
		},
		Cloudlets: []string{"cloudlet1"},
	}
	nodeMgr.CloudletPoolLookup.GetCloudletPoolCache(NoRegion).Update(ctx, &pool, 0)
	cloudletKey := edgeproto.CloudletKey{
		Organization: operOrg,
		Name:         "cloudlet1",
	}
	cpc, ok := nodeMgr.CloudletPoolLookup.(*CloudletPoolCache)
	require.True(t, ok)
	require.True(t, cpc.PoolsByCloudlet.HasRef(cloudletKey))

	// event with two allowed orgs, developer and operator due to CloudletPool
	ts = ts.Add(time.Minute)
	keyTags[edgeproto.CloudletKeyTagName] = "cloudlet1"
	nodeMgr.EventAtTime(ctx, "update AppInst", org, "event", keyTags, nil, ts)

	//
	// ---------------------------------------------------
	// Span logs test data
	// ---------------------------------------------------
	//
	span := log.StartSpan(log.DebugLevelInfo, "span1")
	sctx := log.ContextWithSpan(context.Background(), span)
	log.SpanLog(sctx, log.DebugLevelInfo, "span1-msg1", "key1", "somevalue")
	log.SpanLog(sctx, log.DebugLevelInfo, "msg2")
	span.Finish()

	span = log.StartSpan(log.DebugLevelInfo, "span2")
	sctx = log.ContextWithSpan(context.Background(), span)
	log.SpanLog(sctx, log.DebugLevelInfo, "span2-msg1", "key2", "foooobar")
	log.SpanLog(sctx, log.DebugLevelInfo, "msg2")
	span.Finish()

	span = log.StartSpan(log.DebugLevelInfo, "span3")
	sctx = log.ContextWithSpan(context.Background(), span)
	log.SpanLog(sctx, log.DebugLevelInfo, "msg3")
	log.SpanLog(sctx, log.DebugLevelInfo, "msg3", "key2", "foooobar")
	span.Finish()

	span = log.StartSpan(log.DebugLevelInfo, "span4")
	sctx = log.ContextWithSpan(context.Background(), span)
	log.SpanLog(sctx, log.DebugLevelInfo, "span4-msg1", "anotherkey", "anothervalue")
	log.SpanLog(sctx, log.DebugLevelInfo, "msg3")
	span.Finish()

	// wait for queued events to be written to ES
	waitEvents(t, &nodeMgr, 7)
	// wait for queued spans to be written
	waitSpans(t, 5) // one extra because nodeMgr creates one

	endtime := time.Now()

	// for some reason ES is not ready immediately for searching
	time.Sleep(3 * time.Second)

	//
	// ---------------------------------------------------
	// Tests for term aggregations
	// ---------------------------------------------------
	//

	aggr := func(name string, count int) AggrVal {
		return AggrVal{
			Key:      name,
			DocCount: count,
		}
	}

	// check terms aggregations over all events
	search := EventSearch{
		TimeRange: edgeproto.TimeRange{
			StartTime: starttime,
			EndTime:   endtime,
		},
		Limit: 100,
	}
	terms, err := nodeMgr.EventTerms(ctx, &search)
	require.Nil(t, err)
	expectedTerms := EventTerms{
		Names: []AggrVal{
			aggr("create AppInst", 2),
			aggr("cloudlet online", 1),
			aggr("controller start", 1),
			aggr("delete AppInst", 1),
			aggr("test start", 1),
			aggr("update AppInst", 1),
		},
		Orgs: []AggrVal{
			aggr(org, 4),
			aggr(NoOrg, 2),
			aggr(operOrg, 2),
		},
		Types:   []AggrVal{aggr("event", 7)},
		Regions: []AggrVal{aggr("unit-test", 7)},
		TagKeys: []AggrVal{
			aggr("hostname", 7),
			aggr("lineno", 7),
			aggr("spanid", 7),
			aggr("traceid", 7),
			aggr(edgeproto.CloudletKeyTagName, 6),
			aggr(edgeproto.CloudletKeyTagOrganization, 6),
			aggr(edgeproto.AppKeyTagName, 4),
			aggr(edgeproto.AppKeyTagOrganization, 4),
			aggr(edgeproto.AppKeyTagVersion, 4),
			aggr(edgeproto.ClusterKeyTagName, 4),
			aggr(edgeproto.ClusterInstKeyTagOrganization, 4),
			aggr("the reason", 2),
			aggr("node", 1),
			aggr("noderegion", 1),
			aggr("nodetype", 1),
		},
	}
	require.Equal(t, expectedTerms, *terms)

	// check terms aggregations filtered by allowed org
	es := search
	es.AllowedOrgs = []string{org}
	terms, err = nodeMgr.EventTerms(ctx, &es)
	require.Nil(t, err)
	expectedTerms = EventTerms{
		Names: []AggrVal{
			aggr("create AppInst", 2),
			aggr("delete AppInst", 1),
			aggr("update AppInst", 1),
		},
		Orgs: []AggrVal{
			aggr(org, 4),
			aggr(operOrg, 1),
		},
		Types:   []AggrVal{aggr("event", 4)},
		Regions: []AggrVal{aggr("unit-test", 4)},
		TagKeys: []AggrVal{
			aggr(edgeproto.AppKeyTagName, 4),
			aggr(edgeproto.AppKeyTagOrganization, 4),
			aggr(edgeproto.AppKeyTagVersion, 4),
			aggr(edgeproto.CloudletKeyTagName, 4),
			aggr(edgeproto.CloudletKeyTagOrganization, 4),
			aggr(edgeproto.ClusterKeyTagName, 4),
			aggr(edgeproto.ClusterInstKeyTagOrganization, 4),
			aggr("hostname", 4),
			aggr("lineno", 4),
			aggr("spanid", 4),
			aggr("traceid", 4),
			aggr("the reason", 2),
		},
	}
	require.Equal(t, expectedTerms, *terms)

	// check terms aggregations filtered by allowed org
	es = search
	es.AllowedOrgs = []string{operOrg}
	terms, err = nodeMgr.EventTerms(ctx, &es)
	require.Nil(t, err)
	expectedTerms = EventTerms{
		Names: []AggrVal{
			aggr("cloudlet online", 1),
			aggr("update AppInst", 1),
		},
		Orgs: []AggrVal{
			aggr(operOrg, 2),
			aggr(org, 1),
		},
		Types:   []AggrVal{aggr("event", 2)},
		Regions: []AggrVal{aggr("unit-test", 2)},
		TagKeys: []AggrVal{
			aggr(edgeproto.CloudletKeyTagName, 2),
			aggr(edgeproto.CloudletKeyTagOrganization, 2),
			aggr("hostname", 2),
			aggr("lineno", 2),
			aggr("spanid", 2),
			aggr("traceid", 2),
			aggr(edgeproto.AppKeyTagName, 1),
			aggr(edgeproto.AppKeyTagOrganization, 1),
			aggr(edgeproto.AppKeyTagVersion, 1),
			aggr(edgeproto.ClusterKeyTagName, 1),
			aggr(edgeproto.ClusterInstKeyTagOrganization, 1),
		},
	}
	require.Equal(t, expectedTerms, *terms)

	//
	// ---------------------------------------------------
	// Tests for span term aggregations
	// ---------------------------------------------------
	//
	spansearch := SpanSearch{
		TimeRange: edgeproto.TimeRange{
			StartTime: starttime,
			EndTime:   endtime,
		},
		Limit: 100,
	}
	sterms, err := nodeMgr.SpanTerms(ctx, &spansearch)
	require.Nil(t, err)
	expectedSpanTerms := &SpanTerms{
		Operations: []AggrVal{
			aggr("init-es-events", 1),
			aggr("span1", 1),
			aggr("span2", 1),
			aggr("span3", 1),
			aggr("span4", 1),
		},
		Services: []AggrVal{
			aggr("node.test", 5),
		},
		Msgs: []AggrVal{
			aggr("msg3", 3),
			aggr("msg2", 2),
			aggr("queued event", 1),
			aggr("span1-msg1", 1),
			aggr("span2-msg1", 1),
			aggr("span4-msg1", 1),
			aggr("write event-log index template", 1),
		},
	}
	// ignore tags and hostnames
	sterms.Tags = nil
	sterms.Hostnames = nil
	require.Equal(t, expectedSpanTerms, sterms)

	//
	// ---------------------------------------------------
	// Tests for filter searches
	// ---------------------------------------------------
	//

	// limit time range to just our test events.
	// this avoids the startup event added by nodeMgr.Init().
	search = EventSearch{
		TimeRange: edgeproto.TimeRange{
			StartTime: starttime,
			EndTime:   starttime.Add(time.Hour),
		},
		Limit: 100,
	}

	// find all events
	events, err := nodeMgr.ShowEvents(ctx, &search)
	require.Nil(t, err)
	require.Equal(t, 6, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	require.Equal(t, "cloudlet online", events[4].Name)
	require.Equal(t, "test start", events[5].Name)

	// find all events (wildcard)
	es = search
	es.Match.Names = []string{"*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 6, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	require.Equal(t, "cloudlet online", events[4].Name)
	require.Equal(t, "test start", events[5].Name)

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
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)

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
	es.Match.Orgs = []string{org}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
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
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)

	out, err := yaml.Marshal(events[0])
	require.Nil(t, err)
	fmt.Printf("%s\n", string(out))

	// search by operator org for CloudletPool-based Cloudlet events
	es = search
	es.Match.Orgs = []string{operOrg}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "cloudlet online", events[1].Name)

	// search by tag
	es = search
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName: "myapp",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
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
		edgeproto.AppKeyTagName:      "myapp",
		edgeproto.CloudletKeyTagName: "cloudlet1",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	// search by multiple tags must include all
	es = search
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName:      "myapp",
		edgeproto.CloudletKeyTagName: "cloudlet2",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	// search by tag value can be wildcard
	es = search
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName:      "myapp",
		edgeproto.CloudletKeyTagName: "cloudlet*",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	// verify lineno tag is set correctly
	es = search
	es.Match.Tags = map[string]string{
		"lineno": "*events_test.go*",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 6, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	require.Equal(t, "cloudlet online", events[4].Name)
	require.Equal(t, "test start", events[5].Name)

	// verify allowedOrgs enforcement
	es = search
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	es = search
	es.Match.Orgs = []string{org}
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))

	// find all events for multiple allowed orgs
	es = search
	es.AllowedOrgs = []string{org, operOrg}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 5, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	require.Equal(t, "cloudlet online", events[4].Name)

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
	es.Match.Orgs = []string{org, operOrg}
	es.Match.Error = "failed"
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName: "myapp",
		"the reason":            "because",
	}
	es.Match.Names = []string{"*create*"}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 6, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "update AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	require.Equal(t, "", events[3].Error) // should be empty
	require.Equal(t, "cloudlet online", events[4].Name)
	require.Equal(t, "test start", events[5].Name)

	// search looking for failed autoprov
	es = search
	es.Match.Orgs = []string{org, operOrg}
	es.Match.Failed = true
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName: "myapp",
		"the reason":            "autoprov",
	}
	es.Match.Names = []string{"*update*"}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 6, len(events))
	require.Equal(t, "create AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "update AppInst", events[2].Name)
	require.Equal(t, "create AppInst", events[3].Name)
	require.Equal(t, "", events[3].Error) // should be empty
	require.Equal(t, "cloudlet online", events[4].Name)
	require.Equal(t, "test start", events[5].Name)

	// search for autoprov creates
	es = search
	es.Match.Orgs = []string{org, operOrg}
	es.Match.Names = []string{"*create*"}
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName:      "myapp",
		edgeproto.CloudletKeyTagName: "cloudlet1",
	}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 6, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "", events[1].Error) // should be empty
	require.Equal(t, "create AppInst", events[2].Name)
	require.Equal(t, "delete AppInst", events[3].Name)
	require.Equal(t, "cloudlet online", events[4].Name)
	require.Equal(t, "test start", events[5].Name)

	// verify allowedOrgs enforcement
	es = search
	es.AllowedOrgs = []string{"otherOrg"}
	events, err = nodeMgr.FindEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 0, len(events))
	es = search
	es.Match.Orgs = []string{org}
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
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "delete AppInst", events[1].Name)
	require.Equal(t, "cloudlet online", events[2].Name)
	require.Equal(t, "test start", events[3].Name)

	// tags plus not failed
	es = search
	es.Match.Tags = map[string]string{
		edgeproto.AppKeyTagName: "myapp",
	}
	es.NotMatch.Failed = true
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 2, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)

	// not tags
	es = search
	es.NotMatch.Tags = map[string]string{
		edgeproto.CloudletKeyTagName: "cloudlet2",
	}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 4, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "create AppInst", events[1].Name)
	require.Equal(t, "cloudlet online", events[2].Name)
	require.Equal(t, "test start", events[3].Name)

	es = search
	es.NotMatch.Names = []string{"create App*", "delete App*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 3, len(events))
	require.Equal(t, "update AppInst", events[0].Name)
	require.Equal(t, "cloudlet online", events[1].Name)
	require.Equal(t, "test start", events[2].Name)

	//
	// ---------------------------------------------------
	// Test failures
	// ---------------------------------------------------
	//

	// test error check for -, should be ok for keywords
	es = search
	es.Match.Names = []string{"create-App*", "delete App*"}
	events, err = nodeMgr.ShowEvents(ctx, &es)
	require.Nil(t, err)
	require.Equal(t, 1, len(events))
	require.Equal(t, "delete AppInst", events[0].Name)
	// test error check for - in text
	es = search
	es.Match.Tags = map[string]string{
		"somekey": "bad-wildcard*",
	}
	_, err = nodeMgr.ShowEvents(ctx, &es)
	require.NotNil(t, err)
}

func waitEvents(t *testing.T, nm *NodeMgr, num uint64) {
	for ii := 0; ii < 20; ii++ {
		fmt.Printf("waitEvents %d: %d\n", ii, nm.ESWroteEvents)
		if nm.ESWroteEvents == num {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
	require.Equal(t, num, nm.ESWroteEvents)
}

func waitSpans(t *testing.T, num int64) {
	name := "jaeger.tracer.reporter_spans|result=ok"
	counters := map[string]int64{}
	for ii := 0; ii < 20; ii++ {
		counters, _ = log.ReporterMetrics.Snapshot()
		fmt.Printf("waitSpans %d: %v\n", ii, counters)
		if counters[name] == num {
			break
		}
		time.Sleep(time.Second)
	}
	require.Equal(t, num, counters[name])
}
