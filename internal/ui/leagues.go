package ui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/matthieugusmini/go-lolesports"
)

type leagueItem struct {
	id         string
	leagueName string
}

func (i leagueItem) Title() string {
	flags := strings.Join(flagsByLeagueName[i.leagueName], separatorBullet)
	return i.leagueName + separatorBullet + flags
}

func (i leagueItem) Description() string { return "" }

func (i leagueItem) FilterValue() string { return i.leagueName }

type leagueItemStyles struct {
	normalTitle   lipgloss.Style
	selectedTitle lipgloss.Style
}

func newDefaultLeageItemStyles(theme theme) (s leagueItemStyles) {
	s.normalTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 2).
		Foreground(theme.textPrimary)

	s.selectedTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 1).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(theme.selected).
		Foreground(theme.selected).
		Bold(true)

	return s
}

type leagueItemDelegate struct {
	styles leagueItemStyles
}

func newLeagueItemDelegate(theme theme) leagueItemDelegate {
	return leagueItemDelegate{
		styles: newDefaultLeageItemStyles(theme),
	}
}

func (d leagueItemDelegate) Height() int { return 1 }

func (d leagueItemDelegate) Spacing() int { return 1 }

func (d leagueItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d leagueItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	leagueItem, ok := item.(leagueItem)
	if !ok {
		return
	}

	var (
		title      string
		isSelected = index == m.Index()
	)
	if isSelected {
		title = d.styles.selectedTitle.Render(leagueItem.Title())
	} else {
		title = d.styles.normalTitle.Render(leagueItem.Title())
	}

	fmt.Fprint(w, title)
}

func newLeagueOptionsList(
	leagues []lolesports.League,
	width, height int,
	theme theme,
) list.Model {
	leagueItems := make([]list.Item, len(leagues))
	for i, l := range leagues {
		leagueItems[i] = leagueItem{
			id:         l.ID,
			leagueName: l.Name,
		}
	}

	l := list.New(leagueItems, newLeagueItemDelegate(theme), width, height)
	applyLeagueOptionsTheme(&l, theme)
	l.Title = "LEAGUES"
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return l
}

func applyLeagueOptionsTheme(l *list.Model, theme theme) {
	applyListTheme(l, theme)
	l.SetDelegate(newLeagueItemDelegate(theme))
	l.Styles.Title = selectionListTitleStyle(theme)
}
