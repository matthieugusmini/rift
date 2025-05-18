package rift

import (
	"context"
	"log/slog"
	"strings"

	"github.com/matthieugusmini/go-lolesports"
)

// LoLEsportsAPIClient represents an API client to retrieve data from LoL Esports.
type LoLEsportsAPIClient interface {
	GetStandings(ctx context.Context, tournamentIDs []string) ([]lolesports.Standings, error)
	GetCurrentSeasonSplits(ctx context.Context) ([]lolesports.Split, error)
	GetSchedule(
		ctx context.Context,
		opts *lolesports.GetScheduleOptions,
	) (lolesports.Schedule, error)
}

// LoLEsportsLoader handles loading LoL Esports data from multiple sources.
type LoLEsportsLoader struct {
	LoLEsportsAPIClient

	standingsCache Cache[[]lolesports.Standings]
	logger         *slog.Logger
}

// NewLoLEsportsLoader creates a new instance of a [Loader] which loads LoLEsports
// data from different caches or apiClient to fetch the data from
// the LoLEsports API.
func NewLoLEsportsLoader(
	apiClient LoLEsportsAPIClient,
	standingsCache Cache[[]lolesports.Standings],
	logger *slog.Logger,
) *LoLEsportsLoader {
	return &LoLEsportsLoader{
		LoLEsportsAPIClient: apiClient,
		standingsCache:      standingsCache,
		logger:              logger,
	}
}

// LoadStandingsByTournamentIDs tries to load all the standings for all the tournamentIDs
// from the underlying cache first and if not found, fetches it using the client.
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

	//nolint:staticcheck
	// The client is embedded to satisfy the ui.LoLEsportsLoader interface
	// for the moment but will changed to a named field later.
	standings, err = l.LoLEsportsAPIClient.GetStandings(ctx, tournamentIDs)
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

func makeStandingsCacheKey(tournamentIDs []string) string {
	return strings.Join(tournamentIDs, ":")
}
