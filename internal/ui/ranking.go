package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/internal/timeutils"
)

func newStandingsViewport(stage lolesports.Stage, width, height int) viewport.Model {
	standingsTables := make([]*table.Table, len(stage.Sections))
	for i, section := range stage.Sections {
		standingsTables[i] = newStandingsTable(section.Rankings, width)
	}

	var sb strings.Builder

	for i, t := range standingsTables {
		title := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Padding(0, 1).
			Bold(true).
			Background(secondaryBackgroundColor).
			Foreground(textPrimaryColor).
			Render(stage.Sections[i].Name)
		sb.WriteString(title + "\n")
		sb.WriteString(t.Render())
		sb.WriteString("\n\n")
	}

	v := viewport.New(width, height)
	v.SetContent(sb.String())

	return v
}

func newStandingsTable(rankings []lolesports.Ranking, width int) *table.Table {
	headers := []string{"Ranking", "Team", "Series Win / Loss", "Win / Loss %"}

	var rows [][]string
	for _, ranking := range rankings {
		for _, team := range ranking.Teams {
			seriesWinAndLoss := fmt.Sprintf("%dW - %dL", team.Record.Wins, team.Record.Losses)
			winrate := fmt.Sprintf("%d%%", calculateWinrate(team.Record.Wins, team.Record.Losses))
			row := []string{
				strconv.Itoa(ranking.Ordinal),
				team.Code,
				seriesWinAndLoss,
				winrate,
			}
			rows = append(rows, row)
		}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(selectedBorderColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			// Headers
			case row != 0:
				return lipgloss.NewStyle().
					Align(lipgloss.Center).
					Foreground(textPrimaryColor).
					Bold(true)

			// Rows
			default:
				return lipgloss.NewStyle().
					Align(lipgloss.Center).
					Foreground(textSecondaryColor).
					Bold(true)
			}
		}).
		Headers(headers...).
		Rows(rows...).
		Width(width)
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
	return fmt.Sprintf("%s-%s", startMonth, endMonth)
}

type tournmaentState string

const (
	tournmaentStateUnknown    = "UNKNOWN"
	tournmaentStateNotStarted = "UPCOMING"
	tournmaentStateInProgress = "IN PROGRESS"
	tournmaentStateCompleted  = "COMPLETED"
)

func computeTournamentState(startDate, endDate time.Time) tournmaentState {
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
