package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
)

type ClusterCheckpoint struct {
	Timestamp time.Time
	Org       string
	Keys      []*edgeproto.ClusterInstKey
	Status    []string // either cloudcommon.InstanceUp or cloudcommon.InstanceDown
}

var ClusterUsageInfluxQueryTemplate = `SELECT %s from "%s" WHERE "clusterorg"='%s' AND "cluster"='%s' AND "cloudlet"='%s' AND "cloudletorg"='%s' %sAND time >= '%s' AND time < '%s' order by time desc`
var GetCheckpointInfluxQueryTemplate = `SELECT %s from "%s" WHERE "org"='%s' AND "checkpoint"='CHECKPOINT' AND time <= '%s' order by time desc`

var CreateClusterCheckpointInfluxQueryTemplate = `SELECT %s from "%s" WHERE time >= '%s' AND time < '%s' order by time desc`
var CreateClusterCheckpointInfluxUsageQueryTemplate = `SELECT %s from "%s" WHERE checkpoint='CHECKPOINT' AND time >= '%s' AND time < '%s' order by time desc`

func runClusterCheckpoints(ctx context.Context) {
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

func CreateClusterUsageRecord(ctx context.Context, cluster *edgeproto.ClusterInst, endTime time.Time, usageEvent cloudcommon.Usage_event) error {
	var metric *edgeproto.Metric
	// query from the checkpoint up to the event
	selectors := []string{"\"event\"", "\"status\""}
	reservedByOption := ""
	org := cluster.Key.Organization
	if cluster.Key.Organization == cloudcommon.OrganizationMobiledgeX && cluster.ReservedBy != "" {
		reservedByOption = fmt.Sprintf(`AND "reservedBy"='%s' `, cluster.ReservedBy)
		org = cluster.ReservedBy
	}
	checkpoint, err := GetClusterCheckpoint(ctx, org, endTime)
	if err != nil {
		return fmt.Errorf("unable to retrieve Checkpoint: %v", err)
	}
	influxLogQuery := fmt.Sprintf(ClusterUsageInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstEvent,
		cluster.Key.Organization,
		cluster.Key.ClusterKey.Name,
		cluster.Key.CloudletKey.Name,
		cluster.Key.CloudletKey.Organization,
		reservedByOption,
		checkpoint.Timestamp.Format(time.RFC3339),
		endTime.Format(time.RFC3339))
	logs, err := services.events.QueryDB(influxLogQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}
	var runTime, downTime time.Duration
	var downTimeUpper = endTime
	startTime := checkpoint.Timestamp
	checkpointStatus := cloudcommon.InstanceDown
	for i, key := range checkpoint.Keys {
		if key.Matches(&cluster.Key) {
			checkpointStatus = checkpoint.Status[i]
		}
	}

	empty, err := checkInfluxQueryOutput(logs, cloudcommon.ClusterInstEvent)
	if err != nil {
		return err
	} else if empty {
		//there are no logs between endTime and the checkpoint
		var totalRunTime time.Duration
		if checkpointStatus == cloudcommon.InstanceDown {
			totalRunTime = time.Duration(0)
		} else {
			totalRunTime = endTime.Sub(checkpoint.Timestamp)
		}
		metric = createClusterUsageMetric(cluster, checkpoint.Timestamp, endTime, totalRunTime, usageEvent, checkpointStatus)
		services.events.AddMetric(metric)
		return nil
	}

	var latestStatus = ""
	for _, values := range logs[0].Series[0].Values {
		// value should be of the format [timestamp event status]
		if len(values) != len(selectors)+1 {
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
		if timestamp.Before(checkpoint.Timestamp) {
			// Check to see if it was up or down at the checkpoint and set the downTime accordingly and then calculate the runTime
			if checkpointStatus == cloudcommon.InstanceDown {
				downTime = downTime + downTimeUpper.Sub(checkpoint.Timestamp)
			}
			runTime = endTime.Sub(checkpoint.Timestamp) - downTime
			latestStatus = checkpointStatus
			break
		}
		// if we reached the creation event calculate the runTime
		if event == cloudcommon.CREATED || event == cloudcommon.RESERVED {
			if status == cloudcommon.InstanceDown { // don't think this scenario would ever happen but just in case
				downTime = downTime + downTimeUpper.Sub(timestamp)
			}
			runTime = endTime.Sub(timestamp) - downTime
			startTime = timestamp
			break
		}
		// add to the downtime
		if status == cloudcommon.InstanceDown {
			downTime = downTime + downTimeUpper.Sub(timestamp)
		}
		downTimeUpper = timestamp
	}

	// write the usage record to influx
	metric = createClusterUsageMetric(cluster, startTime, endTime, runTime, usageEvent, latestStatus)

	services.events.AddMetric(metric)
	return nil
}

func createClusterUsageMetric(cluster *edgeproto.ClusterInst, startTime, endTime time.Time, runTime time.Duration, usageEvent cloudcommon.Usage_event, status string) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.ClusterInstUsage
	now := time.Now()
	ts, _ := types.TimestampProto(now)
	metric.Timestamp = *ts
	utc, _ := time.LoadLocation("UTC")
	//start and endtimes end up being put into different timezones somehow when going through calculations so force them both to the same here
	startUTC := startTime.In(utc)
	endUTC := endTime.In(utc)

	// influx requires that at least one field must be specified when querying so these cant be all tags
	metric.AddStringVal("cloudletorg", cluster.Key.CloudletKey.Organization)
	metric.AddTag("cloudlet", cluster.Key.CloudletKey.Name)
	metric.AddTag("cluster", cluster.Key.ClusterKey.Name)
	metric.AddTag("clusterorg", cluster.Key.Organization)
	metric.AddTag("flavor", cluster.Flavor.Name)
	metric.AddStringVal("start", startUTC.Format(time.RFC3339))
	metric.AddStringVal("end", endUTC.Format(time.RFC3339))
	metric.AddDoubleVal("uptime", runTime.Seconds())
	if cluster.ReservedBy != "" && cluster.Key.Organization == cloudcommon.OrganizationMobiledgeX {
		metric.AddTag("org", cluster.ReservedBy)
	} else {
		metric.AddTag("org", cluster.Key.Organization)
	}
	checkpointVal := ""
	writeStatus := ""
	if usageEvent == cloudcommon.USAGE_EVENT_CHECKPOINT {
		checkpointVal = "CHECKPOINT"
		writeStatus = status // only care about the status if its a checkpoint
	}
	metric.AddTag("checkpoint", checkpointVal)
	metric.AddTag("status", writeStatus)
	return &metric
}

// This is checkpointing for BILLING PERIODS ONLY, custom checkpointing, coming later, will not create(or at least store) usage records upon checkpointing
func CreateClusterCheckpoint(ctx context.Context, timestamp time.Time) error {
	now := time.Now()
	if timestamp.After(now) { // we dont know if there will be more creates and deletes before the timestamp occurs
		return fmt.Errorf("Cannot create a checkpoint for the future, checkpointTimestamp: %s, now:%s", timestamp.Format(time.RFC3339), now.Format(time.RFC3339))
	}
	defer services.events.DoPush() // flush these right away for subsequent calls to GetClusterCheckpoint
	// get all running clusterinsts and create a usage record of them
	selectors := []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"event\""}
	influxLogQuery := fmt.Sprintf(CreateClusterCheckpointInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstEvent,
		PrevCheckpoint.Format(time.RFC3339),
		timestamp.Format(time.RFC3339))
	logs, err := services.events.QueryDB(influxLogQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}

	empty, err := checkInfluxQueryOutput(logs, cloudcommon.ClusterInstEvent)
	skipLogCheck := false
	if err != nil {
		return err
	} else if empty {
		//there are no logs between endTime and the checkpoint, just copy over the checkpoint
		skipLogCheck = true
	}

	seenClusters := make(map[edgeproto.ClusterInstKey]bool)
	if !skipLogCheck {
		for _, values := range logs[0].Series[0].Values {
			// value should be of the format [timestamp cluster clusterorg cloudlet cloudletorg event status]
			if len(values) != len(selectors)+1 {
				return fmt.Errorf("Error parsing influx response")
			}
			cluster := fmt.Sprintf("%v", values[1])
			clusterorg := fmt.Sprintf("%v", values[2])
			cloudlet := fmt.Sprintf("%v", values[3])
			cloudletorg := fmt.Sprintf("%v", values[4])
			event := cloudcommon.InstanceEvent(fmt.Sprintf("%v", values[5]))
			key := edgeproto.ClusterInstKey{
				ClusterKey:   edgeproto.ClusterKey{Name: cluster},
				Organization: clusterorg,
				CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
			}
			// only care about each clusterinsts most recent log
			if _, exists := seenClusters[key]; exists {
				continue
			}
			seenClusters[key] = true
			// if its still up, record it
			if event != cloudcommon.DELETED && event != cloudcommon.UNRESERVED {
				info := edgeproto.ClusterInst{}
				if !clusterInstApi.cache.Get(&key, &info) {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Could not find clusterinst even though event log indicates it is up", "cluster", key)
					continue
				}
				//record the usage up to this point
				err = CreateClusterUsageRecord(ctx, &info, timestamp, cloudcommon.USAGE_EVENT_CHECKPOINT)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create cluster usage record of checkpointed cluster", "cluster", key, "err", err)
				}
			}
		}
	}

	// check for apps that got checkpointed but did not have any log events between PrevCheckpoint and this one
	selectors = []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\""}
	influxCheckpointQuery := fmt.Sprintf(CreateClusterCheckpointInfluxUsageQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstUsage,
		PrevCheckpoint.Add(-1*time.Minute).Format(time.RFC3339), //small delta to account for conversion rounding inconsistencies
		PrevCheckpoint.Add(time.Minute).Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}

	empty, err = checkInfluxQueryOutput(checkpoints, cloudcommon.ClusterInstUsage)
	if err != nil {
		return err
	} else if empty {
		// no checkpoints made yet, or nothing got checkpointed last time, dont need to do this check
		return nil
	}

	for _, values := range checkpoints[0].Series[0].Values {
		// value should be of the format [timestamp cluster clusterorg cloudlet cloudletorg org status]
		if len(values) != len(selectors)+1 {
			return fmt.Errorf("Error parsing influx response")
		}
		cluster := fmt.Sprintf("%v", values[1])
		clusterorg := fmt.Sprintf("%v", values[2])
		cloudlet := fmt.Sprintf("%v", values[3])
		cloudletorg := fmt.Sprintf("%v", values[4])
		key := edgeproto.ClusterInstKey{
			ClusterKey:   edgeproto.ClusterKey{Name: cluster},
			Organization: clusterorg,
			CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
		}
		// only care about each clusterinsts most recent log
		if _, exists := seenClusters[key]; exists {
			continue
		}
		seenClusters[key] = true
		// record it
		info := edgeproto.ClusterInst{}
		if !clusterInstApi.cache.Get(&key, &info) {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Could not find clusterinst even though event log indicates it is up", "cluster", key)
			continue
		}
		err = CreateClusterUsageRecord(ctx, &info, timestamp, cloudcommon.USAGE_EVENT_CHECKPOINT)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create cluster usage record of checkpointed cluster", "cluster", key, "err", err)
		}
	}

	return nil
}

