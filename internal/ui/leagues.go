package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
)

type leagueItem struct {
	id         string
	leagueName string
}

func (i leagueItem) Title() string {
	flags := strings.Join(flagsByLeagueName[i.leagueName], separatorBullet)
	title := flags + " " + i.leagueName
	return title
}

func (i leagueItem) Description() string { return "" }

func (i leagueItem) FilterValue() string { return i.leagueName }

type leagueItemStyles struct {
	normalTitle   lipgloss.Style
	selectedTitle lipgloss.Style
}

func newDefaultLeageItemStyles() (s leagueItemStyles) {
	s.normalTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 2).
		Foreground(white).
		Faint(true)

	s.selectedTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 1).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(white).
		Foreground(white).
		Bold(true)

	return s
}

type leagueItemDelegate struct {
	styles leagueItemStyles
}

func newLeagueItemDelegate() leagueItemDelegate {
	return leagueItemDelegate{
		styles: newDefaultLeageItemStyles(),
	}
}

func (d leagueItemDelegate) Height() int { return 1 }

func (d leagueItemDelegate) Spacing() int { return 1 }

func (d leagueItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d leagueItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var title string
	if leagueItem, ok := item.(leagueItem); ok {
		title = leagueItem.Title()
	} else {
		return
	}

	if m.Width() <= 0 {
		return
	}

	isSelected := index == m.Index()
	if isSelected {
		title = d.styles.selectedTitle.Render(title)
	} else {
		title = d.styles.normalTitle.Render(title)
	}
	fmt.Fprint(w, title)
}

func newLeagueOptionsList(leagues []*lolesports.League, width, height int) list.Model {
	leagueItems := make([]list.Item, len(leagues))
	for i, l := range leagues {
		leagueItems[i] = leagueItem{
			id:         l.ID,
			leagueName: l.Name,
		}
	}

	l := list.New(leagueItems, newLeagueItemDelegate(), width, height)
	l.Title = "LEAGUES"
	// TODO: Where should we define this style?
	l.Styles.Title = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(white).
		Background(charcoal).
		Bold(true)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return l
}
