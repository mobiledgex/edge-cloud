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

// earliest possible timestamp influx can handle (64-bit int min in time form)
var InfluxMinimumTimestamp, _ = time.Parse(time.RFC3339, "1677-09-21T00:13:44Z")
var PrevCheckpoint = InfluxMinimumTimestamp
var NextCheckpoint time.Time

func InitUsage() error {
	// set the first NextCheckpoint,
	NextCheckpoint = time.Now().Truncate(time.Minute).Add(*checkpointInterval)
	//set PrevCheckpoint, should not necessarily start at InfluxMinimumTimestamp if controller was restarted halway through operation
	influxQuery := fmt.Sprintf(`SELECT * from "%s" order by time desc limit 1`, cloudcommon.ClusterInstCheckpoint)
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
		checkpoint[0].Series[0].Name != cloudcommon.ClusterInstCheckpoint {
		// should only be 1 series, the 'clusterinst' one
		return fmt.Errorf("Error parsing influx response")
	}
	//we don't care aobut what the checkpoint actually is, just the timestamp of it
	PrevCheckpoint, err = time.Parse(time.RFC3339, fmt.Sprintf("%v", checkpoint[0].Series[0].Values[0][0]))
	if err != nil {
		PrevCheckpoint = InfluxMinimumTimestamp
		return fmt.Errorf("Error creating parsing checkpoint time: %v", err)
	}
	return nil
}

