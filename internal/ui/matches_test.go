package ui //nolint:testpackage // White-box tests verify match card geometry.

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchItemRenderFitsListWidth(t *testing.T) {
	const (
		width          = 40
		testLeagueName = "Test League"
		testBlockName  = "Week 1"
		testStrategy   = "Bo3"
	)

	tests := []struct {
		name string
		item matchItem
	}{
		{
			name: "revealed score",
			item: matchItem{
				team1:                team{name: "AAA", gameWins: 3},
				team2:                team{name: "BBB", gameWins: 1},
				leagueName:           testLeagueName,
				blockName:            testBlockName,
				strategy:             testStrategy,
				flags:                "🇺🇸",
				isCompleted:          true,
				spoilerBlockRevealed: true,
			},
		},
		{
			name: "oversized labels",
			item: matchItem{
				team1:                team{name: strings.Repeat("A", width), gameWins: 3},
				team2:                team{name: strings.Repeat("B", width), gameWins: 1},
				leagueName:           "An Extremely Long Regional Championship",
				blockName:            "An Equally Long Elimination Stage",
				strategy:             "Best of an unreasonable number of games",
				flags:                strings.Repeat("🇺🇸", width),
				isCompleted:          true,
				spoilerBlockRevealed: true,
			},
		},
		{
			name: "upcoming oversized teams",
			item: matchItem{
				team1:      team{name: strings.Repeat("A", width)},
				team2:      team{name: strings.Repeat("B", width)},
				startTime:  time.Now().Add(time.Hour),
				leagueName: testLeagueName,
				blockName:  testBlockName,
				strategy:   testStrategy,
			},
		},
		{
			name: "hidden oversized teams",
			item: matchItem{
				team1:       team{name: strings.Repeat("A", width)},
				team2:       team{name: strings.Repeat("B", width)},
				leagueName:  testLeagueName,
				blockName:   testBlockName,
				strategy:    testStrategy,
				isCompleted: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			delegate := newMatchItemDelegate(newTheme(true))
			model := list.New([]list.Item{test.item}, delegate, width, 10)

			var rendered bytes.Buffer
			delegate.Render(&rendered, model, 0, test.item)

			lines := strings.Split(ansi.Strip(rendered.String()), "\n")
			require.Len(t, lines, matchItemHeight)

			for i, line := range lines {
				assert.Equal(t, width, lipgloss.Width(line), "line %d", i)
			}

			assert.Equal(t, width-2, strings.Count(lines[2], "─"))
		})
	}
}
