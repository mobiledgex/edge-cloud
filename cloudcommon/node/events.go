package node

import (
	"context"
	"encoding/json"
	fmt "fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"go.uber.org/zap/zapcore"
)

// Mapping definition for events.
// Keyword types are exact matches, while text types use word matching (not grep)
var eventMapping = `
{
  "index_patterns": "%s-*",
  "settings":{
    "index.number_of_shards": 5,
    "index.number_of_replicas": 1,
    "index.mapping.nested_fields.limit": 50,
    "index.requests.cache.enable": true
  },
  "mappings":{
    "properties":{
      "name":{
        "type": "keyword",
        "ignore_above": 256
      },
      "org":{
        "type": "keyword",
        "ignore_above": 256
      },
      "type":{
        "type": "keyword",
        "ignore_above": 256
      },
      "region":{
        "type": "keyword",
        "ignore_above": 256
      },
      "timestamp":{
        "type": "date_nanos"
      },
      "error":{
        "type": "text"
      },
      "tags":{
        "type": "nested",
        "properties":{
          "key":{
            "type": "keyword",
            "ignore_above": 256
          },
          "value":{
            "type": "text"
          }
        }
      }
    }
  }
}
`

var (
	esIndexDateFormat = "2006-01"
	esEventLog        = "events-log"
	eventDebug        = false

	NoOrg = ""

	EventType = "event"
	AuditType = "audit"

	DefaultTimeDuration = 48 * time.Hour
)

