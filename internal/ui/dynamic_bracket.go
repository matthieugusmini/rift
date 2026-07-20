package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
)

const (
	dynamicMatchHeight  = 5
	dynamicMatchGap     = 2
	dynamicCellGap      = 1
	dynamicGutterWidth  = 5
	dynamicHeaderHeight = 3
)

type dynamicNodeRef struct {
	column int
	match  int
}

type dynamicMatchLayout struct {
	match      lolesportsgraphql.Match
	top        int
	startsCell bool
}

type dynamicColumnLayout struct {
	title   string
	matches []dynamicMatchLayout
	left    int
	width   int
}

type dynamicEdgeOutcome int

const (
	dynamicEdgeOutcomeWin dynamicEdgeOutcome = iota
	dynamicEdgeOutcomeLoss
)

type dynamicEdgeLayout struct {
	from    dynamicNodeRef
	to      dynamicNodeRef
	outcome dynamicEdgeOutcome
	slot    int
}

type dynamicBracketLayout struct {
	columns       []dynamicColumnLayout
	edges         []dynamicEdgeLayout
	contentHeight int
}

// RenderDynamicBracket renders a bracket section using the topology supplied by LoL Esports.
func RenderDynamicBracket(section lolesportsgraphql.Section) string {
	return renderDynamicBracket(section, newDefaultBracketPageStyles())
}

