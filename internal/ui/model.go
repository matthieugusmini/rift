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
}

type Model struct {
	headers       tea.Model
	scheduleModel tea.Model
	state         state

	lolesportClient LoLEsportClient
}

func NewModel(lolesportClient LoLEsportClient) Model {
	return Model{
		// TODO: Probably inject the UI components as dependencies.
		headers:       newHeadersModel(),
		scheduleModel: newScheduleModel(lolesportClient),
		state:         stateShowSchedule,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
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
		// TODO: Implement the standings model.
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(m.headers.View())
	sb.WriteString("\n")
	sb.WriteString(m.scheduleModel.View())
	return sb.String()
}
