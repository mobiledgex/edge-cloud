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

type AppCheckpoint struct {
	Timestamp time.Time
	Org       string
	Keys      []*edgeproto.AppInstKey
	Status    []string // either cloudcommon.InstanceUp or cloudcommon.InstanceDown
}

var AppUsageInfluxQueryTemplate = `SELECT %s from "%s" WHERE "apporg"='%s' AND "app"='%s' AND "ver"='%s' AND "cluster"='%s' AND "clusterorg"='%s' AND "cloudlet"='%s' AND "cloudletorg"='%s' AND time >= '%s' AND time < '%s' order by time desc`

func CreateAppUsageRecord(ctx context.Context, app *edgeproto.AppInst, endTime time.Time, usageEvent cloudcommon.Usage_event) error {
	fmt.Printf("CreateUsage record called...\n")
	var metric *edgeproto.Metric
	// query from the checkpoint up to the event
	selectors := []string{"\"event\"", "\"status\""}
	org := app.Key.AppKey.Organization

	checkpoint, err := GetAppCheckpoint(ctx, org, endTime)
	if err != nil {
		return fmt.Errorf("unable to retrieve Checkpoint: %v", err)
	}
	influxLogQuery := fmt.Sprintf(AppUsageInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.AppInstEvent,
		app.Key.AppKey.Organization,
		app.Key.AppKey.Name,
		app.Key.AppKey.Version,
		app.Key.ClusterInstKey.ClusterKey.Name,
		app.Key.ClusterInstKey.Organization,
		app.Key.ClusterInstKey.CloudletKey.Name,
		app.Key.ClusterInstKey.CloudletKey.Organization,
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
	fmt.Printf("checkpoints: ")
	for i, key := range checkpoint.Keys {
		fmt.Printf(" %+v ", key)
		if key.Matches(&app.Key) {
			checkpointStatus = checkpoint.Status[i]
		}
	}
	fmt.Printf("\n")

	empty, err := checkInfluxQueryOutput(logs, cloudcommon.AppInstEvent)
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
		metric = createAppUsageMetric(app, checkpoint.Timestamp, endTime, totalRunTime, usageEvent, checkpointStatus)
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
		// go until we hit the creation/reservation of this clusterinst, OR until we hit the checkpoint
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
	metric = createAppUsageMetric(app, startTime, endTime, runTime, usageEvent, latestStatus)

	services.events.AddMetric(metric)
	return nil
}

func createAppUsageMetric(app *edgeproto.AppInst, startTime, endTime time.Time, runTime time.Duration, usageEvent cloudcommon.Usage_event, status string) *edgeproto.Metric {
	fmt.Printf("creating record...\n")
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.VMAppInstUsage
	now := time.Now()
	ts, _ := types.TimestampProto(now)
	metric.Timestamp = *ts
	utc, _ := time.LoadLocation("UTC")
	//start and endtimes end up being put into different timezones somehow when going through calculations so force them both to the same here
	startUTC := startTime.In(utc)
	endUTC := endTime.In(utc)

	// influx requires that at least one field must be specified when querying so these cant be all tags
	metric.AddStringVal("cloudletorg", app.Key.ClusterInstKey.CloudletKey.Organization)
	metric.AddTag("cloudlet", app.Key.ClusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", app.Key.ClusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", app.Key.ClusterInstKey.Organization)
	metric.AddTag("org", app.Key.AppKey.Organization)
	metric.AddTag("app", app.Key.AppKey.Name)
	metric.AddTag("ver", app.Key.AppKey.Version)
	metric.AddTag("flavor", app.VmFlavor)
	metric.AddStringVal("start", startUTC.Format(time.RFC3339))
	metric.AddStringVal("end", endUTC.Format(time.RFC3339))
	metric.AddDoubleVal("uptime", runTime.Seconds())
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

// This is checkpointing for VM based app BILLING PERIODS ONLY, custom checkpointing, coming later, will not create(or at least store) usage records upon checkpointing
func CreateAppCheckpoint(ctx context.Context, timestamp time.Time) error {
	fmt.Printf("creating checkpoint....\n")
	now := time.Now()
	if timestamp.After(now) { // we dont know if there will be more creates and deletes before the timestamp occurs
		return fmt.Errorf("Cannot create a checkpoint for the future, checkpointTimestamp: %s, now:%s", timestamp.Format(time.RFC3339), now.Format(time.RFC3339))
	}
	defer services.events.DoPush() // flush these right away for subsequent calls to GetClusterCheckpoint
	// get all running clusterinsts and create a usage record of them

	selectors := []string{"\"app\"", "\"apporg\"", "\"ver\"", "\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"event\""}
	influxLogQuery := fmt.Sprintf(CreateCheckpointInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.AppInstEvent,
		PrevCheckpoint.Format(time.RFC3339),
		timestamp.Format(time.RFC3339))
	logs, err := services.events.QueryDB(influxLogQuery)
	fmt.Printf("influx query: %s\n", influxLogQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}

	empty, err := checkInfluxQueryOutput(logs, cloudcommon.AppInstEvent)
	skipLogCheck := false
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	} else if empty {
		//there are no logs between endTime and the checkpoint, just copy over the checkpoint
		skipLogCheck = true
		fmt.Printf("empty results\n")
	}
	fmt.Printf("results: %+v\n", logs)

	seenApps := make(map[edgeproto.AppInstKey]bool)
	if !skipLogCheck {
		for _, values := range logs[0].Series[0].Values {
			// value should be of the format [timestamp app apporg ver cluster clusterorg cloudlet cloudletorg event]
			if len(values) != len(selectors)+1 {
				return fmt.Errorf("Error parsing influx response")
			}
			app := fmt.Sprintf("%v", values[1])
			apporg := fmt.Sprintf("%v", values[2])
			ver := fmt.Sprintf("%v", values[3])
			cluster := fmt.Sprintf("%v", values[4])
			clusterorg := fmt.Sprintf("%v", values[5])
			cloudlet := fmt.Sprintf("%v", values[6])
			cloudletorg := fmt.Sprintf("%v", values[7])
			event := cloudcommon.InstanceEvent(fmt.Sprintf("%v", values[8]))
			key := edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Organization: apporg,
					Name:         app,
					Version:      ver,
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					ClusterKey:   edgeproto.ClusterKey{Name: cluster},
					Organization: clusterorg,
					CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
				},
			}
			// only care about each clusterinsts most recent log
			if _, exists := seenApps[key]; exists {
				continue
			}
			seenApps[key] = true
			// if its still up, record it
			if event != cloudcommon.DELETED {
				info := edgeproto.AppInst{}
				if !appInstApi.cache.Get(&key, &info) {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Could not find appinst even though event log indicates it is up", "app", key)
					continue
				}
				//record the usage up to this point
				err = CreateAppUsageRecord(ctx, &info, timestamp, cloudcommon.USAGE_EVENT_CHECKPOINT)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create app usage record of checkpointed app", "app", key, "err", err)
				}
			}
		}
	}

	// check for apps that got checkpointed but did not have any log events between PrevCheckpoint and this one
	selectors = []string{"\"app\"", "\"org\"", "\"ver\"", "\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\""}
	influxCheckpointQuery := fmt.Sprintf(CreateCheckpointInfluxUsageQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.VMAppInstUsage,
		PrevCheckpoint.Add(-1*time.Minute).Format(time.RFC3339), //small delta to account for conversion rounding inconsistencies
		PrevCheckpoint.Add(time.Minute).Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return fmt.Errorf("Unable to query influx: %v", err)
	}

	empty, err = checkInfluxQueryOutput(checkpoints, cloudcommon.VMAppInstUsage)
	if err != nil {
		return err
	} else if empty {
		// no checkpoints made yet, or nothing got checkpointed last time, dont need to do this check
		return nil
	}

	for _, values := range checkpoints[0].Series[0].Values {
		// value should be of the format [timestamp app apporg ver cluster clusterorg cloudlet cloudletorg]
		if len(values) != len(selectors)+1 {
			return fmt.Errorf("Error parsing influx response")
		}
		app := fmt.Sprintf("%v", values[1])
		org := fmt.Sprintf("%v", values[2])
		ver := fmt.Sprintf("%v", values[3])
		cluster := fmt.Sprintf("%v", values[4])
		clusterorg := fmt.Sprintf("%v", values[5])
		cloudlet := fmt.Sprintf("%v", values[6])
		cloudletorg := fmt.Sprintf("%v", values[7])
		key := edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Organization: org,
				Name:         app,
				Version:      ver,
			},
			ClusterInstKey: edgeproto.ClusterInstKey{
				ClusterKey:   edgeproto.ClusterKey{Name: cluster},
				Organization: clusterorg,
				CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
			},
		}
		// only care about each clusterinsts most recent log
		if _, exists := seenApps[key]; exists {
			continue
		}
		seenApps[key] = true
		// record it
		info := edgeproto.AppInst{}
		if !appInstApi.cache.Get(&key, &info) {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Could not find appinst even though event log indicates it is up", "app", key)
			continue
		}
		err = CreateAppUsageRecord(ctx, &info, timestamp, cloudcommon.USAGE_EVENT_CHECKPOINT)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to create app usage record of checkpointed app", "app", key, "err", err)
		}
	}

	return nil
}

