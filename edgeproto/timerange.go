package edgeproto

import (
	"fmt"
	"time"
)

// Used for search queries to specify a time range
type TimeRange struct {
	// Start time of the time range
	StartTime time.Time `json:"starttime"`
	// End time of the time range
	EndTime time.Time `json:"endtime"`
	// Start age relative to now of the time range
	StartAge Duration `json:"startage"`
	// End age relative to now of the time range
	EndAge Duration `json:"endage"`
}

// Resolve possible arguments to ensure StartTime and EndTime are set.
func (s *TimeRange) Resolve(defaultDuration time.Duration) error {
	now := time.Now()
	startName := "start time"
	endName := "end time"
	if s.StartTime.IsZero() {
		// derive start time from start age
		if s.StartAge == 0 {
			// default duration
			s.StartAge = Duration(defaultDuration)
			// if only end time is specified, use that instead of "now"
			if !s.EndTime.IsZero() {
				now = s.EndTime
			}
		} else {
			startName = "start age"
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
		if s.EndAge != 0 {
			endName = "end age"
		}
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
		return fmt.Errorf("%s must be before (older than) %s", startName, endName)
	}
	return nil
}
