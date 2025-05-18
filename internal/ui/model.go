package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/rift/internal/rift"
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

func newDefaultModelStyles() (s modelStyles) {
	s.logo = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(textTitleColor).
		Bold(true)

	s.normalNavItem = lipgloss.NewStyle().
		Foreground(textPrimaryColor).
		Faint(true)

	s.selectedNavItem = lipgloss.NewStyle().
		Foreground(selectedColor).
		Bold(true)

	s.separator = lipgloss.NewStyle().
		Foreground(textSecondaryColor).
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

	// GetCurrentSeasonSplits fetches and returns all the LoL Esports splits
	// for the current season.
	GetCurrentSeasonSplits(ctx context.Context) ([]lolesports.Split, error)
}

// BracketTemplateLoader loads bracket templates.
type BracketTemplateLoader interface {
	// ListAvailableStageIDs returns the list of ids of all the stages
	// that have a bracket template associated with them.
	ListAvailableStageIDs(ctx context.Context) ([]string, error)

	// Load returns the [rift.BracketTemplate] associated with stageID.
	Load(ctx context.Context, stageID string) (rift.BracketTemplate, error)
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
}

// Model implements the [github.com/charmbracelet/bubbletea.Model] interface.
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
}

// NewModel returns a new [Model] initialized with all its sub-models
// and default styles.
func NewModel(
	lolesportsLoader LoLEsportsLoader,
	bracketLoader BracketTemplateLoader,
	logger *slog.Logger,
) Model {
	schedulePage := newSchedulePage(lolesportsLoader, logger)
	standingsPage := newStandingsPage(lolesportsLoader, bracketLoader, logger)

	pages := map[state]page{
		stateShowSchedule:  schedulePage,
		stateShowStandings: standingsPage,
	}

	return Model{
		currentPage: schedulePage,
		pages:       pages,
		styles:      newDefaultModelStyles(),
	}
}

// Init implements the [github.com/charmbracelet/bubbletea.Model] interface.
func (m Model) Init() tea.Cmd {
	return m.currentPage.Init()
}

// Update implements the [github.com/charmbracelet/bubbletea.Model] interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
	}

	var cmd tea.Cmd
	m.currentPage, cmd = m.currentPage.Update(msg)

	return m, cmd
}

// View implements the [github.com/charmbracelet/bubbletea.Model] interface.
func (m Model) View() string {
	navBar := m.viewNavbar(navItems, m.selectedNavIndex, m.pageWidth)

	content := m.currentPage.View()

	view := lipgloss.JoinVertical(lipgloss.Left, navBar, content)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		Render(view)
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
