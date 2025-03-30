package dateutils

import "time"

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

func compareDay(date1 time.Time, date2 time.Time) bool {
	truncatedDate1 := date1.Truncate(24 * time.Hour)
	truncatedDate2 := date2.Truncate(24 * time.Hour)
	return truncatedDate1.Equal(truncatedDate2)
}
