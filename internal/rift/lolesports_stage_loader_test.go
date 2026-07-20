package rift_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
	"github.com/matthieugusmini/rift/internal/rift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoLEsportsStageLoader_GetStage(t *testing.T) {
	const stageID = "stage-1"

	want := lolesportsgraphql.Stage{ID: stageID, Name: "Playoffs"}

	t.Run("returns a cached stage", func(t *testing.T) {
		client := &stubLoLEsportsStageClient{}
		stageCache := newFakeCacheWith(map[string]lolesportsgraphql.Stage{stageID: want})
		loader := rift.NewLoLEsportsStageLoader(client, stageCache, slog.Default())

		got, err := loader.GetStage(t.Context(), stageID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Zero(t, client.calls)
	})

	t.Run("fetches and caches a stage", func(t *testing.T) {
		client := &stubLoLEsportsStageClient{stage: want}
		stageCache := newFakeCache[lolesportsgraphql.Stage]()
		loader := rift.NewLoLEsportsStageLoader(client, stageCache, slog.Default())

		got, err := loader.GetStage(t.Context(), stageID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, want, stageCache.entries[stageID])
		assert.Equal(t, 1, client.calls)
	})

	t.Run("returns client errors", func(t *testing.T) {
		clientErr := errors.New("stage API unavailable")
		client := &stubLoLEsportsStageClient{err: clientErr}
		stageCache := newFakeCache[lolesportsgraphql.Stage]()
		loader := rift.NewLoLEsportsStageLoader(client, stageCache, slog.Default())

		_, err := loader.GetStage(t.Context(), stageID)

		require.ErrorIs(t, err, clientErr)
		assert.NotContains(t, stageCache.entries, stageID)
	})

	t.Run("cache failures do not prevent loading", func(t *testing.T) {
		client := &stubLoLEsportsStageClient{stage: want}
		stageCache := newFakeCache[lolesportsgraphql.Stage]()
		stageCache.getErr = errCacheGet
		stageCache.setErr = errCacheSet
		loader := rift.NewLoLEsportsStageLoader(client, stageCache, slog.Default())

		got, err := loader.GetStage(t.Context(), stageID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

type stubLoLEsportsStageClient struct {
	stage lolesportsgraphql.Stage
	err   error
	calls int
}

func (c *stubLoLEsportsStageClient) GetStage(
	_ context.Context,
	_ string,
) (lolesportsgraphql.Stage, error) {
	c.calls++

	return c.stage, c.err
}
