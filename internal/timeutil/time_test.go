package timeutil_test

import (
	"testing"
	"time"

	"github.com/matthieugusmini/rift/internal/timeutil"
)

var (
	worldWarIIStartDate = time.Date(1939, time.September, 1, 0, 0, 0, 0, time.UTC)
	worldWarIIEndDate   = time.Date(1945, time.September, 2, 0, 0, 0, 0, time.UTC)
)

func TestIsCurrentTimeBetween(t *testing.T) {
	tt := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		want      bool
	}{
		{
			name:      "with range between yesterday and tomorrow returns true",
			startTime: time.Now().AddDate(0, 0, -1),
			endTime:   time.Now().AddDate(0, 0, 1),
			want:      true,
		},
		{
			name:      "with range between World War II start and end date returns false",
			startTime: worldWarIIStartDate,
			endTime:   worldWarIIEndDate,
			want:      false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := timeutil.IsCurrentTimeBetween(tc.startTime, tc.endTime)

			if got != tc.want {
				t.Errorf(
					"IsCurrentTimeBetween(%s, %s) = %t, want %t",
					tc.startTime,
					tc.endTime,
					got,
					tc.want,
				)
			}
		})
	}
}
