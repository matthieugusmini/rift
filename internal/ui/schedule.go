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
	gold     = "#FFD700"
)

const (
	iconStrokeEye = "\uf070"
)

type scheduleStyles struct {
	doc   lipgloss.Style
	title lipgloss.Style
}

func newDefaultScheduleStyles() scheduleStyles {
	docStyle := lipgloss.NewStyle().Padding(1, 2)

	return scheduleStyles{
		doc: docStyle,
	}
}

type scheduleModel struct {
	lolesportClient LoLEsportClient

	matches list.Model

	width, height int
	styles        scheduleStyles
}

func newScheduleModel(lolesportClient LoLEsportClient) *scheduleModel {
	return &scheduleModel{
		lolesportClient: lolesportClient,
		styles:          newDefaultScheduleStyles(),
	}
}

func (m *scheduleModel) Init() tea.Cmd { return nil }

func (m *scheduleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *lolesport.Schedule:
		m.matches = newMatchList(msg.Events, m.width, m.height)

	case tea.WindowSizeMsg:
		h, v := m.styles.doc.GetFrameSize()
		m.setSize(msg.Width-h, msg.Height-v-headersHeight)
		// The list uses the full screen size so we have to wait
		// for tea.WindowSizeMsg before creating the list.
		// TODO: Don't reload the page when resizing the window.
		return m, m.getSchedule()

	case state:
		if msg == stateShowSchedule {
			return m, m.getSchedule()
		}
		return m, nil
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
	return m.styles.doc.Render(m.matches.View())
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

		// Put the list cursor on the current or the closest match.
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
	// Title
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

func newDefaultmatchItemStyles(event lolesport.Event, width int) matchItemStyles {
	// Title
	const startTimeWidth, padding = 5, 2
	startTimeStyle := lipgloss.NewStyle().
		Width(startTimeWidth+padding).
		Padding(0, 1).
		Align(lipgloss.Left).
		Foreground(lipgloss.Color(white)).
		Bold(true)

	teamNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(gray))

	upcomingMatchScoreStyle := lipgloss.NewStyle().
		Width(width - startTimeStyle.GetWidth()*2).
		Align(lipgloss.Center)

	completedMatchScoreStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	// Description
	descStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(gray)).
		Bold(true)

	// TODO(high effort): Develop a grid layout
	sideColumnWidth := max(
		lipgloss.Width(strings.Join(flagsByLeagueName[event.League.Name], bulletSeparator)),
		len(formatMatchStrategy(event.Match.Strategy)),
	) + padding

	flagsStyle := descStyle.
		Width(sideColumnWidth).
		Align(lipgloss.Left)

	leagueAndBlockNameStyle := descStyle.
		Width(width - sideColumnWidth*2).
		Align(lipgloss.Center)

	strategyStyle := descStyle.
		Width(sideColumnWidth).
		Align(lipgloss.Right)

	return matchItemStyles{
		completedMatchScore: completedMatchScoreStyle,
		upcomingMatchScore:  upcomingMatchScoreStyle,
		startTime:           startTimeStyle,
		teamName:            teamNameStyle,
		separator:           separatorStyle,

		flags:              flagsStyle,
		leagueAndBlockName: leagueAndBlockNameStyle,
		strategy:           strategyStyle,
	}
}

type matchItem struct {
	team1      team
	team2      team
	startTime  time.Time
	leagueName string
	blockName  string
	strategy   string

	isCompleted    bool
	spoilerRemoved bool

	width  int
	styles matchItemStyles
}

func newMatchItem(event lolesport.Event, width int) matchItem {
	return matchItem{
		team1:       newTeam(event.Match.Teams[0]),
		team2:       newTeam(event.Match.Teams[1]),
		startTime:   event.StartTime.Local(),
		leagueName:  event.League.Name,
		blockName:   event.BlockName,
		strategy:    formatMatchStrategy(event.Match.Strategy),
		isCompleted: event.State == lolesport.EventStateCompleted,
		width:       width,
		styles:      newDefaultmatchItemStyles(event, width),
	}
}

