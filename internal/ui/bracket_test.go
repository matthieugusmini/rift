package ui //nolint:testpackage // White-box tests exercise viewport mouse handling.

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
	"github.com/stretchr/testify/assert"
)

func TestBracketPageScrollsHorizontallyWithMouseWheel(t *testing.T) {
	stage := lolesportsgraphql.Stage{
		Sections: []lolesportsgraphql.Section{
			{
				Columns: []lolesportsgraphql.Column{
					columnWithMatchCount(1),
					columnWithMatchCount(1),
					columnWithMatchCount(1),
				},
			},
		},
	}
	page := newDynamicBracketPage(stage, 20, 20)

	assert.Zero(t, page.viewport.HorizontalScrollPercent())

	page, _ = page.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelRight,
	})

	assert.Positive(t, page.viewport.HorizontalScrollPercent())

	page, _ = page.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelLeft,
	})

	assert.Zero(t, page.viewport.HorizontalScrollPercent())
}
