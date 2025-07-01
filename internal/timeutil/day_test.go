package timeutil_test

import (
	"testing"
	"time"

	"github.com/matthieugusmini/rift/internal/timeutil"
	"github.com/stretchr/testify/require"
)

var goatBirthday = time.Date(1996, time.May, 7, 0, 0, 0, 0, time.UTC)

func TestIsYesterday(t *testing.T) {
	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "yesterday returns true",
			date: time.Now().AddDate(0, 0, -1),
			want: true,
		},
		{
			name: "Faker birthday returns false",
			date: goatBirthday,
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := timeutil.IsYesterday(tc.date)

			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsToday(t *testing.T) {
	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "current date returns true",
			date: time.Now(),
			want: true,
		},
		{
			name: "Faker birthday returns false",
			date: goatBirthday,
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := timeutil.IsToday(tc.date)

			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsBeforeToday(t *testing.T) {
	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "yesterday returns true",
			date: time.Now().AddDate(0, 0, -1),
			want: true,
		},
		{
			name: "today returns false",
			date: time.Now(),
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := timeutil.IsBeforeToday(tc.date)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsTomorrow(t *testing.T) {
	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "tomorrow returns true",
			date: time.Now().AddDate(0, 0, 1),
			want: true,
		},
		{
			name: "Faker birthday returns false",
			date: goatBirthday,
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := timeutil.IsTomorrow(tc.date)
			require.Equal(t, tc.want, got)
		})
	}
}
