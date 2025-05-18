package rift

import (
	"context"
	"log/slog"
)

// BracketTemplateClient represents a client to retrieve bracket templates
// performing I/O (e.g. network).
type BracketTemplateClient interface {
	// GetTemplateByStageID shoulds return the BracketTemplate
	// associated to the given stage id.
	GetTemplateByStageID(ctx context.Context, stageID string) (BracketTemplate, error)

	// ListAvailableStageIDs returns the list of ids of all stages
	// which have an associated bracket template.
	ListAvailableStageIDs(ctx context.Context) ([]string, error)
}

// BracketTemplateLoader handles loading bracket templates from multiple sources.
type BracketTemplateLoader struct {
	client BracketTemplateClient
	cache  Cache[BracketTemplate]
	logger *slog.Logger
}

// NewBracketTemplateLoader creates a new instance of BracketTemplateLoader.
func NewBracketTemplateLoader(
	bracketTemplateClient BracketTemplateClient,
	cache Cache[BracketTemplate],
	logger *slog.Logger,
) *BracketTemplateLoader {
	return &BracketTemplateLoader{
		client: bracketTemplateClient,
		cache:  cache,
		logger: logger.WithGroup("bracketTemplateLoader"),
	}
}

// ListAvailableStageIDs returns the list of available stage ids in the server.
//
// An error is returned if it cannot fetch the data.
func (l *BracketTemplateLoader) ListAvailableStageIDs(ctx context.Context) ([]string, error) {
	stageIDs, err := l.client.ListAvailableStageIDs(ctx)
	if err != nil {
		return nil, err
	}
	return stageIDs, nil
}

// Load tries to load the bracket template associated to the given stage ID
// from the underlying cache first and if not found fetches it using the client.
//
// An error is returned only if the client cannot load the template.
// Errors returned by the cache are not forwarded and are just logged instead.
func (l *BracketTemplateLoader) Load(
	ctx context.Context,
	stageID string,
) (BracketTemplate, error) {
	var tmpl BracketTemplate

	tmpl, ok, err := l.cache.Get(stageID)
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
		return BracketTemplate{}, err
	}

	if err := l.cache.Set(stageID, tmpl); err != nil {
		l.logger.Warn(
			"Failed to cache bracket template",
			slog.Any("err", err),
			slog.String("stageId", stageID),
		)
	}

	return tmpl, nil
}
