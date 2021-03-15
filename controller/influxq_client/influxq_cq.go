package influxq

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
)

// Struct with information used to create Continuous Query
type ContinuousQuerySettings struct {
	Measurement               string
	AggregationFunctions      map[string]string // maps new field name to an aggregation function
	NewDbName                 string
	CollectionInterval        time.Duration
	RetentionPolicyName       string
	NewRetentionPolicyCreated <-chan *RetentionPolicyCreationResult // channel that receives results of creation of new default retention policy
}

// Result of Continuous Query creation
type ContinuousQueryCreationResult struct {
	Err            error
	CqName         string
	CqTime         time.Duration
	NewMeasurement string
}

// Parameters: continuous query name, currDbName, appended list of op("field"), newmeasurementname, currmeasurement, time interval
var ContinuousQueryTemplate = "CREATE CONTINUOUS QUERY \"%s\" ON \"%s\" " +
	"BEGIN SELECT %s " +
	"INTO %s FROM \"%s\" " +
	"GROUP BY time(%s),* END"

// Adds a continuous query to current db once db is created
// Call before Start()
// rpdone is the channel returned from AddRetentionPolicy
// Continuous query will be created once current db has been created and the corresponding retention policy for rpdone has been created
func (q *InfluxQ) AddContinuousQuery(cq *ContinuousQuerySettings, numReceivers int) <-chan *ContinuousQueryCreationResult {
	if numReceivers < 1 {
		numReceivers = 1
	}
	done := make(chan *ContinuousQueryCreationResult, numReceivers)
	q.continuousQuerySettings[cq] = done
	return done
}

// Create a continuous query
// TODO: Update CQs by dropping unwanted cqs
func (q *InfluxQ) CreateContinuousQuery(cq *ContinuousQuerySettings) *ContinuousQueryCreationResult {
	res := &ContinuousQueryCreationResult{}
	if q.done {
		res.Err = fmt.Errorf("continuous query creation failed - %s db client finished", q.dbName)
		return res
	}
	if !q.dbcreated {
		res.Err = fmt.Errorf("continuous query creation failed - %s db not created yet", q.dbName)
		return res
	}
	if cq.NewRetentionPolicyCreated != nil {
		select {
		case rpRes := <-cq.NewRetentionPolicyCreated:
			if rpRes.Err != nil {
				res.Err = fmt.Errorf("continuous query requires specific retention policy - error creating retention policy: %s", rpRes.Err)
				return res
			} else {
				cq.RetentionPolicyName = rpRes.RpName
			}
		case <-time.After(5 * time.Second):
			res.Err = fmt.Errorf("continuous query requires specific retention policy - retention policy timed out on creation")
			return res
		}
	}
	selectors := ""
	firstIter := true
	for newfield, aggfunction := range cq.AggregationFunctions {
		layout := "%s AS \"%s\""
		if !firstIter {
			layout = `,` + layout
		} else {
			firstIter = false
		}
		selectors += fmt.Sprintf(layout, aggfunction, newfield)
	}
	collectionIntervalName := cq.CollectionInterval.String()
	newMeasurementName := cq.Measurement + "-" + collectionIntervalName
	fullyQualifiedMeasurementName := fmt.Sprintf("\"%s\".\"%s\".\"%s\"", cq.NewDbName, cq.RetentionPolicyName, newMeasurementName)
	cqName := newMeasurementName
	res.CqName = cqName
	res.CqTime = cq.CollectionInterval
	res.NewMeasurement = fullyQualifiedMeasurementName
	query := fmt.Sprintf(ContinuousQueryTemplate, cqName, q.dbName, selectors, fullyQualifiedMeasurementName, cq.Measurement, collectionIntervalName)
	_, err := q.QueryDB(query)
	if err != nil {
		log.DebugLog(log.DebugLevelMetrics,
			"error trying to downsample", "err", err)
		res.Err = err
		return res
	}
	return res
}

// Aggregation functions for EdgeEvents latency stats continuous queries
var LatencyAggregationFunctions = map[string]string{
	"0s":         "sum(\"0s\")",
	"5ms":        "sum(\"5ms\")",
	"10ms":       "sum(\"10ms\")",
	"25ms":       "sum(\"25ms\")",
	"50ms":       "sum(\"50ms\")",
	"100ms":      "sum(\"100ms\")",
	"min":        "min(\"min\")",
	"max":        "max(\"max\")",
	"avg":        "sum(\"total\") / sum(\"numsamples\")",
	"numsamples": "sum(\"numsamples\")",
}

func CreateLatencyContinuousQuerySettings(collectionInterval time.Duration, newDbName string, rpDone <-chan *RetentionPolicyCreationResult) *ContinuousQuerySettings {
	return &ContinuousQuerySettings{
		Measurement:               cloudcommon.LatencyMetric,
		AggregationFunctions:      LatencyAggregationFunctions,
		NewDbName:                 newDbName,
		CollectionInterval:        collectionInterval,
		RetentionPolicyName:       "autogen",
		NewRetentionPolicyCreated: rpDone,
	}
}

// Aggregation functions for EdgeEvents device info stats continuous queries
var DeviceInfoAggregationFunctions = map[string]string{
	"numsessions": "sum(\"numsessions\")",
}

func CreateDeviceInfoContinuousQuerySettings(collectionInterval time.Duration, newDbName string, rpDone <-chan *RetentionPolicyCreationResult) *ContinuousQuerySettings {
	return &ContinuousQuerySettings{
		Measurement:               cloudcommon.DeviceMetric,
		AggregationFunctions:      DeviceInfoAggregationFunctions,
		NewDbName:                 newDbName,
		CollectionInterval:        collectionInterval,
		RetentionPolicyName:       "autogen",
		NewRetentionPolicyCreated: rpDone,
	}
}