func (i matchItem) Title() string {
	if !i.isCompleted {
		return i.titleWithStartTime()
	} else if !i.spoilerRemoved {
		return i.titleWithScoreSpoilerBlock()
	} else {
		return i.titleWithScore()
	}
}

const matchStartTimeLayout = "15:04"

func (i matchItem) titleWithStartTime() string {
	startTime := i.styles.startTime.Render(i.startTime.Format(matchStartTimeLayout))

	team1Name := i.styles.teamName.Render(i.team1.name)
	team2Name := i.styles.teamName.Render(i.team2.name)
	sep := i.styles.separator.Render(" / ")
	score := i.styles.upcomingMatchScore.Render(team1Name + sep + team2Name)

	filler := strings.Repeat(" ", lipgloss.Width(startTime))

	return startTime + score + filler
}

func (i matchItem) titleWithScoreSpoilerBlock() string {
	team1Name := i.styles.teamName.Render(i.team1.name)
	team2Name := i.styles.teamName.Render(i.team2.name)
	sep := i.styles.separator.Render(iconStrokeEye)

	return i.styles.completedMatchScore.Render(fmt.Sprintf("%s %s %s", team1Name, sep, team2Name))
}

func (i matchItem) titleWithScore() string {
	team1NameAndScore := i.styles.teamName.Render(fmt.Sprintf("%s %d", i.team1.name, i.team1.gameWins))
	team2NameAndScore := i.styles.teamName.Render(fmt.Sprintf("%d %s ", i.team2.gameWins, i.team2.name))
	sep := i.styles.separator.Render(" / ")

	return i.styles.completedMatchScore.Render(team1NameAndScore + sep + team2NameAndScore)
}

func (i matchItem) Description() string {
	flags := i.styles.flags.Render(strings.Join(flagsByLeagueName[i.leagueName], bulletSeparator))
	leagueAndBlockName := i.styles.leagueAndBlockName.Render(fmt.Sprintf("%s • %s", i.leagueName, i.blockName))
	strategy := i.styles.strategy.Render(i.strategy)
	return flags + leagueAndBlockName + strategy
}

func (i matchItem) FilterValue() string {
	return strings.Join([]string{
		i.team1.name,
		i.team2.name,
		i.leagueName,
	}, "_")
}

const matchItemHeight = 5

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
		Foreground(lipgloss.Color(gold)).
		BorderForeground(lipgloss.Color(gold))

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

func (d matchItemDelegate) Height() int { return matchItemHeight }

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

const (
	bullet          = "•"
	bulletSeparator = " " + bullet + " "
)

var flagsByLeagueName = map[string][]string{
	"LJL":                     {"🇯🇵"},
	"LEC":                     {"🇪🇺"},
	"LTA North":               {"🇺🇸"},
	"NACL":                    {"🇺🇸"},
	"LTA South":               {"🇧🇷"},
	"Circuito Desafiante":     {"🇧🇷"},
	"LCK":                     {"🇰🇷"},
	"LCK Challengers":         {"🇰🇷"},
	"LPL":                     {"🇨🇳"},
	"LCP":                     {"🇹🇼"},
	"LoL Italian Tournament":  {"🇮🇹"},
	"Hellenic Legends League": {"🇬🇷"},
	"TCL":                     {"🇹🇷"},
	"PCS":                     {"🇭🇰", "🇹🇼"},
	"Hitpoint Masters":        {"🇨🇿"},
	"LRN":                     {"🇲🇽", "🇨🇴"},
	// "NLC":                     {"🇩🇰", "🇫🇮", "🇸🇪", "🇳🇴", "🇬🇧", "🇮🇪"},
	"NLC":          {"🇬🇧", "🇮🇪"},
	"Rift Legends": {"🇵🇱"},
}
