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
		time.Sleep(time.Second)
	}

	shard := time.Duration(24 * time.Hour)
	if retentionTime < shard {
		shard = retentionTime
	}
	// generate retention policy name depending on if default rp or not
	var rpName string
	if rpType == DefaultRetentionPolicy {
		rpName = getDefaultRetentionPolicyName(q.dbName)
	} else {
		rpName = getNonDefaultRetentionPolicyName(q.dbName, retentionTime)
	}

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
	q.retentionPolicies[rpName] = struct{}{}
	return nil
}

// Parameters: retention policy name and DbName
var DropRetentionPolicyTemplate = "DROP RETENTION POLICY \"%s\" ON \"%s\""

// TODO: "Dropping a retention policy will permanently delete all measurements and data stored in the retention policy." - influxdb docs
func (q *InfluxQ) DropRetentionPolicy(rpName string) error {
	query := fmt.Sprintf(DropRetentionPolicyTemplate, rpName, q.dbName)
	if _, err := q.QueryDB(query); err != nil {
		return err
	}
	delete(q.retentionPolicies, rpName)
	return nil
}

func getNonDefaultRetentionPolicyName(dbName string, rp time.Duration) string {
	return fmt.Sprintf("%s_%s", dbName, rp.String())
}

func getDefaultRetentionPolicyName(dbName string) string {
	return fmt.Sprintf("%s_default", dbName)
}
