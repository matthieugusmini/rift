package timeutils

import "time"

// IsCurrentTimeBetween returns true if the current time is later than startTime but earlier than endTime, false otherwise.
func IsCurrentTimeBetween(startTime, endTime time.Time) bool {
	now := time.Now()
	return now.After(startTime) && now.Before(endTime)
}
