package lolesports

import (
	"context"
	"log/slog"
	"strings"

	"github.com/matthieugusmini/go-lolesports"
)

// Cache represents a key/value store with fast read access.
type Cache[T any] interface {
	Get(key string) (T, bool, error)
	Set(key string, value T) error
}

// Loader handles loading LoLEsports data from multiple sources.
type Loader struct {
	*Client

	standingsCache Cache[[]lolesports.Standings]
	logger         *slog.Logger
}

// NewLoader creates a new instance of a Loader which loads LoLEsports
// data from different caches or apiClient to fetch the data from
// the LoLEsports API.
func NewLoader(
	apiClient *Client,
	standingsCache Cache[[]lolesports.Standings],
	logger *slog.Logger,
) *Loader {
	return &Loader{
		Client:         apiClient,
		standingsCache: standingsCache,
		logger:         logger,
	}
}

// LoadStandingsByTournamentIDs tries to load all the standings of all the tournamentIDs
// from the underlying cache first and if not found fetches it using the client.
//
// An error is returned only if the client cannot load the standings.
// Errors returned by the cache are not forwarded and are just logged instead.
func (l *Loader) LoadStandingsByTournamentIDs(
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

	//nolint:staticcheck // The client is embedded to satisfy the ui.LoLEsportsLoader interface
	// for the moment but will changed to a named field later.
	standings, err = l.Client.GetStandings(ctx, tournamentIDs)
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
