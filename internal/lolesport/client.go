package lolesport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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

type Standings struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	ID       string    `json:"id"`
	Name     string    `json:"name,omitempty"`
	Type     string    `json:"type,omitempty"`
	Slug     string    `json:"slug,omitempty"`
	Sections []Section `json:"sections,omitempty"`
}

type Section struct {
	Name     string    `json:"name"`
	Matches  []Match   `json:"matches"`
	Rankings []Ranking `json:"rankings"`
}
type MatchStrategyType string

const (
	MatchStrategyTypeBestOf = "bestOf"
)

type EventState string

const (
	EventStateUnstarted  = "unstarted"
	EventStateInProgress = "inProgress"
	EventStateCompleted  = "completed"
)

type Match struct {
	ID               string   `json:"id"`
	PreviousMatchIDs []string `json:"previousMatchIds"`
	Flags            []string `json:"flags"`
	Teams            []Team   `json:"teams"`
	Strategy         Strategy `json:"strategy"`
}

type Team struct {
	ID     string  `json:"id,omitempty"`
	Slug   string  `json:"slug,omitempty"`
	Name   string  `json:"name"`
	Code   string  `json:"code"`
	Image  string  `json:"image"`
	Result *Result `json:"result,omitempty"`
	Record *Record `json:"record,omitempty"`
}

type Result struct {
	Outcome  *string `json:"outcome"`
	GameWins int     `json:"gameWins"`
}

type Ranking struct {
	Ordinal int    `json:"ordinal"`
	Teams   []Team `json:"teams"`
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

func (c *Client) GetStandings(ctx context.Context, tournamentID string) ([]Standings, error) {
	req, err := newRequest(ctx, "getStandings", map[string]string{"tournamentId": tournamentID})
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	var responseBody struct {
		Data struct {
			Standings []Standings `json:"schedule"`
		} `json:"data"`
	}
	err = c.doRequest(req, &responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody.Data.Standings, nil
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

// Define the structs to match the JSON structure
type Schedule struct {
	Updated time.Time `json:"updated"`
	Pages   Pages     `json:"pages"`
	Events  []Event   `json:"events"`
}

type Pages struct {
	Older string `json:"older"`
	Newer string `json:"newer"`
}

type EventType string

const (
	EventTypeMatch = "match"
	EventTypeShow  = "show"
)

type Event struct {
	StartTime time.Time  `json:"startTime"`
	BlockName string     `json:"blockName"`
	Match     Match      `json:"match"`
	State     EventState `json:"state"`
	Type      string     `json:"type"`
	League    League     `json:"league"`
}

type Record struct {
	Losses int `json:"losses"`
	Wins   int `json:"wins"`
}

type Strategy struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
}

type League struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
