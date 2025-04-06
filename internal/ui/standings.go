package ui

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/lolesport/internal/lolesport"
)

type standingsStyles struct {
	tableBorder      lipgloss.Style
	stageName        lipgloss.Style
	tournamentState  lipgloss.Style
	tournamentPeriod lipgloss.Style
	tournamentType   lipgloss.Style
}

func newDefaultStandingsStyles() standingsStyles {
	stageNameStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(white)).
		Bold(true)

	tournmaentStateStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color(white))

	tournamentPeriodStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(white))

	tournamentTypeStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(white))

	tableBorderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return standingsStyles{
		stageName:        stageNameStyle,
		tournamentState:  tournmaentStateStyle,
		tournamentPeriod: tournamentPeriodStyle,
		tournamentType:   tournamentTypeStyle,
		tableBorder:      tableBorderStyle,
	}
}

type standingsPageState int

const (
	standingsPageStateUnknown standingsPageState = iota
	standingsPageStateLeagueChoice
	standingsPageStateStandings
)

type standingsModel struct {
	lolesportClient LoLEsportClient

	leagueChoices list.Model

	stageName     string
	startDate     time.Time
	endDate       time.Time
	table         table.Model
	height, width int

	styles standingsStyles

	state standingsPageState
}

func newStandingsModel(lolesportClient LoLEsportClient) *standingsModel {
	return &standingsModel{
		lolesportClient: lolesportClient,
		styles:          newDefaultStandingsStyles(),
		state:           standingsPageStateLeagueChoice,
	}
}

func (m *standingsModel) Init() tea.Cmd { return nil }

func (m *standingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []*lolesport.Tournament:
		now := time.Now()
		i := slices.IndexFunc(msg, func(tournament *lolesport.Tournament) bool {
			if now.After(tournament.StartDate.Time) && now.Before(tournament.EndDate.Time) {
				return true
			}
			return false
		})
		tournament := msg[i]

		m.stageName = tournament.Slug
		m.startDate = tournament.StartDate.Time
		m.endDate = tournament.EndDate.Time

		tournamentID := msg[i].ID
		return m, m.getStandings(tournamentID)

	case []*lolesport.League:
		m.leagueChoices = newLeagueChoicesList(msg, m.width, m.height)

	case []*lolesport.Standings:
		m.table = newStandingsTable(msg[0].Stages[0].Sections[0].Rankings, m.width)

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.state = standingsPageStateStandings
			return m, m.getTournamentsForLeague(m.leagueChoices.SelectedItem().(leagueItem).id)

		case "esc":
			m.state = standingsPageStateLeagueChoice
			return m, m.getLeagues()
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.setSize(msg.Width-h, msg.Height-v-headersHeight)
	// TODO: Trigger update of the table

	// TODO: Search for a better way to do this.
	case state:
		if msg == stateShowStandings {
			m.state = standingsPageStateLeagueChoice
			return m, m.getLeagues()
		}
	}

	switch m.state {
	case standingsPageStateLeagueChoice:
		m.leagueChoices, cmd = m.leagueChoices.Update(msg)
	case standingsPageStateStandings:
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m *standingsModel) View() string {
	if m.state == standingsPageStateLeagueChoice {
		return m.leagueChoices.View()
	}

	var sb strings.Builder

	stageName := m.styles.stageName.Width(m.width).Render(m.stageName)
	sb.WriteString(stageName)
	sb.WriteString("\n\n")

	tournamentState := computeTournamentState(m.startDate, m.endDate)
	tournamentState = m.styles.tournamentState.Render(tournamentState)

	tournamentPeriod := formatTournamentPeriod(m.startDate, m.endDate)
	tournamentPeriod = m.styles.tournamentPeriod.Render(tournamentPeriod)

	tournamentType := m.styles.tournamentType.Render("🌍 REGIONAL")

	info := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(
			strings.Join([]string{
				tournamentState,
				tournamentPeriod,
				tournamentType,
			}, " • "),
		)
	sb.WriteString(info)
	sb.WriteString("\n\n")

	table := m.styles.tableBorder.Render(m.table.View())
	sb.WriteString(table)

	docStyle := lipgloss.NewStyle().
		Height(m.height).
		AlignVertical(lipgloss.Center)

	return docStyle.Render(sb.String())
}

