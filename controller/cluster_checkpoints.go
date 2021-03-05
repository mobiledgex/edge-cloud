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
)

type ClusterCheckpoint struct {
	Timestamp time.Time
	Org       string
	Keys      []*edgeproto.ClusterInstKey
	Status    []string // either cloudcommon.InstanceUp or cloudcommon.InstanceDown
}

var ClusterUsageInfluxQueryTemplate = `SELECT %s from "%s" WHERE "clusterorg"='%s' AND "cluster"='%s' AND "cloudlet"='%s' AND "cloudletorg"='%s' %sAND time >= '%s' AND time < '%s' order by time desc`

func CreateClusterUsageRecord(ctx context.Context, cluster *edgeproto.ClusterInst, endTime time.Time) error {
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
	fmt.Printf("got checkpoint: %+v\n", checkpoint)
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

	stats := RunTimeStats{
		end: endTime,
	}
	err = GetRunTimeStats(usageTypeCluster, *checkpoint, cluster.Key, logs, &stats)
	if err != nil {
		return err
	}

	// write the usage record to influx
	metric = createClusterUsageMetric(cluster, stats.start, stats.end, stats.upTime, stats.status)

	services.events.AddMetric(metric)
	return nil
}

func createClusterUsageMetric(cluster *edgeproto.ClusterInst, startTime, endTime time.Time, runTime time.Duration, status string) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.ClusterInstCheckpoints
	ts, _ := types.TimestampProto(endTime)
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
	metric.AddStringVal("flavor", cluster.Flavor.Name)
	metric.AddIntVal("nodecount", uint64(cluster.NumMasters+cluster.NumNodes))
	metric.AddStringVal("ipaccess", cluster.IpAccess.String())
	metric.AddStringVal("start", startUTC.Format(time.RFC3339))
	metric.AddStringVal("end", endUTC.Format(time.RFC3339))
	metric.AddDoubleVal("uptime", runTime.Seconds())
	if cluster.ReservedBy != "" && cluster.Key.Organization == cloudcommon.OrganizationMobiledgeX {
		metric.AddTag("org", cluster.ReservedBy)
	} else {
		metric.AddTag("org", cluster.Key.Organization)
	}
	metric.AddStringVal("status", status)
	return &metric
}

// This is checkpointing for the usage api, from month to month
func CreateClusterCheckpoint(ctx context.Context, timestamp time.Time) error {
	if err := checkpointTimeValid(timestamp); err != nil { // we dont know if there will be more creates and deletes before the timestamp occurs
		return err
	}
	defer services.events.DoPush() // flush these right away for subsequent calls to GetClusterCheckpoint
	// get all running clusterinsts and create a usage record of them
	selectors := []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"event\""}
	influxLogQuery := fmt.Sprintf(CreateCheckpointInfluxQueryTemplate,
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
			// value should be of the format [timestamp cluster clusterorg cloudlet cloudletorg event]
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
				err = CreateClusterUsageRecord(ctx, &info, timestamp)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create cluster usage record of checkpointed cluster", "cluster", key, "err", err)
				}
			}
		}
	}

	// check for clusters that got checkpointed but did not have any log events between PrevCheckpoint and this one
	selectors = []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\""}
	influxCheckpointQuery := fmt.Sprintf(CreateCheckpointInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstCheckpoints,
		PrevCheckpoint.Add(-1*time.Minute).Format(time.RFC3339), //small delta to account for conversion rounding inconsistencies
		PrevCheckpoint.Add(time.Minute).Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}

	empty, err = checkInfluxQueryOutput(checkpoints, cloudcommon.ClusterInstCheckpoints)
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
		err = CreateClusterUsageRecord(ctx, &info, timestamp)
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
	selectors := []string{"\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"status\"", "\"end\""}
	influxCheckpointQuery := fmt.Sprintf(GetCheckpointInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.ClusterInstCheckpoints,
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

	empty, err := checkInfluxQueryOutput(checkpoints, cloudcommon.ClusterInstCheckpoints)
	if err != nil {
		return nil, err
	} else if empty {
		return &result, nil
	}

	for i, values := range checkpoints[0].Series[0].Values {
		// value should be of the format [measurementTime cluster clusterorg cloudlet cloudletorg status end]
		if len(values) != len(selectors)+1 {
			return nil, fmt.Errorf("Error parsing influx response")
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

		measurementTime, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[6]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse timestamp of checkpoint")
		}

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
