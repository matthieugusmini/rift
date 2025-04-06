package lolesport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	baseURL = "https://esports-api.lolesports.com/persisted/gw/"

	headerAPIKey = "X-Api-Key"
	apiKey       = "0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: http.DefaultClient,
	}
}

type GetScheduleOptions struct {
	LeagueIDs []string
}

func (c *Client) GetSchedule(ctx context.Context, opts GetScheduleOptions) (*Schedule, error) {
	params := make(map[string]string)
	for _, leagueID := range opts.LeagueIDs {
		params["leagueId"] = strings.Join([]string{params["leagueID"], leagueID}, ",")
	}
	req, err := newRequest(ctx, "getSchedule", params)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	var responseData struct {
		Data struct {
			Schedule Schedule `json:"schedule"`
		} `json:"data"`
	}
	err = c.doRequest(req, &responseData)
	if err != nil {
		return nil, err
	}
	return &responseData.Data.Schedule, nil
}

func (c *Client) GetStandings(ctx context.Context, tournamentID string) ([]*Standings, error) {
	req, err := newRequest(ctx, "getStandings", map[string]string{"tournamentId": tournamentID})
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	var responseBody struct {
		Data struct {
			Standings []*Standings `json:"standings"`
		} `json:"data"`
	}
	err = c.doRequest(req, &responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody.Data.Standings, nil
}

func (c *Client) GetLeagues(ctx context.Context) ([]*League, error) {
	req, err := newRequest(ctx, "getLeagues", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	var responseBody struct {
		Data struct {
			Leagues []*League `json:"leagues"`
		} `json:"data"`
	}
	err = c.doRequest(req, &responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody.Data.Leagues, nil
}

func (c *Client) GetTournamentsForLeague(ctx context.Context, leagueID string) ([]*Tournament, error) {
	req, err := newRequest(ctx, "getTournamentsForLeague", map[string]string{"leagueId": leagueID})
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	var responseBody struct {
		Data struct {
			Leagues []*League `json:"leagues"`
		} `json:"data"`
	}
	err = c.doRequest(req, &responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody.Data.Leagues[0].Tournaments, nil
}

func (c *Client) doRequest(req *http.Request, response any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("could not decode the response body: %w", err)
	}
	return nil
}

func newRequest(ctx context.Context, endpoint string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set(headerAPIKey, apiKey)
	q := req.URL.Query()
	q.Add("hl", "en-US")
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}
