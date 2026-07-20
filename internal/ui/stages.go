package ui

import (
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
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
	width, height int,
	theme theme,
) list.Model {
	stageItems := make([]list.Item, len(stages))
	for i, stage := range stages {
		item := stageItem{
			name:      stage.Name,
			stageType: getStageType(stage),
		}
		stageItems[i] = item
	}

	stageItemDelegate := newStageItemDelegate(theme)

	l := list.New(stageItems, stageItemDelegate, width, height)
	applyStageOptionsTheme(&l, theme)
	l.Title = "STAGES"
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetSpinner(spinner.Meter)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.StatusMessageLifetime = time.Second * 2

	return l
}

func applyStageOptionsTheme(l *list.Model, theme theme) {
	applyListTheme(l, theme)
	l.SetDelegate(newStageItemDelegate(theme))
	l.Styles.Title = selectionListTitleStyle(theme)
}

type stageItemStyles struct {
	list.DefaultItemStyles

	disabledTitle         lipgloss.Style
	disabledDesc          lipgloss.Style
	disabledSelectedTitle lipgloss.Style
	disabledSelectedDesc  lipgloss.Style
}

func newStageItemStyles(theme theme) (s stageItemStyles) {
	defaultStyles := list.NewDefaultItemStyles(theme.isDark)

	s.DefaultItemStyles = defaultStyles

	// Selected
	s.SelectedTitle = defaultStyles.SelectedTitle.
		Foreground(theme.selected).
		Bold(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.selected)

	s.SelectedDesc = defaultStyles.SelectedDesc.
		Foreground(theme.textSecondary).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.selected)

	// Disabled but selected
	s.disabledSelectedTitle = defaultStyles.SelectedTitle.
		Foreground(theme.textDisabled).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.textDisabled)

	s.disabledSelectedDesc = defaultStyles.SelectedDesc.
		Foreground(theme.textDisabled).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.textDisabled)

	// Disabled not selected
	s.disabledTitle = defaultStyles.NormalTitle.
		Foreground(theme.textDisabled)

	s.disabledDesc = defaultStyles.NormalDesc.
		Foreground(theme.textDisabled)

	return s
}

type stageItemDelegate struct {
	list.DefaultDelegate

	styles stageItemStyles
}

func newStageItemDelegate(theme theme) stageItemDelegate {
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.Styles = list.NewDefaultItemStyles(theme.isDark)

	return stageItemDelegate{
		DefaultDelegate: defaultDelegate,
		styles:          newStageItemStyles(theme),
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
