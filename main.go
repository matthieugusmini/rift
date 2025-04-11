package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gololesports "github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/lolesport/internal/lolesports"
	"github.com/matthieugusmini/lolesport/internal/ui"
)

func main() {
	f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	lolesportsClient := lolesports.NewClient(gololesports.NewClient())

	m := ui.NewModel(lolesportsClient)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}
