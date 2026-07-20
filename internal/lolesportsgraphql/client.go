package lolesportsgraphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL = "https://lolesports.com/api/gql"

	getSeasonStageOperationName = "GetSeasonStage"
	getSeasonStageOperationID   = "1b5a17d767e7b415b56d2319238468791089bc25cb782138697ae5eefbbea668"

	headerApolloClientName    = "Apollographql-Client-Name"
	headerApolloClientVersion = "Apollographql-Client-Version"
	locale                    = "en-US"
)

var ErrPersistedQueryUnavailable = errors.New("persisted GraphQL query is unavailable")

type ClientOption func(*Client)

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(httpClient *http.Client, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: httpClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) GetStage(ctx context.Context, stageID string) (Stage, error) {
	variables, err := json.Marshal(struct {
		Locale   string   `json:"hl"`
		StageIDs []string `json:"stageIds"`
	}{
		Locale:   locale,
		StageIDs: []string{stageID},
	})
	if err != nil {
		return Stage{}, fmt.Errorf("could not encode variables: %w", err)
	}

	extensions, err := json.Marshal(struct {
		PersistedQuery struct {
			Version int    `json:"version"`
			Hash    string `json:"sha256Hash"`
		} `json:"persistedQuery"`
	}{
		PersistedQuery: struct {
			Version int    `json:"version"`
			Hash    string `json:"sha256Hash"`
		}{
			Version: 1,
			Hash:    getSeasonStageOperationID,
		},
	})
	if err != nil {
		return Stage{}, fmt.Errorf("could not encode extensions: %w", err)
	}

	requestURL, err := url.Parse(c.baseURL)
	if err != nil {
		return Stage{}, fmt.Errorf("could not parse base URL: %w", err)
	}

	query := requestURL.Query()
	query.Set("operationName", getSeasonStageOperationName)
	query.Set("variables", string(variables))
	query.Set("extensions", string(extensions))
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return Stage{}, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Accept", "application/graphql-response+json,application/json;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerApolloClientName, "Rift")
	req.Header.Set(headerApolloClientVersion, "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Stage{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Stage{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data struct {
			Stages []Stage `json:"stages"`
		} `json:"data"`
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return Stage{}, fmt.Errorf("could not decode response body: %w", err)
	}

	if len(response.Errors) > 0 {
		code := response.Errors[0].Extensions.Code
		if code == "PERSISTED_QUERY_NOT_IN_LIST" ||
			code == "PERSISTED_QUERY_NOT_FOUND" ||
			code == "PERSISTED_QUERY_ID_REQUIRED" {
			return Stage{}, fmt.Errorf(
				"%w: %s",
				ErrPersistedQueryUnavailable,
				response.Errors[0].Message,
			)
		}

		return Stage{}, fmt.Errorf("GraphQL request failed: %s", response.Errors[0].Message)
	}

	if len(response.Data.Stages) == 0 {
		return Stage{}, errors.New("stage not found")
	}

	return response.Data.Stages[0], nil
}
