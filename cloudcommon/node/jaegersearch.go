package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/jaegertracing/jaeger/plugin/storage/es/spanstore/dbmodel"
	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

type SpanMatch struct {
	TraceIDs     []string          `json:"traceids"`
	SpanIDs      []string          `json:"spanids"`
	Services     []string          `json:"services"`
	Operations   []string          `json:"operations"`
	Hostnames    []string          `json:"hostnames"`
	TagValues    []string          `json:"tagvalues"`
	TagKeyValues map[string]string `json:"tagkeyvalues"`
	LogMsgs      []string          `json:"logmsgs"`
	LogValues    []string          `json:"logvalues"`
	LogKeyValues map[string]string `json:"logkeyvalues"`
}

type SpanSearch struct {
	Match               SpanMatch `json:"match,omitempty"`
	NotMatch            SpanMatch `json:"notmatch,omitempty"`
	edgeproto.TimeRange `json:",inline"`
	From                int  `json:"from,omitempty"`  // start document offset
	Limit               int  `json:"limit,omitempty"` // number of documents to return
	SearchByRelevance   bool `json:"searchbyrelevance"`
}

type SpanOutCondensed struct {
	TraceID   string                 `json:"traceid,omitempty"`
	SpanID    string                 `json:"spanid,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	StartTime time.Time              `json:"starttime,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Hostname  string                 `json:"hostname,omitempty"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	Logs      []SpanLogOut           `json:"logs,omitempty"`
}

type SpanLogOut struct {
	Msg       string                 `json:"msg"`
	Timestamp time.Time              `json:"timestamp"`
	Lineno    string                 `json:"lineno"`
	KeyValues map[string]interface{} `json:"keyvalues"`
}

type SpanTerms struct {
	Operations []AggrVal `json:"operations"`
	Services   []AggrVal `json:"services"`
	Hostnames  []AggrVal `json:"hostnames"`
	Msgs       []AggrVal `json:"msgs"`
	Tags       []AggrVal `json:"tags"`
}

func (s *NodeMgr) ShowSpans(ctx context.Context, search *SpanSearch) ([]dbmodel.Span, error) {
	return s.searchSpans(ctx, search)
}

func (s *NodeMgr) ShowSpansCondensed(ctx context.Context, search *SpanSearch) ([]SpanOutCondensed, error) {
	spans, err := s.searchSpans(ctx, search)
	if err != nil {
		return nil, err
	}
	// The dbmodel Span used to write to ElasticSearch isn't very easy to for
	// humans to read. Condense and convert it to something more compact
	// and easier for a human to read.
	out := []SpanOutCondensed{}
	for _, span := range spans {
		sp := SpanOutCondensed{}
		sp.TraceID = string(span.TraceID)
		sp.SpanID = string(span.SpanID)
		sp.Service = span.Process.ServiceName
		sp.Operation = span.OperationName
		sp.StartTime = util.TimeFromEpochMicros(int64(span.StartTime))
		sp.Duration = time.Duration(span.Duration * 1e3)
		sp.Tags = make(map[string]interface{})
		for _, kv := range span.Process.Tags {
			if kv.Key == "hostname" {
				val, ok := kv.Value.(string)
				if ok {
					sp.Hostname = val
				}
				break
			}
		}
		for _, kv := range span.Tags {
			if log.IgnoreSpanTag(kv.Key) {
				continue
			}
			sp.Tags[kv.Key] = kv.Value
		}
		for _, log := range span.Logs {
			// SpanLog always sets the first field as the "msg",
			// and the second field as the "lineno".
			// See spanlog.go:SpanLog().
			if len(log.Fields) < 2 {
				continue
			}
			lo := SpanLogOut{}
			lo.Timestamp = util.TimeFromEpochMicros(int64(log.Timestamp))
			if val, ok := log.Fields[0].Value.(string); ok {
				lo.Msg = val
			}
			if val, ok := log.Fields[1].Value.(string); ok {
				lo.Lineno = val
			}
			lo.KeyValues = make(map[string]interface{})
			for ii := 2; ii < len(log.Fields); ii++ {
				lo.KeyValues[log.Fields[ii].Key] = log.Fields[ii].Value
			}
			sp.Logs = append(sp.Logs, lo)
		}
		out = append(out, sp)
	}
	return out, nil
}

