package lolesport

import "time"

type Schedule struct {
	Updated time.Time `json:"updated"`
	Pages   Pages     `json:"pages"`
	Events  []Event   `json:"events"`
}

type Pages struct {
	Older string `json:"older"`
	Newer string `json:"newer"`
}

type EventType string

const (
	EventTypeMatch = "match"
	EventTypeShow  = "show"
)

type EventState string

const (
	EventStateUnstarted  = "unstarted"
	EventStateInProgress = "inProgress"
	EventStateCompleted  = "completed"
)

type Event struct {
	StartTime time.Time  `json:"startTime"`
	BlockName string     `json:"blockName"`
	Match     Match      `json:"match"`
	State     EventState `json:"state"`
	Type      string     `json:"type"`
	League    League     `json:"league"`
}

type Match struct {
	ID               string   `json:"id"`
	PreviousMatchIDs []string `json:"previousMatchIds"`
	Flags            []string `json:"flags"`
	Teams            []Team   `json:"teams"`
	Strategy         Strategy `json:"strategy"`
}

type Team struct {
	ID     string  `json:"id,omitempty"`
	Slug   string  `json:"slug,omitempty"`
	Name   string  `json:"name"`
	Code   string  `json:"code"`
	Image  string  `json:"image"`
	Result *Result `json:"result,omitempty"`
	Record *Record `json:"record,omitempty"`
}

type Result struct {
	Outcome  *string `json:"outcome"`
	GameWins int     `json:"gameWins"`
}

type Record struct {
	Losses int `json:"losses"`
	Wins   int `json:"wins"`
}

type Strategy struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
}

type MatchStrategyType string

const (
	MatchStrategyTypeBestOf = "bestOf"
)

type Standings struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	ID       string    `json:"id"`
	Name     string    `json:"name,omitempty"`
	Type     string    `json:"type,omitempty"`
	Slug     string    `json:"slug,omitempty"`
	Sections []Section `json:"sections,omitempty"`
}

type Section struct {
	Name     string    `json:"name"`
	Matches  []Match   `json:"matches"`
	Rankings []Ranking `json:"rankings"`
}

type Ranking struct {
	Ordinal int    `json:"ordinal"`
	Teams   []Team `json:"teams"`
}

type League struct {
	ID              string          `json:"id"`
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Region          string          `json:"region"`
	Image           string          `json:"image"`
	Priority        int             `json:"priority"`
	DisplayPriority DisplayPriority `json:"displayPriority"`
	Tournaments     []*Tournament   `json:"tournaments"`
}

type DisplayPriority struct {
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type Tournament struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	StartDate Date   `json:"startDate"`
	EndDate   Date   `json:"endDate"`
}