func (m *standingsModel) setSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *standingsModel) getStandings(tournamentID string) tea.Cmd {
	return func() tea.Msg {
		standings, err := m.lolesportClient.GetStandings(context.Background(), tournamentID)
		if err != nil {
			return err
		}
		return standings
	}
}

func (m *standingsModel) getLeagues() tea.Cmd {
	return func() tea.Msg {
		leagues, err := m.lolesportClient.GetLeagues(context.Background())
		if err != nil {
			return err
		}
		return leagues
	}
}

func (m *standingsModel) getTournamentsForLeague(leagueID string) tea.Cmd {
	return func() tea.Msg {

		tournaments, err := m.lolesportClient.GetTournamentsForLeague(context.Background(), leagueID)
		if err != nil {
			return err
		}
		return tournaments
	}
}

func newStandingsTable(rankings []lolesport.Ranking, width int) table.Model {
	var (
		headerTitles = []string{"Ranking", "Team", "Series Win / Loss", "Win / Loss %"}
		columnWidth  = width / len(headerTitles)
		columns      []table.Column
	)
	for _, title := range headerTitles {
		columns = append(columns, table.Column{Title: title, Width: columnWidth})
	}

	var rows []table.Row
	for _, ranking := range rankings {
		for _, team := range ranking.Teams {
			seriesWinAndLoss := fmt.Sprintf("%dW - %dL", team.Record.Wins, team.Record.Losses)
			winrate := fmt.Sprintf("%d%%", calculateWinrate(team.Record.Wins, team.Record.Losses))
			row := table.Row{
				strconv.Itoa(ranking.Ordinal),
				team.Code,
				seriesWinAndLoss,
				winrate,
			}
			rows = append(rows, row)
		}
	}

	tableStyles := table.DefaultStyles()
	tableStyles.Selected = tableStyles.Selected.Foreground(lipgloss.Color(gold))
	tableStyles.Cell = tableStyles.Cell.
		Width(columnWidth).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(white)).
		Bold(true)
	tableStyles.Header = tableStyles.Header.
		Align(lipgloss.Center).
		Width(columnWidth).
		Background(lipgloss.Color(charcoal)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithWidth(width),
		table.WithHeight(len(rows)+1),
		table.WithStyles(tableStyles),
		table.WithFocused(true), // FIX: Cannot move in the table
	)

	return t
}

func calculateWinrate(wins, losses int) int {
	totalGames := wins + losses
	if totalGames == 0 {
		return 0
	}
	return int(float64(wins) / float64(totalGames) * 100)
}

func formatTournamentPeriod(startDate, endDate time.Time) string {
	startMonth := startDate.Format("January")
	endMonth := endDate.Format("January")
	return fmt.Sprintf("📅 %3s-%3s", startMonth, endMonth)
}

type tournmaentState string

const (
	tournmaentStateUnknown    = "UNKNOWN"
	tournmaentStateNotStarted = "UPCOMING"
	tournmaentStateInProgress = "IN PROGRESS"
	tournmaentStateCompleted  = "COMPLETED"
)

func computeTournamentState(startDate, endDate time.Time) string {
	now := time.Now()
	switch {
	case now.Before(startDate):
		return tournmaentStateNotStarted
	case now.After(startDate) && now.Before(endDate):
		return tournmaentStateInProgress
	case now.After(endDate):
		return tournmaentStateCompleted
	default:
		return tournmaentStateUnknown
	}
}

type leagueItem struct {
	id         string
	leagueName string
}

func (i leagueItem) Title() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(white)).
		Bold(true)

	title := strings.Join(flagsByLeagueName[i.leagueName], bulletSeparator) + " " + i.leagueName
	return titleStyle.Render(title)
}

func (i leagueItem) Description() string { return "" }

func (i leagueItem) FilterValue() string { return i.leagueName }

func newLeagueChoicesList(leagues []*lolesport.League, width, height int) list.Model {
	isShownLeagues := map[string]bool{
		"LEC":       true,
		"LTA North": true,
		"LTA South": true,
		"LCK":       true,
		"LPL":       true,
		"LCP":       true,
	}

	var leagueItems []list.Item
	for _, l := range leagues {
		if !isShownLeagues[l.Name] {
			continue
		}

		leagueItems = append(leagueItems, leagueItem{
			id:         l.ID,
			leagueName: l.Name,
		})
	}

	l := list.New(leagueItems, list.NewDefaultDelegate(), width, height)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	return l
}
