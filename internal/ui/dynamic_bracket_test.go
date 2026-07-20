package ui //nolint:testpackage // White-box tests verify the internal layout coordinates.

import (
	"strings"
	"testing"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayoutDynamicBracket(t *testing.T) {
	t.Run("aligns a single-elimination bracket from destination IDs", func(t *testing.T) {
		section := lolesportsgraphql.Section{
			Columns: []lolesportsgraphql.Column{
				columnWithMatches(
					matchAdvancingTo("quarterfinal-1", "semifinal-1"),
					matchAdvancingTo("quarterfinal-2", "semifinal-1"),
					matchAdvancingTo("quarterfinal-3", "semifinal-2"),
					matchAdvancingTo("quarterfinal-4", "semifinal-2"),
				),
				columnWithMatches(
					matchAdvancingTo("semifinal-1", "final"),
					matchAdvancingTo("semifinal-2", "final"),
				),
				columnWithMatches(matchAdvancingTo("final", "")),
			},
		}

		layout := layoutDynamicBracket(section)

		require.Len(t, layout.columns, 3)
		assert.Equal(t, []int{0, 7, 14, 21}, dynamicMatchTops(layout.columns[0]))
		assert.Equal(t, []int{3, 17}, dynamicMatchTops(layout.columns[1]))
		assert.Equal(t, []int{10}, dynamicMatchTops(layout.columns[2]))
		assert.Len(t, layout.edges, 6)
		assert.Equal(t, 26, layout.contentHeight)
		assertDynamicMatchesDoNotOverlap(t, layout)
	})

	t.Run("centers Swiss rounds without explicit destinations", func(t *testing.T) {
		section := lolesportsgraphql.Section{
			Columns: []lolesportsgraphql.Column{
				columnWithMatchCount(4),
				columnWithMatchCount(4),
				columnWithMatchCount(4),
				columnWithMatchCount(3),
				columnWithMatchCount(1),
			},
		}

		layout := layoutDynamicBracket(section)

		assert.Empty(t, layout.edges)
		assert.Equal(t, []int{3, 10, 17}, dynamicMatchTops(layout.columns[3]))
		assert.Equal(t, []int{10}, dynamicMatchTops(layout.columns[4]))
		assertDynamicMatchesDoNotOverlap(t, layout)
	})

	t.Run("lays out winner and loser paths in double elimination", func(t *testing.T) {
		section := lolesportsgraphql.Section{
			Columns: []lolesportsgraphql.Column{
				columnWithMatches(
					matchAdvancingToWinAndLoss("round-1-a", "upper-final", "lower-semifinal"),
					matchAdvancingToWinAndLoss("round-1-b", "upper-final", "lower-semifinal"),
				),
				columnWithMatches(
					matchAdvancingToWinAndLoss("upper-final", "grand-final", "lower-final"),
					matchAdvancingTo("lower-semifinal", "lower-final"),
				),
				columnWithMatches(matchAdvancingTo("lower-final", "grand-final")),
				columnWithMatches(matchAdvancingTo("grand-final", "")),
			},
		}

		layout := layoutDynamicBracket(section)
		view := renderDynamicBracket(section, bracketPageStyles{})

		assert.Len(t, layout.edges, 8)
		assert.Equal(t, 5, dynamicEdgeCount(layout.edges, dynamicEdgeOutcomeWin))
		assert.Equal(t, 3, dynamicEdgeCount(layout.edges, dynamicEdgeOutcomeLoss))
		assert.True(t, dynamicEdgeExists(
			layout.edges,
			dynamicNodeRef{column: 0, match: 0},
			dynamicNodeRef{column: 1, match: 1},
			dynamicEdgeOutcomeLoss,
		))
		assert.True(t, dynamicEdgeExists(
			layout.edges,
			dynamicNodeRef{column: 1, match: 1},
			dynamicNodeRef{column: 2, match: 0},
			dynamicEdgeOutcomeWin,
		))
		assert.Len(t, layout.columns[0].matches, 2)
		assert.Len(t, layout.columns[1].matches, 2)
		assert.Len(t, layout.columns[2].matches, 1)
		assert.Len(t, layout.columns[3].matches, 1)
		assertDynamicMatchesDoNotOverlap(t, layout)

		lowerMatch := layout.columns[1].matches[1]
		lowerMatchLeft := layout.columns[1].left +
			(layout.columns[1].width-matchWidth)/2
		lowerMatchCenterRow := dynamicHeaderHeight + lowerMatch.top + dynamicMatchHeight/2
		renderedRow := []rune(strings.Split(view, "\n")[lowerMatchCenterRow])
		assert.Equal(t, ' ', renderedRow[lowerMatchLeft-1])
		assert.Equal(t, '─', renderedRow[lowerMatchLeft+matchWidth])
	})

	t.Run("preserves separation between cells", func(t *testing.T) {
		section := lolesportsgraphql.Section{
			Columns: []lolesportsgraphql.Column{
				{
					Cells: []lolesportsgraphql.Cell{
						{
							Name:    "Upper",
							Matches: []lolesportsgraphql.Match{{StructuralID: "upper"}},
						},
						{
							Name:    "Lower",
							Matches: []lolesportsgraphql.Match{{StructuralID: "lower"}},
						},
					},
				},
			},
		}

		layout := layoutDynamicBracket(section)

		assert.Equal(t, "Upper / Lower", layout.columns[0].title)
		assert.Equal(t, []int{0, 8}, dynamicMatchTops(layout.columns[0]))
	})
}

func TestRenderDynamicBracket(t *testing.T) {
	section := lolesportsgraphql.Section{
		Columns: []lolesportsgraphql.Column{
			{
				Cells: []lolesportsgraphql.Cell{
					{
						Name: "Semifinals",
						Matches: []lolesportsgraphql.Match{
							matchWithTeams("semifinal-1", "final", "AAA", "BBB"),
							matchWithTeams("semifinal-2", "final", "CCC", "DDD"),
						},
					},
				},
			},
			{
				Cells: []lolesportsgraphql.Cell{
					{
						Name:    "Finals",
						Matches: []lolesportsgraphql.Match{{StructuralID: "final"}},
					},
				},
			},
		},
	}

	view := renderDynamicBracket(section, bracketPageStyles{})

	assert.Contains(t, view, "Semifinals")
	assert.Contains(t, view, "Finals")
	assert.Contains(t, view, "AAA")
	assert.Contains(t, view, "DDD")
	assert.Equal(t, 2, strings.Count(view, "TBD"))
	assert.Contains(t, view, "─")
	assert.Contains(t, view, "│")

	lines := strings.Split(view, "\n")
	assert.Len(t, lines, 15)

	for _, line := range lines {
		assert.Len(t, []rune(line), 45)
	}
}

func TestRenderDynamicBracketTargetsDestinationTeamSlots(t *testing.T) {
	section := lolesportsgraphql.Section{
		Columns: []lolesportsgraphql.Column{
			columnWithMatches(
				matchAdvancingToSlot("semifinal-1", "final", 1),
				matchAdvancingToSlot("semifinal-2", "final", 2),
			),
			columnWithMatches(matchAdvancingTo("final", "")),
		},
	}

	layout := layoutDynamicBracket(section)
	view := renderDynamicBracket(section, bracketPageStyles{})
	finalColumn := layout.columns[1]
	finalMatch := finalColumn.matches[0]
	finalLeft := finalColumn.left + (finalColumn.width-matchWidth)/2
	lines := strings.Split(view, "\n")
	upperTeamRow := dynamicHeaderHeight + finalMatch.top + 1
	dividerRow := dynamicHeaderHeight + finalMatch.top + dynamicMatchHeight/2
	lowerTeamRow := dynamicHeaderHeight + finalMatch.top + 3

	assert.NotEqual(t, ' ', []rune(lines[upperTeamRow])[finalLeft-1])
	assert.Equal(t, ' ', []rune(lines[dividerRow])[finalLeft-1])
	assert.NotEqual(t, ' ', []rune(lines[lowerTeamRow])[finalLeft-1])
}

func TestRenderDynamicBracketDisplaysFullRoundTitle(t *testing.T) {
	const title = "Upper Bracket - Finals"

	section := lolesportsgraphql.Section{
		Columns: []lolesportsgraphql.Column{
			{
				Cells: []lolesportsgraphql.Cell{
					{
						Name:    title,
						Matches: []lolesportsgraphql.Match{{StructuralID: "final"}},
					},
				},
			},
		},
	}

	layout := layoutDynamicBracket(section)
	view := renderDynamicBracket(section, bracketPageStyles{})

	assert.Greater(t, layout.columns[0].width, matchWidth)
	assert.Contains(t, view, title)
	assert.NotContains(t, view, "Upper Bracket - Fina\n")
}

func TestDefaultBracketPageStylesHighlightsWinner(t *testing.T) {
	styles := newDefaultBracketPageStyles(newTheme(true))

	assert.True(t, styles.winnerTeamName.GetBold())
	assert.False(t, styles.winnerTeamName.GetFaint())
	assert.True(t, styles.loserTeamName.GetFaint())
}

func TestRenderDynamicStage(t *testing.T) {
	stage := lolesportsgraphql.Stage{
		Sections: []lolesportsgraphql.Section{
			{
				Name: "Group A",
				Columns: []lolesportsgraphql.Column{
					columnWithMatches(matchWithTeams("group-a", "", "AAA", "BBB")),
				},
			},
			{
				Name: "Group B",
				Columns: []lolesportsgraphql.Column{
					columnWithMatches(matchWithTeams("group-b", "", "CCC", "DDD")),
				},
			},
		},
	}

	view := renderDynamicStage(stage, bracketPageStyles{})

	assert.Contains(t, view, "Group A")
	assert.Contains(t, view, "Group B")
	assert.Contains(t, view, "AAA")
	assert.Contains(t, view, "DDD")
}

func TestFindDynamicRouteRow(t *testing.T) {
	section := lolesportsgraphql.Section{
		Columns: []lolesportsgraphql.Column{
			columnWithMatches(matchAdvancingTo("source", "target")),
			columnWithMatches(
				matchAdvancingTo("middle-1", ""),
				matchAdvancingTo("middle-2", ""),
			),
			columnWithMatches(matchAdvancingTo("target", "")),
		},
	}
	layout := layoutDynamicBracket(section)
	require.Len(t, layout.edges, 1)

	routeRow := findDynamicRouteRow(
		layout,
		layout.edges[0],
		layout.columns[0].matches[0].top,
		layout.columns[2].matches[0].top,
	)

	assert.True(t, dynamicRouteRowIsFree(layout, layout.edges[0], routeRow))
}

func columnWithMatches(matches ...lolesportsgraphql.Match) lolesportsgraphql.Column {
	return lolesportsgraphql.Column{
		Cells: []lolesportsgraphql.Cell{{Name: "Round", Matches: matches}},
	}
}

func columnWithMatchCount(count int) lolesportsgraphql.Column {
	matches := make([]lolesportsgraphql.Match, count)
	for index := range matches {
		matches[index].StructuralID = "match-" + string(rune('a'+index))
	}

	return columnWithMatches(matches...)
}

func matchAdvancingTo(structuralID, destinationID string) lolesportsgraphql.Match {
	match := lolesportsgraphql.Match{StructuralID: structuralID}
	if destinationID == "" {
		return match
	}

	match.Destinations = &lolesportsgraphql.Destinations{
		Win: lolesportsgraphql.Destination{
			Type:         lolesportsgraphql.DestinationTypeMatch,
			StructuralID: new(destinationID),
		},
	}

	return match
}

func matchAdvancingToSlot(
	structuralID, destinationID string,
	slot int,
) lolesportsgraphql.Match {
	match := matchAdvancingTo(structuralID, destinationID)
	match.Destinations.Win.Slot = slot

	return match
}

func matchAdvancingToWinAndLoss(
	structuralID, winDestinationID, lossDestinationID string,
) lolesportsgraphql.Match {
	match := matchAdvancingTo(structuralID, winDestinationID)
	match.Destinations.Loss = lolesportsgraphql.Destination{
		Type:         lolesportsgraphql.DestinationTypeMatch,
		StructuralID: new(lossDestinationID),
	}

	return match
}

func matchWithTeams(structuralID, destinationID, team1, team2 string) lolesportsgraphql.Match {
	match := matchAdvancingTo(structuralID, destinationID)
	match.Teams = []lolesportsgraphql.MatchTeam{
		{Code: team1},
		{Code: team2},
	}

	return match
}

func dynamicMatchTops(column dynamicColumnLayout) []int {
	tops := make([]int, len(column.matches))
	for index, match := range column.matches {
		tops[index] = match.top
	}

	return tops
}

func assertDynamicMatchesDoNotOverlap(t *testing.T, layout dynamicBracketLayout) {
	t.Helper()

	for _, column := range layout.columns {
		for index := 1; index < len(column.matches); index++ {
			previous := column.matches[index-1]
			current := column.matches[index]
			assert.GreaterOrEqual(
				t,
				current.top,
				previous.top+dynamicMatchHeight+dynamicMatchGap,
			)
		}
	}
}

func dynamicEdgeCount(edges []dynamicEdgeLayout, outcome dynamicEdgeOutcome) int {
	count := 0

	for _, edge := range edges {
		if edge.outcome == outcome {
			count++
		}
	}

	return count
}

func dynamicEdgeExists(
	edges []dynamicEdgeLayout,
	from, to dynamicNodeRef,
	outcome dynamicEdgeOutcome,
) bool {
	for _, edge := range edges {
		if edge.from == from && edge.to == to && edge.outcome == outcome {
			return true
		}
	}

	return false
}
