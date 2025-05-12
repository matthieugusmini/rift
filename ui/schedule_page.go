package ui

import (
	"context"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/timeutil"
)

const (
	schedulePageShortHelpHeight = 1
	schedulePageFullHelpHeight  = 7
)

const (
	errMessageFetchInitialPage = "Oups! Looks like something went wrong...\nPress any key to try your luck again"
	errMessageFetchNextPage    = "Failed to fetch next events. Retry in a moment"
	errMessageFetchPrevPage    = "Failed to fetch previous events. Retry in a moment"
)

type schedulePageStyles struct {
	doc     lipgloss.Style
	title   lipgloss.Style
	spinner lipgloss.Style
	help    lipgloss.Style
	error   lipgloss.Style
}

func newDefaultSchedulePageStyles() (s schedulePageStyles) {
	s.doc = lipgloss.NewStyle().Padding(1, 2)

	s.title = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(textPrimaryColor).
		Background(secondaryBackgroundColor).
		Bold(true)

	s.spinner = lipgloss.NewStyle().Foreground(spinnerColor)

	s.help = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	s.error = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(textPrimaryColor).
		Italic(true)

	return s
}

type schedulePageKeyMap struct {
	list.KeyMap

	RevealSpoiler key.Binding
	NextPage      key.Binding
	PrevPage      key.Binding
}

func newDefaultSchedulePageKeyMap() schedulePageKeyMap {
	return schedulePageKeyMap{
		KeyMap: list.DefaultKeyMap(),
		RevealSpoiler: key.NewBinding(
			key.WithKeys("enter", "right"),
			key.WithHelp("enter", "reveal spoiler"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next page"),
		),
	}
}

type paginationState struct {
	prevPageToken   string
	loadingPrevPage bool

	nextPageToken   string
	loadingNextPage bool
}

func (s paginationState) hasNextPage() bool { return s.nextPageToken != "" }

func (s paginationState) hasPrevPage() bool { return s.prevPageToken != "" }

type schedulePage struct {
	lolesportsClient LoLEsportsLoader
	logger           *slog.Logger

	width, height int

	// List of all matches fetched so far, filtered to include only events
	// with the "match" type. Storing matches in memory maintains a clear
	// separation between data and UI, avoiding the need for type assertions
	// on the list.Model items.
	matches []lolesports.Event

	// UI representation of the matches
	matchList list.Model

	// Contains the information required to fetch schedule pages.
	paginationState paginationState

	// Indicates whether the schedule events have been fetched.
	loaded bool

	// Error message displayed to the user when failed to
	// load the initial page data.
	errMsg string

	help help.Model

	spinner spinner.Model

	keyMap schedulePageKeyMap
	styles schedulePageStyles
}

func newSchedulePage(lolesportsClient LoLEsportsLoader, logger *slog.Logger) *schedulePage {
	styles := newDefaultSchedulePageStyles()

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(styles.spinner),
	)

	return &schedulePage{
		lolesportsClient: lolesportsClient,
		logger:           logger,
		spinner:          sp,
		styles:           styles,
		keyMap:           newDefaultSchedulePageKeyMap(),
		help:             help.New(),
	}
}

func (p *schedulePage) Init() tea.Cmd {
	return tea.Batch(p.spinner.Tick, p.fetchEvents(pageDirectionInitial))
}

func (p *schedulePage) Update(msg tea.Msg) (*schedulePage, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// When an error is displayed after failing to fetch the initial schedule data,
		// any keypress should trigger a refetch of the initial data again.
		if p.errMsg != "" {
			p.errMsg = ""
			if !p.loaded {
				return p, tea.Batch(p.fetchEvents(pageDirectionInitial))
			}
		}

		switch {
		case key.Matches(msg, p.keyMap.ShowFullHelp),
			key.Matches(msg, p.keyMap.CloseFullHelp):
			p.toggleHelp()

		case msg.String() == "down":
			if p.shouldFetchNextPage() {
				p.paginationState.loadingNextPage = true
				cmds = append(cmds, p.matchList.StartSpinner(), p.fetchEventsNextPage())
			}

		case msg.String() == "up":
			if p.shouldFetchPreviousPage() {
				p.paginationState.loadingPrevPage = true
				cmds = append(cmds, p.matchList.StartSpinner(), p.fetchEventsPreviousPage())
			}
		}

	case spinner.TickMsg:
		if !p.loaded {
			var cmd tea.Cmd
			p.spinner, cmd = p.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case fetchedEventsMessage:
		p.handleFetchedEvents(msg)

	case fetchErrorMessage:
		cmd := p.handleFetchError(msg)
		cmds = append(cmds, cmd)
	}

	if !p.loaded {
		return p, tea.Batch(cmds...)
	}

	var cmd tea.Cmd

	p.updateMatchListTitle()

	p.matchList, cmd = p.matchList.Update(msg)
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *schedulePage) View() string {
	if p.errMsg != "" {
		return p.viewError()
	}

	var sections []string

	if !p.loaded {
		sections = append(sections, p.viewSpinner())
	} else {
		sections = append(sections, p.matchList.View())
	}
	sections = append(sections, p.viewHelp())

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return p.styles.doc.Render(view)
}

