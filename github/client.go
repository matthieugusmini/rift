package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/matthieugusmini/lolesport/rift"
)

const (
	baseURL = "https://raw.githubusercontent.com/matthieugusmini/lolesports-bracket-templates/refs/heads/main/"

	bracketTypeByStageIDFilename = "bracket-type-by-stage-id.json"
)

// BracketTemplateLoader handles loading bracket templates from JSON config files
// stored in a GitHub repository.
//
// Example: https://raw.githubusercontent.com/matthieugusmini/lolesports-bracket-templates/refs/heads/main/8SE.json
type BracketTemplateLoader struct {
	httpClient *http.Client
}

// NewBracketTemplateLoader creates a new instance of BracketTemplateLoader.
func NewBracketTemplateLoader(httpClient *http.Client) *BracketTemplateLoader {
	return &BracketTemplateLoader{
		httpClient: httpClient,
	}
}

// Load fetches and returns the bracket template for the given stage ID.
func (l *BracketTemplateLoader) Load(
	ctx context.Context,
	stageID string,
) (rift.BracketTemplate, error) {
	var bracketTypeByStageID map[string]string
	bracketTypeByStageIDURL := baseURL + bracketTypeByStageIDFilename
	if err := l.get(ctx, bracketTypeByStageIDURL, &bracketTypeByStageID); err != nil {
		return rift.BracketTemplate{}, err
	}

	bracketType, ok := bracketTypeByStageID[stageID]
	if !ok {
		return rift.BracketTemplate{}, fmt.Errorf("stage ID %q is unsupported", stageID)
	}

	var tmpl rift.BracketTemplate
	bracketTemplateURL := fmt.Sprintf("%s%s.json", baseURL, bracketType)
	if err := l.get(ctx, bracketTemplateURL, &tmpl); err != nil {
		return rift.BracketTemplate{}, err
	}

	return tmpl, nil
}

func (l *BracketTemplateLoader) get(ctx context.Context, url string, data any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not create new request: %w", err)
	}
	resp, err := l.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(data)
}