func (s *NodeMgr) searchSpans(ctx context.Context, search *SpanSearch) ([]dbmodel.Span, error) {
	searchType := searchTypeFilter
	if search.SearchByRelevance {
		searchType = searchTypeRelevance
	}

	query, err := s.getSpanQuery(ctx, searchType, search)
	if err != nil {
		return nil, err
	}
	dat, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	req := esapi.SearchRequest{
		Index: []string{"jaeger-span-*"},
		Body:  strings.NewReader(string(dat)),
	}
	if s.unitTestMode && searchType == searchTypeRelevance {
		// For relevance queries, multiple shards cause non-deterministic
		// ordering, so force ordering for results between shards.
		// This should not be used for production.
		req.SearchType = "dfs_query_then_fetch"
	}
	res, err := req.Do(ctx, s.ESClient)
	if err == nil && res.StatusCode/100 != http.StatusOK/100 {
		defer res.Body.Close()
		err = fmt.Errorf("%v", res)
	}
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.SpanLog(ctx, log.DebugLevelEvents, "span search response", "res", res)

	resp := SearchResp{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	if resp.Hits.Total.Value == 0 || resp.Hits.Hits == nil {
		return []dbmodel.Span{}, nil
	}
	out := []dbmodel.Span{}
	for _, hit := range resp.Hits.Hits {
		span := dbmodel.Span{}
		err = json.Unmarshal(*hit.Source, &span)
		if err != nil {
			return nil, err
		}
		if s.unitTestMode {
			log.SpanLog(ctx, log.DebugLevelInfo, "searchSpans result", "span", span.OperationName, "score", hit.Score)
		}
		out = append(out, span)
	}
	return out, nil
}

func (s *NodeMgr) getSpanQuery(ctx context.Context, searchType string, search *SpanSearch) (map[string]interface{}, error) {
	if searchType != searchTypeFilter && searchType != searchTypeRelevance {
		return nil, fmt.Errorf("invalid event search type %s", searchType)
	}
	// get match criteria
	smaps, err := s.getSpanMatchQueries(ctx, searchType, &search.Match)
	if err != nil {
		return nil, err
	}
	// get not match criteria
	mustnot, err := s.getSpanMatchQueries(ctx, searchType, &search.NotMatch)
	if err != nil {
		return nil, err
	}

	filter := []map[string]interface{}{}
	// time range is always specified
	// Resolve{} is idempotent so may already have been done.
	err = search.TimeRange.Resolve(DefaultTimeDuration)
	if err != nil {
		return nil, err
	}
	st := util.GetEpochMillis(search.StartTime.UTC())
	et := util.GetEpochMillis(search.EndTime.UTC())
	timerange := map[string]interface{}{
		"range": map[string]interface{}{
			"startTimeMillis": map[string]interface{}{
				"gte": st,
				"lt":  et,
			},
		},
	}
	filter = append(filter, timerange)

	return s.getQueryCommon(ctx, searchType, search.TimeRange, search.From, search.Limit, "startTimeMillis", smaps, mustnot, filter)
}

func (s *NodeMgr) getSpanMatchQueries(ctx context.Context, searchType string, match *SpanMatch) ([]map[string]interface{}, error) {
	nestedQuery := "filter"
	if searchType == searchTypeRelevance {
		nestedQuery = "should"
	}
	// Search parameters (search maps)
	smaps := []map[string]interface{}{}
	if len(match.TraceIDs) > 0 {
		smaps = append(smaps, esSearchKeywords("traceID", match.TraceIDs...))
	}
	if len(match.SpanIDs) > 0 {
		smaps = append(smaps, esSearchKeywords("spanID", match.SpanIDs...))
	}
	if len(match.Services) > 0 {
		smaps = append(smaps, esSearchKeywords("process.serviceName", match.Services...))
	}
	if len(match.Operations) > 0 {
		smaps = append(smaps, esSearchKeywords("operationName", match.Operations...))
	}
	if len(match.Hostnames) > 0 {
		for _, hostname := range match.Hostnames {
			qs := []map[string]interface{}{}
			qs = append(qs,
				esSearchKeywords("process.tags.key", "hostname"),
				esSearchKeywords("process.tags.value", hostname),
			)
			nested := map[string]interface{}{
				"nested": map[string]interface{}{
					"path":            "process.tags",
					"ignore_unmapped": "true",
					"query": map[string]interface{}{
						"bool": map[string]interface{}{
							nestedQuery: qs,
						},
					},
				},
			}
			smaps = append(smaps, nested)
		}
	}
	if len(match.TagValues) > 0 {
		for _, v := range match.TagValues {
			smaps = append(smaps, getNestedSpanTagQuery("", v, nestedQuery))
		}
	}
	if len(match.TagKeyValues) > 0 {
		for k, v := range match.TagKeyValues {
			smaps = append(smaps, getNestedSpanTagQuery(k, v, nestedQuery))
		}
	}
	if len(match.LogMsgs) > 0 {
		for _, v := range match.LogMsgs {
			smaps = append(smaps, getNestedSpanLogQuery("msg", v, nestedQuery))
		}
	}
	if len(match.LogValues) > 0 {
		for _, v := range match.LogValues {
			smaps = append(smaps, getNestedSpanLogQuery("", v, nestedQuery))
		}
	}
	if len(match.LogKeyValues) > 0 {
		for k, v := range match.LogKeyValues {
			smaps = append(smaps, getNestedSpanLogQuery(k, v, nestedQuery))
		}
	}
	return smaps, nil
}

func getNestedSpanTagQuery(key, val, nestedQuery string) map[string]interface{} {
	qs := []map[string]interface{}{}
	if key != "" {
		qs = append(qs, esSearchKeywords("tags.key", key))
	}
	qs = append(qs, esSearchKeywords("tags.value", val))

	// note that this is double nested.
	nested := map[string]interface{}{
		"nested": map[string]interface{}{
			"path":            "tags",
			"ignore_unmapped": "true",
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					nestedQuery: qs,
				},
			},
		},
	}
	return nested
}

