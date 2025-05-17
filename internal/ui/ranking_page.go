package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
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
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
	separator        lipgloss.Style
	help             lipgloss.Style
}

func newDefaultRankingPageStyles() (s rankingPageStyles) {
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
	v := &rankingPage{
		width:  width,
		height: height,
		split:  split,
		league: league,
		stage:  stage,
		help:   help.New(),
		keyMap: newDefaultRankingPageKeyMap(),
		styles: newDefaultRankingPageStyles(),
	}

	v.viewport = newRankingViewport(stage, width, v.contentHeight())

	return v
}

func (v *rankingPage) Init() tea.Cmd { return nil }

func (v *rankingPage) Update(msg tea.Msg) (*rankingPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, v.keyMap.ShowFullHelp),
			key.Matches(msg, v.keyMap.CloseFullHelp):
			v.toggleFullHelp()
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *rankingPage) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		v.viewHeader(),
		v.viewport.View(),
		v.viewHelp(),
	)
}

func (v *rankingPage) viewHeader() string {
	stageName := v.styles.stageName.Render(
		fmt.Sprintf("%s: %s Standings", v.split.Name, v.league.Name),
	)

	tournamentState := computeTournamentState(v.split.StartTime, v.split.EndTime)
	tournamentPeriod := formatTournamentPeriod(v.split.StartTime, v.split.EndTime)
	tournamentType := v.split.Region
	stageInfo := strings.Join(
		[]string{
			v.styles.tournamentState.Render(string(tournamentState)),
			v.styles.tournamentPeriod.Render(tournamentPeriod),
			v.styles.tournamentType.Render(tournamentType),
		},
		separatorBullet,
	)

	sep := v.styles.separator.Render(strings.Repeat(separatorLine, v.width))

	return fmt.Sprintf("%s\n\n%s\n%s\n", stageName, stageInfo, sep)
}

func (v *rankingPage) viewHelp() string {
	return v.styles.help.Render(v.help.View(v))
}

func (v *rankingPage) ShortHelp() []key.Binding {
	return []key.Binding{
		v.keyMap.Up,
		v.keyMap.Down,
		v.keyMap.Previous,
		v.keyMap.Quit,
		v.keyMap.ShowFullHelp,
	}
}

func (v *rankingPage) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Motions
		{
			v.keyMap.Up,
			v.keyMap.Down,
			v.keyMap.Previous,
		},
		// App navigation
		{
			v.keyMap.NextPage,
			v.keyMap.PrevPage,
		},
		// Others
		{
			v.keyMap.Quit,
			v.keyMap.CloseFullHelp,
		},
	}
}

func (v *rankingPage) toggleFullHelp() {
	v.help.ShowAll = !v.help.ShowAll
	// Resize the viewport as the full help takes up more space.
	v.viewport = newRankingViewport(v.stage, v.width, v.contentHeight())
}

func (v *rankingPage) setSize(width, height int) {
	v.width, v.height = width, height
	v.viewport = newRankingViewport(v.stage, width, v.contentHeight())
}

func (v *rankingPage) contentHeight() int {
	return v.height - rankingPageHeaderHeight - v.helpHeight()
}

func (v *rankingPage) helpHeight() int {
	padding := v.styles.help.GetVerticalPadding()
	if v.help.ShowAll {
		return rankingPageFullHelpHeight + padding
	}
	return rankingPageShortHelpHeight + padding
}
