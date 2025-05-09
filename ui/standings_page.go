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
	selectionListCount       = 3
	minListHeight            = 18
	minSelectionPromptHeight = 3

	rankingViewHeaderHeight = 5

	standingsPageShortHelpHeight = 1
	standingsPageFullHelpHeight  = 6
)

const (
	captionSelectSplit  = "SELECT A SPLIT"
	captionSelectLeague = "SELECT A LEAGUE"
	captionSelectStage  = "SELECT A STAGE"
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
	caption          lipgloss.Style
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

	s.caption = lipgloss.NewStyle().
		Foreground(textPrimaryColor).
		Bold(true)

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
	lolesportsClient      LoLEsportsLoader
	bracketTemplateLoader BracketTemplateLoader

	state standingsPageState

	splits  []lolesports.Split
	leagues []lolesports.League
	stages  []lolesports.Stage

	splitOptions  list.Model
	leagueOptions list.Model
	stageOptions  list.Model

	ranking viewport.Model
	bracket *bracketModel

	err error

	spinner spinner.Model

	keyMap standingsPageKeyMap
	help   help.Model

	height, width int

	styles standingsStyles
}

func newStandingsPage(
	lolesportsClient LoLEsportsLoader,
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

	case loadedStandingsMessage:
		p.state = standingsPageStateStageSelection
		p.stages = listStagesFromStandings(msg.standings)
		p.stageOptions = newStageOptionsList(p.stages, p.listWidth(), p.listHeight())

	case fetchedCurrentSeasonSplitsMessage:
		p.state = standingsPageStateSplitSelection
		p.splits = msg.splits
		p.splitOptions = newSplitOptionsList(p.splits, p.listWidth(), p.listHeight())

	case loadedBracketStageTemplateMessage:
		p.state = standingsPageStateShowBracket
		selectedStage := p.stages[p.stageOptions.Index()]
		p.bracket = newBracketModel(msg.template, selectedStage, p.width, p.contentHeight())

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
	case standingsPageStateShowBracket:
		p.bracket, cmd = p.bracket.Update(msg)
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

		showPrompt := p.contentHeight() >= minListHeight+minSelectionPromptHeight
		if showPrompt {
			sections = append(sections, p.viewSelectionPrompt())
		}

	case standingsPageStateShowBracket:
		sections = append(sections, p.viewBracket())

	case standingsPageStateShowRanking:
		sections = append(sections, p.viewRanking())
	}

	sections = append(sections, p.viewHelp())

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return p.styles.doc.Render(view)
}

func (p *standingsPage) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Select,
		p.keyMap.NextPage,
		p.keyMap.Quit,
		p.keyMap.ShowFullHelp,
	}
}

func (p *standingsPage) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			p.keyMap.Up,
			p.keyMap.Down,
			p.keyMap.Select,
			p.keyMap.Previous,
			p.keyMap.NextPage,
			p.keyMap.PrevPage,
		},
		{
			p.keyMap.Quit,
			p.keyMap.CloseFullHelp,
		},
	}
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
	listStyle := lipgloss.NewStyle().
		Width(p.listWidth()).
		Height(p.listHeight()).
		Align(lipgloss.Center)

	var (
		splitOptionsView  string
		leagueOptionsView string
		stageOptionsView  string
	)
	switch p.state {
	case standingsPageStateLoadingSplits:
		splitOptionsView = listStyle.Render(p.spinner.View())

	case standingsPageStateSplitSelection:
		splitOptionsView = listStyle.Render(p.splitOptions.View())

	case standingsPageStateLeagueSelection:
		splitOptionsView = listStyle.Render(p.splitOptions.View())
		leagueOptionsView = listStyle.Render(p.leagueOptions.View())

	case standingsPageStateLoadingStages:
		splitOptionsView = listStyle.Render(p.splitOptions.View())
		leagueOptionsView = listStyle.Render(p.leagueOptions.View())
		stageOptionsView = listStyle.Render(p.spinner.View())

	case standingsPageStateStageSelection:
		splitOptionsView = listStyle.Render(p.splitOptions.View())
		leagueOptionsView = listStyle.Render(p.leagueOptions.View())
		stageOptionsView = listStyle.Render(p.stageOptions.View())
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		splitOptionsView,
		leagueOptionsView,
		stageOptionsView,
	)
}

func (p *standingsPage) viewSelectionPrompt() string {
	promptHeight := p.contentHeight() - p.listHeight()

	var prompt string

	switch p.state {
	case standingsPageStateSplitSelection:
		prompt = p.styles.caption.Render(captionSelectSplit)
	case standingsPageStateLeagueSelection:
		prompt = p.styles.caption.Render(captionSelectLeague)
	case standingsPageStateStageSelection:
		prompt = p.styles.caption.Render(captionSelectStage)
	}

	return lipgloss.Place(
		p.width,
		promptHeight,
		lipgloss.Center,
		lipgloss.Center,
		prompt,
	)
}

func (p *standingsPage) viewHelp() string {
	return p.styles.help.Render(p.help.View(p))
}

