package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/lolesport/internal/dateutils"
	"github.com/matthieugusmini/lolesport/internal/lolesport"
)

const (
	white    = "#FFFFFF"
	gray     = "#777777"
	charcoal = "#333333"
	cyan     = "#00FFFF"
)

const (
	iconStrokeEye = "\uf070"
)

var docStyle = lipgloss.NewStyle().Padding(1, 2)

type scheduleModel struct { 
	lolesportClient LoLEsportClient

	matches       list.Model
	width, height int
}

func newScheduleModel(lolesportClient LoLEsportClient) *scheduleModel {
	return &scheduleModel{
		lolesportClient: lolesportClient,
	}
}

// The list uses the full screen size so we have to wait
// for tea.WindowSizeMsg before creating the list.
func (m *scheduleModel) Init() tea.Cmd { return nil }

func (m *scheduleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *lolesport.Schedule:
		m.matches = newMatchList(msg.Events, m.width, m.height)

	case tea.KeyMsg:
		switch msg.String() {
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.setSize(msg.Width-h, msg.Height-v-headersHeight)

		return m, m.getSchedule()
	}

	var cmd tea.Cmd
	m.matches, cmd = m.matches.Update(msg)

	if item, ok := m.matches.SelectedItem().(matchItem); ok {
		title := formatDateTitle(item.startTime)
		m.matches.Title = lipgloss.PlaceHorizontal(m.width-4, lipgloss.Center, title)
		m.matches.Styles.Title = lipgloss.NewStyle().
			Foreground(lipgloss.Color(white)).
			Background(lipgloss.Color(charcoal)).
			Bold(true)
	}

	return m, cmd
}

func (m *scheduleModel) View() string {
	return docStyle.Render(m.matches.View())
}

func (m *scheduleModel) getSchedule() tea.Cmd {
	return func() tea.Msg {
		schedule, err := m.lolesportClient.GetSchedule(context.Background(), lolesport.GetScheduleOptions{})
		if err != nil {
			return err
		}
		return schedule
	}
}

func (m *scheduleModel) setSize(width, height int) {
	m.width, m.height = width, height
}

func newMatchList(events []lolesport.Event, width, height int) list.Model {
	var (
		items       []list.Item
		cursorIndex int
	)
	for i, event := range events {
		if event.Type != lolesport.EventTypeMatch {
			continue
		}

		item := newMatchItem(event, width)
		items = append(items, item)

		if cursorIndex == 0 && !item.isCompleted {
			cursorIndex = i
		}
	}

	l := list.New(items, newMatchItemDelegate(), width, height)
	l.Select(cursorIndex)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)

	return l
}

type team struct {
	name     string
	gameWins int
}

func newTeam(t lolesport.Team) team {
	var gameWins int
	if t.Result != nil {
		gameWins = t.Result.GameWins
	}
	return team{
		name:     t.Code,
		gameWins: gameWins,
	}
}

type matchItemStyles struct {
	score     lipgloss.Style
	teamName  lipgloss.Style
	separator lipgloss.Style
	startTime lipgloss.Style
	desc      lipgloss.Style
}

func newDefaultmatchItemStyles() matchItemStyles {
	scoreStyle := lipgloss.NewStyle().Align(lipgloss.Center)

	startTimeStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Align(lipgloss.Left).
		Foreground(lipgloss.Color(white)).
		Bold(true)

	teamNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(gray))

	descStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(gray)).
		Bold(true)

	return matchItemStyles{
		score:     scoreStyle,
		startTime: startTimeStyle,
		teamName:  teamNameStyle,
		separator: separatorStyle,
		desc:      descStyle,
	}
}

type matchItem struct {
	team1          team
	team2          team
	startTime      time.Time
	leagueName     string
	blockName      string
	strategy       string
	isCompleted    bool
	spoilerRemoved bool

	width  int
	styles matchItemStyles
}

func newMatchItem(event lolesport.Event, width int) matchItem {
	return matchItem{
		team1:       newTeam(event.Match.Teams[0]),
		team2:       newTeam(event.Match.Teams[1]),
		startTime:   event.StartTime,
		leagueName:  event.League.Name,
		blockName:   event.BlockName,
		strategy:    formatMatchStrategy(event.Match.Strategy),
		isCompleted: event.State == lolesport.EventStateCompleted,
		width:       width,
		styles:      newDefaultmatchItemStyles(),
	}
}

