package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
)

const selectionListCount = 3

type standingsPageState int

const (
	standingsPageStateLoadingSplits standingsPageState = iota
	standingsPageStateSplitSelection
	standingsPageStateLeagueSelection
	standingsPageStateLoadingStages
	standingsPageStateStageSelection
	standingsPageStateShowStandings
)

type standingsStyles struct {
	doc              lipgloss.Style
	tableBorder      lipgloss.Style
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
}

func newDefaultStandingsStyles() (s standingsStyles) {
	s.doc = lipgloss.NewStyle().Padding(1, 2)

	s.stageName = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(textPrimaryColor).
		Bold(true)

	s.tournamentState = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("#000000")).
		Background(textPrimaryColor)

	s.tournamentPeriod = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(textPrimaryColor)

	s.tournamentType = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(textPrimaryColor)

	s.tableBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return s
}

type standingsPage struct {
	lolesportsClient LoLEsportsClient

	state standingsPageState

	splits         []lolesports.Split
	leagues        []lolesports.League
	stages         []lolesports.Stage
	standingsCache map[string][]lolesports.Standings

	// Selection phase
	splitOptions  list.Model
	leagueOptions list.Model
	stageOptions  list.Model

	// Standings page
	standings viewport.Model

	err error

	spinner spinner.Model

	height, width int

	styles standingsStyles
}

func newStandingsPage(lolesportsClient LoLEsportsClient) *standingsPage {
	return &standingsPage{
		lolesportsClient: lolesportsClient,
		styles:           newDefaultStandingsStyles(),
		standingsCache:   map[string][]lolesports.Standings{},
		spinner:          spinner.New(spinner.WithSpinner(spinner.Monkey)),
	}
}

func (p *standingsPage) Init() tea.Cmd {
	if p.state != standingsPageStateLoadingSplits {
		return nil
	}
	return tea.Batch(p.spinner.Tick, p.fetchCurrentSeasonSplits())
}

func (p *standingsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			p.goToPreviousStep()

		case "enter":
			return p.handleSelection()
		}

	case spinner.TickMsg:
		if p.isLoading() {
			var cmd tea.Cmd
			p.spinner, cmd = p.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case fetchedStandingsMessage:
		p.state = standingsPageStateStageSelection
		p.setStandingsInCache(msg.standings)
		p.stages = listStagesFromStandings(msg.standings)
		p.stageOptions = newStageOptionsList(p.stages, p.optionListWidth(), p.height)

	case fetchedCurrentSeasonSplitsMessage:
		p.state = standingsPageStateSplitSelection
		p.splits = msg.splits
		p.splitOptions = newSplitOptionsList(p.splits, p.optionListWidth(), p.height)

	case fetchErrorMessage:
		p.err = msg.err
	}

	var cmd tea.Cmd
	switch p.state {
	case standingsPageStateSplitSelection:
		p.splitOptions, cmd = p.splitOptions.Update(msg)
	case standingsPageStateLeagueSelection:
		p.leagueOptions, cmd = p.leagueOptions.Update(msg)
	case standingsPageStateStageSelection:
		p.stageOptions, cmd = p.stageOptions.Update(msg)
	case standingsPageStateShowStandings:
		p.standings, cmd = p.standings.Update(msg)
	}
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *standingsPage) View() string {
	if p.err != nil {
		return p.err.Error()
	}
	if p.state == standingsPageStateShowStandings {
		return p.styles.tableBorder.Render(p.standings.View())
	}
	return p.viewSelection()
}

func (p *standingsPage) viewSelection() string {
	style := lipgloss.NewStyle().
		Width(p.optionListWidth()).
		Align(lipgloss.Center)

	var (
		splitOptionsView  string
		leagueOptionsView string
		stageOptionsView  string
	)
	switch p.state {
	case standingsPageStateLoadingSplits:
		splitOptionsView = style.Render(p.viewSpinner())

	case standingsPageStateSplitSelection:
		splitOptionsView = style.Render(p.splitOptions.View())

	case standingsPageStateLeagueSelection:
		splitOptionsView = style.Render(p.splitOptions.View())
		leagueOptionsView = style.Render(p.leagueOptions.View())

	case standingsPageStateLoadingStages:
		splitOptionsView = style.Render(p.splitOptions.View())
		leagueOptionsView = style.Render(p.leagueOptions.View())
		stageOptionsView = style.Render(p.viewSpinner())

	case standingsPageStateStageSelection:
		splitOptionsView = style.Render(p.splitOptions.View())
		leagueOptionsView = style.Render(p.leagueOptions.View())
		stageOptionsView = style.Render(p.stageOptions.View())
	}

	allOptionLists := lipgloss.JoinHorizontal(
		lipgloss.Top,
		splitOptionsView,
		leagueOptionsView,
		stageOptionsView,
	)

	return p.styles.doc.Render(allOptionLists)
}

func (p *standingsPage) viewSpinner() string {
	return p.spinner.View() + " Wukong is preparing your data…"
}

func (p *standingsPage) SetSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v
}

func (p *standingsPage) isLoading() bool {
	return p.state == standingsPageStateLoadingSplits ||
		p.state == standingsPageStateLoadingStages
}

