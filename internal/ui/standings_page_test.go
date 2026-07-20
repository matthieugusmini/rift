package ui //nolint:testpackage // White-box tests exercise Bubble Tea command messages.

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

func TestStandingsPage_LoadBracketStage(t *testing.T) {
	t.Run("returns dynamic stage data without loading a template", func(t *testing.T) {
		stageClient := &stubStageClient{
			stage: lolesportsgraphql.Stage{ID: "stage-1"},
		}
		bracketLoader := &stubBracketLoader{}
		page := newStandingsPage(nil, stageClient, bracketLoader, slog.Default())

		msg := page.loadBracketStage("stage-1")()

		loaded, ok := msg.(loadedDynamicStageMessage)
		require.True(t, ok)
		assert.Equal(t, "stage-1", loaded.stage.ID)
		assert.Equal(t, 1, stageClient.calls)
		assert.Zero(t, bracketLoader.calls)
	})

	t.Run("falls back to a bracket template", func(t *testing.T) {
		stageClient := &stubStageClient{err: errors.New("dynamic stage unavailable")}
		bracketLoader := &stubBracketLoader{
			template: rift.BracketTemplate{Rounds: []rift.Round{{Title: "Finals"}}},
		}
		page := newStandingsPage(nil, stageClient, bracketLoader, slog.Default())

		msg := page.loadBracketStage("stage-1")()

		loaded, ok := msg.(loadedBracketStageTemplateMessage)
		require.True(t, ok)
		assert.Equal(t, "Finals", loaded.template.Rounds[0].Title)
		assert.Equal(t, 1, stageClient.calls)
		assert.Equal(t, 1, bracketLoader.calls)
	})

	t.Run("returns both dynamic and fallback errors", func(t *testing.T) {
		stageClient := &stubStageClient{err: errors.New("dynamic stage unavailable")}
		bracketLoader := &stubBracketLoader{err: errors.New("template unavailable")}
		page := newStandingsPage(nil, stageClient, bracketLoader, slog.Default())

		msg := page.loadBracketStage("stage-1")()

		failed, ok := msg.(fetchErrorMessage)
		require.True(t, ok)
		require.ErrorContains(t, failed.err, "dynamic stage unavailable")
		require.ErrorContains(t, failed.err, "template unavailable")
	})
}

func TestStandingsPage_HandleDynamicStageLoaded(t *testing.T) {
	page := newStandingsPage(nil, &stubStageClient{}, &stubBracketLoader{}, slog.Default())
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

type stubBracketLoader struct {
	template rift.BracketTemplate
	err      error
	calls    int
}

func (l *stubBracketLoader) Load(
	_ context.Context,
	_ string,
) (rift.BracketTemplate, error) {
	l.calls++

	return l.template, l.err
}
