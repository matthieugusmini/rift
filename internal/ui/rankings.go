package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/rift/internal/timeutil"
)

const (
	rankingPageHeaderHeight = 5

	rankingPageShortHelpHeight = 1
	rankingPageFullHelpHeight  = 3
)

type rankingPageKeyMap struct {
	baseKeyMap

	Up       key.Binding
	Down     key.Binding
	Previous key.Binding
}

func newDefaultRankingPageKeyMap() rankingPageKeyMap {
	return rankingPageKeyMap{
		baseKeyMap: newBaseKeyMap(),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Previous: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "previous"),
		),
	}
}

type rankingPageStyles struct {
	// Header
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
	separator        lipgloss.Style

	// Content
	tableTitle  lipgloss.Style
	tableHeader lipgloss.Style
	tableRow    lipgloss.Style

	// Footer
	help lipgloss.Style
}

func newDefaultRankingPageStyles() (s rankingPageStyles) {
	// Header
	s.stageName = lipgloss.NewStyle().
		Foreground(textPrimaryColor).
		Bold(true)

	s.tournamentState = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(black)).
		Background(lipgloss.Color(antiFlashWhite))

	s.tournamentPeriod = lipgloss.NewStyle().
		Foreground(textPrimaryColor)

	s.tournamentType = lipgloss.NewStyle().
		Foreground(textPrimaryColor)

	s.separator = lipgloss.NewStyle().Foreground(borderSecondaryColor)

	// Content
	s.tableTitle = lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Background(lipgloss.Color(antiFlashWhite)).
		Foreground(lipgloss.Color(black))

	s.tableHeader = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(textSecondaryColor).
		Bold(true)

	s.tableRow = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(textPrimaryColor).
		Bold(true)

	// Footer
	s.help = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	return s
}

type rankingPage struct {
	width, height int
	stage         lolesports.Stage
	split         lolesports.Split
	league        lolesports.League
	viewport      viewport.Model
	help          help.Model
	keyMap        rankingPageKeyMap
	styles        rankingPageStyles
}

func newRankingPage(
	split lolesports.Split,
	league lolesports.League,
	stage lolesports.Stage,
	width, height int,
) *rankingPage {
	p := &rankingPage{
		width:  width,
		height: height,
		split:  split,
		league: league,
		stage:  stage,
		help:   help.New(),
		keyMap: newDefaultRankingPageKeyMap(),
		styles: newDefaultRankingPageStyles(),
	}

	p.initViewport()

	return p
}

func (p *rankingPage) Init() tea.Cmd { return nil }

func (p *rankingPage) Update(msg tea.Msg) (*rankingPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keyMap.ShowFullHelp),
			key.Matches(msg, p.keyMap.CloseFullHelp):
			p.toggleFullHelp()
		}
	}

	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)

	return p, cmd
}

func (p *rankingPage) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		p.viewHeader(),
		p.viewport.View(),
		p.viewHelp(),
	)
}

func (p *rankingPage) viewHeader() string {
	stageName := p.styles.stageName.Render(
		fmt.Sprintf("%s: %s Standings", p.split.Name, p.league.Name),
	)

	tournamentState := computeTournamentState(p.split.StartTime, p.split.EndTime)
	tournamentPeriod := formatTournamentPeriod(p.split.StartTime, p.split.EndTime)
	tournamentType := p.split.Region
	stageInfo := strings.Join(
		[]string{
			p.styles.tournamentState.Render(string(tournamentState)),
			p.styles.tournamentPeriod.Render(tournamentPeriod),
			p.styles.tournamentType.Render(tournamentType),
		},
		separatorBullet,
	)

	sep := p.styles.separator.Render(strings.Repeat(separatorLine, p.width))

	return fmt.Sprintf("%s\n\n%s\n%s\n", stageName, stageInfo, sep)
}

func (p *rankingPage) viewHelp() string {
	return p.styles.help.Render(p.help.View(p))
}

