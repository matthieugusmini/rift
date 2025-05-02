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
)

type BracketTemplateLoader struct {
	httpClient *http.Client
}

func NewBracketTemplateLoader(httpClient *http.Client) *BracketTemplateLoader {
	return &BracketTemplateLoader{
		httpClient: httpClient,
	}
}

func (l *BracketTemplateLoader) Load(
	ctx context.Context,
	stageID string,
) (rift.BracketTemplate, error) {
	var (
		formatsByStageID    map[string]string
		formatsByStageIDURL = baseURL + "formats_by_stage_id.json"
	)
	if err := l.fetch(ctx, formatsByStageIDURL, &formatsByStageID); err != nil {
		return rift.BracketTemplate{}, err
	}

	format, ok := formatsByStageID[stageID]
	if !ok {
		return rift.BracketTemplate{}, fmt.Errorf("format for stage ID %q not found", stageID)
	}

	var template rift.BracketTemplate
	formatURL := fmt.Sprintf("%s%s.json", baseURL, format)
	if err := l.fetch(ctx, formatURL, &template); err != nil {
		return rift.BracketTemplate{}, err
	}

	return template, nil
}

func (l *BracketTemplateLoader) fetch(ctx context.Context, url string, data any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := l.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(data)
}
