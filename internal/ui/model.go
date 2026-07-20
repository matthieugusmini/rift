package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
)

const (
	logo = "Rift"

	navItemLabelSchedule  = "Schedule"
	navItemLabelStandings = "Standings"

	navbarHeight = 2

	maxWidth = 120
)

type navItem struct {
	label string
	state state
}

var navItems = []navItem{
	{label: navItemLabelSchedule, state: stateShowSchedule},
	{label: navItemLabelStandings, state: stateShowStandings},
}

type state int

const (
	stateShowSchedule state = iota
	stateShowStandings
)

type modelStyles struct {
	logo            lipgloss.Style
	normalNavItem   lipgloss.Style
	selectedNavItem lipgloss.Style
	separator       lipgloss.Style
}

func newDefaultModelStyles(theme theme) (s modelStyles) {
	s.logo = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(theme.textTitle).
		Bold(true)

	s.normalNavItem = lipgloss.NewStyle().
		Foreground(theme.textPrimary).
		Faint(true)

	s.selectedNavItem = lipgloss.NewStyle().
		Foreground(theme.selected).
		Bold(true)

	s.separator = lipgloss.NewStyle().
		Foreground(theme.textSecondary).
		Bold(true)

	return s
}

// LoLEsportsLoader loads LoL Esports data.
type LoLEsportsLoader interface {
	// GetSchedule fetches and returns the LoL Esports schedule data.
	//
	// Optionnaly you can use the [github.com/matthieugusmini/rift/internal/rift.GetScheduleOptions] to:
	// - Fetch events only for specific leagues
	// - Specify which page to fetch
	GetSchedule(
		ctx context.Context,
		opts *lolesports.GetScheduleOptions,
	) (lolesports.Schedule, error)

	// LoadStandingsByTournamentIDs loads the standings associated with
	// each given tournament ids.
	LoadStandingsByTournamentIDs(
		ctx context.Context,
		tournamentIDs []string,
	) ([]lolesports.Standings, error)

	// LoadCurrentSeasonSplits loads and returns all the LoL Esports splits
	// for the current season.
	LoadCurrentSeasonSplits(ctx context.Context) ([]lolesports.Split, error)
}

// LoLEsportsStageClient loads detailed stage topology from LoL Esports.
type LoLEsportsStageClient interface {
	GetStage(ctx context.Context, stageID string) (lolesportsgraphql.Stage, error)
}

// page is similar to a tea.Model but with the added ability to set its size.
// It's particularly useful for managing sub-models that need to be displayed
// in specific screen areas (e.g., between a navbar and footer).
//
// Note: Cannot embed tea.Model as Update should return a page.
type page interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (page, tea.Cmd)
	View() string

	setSize(width, height int)
	setTheme(theme theme)
}

// Model implements the [charm.land/bubbletea/v2.Model] interface.
//
// It is the main model of the application which dictate which sub-model
// should be displayed and how to navigate between pages.
type Model struct {
	// Actual terminal size
	width, height int

	// Represents the width the application actually uses for display.
	// Some of the pages renders horrendously when they take too much
	// space we keep this values below maxWidth.
	pageWidth int

	// Indicates which sub-model to display.
	state state

	// Index of the selected item in the navbar
	selectedNavIndex int

	currentPage page
	pages       map[state]page

	styles modelStyles
	theme  theme
}

// NewModel returns a new [Model] initialized with all its sub-models
// and default styles.
func NewModel(
	lolesportsLoader LoLEsportsLoader,
	lolesportsStageClient LoLEsportsStageClient,
	logger *slog.Logger,
) Model {
	theme := newTheme(true)
	schedulePage := newSchedulePage(lolesportsLoader, logger, theme)
	standingsPage := newStandingsPage(
		lolesportsLoader,
		lolesportsStageClient,
		logger,
		theme,
	)

	pages := map[state]page{
		stateShowSchedule:  schedulePage,
		stateShowStandings: standingsPage,
	}

	return Model{
		currentPage: schedulePage,
		pages:       pages,
		styles:      newDefaultModelStyles(theme),
		theme:       theme,
	}
}

// Init implements the [charm.land/bubbletea/v2.Model] interface.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.currentPage.Init(), tea.RequestBackgroundColor)
}

// Update implements the [charm.land/bubbletea/v2.Model] interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			return m.navigateRight()
		case "shift+tab":
			return m.navigateLeft()
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.pageWidth = min(msg.Width, maxWidth)
		for _, page := range m.pages {
			page.setSize(m.pageWidth, msg.Height-navbarHeight)
		}

	case tea.BackgroundColorMsg:
		m.setTheme(newTheme(msg.IsDark()))
	}

	var cmd tea.Cmd
	m.currentPage, cmd = m.currentPage.Update(msg)

	return m, cmd
}

// View implements the [charm.land/bubbletea/v2.Model] interface.
func (m Model) View() tea.View {
	navBar := m.viewNavbar(navItems, m.selectedNavIndex, m.pageWidth)

	pageContent := m.currentPage.View()

	view := lipgloss.JoinVertical(lipgloss.Left, navBar, pageContent)

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		Render(view)

	result := tea.NewView(content)
	result.AltScreen = true
	result.MouseMode = tea.MouseModeCellMotion

	return result
}

func (m *Model) setTheme(theme theme) {
	if m.theme.isDark == theme.isDark {
		return
	}

	m.theme = theme
	m.styles = newDefaultModelStyles(theme)

	for _, page := range m.pages {
		page.setTheme(theme)
	}
}

func (m Model) viewNavbar(
	navItems []navItem,
	selectedNavIndex int,
	width int,
) string {
	logo := m.styles.logo.Render(logo)

	// Add a padding with the same size as the logo on the other end
	// of the navbar to center the nav items.
	padding := strings.Repeat(" ", lipgloss.Width(logo))

	styledNavItems := make([]string, len(navItems))
	for i, navItem := range navItems {
		isSelected := selectedNavIndex == i
		if isSelected {
			styledNavItems[i] = m.styles.selectedNavItem.Render(navItem.label)
		} else {
			styledNavItems[i] = m.styles.normalNavItem.Render(navItem.label)
		}
	}

	availWidth := width - lipgloss.Width(logo) - lipgloss.Width(padding)
	navItemsStyle := lipgloss.NewStyle().
		Width(availWidth).
		Align(lipgloss.Center)
	renderedNavItems := navItemsStyle.Render(strings.Join(styledNavItems, separatorBullet))

	navbar := logo + renderedNavItems + padding

	separator := m.styles.separator.Render(strings.Repeat(separatorLine, width))

	return fmt.Sprintf("%s\n%s", navbar, separator)
}

func (m Model) navigateRight() (Model, tea.Cmd) {
	m.selectedNavIndex = moveNavigationBarCursorRight(m.selectedNavIndex)
	return m.updateCurrentPage()
}

func (m Model) navigateLeft() (Model, tea.Cmd) {
	m.selectedNavIndex = moveNavigationBarCursorLeft(m.selectedNavIndex)
	return m.updateCurrentPage()
}

func (m Model) updateCurrentPage() (Model, tea.Cmd) {
	m.state = navItems[m.selectedNavIndex].state
	m.currentPage = m.pages[m.state]
	return m, m.currentPage.Init()
}

func moveNavigationBarCursorLeft(current int) int {
	return moveCursor(current, -1, len(navItems))
}

func moveNavigationBarCursorRight(current int) int {
	return moveCursor(current, 1, len(navItems))
}

func moveCursor(current, delta, upperBound int) int {
	if upperBound == 0 {
		return 0
	}
	return (current + delta + upperBound) % upperBound
}
