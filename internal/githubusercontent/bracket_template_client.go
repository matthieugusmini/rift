package githubusercontent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/matthieugusmini/rift/internal/rift"
)

const (
	baseURL = "https://raw.githubusercontent.com/matthieugusmini/lolesports-bracket-templates/refs/heads/main"

	bracketTypeByStageIDFilename = "bracket-type-by-stage-id.json"
)

type BracketTemplateClientOption func(*BracketTemplateClient)

func WithBaseURL(url string) BracketTemplateClientOption {
	return func(c *BracketTemplateClient) {
		c.baseURL = url
	}
}

// BracketTemplateClient handles fetching bracket templates from JSON config files
// stored in a GitHub repository.
//
// Example: https://raw.githubusercontent.com/matthieugusmini/lolesports-bracket-templates/refs/heads/main/8SE.json
type BracketTemplateClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBracketTemplateClient creates a new instance of [BracketTemplateClient].
func NewBracketTemplateClient(
	httpClient *http.Client,
	opts ...BracketTemplateClientOption,
) *BracketTemplateClient {
	c := &BracketTemplateClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetAvailableStageTemplate fetches the list of stage ids
// which have a bracket template associated with them.
//
// An error is returned in case of HTTP error.
func (c *BracketTemplateClient) ListAvailableStageIDs(ctx context.Context) ([]string, error) {
	var bracketTypeByStageID map[string]string
	bracketTypeByStageID, err := c.getBracketTemplateMapper(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the bracket type mapper: %w", err)
	}

	keys := make([]string, 0, len(bracketTypeByStageID))
	for k := range bracketTypeByStageID {
		keys = append(keys, k)
	}

	return keys, nil
}

// GetTemplateByStageID fetches the bracket template associated with the given stage id.
//
// An error is returned if no bracket template is associated with the stage id or
// in case of HTTP error.
func (c *BracketTemplateClient) GetTemplateByStageID(
	ctx context.Context,
	stageID string,
) (rift.BracketTemplate, error) {
	var (
		bracketTypeByStageID map[string]string
		data                 rift.BracketTemplate
	)
	bracketTypeByStageID, err := c.getBracketTemplateMapper(ctx)
	if err != nil {
		return rift.BracketTemplate{}, fmt.Errorf(
			"could not fetch the bracket type mapper: %w",
			err,
		)
	}

	bracketType, ok := bracketTypeByStageID[stageID]
	if !ok {
		return rift.BracketTemplate{}, fmt.Errorf("stage ID %q is not supported", stageID)
	}

	bracketTypeFilename := bracketType + ".json"
	bracketTemplateURL, err := url.JoinPath(c.baseURL, bracketTypeFilename)
	if err != nil {
		return rift.BracketTemplate{}, err
	}

	if err := c.get(ctx, bracketTemplateURL, &data); err != nil {
		return rift.BracketTemplate{}, err
	}

	return data, nil
}

func (c *BracketTemplateClient) getBracketTemplateMapper(
	ctx context.Context,
) (map[string]string, error) {
	bracketTypeByStageIDURL, err := url.JoinPath(c.baseURL, bracketTypeByStageIDFilename)
	if err != nil {
		return nil, err
	}

	var data map[string]string
	if err := c.get(ctx, bracketTypeByStageIDURL, &data); err != nil {
		return map[string]string{}, err
	}

	return data, nil
}

func (c *BracketTemplateClient) get(ctx context.Context, url string, data any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not create new request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(data)
}