type EventData struct {
	Name      string            `json:"name"`
	Org       []string          `json:"org"`
	Type      string            `json:"type"`
	Region    string            `json:"region,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Error     string            `json:"error,omitempty"`
	Tags      []EventTag        `json:"tags,omitempty"`  // this is needed for writing to elasticsearch
	Mtags     map[string]string `json:"mtags,omitempty"` // used for show output
}

type EventTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type EventMatch struct {
	Names   []string          `json:"names"`
	Orgs    []string          `json:"orgs"`
	Types   []string          `json:"types"`
	Regions []string          `json:"regions,omitempty"`
	Error   string            `json:"error,omitempty"`
	Tags    map[string]string `json:"tags"`
	Failed  bool              `json:"failed,omitempty"`
}

type EventSearch struct {
	Match          EventMatch `json:"match,omitempty"`
	NotMatch       EventMatch `json:"notmatch,omitempty"`
	AllowedOrgs    []string   `json:"allowedorgs"` // to enforce rbac
	util.TimeRange `json:",inline"`
	From           int `json:"from,omitempty"`  // start document offset
	Limit          int `json:"limit,omitempty"` // number of documents to return
}

type EventTerms struct {
	Names   []string `json:"names,omitempty"`
	Orgs    []string `json:"orgs,omitempty"`
	Types   []string `json:"types,omitempty"`
	Regions []string `json:"regions,omitempty"`
	TagKeys []string `json:"tagkeys,omitempty"`
}

// These search structs are just for parsing the search response from ElasticSearch
type SearchResp struct {
	TimedOut bool       `json:"timed_out"`
	Hits     SearchHits `json:"hits"`
}
type SearchHits struct {
	Total SearchVal      `json:"total"`
	Hits  []SearchResult `json:"hits"`
}
type SearchVal struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}
type SearchResult struct {
	Source *json.RawMessage `json:"_source"`
	Score  float32          `json:"_score"`
}

// These aggr structs are just for parsing the aggr response from ElasticSearch
type AggrResult struct {
	Buckets []AggrVal `json:"buckets"`
}
type AggrVal struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

func (s *NodeMgr) initEvents(ctx context.Context, opts *NodeOptions) error {
	s.esEvents = make(chan esapi.IndexRequest, 100)
	s.esEventsDone = make(chan struct{})
	// ES_SERVER_URLS should be set in environment, it exists as an parameter
	// option just for unit-tests.
	if opts.esUrls == "" {
		opts.esUrls = os.Getenv("ES_SERVER_URLS")
	}
	if opts.esUrls == "" {
		return nil
	}
	if s.DeploymentTag != "" {
		esEventLog = esEventLog + "-" + s.DeploymentTag
	}

	log.SpanLog(ctx, log.DebugLevelInfo, "new elastic client", "esurls", opts.esUrls)
	config := elasticsearch.Config{
		Addresses: strings.Split(opts.esUrls, ","),
	}
	tlsConfig, err := s.GetPublicClientTlsConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to get elastic search client tls config", "err", err)
		return err
	}
	if tlsConfig != nil {
		transport := http.Transport{
			// settings from http.DefaultTransport
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// custom TLS
			TLSClientConfig: tlsConfig,
		}
		config.Transport = &transport
	}

	s.ESClient, err = elasticsearch.NewClient(config)
	if err != nil {
		return err
	}
	go s.writeEvents()
	return nil
}

func (s *NodeMgr) writeEvents() {
	// initialize index
	startTime := time.Now()
	done := false
	for {
		// retry loop in case ElasticSearch not online/ready right away
		span := log.StartSpan(log.DebugLevelInfo, "init-es-events")
		ctx := log.ContextWithSpan(context.Background(), span)
		err := s.writeIndex(ctx)
		if err == nil {
			// log start once we can talk to ES
			s.EventAtTime(ctx, s.MyNode.Key.Type+" start", NoOrg, "event", s.MyNode.Key.GetTags(), nil, startTime)
			span.Finish()
			break
		}
		log.SpanLog(ctx, log.DebugLevelInfo, "write event-log failed, will retry", "err", err)
		span.Finish()
		select {
		case <-time.After(10 * time.Second):
		case <-s.esEventsDone:
			done = true
			break
		}
	}
	// write events
	for !done {
		var req esapi.IndexRequest
		select {
		case req = <-s.esEvents:
		case <-s.esEventsDone:
			done = true
			break
		}
		res, err := req.Do(context.Background(), s.ESClient)
		if err == nil && res.StatusCode/100 != http.StatusOK/100 {
			res.Body.Close()
			err = fmt.Errorf("%v", res)
		}
		if err != nil {
			span := log.StartSpan(log.DebugLevelEvents, "es-write-event")
			ctx := log.ContextWithSpan(context.Background(), span)
			log.SpanLog(ctx, log.DebugLevelEvents, "failed to log event", "err", err)
			span.Finish()
			continue
		}
		res.Body.Close()
	}
}

func (s *NodeMgr) writeIndex(ctx context.Context) error {
	mapping := fmt.Sprintf(eventMapping, esEventLog)
	req := esapi.IndicesPutTemplateRequest{
		Name: esEventLog,
		Body: strings.NewReader(mapping),
	}
	res, err := req.Do(ctx, s.ESClient)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != http.StatusOK/100 {
		return fmt.Errorf("Bad status: %v", res)
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "write event-log index template", "res", res)
	return nil
}

func (s *NodeMgr) Event(ctx context.Context, name, org string, keyTags map[string]string, err error, keysAndValues ...string) {
	s.EventAtTime(ctx, name, org, "event", keyTags, err, time.Now(), keysAndValues...)
}

func (s *NodeMgr) TimedEvent(ctx context.Context, name, org, typ string, keyTags map[string]string, err error, startTime, endTime time.Time, keysAndValues ...string) {
	keysAndValues = append(keysAndValues,
		"duration", endTime.Sub(startTime).String(),
	)
	s.event(ctx, name, org, typ, keyTags, err, startTime, keysAndValues...)
}

// EventAtTime is the same as event(), but we need the extra level of call
// stack to get the runtime.Caller() lineno correctly in all cases.
func (s *NodeMgr) EventAtTime(ctx context.Context, name, org, typ string, keyTags map[string]string, err error, ts time.Time, keysAndValues ...string) {
	s.event(ctx, name, org, typ, keyTags, err, ts, keysAndValues...)
}

func (s *NodeMgr) event(ctx context.Context, name, org, typ string, keyTags map[string]string, err error, ts time.Time, keysAndValues ...string) {
	if s.ESClient == nil {
		return
	}
	event := EventData{
		Name:      name,
		Org:       []string{org},
		Type:      typ,
		Region:    s.Region,
		Timestamp: ts,
	}
	if err != nil {
		event.Error = err.Error()
	}

	ospan := log.SpanFromContext(ctx)
	if ospan != nil {
		if span, ok := ospan.(*log.Span); ok {
			spanctx := span.SpanContext()
			event.Tags = append(event.Tags,
				EventTag{"traceid", spanctx.TraceID().String()},
				EventTag{"spanid", spanctx.SpanID().String()},
			)
		}
	}
	eventRegion := event.Region
	cloudletKey := edgeproto.CloudletKey{}
	for k, v := range keyTags {
		event.Tags = append(event.Tags, EventTag{k, v})
		if k == edgeproto.CloudletKeyTagOrganization {
			cloudletKey.Organization = v
		} else if k == edgeproto.CloudletKeyTagName {
			cloudletKey.Name = v
		} else if k == "region" && eventRegion == "" {
			eventRegion = v
		}
	}
	if len(keysAndValues)%2 != 0 {
		panic(fmt.Sprintf("non-even keyAndValues len: %d", len(keysAndValues)))
	}
	for ii := 0; ii*2 < len(keysAndValues); ii++ {
		key := keysAndValues[ii*2]
		val := keysAndValues[ii*2+1]
		event.Tags = append(event.Tags, EventTag{key, val})
	}
	// To allow operators to see events on Cloudlets in CloudletPools,
	// add the operator org to the allowed orgs if the event is tagged
	// with such a cloudlet.
	if cloudletKey.Organization != "" && cloudletKey.Name != "" &&
		cloudletKey.Organization != org {
		if s.CloudletPoolLookup.InPool(eventRegion, cloudletKey) {
			event.Org = append(event.Org, cloudletKey.Organization)
		}
	}

	ec := zapcore.NewEntryCaller(runtime.Caller(2))
	event.Tags = append(event.Tags, EventTag{"lineno", ec.TrimmedPath()})
	event.Tags = append(event.Tags, EventTag{"hostname", s.MyNode.Hostname})

	dat, err := json.Marshal(event)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to marshal event", "event", event, "err", err)
		return
	}

	req := esapi.IndexRequest{
		Index: esEventLog + "-" + indexTime(ts),
		Body:  strings.NewReader(string(dat)),
	}
	select {
	case s.esEvents <- req:
		log.SpanLog(ctx, log.DebugLevelEvents, "queued event", "event", string(dat))
	default:
		log.SpanLog(ctx, log.DebugLevelEvents, "dropped event", "event", string(dat))
	}
}

func (s *NodeMgr) ShowEvents(ctx context.Context, search *EventSearch) ([]EventData, error) {
	return s.searchEvents(ctx, searchTypeFilter, search)
}

func (s *NodeMgr) FindEvents(ctx context.Context, search *EventSearch) ([]EventData, error) {
	return s.searchEvents(ctx, searchTypeRelevance, search)
}

var (
	searchTypeFilter    = "filter"
	searchTypeRelevance = "relevance"
)

func (s *NodeMgr) searchEvents(ctx context.Context, searchType string, search *EventSearch) ([]EventData, error) {
	query, err := s.getQuery(ctx, searchType, search)
	if err != nil {
		return nil, err
	}
	dat, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelEvents, "event search", "search", string(dat))

	req := esapi.SearchRequest{
		Index: []string{esEventLog + "-*"},
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

	log.SpanLog(ctx, log.DebugLevelEvents, "event search response", "res", res)

	resp := SearchResp{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	if resp.Hits.Total.Value == 0 || resp.Hits.Hits == nil {
		return []EventData{}, nil
	}
	out := []EventData{}
	for _, hit := range resp.Hits.Hits {
		ed := EventData{}
		err = json.Unmarshal(*hit.Source, &ed)
		if err != nil {
			return nil, err
		}
		if s.unitTestMode {
			log.SpanLog(ctx, log.DebugLevelInfo, "searchEvents result", "event", ed.Name, "score", hit.Score)
		}
		if len(ed.Tags) > 0 {
			ed.Mtags = make(map[string]string)
			for _, tag := range ed.Tags {
				ed.Mtags[tag.Key] = tag.Value
			}
			ed.Tags = nil
		}
		out = append(out, ed)
	}
	return out, nil
}

func (s *NodeMgr) getQuery(ctx context.Context, searchType string, search *EventSearch) (map[string]interface{}, error) {
	// ElasticSearch is super flexible in terms of how it can search.
	// To simplify things, we offer only the two opposite extremes.

	// The "filter" search returns the AND of all search criteria, sorted
	// by timestamp. This is probably the typical use case where one wants
	// to see all events for a particular Org or AppInst in chronological
	// order. All specified criteria must match or the event is filtered out.

	// The "relevance" search returns the OR of all search criteria,
	// sorted by matching score. This is typically used when you're not
	// exactly sure what you're looking for, but want to specify a few
	// things that might be relevant. Not all search criteria needs to match,
	// except for a few things like the time range.

	// Of note is that the "filter" searches can be cached by ElasticSearch
	// and are faster.
	if searchType != searchTypeFilter && searchType != searchTypeRelevance {
		return nil, fmt.Errorf("invalid event search type %s", searchType)
	}

	// get match criteria
	smaps, err := s.getMatchQueries(ctx, searchType, &search.Match)
	if err != nil {
		return nil, err
	}
	// get not match criteria
	mustnot, err := s.getMatchQueries(ctx, searchTypeFilter, &search.NotMatch)
	if err != nil {
		return nil, err
	}

	// search params that are always filtered
	// allowed orgs for rbac enforcement
	filter := []map[string]interface{}{}
	if len(search.AllowedOrgs) > 0 {
		filter = append(filter, esSearchKeywords("org", search.AllowedOrgs))
	}
	// time range is always specified
	// Resolve{} is idempotent so may already have been done.
	err = search.TimeRange.Resolve(DefaultTimeDuration)
	if err != nil {
		return nil, err
	}
	st := search.StartTime.UTC().Format(time.RFC3339Nano)
	et := search.EndTime.UTC().Format(time.RFC3339Nano)
	timerange := map[string]interface{}{
		"range": map[string]interface{}{
			"timestamp": map[string]interface{}{
				"gte": st,
				"lt":  et,
			},
		},
	}
	filter = append(filter, timerange)

	query := map[string]interface{}{}
	if searchType == searchTypeFilter {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"filter":   append(smaps, filter...),
				"must_not": mustnot,
			},
		}
		// sort by timestamp
		query["sort"] = []map[string]interface{}{
			map[string]interface{}{
				"timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	} else {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"filter":   filter,
				"should":   smaps,
				"must_not": mustnot,
			},
		}
	}

	if search.From != 0 {
		query["from"] = search.From
	}
	if search.Limit != 0 {
		query["size"] = search.Limit
	}
	return query, nil
}

func (s *NodeMgr) getMatchQueries(ctx context.Context, searchType string, match *EventMatch) ([]map[string]interface{}, error) {
	nestedQuery := "filter"
	if searchType == searchTypeRelevance {
		nestedQuery = "should"
	}
	// Search parameters (search maps)
	smaps := []map[string]interface{}{}
	if len(match.Names) > 0 {
		smaps = append(smaps, esSearchKeywords("name", match.Names))
	}
	if len(match.Regions) > 0 {
		smaps = append(smaps, esSearchKeywords("region", match.Regions))
	}
	if len(match.Orgs) > 0 {
		smaps = append(smaps, esSearchKeywords("org", match.Orgs))
	}
	if len(match.Types) > 0 {
		smaps = append(smaps, esSearchKeywords("type", match.Types))
	}
	if match.Error != "" {
		smaps = append(smaps, esSearchText("match", "error", match.Error, searchType))
	}
	if match.Failed {
		smaps = append(smaps, esSearchText("exists", "field", "error", searchType))
	}
	if len(match.Tags) > 0 {
		for key, value := range match.Tags {
			if strings.ContainsRune(value, '*') && strings.ContainsRune(value, '-') {
				// the standard analyzer splits words based on
				// whitespace and other chars (like -), so specifying
				// - will never match because for text fields,
				// we are only matching words in the field, not
				// the whole field value.
				return nil, fmt.Errorf("for wildcard matching, don't use '-' because the value is split up by '-' and wildcard is matching words, not the full text")
			}
			qs := []map[string]interface{}{}
			qs = append(qs,
				esSearchText("term", "tags.key", key, searchType),
				esSearchText("match", "tags.value", value, searchType),
			)
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
			smaps = append(smaps, nested)
		}
	}
	return smaps, nil
}

func esSearchText(typ, field, val, searchType string) map[string]interface{} {
	var q interface{}
	if typ == "match" && strings.ContainsRune(val, '*') {
		// if user specified * they probably want wildcard matching
		// for some reason text must be lower case or it won't match
		typ = "wildcard"
		qmap := make(map[string]interface{})
		qmap["value"] = strings.ToLower(val)
		q = qmap
	} else if typ == "match" {
		qmap := make(map[string]interface{})
		qmap["query"] = val
		if searchType == searchTypeFilter {
			qmap["operator"] = "AND"
		}
		q = qmap
	} else {
		// term, wildcard, exists (short form without extra map hierarchy)
		q = val
	}
	return map[string]interface{}{
		typ: map[string]interface{}{
			field: q,
		},
	}
}

func esSearchKeywords(field string, vals []string) map[string]interface{} {
	// multiple values for keywords are treated as an OR
	smaps := []map[string]interface{}{}
	for _, val := range vals {
		q := make(map[string]interface{})
		if strings.ContainsRune(val, '*') {
			// wildcard matching.
			q["wildcard"] = map[string]interface{}{
				field: val,
			}
		} else {
			q["term"] = map[string]interface{}{
				field: val,
			}
		}
		smaps = append(smaps, q)
	}
	return map[string]interface{}{
		"bool": map[string]interface{}{
			"should":               smaps,
			"minimum_should_match": 1,
		},
	}
}

func indexTime(t time.Time) string {
	return t.UTC().Format(esIndexDateFormat)
}

func (s *NodeMgr) EventTerms(ctx context.Context, search *EventSearch) (*EventTerms, error) {
	query, err := s.getQuery(ctx, searchTypeFilter, search)
	if err != nil {
		return nil, err
	}
	// Set size to 0 because we don't need the overhead of returned results,
	// because we're going to ignore them. This does however disable pagination
	// results for event terms, but it's unlikely pagination will be needed
	// in this context.
	query["size"] = 0
	query["aggs"] = map[string]interface{}{
		"names": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "name",
				"size":  100,
			},
		},
		"orgs": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "org",
				"size":  100,
			},
		},
		"types": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "type",
				"size":  100,
			},
		},
		"regions": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "region",
				"size":  100,
			},
		},
		"tagkeys": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "tags",
			},
			"aggs": map[string]interface{}{
				"tagkeys": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "tags.key",
						"size":  500,
					},
				},
			},
		},
	}

	dat, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelEvents, "event terms", "query", string(dat))

	req := esapi.SearchRequest{
		Index: []string{esEventLog + "-*"},
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

	log.SpanLog(ctx, log.DebugLevelEvents, "event terms response", "res", res)

	resp := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	aggrs, err := getMap(resp, "aggregations")
	if err != nil {
		return nil, err
	}
	terms := &EventTerms{}
	for k, v := range aggrs {
		m, ok := v.(map[string]interface{})
		if !ok {
			log.SpanLog(ctx, log.DebugLevelEvents, "value not a struct", "key", k, "value", v)
			continue
		}
		if k == "tagkeys" {
			// nested fields have an extra layer of hierarchy
			m, err = getMap(m, k)
			if err != nil {
				return nil, err
			}
		}
		aggr := AggrResult{}
		err = mapstructure.WeakDecode(m, &aggr)
		if err != nil {
			return nil, err
		}
		vals := []string{}
		for _, bucket := range aggr.Buckets {
			vals = append(vals, bucket.Key)
		}
		sort.Strings(vals)
		switch k {
		case "names":
			terms.Names = vals
		case "orgs":
			terms.Orgs = vals
		case "types":
			terms.Types = vals
		case "regions":
			terms.Regions = vals
		case "tagkeys":
			terms.TagKeys = vals
		}
	}
	return terms, nil
}

func getMap(source map[string]interface{}, key string) (map[string]interface{}, error) {
	i, found := source[key]
	if !found {
		return nil, fmt.Errorf("key %s not found in %v", key, source)
	}
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %s not a map in %v", key, source)
	}
	return m, nil
}
