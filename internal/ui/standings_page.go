package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/internal/rift"
)

const (
	selectionListCount        = 3
	standingsViewHeaderHeight = 5
)

type standingsPageState int

const (
	standingsPageStateLoadingSplits standingsPageState = iota
	standingsPageStateSplitSelection
	standingsPageStateLeagueSelection
	standingsPageStateLoadingStages
	standingsPageStateStageSelection
	standingsPageStateShowStandings
	standingsPageStateShowBracket
)

type standingsStyles struct {
	doc              lipgloss.Style
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
	separator        lipgloss.Style
}

func newDefaultStandingsStyles() (s standingsStyles) {
	s.doc = lipgloss.NewStyle().Padding(1, 2)

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

	return s
}

type standingsPage struct {
	lolesportsClient      LoLEsportsClient
	bracketTemplateLoader BracketTemplateLoader

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

	// Bracket page
	bracket BracketModel

	err error

	spinner *wukongSpinner

	height, width int

	styles standingsStyles
}

func newStandingsPage(
	lolesportsClient LoLEsportsClient,
	bracketLoader BracketTemplateLoader,
) *standingsPage {
	return &standingsPage{
		lolesportsClient:      lolesportsClient,
		bracketTemplateLoader: bracketLoader,
		styles:                newDefaultStandingsStyles(),
		standingsCache:        map[string][]lolesports.Standings{},
		spinner:               newWukongSpinner(),
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
		p.spinner.refreshQuote()
		p.setStandingsInCache(msg.standings)
		p.stages = listStagesFromStandings(msg.standings)
		p.stageOptions = newStageOptionsList(p.stages, p.optionListWidth(), p.height)

	case fetchedCurrentSeasonSplitsMessage:
		p.state = standingsPageStateSplitSelection
		p.spinner.refreshQuote()
		p.splits = msg.splits
		p.splitOptions = newSplitOptionsList(p.splits, p.optionListWidth(), p.height)

	case fetchedBracketStageTemplateMessage:
		p.state = standingsPageStateShowBracket
		selectedStage := p.stages[p.stageOptions.Index()]
		p.bracket = NewBracketModel(msg.template, selectedStage, p.width, p.height)

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
	if p.state == standingsPageStateShowBracket {
		return p.styles.doc.Render(p.bracket.View())
	}
	if p.state == standingsPageStateShowStandings {
		return p.viewStandings()
	}
	return p.viewSelection()
}

func (p *standingsPage) viewStandings() string {
	selectedSplit := p.splits[p.splitOptions.Index()]
	selectedLeague := p.leagues[p.leagueOptions.Index()]
	stageName := p.styles.stageName.Render(
		fmt.Sprintf("%s: %s Standings", selectedSplit.Name, selectedLeague.Name),
	)

	tournamentState := computeTournamentState(selectedSplit.StartTime, selectedSplit.EndTime)
	tournamentPeriod := formatTournamentPeriod(selectedSplit.StartTime, selectedSplit.EndTime)
	tournamentType := selectedSplit.Region
	stageInfo := strings.Join(
		[]string{
			p.styles.tournamentState.Render(string(tournamentState)),
			p.styles.tournamentPeriod.Render(tournamentPeriod),
			p.styles.tournamentType.Render(tournamentType),
		},
		separatorBullet,
	)

	sep := p.styles.separator.Render(strings.Repeat(separatorLine, p.width))

	header := fmt.Sprintf("%s\n\n%s\n%s\n", stageName, stageInfo, sep)
	return p.styles.doc.Render(header + p.standings.View())
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
		splitOptionsView = style.Render(p.spinner.View())

	case standingsPageStateSplitSelection:
		splitOptionsView = style.Render(p.splitOptions.View())

	case standingsPageStateLeagueSelection:
		splitOptionsView = style.Render(p.splitOptions.View())
		leagueOptionsView = style.Render(p.leagueOptions.View())

	case standingsPageStateLoadingStages:
		splitOptionsView = style.Render(p.splitOptions.View())
		leagueOptionsView = style.Render(p.leagueOptions.View())
		stageOptionsView = style.Render(p.spinner.View())

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
	selectedStage := p.stages[p.stageOptions.Index()]

	stageType := getStageType(selectedStage)
	switch stageType {
	case stageTypeGroups:
		p.state = standingsPageStateShowStandings
		p.standings = newStandingsViewport(
			selectedStage,
			p.width,
			p.height-standingsViewHeaderHeight,
		)

	case stageTypeBracket:
		return p, p.fetchBracketStageTemplate(selectedStage.ID)
	}

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

	case standingsPageStateShowStandings, standingsPageStateShowBracket:
		p.state = standingsPageStateStageSelection
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

type fetchedBracketStageTemplateMessage struct {
	template rift.BracketTemplate
}

func (p *standingsPage) fetchBracketStageTemplate(stageID string) tea.Cmd {
	return func() tea.Msg {
		tmpl, err := p.bracketTemplateLoader.Load(context.Background(), stageID)
		if err != nil {
			return fetchErrorMessage{err}
		}
		return fetchedBracketStageTemplateMessage{tmpl}
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
		stages = append(stages, standing.Stages...)
	}
	return stages
}

func makeStandingsCacheKey(splitID, leagueID string) string {
	return fmt.Sprintf("%s-%s", splitID, leagueID)
}
