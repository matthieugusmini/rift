package ui

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/rift/internal/timeutil"
)

const (
	matchStartTimeLayout = "15:04"

	matchItemHeight = 5
)

type team struct {
	name     string
	gameWins int
}

func newTeam(t lolesports.Team) team {
	var gameWins int
	if t.Result != nil {
		gameWins = t.Result.GameWins
	}
	return team{
		name:     t.Code,
		gameWins: gameWins,
	}
}

type matchItem struct {
	team1      team
	team2      team
	startTime  time.Time
	leagueName string
	blockName  string
	strategy   string
	flags      string

	isCompleted          bool
	spoilerBlockRevealed bool
}

func newMatchItem(event lolesports.Event) matchItem {
	return matchItem{
		team1:       newTeam(event.Match.Teams[0]),
		team2:       newTeam(event.Match.Teams[1]),
		startTime:   event.StartTime.Local(),
		leagueName:  event.League.Name,
		blockName:   event.BlockName,
		strategy:    formatMatchStrategy(event.Match.Strategy),
		isCompleted: event.State == lolesports.EventStateCompleted,
		flags:       strings.Join(flagsByLeagueName[event.League.Name], separatorBullet),
	}
}

func (i matchItem) FilterValue() string {
	return strings.Join([]string{
		i.team1.name,
		i.team2.name,
		i.leagueName,
	}, "_")
}

func newMatchListItems(events []lolesports.Event) []list.Item {
	items := make([]list.Item, len(events))

	for i, event := range events {
		items[i] = newMatchItem(event)
	}

	return items
}

func newMatchList(events []lolesports.Event, width, height int) list.Model {
	items := newMatchListItems(events)

	l := list.New(items, newMatchItemDelegate(), width, height)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.StatusMessageLifetime = time.Second * 2
	l.SetSpinner(spinner.MiniDot)
	l.SetShowHelp(false)

	cursorStartingPos := indexMatchListInitialCursor(events)
	l.Select(cursorStartingPos)

	return l
}

func indexMatchListInitialCursor(events []lolesports.Event) int {
	// The starting position of the cursor is the first match of the first day
	// with a match starting from today.
	cursorStartingPos := slices.IndexFunc(events, func(event lolesports.Event) bool {
		return !timeutil.IsBeforeToday(event.StartTime)
	})
	// Fallback to 0 but it should never happen. Right?
	if cursorStartingPos == -1 {
		return 0
	}
	return cursorStartingPos
}

type matchItemStyles struct {
	// Item
	normalItem   lipgloss.Style
	selectedItem lipgloss.Style

	// Title
	title               lipgloss.Style
	completedMatchScore lipgloss.Style
	upcomingMatchScore  lipgloss.Style
	teamName            lipgloss.Style
	separator           lipgloss.Style
	startTime           lipgloss.Style

	// Description
	flags              lipgloss.Style
	leagueAndBlockName lipgloss.Style
	strategy           lipgloss.Style
}

func newDefaultMatchItemStyles() (s matchItemStyles) {
	// Item
	itemStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	s.normalItem = itemStyle.
		Foreground(textSecondaryColor).
		BorderForeground(borderPrimaryColor)

	s.selectedItem = itemStyle.
		Foreground(selectedColor).
		BorderForeground(selectedColor)

	// Title
	s.title = lipgloss.NewStyle().Padding(0, 1)

	s.startTime = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(textPrimaryColor).
		Bold(true)

	s.teamName = lipgloss.NewStyle().
		Foreground(textPrimaryColor).
		Bold(true)

	s.separator = lipgloss.NewStyle().
		Foreground(textSecondaryColor)

	s.upcomingMatchScore = lipgloss.NewStyle().
		Align(lipgloss.Center)

	s.completedMatchScore = lipgloss.NewStyle().
		Align(lipgloss.Center)

	// Description
	s.flags = lipgloss.NewStyle().
		Padding(0, 1).
		Align(lipgloss.Left).
		Foreground(textSecondaryColor).
		Bold(true)

	s.leagueAndBlockName = lipgloss.NewStyle().
		Padding(0, 1).
		Align(lipgloss.Center).
		Foreground(textSecondaryColor).
		Bold(true)

	s.strategy = lipgloss.NewStyle().
		Padding(0, 1).
		Align(lipgloss.Right).
		Foreground(textSecondaryColor).
		Bold(true)

	return s
}

type matchItemDelegate struct {
	styles matchItemStyles
}

func newMatchItemDelegate() matchItemDelegate {
	return matchItemDelegate{
		styles: newDefaultMatchItemStyles(),
	}
}