func (p *standingsPage) handleSelection() (tea.Model, tea.Cmd) {
	switch p.state {
	case standingsPageStateSplitSelection:
		return p.selectSplit()
	case standingsPageStateLeagueSelection:
		return p.selectLeague()
	case standingsPageStateStageSelection:
		return p.selectStage()
	default:
		return p, nil
	}
}

func (p *standingsPage) selectSplit() (tea.Model, tea.Cmd) {
	p.state = standingsPageStateLeagueSelection

	selectedSplit := p.splits[p.splitOptions.Index()]
	p.leagues = listLeaguesFromTournaments(selectedSplit.Tournaments)
	p.leagueOptions = newLeagueOptionsList(p.leagues, p.optionListWidth(), p.height)

	return p, nil
}

func (p *standingsPage) selectLeague() (tea.Model, tea.Cmd) {
	if standings, ok := p.standingsFromCache(); ok {
		p.state = standingsPageStateStageSelection
		p.stages = listStagesFromStandings(standings)
		p.stageOptions = newStageOptionsList(p.stages, p.optionListWidth(), p.height)
		return p, nil
	}

	p.state = standingsPageStateLoadingStages

	selectedSplit := p.splits[p.splitOptions.Index()]
	selectedLeague := p.leagues[p.leagueOptions.Index()]
	tournamentIDs := listTournamentIDsForLeague(selectedSplit.Tournaments, selectedLeague.ID)
	return p, tea.Batch(p.spinner.Tick, p.fetchStandings(tournamentIDs))
}

func (p *standingsPage) selectStage() (tea.Model, tea.Cmd) {
	p.state = standingsPageStateShowStandings

	selectedStage := p.stages[p.stageOptions.Index()]
	p.standings = newStandingsViewport(selectedStage, p.width, p.height)

	return p, nil
}

func (p *standingsPage) setStandingsInCache(standings []lolesports.Standings) {
	selectedSplit := p.splits[p.splitOptions.Index()]
	selectedLeague := p.leagues[p.leagueOptions.Index()]
	cacheKey := makeStandingsCacheKey(selectedSplit.ID, selectedLeague.ID)
	p.standingsCache[cacheKey] = standings
}

func (p *standingsPage) standingsFromCache() ([]lolesports.Standings, bool) {
	selectedSplit := p.splits[p.splitOptions.Index()]
	selectedLeague := p.leagues[p.leagueOptions.Index()]
	cacheKey := makeStandingsCacheKey(selectedSplit.ID, selectedLeague.ID)
	standings, ok := p.standingsCache[cacheKey]
	return standings, ok
}

func (p *standingsPage) goToPreviousStep() {
	switch p.state {
	case standingsPageStateLeagueSelection:
		p.state = standingsPageStateSplitSelection
		p.leagueOptions = list.Model{}

	case standingsPageStateStageSelection:
		p.state = standingsPageStateLeagueSelection
		p.stageOptions = list.Model{}

	case standingsPageStateShowStandings:
		p.state = standingsPageStateSplitSelection
		p.leagueOptions = list.Model{}
		p.stageOptions = list.Model{}
	}
}

func (p *standingsPage) optionListWidth() int {
	return p.width / selectionListCount
}

type fetchedStandingsMessage struct {
	standings []lolesports.Standings
}

func (p *standingsPage) fetchStandings(tournamentIDs []string) tea.Cmd {
	return func() tea.Msg {
		standings, err := p.lolesportsClient.GetStandings(context.Background(), tournamentIDs)
		if err != nil {
			return fetchErrorMessage{err}
		}
		return fetchedStandingsMessage{standings}
	}
}

type fetchedCurrentSeasonSplitsMessage struct {
	splits []lolesports.Split
}

func (p *standingsPage) fetchCurrentSeasonSplits() tea.Cmd {
	return func() tea.Msg {
		splits, err := p.lolesportsClient.GetCurrentSeasonSplits(context.Background())
		if err != nil {
			return fetchErrorMessage{err}
		}
		return fetchedCurrentSeasonSplitsMessage{splits}
	}
}

func listLeaguesFromTournaments(tournaments []lolesports.Tournament) []lolesports.League {
	var (
		leagues     []lolesports.League
		seenLeagues = map[string]bool{}
	)
	for _, tournament := range tournaments {
		if _, ok := seenLeagues[tournament.League.ID]; !ok {
			leagues = append(leagues, tournament.League)
			seenLeagues[tournament.League.ID] = true
		}
	}
	return leagues
}

func listTournamentIDsForLeague(tournaments []lolesports.Tournament, leagueID string) []string {
	var tournamentIDs []string
	for _, tournament := range tournaments {
		if tournament.League.ID == leagueID {
			tournamentIDs = append(tournamentIDs, tournament.ID)
		}
	}
	return tournamentIDs
}

func listStagesFromStandings(standings []lolesports.Standings) []lolesports.Stage {
	var stages []lolesports.Stage
	for _, standing := range standings {
		for _, stage := range standing.Stages {
			stages = append(stages, stage)
		}
	}
	return stages
}

func makeStandingsCacheKey(splitID, leagueID string) string {
	return fmt.Sprintf("%s-%s", splitID, leagueID)
}
