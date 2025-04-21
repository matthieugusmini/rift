package ui

import "github.com/charmbracelet/lipgloss"

var (
	white       = "#ffffff"
	eerieBlack  = "#1f1f1f"
	lightGray   = "#848482"
	sonicSilver = "#757575"
	gold        = "#ffd700"
	dimGrey     = "#696969"
	darkVanilla = "#d1bea8"
	imperialRed = "#ed2939"
	crimson     = "#dc143c"
	neonFuchsia = "#fe4164"
)

var (
	textPrimaryColor         = lipgloss.AdaptiveColor{Light: eerieBlack, Dark: white}
	textSecondaryColor       = lipgloss.AdaptiveColor{Light: sonicSilver, Dark: lightGray}
	borderPrimaryColor       = lipgloss.AdaptiveColor{Light: eerieBlack, Dark: white}
	borderSecondaryColor     = lipgloss.AdaptiveColor{Light: lightGray, Dark: dimGrey}
	selectedBorderColor      = lipgloss.AdaptiveColor{Light: neonFuchsia, Dark: gold}
	secondaryBackgroundColor = lipgloss.AdaptiveColor{Light: darkVanilla, Dark: dimGrey}
	red                      = lipgloss.AdaptiveColor{Light: crimson, Dark: imperialRed}
)
