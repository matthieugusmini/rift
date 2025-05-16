package ui

import (
	"time"

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
			name:      stage.Name,
			stageType: getStageType(stage),
		}
		stageItems[i] = item
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(selectedColor).
		Bold(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(selectedColor)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(textSecondaryColor).
		Bold(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(selectedColor)

	l := list.New(stageItems, delegate, width, height)
	l.Title = "STAGES"
	l.Styles.Title = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(textTitleColor).
		Background(secondaryBackgroundColor).
		Bold(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetSpinner(spinner.Meter)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.StatusMessageLifetime = time.Second * 2

	return l
}

func getStageType(stage lolesports.Stage) stageType {
	if len(stage.Sections) > 0 && len(stage.Sections[0].Rankings) == 0 {
		return stageTypeBracket
	}
	return stageTypeGroups
}
