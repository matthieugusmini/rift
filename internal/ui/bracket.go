package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/rift/internal/rift"
)

const (
	matchWidth = 16
	linkWidth  = 3

	horizontalLine    = "─"
	verticalLine      = "│"
	topRightCorner    = "┐"
	topLeftCorner     = "┌"
	bottomRightCorner = "┘"
	bottomLeftCorner  = "└"
)

type bracketModelStyles struct {
	roundTitle       lipgloss.Style
	match            lipgloss.Style
	noTeamResult     lipgloss.Style
	loserTeamName    lipgloss.Style
	loserTeamResult  lipgloss.Style
	winnerTeamName   lipgloss.Style
	winnerTeamResult lipgloss.Style
	link             lipgloss.Style
}

func newDefaultBracketModelStyles() (s bracketModelStyles) {
	s.roundTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(black)).
		Background(lipgloss.Color(antiFlashWhite)).
		Padding(0, 1).
		Bold(true)

	s.match = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderPrimaryColor)

	s.noTeamResult = lipgloss.NewStyle().
		Foreground(textPrimaryColor)

	s.loserTeamName = lipgloss.NewStyle().
		Foreground(textSecondaryColor).
		Faint(true)

	s.loserTeamResult = lipgloss.NewStyle().
		Foreground(textSecondaryColor).
		Bold(true)

	s.winnerTeamName = lipgloss.NewStyle().
		Foreground(selectedColor).
		Faint(true)

	s.winnerTeamResult = lipgloss.NewStyle().
		Foreground(selectedColor).
		Bold(true)

	s.link = lipgloss.NewStyle().Foreground(borderSecondaryColor)

	return s
}

type bracketModel struct {
	template rift.BracketTemplate
	matches  []lolesports.Match

	vp viewport.Model

	width, height int
	styles        bracketModelStyles
}

func newBracketModel(
	template rift.BracketTemplate,
	matches []lolesports.Match,
	width, height int,
) *bracketModel {
	styles := newDefaultBracketModelStyles()

	vp := newBracketViewport(template, matches, width, height, styles)

	return &bracketModel{
		template: template,
		matches:  matches,
		width:    width,
		height:   height,
		vp:       vp,
		styles:   styles,
	}
}

func newBracketViewport(
	tmpl rift.BracketTemplate,
	matches []lolesports.Match,
	width, height int,
	styles bracketModelStyles,
) viewport.Model {
	nbRounds := len(tmpl.Rounds)
	nbLinkColumns := nbRounds - 1

	var (
		sections     = make([]string, nbRounds+nbLinkColumns)
		sectionIndex int
		matchIndex   int
	)
	for _, round := range tmpl.Rounds {
		if len(round.Links) > 0 {
			links := drawLinks(round.Links, styles)

			sections[sectionIndex] = links
			sectionIndex++
		}

		roundView := lipgloss.PlaceHorizontal(
			matchWidth,
			lipgloss.Center,
			styles.roundTitle.Render(round.Title),
			lipgloss.WithWhitespaceBackground(lipgloss.Color(antiFlashWhite)),
		)
		roundView += "\n\n"

		for i, match := range round.Matches {
			roundView += strings.Repeat("\n", match.Above)

			switch match.DisplayType {
			case rift.DisplayTypeMatch:
				roundView += drawMatch(matches[matchIndex], matchWidth, styles)
				matchIndex++
			case rift.DisplayTypeHorizontalLine:
				line := styles.link.Render(horizontalLine)
				roundView += strings.Repeat(line, matchWidth)
			}

			if i < len(round.Matches)-1 {
				roundView += "\n\n"
			}
		}

		roundView = lipgloss.NewStyle().
			Width(matchWidth).
			Height(height).
			Render(roundView)

		sections[sectionIndex] = roundView
		sectionIndex++
	}

	view := lipgloss.JoinHorizontal(lipgloss.Top, sections...)

	view = lipgloss.NewStyle().
		Width(max(lipgloss.Width(view), width)).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(view)

	vp := viewport.New(width, height)
	vp.SetContent(view)
	vp.SetHorizontalStep(5)

	return vp
}

func (m *bracketModel) Update(msg tea.Msg) (*bracketModel, tea.Cmd) {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m *bracketModel) View() string {
	return m.vp.View()
}

func drawLinks(links []rift.Link, styles bracketModelStyles) string {
	var linksView string

	for _, link := range links {
		linksView += strings.Repeat("\n", link.Above)
		linksView += styles.link.Render(drawLink(link))
	}

	return linksView
}

func drawMatch(match lolesports.Match, width int, styles bracketModelStyles) string {
	borderWidth := styles.match.GetHorizontalBorderSize()
	rowWidth := width - borderWidth
	if rowWidth <= 0 {
		return ""
	}

	var (
		team1Style       = styles.noTeamResult
		team2Style       = styles.noTeamResult
		team2ResultStyle lipgloss.Style
		team1ResultStyle lipgloss.Style
	)
	if teamHasWon(match.Teams[0]) {
		team1Style = styles.winnerTeamName
		team1ResultStyle = styles.winnerTeamResult

		team2Style = styles.loserTeamName
		team2ResultStyle = styles.loserTeamResult
	} else if teamHasWon(match.Teams[1]) {
		team1Style = styles.loserTeamName
		team1ResultStyle = styles.loserTeamResult

		team2Style = styles.winnerTeamName
		team2ResultStyle = styles.winnerTeamResult
	}

	rowStyle := lipgloss.NewStyle().
		Width(rowWidth).
		Align(lipgloss.Center)

	team1Row := formatTeamRow(match.Teams[0])
	team1Row = lipgloss.StyleRanges(
		team1Row,
		lipgloss.NewRange(0, len(match.Teams[0].Code), team1Style),
		lipgloss.NewRange(len(match.Teams[0].Code), len(team1Row), team1ResultStyle),
	)

	team2Row := formatTeamRow(match.Teams[1])
	team2Row = lipgloss.StyleRanges(
		team2Row,
		lipgloss.NewRange(0, len(match.Teams[1].Code), team2Style),
		lipgloss.NewRange(len(match.Teams[1].Code), len(team2Row), team2ResultStyle),
	)

	content := fmt.Sprintf(
		"%s\n%s\n%s",
		rowStyle.Render(team1Row),
		styles.link.Render(strings.Repeat(horizontalLine, rowWidth)),
		rowStyle.Render(team2Row),
	)

	return styles.match.Render(content)
}

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

	default:
		sb.WriteString(strings.Repeat(" ", linkWidth))
	}
	return sb.String()
}

func (m *bracketModel) setSize(width, height int) {
	m.width, m.height = width, height

	// Setting the Height and Width field doesn't seem to work
	// so we recreate it with the right size.
	m.vp = newBracketViewport(m.template, m.matches, width, height, m.styles)
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
