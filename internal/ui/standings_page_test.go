package ui //nolint:testpackage // White-box tests exercise Bubble Tea command messages.

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandingsPageLoadBracketStage(t *testing.T) {
	t.Run("returns dynamic stage data", func(t *testing.T) {
		stageClient := &stubStageClient{
			stage: lolesportsgraphql.Stage{ID: "stage-1"},
		}
		page := newStandingsPage(nil, stageClient, slog.Default())

		msg := page.loadBracketStage("stage-1")()

		loaded, ok := msg.(loadedDynamicStageMessage)
		require.True(t, ok)
		assert.Equal(t, "stage-1", loaded.stage.ID)
		assert.Equal(t, 1, stageClient.calls)
	})

	t.Run("returns dynamic stage errors", func(t *testing.T) {
		stageClient := &stubStageClient{err: errors.New("dynamic stage unavailable")}
		page := newStandingsPage(nil, stageClient, slog.Default())

		msg := page.loadBracketStage("stage-1")()

		failed, ok := msg.(fetchErrorMessage)
		require.True(t, ok)
		require.ErrorContains(t, failed.err, "dynamic stage unavailable")
	})
}

func TestStandingsPage_HandleDynamicStageLoaded(t *testing.T) {
	page := newStandingsPage(nil, &stubStageClient{}, slog.Default())
	page.width = 100
	page.height = 40
	stage := lolesportsgraphql.Stage{
		ID: "stage-1",
		Sections: []lolesportsgraphql.Section{
			{
				Columns: []lolesportsgraphql.Column{columnWithMatchCount(1)},
			},
		},
	}

	page.handleDynamicStageLoaded(loadedDynamicStageMessage{stage: stage})

	assert.Equal(t, standingsPageStateShowBracketPage, page.state)
	require.NotNil(t, page.bracket)
	require.NotNil(t, page.bracket.dynamicStage)
	assert.Equal(t, "stage-1", page.bracket.dynamicStage.ID)
}

type stubStageClient struct {
	stage lolesportsgraphql.Stage
	err   error
	calls int
}

func (c *stubStageClient) GetStage(
	_ context.Context,
	_ string,
) (lolesportsgraphql.Stage, error) {
	c.calls++

	return c.stage, c.err
}
