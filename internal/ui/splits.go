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

type splitType string

const (
	splitTypeRegional splitType = "REGIONAL"
	splitTypeGlobal   splitType = "GLOBAL"
)

type splitItem struct {
	id     string
	name   string
	region splitType

	// TODO: Maybe move somewhere else
	startTime   time.Time
	endTime     time.Time
	tournaments []*lolesports.Tournament
}

func (i splitItem) FilterValue() string {
	return i.name
}

func (i splitItem) Title() string {
	return i.name
}

func (i splitItem) Description() string {
	return string(i.region) + " EVENT"
}

type splitItemStyles struct {
	normalTitle       lipgloss.Style
	normalDescription lipgloss.Style

	upcomingEventNormalTitle       lipgloss.Style
	upcomingEventNormalDescription lipgloss.Style

	selectedTitle       lipgloss.Style
	selectedDescription lipgloss.Style
}

func newSplitItemStyles() (s splitItemStyles) {
	baseTitleStyle := lipgloss.NewStyle().
		Padding(0, 0, 0, 2).
		BorderLeft(true).
		Bold(true)

	baseDescriptionStyle := lipgloss.NewStyle().
		Padding(0, 0, 0, 2).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		Bold(true)

	s.normalTitle = baseTitleStyle.
		BorderStyle(lipgloss.Border{Left: "◉"}).
		Foreground(white)

	s.normalDescription = baseDescriptionStyle.
		Foreground(gray)

	s.upcomingEventNormalTitle = baseTitleStyle.
		BorderStyle(lipgloss.Border{Left: "◯"}).
		BorderForeground(charcoal).
		Foreground(white)

	s.upcomingEventNormalDescription = baseDescriptionStyle.
		BorderForeground(charcoal).
		Foreground(gray)

	s.selectedTitle = baseTitleStyle.
		BorderStyle(lipgloss.Border{Left: "◉"}).
		BorderForeground(red).
		Foreground(white)

	s.selectedDescription = baseDescriptionStyle.
		Foreground(gray)

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
	)
	if isSelected {
		titleStyle = d.styles.selectedTitle
		if isCurrent || isUpcoming {
			descStyle = d.styles.upcomingEventNormalDescription
		} else {
			descStyle = d.styles.selectedDescription
		}
	} else {
		if isCurrent {
			descStyle = d.styles.upcomingEventNormalDescription
		} else if isUpcoming {
			titleStyle = d.styles.upcomingEventNormalTitle
			descStyle = d.styles.upcomingEventNormalDescription
		}
	}

	fmt.Fprintf(w, "%s\n%s", titleStyle.Render(title), descStyle.Render(desc))
}

func newSplitChoices(splits []*lolesports.Split, width, height int) list.Model {
	var (
		items       = make([]list.Item, len(splits))
		cursorIndex int
	)
	for i, split := range splits {
		item := splitItem{
			id:          split.ID,
			name:        split.Name,
			region:      splitType(split.Region),
			tournaments: split.Tournaments,
			startTime:   split.StartTime,
			endTime:     split.EndTime,
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
