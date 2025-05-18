package rift_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/rift/internal/rift"
)

func TestLoLEsportsLoader_LoadStandingsByTournamentIDs(t *testing.T) {
	tournamentIDs := []string{"msi-2019", "worlds-2019"}
	cacheKey := "msi-2019:worlds-2019"
	want := testStandings

	t.Run("returns from cache", func(t *testing.T) {
		stubLoLEsportsAPIClient := newStubLoLEsportsAPIClient()
		fakeStandingsCache := newFakeCacheWith(map[string][]lolesports.Standings{
			cacheKey: testStandings,
		})
		fakeSplitsCache := newFakeCache[[]lolesports.Split]()
		loader := rift.NewLoLEsportsLoader(
			stubLoLEsportsAPIClient,
			fakeStandingsCache,
			fakeSplitsCache,
			slog.Default(),
		)

		got := mustLoadStandings(t, loader, tournamentIDs)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"LoLEsportsLoader.LoadStandingsByTournamentIDs(tournamentIDs) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}
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

		got := mustLoadStandings(t, loader, tournamentIDs)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"LoLEsportsLoader.LoadStandingsByTournamentIDs(tournamentIDs) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}

		// Assert that the cache has been updated
		cacheEntry, ok := fakeStandingsCache.entries[cacheKey]
		if !ok {
			t.Fatalf("Bracket template should be cached after loading")
		}
		if diff := cmp.Diff(want, cacheEntry); diff != "" {
			t.Errorf("Cache[stageID] has unexpected diffs(-want +got):\n%s", diff)
		}
	},
	)

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
		if err == nil {
			t.Error("expected an error, got nil")
		}
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

		got := mustLoadStandings(t, loader, tournamentIDs)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"LoLEsportsLoader.LoadStandingsByTournamentIDs(tournamentIDs) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}

		// Assert that the cache has been updated
		cacheEntry, ok := fakeStandingsCache.entries[cacheKey]
		if !ok {
			t.Fatalf("Bracket template should be cached after loading")
		}
		if diff := cmp.Diff(want, cacheEntry); diff != "" {
			t.Errorf("Cache[stageID] has unexpected diffs(-want +got):\n%s", diff)
		}
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

		got := mustLoadStandings(t, loader, tournamentIDs)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"LoLEsportsLoader.LoadStandingsByTournamentIDs(tournamentIDs) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}
	})
}

func mustLoadStandings(
	t *testing.T,
	loader *rift.LoLEsportsLoader,
	tournamentIDs []string,
) []lolesports.Standings {
	t.Helper()

	got, err := loader.LoadStandingsByTournamentIDs(t.Context(), tournamentIDs)
	if err != nil {
		t.Fatalf("got unexpected error %q, want nil", err)
	}

	return got
}

var testStandings = []lolesports.Standings{
	{
		Stages: []lolesports.Stage{
			{
				ID:   "",
				Name: "",
				Type: "",
				Slug: "",
				Sections: []lolesports.Section{
					{
						Name: "",
						Matches: []lolesports.Match{
							{},
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
	return nil, nil
}

func (c *stubLoLEsportsAPIClient) GetSchedule(
	ctx context.Context,
	opts *lolesports.GetScheduleOptions,
) (lolesports.Schedule, error) {
	return lolesports.Schedule{}, nil
}
