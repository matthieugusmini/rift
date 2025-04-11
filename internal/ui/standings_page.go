package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
)

type standingsStyles struct {
	docStyle         lipgloss.Style
	tableBorder      lipgloss.Style
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
}

func newDefaultStandingsStyles() (s standingsStyles) {
	s.docStyle = lipgloss.NewStyle().Padding(1, 2)

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

type standingsPageState int

const (
	standingsPageStateSplitSelection standingsPageState = iota
	standingsPageStateLeagueSelection
	standingsPageStateStageSelection
	standingsPageStateShowStandings
)

type standingsPage struct {
	lolesportsClient LoLEsportClient

	state standingsPageState

	currentSeason            *lolesports.Season
	selectedSplitTournaments []*lolesports.Tournament
	selectedStandings        []*lolesports.Standings

	splitChoices  list.Model
	leagueChoices list.Model
	stageChoices  list.Model

	splitName     string
	startDate     time.Time
	endDate       time.Time
	standings     viewport.Model
	height, width int

	styles standingsStyles
}

func newStandingsPage(lolesportsClient LoLEsportClient) *standingsPage {
	return &standingsPage{
		lolesportsClient: lolesportsClient,
		styles:           newDefaultStandingsStyles(),
	}
}

func (m *standingsPage) Init() tea.Cmd {
	return m.getSeasons()
}

func (m *standingsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			switch m.state {
			case standingsPageStateLeagueSelection:
				m.state = standingsPageStateSplitSelection
			case standingsPageStateStageSelection:
				m.state = standingsPageStateLeagueSelection
			case standingsPageStateShowStandings:
				m.state = standingsPageStateSplitSelection
			}

		// TODO: Move to item delegate
		case "enter":
			switch m.state {
			case standingsPageStateSplitSelection:
				m.state = standingsPageStateLeagueSelection

				var (
					leagues           []*lolesports.League
					selectedSplit     = m.splitChoices.SelectedItem().(splitItem)
					alreadySeenLeague = make(map[string]bool)
				)
				m.selectedSplitTournaments = selectedSplit.tournaments
				for _, tournament := range selectedSplit.tournaments {
					if _, ok := alreadySeenLeague[tournament.League.ID]; !ok {
						leagues = append(leagues, &tournament.League)
						alreadySeenLeague[tournament.League.ID] = true
					}
				}
				m.leagueChoices = newLeagueChoicesList(leagues, m.width/3, m.height)

				m.splitName = selectedSplit.name
				m.startDate = selectedSplit.startTime
				m.endDate = selectedSplit.endTime

			case standingsPageStateLeagueSelection:
				m.state = standingsPageStateStageSelection

				selectedLeague := m.leagueChoices.SelectedItem().(leagueItem)
				var tournamentIDs []string
				for _, tournament := range m.selectedSplitTournaments {
					if tournament.League.ID == selectedLeague.id {
						tournamentIDs = append(tournamentIDs, tournament.ID)
					}
				}
				return m, m.getStandings(tournamentIDs)
			case standingsPageStateStageSelection:
				m.state = standingsPageStateShowStandings

				var stage lolesports.Stage
				for _, standing := range m.selectedStandings {
					for _, s := range standing.Stages {
						if m.stageChoices.SelectedItem().(stageItem).id == s.ID {
							stage = s
						}
					}
				}

				m.standings = newStandingsViewport(stage, m.width, m.height-navigationBarHeight)
				// m.table = newStandingsViewport(rankings, m.width)
			}
		}

	case tea.WindowSizeMsg:
		if m.currentSeason == nil {
			return m, m.getSeasons()
		}
		return m, nil

	case fetchedStandingsMessage:
		m.selectedStandings = msg.standings
		m.stageChoices = newStageChoices(msg.standings, m.width/3, m.height)

	case fetchedCurrentSeasonMessage:
		m.currentSeason = msg.currentSeason
		m.splitChoices = newSplitChoices(msg.currentSeason.Splits, m.width/3, m.height)
	}

	var cmds []tea.Cmd

	if m.state == standingsPageStateSplitSelection {
		m.splitChoices, cmd = m.splitChoices.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == standingsPageStateLeagueSelection {
		m.leagueChoices, cmd = m.leagueChoices.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == standingsPageStateStageSelection {
		m.stageChoices, cmd = m.stageChoices.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == standingsPageStateShowStandings {
		m.standings, cmd = m.standings.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *standingsPage) View() string {
	if m.state == standingsPageStateShowStandings {
		return m.styles.tableBorder.Render(m.standings.View())
	}

	var (
		splitChoices  string
		leagueChoices string
		stageChoices  string

		style = lipgloss.NewStyle().
			Width(m.width / 3).
			Align(lipgloss.Center)
	)
	switch m.state {
	case standingsPageStateSplitSelection:
		splitChoices = style.Render(m.splitChoices.View())
	case standingsPageStateLeagueSelection:
		splitChoices = style.Render(m.splitChoices.View())
		leagueChoices = style.Render(m.leagueChoices.View())
	case standingsPageStateStageSelection:
		splitChoices = style.Render(m.splitChoices.View())
		leagueChoices = style.Render(m.leagueChoices.View())
		stageChoices = style.Render(m.stageChoices.View())
	}

	selectionLists := lipgloss.JoinHorizontal(lipgloss.Center, splitChoices, leagueChoices, stageChoices)

	return m.styles.docStyle.Render(selectionLists)
}

func (m *standingsPage) SetSize(width, height int) {
	h, v := m.styles.docStyle.GetFrameSize()
	m.width, m.height = width-h, height-v
}

type fetchedStandingsMessage struct {
	standings []*lolesports.Standings
}

func (m *standingsPage) getStandings(tournamentIDs []string) tea.Cmd {
	return func() tea.Msg {
		standings, err := m.lolesportsClient.GetStandings(context.Background(), tournamentIDs)
		if err != nil {
			return err
		}
		return fetchedStandingsMessage{standings}
	}
}

type fetchedCurrentSeasonMessage struct {
	currentSeason *lolesports.Season
}

func (m *standingsPage) getSeasons() tea.Cmd {
	return func() tea.Msg {
		currentSeason, err := m.lolesportsClient.GetCurrentSeason(context.Background())
		if err != nil {
			return err
		}
		return fetchedCurrentSeasonMessage{currentSeason}
	}
}
