package github

import (
	"context"
	"log/slog"

	"github.com/matthieugusmini/lolesport/rift"
)

type BracketTemplateLoaderCache interface {
	GetBracketTemplate(key string) (rift.BracketTemplate, bool, error)
	SetBracketTemplate(key string, value rift.BracketTemplate) error
}

// BracketTemplateLoader handles loading bracket templates from JSON config files
// stored in a GitHub repository.
// It also use a cache package to not fetch the files each time the app reload
//
// Example: https://raw.githubusercontent.com/matthieugusmini/lolesports-bracket-templates/refs/heads/main/8SE.json
type BracketTemplateLoader struct {
	client *BracketTemplateClient
	cache  BracketTemplateLoaderCache
	logger *slog.Logger
}

// NewBracketTemplateLoader creates a new instance of BracketTemplateLoader.
func NewBracketTemplateLoader(
	bracketTemplateClient *BracketTemplateClient,
	cache BracketTemplateLoaderCache,
	logger *slog.Logger,
) *BracketTemplateLoader {
	return &BracketTemplateLoader{
		client: bracketTemplateClient,
		cache:  cache,
	}
}

// Load check the cache for the file
// if present, returns the bracket template
// if not present fetches the file, keep it in the cache and returns the bracket template for the given stage ID.
func (l *BracketTemplateLoader) Load(
	ctx context.Context,
	stageID string,
) (rift.BracketTemplate, error) {
	var tmpl rift.BracketTemplate

	tmpl, ok, err := l.cache.GetBracketTemplate(stageID)
	if err != nil {
		l.logger.Debug(
			"Bracket template not present in cache",
			slog.Any("err", err),
			slog.String("stageId", stageID),
		)
	}
	if ok {
		return tmpl, nil
	}

	tmpl, err = l.client.GetTemplateByStageID(ctx, stageID)
	if err != nil {
		return rift.BracketTemplate{}, err
	}

	if err := l.cache.SetBracketTemplate(stageID, tmpl); err != nil {
		l.logger.Debug(
			"Failed to cache bracket template",
			slog.Any("err", err),
			slog.String("stageId", stageID),
		)
	}

	return tmpl, nil
}