func (p *rankingPage) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Up,
		p.keyMap.Down,
		p.keyMap.Previous,
		p.keyMap.Quit,
		p.keyMap.ShowFullHelp,
	}
}

func (p *rankingPage) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Motions
		{
			p.keyMap.Up,
			p.keyMap.Down,
			p.keyMap.Previous,
		},
		// App navigation
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

func (p *rankingPage) toggleFullHelp() {
	p.help.ShowAll = !p.help.ShowAll
	// Resize the viewport as the full help takes up more space.
	p.initViewport()
}

func (p *rankingPage) setSize(width, height int) {
	p.width, p.height = width, height
	p.initViewport()
}

func (p *rankingPage) initViewport() {
	content := renderRankings(p.stage, p.width, p.styles)
	p.viewport = viewport.New(p.width, p.contentHeight())
	p.viewport.SetContent(content)
}

func (p *rankingPage) contentHeight() int {
	return p.height - rankingPageHeaderHeight - p.helpHeight()
}

func (p *rankingPage) helpHeight() int {
	padding := p.styles.help.GetVerticalPadding()
	if p.help.ShowAll {
		return rankingPageFullHelpHeight + padding
	}
	return rankingPageShortHelpHeight + padding
}

func renderRankings(stage lolesports.Stage, width int, styles rankingPageStyles) string {
	rankingTable := make([]*table.Table, len(stage.Sections))
	for i, section := range stage.Sections {
		rankingTable[i] = newRankingTable(section.Rankings, width, styles)
	}

	var sb strings.Builder

	for i, t := range rankingTable {
		title := lipgloss.PlaceHorizontal(
			width,
			lipgloss.Center,
			styles.tableTitle.Render(stage.Sections[i].Name),
			lipgloss.WithWhitespaceBackground(lipgloss.Color(antiFlashWhite)),
		)
		sb.WriteString(title + "\n")
		sb.WriteString(t.Render())

		if i < len(rankingTable)-1 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

func newRankingTable(
	rankings []lolesports.Ranking,
	width int,
	styles rankingPageStyles,
) *table.Table {
	headers := []string{"Ranking", "Team", "Series Win / Loss", "Win / Loss %"}

	var rows [][]string
	for _, ranking := range rankings {
		for _, team := range ranking.Teams {
			seriesWinAndLoss := fmt.Sprintf("%dW - %dL", team.Record.Wins, team.Record.Losses)

			winrate := fmt.Sprintf("%d%%", calculateWinrate(team.Record.Wins, team.Record.Losses))

			row := []string{
				strconv.Itoa(ranking.Ordinal),
				team.Code,
				seriesWinAndLoss,
				winrate,
			}
			rows = append(rows, row)
		}
	}

	return table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(selectedColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return styles.tableHeader
			default:
				return styles.tableRow
			}
		}).
		Headers(headers...).
		Rows(rows...).
		Width(width)
}

func calculateWinrate(wins, losses int) int {
	totalGames := wins + losses
	if totalGames == 0 {
		return 0
	}
	return int(float64(wins) / float64(totalGames) * 100)
}

func formatTournamentPeriod(startDate, endDate time.Time) string {
	startMonth := startDate.Format("January")
	endMonth := endDate.Format("January")
	return fmt.Sprintf("%s-%s", startMonth, endMonth)
}

type tournmaentState string

const (
	tournmaentStateUnknown    = "UNKNOWN"
	tournmaentStateNotStarted = "UPCOMING"
	tournmaentStateInProgress = "IN PROGRESS"
	tournmaentStateCompleted  = "COMPLETED"
)

func computeTournamentState(startDate, endDate time.Time) tournmaentState {
	now := time.Now()
	switch {
	case now.Before(startDate):
		return tournmaentStateNotStarted
	case timeutil.IsCurrentTimeBetween(startDate, endDate):
		return tournmaentStateInProgress
	case now.After(endDate):
		return tournmaentStateCompleted
	default:
		return tournmaentStateUnknown
	}
}
