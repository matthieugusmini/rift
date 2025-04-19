package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
)

type stageType string

const (
	stageTypeGroups  stageType = "GROUPS"
	stageTypeBracket stageType = "BRACKET"
)

type stageItem struct {
	id        string
	name      string
	stageType stageType
}

func (i stageItem) Title() string { return i.name }

func (i stageItem) Description() string { return string(i.stageType) }

func (i stageItem) FilterValue() string { return i.name }

func newStageOptionsList(stages []lolesports.Stage, width, height int) list.Model {
	stageItems := make([]list.Item, len(stages))
	for i, stage := range stages {
		item := stageItem{
			id:        stage.ID,
			name:      stage.Name,
			stageType: getStageType(stage),
		}
		stageItems[i] = item
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(white).
		Bold(true).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(white)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(white).
		Bold(true).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(white)
	l := list.New(stageItems, delegate, width, height)
	l.Title = "STAGES"
	l.Styles.Title = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(white).
		Background(charcoal).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetSpinner(spinner.Meter)

	return l
}

func getStageType(stage lolesports.Stage) stageType {
	if len(stage.Sections) > 0 && len(stage.Sections[0].Rankings) == 0 {
		return stageTypeBracket
	} else {
		return stageTypeGroups
	}
}
