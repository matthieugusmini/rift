package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
)

const (
	matchWidth = 20
)

const (
	bracketPageShortHelpHeight = 1
	bracketPageFullHelpHeight  = 5
)

type bracketPageKeyMap struct {
	baseKeyMap

	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Previous key.Binding
}

func newDefaultBracketPageKeyMap() bracketPageKeyMap {
	return bracketPageKeyMap{
		baseKeyMap: newBaseKeyMap(),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Previous: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "previous"),
		),
	}
}

type bracketPageStyles struct {
	roundTitle     lipgloss.Style
	matchBorder    lipgloss.Style
	noTeamResult   lipgloss.Style
	loserTeamName  lipgloss.Style
	winnerTeamName lipgloss.Style
	link           lipgloss.Style
	help           lipgloss.Style
}

func newDefaultBracketPageStyles() (s bracketPageStyles) {
	s.roundTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(black)).
		Background(lipgloss.Color(antiFlashWhite)).
		Padding(0, 1).
		Bold(true)

	s.matchBorder = lipgloss.NewStyle().Foreground(borderPrimaryColor)

	s.noTeamResult = lipgloss.NewStyle().
		Foreground(textPrimaryColor)

	s.loserTeamName = lipgloss.NewStyle().
		Foreground(textSecondaryColor).
		Faint(true)

	s.winnerTeamName = lipgloss.NewStyle().
		Foreground(selectedColor).
		Bold(true)

	s.link = lipgloss.NewStyle().Foreground(borderSecondaryColor)

	s.help = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	return s
}

type bracketPage struct {
	width, height int
	dynamicStage  *lolesportsgraphql.Stage
	viewport      viewport.Model
	help          help.Model
	keyMap        bracketPageKeyMap
	styles        bracketPageStyles
}

func newDynamicBracketPage(
	stage lolesportsgraphql.Stage,
	width, height int,
) *bracketPage {
	m := &bracketPage{
		dynamicStage: &stage,
		width:        width,
		height:       height,
		help:         help.New(),
		keyMap:       newDefaultBracketPageKeyMap(),
		styles:       newDefaultBracketPageStyles(),
	}

	m.initViewport()

	return m
}

func (m *bracketPage) Update(msg tea.Msg) (*bracketPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ShowFullHelp),
			key.Matches(msg, m.keyMap.CloseFullHelp):
			m.toggleFullHelp()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *bracketPage) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
		m.viewHelp(),
	)
}

func (m *bracketPage) viewHelp() string {
	return m.styles.help.Render(m.help.View(m))
}

func (m *bracketPage) setSize(width, height int) {
	m.width, m.height = width, height

	// Setting the Height and Width field doesn't seem to work
	// so we recreate it with the right size.
	m.initViewport()
}

func (p *bracketPage) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Right,
		p.keyMap.Left,
		p.keyMap.Previous,
		p.keyMap.Quit,
		p.keyMap.ShowFullHelp,
	}
}

func (p *bracketPage) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Motions
		{
			p.keyMap.Up,
			p.keyMap.Down,
			p.keyMap.Right,
			p.keyMap.Left,
			p.keyMap.Previous,
		},
		// Navigation
		{
			p.keyMap.NextPage,
			p.keyMap.PrevPage,
		},
		// Others
		{
			p.keyMap.Quit,
			p.keyMap.CloseFullHelp,
		},
	}
}

func (m *bracketPage) toggleFullHelp() {
	m.help.ShowAll = !m.help.ShowAll
	// Resize the viewport as the full help takes up more space.
	m.initViewport()
}

func (m *bracketPage) initViewport() {
	content := renderDynamicStage(*m.dynamicStage, m.styles)
	m.viewport = viewport.New(m.width, m.contentHeight())
	m.viewport.SetContent(content)
	m.viewport.SetHorizontalStep(5)
}

func (m *bracketPage) contentHeight() int {
	return m.height - m.helpHeight()
}

func (m *bracketPage) helpHeight() int {
	padding := m.styles.help.GetVerticalPadding()
	if m.help.ShowAll {
		return bracketPageFullHelpHeight + padding
	}
	return bracketPageShortHelpHeight + padding
}
