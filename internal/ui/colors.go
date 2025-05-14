package ui

import "github.com/charmbracelet/lipgloss"

var (
	black          = "#000000"
	white          = "#ffffff"
	antiFlashWhite = "#f2f3f4"
	eerieBlack     = "#1f1f1f"
	lightGray      = "#848482"
	sonicSilver    = "#757575"
	gold           = "#ffd700"
	dimGrey        = "#696969"
	darkVanilla    = "#d1bea8"
	imperialRed    = "#ed2939"
	crimson        = "#dc143c"
	neonFuchsia    = "#fe4164"
)

var (
	blackAndWhite            = lipgloss.AdaptiveColor{Light: black, Dark: white}
	textPrimaryColor         = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}
	textSecondaryColor       = lipgloss.AdaptiveColor{Light: sonicSilver, Dark: lightGray}
	textDimmedSecondaryColor = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}
	borderPrimaryColor       = lipgloss.AdaptiveColor{Light: eerieBlack, Dark: white}
	borderSecondaryColor     = lipgloss.AdaptiveColor{Light: lightGray, Dark: dimGrey}
	selectedColor            = lipgloss.AdaptiveColor{Light: neonFuchsia, Dark: gold}
	secondaryBackgroundColor = lipgloss.AdaptiveColor{Light: darkVanilla, Dark: imperialRed}
	red                      = lipgloss.AdaptiveColor{Light: crimson, Dark: imperialRed}
	spinnerColor             = lipgloss.AdaptiveColor{Light: neonFuchsia, Dark: gold}
)
