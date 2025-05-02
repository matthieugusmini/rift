package main

import (
	"fmt"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gololesports "github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/internal/github"
	"github.com/matthieugusmini/lolesport/internal/lolesports"
	"github.com/matthieugusmini/lolesport/internal/ui"
)

func main() {
	l := github.NewBracketTemplateLoader(http.DefaultClient)

	lolesportsClient := lolesports.NewClient(gololesports.NewClient())

	m := ui.NewModel(lolesportsClient, l)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}
