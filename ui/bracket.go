package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/rift"
)

const (
	horizontalLine    = "─"
	verticalLine      = "│"
	topRightCorner    = "┐"
	bottomLeftCorner  = "└"
	topLeftCorner     = "┌"
	bottomRightCorner = "┘"
	topTShape         = "┬"
	bottomTShape      = "┴"
)

type BracketModelStyles struct {
	roundTitle       lipgloss.Style
	match            lipgloss.Style
	noTeamResult     lipgloss.Style
	loserTeamResult  lipgloss.Style
	winnerTeamResult lipgloss.Style
	link             lipgloss.Style
}

func NewDefaultBracketModelStyles() (s BracketModelStyles) {
	s.roundTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(black)).
		Background(lipgloss.Color(antiFlashWhite)).
		Padding(0, 1).
		Bold(true)

	s.match = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderSecondaryColor)

	s.noTeamResult = lipgloss.NewStyle().
		Foreground(textPrimaryColor)

	s.loserTeamResult = lipgloss.NewStyle().
		Foreground(textSecondaryColor)

	s.winnerTeamResult = lipgloss.NewStyle().
		Foreground(selectedBorderColor).
		Bold(true)

	s.link = lipgloss.NewStyle().Foreground(textSecondaryColor)

	return s
}

type BracketModel struct {
	template rift.BracketTemplate
	matches  []lolesports.Match

	width, height int
	styles        BracketModelStyles
}

func NewBracketModel(
	template rift.BracketTemplate,
	stage lolesports.Stage,
	width, height int,
) BracketModel {
	return BracketModel{
		template: template,
		matches:  stage.Sections[0].Matches,
		width:    width,
		height:   height,
		styles:   NewDefaultBracketModelStyles(),
	}
}

func (m BracketModel) View() string {
	nbRounds := len(m.template.Rounds)
	nbLinkColumns := nbRounds - 1
	widthWithoutLinks := m.width - nbLinkColumns*linkWidth
	if widthWithoutLinks <= 0 {
		return ""
	}
	columnsWidth := widthWithoutLinks / nbRounds

	var (
		roundViewsIndex int
		matchIndex      int
		roundViews      = make([]string, nbRounds+nbLinkColumns)
	)

	for _, round := range m.template.Rounds {
		if len(round.Links) > 0 {
			links := m.viewLinks(round.Links)
			roundViews[roundViewsIndex] = links
			roundViewsIndex++
		}

		roundView := lipgloss.PlaceHorizontal(
			columnsWidth,
			lipgloss.Center,
			m.styles.roundTitle.Render(round.Title),
			lipgloss.WithWhitespaceBackground(lipgloss.Color(antiFlashWhite)),
		)

		roundView += "\n\n"
		for _, match := range round.Matches {
			roundView += strings.Repeat("\n", match.Above)

			var card string
			switch match.DisplayType {
			case rift.DisplayTypeMatch:
				card = m.drawMatch(m.matches[matchIndex], columnsWidth)
				matchIndex++
			case rift.DisplayTypeHorizontalLine:
				line := m.styles.link.Render(horizontalLine)
				card = strings.Repeat(line, columnsWidth)
			}

			roundView += card + "\n\n"
		}
		roundView = lipgloss.NewStyle().
			Width(columnsWidth).
			Height(m.height).
			Render(roundView)
		roundViews[roundViewsIndex] = roundView
		roundViewsIndex++
	}

	view := lipgloss.JoinHorizontal(lipgloss.Top, roundViews...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(view)
}

func (m BracketModel) viewLinks(links []rift.Link) string {
	var linksView string
	for _, link := range links {
		linksView += strings.Repeat("\n", link.Above)
		linksView += m.styles.link.Render(drawLink(link))
	}
	return linksView
}

func (m BracketModel) drawMatch(match lolesports.Match, width int) string {
	var (
		team1Style = m.styles.noTeamResult
		team2Style = m.styles.noTeamResult
	)
	if teamHasWon(match.Teams[0]) {
		team1Style = m.styles.winnerTeamResult
		team2Style = m.styles.loserTeamResult
	} else if teamHasWon(match.Teams[1]) {
		team1Style = m.styles.loserTeamResult
		team2Style = m.styles.winnerTeamResult
	}

	borderWidth := m.styles.match.GetHorizontalFrameSize()
	rowWidth := width - borderWidth
	if rowWidth <= 0 {
		return ""
	}

	rowStyle := lipgloss.NewStyle().
		Width(rowWidth).
		Align(lipgloss.Center)

	team1Row := team1Style.Render(formatTeamRow(match.Teams[0]))
	team2Row := team2Style.Render(formatTeamRow(match.Teams[1]))

	content := fmt.Sprintf(
		"%s\n%s\n%s",
		rowStyle.Render(team1Row),
		m.styles.link.Render(strings.Repeat(horizontalLine, rowWidth)),
		rowStyle.Render(team2Row),
	)

	return m.styles.match.Render(content)
}

const linkWidth = 3

func drawLink(link rift.Link) string {
	var sb strings.Builder
	switch link.Type {
	// ┌
	// │
	// ┘
	case rift.LinkTypeZDown:
		sb.WriteString(horizontalLine + topRightCorner + "\n")
		sb.WriteString(strings.Repeat(" "+verticalLine+"\n", link.Height))
		sb.WriteString(" " + bottomLeftCorner + horizontalLine + "\n")

	// ┐
	// │
	// └
	case rift.LinkTypeZUp:
		sb.WriteString(" " + topLeftCorner + horizontalLine + "\n")
		sb.WriteString(strings.Repeat(" "+verticalLine+"\n", link.Height))
		sb.WriteString(horizontalLine + bottomRightCorner + " ")

	// ───
	case rift.LinkTypeHorizontal:
		sb.WriteString(strings.Repeat(horizontalLine, linkWidth))

	case rift.LinkTypeReseed:
		sb.WriteString(strings.Repeat(" ", linkWidth))
	}
	return sb.String()
}

func formatTeamRow(team lolesports.Team) string {
	row := team.Code
	if team.Result != nil {
		row += " " + strconv.Itoa(team.Result.GameWins)
	}
	return row
}

func teamHasWon(team lolesports.Team) bool {
	if team.Result != nil && team.Result.Outcome != nil && *team.Result.Outcome == "win" {
		return true
	}
	return false
}
