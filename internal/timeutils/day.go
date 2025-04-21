package timeutils

import (
	"time"
)

func IsYesterday(date time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return compareDay(date, yesterday)
}

func IsToday(date time.Time) bool {
	today := time.Now()
	return compareDay(date, today)
}

func IsTomorrow(date time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return compareDay(date, tomorrow)
}

func compareDay(date1, date2 time.Time) bool {
	return date1.Year() == date2.Year() &&
		date1.Month() == date2.Month() &&
		date1.Day() == date2.Day()
}
