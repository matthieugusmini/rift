package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	headersHeight = 2
)

type headersStyles struct {
	lolStyle    lipgloss.Style
	esportStyle lipgloss.Style

	normalHeaderStyle   lipgloss.Style
	selectedHeaderStyle lipgloss.Style

	separatorStyle lipgloss.Style
}

func newDefaultHeadersStyles() headersStyles {
	lolStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)
	esportStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cyan)).
		Bold(true)

	normalHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Faint(true)
	selectedHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(gold)).
		Bold(true)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)

	return headersStyles{
		lolStyle:    lolStyle,
		esportStyle: esportStyle,

		normalHeaderStyle:   normalHeaderStyle,
		selectedHeaderStyle: selectedHeaderStyle,

		separatorStyle: separatorStyle,
	}
}

type headersModel struct {
	headers []string
	cursor  int
	width   int
	styles  headersStyles
}

func newHeadersModel() *headersModel {
	return &headersModel{
		headers: []string{"Schedule", "Standings"},
		styles:  newDefaultHeadersStyles(),
	}
}

func (m *headersModel) Init() tea.Cmd { return nil }

func (m *headersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	states := []state{
		stateShowSchedule,
		stateShowStandings,
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "shift+tab":
			m.moveCursorLeft()
			return m, changeSection(states[m.cursor])
		case "tab":
			m.moveCursorRight()
			return m, changeSection(states[m.cursor])
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	return m, nil
}

func (m *headersModel) View() string {
	lol := m.styles.lolStyle.Render("lol")
	esport := m.styles.esportStyle.Render("esport")
	lolesport := lipgloss.NewStyle().Padding(0, 1).Render(lol + esport)

	headers := make([]string, len(m.headers))
	for i, header := range m.headers {
		isSelected := m.cursor == i
		if isSelected {
			headers[i] = m.styles.selectedHeaderStyle.Render(header)
		} else {
			headers[i] = m.styles.normalHeaderStyle.Render(header)
		}
	}

	renderedHeaders := lipgloss.NewStyle().
		Width(m.width - lipgloss.Width(lolesport)*2).
		Align(lipgloss.Center).
		Render(strings.Join(headers, " • "))

	filler := lipgloss.NewStyle().
		Width(lipgloss.Width(lolesport)).
		Render(strings.Repeat(" ", lipgloss.Width(lolesport)))

	separator := m.styles.separatorStyle.
		Render(strings.Repeat("━", m.width))

	return fmt.Sprintf(
		"%s\n%s",
		lolesport+renderedHeaders+filler,
		separator,
	)
}

func (m *headersModel) moveCursorRight() {
	m.cursor = (m.cursor + 1) % len(m.headers)
}

func (m *headersModel) moveCursorLeft() {
	m.cursor = (m.cursor - 1 + len(m.headers)) % len(m.headers)
}

func changeSection(state state) tea.Cmd {
	return func() tea.Msg {
		return state
	}
}
