package timeutils

import "time"

func IsCurrentTimeBetween(startTime, endTime time.Time) bool {
	now := time.Now()
	return now.After(startTime) && now.Before(endTime)
}