// returns all the checkpointed clusterinsts of the most recent checkpoint with regards to timestamp
func GetClusterCheckpoint(ctx context.Context, org string, timestamp time.Time) (*ClusterCheckpoint, error) {
	// wait until the current checkpoint is done if we want to access it, to prevent race conditions with CreateCheckpoint
	for timestamp.After(NextCheckpoint) {
		time.Sleep(time.Second)
	}
	// query from the checkpoint up to the delete
	selectors := []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"status\""}
	influxCheckpointQuery := fmt.Sprintf(GetCheckpointInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstUsage,
		org,
		timestamp.Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return nil, fmt.Errorf("Unable to query influx: %v", err)
	}
	result := ClusterCheckpoint{
		Timestamp: PrevCheckpoint,
		Org:       org,
		Keys:      make([]*edgeproto.ClusterInstKey, 0),
		Status:    make([]string, 0),
	}

	empty, err := checkInfluxQueryOutput(checkpoints, cloudcommon.ClusterInstUsage)
	if err != nil {
		return nil, err
	} else if empty {
		return &result, nil
	}

	for i, values := range checkpoints[0].Series[0].Values {
		// value should be of the format [measurementTime cluster clusterorg cloudlet cloudletorg status]
		if len(values) != len(selectors)+1 {
			return nil, fmt.Errorf("Error parsing influx response")
		}
		measurementTime, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[0]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse timestamp of checkpoint")
		}
		cluster := fmt.Sprintf("%v", values[1])
		clusterorg := fmt.Sprintf("%v", values[2])
		cloudlet := fmt.Sprintf("%v", values[3])
		cloudletorg := fmt.Sprintf("%v", values[4])
		status := fmt.Sprintf("%v", values[5])
		key := edgeproto.ClusterInstKey{
			ClusterKey:   edgeproto.ClusterKey{Name: cluster},
			Organization: clusterorg,
			CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
		}
		result.Keys = append(result.Keys, &key)
		result.Status = append(result.Status, status)

		if i == 0 {
			result.Timestamp = measurementTime
		} else { // all entries should have the same timestamp, if not equal, we ran through the whole checkpoint and moved onto an older one
			if !result.Timestamp.Equal(measurementTime) {
				break
			}
		}
	}
	return &result, nil
}
