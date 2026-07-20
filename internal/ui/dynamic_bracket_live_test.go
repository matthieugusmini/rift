package ui_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
	"github.com/matthieugusmini/rift/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicBracketLive(t *testing.T) {
	if os.Getenv("RIFT_LIVE_API_TEST") == "" {
		t.Skip("set RIFT_LIVE_API_TEST to exercise the live LoL Esports API")
	}

	client := lolesportsgraphql.NewClient(http.DefaultClient)
	stage, err := client.GetStage(t.Context(), "115548681802750748")
	require.NoError(t, err)
	require.NotEmpty(t, stage.Sections)

	view := ui.RenderDynamicBracket(stage.Sections[0])

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "╭")
}
