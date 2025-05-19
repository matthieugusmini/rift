package timeutil_test

import (
	"testing"
	"time"

	"github.com/matthieugusmini/rift/internal/timeutil"
	"github.com/stretchr/testify/require"
)

var (
	shacoReleaseDate = time.Date(2009, time.October, 10, 0, 0, 0, 0, time.UTC)
	yuumiReleaseDate = time.Date(2019, time.May, 14, 0, 0, 0, 0, time.UTC)
)

func TestIsCurrentTimeBetween(t *testing.T) {
	tt := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		want      bool
	}{
		{
			name:      "returns true if range from yesterday to tomorrow",
			startTime: time.Now().AddDate(0, 0, -1),
			endTime:   time.Now().AddDate(0, 0, 1),
			want:      true,
		},
		{
			name:      "returns false if range from Shaco release date to Yuumi release date",
			startTime: yuumiReleaseDate,
			endTime:   shacoReleaseDate,
			want:      false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := timeutil.IsCurrentTimeBetween(tc.startTime, tc.endTime)
			require.Equal(t, tc.want, got)
		})
	}
}