func (d matchItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	matchItem, ok := item.(matchItem)
	if !ok {
		return
	}

	itemWidth := m.Width() - d.styles.normalItem.GetHorizontalFrameSize()
	if itemWidth <= 0 {
		return
	}

	var title string
	// Some matches are completed but unstarted somehow so we render those
	// with their score.
	if !matchItem.isCompleted && matchItem.startTime.After(time.Now()) {
		title = d.viewTitleWithStartTime(matchItem, itemWidth)
	} else if !matchItem.spoilerBlockRevealed {
		title = d.viewTitleWithScoreSpoilerBlock(matchItem, itemWidth)
	} else {
		title = d.viewTitleWithScore(matchItem, itemWidth)
	}

	desc := d.viewDescription(matchItem, itemWidth)

	content := fmt.Sprintf("%s\n%s\n%s", title, strings.Repeat("─", itemWidth), desc)

	var (
		matchItemStyle = d.styles.normalItem.Width(itemWidth)
		isSelected     = index == m.Index()
	)
	if isSelected {
		matchItemStyle = d.styles.selectedItem.Width(itemWidth)
	}

	fmt.Fprintf(w, "%s", matchItemStyle.Render(content))
}

func (d matchItemDelegate) Height() int { return matchItemHeight }

func (d matchItemDelegate) Spacing() int { return 0 }

func (d matchItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			d.revealSpoiler(m)
		}
	}
	return nil
}

func (d matchItemDelegate) revealSpoiler(m *list.Model) {
	item := m.SelectedItem().(matchItem)
	item.spoilerBlockRevealed = true
	m.SetItem(m.Index(), item)
}

//	┌──────┬────────────────────────────────────────┬──────┐
//	│ time │             TEAM1 / TEAM2              │      │
//	└──────┴────────────────────────────────────────┴──────┘
//
// Used for upcoming matches.
func (d matchItemDelegate) viewTitleWithStartTime(item matchItem, width int) string {
	padding := d.styles.title.GetHorizontalFrameSize()

	startTime := d.styles.startTime.Render(item.startTime.Format(matchStartTimeLayout))

	team1Name := d.styles.teamName.Render(item.team1.name)
	team2Name := d.styles.teamName.Render(item.team2.name)
	sep := d.styles.separator.Render(separatorSlash)
	score := d.styles.upcomingMatchScore.
		Width(width - padding - lipgloss.Width(startTime)*2).
		Render(team1Name + sep + team2Name)

	// Place a filler with the same size as startTime to maintain alignment.
	filler := strings.Repeat(" ", lipgloss.Width(startTime))

	return d.styles.title.Render(startTime + score + filler)
}

//	┌──────────────────────────────────────────────────────┐
//	│                  TEAM1 [ EYE ] TEAM2                 │
//	└──────────────────────────────────────────────────────┘
//
// Used for completed matches with spoiler block still hidden.
func (d matchItemDelegate) viewTitleWithScoreSpoilerBlock(item matchItem, width int) string {
	team1Name := d.styles.teamName.Render(item.team1.name)
	team2Name := d.styles.teamName.Render(item.team2.name)
	sep := d.styles.separator.Render(separataorStrokeEye)

	title := team1Name + sep + team2Name

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
}

//	┌──────────────────────────────────────────────────────┐
//	│              TEAM1 SCORE / SCORE TEAM2               │
//	└──────────────────────────────────────────────────────┘
//
// Used for completed matches with scores revealed.
func (d matchItemDelegate) viewTitleWithScore(item matchItem, width int) string {
	team1NameAndScore := d.styles.teamName.Render(
		fmt.Sprintf("%s %d", item.team1.name, item.team1.gameWins),
	)
	team2NameAndScore := d.styles.teamName.Render(
		fmt.Sprintf("%d %s ", item.team2.gameWins, item.team2.name),
	)
	sep := d.styles.separator.Render(separatorSlash)

	title := team1NameAndScore + sep + team2NameAndScore

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
}

//	┌────────────┬────────────────────────────┬────────────┐
//	│   FLAGS    │    LEAGUE • BLOCK NAME     │  STRATEGY  │
//	└────────────┴────────────────────────────┴────────────┘
//
// LEAGUE & BLOCK NAME gets content-based width + its padding.
// The remaining space is split evenly between FLAGS and STRATEGY.
// FLAGS is truncated with ellipsis when necessary.
func (d matchItemDelegate) viewDescription(item matchItem, width int) string {
	leagueAndBlockName := d.styles.leagueAndBlockName.Render(
		item.leagueName + separatorBullet + item.blockName,
	)

	availWidth := width - lipgloss.Width(leagueAndBlockName)

	sideColumnWidth := availWidth / 2
	strategy := d.styles.strategy.
		Width(sideColumnWidth).
		Render(item.strategy)

	// We don't use sideColumnWidth as it would be incorrect when
	// availWidth is an odd number.
	flagsMaxWidth := availWidth - lipgloss.Width(strategy)
	flags := d.styles.flags.
		Width(flagsMaxWidth).
		Render(item.flags)
	flags = ansi.Truncate(flags, flagsMaxWidth, "…")

	return flags + leagueAndBlockName + strategy
}

func formatMatchStrategy(strategy lolesports.Strategy) string {
	switch strategy.Type {
	case lolesports.MatchStrategyTypeBestOf:
		return fmt.Sprintf("Bo%d", strategy.Count)
	default:
		return "Unknown"
	}
}