func (p *standingsPage) setSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v

	p.help.Width = p.width

	switch p.state {
	case standingsPageStateSplitSelection:
		p.splitOptions.SetSize(p.listSize())

	case standingsPageStateLeagueSelection:
		listWidth, listHeight := p.listSize()
		p.splitOptions.SetSize(listWidth, listHeight)
		p.leagueOptions.SetSize(listWidth, listHeight)

	case standingsPageStateStageSelection:
		listWidth, listHeight := p.listSize()
		p.splitOptions.SetSize(listWidth, listHeight)
		p.leagueOptions.SetSize(listWidth, listHeight)
		p.stageOptions.SetSize(listWidth, listHeight)

	case standingsPageStateShowRanking:
		selectedStage := p.stages[p.stageOptions.Index()]
		p.ranking = newRankingViewport(
			selectedStage,
			p.width,
			p.contentHeight()-rankingViewHeaderHeight,
		)

	case standingsPageStateShowBracket:
		p.bracket.setSize(p.width, p.contentHeight())
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
	p.leagueOptions = newLeagueOptionsList(p.leagues, p.listWidth(), p.listHeight())

	return p, nil
}

func (p *standingsPage) selectLeague() (tea.Model, tea.Cmd) {
	p.state = standingsPageStateLoadingStages

	selectedSplit := p.splits[p.splitOptions.Index()]
	selectedLeague := p.leagues[p.leagueOptions.Index()]
	tournamentIDs := listTournamentIDsForLeague(selectedSplit.Tournaments, selectedLeague.ID)

	return p, tea.Batch(p.spinner.Tick, p.loadStandings(tournamentIDs))
}

func (p *standingsPage) selectStage() (tea.Model, tea.Cmd) {
	selectedStage := p.stages[p.stageOptions.Index()]

	stageType := getStageType(selectedStage)
	switch stageType {
	case stageTypeGroups:
		p.ranking = newRankingViewport(
			selectedStage,
			p.width,
			p.contentHeight()-rankingViewHeaderHeight,
		)
		p.state = standingsPageStateShowRanking

	case stageTypeBracket:
		return p, p.loadBracketStageTemplate(selectedStage.ID)
	}

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

	case standingsPageStateShowRanking, standingsPageStateShowBracket:
		p.state = standingsPageStateStageSelection
	}
}

func (p *standingsPage) toggleFullHelp() {
	p.help.ShowAll = !p.help.ShowAll
	p.updateContentViewHeight()
}

func (p *standingsPage) updateContentViewHeight() {
	listHeight := p.listHeight()

	switch p.state {
	case standingsPageStateSplitSelection:
		p.splitOptions.SetHeight(listHeight)

	case standingsPageStateLeagueSelection:
		p.splitOptions.SetHeight(listHeight)
		p.leagueOptions.SetHeight(listHeight)

	case standingsPageStateStageSelection:
		p.splitOptions.SetHeight(listHeight)
		p.leagueOptions.SetHeight(listHeight)
		p.stageOptions.SetHeight(listHeight)

	case standingsPageStateShowRanking:
		selectedStage := p.stages[p.stageOptions.Index()]
		p.ranking = newRankingViewport(
			selectedStage,
			p.width,
			p.contentHeight()-rankingViewHeaderHeight,
		)

	case standingsPageStateShowBracket:
		p.bracket.setSize(p.width, p.contentHeight())
	}
}

func (p *standingsPage) contentHeight() int {
	return p.height - p.helpHeight()
}

func (p *standingsPage) helpHeight() int {
	padding := p.styles.help.GetVerticalPadding()
	if p.help.ShowAll {
		return standingsPageFullHelpHeight + padding
	}
	return standingsPageShortHelpHeight + padding
}

func (p *standingsPage) listSize() (width, height int) {
	return p.listWidth(), p.listHeight()
}

func (p *standingsPage) listWidth() int {
	return p.width / selectionListCount
}

func (p *standingsPage) listHeight() int {
	showMessage := p.contentHeight() >= minListHeight+minSelectionPromptHeight
	if showMessage {
		return max(p.contentHeight()/2, minListHeight)
	} else {
		return p.contentHeight()
	}
}

type loadedStandingsMessage struct {
	standings []lolesports.Standings
}

func (p *standingsPage) loadStandings(tournamentIDs []string) tea.Cmd {
	return func() tea.Msg {
		standings, err := p.lolesportsClient.LoadStandingsByTournamentIDs(
			context.Background(),
			tournamentIDs,
		)
		if err != nil {
			return fetchErrorMessage{err}
		}
		return loadedStandingsMessage{standings}
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

type loadedBracketStageTemplateMessage struct {
	template rift.BracketTemplate
}

func (p *standingsPage) loadBracketStageTemplate(stageID string) tea.Cmd {
	return func() tea.Msg {
		tmpl, err := p.bracketTemplateLoader.Load(context.Background(), stageID)
		if err != nil {
			return fetchErrorMessage{err}
		}
		return loadedBracketStageTemplateMessage{tmpl}
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
