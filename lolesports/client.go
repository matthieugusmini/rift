package lolesports

import (
	"context"

	lolesports "github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/timeutil"
)

// Client is an adapter of the unofficial lolesports HTTP client.
type Client struct {
	lolesportsClient *lolesports.Client
}

// NewClient returns a newly instanciated Client using the given lolesportsClient under the hood.
func NewClient(lolesportsClient *lolesports.Client) *Client {
	return &Client{
		lolesportsClient: lolesportsClient,
	}
}

// GetSchedule returns a lolesports.Schedule fetched from the lolesports API.
func (c *Client) GetSchedule(
	ctx context.Context,
	opts *lolesports.GetScheduleOptions,
) (lolesports.Schedule, error) {
	return c.lolesportsClient.GetSchedule(ctx, opts)
}

// GetCurrentSeasonSplits returns the list of all the lolesports.Split for the current season.
func (c *Client) GetCurrentSeasonSplits(ctx context.Context) ([]lolesports.Split, error) {
	seasons, err := c.lolesportsClient.GetSeasons(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	var currentSeason lolesports.Season
	for _, season := range seasons {
		if isCurrentSeason(season) {
			currentSeason = season
		}
	}

	return currentSeason.Splits, nil
}

// GetStandings returns a list of all the standings associated with the given tournament ids.
func (c *Client) GetStandings(
	ctx context.Context,
	tournamentIDs []string,
) ([]lolesports.Standings, error) {
	return c.lolesportsClient.GetStandings(ctx, tournamentIDs)
}

func isCurrentSeason(season lolesports.Season) bool {
	return season.Name == "lolesports" &&
		timeutil.IsCurrentTimeBetween(season.StartTime, season.EndTime)
}
