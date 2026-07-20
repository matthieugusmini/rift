package ui

import (
	"image/color"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

const (
	// Black
	black       = "#000000"
	almostBlack = "#1a1a1a"
	eerieBlack  = "#1f1f1f"

	// White
	white          = "#ffffff"
	antiFlashWhite = "#f2f3f4"
	darkVanilla    = "#d1bea8"

	// Grey
	grey        = "#808080"
	lightGrey   = "#dddddd"
	dimGrey     = "#777777"
	sonicSilver = "#757575"

	// Red
	imperialRed = "#ed2939"
	crimson     = "#dc143c"
	neonFuchsia = "#fe4164"

	// Yellow
	gold = "#ffd700"
)

type theme struct {
	isDark bool

	textPrimary         color.Color
	textSecondary       color.Color
	textDimmedSecondary color.Color
	textDisabled        color.Color
	textTitle           color.Color

	borderPrimary   color.Color
	borderSecondary color.Color

	secondaryBackground color.Color
	selected            color.Color
	red                 color.Color
	spinner             color.Color
}

func newTheme(isDark bool) theme {
	lightDark := lipgloss.LightDark(isDark)

	return theme{
		isDark: isDark,

		textPrimary:         lightDark(lipgloss.Color(almostBlack), lipgloss.Color(lightGrey)),
		textSecondary:       lightDark(lipgloss.Color(sonicSilver), lipgloss.Color(grey)),
		textDimmedSecondary: lightDark(lipgloss.Color("#A49FA5"), lipgloss.Color(dimGrey)),
		textDisabled:        lightDark(lipgloss.Color("#555156"), lipgloss.Color("#505050")),
		textTitle:           lightDark(lipgloss.Color(black), lipgloss.Color(white)),

		borderPrimary:   lightDark(lipgloss.Color(eerieBlack), lipgloss.Color(white)),
		borderSecondary: lightDark(lipgloss.Color(grey), lipgloss.Color(dimGrey)),

		secondaryBackground: lightDark(lipgloss.Color(darkVanilla), lipgloss.Color(imperialRed)),
		selected:            lightDark(lipgloss.Color(neonFuchsia), lipgloss.Color(gold)),
		red:                 lightDark(lipgloss.Color(crimson), lipgloss.Color(imperialRed)),
		spinner:             lightDark(lipgloss.Color(neonFuchsia), lipgloss.Color(gold)),
	}
}

func newHelp(theme theme) help.Model {
	m := help.New()
	m.Styles = help.DefaultStyles(theme.isDark)

	return m
}

func applyListTheme(m *list.Model, theme theme) {
	m.Styles = list.DefaultStyles(theme.isDark)
	m.Help.Styles = help.DefaultStyles(theme.isDark)
	m.FilterInput.SetStyles(textinput.DefaultStyles(theme.isDark))
}

func selectionListTitleStyle(theme theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(theme.textTitle).
		Background(theme.secondaryBackground).
		Bold(true)
}
