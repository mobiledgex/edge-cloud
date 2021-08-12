package influxq

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
)

// Struct with information used to create Continuous Query
type ContinuousQuerySettings struct {
	Measurement          string
	AggregationFunctions map[string]string // maps new field name to an aggregation function
	CollectionInterval   time.Duration
	RetentionPolicyTime  time.Duration
}

// Parameters: continuous query name, destinationDbName, appended list of op("field"), newmeasurementname, originmeasurement, time interval
var ContinuousQueryTemplate = "CREATE CONTINUOUS QUERY \"%s\" ON \"%s\" " +
	"BEGIN SELECT %s " +
	"INTO %s FROM \"%s\" " +
	"GROUP BY time(%s),* END"

func CreateContinuousQuery(origin *InfluxQ, dest *InfluxQ, cq *ContinuousQuerySettings) error {
	// make sure db is running
	if origin.done {
		return fmt.Errorf("continuous query creation failed - %s db client finished", origin.dbName)
	}

	// make sure db has been created before moving on
	numTries := 3
	for i := 1; i <= numTries; i++ {
		if origin.dbcreated {
			break
		}
		if i == numTries {
			return fmt.Errorf("continuous query creation failed - %s db not created yet", origin.dbName)
		}
		time.Sleep(time.Second)
	}

	// create retention policy if specified
	var rpName string
	if cq.RetentionPolicyTime == 0 { // use default retention policy
		rpName = getDefaultRetentionPolicyName(dest.dbName)
	} else {
		err := dest.CreateRetentionPolicy(cq.RetentionPolicyTime, NonDefaultRetentionPolicy)
		if err != nil {
			return fmt.Errorf("continuous query createion failed - unable to create retention policy for continuous qyer, error is %s", err.Error())
		}
		rpName = getNonDefaultRetentionPolicyName(dest.dbName, cq.RetentionPolicyTime)
	}

	// create continuous query with created retention
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
	newMeasurementName := cloudcommon.CreateInfluxMeasurementName(cq.Measurement, cq.CollectionInterval)
	fullyQualifiedMeasurementName := fmt.Sprintf("\"%s\".\"%s\".\"%s\"", dest.dbName, rpName, newMeasurementName)
	cqName := newMeasurementName
	query := fmt.Sprintf(ContinuousQueryTemplate, cqName, origin.dbName, selectors, fullyQualifiedMeasurementName, cq.Measurement, cq.CollectionInterval.String())
	if _, err := origin.QueryDB(query); err != nil {
		log.DebugLog(log.DebugLevelMetrics,
			"error trying to downsample", "err", err)
		return err
	}
	origin.continuousQueries[cqName] = struct{}{}
	return nil
}

// Parameters: continuous query name and DbName
var DropContinuousQueryTemplate = "DROP CONTINUOUS QUERY \"%s\" ON \"%s\""

// Drop ContinuousQuery
func DropContinuousQuery(origin *InfluxQ, dest *InfluxQ, measurement string, interval time.Duration, retention time.Duration) error {
	// drop old retention policy
	if err := dest.DropRetentionPolicy(getNonDefaultRetentionPolicyName(dest.dbName, retention)); err != nil {
		return err
	}

	// drop old cq
	cqName := cloudcommon.CreateInfluxMeasurementName(measurement, interval)
	query := fmt.Sprintf(DropContinuousQueryTemplate, cqName, origin.dbName)
	if _, err := origin.QueryDB(query); err != nil {
		return err
	}
	delete(origin.continuousQueries, cqName)
	return nil
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
	"total":      "sum(\"total\")",
	"avg":        "sum(\"total\") / sum(\"numsamples\")",
	"numsamples": "sum(\"numsamples\")",
}

func CreateLatencyContinuousQuerySettings(collectionInterval time.Duration, retention time.Duration) *ContinuousQuerySettings {
	return &ContinuousQuerySettings{
		Measurement:          cloudcommon.LatencyMetric,
		AggregationFunctions: LatencyAggregationFunctions,
		CollectionInterval:   collectionInterval,
		RetentionPolicyTime:  retention,
	}
}

// Aggregation functions for EdgeEvents device info stats continuous queries
var DeviceInfoAggregationFunctions = map[string]string{
	"numsessions": "sum(\"numsessions\")",
}

func CreateDeviceInfoContinuousQuerySettings(collectionInterval time.Duration, retention time.Duration) *ContinuousQuerySettings {
	return &ContinuousQuerySettings{
		Measurement:          cloudcommon.DeviceMetric,
		AggregationFunctions: DeviceInfoAggregationFunctions,
		CollectionInterval:   collectionInterval,
		RetentionPolicyTime:  retention,
	}
}
