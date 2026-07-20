package ui //nolint:testpackage // White-box tests exercise viewport mouse handling.

import (
	"testing"

	tea "charm.land/bubbletea/v2"

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
	page := newDynamicBracketPage(stage, 20, 20, newTheme(true))

	assert.Zero(t, page.viewport.HorizontalScrollPercent())

	page, _ = page.Update(tea.MouseWheelMsg{
		Button: tea.MouseWheelRight,
	})

	assert.Positive(t, page.viewport.HorizontalScrollPercent())

	page, _ = page.Update(tea.MouseWheelMsg{
		Button: tea.MouseWheelLeft,
	})

	assert.Zero(t, page.viewport.HorizontalScrollPercent())
}
