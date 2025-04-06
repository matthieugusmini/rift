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
	lol            lipgloss.Style
	esport         lipgloss.Style
	headers        lipgloss.Style
	normalHeader   lipgloss.Style
	selectedHeader lipgloss.Style
	separatorLine  lipgloss.Style
	filler         lipgloss.Style
}

func newDefaultHeadersStyles(width int) headersStyles {
	lolStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)
	esportStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cyan)).
		Bold(true)

	headersStyle := lipgloss.NewStyle().
		Width(width - (len("lolesport")+2)*2).
		Align(lipgloss.Center)
	normalHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Faint(true)
	selectedHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(gold)).
		Bold(true)

	fillerStyle := lipgloss.NewStyle().Width(width)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)

	return headersStyles{
		lol:            lolStyle,
		esport:         esportStyle,
		headers:        headersStyle,
		normalHeader:   normalHeaderStyle,
		selectedHeader: selectedHeaderStyle,
		separatorLine:  separatorStyle,
		filler:         fillerStyle,
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
		styles:  newDefaultHeadersStyles(0),
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
		m.styles = newDefaultHeadersStyles(m.width)
	}

	return m, nil
}

func (m *headersModel) View() string {
	lol := m.styles.lol.Render("lol")
	esport := m.styles.esport.Render("esport")
	lolesport := lipgloss.NewStyle().Padding(0, 1).Render(lol + esport)

	headers := make([]string, len(m.headers))
	for i, header := range m.headers {
		isSelected := m.cursor == i
		if isSelected {
			headers[i] = m.styles.selectedHeader.Render(header)
		} else {
			headers[i] = m.styles.normalHeader.Render(header)
		}
	}

	renderedHeaders := m.styles.headers.Render(strings.Join(headers, bulletSeparator))
	filler := m.styles.filler.Render(strings.Repeat(" ", lipgloss.Width(lolesport)))
	separator := m.styles.separatorLine.Render(strings.Repeat("━", m.width))

	return fmt.Sprintf("%s\n%s", lolesport+renderedHeaders+filler, separator)
}

func (m *headersModel) moveCursorRight() {
	m.cursor = (m.cursor + 1) % len(m.headers)
}

func (m *headersModel) moveCursorLeft() {
	// Traverse the headers like a ring.
	m.cursor = (m.cursor - 1 + len(m.headers)) % len(m.headers)
}

func changeSection(state state) tea.Cmd {
	return func() tea.Msg {
		return state
	}
}
