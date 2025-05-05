package main

import (
	"fmt"
	"log/slog"
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

const logFile = "app.log"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	logFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}
	//nolint
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	cacheDB, err := bbolt.Open("rift.db", 0o600, bbolt.DefaultOptions)
	if err != nil {
		return fmt.Errorf("could not open the cache database: %w", err)
	}
	defer cacheDB.Close() //nolint

	bracketTemplateCache := cache.NewCache(cacheDB, 10*time.Hour, logger)

	bracketTemplateClient := github.NewBracketTemplateClient(http.DefaultClient)
	bracketTemplateLoader := github.NewBracketTemplateLoader(
		bracketTemplateClient,
		bracketTemplateCache,
		logger.WithGroup("bracketTemplateLoader"),
	)

	lolesportsClient := lolesports.NewClient(gololesports.NewClient())

	m := ui.NewModel(lolesportsClient, bracketTemplateLoader)

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
