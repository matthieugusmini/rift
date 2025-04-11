package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/list"
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

type schedulePage struct {
	lolesportsClient LoLEsportClient

	paginationState paginationState
	matches         list.Model

	errorMessage string

	width, height int
	styles        schedulePageStyles
}

func newSchedulePage(lolesportsClient LoLEsportClient) *schedulePage {
	return &schedulePage{
		lolesportsClient: lolesportsClient,
		styles:           newDefaultSchedulePageStyles(),
	}
}

func (m *schedulePage) Init() tea.Cmd {
	return m.getEvents(pageDirectionNone)
}

func (m *schedulePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// The size of this page should be managed by the parent model to ensure
	// that the page remains agnostic to its parent's layout.
	case tea.WindowSizeMsg:
		if m.matches.Items() == nil {
			return m, nil
		}
		m.matches.SetSize(m.width, m.height)

	case fetchedEventsMessage:
		m.handleFetchedEvents(msg)

	case fetchErrorMessage:
		m.errorMessage = msg.err.Error()
		return m, nil
	}

	if m.matches.Items() == nil {
		return m, nil
	}

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	m.matches, cmd = m.matches.Update(msg)
	cmds = append(cmds, cmd)

	m.updateMatchListTitle()

	if m.shouldFetchNextPage() {
		m.paginationState.loadingNextPage = true
		cmds = append(cmds, m.getEventsNextPage())
	}
	if m.shouldFetchPreviousPage() {
		m.paginationState.loadingPrevPage = true
		cmds = append(cmds, m.getEventsPreviousPage())
	}

	return m, tea.Batch(cmds...)
}

func (m *schedulePage) View() string {
	if m.errorMessage != "" {
		return m.errorView()
	}
	return m.styles.doc.Render(m.matches.View())
}

func (m *schedulePage) errorView() string {
	return m.styles.doc.Render(m.errorMessage)
}

func (m *schedulePage) SetSize(width, height int) {
	h, v := m.styles.doc.GetFrameSize()
	m.width, m.height = width-h, height-v
}

func (m *schedulePage) shouldFetchNextPage() bool {
	return m.onLastItem() &&
		m.paginationState.nextPageToken != "" &&
		!m.paginationState.loadingNextPage
}

func (m *schedulePage) shouldFetchPreviousPage() bool {
	return m.onFirstItem() &&
		m.paginationState.prevPageToken != "" &&
		!m.paginationState.loadingPrevPage
}

func (m *schedulePage) onLastItem() bool {
	return m.matches.Index() == len(m.matches.Items())-1
}

func (m *schedulePage) onFirstItem() bool {
	return m.matches.Index() == 0
}

func (m *schedulePage) handleFetchedEvents(msg fetchedEventsMessage) {
	switch msg.pageDirection {
	case pageDirectionNone:
		m.matches = newMatchList(msg.events, m.width, m.height)
		m.paginationState.prevPageToken = msg.prevPageToken
		m.paginationState.nextPageToken = msg.nextPageToken
	case pageDirectionPrev:
		m.prependMatches(msg.events)
		m.paginationState.prevPageToken = msg.prevPageToken
		m.paginationState.loadingPrevPage = false
	case pageDirectionNext:
		m.appendMatches(msg.events)
		m.paginationState.nextPageToken = msg.nextPageToken
		m.paginationState.loadingNextPage = false
	}
}

func (m *schedulePage) prependMatches(events []lolesports.Event) {
	items := newMatchListItems(events)
	m.matches.SetItems(append(items, m.matches.Items()...))
	// We should keep the cursor on the previously selected index.
	m.matches.Select(m.matches.Index() + len(items))
}

func (m *schedulePage) appendMatches(events []lolesports.Event) {
	items := newMatchListItems(events)
	m.matches.SetItems(append(m.matches.Items(), items...))
}

func (m *schedulePage) updateMatchListTitle() {
	if item, ok := m.matches.SelectedItem().(matchItem); ok {
		title := formatDateTitle(item.startTime)
		// TODO: Search why -3 is needed here
		m.matches.Title = lipgloss.PlaceHorizontal(m.width-3, lipgloss.Center, title)
		m.matches.Styles.Title = m.styles.title
	}
}

type pageDirection int

const (
	pageDirectionNone pageDirection = iota
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

func (m *schedulePage) getEventsPreviousPage() tea.Cmd {
	return m.getEvents(pageDirectionPrev)
}

func (m *schedulePage) getEventsNextPage() tea.Cmd {
	return m.getEvents(pageDirectionNext)
}

func (m *schedulePage) getEvents(pageDirection pageDirection) tea.Cmd {
	return func() tea.Msg {
		var opts lolesports.GetScheduleOptions
		if pageDirection == pageDirectionNext {
			opts.PageToken = &m.paginationState.nextPageToken
		} else if pageDirection == pageDirectionPrev {
			opts.PageToken = &m.paginationState.prevPageToken
		}

		schedule, err := m.lolesportsClient.GetSchedule(context.Background(), opts)
		if err != nil {
			return fetchErrorMessage{err}
		}

		var prevPageToken, nextPageToken string
		switch pageDirection {
		case pageDirectionNone:
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
