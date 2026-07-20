package rift

import (
	"context"
	"log/slog"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
)

type LoLEsportsStageClient interface {
	GetStage(ctx context.Context, stageID string) (lolesportsgraphql.Stage, error)
}

type LoLEsportsStageLoader struct {
	client LoLEsportsStageClient
	cache  Cache[lolesportsgraphql.Stage]
	logger *slog.Logger
}

func NewLoLEsportsStageLoader(
	client LoLEsportsStageClient,
	cache Cache[lolesportsgraphql.Stage],
	logger *slog.Logger,
) *LoLEsportsStageLoader {
	return &LoLEsportsStageLoader{
		client: client,
		cache:  cache,
		logger: logger.WithGroup("lolesportsStageLoader"),
	}
}

func (l *LoLEsportsStageLoader) GetStage(
	ctx context.Context,
	stageID string,
) (lolesportsgraphql.Stage, error) {
	stage, ok, err := l.cache.Get(stageID)
	if err != nil {
		l.logger.Debug(
			"Stage not present in cache",
			slog.Any("error", err),
			slog.String("stageId", stageID),
		)
	}

	if ok {
		return stage, nil
	}

	stage, err = l.client.GetStage(ctx, stageID)
	if err != nil {
		return lolesportsgraphql.Stage{}, err
	}

	err = l.cache.Set(stageID, stage)
	if err != nil {
		l.logger.Warn(
			"Failed to cache stage",
			slog.Any("error", err),
			slog.String("stageId", stageID),
		)
	}

	return stage, nil
}
