package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/lolesport/internal/timeutils"
)

func newStandingsViewport(stage lolesports.Stage, width, height int) viewport.Model {
	var (
		standingsTables []table.Model
	)
	for _, section := range stage.Sections {
		standingsTable := newStandingsTable(section.Rankings, width-2)
		standingsTables = append(standingsTables, standingsTable)
	}

	var sb strings.Builder

	for i, t := range standingsTables {
		title := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Margin(1, 2).
			Bold(true).
			Background(white).
			Foreground(black).
			Render(stage.Sections[i].Name)
		sb.WriteString("\n")
		sb.WriteString(title)
		sb.WriteString(
			lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(t.View()),
		)
		sb.WriteString("\n")
	}

	v := viewport.New(width, height)
	v.SetContent(sb.String())

	return v
}

func newStandingsTable(rankings []lolesports.Ranking, width int) table.Model {
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
	tableStyles.Selected = tableStyles.Selected.Foreground(gold)
	tableStyles.Cell = tableStyles.Cell.
		Width(columnWidth).
		Align(lipgloss.Center).
		Foreground(white).
		Bold(true)
	tableStyles.Header = tableStyles.Header.
		Align(lipgloss.Center).
		Width(columnWidth).
		Background(charcoal).
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
	case timeutils.IsCurrentTimeBetween(startDate, endDate):
		return tournmaentStateInProgress
	case now.After(endDate):
		return tournmaentStateCompleted
	default:
		return tournmaentStateUnknown
	}
}