func (p *schedulePage) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.RevealSpoiler,
		p.keyMap.NextPage,
		p.keyMap.Quit,
		p.keyMap.ShowFullHelp,
	}
}

func (p *schedulePage) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			p.keyMap.CursorUp,
			p.keyMap.CursorDown,
			p.keyMap.NextPage,
			p.keyMap.PrevPage,
			p.keyMap.GoToStart,
			p.keyMap.GoToEnd,
			p.keyMap.RevealSpoiler,
		},
		{
			p.keyMap.Filter,
			p.keyMap.ClearFilter,
			p.keyMap.AcceptWhileFiltering,
			p.keyMap.CancelWhileFiltering,
		},
		{
			p.keyMap.Quit,
			p.keyMap.CloseFullHelp,
		},
	}
}

func (p *schedulePage) viewHelp() string {
	return p.styles.help.Render(p.help.View(p))
}

func (p *schedulePage) viewSpinner() string {
	return lipgloss.NewStyle().
		Width(p.width).
		Height(p.contentHeight()).
		Align(lipgloss.Center, lipgloss.Center).
		Render(p.spinner.View())
}

func (p *schedulePage) viewError() string {
	errMsg := p.styles.error.Render(p.errMsg)
	return p.styles.doc.
		Width(p.width).
		Height(p.contentHeight()).
		Align(lipgloss.Center, lipgloss.Center).
		Render(errMsg)
}

func (p *schedulePage) setSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v

	if p.loaded {
		p.matchList.SetSize(p.width, p.contentHeight())
	}

	p.help.Width = p.width
}

func (p *schedulePage) shouldFetchNextPage() bool {
	return p.onLastItem() &&
		p.paginationState.hasNextPage() &&
		!p.paginationState.loadingNextPage &&
		!p.matchList.IsFiltered()
}

func (p *schedulePage) shouldFetchPreviousPage() bool {
	return p.onFirstItem() &&
		p.paginationState.hasPrevPage() &&
		!p.paginationState.loadingPrevPage &&
		!p.matchList.IsFiltered()
}

func (p *schedulePage) onLastItem() bool {
	return p.matchList.Index() == len(p.matchList.Items())-1
}

func (p *schedulePage) onFirstItem() bool {
	return p.matchList.Index() == 0
}

func (p *schedulePage) handleFetchedEvents(msg fetchedEventsMessage) {
	matches := filterMatchEvents(msg.events)

	switch msg.pageDirection {
	case pageDirectionInitial:
		p.loaded = true
		p.matches = matches
		p.matchList = newMatchList(matches, p.width, p.contentHeight())
		p.paginationState.prevPageToken = msg.prevPageToken
		p.paginationState.nextPageToken = msg.nextPageToken

	case pageDirectionPrev:
		p.matchList.StopSpinner()
		p.prependMatches(matches)
		p.paginationState.prevPageToken = msg.prevPageToken
		p.paginationState.loadingPrevPage = false

	case pageDirectionNext:
		p.matchList.StopSpinner()
		p.appendMatches(matches)
		p.paginationState.nextPageToken = msg.nextPageToken
		p.paginationState.loadingNextPage = false
	}
}

func (p *schedulePage) prependMatches(events []lolesports.Event) {
	p.matches = append(events, p.matches...)
	items := newMatchListItems(p.matches)
	p.matchList.SetItems(items)
	// We should keep the cursor on the previously selected index.
	p.matchList.Select(p.matchList.Index() + len(events))
}

