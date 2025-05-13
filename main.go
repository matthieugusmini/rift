package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gololesports "github.com/matthieugusmini/go-lolesports"
	gap "github.com/muesli/go-app-paths"
	"go.etcd.io/bbolt"

	"github.com/matthieugusmini/rift/internal/cache"
	"github.com/matthieugusmini/rift/internal/github"
	"github.com/matthieugusmini/rift/internal/lolesports"
	"github.com/matthieugusmini/rift/internal/rift"
	"github.com/matthieugusmini/rift/internal/ui"
)

var (
	Version string

	Commit string
)

const appName = "rift"

const logFilename = "rift.log"

const (
	cacheFile = "rift.db"

	bucketBracketTemplate = "bracketTemplate"
	bucketStandings       = "standings"
	bucketSchedule        = "schedule"

	cacheDefaultTTL = 12 * time.Hour
)

const (
	httpClientDefaultTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	scope := gap.NewScope(gap.User, appName)

	logger, logFile, err := initLogger(scope)
	if err != nil {
		return fmt.Errorf("could not initialize the logger: %w", err)
	}
	defer logFile.Close()

	cacheDB, err := initCache(scope)
	if err != nil {
		return fmt.Errorf("could not initialize the cache: %w", err)
	}
	defer cacheDB.Close()

	httpClient := &http.Client{
		Timeout: httpClientDefaultTimeout,
	}

	bracketTemplateLoader := initBracketTemplateLoader(httpClient, cacheDB, logger)

	lolesportsLoader := initLoLEsportsLoader(httpClient, cacheDB, logger)

	m := ui.NewModel(lolesportsLoader, bracketTemplateLoader, logger)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func initLogger(scope *gap.Scope) (*slog.Logger, io.Closer, error) {
	logPath, err := scope.LogPath(logFilename)
	if err != nil {
		return nil, nil, fmt.Errorf("could not retrieve the log file path: %w", err)
	}

	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, nil, fmt.Errorf("could not make a new log directory in filesystem: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open log file: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	return logger, logFile, nil
}

func initCache(scope *gap.Scope) (*bbolt.DB, error) {
	cacheDir, err := scope.CacheDir()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the user cache directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("could not make a new cache directory in filesystem: %w", err)
	}

	cachePath := filepath.Join(cacheDir, cacheFile)
	cacheDB, err := bbolt.Open(cachePath, 0o600, bbolt.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("could not open the cache database: %w", err)
	}

	return cacheDB, nil
}

func initBracketTemplateLoader(
	httpClient *http.Client,
	cacheDB *bbolt.DB,
	logger *slog.Logger,
) *rift.BracketTemplateLoader {
	bracketTemplateClient := github.NewBracketTemplateClient(httpClient)

	bracketTemplateCache := cache.New[rift.BracketTemplate](
		cacheDB,
		bucketBracketTemplate,
		cacheDefaultTTL,
	)

	return rift.NewBracketTemplateLoader(
		bracketTemplateClient,
		bracketTemplateCache,
		logger,
	)
}

func initLoLEsportsLoader(
	_ *http.Client, // TODO: Add an option to configure go-lolesports internal http.Client.
	cacheDB *bbolt.DB,
	logger *slog.Logger,
) *rift.LoLEsportsLoader {
	lolesportsAPIClient := lolesports.NewClient(gololesports.NewClient())

	standingsCache := cache.New[[]gololesports.Standings](
		cacheDB,
		bucketStandings,
		cacheDefaultTTL,
	)

	return rift.NewLoLEsportsLoader(lolesportsAPIClient, standingsCache, logger)
}
