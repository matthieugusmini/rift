package ui //nolint:testpackage // White-box tests verify stage cursor geometry.

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStageItemSelectedCursorSegmentsAlign(t *testing.T) {
	item := stageItem{name: "Playoffs", stageType: stageTypeBracket}
	delegate := newStageItemDelegate(newTheme(true))
	model := list.New([]list.Item{item}, delegate, 30, 10)

	var rendered bytes.Buffer
	delegate.Render(&rendered, model, 0, item)

	lines := strings.Split(ansi.Strip(rendered.String()), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, cursorColumn(lines[0]), cursorColumn(lines[1]))
	assert.True(t, strings.HasPrefix(lines[0], "┃ "))
	assert.True(t, strings.HasPrefix(lines[1], "┃ "))
}

func cursorColumn(line string) int {
	before, _, found := strings.Cut(line, "┃")
	if !found {
		return -1
	}

	return lipgloss.Width(before)
}