func (p *schedulePage) appendMatches(events []lolesports.Event) {
	p.matches = append(p.matches, events...)
	items := newMatchListItems(p.matches)
	p.matchList.SetItems(items)
}

func (p *schedulePage) handleFetchError(msg fetchErrorMessage) tea.Cmd {
	var cmd tea.Cmd

	// We log the error received after a failed fetch for debugging purpose,
	// but we display a more user-friendly message to help the user.
	if !p.loaded {
		p.errMsg = errMessageFetchInitialPage
	} else {
		p.matchList.StopSpinner()

		var statusMessage string
		switch msg.pageDirection {
		case pageDirectionNext:
			statusMessage = errMessageFetchNextPage
			p.paginationState.loadingNextPage = false
		case pageDirectionPrev:
			statusMessage = errMessageFetchPrevPage
			p.paginationState.loadingPrevPage = false
		}

		cmd = p.matchList.NewStatusMessage(statusMessage)
	}

	p.logger.Error("Failed to fetch data", slog.Any("error", msg.err))

	return cmd
}

func (p *schedulePage) updateMatchListTitle() {
	selectedIndex := p.matchList.Index()
	selectedEvent := p.matches[selectedIndex]
	title := formatDateTitle(selectedEvent.StartTime)

	p.matchList.Title = title
	p.matchList.Styles.Title = p.styles.title
}

func (p *schedulePage) contentHeight() int {
	return p.height - p.helpHeight()
}

func (p *schedulePage) helpHeight() int {
	padding := p.styles.help.GetVerticalPadding()
	if p.help.ShowAll {
		return schedulePageFullHelpHeight + padding
	}
	return schedulePageShortHelpHeight + padding
}

func (p *schedulePage) toggleHelp() {
	p.help.ShowAll = !p.help.ShowAll

	// Need to resize the list of matches as the help now
	// takes up more space.
	p.matchList.SetSize(p.width, p.contentHeight())
}

type pageDirection int

const (
	pageDirectionInitial pageDirection = iota
	pageDirectionNext
	pageDirectionPrev
)

type fetchedEventsMessage struct {
	events        []lolesports.Event
	pageDirection pageDirection
	nextPageToken string
	prevPageToken string
}

type fetchErrorMessage struct {
	err           error
	pageDirection pageDirection
}

func (p *schedulePage) fetchEventsPreviousPage() tea.Cmd {
	return p.fetchEvents(pageDirectionPrev)
}

func (p *schedulePage) fetchEventsNextPage() tea.Cmd {
	return p.fetchEvents(pageDirectionNext)
}

func (p *schedulePage) fetchEvents(pageDirection pageDirection) tea.Cmd {
	return func() tea.Msg {
		var opts lolesports.GetScheduleOptions
		switch pageDirection {
		case pageDirectionNext:
			opts.PageToken = &p.paginationState.nextPageToken
		case pageDirectionPrev:
			opts.PageToken = &p.paginationState.prevPageToken
		}

		schedule, err := p.lolesportsClient.GetSchedule(context.Background(), &opts)
		if err != nil {
			return fetchErrorMessage{
				err:           err,
				pageDirection: pageDirection,
			}
		}

		var prevPageToken, nextPageToken string
		switch pageDirection {
		case pageDirectionInitial:
			prevPageToken = schedule.Pages.Older
			nextPageToken = schedule.Pages.Newer
		case pageDirectionNext:
			nextPageToken = schedule.Pages.Newer
		case pageDirectionPrev:
			prevPageToken = schedule.Pages.Older
		}

		return fetchedEventsMessage{
			events:        schedule.Events,
			pageDirection: pageDirection,
			prevPageToken: prevPageToken,
			nextPageToken: nextPageToken,
		}
	}
}

func formatDateTitle(date time.Time) string {
	switch {
	case timeutil.IsYesterday(date):
		return "Yesterday"
	case timeutil.IsToday(date):
		if date.After(time.Now()) {
			return "Later Today"
		}
		return "Earlier Today"
	case timeutil.IsTomorrow(date):
		return "Tomorrow"
	default:
		return date.Format("Monday 02 Jan")
	}
}

func filterMatchEvents(events []lolesports.Event) []lolesports.Event {
	matches := make([]lolesports.Event, 0, len(events))

	for _, event := range events {
		if event.Type != lolesports.EventTypeMatch {
			continue
		}

		matches = append(matches, event)
	}

	return matches
}
