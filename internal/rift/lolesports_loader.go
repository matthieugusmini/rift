package rift

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/rift/internal/timeutil"
)

const currentSeasonSplitsCacheKey = "current_splits"

// LoLEsportsAPIClient represents an API client to retrieve data from LoL Esports.
type LoLEsportsAPIClient interface {
	GetStandings(ctx context.Context, tournamentIDs []string) ([]lolesports.Standings, error)
	GetSeasons(ctx context.Context, opts *lolesports.GetSeasonsOptions) ([]lolesports.Season, error)
	GetSchedule(
		ctx context.Context,
		opts *lolesports.GetScheduleOptions,
	) (lolesports.Schedule, error)
}

// LoLEsportsLoader handles loading LoL Esports data from multiple sources.
type LoLEsportsLoader struct {
	apiClient      LoLEsportsAPIClient
	standingsCache Cache[[]lolesports.Standings]
	splitsCache    Cache[[]lolesports.Split]
	logger         *slog.Logger
}

// NewLoLEsportsLoader creates a new instance of a [Loader] which loads LoLEsports
// data from different caches or apiClient to fetch the data from
// the LoLEsports API.
func NewLoLEsportsLoader(
	apiClient LoLEsportsAPIClient,
	standingsCache Cache[[]lolesports.Standings],
	splitsCache Cache[[]lolesports.Split],
	logger *slog.Logger,
) *LoLEsportsLoader {
	return &LoLEsportsLoader{
		apiClient:      apiClient,
		standingsCache: standingsCache,
		splitsCache:    splitsCache,
		logger:         logger,
	}
}

// LoadStandingsByTournamentIDs tries to load all the standings for all the tournamentIDs
// from the underlying cache first and if not found, fetches them from the API.
//
// An error is returned only if the client cannot load the standings.
// Errors returned by the cache are not forwarded and are just logged instead.
func (l *LoLEsportsLoader) LoadStandingsByTournamentIDs(
	ctx context.Context,
	tournamentIDs []string,
) ([]lolesports.Standings, error) {
	var standings []lolesports.Standings

	key := makeStandingsCacheKey(tournamentIDs)
	standings, ok, err := l.standingsCache.Get(key)
	if err != nil {
		l.logger.Debug(
			"Standings not present in cache",
			slog.Any("err", err),
			slog.Any("tournamentIds", tournamentIDs),
		)
	}
	if ok {
		return standings, nil
	}

	standings, err = l.apiClient.GetStandings(ctx, tournamentIDs)
	if err != nil {
		return nil, err
	}

	if err := l.standingsCache.Set(key, standings); err != nil {
		l.logger.Warn(
			"Failed to set standings in cache",
			slog.Any("err", err),
			slog.Any("tournamentIds", tournamentIDs),
		)
	}

	return standings, nil
}

// LoadCurrentSeasonSplits tries to load all the splits for the current season
// from the underlying cache first and if not found, fetches them from the API.
//
// An error is returned only if the client cannot load the standings.
// Errors returned by the cache are not forwarded and are just logged instead.
func (l *LoLEsportsLoader) LoadCurrentSeasonSplits(
	ctx context.Context,
) ([]lolesports.Split, error) {
	splits, ok, err := l.splitsCache.Get(currentSeasonSplitsCacheKey)
	if err != nil {
		l.logger.Debug(
			"Current season splits not present in cache",
			slog.Any("err", err),
		)
	}
	if ok {
		return splits, nil
	}

	seasons, err := l.apiClient.GetSeasons(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch seasons: %w", err)
	}

	var currentSeason lolesports.Season
	for _, season := range seasons {
		if isCurrentSeason(season) {
			currentSeason = season
		}
	}

	if err := l.splitsCache.Set(currentSeasonSplitsCacheKey, currentSeason.Splits); err != nil {
		l.logger.Warn(
			"Failed to set splits in cache",
			slog.Any("err", err),
			slog.Any("splits", currentSeason.Splits),
		)
	}

	return currentSeason.Splits, nil
}

// GetSchedule fetches the schedule from the API.
//
// Optionally options can be passed to fetch specific pages or
// to fetch only events related to certain leagues.
//
// An error is returned if it cannot fetch the data.
func (l *LoLEsportsLoader) GetSchedule(
	ctx context.Context,
	opts *lolesports.GetScheduleOptions,
) (lolesports.Schedule, error) {
	return l.apiClient.GetSchedule(ctx, opts)
}

func makeStandingsCacheKey(tournamentIDs []string) string {
	return strings.Join(tournamentIDs, ":")
}

func isCurrentSeason(season lolesports.Season) bool {
	return season.Name == "lolesports" &&
		timeutil.IsCurrentTimeBetween(season.StartTime, season.EndTime)
}
