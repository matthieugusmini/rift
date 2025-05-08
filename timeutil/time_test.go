package timeutil_test

import (
	"testing"
	"time"

	"github.com/matthieugusmini/lolesport/timeutil"
)

var (
	worldWarIIStartDate = time.Date(1939, time.September, 1, 0, 0, 0, 0, time.UTC)
	worldWarIIEndDate   = time.Date(1945, time.September, 2, 0, 0, 0, 0, time.UTC)
)

func TestIsCurrentTimeBetween(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		want      bool
	}{
		{
			name:      "returns true if range between yesterday and tomorrow",
			startTime: time.Now().AddDate(0, 0, -1),
			endTime:   time.Now().AddDate(0, 0, 1),
			want:      true,
		},
		{
			name:      "returns false with World War II range",
			startTime: worldWarIIStartDate,
			endTime:   worldWarIIEndDate,
			want:      false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
