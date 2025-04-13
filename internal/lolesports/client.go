package lolesports

import (
	"context"

	lolesports "github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/lolesport/internal/timeutils"
)

type Client struct {
	lolesportsClient *lolesports.Client
}

func NewClient(lolesportsClient *lolesports.Client) *Client {
	return &Client{
		lolesportsClient: lolesportsClient,
	}
}

func (c *Client) GetSchedule(ctx context.Context, opts lolesports.GetScheduleOptions) (*lolesports.Schedule, error) {
	return c.lolesportsClient.GetSchedule(ctx, opts)
}

func (c *Client) GetCurrentSeasonSplits(ctx context.Context) ([]*lolesports.Split, error) {
	seasons, err := c.lolesportsClient.GetSeasons(context.Background(), lolesports.GetSeasonsOptions{})
	if err != nil {
		return nil, err
	}
	
	var currentSeason *lolesports.Season
	for _, season := range seasons {
		if season.Name == "lolesports" && timeutils.IsCurrentTimeBetween(season.StartTime, season.EndTime) {
			currentSeason = season
		}
	}

	return currentSeason.Splits, nil
}

func (c *Client) GetStandings(ctx context.Context, tournamentIDs []string) ([]*lolesports.Standings, error) {
	return c.lolesportsClient.GetStandings(ctx, tournamentIDs)
}