func getNestedSpanLogQuery(key, val, nestedQuery string) map[string]interface{} {
	qs := []map[string]interface{}{}
	if key != "" {
		qs = append(qs, esSearchKeywords("logs.fields.key", key))
	}
	qs = append(qs, esSearchKeywords("logs.fields.value", val))

	// note that this is double nested.
	nested := map[string]interface{}{
		"nested": map[string]interface{}{
			"path":            "logs",
			"ignore_unmapped": "true",
			"query": map[string]interface{}{
				"nested": map[string]interface{}{
					"path":            "logs.fields",
					"ignore_unmapped": "true",
					"query": map[string]interface{}{
						"bool": map[string]interface{}{
							nestedQuery: qs,
						},
					},
				},
			},
		},
	}
	return nested
}

func (s *NodeMgr) SpanTerms(ctx context.Context, search *SpanSearch) (*SpanTerms, error) {
	query, err := s.getSpanQuery(ctx, searchTypeFilter, search)
	if err != nil {
		return nil, err
	}
	if search.Limit == 0 {
		search.Limit = 100
	}
	// Set size to 0 because we don't need the overhead of returned results,
	// because we're going to ignore them. This does however disable pagination
	// results for event terms, but it's unlikely pagination will be needed
	// in this context.
	query["size"] = 0
	query["aggs"] = map[string]interface{}{
		"operations": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "operationName",
				"size":  search.Limit,
			},
		},
		"services": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "process.serviceName",
				"size":  search.Limit,
			},
		},
		"hostnames": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "process.tags",
			},
			"aggs": map[string]interface{}{
				"hostnames": map[string]interface{}{
					"filter": map[string]interface{}{
						"term": map[string]interface{}{
							"process.tags.key": "hostname",
						},
					},
					"aggs": map[string]interface{}{
						"tagkeys": map[string]interface{}{
							"terms": map[string]interface{}{
								"field": "process.tags.value",
								"size":  search.Limit,
							},
						},
					},
				},
			},
		},
		"tags": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "tags",
			},
			"aggs": map[string]interface{}{
				"tags": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "tags.key",
						"size":  search.Limit,
					},
				},
			},
		},
		"msgs": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "logs",
			},
			"aggs": map[string]interface{}{
				"logfields": map[string]interface{}{
					"nested": map[string]interface{}{
						"path": "logs.fields",
					},
					"aggs": map[string]interface{}{
						"logfieldsmsgs": map[string]interface{}{
							"filter": map[string]interface{}{
								"term": map[string]interface{}{
									"logs.fields.key": "msg",
								},
							},
							"aggs": map[string]interface{}{
								"logkeys": map[string]interface{}{
									"terms": map[string]interface{}{
										"field": "logs.fields.value",
										"size":  search.Limit,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	dat, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelEvents, "log terms", "query", string(dat))

	req := esapi.SearchRequest{
		Index: []string{"jaeger-span-*"},
		Body:  strings.NewReader(string(dat)),
	}
	res, err := req.Do(ctx, s.ESClient)
	if err == nil && res.StatusCode/100 != http.StatusOK/100 {
		defer res.Body.Close()
		err = fmt.Errorf("%v", res)
	}
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.SpanLog(ctx, log.DebugLevelEvents, "log terms response", "res", res)

	resp := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	aggrs, err := getMap(resp, "aggregations")
	if err != nil {
		return nil, err
	}
	terms := &SpanTerms{}
	for k, v := range aggrs {
		m, ok := v.(map[string]interface{})
		if !ok {
			log.SpanLog(ctx, log.DebugLevelEvents, "value not a struct", "key", k, "value", v)
			continue
		}
		if k == "hostnames" {
			// nested and filters add extra layers
			m, err = getMap(m, "hostnames", "tagkeys")
			if err != nil {
				return nil, err
			}
		}
		if k == "msgs" {
			// nested and filters add extra layers
			m, err = getMap(m, "logfields", "logfieldsmsgs", "logkeys")
			if err != nil {
				return nil, err
			}
		}
		if k == "tags" {
			m, err = getMap(m, "tags")
			if err != nil {
				return nil, err
			}
		}
		aggr := AggrResult{}
		err = mapstructure.WeakDecode(m, &aggr)
		if err != nil {
			return nil, err
		}
		vals := aggr.Buckets
		switch k {
		case "operations":
			terms.Operations = vals
		case "services":
			terms.Services = vals
		case "hostnames":
			terms.Hostnames = vals
		case "msgs":
			terms.Msgs = vals
		case "tags":
			terms.Tags = vals
		}
	}
	return terms, nil
}
