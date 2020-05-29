package main

import (
	"context"
	"fmt"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
)

// influx timestamp ranges can handle (64-bit int min in time form)
var InfluxMinimumTimestamp, _ = time.Parse(time.RFC3339, "1677-09-21T00:13:44Z")
var InfluxMaximumTimestamp, _ = time.Parse(time.RFC3339, "2262-04-11T23:47:15Z")

var PrevCheckpoint = InfluxMinimumTimestamp
var NextCheckpoint = InfluxMaximumTimestamp //for unit tests, so getClusterCheckpoint will never sleep

func InitUsage() error {
	// set the first NextCheckpoint,
	NextCheckpoint = time.Now().Truncate(time.Minute).Add(*checkpointInterval)
	//set PrevCheckpoint, should not necessarily start at InfluxMinimumTimestamp if controller was restarted halway through operation
	influxQuery := fmt.Sprintf(`SELECT * from "%s" WHERE "checkpoint"='CHECKPOINT' order by time desc limit 1`, cloudcommon.ClusterInstUsage)
	checkpoint, err := services.events.QueryDB(influxQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}
	if len(checkpoint) == 0 || len(checkpoint[0].Series) == 0 {
		//nothing in checkpoint db, use influxMinimumTimestamp
		return nil
	} else if len(checkpoint) != 1 ||
		len(checkpoint[0].Series) != 1 ||
		len(checkpoint[0].Series[0].Values) == 0 ||
		len(checkpoint[0].Series[0].Values[0]) == 0 ||
		checkpoint[0].Series[0].Name != cloudcommon.ClusterInstUsage {
		// should only be 1 series, the 'clusterinst' one
		return fmt.Errorf("Error parsing influx response")
	}
	//we don't care about what the checkpoint actually is, just the timestamp of it
	PrevCheckpoint, err = time.Parse(time.RFC3339, fmt.Sprintf("%v", checkpoint[0].Series[0].Values[0][0]))
	if err != nil {
		PrevCheckpoint = InfluxMinimumTimestamp
		return fmt.Errorf("Error creating parsing checkpoint time: %v", err)
	}
	return nil
}

func runCheckpoints(ctx context.Context) {
	checkpointSpan := log.StartSpan(log.DebugLevelInfo, "Cluster Checkpointing thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
	defer checkpointSpan.Finish()
	err := InitUsage()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error setting up checkpoints", "err", err)
	}
	for {
		select {
		// add 2 seconds to the checkpoint bc this was actually going into the case 1 second before NextCheckpoint,
		// resulting in creating a checkpoint for the future, which is not allowed
		case <-time.After(NextCheckpoint.Add(time.Second * 2).Sub(time.Now())):
			checkpointTime := NextCheckpoint
			err = CreateClusterCheckpoint(ctx, checkpointTime)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Could not create checkpoint", "time", checkpointTime, "err", err)
			}
			// this must be AFTER the checkpoint is created, see the comments about race conditions above GetClusterCheckpoint
			PrevCheckpoint = NextCheckpoint
			NextCheckpoint = NextCheckpoint.Add(*checkpointInterval)
		}
	}
}

// checks the output of the influx log query and checks to see if it is empty(first return value) and if it is not empty then the format is what we expect(second return value)
func checkInfluxQueryOutput(result []client.Result, dbName string) (bool, error) {
	empty := false
	var valid error
	if len(result) == 0 || len(result[0].Series) == 0 {
		empty = true
	} else if len(result) != 1 ||
		len(result[0].Series) != 1 ||
		len(result[0].Series[0].Values) == 0 ||
		len(result[0].Series[0].Values[0]) == 0 ||
		result[0].Series[0].Name != dbName {
		// should only be 1 series, the 'dbName' one
		valid = fmt.Errorf("Error parsing influx, unexpected format")
	}
	return empty, valid
}
