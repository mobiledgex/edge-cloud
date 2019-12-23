package util

import (
	"fmt"
	"time"

	"github.com/cespare/xxhash"
)

// Utility functions for statistics and metrics

// Get the time to wait based on the interval and offset.
// Intervals start at epoch time zero. Returns the time to wait until
// the next interval time. For example, if intervals are every 10 seconds,
// then triggers are at time 10,20,30,40,... If time is now 5, it is 5 seconds
// until trigger 10. If time is now 99, it is 1 second until trigger 100.
//
// The intent is that the stats senders send stats more-or-less around the
// same time, and the aggregator processes those stats at the same interval,
// but some offset later, allowing the aggregator to receive stats from all
// senders for the current interval before processing. This assumes loose time
// synchronization between nodes in the system where the offset can compensate
// for differences in time synchronization.
func GetWaitTime(now time.Time, intervalSec, offsetSec float64) time.Duration {
	// for testing, we allow fractional seconds for offset.
	// do all calculations in millisecond units.
	offsetMsec := int64(offsetSec * 1e3)
	if now.Unix()*1e3 < int64(offsetMsec) {
		// this should never be the case when using real time values.
		panic("offset cannot be greater than the current time")
	}
	alreadyWaitedMs := (now.UnixNano()/1e6 - int64(offsetMsec)) % int64(intervalSec*1e3)
	return time.Duration(int64(intervalSec*1e3)-alreadyWaitedMs) * time.Millisecond
}

// Get the shard index based on the key and number of shards.
func GetShardIndex(key interface{}, numShards uint) uint64 {
	hash := xxhash.Sum64([]byte(fmt.Sprintf("%v", key)))
	return hash % uint64(numShards)
}
