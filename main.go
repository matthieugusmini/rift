package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gololesports "github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/internal/lolesports"
	"github.com/matthieugusmini/lolesport/internal/ui"
)

func main() {
	lolesportsClient := lolesports.NewClient(gololesports.NewClient())

	m := ui.NewModel(lolesportsClient)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}