func CreateClusterUsageRecord(ctx context.Context, cluster *edgeproto.ClusterInst, endTime time.Time) error {
	var metric *edgeproto.Metric
	// query from the checkpoint up to the delete
	influxLogQueryTemplate := `SELECT %s from "%s" WHERE "clusterorg"='%s' AND "cluster"='%s' AND "cloudlet"='%s' AND "cloudletorg"='%s' %sAND time >= '%s' AND time < '%s' order by time desc`
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
	influxLogQuery := fmt.Sprintf(influxLogQueryTemplate,
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
	if len(logs) == 0 || len(logs[0].Series) == 0 {
		//there are no logs between endTime and the checkpoint
		var totalRunTime time.Duration
		if checkpointStatus == cloudcommon.InstanceDown {
			totalRunTime = time.Duration(0)
		} else {
			totalRunTime = endTime.Sub(checkpoint.Timestamp)
		}
		metric = createClusterUsageMetric(cluster, checkpoint.Timestamp, endTime, totalRunTime)
		services.events.AddMetric(metric)
		return nil
	} else if len(logs) != 1 || len(logs[0].Series) != 1 || logs[0].Series[0].Name != cloudcommon.ClusterInstEvent { // should only be 1 series, the 'clusterinst' one
		return fmt.Errorf("Error parsing influx response, too many series")
	}
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
		// go until we hit the creation/reservation of this appinst, OR until we hit the checkpoint
		if timestamp.Before(checkpoint.Timestamp) {
			// Check to see if it was up or down at the checkpoint and set the downTime accordingly and then calculate the runTime
			if checkpointStatus == cloudcommon.InstanceDown {
				downTime = downTime + downTimeUpper.Sub(checkpoint.Timestamp)
			}
			runTime = endTime.Sub(checkpoint.Timestamp) - downTime
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
	metric = createClusterUsageMetric(cluster, startTime, endTime, runTime)

	services.events.AddMetric(metric)
	return nil
}

func createClusterUsageMetric(cluster *edgeproto.ClusterInst, startTime, endTime time.Time, runTime time.Duration) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.ClusterInstUsage
	now := time.Now()
	ts, _ := types.TimestampProto(now)
	metric.Timestamp = *ts
	metric.AddTag("cloudletorg", cluster.Key.CloudletKey.Organization)
	metric.AddTag("cloudlet", cluster.Key.CloudletKey.Name)
	metric.AddTag("cluster", cluster.Key.ClusterKey.Name)
	metric.AddTag("clusterorg", cluster.Key.Organization)
	metric.AddTag("flavor", cluster.Flavor.Name)
	metric.AddStringVal("start", startTime.Format(time.RFC3339))
	metric.AddStringVal("end", endTime.Format(time.RFC3339))
	metric.AddDoubleVal("uptime", runTime.Seconds())
	if cluster.ReservedBy != "" && cluster.Key.Organization == cloudcommon.OrganizationMobiledgeX {
		metric.AddTag("org", cluster.ReservedBy)
	} else {
		metric.AddTag("org", cluster.Key.Organization)
	}
	return &metric
}

// This is checkpointing for BILLING PERIODS ONLY, custom checkpointing, coming later, will not create(or at least store) usage records upon checkpointing
func CreateClusterCheckpoint(ctx context.Context, timestamp time.Time) error {
	skipLogCheck := false
	if timestamp.After(time.Now()) { // we dont know if there will be more creates and deletes before the timestamp occurs
		return fmt.Errorf("Cannot create a checkpoint for the future")
	}
	// query from the previous checkpoint to this one
	influxLogQueryTemplate := `SELECT %s from "%s" WHERE time >= '%s' AND time < '%s' order by time desc`
	selectors := []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"event\"", "\"status\""}
	influxLogQuery := fmt.Sprintf(influxLogQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstEvent,
		PrevCheckpoint.Format(time.RFC3339),
		timestamp.Format(time.RFC3339))
	logs, err := services.events.QueryDB(influxLogQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}
	if len(logs) == 0 || len(logs[0].Series) == 0 {
		//there are no logs between endTime and the checkpoint, just copy over the checkpoint
		skipLogCheck = true
	} else if len(logs) != 1 || len(logs[0].Series) != 1 || logs[0].Series[0].Name != cloudcommon.ClusterInstEvent {
		// should only be 1 series, the 'clusterinst' one
		return fmt.Errorf("Error parsing influx response, too many series")
	}
	seenClusters := make(map[edgeproto.ClusterInstKey]bool)
	metrics := make([]*edgeproto.Metric, 0)
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
			status := fmt.Sprintf("%v", values[6])
			key := edgeproto.ClusterInstKey{
				ClusterKey:   edgeproto.ClusterKey{Name: cluster},
				Organization: clusterorg,
				CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
			}
			// only care about each clusterinsts most recent log
			if _, exists := seenClusters[key]; exists {
				continue
			}
			// if its still up, record it
			if event != cloudcommon.DELETED && event != cloudcommon.UNRESERVED {
				info := edgeproto.ClusterInst{}
				if !clusterInstApi.cache.Get(&key, &info) {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Could not find clusterinst even though event log indicates it is up", "cluster", key)
					continue
				}
				org := clusterorg
				if clusterorg == cloudcommon.OrganizationMobiledgeX && info.ReservedBy != "" {
					org = info.ReservedBy
				}

				metric := edgeproto.Metric{}
				metric.Name = cloudcommon.ClusterInstCheckpoint
				ts, _ := types.TimestampProto(timestamp)
				metric.Timestamp = *ts
				metric.AddTag("cloudletorg", cloudletorg)
				metric.AddTag("cloudlet", cloudlet)
				metric.AddTag("cluster", cluster)
				metric.AddTag("clusterorg", clusterorg)
				metric.AddTag("org", org)
				metric.AddStringVal("status", status)
				metrics = append(metrics, &metric)

				//record the usage up to this point
				err = CreateClusterUsageRecord(ctx, &info, timestamp)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create cluster usage record of checkpointed cluster", "cluster", key, "err", err)
				}

				seenClusters[key] = true
			}
		}
	}

	// check for apps that got checkpointed but did not have any log events between PrevCheckpoint and this one
	influxCheckpointQueryTemplate := `SELECT %s from "%s" WHERE time <= '%s' AND time >= '%s' order by time desc`
	selectors = []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"org\"", "\"status\""}
	influxCheckpointQuery := fmt.Sprintf(influxCheckpointQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstCheckpoint,
		PrevCheckpoint.Add(time.Minute).Format(time.RFC3339), //small delta to account for conversion rounding inconsistencies
		PrevCheckpoint.Add(-1*time.Minute).Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}

	if len(checkpoints) == 0 || len(checkpoints[0].Series) == 0 {
		// no checkpoints made yet, or nothing got checkpointed last time, dont need to do this check
		services.events.AddMetric(metrics...)
		services.events.DoPush() // flush these right away for subsequent calls to GetClusterCheckpoint
		PrevCheckpoint = timestamp
		return nil
		// should only be 1 series, the 'clusterinst-checkpoints' one
	} else if len(checkpoints) != 1 || len(checkpoints[0].Series) != 1 || checkpoints[0].Series[0].Name != cloudcommon.ClusterInstCheckpoint {
		return fmt.Errorf("Error parsing influx response, too many series")
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
		org := fmt.Sprintf("%v", values[5])
		status := fmt.Sprintf("%v", values[6])
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
		org = clusterorg
		if clusterorg == cloudcommon.OrganizationMobiledgeX && info.ReservedBy != "" {
			org = info.ReservedBy
		}

		metric := edgeproto.Metric{}
		metric.Name = cloudcommon.ClusterInstCheckpoint
		ts, _ := types.TimestampProto(timestamp)
		metric.Timestamp = *ts
		metric.AddTag("cloudletorg", cloudletorg)
		metric.AddTag("cloudlet", cloudlet)
		metric.AddTag("cluster", cluster)
		metric.AddTag("clusterorg", clusterorg)
		metric.AddTag("org", org)
		metric.AddStringVal("status", status)
		metrics = append(metrics, &metric)

		//record the usage up to this point
		err = CreateClusterUsageRecord(ctx, &info, timestamp)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create cluster usage record of checkpointed cluster", "cluster", key, "err", err)
		}
	}

	services.events.AddMetric(metrics...)
	services.events.DoPush() // flush these right away for subsequent calls to GetClusterCheckpoint
	PrevCheckpoint = timestamp
	return nil
}

