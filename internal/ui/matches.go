package ui

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/lolesport/internal/timeutils"
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
	var items []list.Item
	for _, event := range events {
		if event.Type != lolesports.EventTypeMatch {
			continue
		}

		item := newMatchItem(event)
		items = append(items, item)
	}
	return items
}

func newMatchList(events []lolesports.Event, width, height int) list.Model {
	items := newMatchListItems(events)

	l := list.New(items, newMatchItemDelegate(), width, height)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)

	i := slices.IndexFunc(items, func(item list.Item) bool {
		return timeutils.IsToday(item.(matchItem).startTime)
	})
	l.Select(i)

	return l
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
	desc               lipgloss.Style
	flags              lipgloss.Style
	leagueAndBlockName lipgloss.Style
	strategy           lipgloss.Style
}

func newDefaultMatchItemStyles() (s matchItemStyles) {
	// Item
	itemStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	s.normalItem = itemStyle.
		Foreground(gray).
		BorderForeground(gray)

	s.selectedItem = itemStyle.
		Foreground(gold).
		BorderForeground(gold)

	// Title
	s.title = lipgloss.NewStyle().Padding(0, 1)

	s.startTime = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(white).
		Bold(true)

	s.teamName = lipgloss.NewStyle().
		Foreground(white).
		Bold(true)

	s.separator = lipgloss.NewStyle().
		Foreground(gray)

	s.upcomingMatchScore = lipgloss.NewStyle().
		Align(lipgloss.Center)

	s.completedMatchScore = lipgloss.NewStyle().
		Align(lipgloss.Center)

	// Description
	s.desc = lipgloss.NewStyle().Padding(0, 1)

	s.flags = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(gray)

	s.leagueAndBlockName = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(gray).
		Bold(true)

	s.strategy = lipgloss.NewStyle().
		Align(lipgloss.Right).
		Foreground(gray).
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

	if m.Width() <= 0 {
		return
	}

	var title string
	// Some matches are completed but unstarted somehow so we render those
	// with their score.
	if !matchItem.isCompleted && matchItem.startTime.After(time.Now()) {
		title = d.viewTitleWithStartTime(matchItem, m.Width())
	} else if !matchItem.spoilerBlockRevealed {
		title = d.viewTitleWithScoreSpoilerBlock(matchItem, m.Width())
	} else {
		title = d.viewTitleWithScore(matchItem, m.Width())
	}

	desc := d.viewDescription(matchItem, m.Width())

	content := fmt.Sprintf("%s\n%s\n%s", title, strings.Repeat("─", m.Width()), desc)

	var (
		matchItemStyle = d.styles.normalItem.Width(m.Width())
		isSelected     = index == m.Index()
	)
	if isSelected {
		matchItemStyle = d.styles.selectedItem.Width(m.Width())
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

func (d matchItemDelegate) viewTitleWithStartTime(item matchItem, width int) string {
	padding := d.styles.title.GetHorizontalFrameSize()

	startTime := d.styles.startTime.Render(item.startTime.Format(matchStartTimeLayout))

	team1Name := d.styles.teamName.Render(item.team1.name)
	team2Name := d.styles.teamName.Render(item.team2.name)
	sep := d.styles.separator.Render(separatorSlash)
	score := d.styles.upcomingMatchScore.
		Width(width - padding - lipgloss.Width(startTime)*2).
		Render(team1Name + sep + team2Name)

	filler := strings.Repeat(" ", lipgloss.Width(startTime))

	return d.styles.title.Render(startTime + score + filler)
}

func (d matchItemDelegate) viewTitleWithScoreSpoilerBlock(item matchItem, width int) string {
	team1Name := d.styles.teamName.Render(item.team1.name)
	team2Name := d.styles.teamName.Render(item.team2.name)
	sep := d.styles.separator.Render(separataorStrokeEye)

	title := team1Name + sep + team2Name

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
}

func (d matchItemDelegate) viewTitleWithScore(item matchItem, width int) string {
	team1NameAndScore := d.styles.teamName.Render(fmt.Sprintf("%s %d", item.team1.name, item.team1.gameWins))
	team2NameAndScore := d.styles.teamName.Render(fmt.Sprintf("%d %s ", item.team2.gameWins, item.team2.name))
	sep := d.styles.separator.Render(separatorSlash)

	title := team1NameAndScore + sep + team2NameAndScore

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
}

func (d matchItemDelegate) viewDescription(item matchItem, width int) string {
	padding := d.styles.desc.GetHorizontalFrameSize()
	sideColumnWidth := max(lipgloss.Width(item.flags), lipgloss.Width(item.strategy))

	flags := d.styles.flags.
		Width(sideColumnWidth).
		Render(item.flags)

	strategy := d.styles.strategy.
		Width(sideColumnWidth).
		Render(item.strategy)

	leagueAndBlockName := d.styles.leagueAndBlockName.
		Width(width - padding - sideColumnWidth*2).
		Render(item.leagueName + separatorBullet + item.blockName)

	return d.styles.desc.Render(flags + leagueAndBlockName + strategy)
}

func formatMatchStrategy(strategy lolesports.Strategy) string {
	switch strategy.Type {
	case lolesports.MatchStrategyTypeBestOf:
		return fmt.Sprintf("Bo%d", strategy.Count)
	default:
		return "Unknown"
	}
}
