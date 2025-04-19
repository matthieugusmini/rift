package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthieugusmini/go-lolesports"
	"github.com/matthieugusmini/lolesport/internal/timeutils"
)

type schedulePageStyles struct {
	doc   lipgloss.Style
	title lipgloss.Style
}

func newDefaultSchedulePageStyles() (s schedulePageStyles) {
	s.doc = lipgloss.NewStyle().Padding(1, 2)

	s.title = lipgloss.NewStyle().
		Foreground(white).
		Background(charcoal).
		Bold(true)

	return s
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
	styles        schedulePageStyles
}

func newSchedulePage(lolesportsClient LoLEsportsClient) *schedulePage {
	return &schedulePage{
		lolesportsClient: lolesportsClient,
		spinner:          spinner.New(spinner.WithSpinner(spinner.Monkey)),
		styles:           newDefaultSchedulePageStyles(),
	}
}

func (p *schedulePage) Init() tea.Cmd {
	return tea.Batch(p.spinner.Tick, p.fetchEvents(pageDirectionInitial))
}

func (p *schedulePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// The size of this page should be managed by the parent model to ensure
	// that the page remains agnostic to its parent's layout.
	case tea.WindowSizeMsg:
		if p.matches.Items() != nil {
			p.matches.SetSize(p.width, p.height)
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
	p.matches, cmd = p.matches.Update(msg)
	cmds = append(cmds, cmd)

	p.updateMatchListTitle()

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
	if !p.loaded {
		return p.viewSpinner()
	}
	if p.err != nil {
		return p.viewError()
	}
	return p.styles.doc.Render(p.matches.View())
}

func (p *schedulePage) viewSpinner() string {
	style := lipgloss.NewStyle().
		Width(p.width).
		Height(p.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(
		p.spinner.View() + " Wukong is looking for the schedule",
	)
}

func (p *schedulePage) viewError() string {
	return p.styles.doc.Render(p.err.Error())
}

func (p *schedulePage) SetSize(width, height int) {
	h, v := p.styles.doc.GetFrameSize()
	p.width, p.height = width-h, height-v
}

func (p *schedulePage) shouldFetchNextPage() bool {
	return p.onLastItem() && p.paginationState.hasNextPage() && !p.paginationState.loadingNextPage
}

func (p *schedulePage) shouldFetchPreviousPage() bool {
	return p.onFirstItem() && p.paginationState.hasPrevPage() && !p.paginationState.loadingPrevPage
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
		p.matches = newMatchList(p.events, p.width, p.height)
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
	p.matches.Select(p.matches.Index() + len(items))
}

func (p *schedulePage) appendMatches(events []lolesports.Event) {
	p.events = append(p.events, events...)
	items := newMatchListItems(events)
	p.matches.SetItems(append(p.matches.Items(), items...))
}

func (p *schedulePage) updateMatchListTitle() {
	selectedIndex := p.matches.Index()
	selectedEvent := p.events[selectedIndex]
	title := formatDateTitle(selectedEvent.StartTime)
	// TODO: Search why -3 is needed here
	p.matches.Title = lipgloss.PlaceHorizontal(p.width-3, lipgloss.Center, title)
	p.matches.Styles.Title = p.styles.title
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
		if pageDirection == pageDirectionNext {
			opts.PageToken = &p.paginationState.nextPageToken
		} else if pageDirection == pageDirectionPrev {
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
		} else {
			return "Earlier Today"
		}
	case timeutils.IsTomorrow(date):
		return "Tomorrow"
	default:
		return date.Format("Monday 02 Jan")
	}
}
