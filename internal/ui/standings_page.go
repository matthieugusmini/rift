package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
)

const selectionListCount = 3

type standingsPageState int

const (
	standingsPageStateSplitSelection standingsPageState = iota
	standingsPageStateLeagueSelection
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
		Foreground(white).
		Bold(true)

	s.tournamentState = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("#000000")).
		Background(white)

	s.tournamentPeriod = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(white)

	s.tournamentType = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(white)

	s.tableBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return s
}

type standingsPage struct {
	lolesportsClient LoLEsportClient

	state standingsPageState

	standingsCache map[string][]*lolesports.Standings
	
	// Selection phase
	splitOptions  list.Model
	leagueOptions list.Model
	stageOptions  list.Model

	// Standings page
	splitName string
	startDate time.Time
	endDate   time.Time
	standings viewport.Model

	errorMessage string

	height, width int

	styles standingsStyles
}

func newStandingsPage(lolesportsClient LoLEsportClient) *standingsPage {
	return &standingsPage{
		lolesportsClient: lolesportsClient,
		styles:           newDefaultStandingsStyles(),
		standingsCache:   map[string][]*lolesports.Standings{},
	}
}

func (p *standingsPage) Init() tea.Cmd {
	return p.fetchCurrentSeasonSplits()
}

func (p *standingsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			p.goToPreviousStep()

		case "enter":
			return p.handleSelection()
		}

	case tea.WindowSizeMsg:
		if !p.hasLoadedSplits() {
			return p, p.fetchCurrentSeasonSplits()
		}
		return p, nil

	case fetchedStandingsMessage:
		splitID := p.splitOptions.SelectedItem().(splitItem).id
		leagueID := p.leagueOptions.SelectedItem().(leagueItem).id
		cacheKey := makeStandingsCacheKey(splitID, leagueID)
		p.standingsCache[cacheKey] = msg.standings

		p.stageOptions = newStageOptionsList(msg.standings, p.optionListWidth(), p.height)
		return p, nil

	case fetchedCurrentSeasonSplitsMessage:
		p.splitOptions = newSplitOptionsList(msg.splits, p.optionListWidth(), p.height)
		return p, nil

	case fetchErrorMessage:
		// TODO: add proper error UI handling
		p.errorMessage = msg.err.Error()
		return p, nil
	}

	if !p.hasLoadedSplits() {
		return p, nil
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
	return p, cmd
}

func (p *standingsPage) View() string {
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
	case standingsPageStateSplitSelection:
		splitOptionsView = style.Render(p.splitOptions.View())
	case standingsPageStateLeagueSelection:
		splitOptionsView = style.Render(p.splitOptions.View())
		leagueOptionsView = style.Render(p.leagueOptions.View())
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

func (p *standingsPage) SetSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v
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

	selectedSplit := p.splitOptions.SelectedItem().(splitItem)
	leagues := listLeaguesFromTournaments(selectedSplit.tournaments)
	p.leagueOptions = newLeagueOptionsList(leagues, p.optionListWidth(), p.height)

	p.splitName = selectedSplit.name
	p.startDate = selectedSplit.startTime
	p.endDate = selectedSplit.endTime

	return p, nil
}

func (p *standingsPage) selectLeague() (tea.Model, tea.Cmd) {
	p.state = standingsPageStateStageSelection

	selectedSplit := p.splitOptions.SelectedItem().(splitItem)
	selectedLeague := p.leagueOptions.SelectedItem().(leagueItem)
	cacheKey := makeStandingsCacheKey(selectedSplit.id, selectedLeague.id)
	if standings, ok := p.standingsCache[cacheKey]; ok {
		p.stageOptions = newStageOptionsList(standings, p.optionListWidth(), p.height)
		return p, nil
	}

	tournamentIDs := listTournamentIDsForLeague(selectedSplit.tournaments, selectedLeague.id)
	return p, p.fetchStandings(tournamentIDs)
}

func (p *standingsPage) selectStage() (tea.Model, tea.Cmd) {
	p.state = standingsPageStateShowStandings

	selectedSplit := p.splitOptions.SelectedItem().(splitItem)
	selectedLeague := p.leagueOptions.SelectedItem().(leagueItem)
	cacheKey := makeStandingsCacheKey(selectedSplit.id, selectedLeague.id)
	standings := p.standingsCache[cacheKey]

	selectedStage := p.stageOptions.SelectedItem().(stageItem)

	stage := findStageByID(standings, selectedStage.id)
	p.standings = newStandingsViewport(stage, p.width, p.height)

	return p, nil
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

func (p *standingsPage) hasLoadedSplits() bool {
	return len(p.splitOptions.Items()) > 0
}

func (p *standingsPage) optionListWidth() int {
	return p.width / selectionListCount
}

type fetchedStandingsMessage struct {
	standings []*lolesports.Standings
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
	splits []*lolesports.Split
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

func listLeaguesFromTournaments(tournaments []*lolesports.Tournament) []*lolesports.League {
	var (
		leagues     []*lolesports.League
		seenLeagues = map[string]bool{}
	)
	for _, tournament := range tournaments {
		if _, ok := seenLeagues[tournament.League.ID]; !ok {
			leagues = append(leagues, &tournament.League)
			seenLeagues[tournament.League.ID] = true
		}
	}
	return leagues
}

func listTournamentIDsForLeague(tournaments []*lolesports.Tournament, leagueID string) []string {
	var tournamentIDs []string
	for _, tournament := range tournaments {
		if tournament.League.ID == leagueID {
			tournamentIDs = append(tournamentIDs, tournament.ID)
		}
	}
	return tournamentIDs
}

func findStageByID(standings []*lolesports.Standings, stageID string) lolesports.Stage {
	for _, standing := range standings {
		for _, s := range standing.Stages {
			if s.ID == stageID {
				return s
			}
		}
	}
	return lolesports.Stage{}
}

func makeStandingsCacheKey(splitID, leagueID string) string {
	return fmt.Sprintf("%s-%s", splitID, leagueID)
}
