package edgeproto

import (
	"fmt"
	"time"
)

// Used for search queries to specify a time range
type TimeRange struct {
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
	StartAge  Duration  `json:"startage"`
	EndAge    Duration  `json:"endage"`
}

// Resolve possible arguments to ensure StartTime and EndTime are set.
func (s *TimeRange) Resolve(defaultDuration time.Duration) error {
	now := time.Now()
	if s.StartTime.IsZero() {
		// derive start time from start age
		if s.StartAge == 0 {
			// default duration
			s.StartAge = Duration(defaultDuration)
		}
		s.StartTime = now.Add(-1 * s.StartAge.TimeDuration())
		// set age to 0 so function can be idempotent
		s.StartAge = 0
	} else {
		if s.StartAge != 0 {
			return fmt.Errorf("may only specify one of start time or start age")
		}
	}
	if s.EndTime.IsZero() {
		// derive end time from end age
		// default end age of 0 will result in end time of now.
		s.EndTime = now.Add(-1 * s.EndAge.TimeDuration())
		// set age to 0 so function can be idempotent
		s.EndAge = 0
	} else {
		if s.EndAge != 0 {
			return fmt.Errorf("may only specify one of end time or end age")
		}
	}
	if !s.StartTime.Before(s.EndTime) {
		return fmt.Errorf("start time must be before (older than) end time")
	}
	return nil
}
