package ui //nolint:testpackage // White-box tests exercise root view configuration.

import (
	"image/color"
	"log/slog"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestModelViewConfiguresTerminalFeatures(t *testing.T) {
	model := NewModel(nil, nil, slog.Default())

	view := model.View()

	assert.True(t, view.AltScreen)
	assert.Equal(t, tea.MouseModeCellMotion, view.MouseMode)
}

func TestModelAppliesTerminalBackgroundTheme(t *testing.T) {
	model := NewModel(nil, nil, slog.Default())

	updated, _ := model.Update(tea.BackgroundColorMsg{Color: color.White})
	got := updated.(Model)

	assert.False(t, got.theme.isDark)
	assert.False(t, got.pages[stateShowSchedule].(*schedulePage).theme.isDark)
	assert.False(t, got.pages[stateShowStandings].(*standingsPage).theme.isDark)
}