// There is a race condition here between GetClusterCheckpoint and CreateClusterCheckpoint,
// if someone calls GetClusterCheckpoint in between the timestamp arg to CreateClusterCheckpoint and the actual call to CreateClusterCheckpoint,
// it will end up referencing the old checkpoint and not the latest one
// So we need to prevent calls to GetClusterCheckpoint with a timestamp after the timestamp passed in CreateClusterCheckpoint until CreateClusterCheckpoint
// actually gets run and finishes its call with that checkpoint.

//ie. timeline: checkpoint 1 time -------------- checkpoint 2 time --- timestampX ----A----B--- ,
// GetClusterCheckpoint gets called with timestampX at time A, and CreateClusterCheckpoint with timestamp=checkpoint2Time doesnt go off until time B (this window should be very small)
// GetClusterCheckpoint will return checkpoint 1 instead of checkpoint 2 and could theoretically result in a usage record greater than the length of a billing period
// as well as usage times being double counted by CreateClusterCheckpoints usage record generation

// returns all the clusterinsts that were running at the time belonging to that org
func GetClusterCheckpoint(ctx context.Context, org string, timestamp time.Time) (*ClusterCheckpoint, error) {
	// wait until the current checkpoint is done if we want to access it, see the above comment about race conditions
	for timestamp.After(NextCheckpoint) {
		time.Sleep(time.Second)
	}
	// query from the checkpoint up to the delete
	influxCheckpointQueryTemplate := `SELECT %s from "%s" WHERE "org"='%s' AND time <= '%s' order by time desc`
	selectors := []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"status\""}
	influxCheckpointQuery := fmt.Sprintf(influxCheckpointQueryTemplate, strings.Join(selectors, ","), cloudcommon.ClusterInstCheckpoint, org, timestamp.Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return nil, fmt.Errorf("Unable to query influx: %v", err)
	}
	result := ClusterCheckpoint{
		Timestamp: InfluxMinimumTimestamp,
		Org:       org,
		Keys:      make([]*edgeproto.ClusterInstKey, 0),
		Status:    make([]string, 0),
	}

	if len(checkpoints) == 0 || len(checkpoints[0].Series) == 0 {
		// no checkpoints made yet
		return &result, nil
		// should only be 1 series, the 'clusterinst-checkpoints' one
	} else if len(checkpoints) != 1 || len(checkpoints[0].Series) != 1 || checkpoints[0].Series[0].Name != cloudcommon.ClusterInstCheckpoint {
		return nil, fmt.Errorf("Error parsing influx response, too many series")
	}

	for i, values := range checkpoints[0].Series[0].Values {
		// value should be of the format [timestamp cluster clusterorg cloudlet cloudletorg status]
		if len(values) != len(selectors)+1 {
			return nil, fmt.Errorf("Error parsing influx response")
		}
		timestamp, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[0]))
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
			result.Timestamp = timestamp
		} else { // all entries should have the same timestamp, if not equal, we ran through the whole checkpoint and moved onto an older one
			if !result.Timestamp.Equal(timestamp) {
				break
			}
		}
	}
	return &result, nil
}

func runClusterCheckpoints(ctx context.Context) {
	checkpointSpan := log.StartSpan(log.DebugLevelInfo, "Cluster Checkpointing thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
	defer checkpointSpan.Finish()
	err := InitUsage()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error setting up checkpoints", "err", err)
	}
	for {
		if time.Now().After(NextCheckpoint) {
			checkpointTime := NextCheckpoint
			err = CreateClusterCheckpoint(ctx, checkpointTime)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Could not create checkpoint", "time", checkpointTime, "err", err)
			}
			// this must be AFTER the checkpoint is created, see the comments about race conditions above GetClusterCheckpoint
			NextCheckpoint = NextCheckpoint.Add(*checkpointInterval)
		}
	}
}
