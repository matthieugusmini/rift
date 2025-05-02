package timeutils_test

import (
	"testing"
	"time"

	"github.com/matthieugusmini/lolesport/timeutils"
)

var goatBirthday = time.Date(1996, time.May, 7, 0, 0, 0, 0, time.UTC)

func TestIsYesterday(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "returns true with yesterday date",
			date: time.Now().AddDate(0, 0, -1),
			want: true,
		},
		{
			name: "returns false with Faker birthday",
			date: goatBirthday,
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := timeutils.IsYesterday(tc.date)

			if got != tc.want {
				t.Errorf("IsYesterday(%s) = %t, want %t", tc.date, got, tc.want)
			}
		})
	}
}

func TestIsToday(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "returns true with current date",
			date: time.Now(),
			want: true,
		},
		{
			name: "returns false with faker birthday",
			date: goatBirthday,
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := timeutils.IsToday(tc.date)

			if got != tc.want {
				t.Errorf("IsToday(%s) = %t, want %t", tc.date, got, tc.want)
			}
		})
	}
}

func TestIsTomorrow(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "returns true with tomorrow date",
			date: time.Now().AddDate(0, 0, 1),
			want: true,
		},
		{
			name: "returns false with faker birthday",
			date: goatBirthday,
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := timeutils.IsTomorrow(tc.date)

			if got != tc.want {
				t.Errorf("IsTomorrow(%s) = %t, want %t", tc.date, got, tc.want)
			}
		})
	}
}
