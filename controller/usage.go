package main

import (
	"fmt"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
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
