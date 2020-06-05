package main

import (
	"context"
	"fmt"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
)

type RunTimeStats struct {
	start  time.Time
	end    time.Time
	upTime time.Duration
	event  cloudcommon.Usage_event
	status string
}

var usageTypeVmApp = "VmApp"
var usageTypeCluster = "Cluster"

// influx timestamp ranges can handle (64-bit int min in time form)
var InfluxMinimumTimestamp, _ = time.Parse(time.RFC3339, "1677-09-21T00:13:44Z")
var InfluxMaximumTimestamp, _ = time.Parse(time.RFC3339, "2262-04-11T23:47:15Z")

var PrevCheckpoint = InfluxMinimumTimestamp
var NextCheckpoint = InfluxMaximumTimestamp //for unit tests, so getClusterCheckpoint will never sleep

var GetCheckpointInfluxQueryTemplate = `SELECT %s from "%s" WHERE "org"='%s' AND "checkpoint"='CHECKPOINT' AND time <= '%s' order by time desc`
var CreateCheckpointInfluxQueryTemplate = `SELECT %s from "%s" WHERE time >= '%s' AND time < '%s' order by time desc`
var CreateCheckpointInfluxUsageQueryTemplate = `SELECT %s from "%s" WHERE checkpoint='CHECKPOINT' AND time >= '%s' AND time < '%s' order by time desc`

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
	checkpointSpan := log.StartSpan(log.DebugLevelInfo, "Checkpointing thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
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
				log.SpanLog(ctx, log.DebugLevelInfo, "Could not create cluster checkpoint", "time", checkpointTime, "err", err)
			}
			err = CreateVmAppCheckpoint(ctx, checkpointTime)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Could not create app checkpoint", "time", checkpointTime, "err", err)
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

func checkpointTimeValid(timestamp time.Time) error {
	now := time.Now()
	if timestamp.After(now) {
		return fmt.Errorf("Cannot create a checkpoint for the future, checkpointTimestamp: %s, now:%s", timestamp.Format(time.RFC3339), now.Format(time.RFC3339))
	}
	return nil
}

func GetRunTimeStats(usageType string, checkpoint, key interface{}, logs []client.Result, stats *RunTimeStats) error {
	var influxMeasurementName string
	var runTime, downTime time.Duration
	var startTime, checkpointTimestamp time.Time
	var downTimeUpper = stats.end
	checkpointStatus := cloudcommon.InstanceDown

	if usageType == usageTypeVmApp {
		influxMeasurementName = cloudcommon.AppInstEvent
		appCheckpoint, ok := checkpoint.(AppCheckpoint)
		if !ok {
			return fmt.Errorf("Unable to cast app checkpoint")
		}
		castedKey, ok := key.(edgeproto.AppInstKey)
		if !ok {
			return fmt.Errorf("Unable to cast appinstkey")
		}
		startTime = appCheckpoint.Timestamp
		checkpointTimestamp = appCheckpoint.Timestamp
		for i, key := range appCheckpoint.Keys {
			if key.Matches(&castedKey) {
				checkpointStatus = appCheckpoint.Status[i]
			}
		}
	} else if usageType == usageTypeCluster {
		influxMeasurementName = cloudcommon.ClusterInstEvent
		clusterCheckpoint, ok := checkpoint.(ClusterCheckpoint)
		if !ok {
			return fmt.Errorf("Unable to cast cluster checkpoint")
		}
		castedKey, ok := key.(edgeproto.ClusterInstKey)
		if !ok {
			return fmt.Errorf("Unable to cast clusterinstkey")
		}
		startTime = clusterCheckpoint.Timestamp
		checkpointTimestamp = clusterCheckpoint.Timestamp
		for i, key := range clusterCheckpoint.Keys {
			if key.Matches(&castedKey) {
				checkpointStatus = clusterCheckpoint.Status[i]
			}
		}
	} else {
		return fmt.Errorf("unknown usage type")
	}

	empty, err := checkInfluxQueryOutput(logs, influxMeasurementName)
	if err != nil {
		return err
	} else if empty {
		//there are no logs between endTime and the checkpoint
		var totalRunTime time.Duration
		if checkpointStatus == cloudcommon.InstanceDown {
			totalRunTime = time.Duration(0)
		} else {
			totalRunTime = stats.end.Sub(checkpointTimestamp)
		}
		stats.start = checkpointTimestamp
		stats.upTime = totalRunTime
		stats.status = checkpointStatus
		return nil
	}

	var latestStatus = ""
	for _, values := range logs[0].Series[0].Values {
		// value should be of the format [timestamp event status]
		if len(values) != 3 {
			return fmt.Errorf("Error parsing influx response")
		}
		timestamp, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[0]))
		if err != nil {
			return fmt.Errorf("Unable to parse timestamp: %v", err)
		}
		event := cloudcommon.InstanceEvent(fmt.Sprintf("%v", values[1]))
		status := fmt.Sprintf("%v", values[2])
		if latestStatus == "" { // keep track of the latest seen status in the logs for checkpointing purposes
			latestStatus = status
		}
		// go until we hit the creation/reservation of this appinst, OR until we hit the checkpoint
		if timestamp.Before(checkpointTimestamp) {
			// Check to see if it was up or down at the checkpoint and set the downTime accordingly and then calculate the runTime
			if checkpointStatus == cloudcommon.InstanceDown {
				downTime = downTime + downTimeUpper.Sub(checkpointTimestamp)
			}
			runTime = stats.end.Sub(checkpointTimestamp) - downTime
			latestStatus = checkpointStatus
			break
		}
		// if we reached the creation event calculate the runTime
		if event == cloudcommon.CREATED || event == cloudcommon.RESERVED {
			if status == cloudcommon.InstanceDown { // don't think this scenario would ever happen but just in case
				downTime = downTime + downTimeUpper.Sub(timestamp)
			}
			runTime = stats.end.Sub(timestamp) - downTime
			startTime = timestamp
			break
		}
		// add to the downtime
		if status == cloudcommon.InstanceDown {
			downTime = downTime + downTimeUpper.Sub(timestamp)
		}
		downTimeUpper = timestamp
	}

	stats.start = startTime
	stats.upTime = runTime
	stats.status = latestStatus
	return nil
}
