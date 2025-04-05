package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matthieugusmini/lolesport/internal/lolesport"
)

type state int

const (
	stateUnknown state = iota
	stateShowSchedule
	stateShowStandings
)

type LoLEsportClient interface {
	GetSchedule(ctx context.Context, opts lolesport.GetScheduleOptions) (*lolesport.Schedule, error)
	GetStandings(ctx context.Context, tournamentID string) ([]*lolesport.Standings, error)
	GetLeagues(ctx context.Context) ([]*lolesport.League, error)
	GetTournamentsForLeague(ctx context.Context, leagueID string) ([]*lolesport.Tournament, error)
}

type Model struct {
	headers        tea.Model
	scheduleModel  tea.Model
	standingsModel tea.Model
	state          state

	lolesportClient LoLEsportClient
}

func NewModel(lolesportClient LoLEsportClient) Model {
	return Model{
		// TODO: Probably inject the UI components as dependencies.
		headers:        newHeadersModel(),
		scheduleModel:  newScheduleModel(lolesportClient),
		standingsModel: newStandingsModel(lolesportClient),
		state:          stateShowSchedule,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.scheduleModel.(*scheduleModel).setSize(msg.Width-h, msg.Height-v-headersHeight)
		m.standingsModel.(*standingsModel).setSize(msg.Width-h, msg.Height-v-headersHeight)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case state:
		m.state = msg
	}

	headers, cmd := m.headers.Update(msg)
	m.headers = headers
	cmds = append(cmds, cmd)

	switch m.state {
	case stateShowSchedule:
		scheduleModel, cmd := m.scheduleModel.Update(msg)
		m.scheduleModel = scheduleModel
		cmds = append(cmds, cmd)
	case stateShowStandings:
		standingsModel, cmd := m.standingsModel.Update(msg)
		m.standingsModel = standingsModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(m.headers.View())
	sb.WriteString("\n")

	switch m.state {
	case stateShowSchedule:
		sb.WriteString(m.scheduleModel.View())
	case stateShowStandings:
		sb.WriteString(m.standingsModel.View())
	}

	return sb.String()
}
