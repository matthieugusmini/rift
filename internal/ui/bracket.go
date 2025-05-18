package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/rift/internal/rift"
)

const (
	matchWidth = 20
)

const (
	linkWidth = 3

	horizontalLine    = "─"
	verticalLine      = "│"
	topRightCorner    = "┐"
	topLeftCorner     = "┌"
	bottomRightCorner = "┘"
	bottomLeftCorner  = "└"
)

const (
	bracketPageShortHelpHeight = 1
	bracketPageFullHelpHeight  = 5
)

type bracketPageKeyMap struct {
	baseKeyMap

	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Previous key.Binding
}

func newDefaultBracketPageKeyMap() bracketPageKeyMap {
	return bracketPageKeyMap{
		baseKeyMap: newBaseKeyMap(),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Previous: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "previous"),
		),
	}
}

type bracketPageStyles struct {
	roundTitle       lipgloss.Style
	match            lipgloss.Style
	noTeamResult     lipgloss.Style
	loserTeamName    lipgloss.Style
	loserTeamResult  lipgloss.Style
	winnerTeamName   lipgloss.Style
	winnerTeamResult lipgloss.Style
	link             lipgloss.Style
	help             lipgloss.Style
}

func newDefaultBracketPageStyles() (s bracketPageStyles) {
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

	s.help = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	return s
}

type bracketPage struct {
	width, height int
	template      rift.BracketTemplate
	matches       []lolesports.Match
	viewport      viewport.Model
	help          help.Model
	keyMap        bracketPageKeyMap
	styles        bracketPageStyles
}

func newBracketPage(
	template rift.BracketTemplate,
	matches []lolesports.Match,
	width, height int,
) *bracketPage {
	m := &bracketPage{
		template: template,
		matches:  matches,
		width:    width,
		height:   height,
		help:     help.New(),
		keyMap:   newDefaultBracketPageKeyMap(),
		styles:   newDefaultBracketPageStyles(),
	}

	m.initViewport()

	return m
}

func renderBracket(
	tmpl rift.BracketTemplate,
	matches []lolesports.Match,
	width, height int,
	styles bracketPageStyles,
) string {
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

	return lipgloss.NewStyle().
		Width(max(lipgloss.Width(view), width)).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(view)
}

func (m *bracketPage) Update(msg tea.Msg) (*bracketPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ShowFullHelp),
			key.Matches(msg, m.keyMap.CloseFullHelp):
			m.toggleFullHelp()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *bracketPage) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
		m.viewHelp(),
	)
}

func (m *bracketPage) viewHelp() string {
	return m.styles.help.Render(m.help.View(m))
}

func (m *bracketPage) setSize(width, height int) {
	m.width, m.height = width, height

	// Setting the Height and Width field doesn't seem to work
	// so we recreate it with the right size.
	m.initViewport()
}

func (p *bracketPage) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Right,
		p.keyMap.Left,
		p.keyMap.Previous,
		p.keyMap.Quit,
		p.keyMap.ShowFullHelp,
	}
}

func (p *bracketPage) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Motions
		{
			p.keyMap.Up,
			p.keyMap.Down,
			p.keyMap.Right,
			p.keyMap.Left,
			p.keyMap.Previous,
		},
		// Navigation
		{
			p.keyMap.NextPage,
			p.keyMap.PrevPage,
		},
		// Others
		{
			p.keyMap.Quit,
			p.keyMap.CloseFullHelp,
		},
	}
}

func (m *bracketPage) toggleFullHelp() {
	m.help.ShowAll = !m.help.ShowAll
	// Resize the viewport as the full help takes up more space.
	m.initViewport()
}

func (m *bracketPage) initViewport() {
	content := renderBracket(m.template, m.matches, m.width, m.contentHeight(), m.styles)
	m.viewport = viewport.New(m.width, m.contentHeight())
	m.viewport.SetContent(content)
	m.viewport.SetHorizontalStep(5)
}

func (m *bracketPage) contentHeight() int {
	return m.height - m.helpHeight()
}

func (m *bracketPage) helpHeight() int {
	padding := m.styles.help.GetVerticalPadding()
	if m.help.ShowAll {
		return bracketPageFullHelpHeight + padding
	}
	return bracketPageShortHelpHeight + padding
}

func drawMatch(match lolesports.Match, width int, styles bracketPageStyles) string {
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

func drawLinks(links []rift.Link, styles bracketPageStyles) string {
	var linksView string

	for _, link := range links {
		linksView += strings.Repeat("\n", link.Above)
		linksView += styles.link.Render(drawLink(link))
	}

	return linksView
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

	// loser-advance, reseed, etc.
	default:
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
