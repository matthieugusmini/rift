package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matthieugusmini/lolesport/internal/lolesport"
	"github.com/matthieugusmini/lolesport/internal/ui"
)

func main() {
	lolesportClient := lolesport.NewClient()

	m := ui.NewModel(lolesportClient)

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}
