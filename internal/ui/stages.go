package ui

import (
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	disabled  bool
}

func (i stageItem) Title() string { return i.name }

func (i stageItem) Description() string { return string(i.stageType) }

func (i stageItem) FilterValue() string { return i.name }

func (i stageItem) isDisabled() bool { return i.disabled }

func newStageOptionsList(
	stages []lolesports.Stage,
	availableStages []string,
	width, height int,
) list.Model {
	stageItems := make([]list.Item, len(stages))
	for i, stage := range stages {
		item := stageItem{
			name:      stage.Name,
			stageType: getStageType(stage),
			disabled:  !isStageAvailable(stage, availableStages),
		}
		stageItems[i] = item
	}

	stageItemDelegate := newStageItemDelegate()

	l := list.New(stageItems, stageItemDelegate, width, height)
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

type stageItemStyles struct {
	list.DefaultItemStyles

	disabledTitle         lipgloss.Style
	disabledDesc          lipgloss.Style
	disabledSelectedTitle lipgloss.Style
	disabledSelectedDesc  lipgloss.Style
}

func newStageItemStyles() (s stageItemStyles) {
	defaultStyles := list.NewDefaultItemStyles()

	s.DefaultItemStyles = defaultStyles

	// Selected
	s.SelectedTitle = defaultStyles.SelectedTitle.
		Foreground(selectedColor).
		Bold(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(selectedColor)

	s.SelectedDesc = defaultStyles.SelectedDesc.
		Foreground(textSecondaryColor).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(selectedColor)

	// Disabled but selected
	s.disabledSelectedTitle = defaultStyles.SelectedTitle.
		Foreground(textDisabledColor).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(textDisabledColor)

	s.disabledSelectedDesc = defaultStyles.SelectedDesc.
		Foreground(textDisabledColor).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(textDisabledColor)

	// Disabled not selected
	s.disabledTitle = defaultStyles.NormalTitle.
		Foreground(textDisabledColor)

	s.disabledDesc = defaultStyles.NormalDesc.
		Foreground(textDisabledColor)

	return s
}

type stageItemDelegate struct {
	list.DefaultDelegate

	styles stageItemStyles
}

func newStageItemDelegate() stageItemDelegate {
	return stageItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          newStageItemStyles(),
	}
}

func (d stageItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if m.Width() <= 0 {
		// short-circuit
		return
	}

	i, ok := item.(stageItem)
	if !ok {
		return
	}

	var (
		title      = i.Title()
		desc       = i.Description()
		isDisabled = i.isDisabled()
		s          = &d.styles
	)

	// Prevent text from exceeding list width
	textWidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()
	title = ansi.Truncate(title, textWidth, "…")
	desc = ansi.Truncate(desc, textWidth, "…")

	isSelected := index == m.Index()
	switch {
	case isDisabled && isSelected:
		title = s.disabledSelectedTitle.Render(title)
		desc = s.disabledSelectedDesc.Render(desc)

	case isDisabled && !isSelected:
		title = s.disabledTitle.Render(title)
		desc = s.disabledDesc.Render(desc)

	case !isDisabled && isSelected:
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)

	case !isDisabled && !isSelected:
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func getStageType(stage lolesports.Stage) stageType {
	if len(stage.Sections) > 0 && len(stage.Sections[0].Rankings) == 0 {
		return stageTypeBracket
	}
	return stageTypeGroups
}

func isStageAvailable(stage lolesports.Stage, availableStages []string) bool {
	if getStageType(stage) == stageTypeBracket {
		return slices.Contains(availableStages, stage.ID)
	}
	return true
}
