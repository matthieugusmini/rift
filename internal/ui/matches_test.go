package ui //nolint:testpackage // White-box tests verify match card geometry.

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

func TestMatchItemRenderFitsListWidth(t *testing.T) {
	const width = 40

	item := matchItem{
		team1:                team{name: "AAA", gameWins: 3},
		team2:                team{name: "BBB", gameWins: 1},
		leagueName:           "Test League",
		blockName:            "Week 1",
		strategy:             "Bo3",
		flags:                "🇺🇸",
		isCompleted:          true,
		spoilerBlockRevealed: true,
	}
	delegate := newMatchItemDelegate(newTheme(true))
	model := list.New([]list.Item{item}, delegate, width, 10)

	var rendered bytes.Buffer
	delegate.Render(&rendered, model, 0, item)

	lines := strings.Split(ansi.Strip(rendered.String()), "\n")
	require.Len(t, lines, matchItemHeight)

	for i, line := range lines {
		assert.Equal(t, width, lipgloss.Width(line), "line %d", i)
	}

	assert.Equal(t, width-2, strings.Count(lines[2], "─"))
}