func renderDynamicStage(
	stage lolesportsgraphql.Stage,
	styles bracketPageStyles,
) string {
	sections := make([]string, 0, len(stage.Sections))
	for _, section := range stage.Sections {
		view := renderDynamicBracket(section, styles)
		if len(stage.Sections) > 1 {
			title := styles.roundTitle.Render(section.Name)
			view = lipgloss.JoinVertical(lipgloss.Left, title, "", view)
		}

		sections = append(sections, view)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func layoutDynamicBracket(section lolesportsgraphql.Section) dynamicBracketLayout {
	layout := dynamicBracketLayout{
		columns: make([]dynamicColumnLayout, len(section.Columns)),
	}

	maxColumnHeight := 0

	for columnIndex, column := range section.Columns {
		matches := flattenDynamicMatches(column)
		positionDynamicMatches(matches)
		columnHeight := dynamicColumnHeight(matches)
		title := dynamicColumnTitle(column)
		maxColumnHeight = max(maxColumnHeight, columnHeight)
		layout.columns[columnIndex] = dynamicColumnLayout{
			title:   title,
			matches: matches,
			width:   max(matchWidth, lipgloss.Width(title)),
		}
	}

	for columnIndex := range layout.columns {
		columnHeight := dynamicColumnHeight(layout.columns[columnIndex].matches)

		offset := (maxColumnHeight - columnHeight) / 2
		for matchIndex := range layout.columns[columnIndex].matches {
			layout.columns[columnIndex].matches[matchIndex].top += offset
		}
	}

	left := 0
	for columnIndex := range layout.columns {
		layout.columns[columnIndex].left = left
		left += layout.columns[columnIndex].width + dynamicGutterWidth
	}

	nodesByStructuralID := make(map[string]dynamicNodeRef)

	for columnIndex, column := range layout.columns {
		for matchIndex, match := range column.matches {
			if match.match.StructuralID != "" {
				nodesByStructuralID[match.match.StructuralID] = dynamicNodeRef{
					column: columnIndex,
					match:  matchIndex,
				}
			}
		}
	}

	layout.edges = listDynamicEdges(layout.columns, nodesByStructuralID)
	alignDynamicMatches(&layout)
	layout.contentHeight = dynamicLayoutHeight(layout.columns)

	return layout
}

func flattenDynamicMatches(column lolesportsgraphql.Column) []dynamicMatchLayout {
	matchCount := 0
	for _, cell := range column.Cells {
		matchCount += len(cell.Matches)
	}

	matches := make([]dynamicMatchLayout, 0, matchCount)

	for _, cell := range column.Cells {
		for matchIndex, match := range cell.Matches {
			matches = append(matches, dynamicMatchLayout{
				match:      match,
				startsCell: len(matches) > 0 && matchIndex == 0,
			})
		}
	}

	return matches
}

func positionDynamicMatches(matches []dynamicMatchLayout) {
	top := 0

	for matchIndex := range matches {
		if matchIndex > 0 {
			top += dynamicMatchHeight + dynamicMatchGap
			if matches[matchIndex].startsCell {
				top += dynamicCellGap
			}
		}

		matches[matchIndex].top = top
	}
}

func dynamicColumnHeight(matches []dynamicMatchLayout) int {
	if len(matches) == 0 {
		return 0
	}

	lastMatch := matches[len(matches)-1]

	return lastMatch.top + dynamicMatchHeight
}

func dynamicColumnTitle(column lolesportsgraphql.Column) string {
	names := make([]string, 0, len(column.Cells))

	seen := make(map[string]bool)
	for _, cell := range column.Cells {
		if cell.Name == "" || seen[cell.Name] {
			continue
		}

		names = append(names, cell.Name)
		seen[cell.Name] = true
	}

	return strings.Join(names, " / ")
}

func listDynamicEdges(
	columns []dynamicColumnLayout,
	nodesByStructuralID map[string]dynamicNodeRef,
) []dynamicEdgeLayout {
	var edges []dynamicEdgeLayout

	for columnIndex, column := range columns {
		for matchIndex, match := range column.matches {
			if match.match.Destinations == nil {
				continue
			}

			from := dynamicNodeRef{column: columnIndex, match: matchIndex}
			edges = appendDynamicEdge(
				edges,
				from,
				match.match.Destinations.Win,
				dynamicEdgeOutcomeWin,
				nodesByStructuralID,
			)
			edges = appendDynamicEdge(
				edges,
				from,
				match.match.Destinations.Loss,
				dynamicEdgeOutcomeLoss,
				nodesByStructuralID,
			)
		}
	}

	return edges
}

func appendDynamicEdge(
	edges []dynamicEdgeLayout,
	from dynamicNodeRef,
	destination lolesportsgraphql.Destination,
	outcome dynamicEdgeOutcome,
	nodesByStructuralID map[string]dynamicNodeRef,
) []dynamicEdgeLayout {
	if destination.Type != lolesportsgraphql.DestinationTypeMatch ||
		destination.StructuralID == nil {
		return edges
	}

	to, ok := nodesByStructuralID[*destination.StructuralID]
	if !ok || to.column <= from.column {
		return edges
	}

	return append(edges, dynamicEdgeLayout{
		from:    from,
		to:      to,
		outcome: outcome,
		slot:    destination.Slot,
	})
}

func alignDynamicMatches(layout *dynamicBracketLayout) {
	incoming := make(map[dynamicNodeRef][]dynamicNodeRef)
	for _, edge := range layout.edges {
		incoming[edge.to] = append(incoming[edge.to], edge.from)
	}

	for columnIndex := 1; columnIndex < len(layout.columns); columnIndex++ {
		minimumTop := 0

		for matchIndex := range layout.columns[columnIndex].matches {
			ref := dynamicNodeRef{column: columnIndex, match: matchIndex}
			match := &layout.columns[columnIndex].matches[matchIndex]

			sources := incoming[ref]
			if len(sources) > 0 {
				centerSum := 0
				for _, source := range sources {
					centerSum += layout.columns[source.column].matches[source.match].top +
						dynamicMatchHeight/2
				}

				match.top = centerSum/len(sources) - dynamicMatchHeight/2
			}

			minimumTop += dynamicMatchCellGap(*match, matchIndex)
			match.top = max(match.top, minimumTop)
			minimumTop = match.top + dynamicMatchHeight + dynamicMatchGap
		}
	}
}

func dynamicMatchCellGap(match dynamicMatchLayout, matchIndex int) int {
	if matchIndex > 0 && match.startsCell {
		return dynamicCellGap
	}

	return 0
}

func dynamicLayoutHeight(columns []dynamicColumnLayout) int {
	height := 0
	for _, column := range columns {
		height = max(height, dynamicColumnHeight(column.matches))
	}

	return height
}

type dynamicCanvasStyle int

const (
	dynamicCanvasStyleNone dynamicCanvasStyle = iota
	dynamicCanvasStyleTitle
	dynamicCanvasStyleBorder
	dynamicCanvasStyleLink
	dynamicCanvasStyleTeam
	dynamicCanvasStyleWinner
	dynamicCanvasStyleLoser
)

type dynamicCanvasCell struct {
	character rune
	style     dynamicCanvasStyle
}

func renderDynamicBracket(
	section lolesportsgraphql.Section,
	styles bracketPageStyles,
) string {
	layout := layoutDynamicBracket(section)
	if len(layout.columns) == 0 {
		return ""
	}

	lastColumn := layout.columns[len(layout.columns)-1]
	canvasWidth := lastColumn.left + lastColumn.width
	canvasHeight := dynamicHeaderHeight + layout.contentHeight
	canvas := newDynamicCanvas(canvasWidth, canvasHeight)

	for _, column := range layout.columns {
		drawDynamicTitle(canvas, column.left, column.width, column.title)
		matchLeft := column.left + (column.width-matchWidth)/2

		for _, match := range column.matches {
			drawDynamicMatch(
				canvas,
				matchLeft,
				dynamicHeaderHeight+match.top,
				match.match,
			)
		}
	}

	drawDynamicConnectors(canvas, layout)

	return renderDynamicCanvas(canvas, styles)
}

func newDynamicCanvas(width, height int) [][]dynamicCanvasCell {
	canvas := make([][]dynamicCanvasCell, height)
	for row := range canvas {
		canvas[row] = make([]dynamicCanvasCell, width)
		for column := range canvas[row] {
			canvas[row][column].character = ' '
		}
	}

	return canvas
}

func drawDynamicTitle(canvas [][]dynamicCanvasCell, left, width int, title string) {
	start := left + max((width-lipgloss.Width(title))/2, 0)
	writeDynamicText(canvas, 0, start, title, dynamicCanvasStyleTitle)
}

func drawDynamicMatch(
	canvas [][]dynamicCanvasCell,
	left, top int,
	match lolesportsgraphql.Match,
) {
	innerWidth := matchWidth - 2

	drawDynamicHorizontalBorder(canvas, left, top, '╭', '╮')
	drawDynamicTeamRow(canvas, left, top+1, innerWidth, dynamicMatchTeam(match, 0))
	drawDynamicHorizontalBorder(canvas, left, top+2, '├', '┤')
	drawDynamicTeamRow(canvas, left, top+3, innerWidth, dynamicMatchTeam(match, 1))
	drawDynamicHorizontalBorder(canvas, left, top+4, '╰', '╯')
}

func drawDynamicHorizontalBorder(
	canvas [][]dynamicCanvasCell,
	left, row int,
	start, end rune,
) {
	writeDynamicCell(canvas, row, left, start, dynamicCanvasStyleBorder)

	for column := 1; column < matchWidth-1; column++ {
		writeDynamicCell(canvas, row, left+column, '─', dynamicCanvasStyleBorder)
	}

	writeDynamicCell(canvas, row, left+matchWidth-1, end, dynamicCanvasStyleBorder)
}

func drawDynamicTeamRow(
	canvas [][]dynamicCanvasCell,
	left, row, innerWidth int,
	team *lolesportsgraphql.MatchTeam,
) {
	writeDynamicCell(canvas, row, left, '│', dynamicCanvasStyleBorder)
	writeDynamicCell(canvas, row, left+matchWidth-1, '│', dynamicCanvasStyleBorder)

	label := "TBD"
	style := dynamicCanvasStyleTeam

	if team != nil {
		label = team.Code
		if label == "" {
			label = team.Name
		}

		if team.Result != nil {
			label += " " + strconv.Itoa(team.Result.GameWins)
			if team.Result.Outcome != nil {
				switch *team.Result.Outcome {
				case "win":
					style = dynamicCanvasStyleWinner
				case "loss":
					style = dynamicCanvasStyleLoser
				}
			}
		}
	}

	label = ansi.Truncate(label, innerWidth, "")
	start := left + 1 + max((innerWidth-lipgloss.Width(label))/2, 0)
	writeDynamicText(canvas, row, start, label, style)
}

func dynamicMatchTeam(
	match lolesportsgraphql.Match,
	index int,
) *lolesportsgraphql.MatchTeam {
	if index >= len(match.Teams) {
		return nil
	}

	return &match.Teams[index]
}

const (
	connectorNorth uint8 = 1 << iota
	connectorEast
	connectorSouth
	connectorWest
)

func drawDynamicConnectors(
	canvas [][]dynamicCanvasCell,
	layout dynamicBracketLayout,
) {
	connections := make([][]uint8, len(canvas))
	for row := range connections {
		connections[row] = make([]uint8, len(canvas[row]))
	}

	for _, edge := range layout.edges {
		if edge.outcome == dynamicEdgeOutcomeLoss {
			continue
		}

		fromColumn := layout.columns[edge.from.column]
		toColumn := layout.columns[edge.to.column]
		from := layout.columns[edge.from.column].matches[edge.from.match]
		to := layout.columns[edge.to.column].matches[edge.to.match]

		fromLeft := fromColumn.left + (fromColumn.width-matchWidth)/2
		toLeft := toColumn.left + (toColumn.width-matchWidth)/2
		fromRow := dynamicHeaderHeight + from.top + dynamicMatchHeight/2
		toRow := dynamicHeaderHeight + dynamicDestinationRow(to.top, edge.slot)
		fromStart := fromLeft + matchWidth
		toEnd := toLeft - 1
		fromLane := fromColumn.left + fromColumn.width + dynamicGutterWidth/2
		toLane := toColumn.left - dynamicGutterWidth/2 - 1
		routeRow := dynamicHeaderHeight + findDynamicRouteRow(layout, edge, from.top, to.top)

		addDynamicHorizontalConnection(connections, fromRow, fromStart, fromLane)
		addDynamicVerticalConnection(connections, fromLane, fromRow, routeRow)
		addDynamicHorizontalConnection(connections, routeRow, fromLane, toLane)
		addDynamicVerticalConnection(connections, toLane, routeRow, toRow)
		addDynamicHorizontalConnection(connections, toRow, toLane, toEnd)
	}

	for row := range connections {
		for column, connection := range connections[row] {
			if connection == 0 || canvas[row][column].character != ' ' {
				continue
			}

			canvas[row][column] = dynamicCanvasCell{
				character: dynamicConnectorCharacter(connection),
				style:     dynamicCanvasStyleLink,
			}
		}
	}
}

func dynamicDestinationRow(matchTop, slot int) int {
	switch slot {
	case 1:
		return matchTop + 1
	case 2:
		return matchTop + 3
	default:
		return matchTop + dynamicMatchHeight/2
	}
}

func findDynamicRouteRow(
	layout dynamicBracketLayout,
	edge dynamicEdgeLayout,
	fromTop, toTop int,
) int {
	fromCenter := fromTop + dynamicMatchHeight/2
	toCenter := toTop + dynamicMatchHeight/2
	desired := (fromCenter + toCenter) / 2

	for distance := 0; distance <= layout.contentHeight; distance++ {
		for _, candidate := range []int{desired - distance, desired + distance} {
			if candidate < 0 || candidate >= layout.contentHeight {
				continue
			}

			if dynamicRouteRowIsFree(layout, edge, candidate) {
				return candidate
			}
		}
	}

	return desired
}

func dynamicRouteRowIsFree(
	layout dynamicBracketLayout,
	edge dynamicEdgeLayout,
	row int,
) bool {
	for columnIndex := edge.from.column + 1; columnIndex < edge.to.column; columnIndex++ {
		for _, match := range layout.columns[columnIndex].matches {
			if row >= match.top && row < match.top+dynamicMatchHeight {
				return false
			}
		}
	}

	return true
}

func addDynamicHorizontalConnection(connections [][]uint8, row, start, end int) {
	if start > end {
		start, end = end, start
	}

	for column := start; column < end; column++ {
		if row < 0 || row >= len(connections) ||
			column < 0 || column+1 >= len(connections[row]) {
			continue
		}

		connections[row][column] |= connectorEast
		connections[row][column+1] |= connectorWest
	}
}

func addDynamicVerticalConnection(connections [][]uint8, column, start, end int) {
	if start > end {
		start, end = end, start
	}

	for row := start; row < end; row++ {
		if row < 0 || row+1 >= len(connections) ||
			column < 0 || column >= len(connections[row]) {
			continue
		}

		connections[row][column] |= connectorSouth
		connections[row+1][column] |= connectorNorth
	}
}

func dynamicConnectorCharacter(connection uint8) rune {
	switch connection {
	case connectorEast, connectorWest, connectorEast | connectorWest:
		return '─'
	case connectorNorth, connectorSouth, connectorNorth | connectorSouth:
		return '│'
	case connectorEast | connectorSouth:
		return '┌'
	case connectorWest | connectorSouth:
		return '┐'
	case connectorEast | connectorNorth:
		return '└'
	case connectorWest | connectorNorth:
		return '┘'
	case connectorEast | connectorWest | connectorSouth:
		return '┬'
	case connectorEast | connectorWest | connectorNorth:
		return '┴'
	case connectorNorth | connectorEast | connectorSouth:
		return '├'
	case connectorNorth | connectorSouth | connectorWest:
		return '┤'
	default:
		return '┼'
	}
}

func writeDynamicText(
	canvas [][]dynamicCanvasCell,
	row, start int,
	text string,
	style dynamicCanvasStyle,
) {
	for offset, character := range []rune(text) {
		writeDynamicCell(canvas, row, start+offset, character, style)
	}
}

func writeDynamicCell(
	canvas [][]dynamicCanvasCell,
	row, column int,
	character rune,
	style dynamicCanvasStyle,
) {
	if row < 0 || row >= len(canvas) || column < 0 || column >= len(canvas[row]) {
		return
	}

	canvas[row][column] = dynamicCanvasCell{character: character, style: style}
}

func renderDynamicCanvas(
	canvas [][]dynamicCanvasCell,
	styles bracketPageStyles,
) string {
	var output strings.Builder

	for rowIndex, row := range canvas {
		for start := 0; start < len(row); {
			end := start + 1
			for end < len(row) && row[end].style == row[start].style {
				end++
			}

			var text strings.Builder
			for _, cell := range row[start:end] {
				text.WriteRune(cell.character)
			}

			output.WriteString(renderDynamicCanvasRun(text.String(), row[start].style, styles))
			start = end
		}

		if rowIndex < len(canvas)-1 {
			output.WriteByte('\n')
		}
	}

	return output.String()
}

func renderDynamicCanvasRun(
	text string,
	style dynamicCanvasStyle,
	styles bracketPageStyles,
) string {
	switch style {
	case dynamicCanvasStyleTitle:
		return styles.roundTitle.Padding(0).Render(text)
	case dynamicCanvasStyleBorder:
		return styles.matchBorder.Render(text)
	case dynamicCanvasStyleLink:
		return styles.link.Render(text)
	case dynamicCanvasStyleTeam:
		return styles.noTeamResult.Render(text)
	case dynamicCanvasStyleWinner:
		return styles.winnerTeamName.Render(text)
	case dynamicCanvasStyleLoser:
		return styles.loserTeamName.Render(text)
	case dynamicCanvasStyleNone:
		return text
	default:
		return text
	}
}
