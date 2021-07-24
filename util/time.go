package util

import "time"

// Get the number of milliseconds since epoch for the time.
// This is used for ElasticSearch.
func GetEpochMillis(t time.Time) int64 {
	return (t.Unix()*1e3 + int64(t.Nanosecond())/1e6)
}

func TimeFromEpochMicros(us int64) time.Time {
	sec := us / 1e6
	ns := (us % 1e6) * 1e3
	return time.Unix(sec, ns)
}
