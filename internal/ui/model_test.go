package ui //nolint:testpackage // White-box tests exercise root view configuration.

import (
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