// returns all the checkpointed clusterinsts of the most recent checkpoint with regards to timestamp
func GetAppCheckpoint(ctx context.Context, org string, timestamp time.Time) (*AppCheckpoint, error) {
	fmt.Printf("calling getAppCheckpoint...\n")
	// wait until the current checkpoint is done if we want to access it, to prevent race conditions with CreateCheckpoint
	for timestamp.After(NextCheckpoint) {
		time.Sleep(time.Second)
	}
	// query from the checkpoint up to the delete
	selectors := []string{"\"app\"", "\"ver\"", "\"cluster\"", "\"clusterorg\"", "\"cloudlet\"", "\"cloudletorg\"", "\"status\"", "\"end\""}
	influxCheckpointQuery := fmt.Sprintf(GetCheckpointInfluxQueryTemplate,
		strings.Join(selectors, ","),
		cloudcommon.VMAppInstUsage,
		org,
		timestamp.Format(time.RFC3339))
	checkpoints, err := services.events.QueryDB(influxCheckpointQuery)
	if err != nil {
		return nil, fmt.Errorf("Unable to query influx: %v", err)
	}
	result := AppCheckpoint{
		Timestamp: PrevCheckpoint,
		Org:       org,
		Keys:      make([]*edgeproto.AppInstKey, 0),
		Status:    make([]string, 0),
	}

	defer fmt.Printf("result: %+v\n", result)

	empty, err := checkInfluxQueryOutput(checkpoints, cloudcommon.VMAppInstUsage)
	if err != nil {
		return nil, err
	} else if empty {
		return &result, nil
	}

	for i, values := range checkpoints[0].Series[0].Values {
		// value should be of the format [timestamp app version cluster clusterorg cloudlet cloudletorg status]
		if len(values) != len(selectors)+1 {
			return nil, fmt.Errorf("Error parsing influx response")
		}
		appname := fmt.Sprintf("%v", values[1])
		version := fmt.Sprintf("%v", values[2])
		cluster := fmt.Sprintf("%v", values[3])
		clusterorg := fmt.Sprintf("%v", values[4])
		cloudlet := fmt.Sprintf("%v", values[5])
		cloudletorg := fmt.Sprintf("%v", values[6])
		status := fmt.Sprintf("%v", values[7])
		key := edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Organization: org,
				Name:         appname,
				Version:      version,
			},
			ClusterInstKey: edgeproto.ClusterInstKey{
				ClusterKey:   edgeproto.ClusterKey{Name: cluster},
				Organization: clusterorg,
				CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
			},
		}
		result.Keys = append(result.Keys, &key)
		result.Status = append(result.Status, status)

		measurementTime, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[8]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse timestamp of checkpoint: %v", err)
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
