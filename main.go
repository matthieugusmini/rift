package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gololesports "github.com/matthieugusmini/go-lolesports"
	"go.etcd.io/bbolt"

	"github.com/matthieugusmini/lolesport/cache"
	"github.com/matthieugusmini/lolesport/github"
	"github.com/matthieugusmini/lolesport/lolesports"
	"github.com/matthieugusmini/lolesport/ui"
)

func main() {
	var logFile = "app.log"

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open or create log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.SetPrefix("[rift] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cacheDB, err := bbolt.Open("rift.db", 0600, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open DB: %v\n", err)
		os.Exit(1)
	}
	defer cacheDB.Close()

	bracketTemplateCache := cache.NewCache(cacheDB, 10*time.Hour)

	bracketTemplateClient := github.NewBracketTemplateClient(http.DefaultClient)
	bracketTemplateLoader := github.NewBracketTemplateLoader(bracketTemplateClient, bracketTemplateCache)

	lolesportsClient := lolesports.NewClient(gololesports.NewClient())

	m := ui.NewModel(lolesportsClient, bracketTemplateLoader)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}
