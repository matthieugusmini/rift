package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/lolesport/internal/timeutils"
)

func newSplitOptionsList(splits []lolesports.Split, width, height int) list.Model {
	var (
		items       = make([]list.Item, len(splits))
		cursorIndex int
	)
	for i, split := range splits {
		item := splitItem{
			// id:          split.ID,
			name:      split.Name,
			splitType: splitType(split.Region),
			// tournaments: split.Tournaments,
			startTime: split.StartTime,
			endTime:   split.EndTime,
		}
		items[i] = item

		if timeutils.IsCurrentTimeBetween(split.StartTime, split.EndTime) {
			cursorIndex = i
		}
	}

	l := list.New(items, newSplitItemDelegate(), width, height)
	l.Select(cursorIndex)
	l.Title = "EVENTS"
	l.Styles.Title = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(white).
		Background(charcoal).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)

	return l
}

type splitType string

const (
	splitTypeRegional splitType = "REGIONAL"
	splitTypeGlobal   splitType = "GLOBAL"
)

type splitItem struct {
	// id        string
	name      string
	splitType splitType

	// TODO: Maybe move somewhere else?
	startTime time.Time
	endTime   time.Time
	// tournaments []*lolesports.Tournament
}

func (i splitItem) FilterValue() string {
	return i.name
}

func (i splitItem) Title() string {
	return i.name
}

func (i splitItem) Description() string {
	return string(i.splitType) + " EVENT"
}

type splitItemStyles struct {
	normalTitle         lipgloss.Style
	upcomingNormalTitle lipgloss.Style
	selectedTitle       lipgloss.Style

	normalDescription         lipgloss.Style
	upcomingNormalDescription lipgloss.Style
	lastNormalDescription     lipgloss.Style
}

func newSplitItemStyles() (s splitItemStyles) {
	baseTitleStyle := lipgloss.NewStyle().
		Padding(0, 0, 0, 2).
		BorderLeft(true).
		Foreground(white).
		Bold(true)

	s.normalTitle = baseTitleStyle.
		BorderStyle(lipgloss.Border{Left: "◉"}).
		BorderForeground(white)

	s.upcomingNormalTitle = baseTitleStyle.
		BorderStyle(lipgloss.Border{Left: "◯"}).
		BorderForeground(charcoal)

	s.selectedTitle = baseTitleStyle.
		BorderStyle(lipgloss.Border{Left: "◉"}).
		BorderForeground(red)

	baseDescStyle := lipgloss.NewStyle().
		Foreground(gray).
		Bold(true)

	s.normalDescription = baseDescStyle.
		Padding(0, 0, 0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(white)

	s.upcomingNormalDescription = baseDescStyle.
		Padding(0, 0, 0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(charcoal)

	s.lastNormalDescription = baseDescStyle.
		Padding(0, 0, 0, 3)

	return s
}

type splitItemDelegate struct {
	styles splitItemStyles
}

func newSplitItemDelegate() splitItemDelegate {
	return splitItemDelegate{
		styles: newSplitItemStyles(),
	}
}

func (d splitItemDelegate) Height() int { return 2 }

func (d splitItemDelegate) Spacing() int { return 0 }

func (d splitItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d splitItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(splitItem)
	if !ok {
		return
	}
	title, desc := i.Title(), i.Description()

	var (
		titleStyle, descStyle = d.styles.normalTitle, d.styles.normalDescription

		now        = time.Now()
		isUpcoming = i.startTime.After(now)
		isCurrent  = timeutils.IsCurrentTimeBetween(i.startTime, i.endTime)
		isSelected = index == m.Index()
		isLast     = index == len(m.Items())-1
	)
	switch {
	case isSelected:
		titleStyle = d.styles.selectedTitle
		if isLast {
			descStyle = d.styles.lastNormalDescription
		} else if isCurrent || isUpcoming {
			descStyle = d.styles.upcomingNormalDescription
		}

	case isLast:
		titleStyle = d.styles.upcomingNormalTitle
		descStyle = d.styles.lastNormalDescription

	case isCurrent:
		descStyle = d.styles.upcomingNormalDescription

	case isUpcoming:
		titleStyle = d.styles.upcomingNormalTitle
		descStyle = d.styles.upcomingNormalDescription
	}

	fmt.Fprintf(w, "%s\n%s", titleStyle.Render(title), descStyle.Render(desc))
}
