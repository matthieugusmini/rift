package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gololesports "github.com/matthieugusmini/go-lolesports"
	gap "github.com/muesli/go-app-paths"
	"go.etcd.io/bbolt"

	"github.com/matthieugusmini/lolesport/cache"
	"github.com/matthieugusmini/lolesport/github"
	"github.com/matthieugusmini/lolesport/lolesports"
	"github.com/matthieugusmini/lolesport/rift"
	"github.com/matthieugusmini/lolesport/ui"
)

const (
	appName     = "rift"
	logFilename = "rift.log"
	cacheFile   = "rift.db"

	bucketBracketTemplate = "bracketTemplate"
	bucketStandings       = "standings"
	bucketSchedule        = "schedule"

	cacheDefaultTTL = time.Hour * 12
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	scope := gap.NewScope(gap.User, appName)

	logPath, err := scope.LogPath(logFilename)
	if err != nil {
		return fmt.Errorf("could not retrieve the log file path: %w", err)
	}

	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not make a new log directory in filesystem: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	cacheDir, err := scope.CacheDir()
	if err != nil {
		return fmt.Errorf("could not retrieve the user cache directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not make a new cache directory in filesystem: %w", err)
	}

	cachePath := filepath.Join(cacheDir, cacheFile)
	cacheDB, err := bbolt.Open(cachePath, 0o600, bbolt.DefaultOptions)
	if err != nil {
		return fmt.Errorf("could not open the cache database: %w", err)
	}
	defer cacheDB.Close()

	bracketTemplateClient := github.NewBracketTemplateClient(http.DefaultClient)
	bracketTemplateCache := cache.New[rift.BracketTemplate](
		cacheDB,
		bucketBracketTemplate,
		cacheDefaultTTL,
	)
	bracketTemplateLoader := github.NewBracketTemplateLoader(
		bracketTemplateClient,
		bracketTemplateCache,
		logger,
	)

	lolesportsAPIClient := lolesports.NewClient(gololesports.NewClient())
	standingsCache := cache.New[[]gololesports.Standings](
		cacheDB,
		bucketStandings,
		cacheDefaultTTL,
	)
	lolesportsLoader := lolesports.NewLoader(lolesportsAPIClient, standingsCache, logger)

	m := ui.NewModel(lolesportsLoader, bracketTemplateLoader)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
