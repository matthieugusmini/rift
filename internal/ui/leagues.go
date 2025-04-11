package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
)

type leagueItem struct {
	id         string
	leagueName string
}

func (i leagueItem) Title() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(white).
		Bold(true)

	flags := strings.Join(flagsByLeagueName[i.leagueName], separatorBullet)
	title := flags + " " + i.leagueName
	return titleStyle.Render(title)
}

func (i leagueItem) Description() string { return "" }

func (i leagueItem) FilterValue() string { return i.leagueName }

func newLeagueChoicesList(leagues []*lolesports.League, width, height int) list.Model {
	var leagueItems []list.Item
	for _, l := range leagues {
		item := leagueItem{
			id:         l.ID,
			leagueName: l.Name,
		}
		leagueItems = append(leagueItems, item)
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Border(lipgloss.ThickBorder(), false, false, false, true).BorderForeground(white)
	l := list.New(leagueItems, delegate, width, height)
	l.Title = "LEAGUES"
	l.Styles.Title = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(white).
		Background(charcoal).
		Bold(true)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return l
}
