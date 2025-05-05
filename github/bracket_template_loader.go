package github

import (
	"context"
	"log/slog"

	"github.com/matthieugusmini/lolesport/rift"
)

// BracketTemplateLoaderCache represents a cache to retrieve bracket templates.
type BracketTemplateLoaderCache interface {
	// GetBracketTemplate should return:
	// - The rift.BracketTemplate associated to the given stage id that is retrieved from the cache
	// - A boolean indicating whether the value was found or not in the cache
	// - An error(e.g. could not invalidate entry, etc.)
	GetBracketTemplate(stageID string) (rift.BracketTemplate, bool, error)

	// SetBracketTemplate should create a new entry in the cache with a
	// rift.BracketTemplate as value and a stageID as key.
	SetBracketTemplate(stageID string, value rift.BracketTemplate) error
}

// BracketTemplateLoader handles loading bracket templates from multiple sources.
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

// Load tries to load the bracket template associated to the given stage ID
// from the underlying cache first and if not found fetches it using the client.
//
// An error is returned only if the client cannot load the template.
// Errors returned by the cache are not forwarded and are just logged instead.
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
