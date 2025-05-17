package ui

import "github.com/charmbracelet/lipgloss"

var (
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

var (
	textPrimaryColor         = lipgloss.AdaptiveColor{Light: almostBlack, Dark: lightGrey}
	textSecondaryColor       = lipgloss.AdaptiveColor{Light: sonicSilver, Dark: grey}
	textDimmedSecondaryColor = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: dimGrey}
	textTitleColor           = lipgloss.AdaptiveColor{Light: black, Dark: white}

	borderPrimaryColor   = lipgloss.AdaptiveColor{Light: eerieBlack, Dark: white}
	borderSecondaryColor = lipgloss.AdaptiveColor{Light: grey, Dark: dimGrey}

	secondaryBackgroundColor = lipgloss.AdaptiveColor{Light: darkVanilla, Dark: imperialRed}

	selectedColor = lipgloss.AdaptiveColor{Light: neonFuchsia, Dark: gold}

	red = lipgloss.AdaptiveColor{Light: crimson, Dark: imperialRed}

	spinnerColor = lipgloss.AdaptiveColor{Light: neonFuchsia, Dark: gold}
)
