package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/rift"
)

const (
	selectionListCount = 3

	rankingViewHeaderHeight = 5

	standingsPageShortHelpHeight = 1
	standingsPageFullHelpHeight  = 6
)

type standingsPageState int

const (
	standingsPageStateLoadingSplits standingsPageState = iota
	standingsPageStateSplitSelection
	standingsPageStateLeagueSelection
	standingsPageStateLoadingStages
	standingsPageStateStageSelection
	standingsPageStateShowRanking
	standingsPageStateShowBracket
)

type standingsStyles struct {
	doc              lipgloss.Style
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
	separator        lipgloss.Style
	help             lipgloss.Style
	spinner          lipgloss.Style
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

	s.help = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	s.spinner = lipgloss.NewStyle().Foreground(spinnerColor)

	return s
}

type standingsPageKeyMap struct {
	Select        key.Binding
	Previous      key.Binding
	Up            key.Binding
	Down          key.Binding
	NextPage      key.Binding
	PrevPage      key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
	Quit          key.Binding
}

func (k standingsPageKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Select,
		k.ShowFullHelp,
	}
}

func (k standingsPageKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.Select,
			k.Previous,
			k.NextPage,
			k.PrevPage,
		},
		{
			k.Quit,
			k.CloseFullHelp,
		},
	}
}

func newDefaultStandingsPageKeyMap() standingsPageKeyMap {
	return standingsPageKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", "right"),
			key.WithHelp("enter/→", "select"),
		),
		Previous: key.NewBinding(
			key.WithKeys("esc", "left"),
			key.WithHelp("esc/←", "previous"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next page"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}

type standingsPage struct {
	lolesportsClient      LoLEsportsClient
	bracketTemplateLoader BracketTemplateLoader

	state standingsPageState

	splits                []lolesports.Split
	leagues               []lolesports.League
	stages                []lolesports.Stage
	standingsCache        map[string][]lolesports.Standings
	bracketTemplatesCache map[string]rift.BracketTemplate

	splitOptions  list.Model
	leagueOptions list.Model
	stageOptions  list.Model

	ranking viewport.Model
	bracket bracketModel

	err error

	spinner spinner.Model

	keyMap standingsPageKeyMap
	help   help.Model

	height, width int

	styles standingsStyles
}

func newStandingsPage(
	lolesportsClient LoLEsportsClient,
	bracketLoader BracketTemplateLoader,
) *standingsPage {
	styles := newDefaultStandingsStyles()

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(styles.spinner),
	)

	return &standingsPage{
		lolesportsClient:      lolesportsClient,
		bracketTemplateLoader: bracketLoader,
		styles:                styles,
		standingsCache:        map[string][]lolesports.Standings{},
		bracketTemplatesCache: map[string]rift.BracketTemplate{},
		spinner:               sp,
		keyMap:                newDefaultStandingsPageKeyMap(),
		help:                  help.New(),
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
		switch {
		case key.Matches(msg, p.keyMap.Quit):
			return p, tea.Quit

		case key.Matches(msg, p.keyMap.ShowFullHelp),
			key.Matches(msg, p.keyMap.CloseFullHelp):
			p.toggleFullHelp()

		case key.Matches(msg, p.keyMap.Previous):
			p.goToPreviousStep()

		case key.Matches(msg, p.keyMap.Select):
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
		p.stageOptions = newStageOptionsList(p.stages, p.optionListWidth(), p.height-p.helpHeight())

	case fetchedCurrentSeasonSplitsMessage:
		p.state = standingsPageStateSplitSelection
		p.splits = msg.splits
		p.splitOptions = newSplitOptionsList(p.splits, p.optionListWidth(), p.height-p.helpHeight())

	case fetchedBracketStageTemplateMessage:
		p.state = standingsPageStateShowBracket
		selectedStage := p.stages[p.stageOptions.Index()]
		p.bracketTemplatesCache[selectedStage.ID] = msg.template
		p.bracket = newBracketModel(msg.template, selectedStage, p.width, p.height-p.helpHeight())

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
	case standingsPageStateShowRanking:
		p.ranking, cmd = p.ranking.Update(msg)
	}
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *standingsPage) View() string {
	if p.width <= 0 {
		return ""
	}

	if p.err != nil {
		return p.err.Error()
	}

	var sections []string

	switch p.state {
	case standingsPageStateSplitSelection,
		standingsPageStateLeagueSelection,
		standingsPageStateStageSelection,
		standingsPageStateLoadingSplits,
		standingsPageStateLoadingStages:
		sections = append(sections, p.viewSelection())

	case standingsPageStateShowBracket:
		sections = append(sections, p.viewBracket())

	case standingsPageStateShowRanking:
		sections = append(sections, p.viewRanking())
	}

	sections = append(sections, p.viewHelp())

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return p.styles.doc.Render(view)
}

func (p *standingsPage) viewBracket() string {
	return p.bracket.View()
}

func (p *standingsPage) viewRanking() string {
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

	return lipgloss.JoinVertical(lipgloss.Left, header, p.ranking.View())
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

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		splitOptionsView,
		leagueOptionsView,
		stageOptionsView,
	)
}

func (p *standingsPage) viewHelp() string {
	return p.styles.help.Render(p.help.View(p.keyMap))
}

func (p *standingsPage) SetSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v

	p.help.Width = p.width

	switch p.state {
	case standingsPageStateShowRanking:
		selectedStage := p.stages[p.stageOptions.Index()]
		p.ranking = newRankingViewport(
			selectedStage,
			p.width,
			p.height-rankingViewHeaderHeight-p.helpHeight(),
		)

	case standingsPageStateShowBracket:
		selectedStage := p.stages[p.stageOptions.Index()]
		tmpl := p.bracketTemplatesCache[selectedStage.ID]
		p.bracket = newBracketModel(tmpl, selectedStage, p.width, p.height-p.helpHeight())
	}
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
	p.leagueOptions = newLeagueOptionsList(p.leagues, p.optionListWidth(), p.height-p.helpHeight())

	return p, nil
}