func (i matchItem) Title() string {
	var title string
	if !i.isCompleted {
		const startTimeWidth, startTimePadding = 5, 2
		startTime := i.styles.startTime.
			Width(startTimeWidth + startTimePadding).
			Render(i.startTime.Format("15:04"))

		team1Name := i.styles.teamName.Render(i.team1.name)
		team2Name := i.styles.teamName.Render(i.team2.name)
		sep := i.styles.separator.Render(" / ")
		score := i.styles.score.
			Width(i.width - lipgloss.Width(startTime)*2).
			Render(team1Name + sep + team2Name)

		filler := strings.Repeat(" ", lipgloss.Width(startTime))

		title = lipgloss.JoinHorizontal(
			lipgloss.Center,
			startTime,
			score,
			filler,
		)
	} else if !i.spoilerRemoved {
		team1Name := i.styles.teamName.Render(i.team1.name)
		team2Name := i.styles.teamName.Render(i.team2.name)
		sep := i.styles.separator.Render(iconStrokeEye)

		title = i.styles.score.
			Width(i.width).
			Render(fmt.Sprintf("%s %s %s", team1Name, sep, team2Name))
	} else {
		team1NameAndScore := i.styles.teamName.Render(fmt.Sprintf("%s %d", i.team1.name, i.team1.gameWins))
		team2NameAndScore := i.styles.teamName.Render(fmt.Sprintf("%d %s ", i.team2.gameWins, i.team2.name))
		sep := i.styles.separator.Render(" / ")

		title = i.styles.score.
			Width(i.width).
			Render(team1NameAndScore + sep + team2NameAndScore)
	}

	return title
}

func (i matchItem) Description() string {
	const padding = 2
	sideColumnWidth := max(len(i.leagueName), len(i.strategy)) + padding

	leagueName := i.styles.desc.
		Width(sideColumnWidth).
		Align(lipgloss.Left).
		Render(i.leagueName)

	blockName := i.styles.desc.
		Width(i.width - sideColumnWidth*2).
		Align(lipgloss.Center).
		Render(i.blockName)

	strategy := i.styles.desc.
		Width(sideColumnWidth).
		Align(lipgloss.Right).
		Render(i.strategy)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		leagueName,
		blockName,
		strategy,
	)
}

func (i matchItem) FilterValue() string {
	return strings.Join([]string{
		i.team1.name,
		i.team2.name,
		i.leagueName,
	}, "_")
}

type matchItemDelegateStyles struct {
	normalItem   lipgloss.Style
	selectedItem lipgloss.Style
}

func newDefaultMatchItemDelegateStyles() matchItemDelegateStyles {
	itemStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	normalItemStyle := itemStyle.
		Foreground(lipgloss.Color(gray)).
		BorderForeground(lipgloss.Color(gray))

	selectedItemStyle := itemStyle.
		Foreground(lipgloss.Color(cyan)).
		BorderForeground(lipgloss.Color(cyan))

	return matchItemDelegateStyles{
		normalItem:   normalItemStyle,
		selectedItem: selectedItemStyle,
	}
}

type matchItemDelegate struct {
	styles matchItemDelegateStyles
}

func newMatchItemDelegate() matchItemDelegate {
	return matchItemDelegate{
		styles: newDefaultMatchItemDelegateStyles(),
	}
}

func (d matchItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	matchItem, ok := item.(matchItem)
	if !ok {
		return
	}

	var (
		matchItemStyle = d.styles.normalItem
		isSelected     = index == m.Index()
	)
	if isSelected {
		matchItemStyle = d.styles.selectedItem
	}

	content := fmt.Sprintf(
		"%s\n%s\n%s",
		matchItem.Title(),
		strings.Repeat("─", matchItem.width),
		matchItem.Description(),
	)
	fmt.Fprintf(w, "%s", matchItemStyle.Render(content))
}

func (d matchItemDelegate) Height() int { return 5 }

func (d matchItemDelegate) Spacing() int { return 0 }

func (d matchItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			item := m.SelectedItem().(matchItem)
			item.spoilerRemoved = true
			m.SetItem(m.Index(), item)
		}
	}
	return nil
}

func formatMatchStrategy(strategy lolesport.Strategy) string {
	switch strategy.Type {
	case lolesport.MatchStrategyTypeBestOf:
		return fmt.Sprintf("Bo%d", strategy.Count)
	default:
		return ""
	}
}

func formatDateTitle(date time.Time) string {
	switch {
	case dateutils.IsYesterday(date):
		return "Yesterday"
	case dateutils.IsToday(date):
		if date.After(time.Now()) {
			return "Later Today"
		} else {
			return "Today"
		}
	case dateutils.IsTomorrow(date):
		return "Tomorrow"
	default:
		return date.Format("Monday 02 Jan")
	}
}
