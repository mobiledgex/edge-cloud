package influxq

import (
	"fmt"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
)

type RetentionPolicyType int

const (
	DefaultRetentionPolicy RetentionPolicyType = iota
	NonDefaultRetentionPolicy
	UnknownRetentionPolicy
)

// Create a retention policy for db
func (q *InfluxQ) CreateRetentionPolicy(retentionTime time.Duration, rpType RetentionPolicyType) error {
	// make sure db is running
	if q.done {
		return fmt.Errorf("retention policy creation failed - %s db finished", q.dbName)
	}

	// make sure db has been created before moving on
	numTries := 3
	for i := 1; i <= numTries; i++ {
		if q.dbcreated {
			break
		}
		if i == numTries {
			return fmt.Errorf("retention policy creation failed - %s db not created yet, ", q.dbName)
		}
		time.Sleep(100 * time.Millisecond)
	}

	shard := time.Duration(24 * time.Hour)
	if retentionTime < shard {
		shard = retentionTime
	}
	// generate retention policy name depending on if default rp or not
	rpName := GetRetentionPolicyName(q.dbName, retentionTime, rpType)

	// create query to create rp
	query := fmt.Sprintf("create retention policy \"%s\" ON \"%s\" duration %s replication 1 shard duration %s", rpName, q.dbName, retentionTime.String(), shard.String())
	if rpType == DefaultRetentionPolicy {
		query += " default"
	}
	_, err := q.QueryDB(query)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.DebugLog(log.DebugLevelMetrics,
				"unable to create default policy", "err", err)
			return err
		}
		// if already exists alter policy instead
		query = fmt.Sprintf("alter retention policy \"%s\" ON \"%s\" duration %s replication 1 shard duration %s", rpName, q.dbName, retentionTime.String(), shard.String())
		if rpType == DefaultRetentionPolicy {
			query += " default"
		}
		if _, err := q.QueryDB(query); err != nil {
			log.DebugLog(log.DebugLevelMetrics,
				"unable to alter policy", "db", q.dbName, "err", err)
			return err
		}
	}
	return nil
}

// Parameters: retention policy name and DbName
var DropRetentionPolicyTemplate = "DROP RETENTION POLICY \"%s\" ON \"%s\""

// Note: "Dropping a retention policy will permanently delete all measurements and data stored in the retention policy." - influxdb docs
func (q *InfluxQ) DropRetentionPolicy(rpName string) error {
	query := fmt.Sprintf(DropRetentionPolicyTemplate, rpName, q.dbName)
	if _, err := q.QueryDB(query); err != nil {
		return err
	}
	return nil
}

func GetRetentionPolicyName(dbName string, retention time.Duration, rpType RetentionPolicyType) string {
	if rpType == DefaultRetentionPolicy || retention == 0 { // if rpType is default or retention is 0, return default rp name
		return fmt.Sprintf("%s_default", dbName)
	} else {
		return fmt.Sprintf("%s_%s", dbName, retention.String())
	}
}
