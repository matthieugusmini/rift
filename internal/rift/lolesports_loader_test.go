package rift_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/rift/internal/rift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoLEsportsLoader_LoadStandingsByTournamentIDs(t *testing.T) {
	tournamentIDs := []string{"msi-2019", "worlds-2019"}
	cacheKey := "msi-2019:worlds-2019"
	want := testStandings

	t.Run("returns from cache", func(t *testing.T) {
		stubLoLEsportsAPIClient := newStubLoLEsportsAPIClient()
		fakeStandingsCache := newFakeCacheWith(
			map[string][]lolesports.Standings{cacheKey: testStandings},
		)
		fakeSplitsCache := newFakeCache[[]lolesports.Split]()
		loader := rift.NewLoLEsportsLoader(
			stubLoLEsportsAPIClient,
			fakeStandingsCache,
			fakeSplitsCache,
			slog.Default(),
		)

		got, err := loader.LoadStandingsByTournamentIDs(t.Context(), tournamentIDs)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("fetches from API and update cache", func(t *testing.T) {
		stubLoLEsportsAPIClient := newStubLoLEsportsAPIClient()
		fakeStandingsCache := newFakeCache[[]lolesports.Standings]()
		fakeSplitsCache := newFakeCache[[]lolesports.Split]()
		loader := rift.NewLoLEsportsLoader(
			stubLoLEsportsAPIClient,
			fakeStandingsCache,
			fakeSplitsCache,
			slog.Default(),
		)

		got, err := loader.LoadStandingsByTournamentIDs(t.Context(), tournamentIDs)

		require.NoError(t, err)
		assert.Equal(t, want, got)
		// Assert that the cache has been updated
		_, ok := fakeStandingsCache.entries[cacheKey]
		assert.True(t, ok)
	})

	t.Run("returns error if not in cache and API fails", func(t *testing.T) {
		stubLoLEsportsAPIClient := newNotFoundLoLEsportsAPIClient()
		fakeStandingsCache := newFakeCache[[]lolesports.Standings]()
		fakeSplitsCache := newFakeCache[[]lolesports.Split]()
		loader := rift.NewLoLEsportsLoader(
			stubLoLEsportsAPIClient,
			fakeStandingsCache,
			fakeSplitsCache,
			slog.Default(),
		)

		_, err := loader.LoadStandingsByTournamentIDs(t.Context(), tournamentIDs)

		assert.Error(t, err)
	})

	t.Run("fetch from API if fails to get in cache", func(t *testing.T) {
		stubLoLEsportsAPIClient := newStubLoLEsportsAPIClient()
		fakeStandingsCache := newFakeCache[[]lolesports.Standings]()
		fakeStandingsCache.getErr = errCacheGet
		fakeSplitsCache := newFakeCache[[]lolesports.Split]()
		loader := rift.NewLoLEsportsLoader(
			stubLoLEsportsAPIClient,
			fakeStandingsCache,
			fakeSplitsCache,
			slog.Default(),
		)

		got, err := loader.LoadStandingsByTournamentIDs(t.Context(), tournamentIDs)

		require.NoError(t, err)
		assert.Equal(t, want, got)
		// Assert that the cache has been updated
		_, ok := fakeStandingsCache.entries[cacheKey]
		assert.True(t, ok)
	})

	t.Run("returns result if cannot update cache", func(t *testing.T) {
		stubLoLEsportsAPIClient := newStubLoLEsportsAPIClient()
		fakeStandingsCache := newFakeCache[[]lolesports.Standings]()
		fakeStandingsCache.setErr = errCacheSet
		fakeSplitsCache := newFakeCache[[]lolesports.Split]()
		loader := rift.NewLoLEsportsLoader(
			stubLoLEsportsAPIClient,
			fakeStandingsCache,
			fakeSplitsCache,
			slog.Default(),
		)

		got, err := loader.LoadStandingsByTournamentIDs(t.Context(), tournamentIDs)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

var testStandings = []lolesports.Standings{
	{
		Stages: []lolesports.Stage{
			{
				ID:   "EU",
				Name: "will",
				Type: "never",
				Slug: "win",
				Sections: []lolesports.Section{
					{
						Name: "Worlds",
						Matches: []lolesports.Match{
							{
								ID:               "",
								PreviousMatchIDs: nil,
								Flags:            nil,
								Teams: []lolesports.Team{
									{
										ID:    "1337",
										Slug:  "M5",
										Name:  "Moscow 5",
										Code:  "M5",
										Image: "https://le-link-pour-get-l-image.com",
										Result: &lolesports.Result{
											Outcome:  pointer("win"),
											GameWins: 3,
										},
									},
								},
								Strategy: lolesports.Strategy{
									Type:  lolesports.MatchStrategyTypeBestOf,
									Count: 5,
								},
							},
						},
						Rankings: []lolesports.Ranking{
							{},
						},
					},
				},
			},
		},
	},
}

type stubLoLEsportsAPIClient struct {
	standings []lolesports.Standings
	seasons   []lolesports.Season
	err       error
}

func newStubLoLEsportsAPIClient() *stubLoLEsportsAPIClient {
	return &stubLoLEsportsAPIClient{standings: testStandings}
}

func newNotFoundLoLEsportsAPIClient() *stubLoLEsportsAPIClient {
	return &stubLoLEsportsAPIClient{err: errAPINotFound}
}

func (c *stubLoLEsportsAPIClient) GetStandings(
	ctx context.Context,
	tournamentIDs []string,
) ([]lolesports.Standings, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.standings, nil
}

func (c *stubLoLEsportsAPIClient) GetSeasons(
	ctx context.Context,
	opts *lolesports.GetSeasonsOptions,
) ([]lolesports.Season, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.seasons, nil
}

func (c *stubLoLEsportsAPIClient) GetSchedule(
	ctx context.Context,
	opts *lolesports.GetScheduleOptions,
) (lolesports.Schedule, error) {
	return lolesports.Schedule{}, nil
}

func pointer[T any](v T) *T { return &v }
