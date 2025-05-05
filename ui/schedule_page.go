package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"

	"github.com/matthieugusmini/lolesport/timeutils"
)

const (
	schedulePageShortHelpHeight = 1
	schedulePageFullHelpHeight  = 7
)

type schedulePageStyles struct {
	doc     lipgloss.Style
	title   lipgloss.Style
	spinner lipgloss.Style
	help    lipgloss.Style
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
	lolesportsClient LoLEsportsClient

	events          []lolesports.Event
	matches         list.Model
	paginationState paginationState

	err error

	spinner spinner.Model
	loaded  bool

	width, height int

	styles schedulePageStyles
	keyMap schedulePageKeyMap
	help   help.Model
}

func newSchedulePage(lolesportsClient LoLEsportsClient) *schedulePage {
	styles := newDefaultSchedulePageStyles()

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(styles.spinner),
	)

	return &schedulePage{
		lolesportsClient: lolesportsClient,
		spinner:          sp,
		styles:           styles,
		keyMap:           newDefaultSchedulePageKeyMap(),
		help:             help.New(),
	}
}

func (p *schedulePage) Init() tea.Cmd {
	return tea.Batch(p.spinner.Tick, p.fetchEvents(pageDirectionInitial))
}

func (p *schedulePage) toggleHelp() {
	p.help.ShowAll = !p.help.ShowAll
	p.matches.SetSize(p.width, p.contentViewHeight())
}

func (p *schedulePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keyMap.ShowFullHelp),
			key.Matches(msg, p.keyMap.CloseFullHelp):
			p.toggleHelp()
		}

	// The size of this page should be managed by the parent model to ensure
	// that the page remains agnostic to its parent's layout.
	case tea.WindowSizeMsg:
		if p.matches.Items() != nil {
			p.matches.SetSize(p.width, p.contentViewHeight())
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
		p.err = msg.err
	}

	if !p.loaded {
		return p, tea.Batch(cmds...)
	}

	var cmd tea.Cmd

	p.updateMatchListTitle()

	p.matches, cmd = p.matches.Update(msg)
	cmds = append(cmds, cmd)

	if p.shouldFetchNextPage() {
		p.paginationState.loadingNextPage = true
		cmds = append(cmds, p.matches.StartSpinner(), p.fetchEventsNextPage())
	}
	if p.shouldFetchPreviousPage() {
		p.paginationState.loadingPrevPage = true
		cmds = append(cmds, p.matches.StartSpinner(), p.fetchEventsPreviousPage())
	}

	return p, tea.Batch(cmds...)
}

func (p *schedulePage) View() string {
	if p.err != nil {
		return p.viewError()
	}

	var sections []string

	if !p.loaded {
		sections = append(sections, p.viewSpinner())
	} else {
		sections = append(sections, p.matches.View())
	}

	sections = append(sections, p.viewHelp())

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return p.styles.doc.Render(view)
}

func (p *schedulePage) viewHelp() string {
	return p.styles.help.Render(p.help.View(p))
}

func (p *schedulePage) viewSpinner() string {
	return lipgloss.NewStyle().
		Width(p.width).
		Height(p.contentViewHeight()).
		Align(lipgloss.Center, lipgloss.Center).
		Render(p.spinner.View())
}

func (p *schedulePage) viewError() string {
	return p.styles.doc.Render(p.err.Error())
}

func (p *schedulePage) setSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v

	p.help.Width = p.width
}

func (p *schedulePage) shouldFetchNextPage() bool {
	return p.onLastItem() &&
		p.paginationState.hasNextPage() &&
		!p.paginationState.loadingNextPage &&
		!p.matches.IsFiltered()
}

func (p *schedulePage) shouldFetchPreviousPage() bool {
	return p.onFirstItem() &&
		p.paginationState.hasPrevPage() &&
		!p.paginationState.loadingPrevPage &&
		!p.matches.IsFiltered()
}

func (p *schedulePage) onLastItem() bool {
	return p.matches.Index() == len(p.matches.Items())-1
}

func (p *schedulePage) onFirstItem() bool {
	return p.matches.Index() == 0
}

func (p *schedulePage) handleFetchedEvents(msg fetchedEventsMessage) {
	switch msg.pageDirection {
	case pageDirectionInitial:
		p.loaded = true
		p.events = msg.events
		p.matches = newMatchList(p.events, p.width, p.contentViewHeight())
		p.paginationState.prevPageToken = msg.prevPageToken
		p.paginationState.nextPageToken = msg.nextPageToken

	case pageDirectionPrev:
		p.matches.StopSpinner()
		p.prependMatches(msg.events)
		p.paginationState.prevPageToken = msg.prevPageToken
		p.paginationState.loadingPrevPage = false

	case pageDirectionNext:
		p.matches.StopSpinner()
		p.appendMatches(msg.events)
		p.paginationState.nextPageToken = msg.nextPageToken
		p.paginationState.loadingNextPage = false
	}
}

func (p *schedulePage) prependMatches(events []lolesports.Event) {
	p.events = append(events, p.events...)
	items := newMatchListItems(p.events)
	p.matches.SetItems(items)
	// We should keep the cursor on the previously selected index.
	p.matches.Select(p.matches.Index() + len(events))
}

func (p *schedulePage) appendMatches(events []lolesports.Event) {
	p.events = append(p.events, events...)
	items := newMatchListItems(p.events)
	p.matches.SetItems(items)
}

func (p *schedulePage) updateMatchListTitle() {
	selectedIndex := p.matches.Index()
	selectedEvent := p.events[selectedIndex]
	title := formatDateTitle(selectedEvent.StartTime)

	p.matches.Title = title
	p.matches.Styles.Title = p.styles.title
}

func (p *schedulePage) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.CursorUp,
		p.keyMap.CursorDown,
		p.keyMap.RevealSpoiler,
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

func (p *schedulePage) helpHeight() int {
	padding := p.styles.help.GetVerticalPadding()
	if p.help.ShowAll {
		return schedulePageFullHelpHeight + padding
	}
	return schedulePageShortHelpHeight + padding
}

func (p *schedulePage) contentViewHeight() int {
	return p.height - p.helpHeight()
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
	err error
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
			return fetchErrorMessage{err}
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
	case timeutils.IsYesterday(date):
		return "Yesterday"
	case timeutils.IsToday(date):
		if date.After(time.Now()) {
			return "Later Today"
		}
		return "Earlier Today"
	case timeutils.IsTomorrow(date):
		return "Tomorrow"
	default:
		return date.Format("Monday 02 Jan")
	}
}