func (p *standingsPage) selectLeague() (tea.Model, tea.Cmd) {
	if standings, ok := p.standingsFromCache(); ok {
		p.state = standingsPageStateStageSelection
		p.stages = listStagesFromStandings(standings)
		p.stageOptions = newStageOptionsList(p.stages, p.optionListWidth(), p.height-p.helpHeight())
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
		p.ranking = newRankingViewport(
			selectedStage,
			p.width,
			p.height-rankingViewHeaderHeight-p.helpHeight(),
		)
		p.state = standingsPageStateShowRanking

	case stageTypeBracket:
		tmpl, ok := p.bracketTemplatesCache[selectedStage.ID]
		if !ok {
			return p, p.fetchBracketStageTemplate(selectedStage.ID)
		}

		p.bracket = newBracketModel(tmpl, selectedStage, p.width, p.height-p.helpHeight())
		p.state = standingsPageStateShowBracket
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

	case standingsPageStateShowRanking, standingsPageStateShowBracket:
		p.state = standingsPageStateStageSelection
	}
}

func (p *standingsPage) optionListWidth() int {
	return p.width / selectionListCount
}

func (p *standingsPage) contentViewHeight() int {
	return p.height - p.helpHeight()
}

func (p *standingsPage) toggleFullHelp() {
	p.help.ShowAll = !p.help.ShowAll
	p.updateCentralViewHeight()
}

func (p *standingsPage) updateCentralViewHeight() {
	switch p.state {
	case standingsPageStateSplitSelection:
		p.splitOptions.SetHeight(p.contentViewHeight())

	case standingsPageStateLeagueSelection:
		p.splitOptions.SetHeight(p.contentViewHeight())
		p.leagueOptions.SetHeight(p.contentViewHeight())

	case standingsPageStateStageSelection:
		p.splitOptions.SetHeight(p.contentViewHeight())
		p.leagueOptions.SetHeight(p.contentViewHeight())
		p.stageOptions.SetHeight(p.contentViewHeight())

	case standingsPageStateShowRanking:
		selectedStage := p.stages[p.stageOptions.Index()]
		p.ranking = newRankingViewport(
			selectedStage,
			p.width,
			p.contentViewHeight()-rankingViewHeaderHeight,
		)

	case standingsPageStateShowBracket:
		selectedStage := p.stages[p.stageOptions.Index()]
		tmpl := p.bracketTemplatesCache[selectedStage.ID]
		p.bracket = newBracketModel(tmpl, selectedStage, p.width, p.contentViewHeight())
	}
}

func (p *standingsPage) helpHeight() int {
	padding := p.styles.help.GetVerticalPadding()
	if p.help.ShowAll {
		return standingsPageFullHelpHeight + padding
	}
	return standingsPageShortHelpHeight + padding
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
