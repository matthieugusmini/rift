package lolesportsgraphql

type SectionType string

const (
	SectionTypeBracket SectionType = "bracket"
	SectionTypeGroup   SectionType = "group"
)

type DestinationType string

const (
	DestinationTypeMatch    DestinationType = "match"
	DestinationTypeStage    DestinationType = "stage"
	DestinationTypeStanding DestinationType = "standing"
)

type Stage struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Slug         string        `json:"slug"`
	DisplayState string        `json:"displayState"`
	Destinations []Destination `json:"destinations"`
	Sections     []Section     `json:"sections"`
}

type Section struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Slug      string      `json:"slug"`
	Type      SectionType `json:"type"`
	Columns   []Column    `json:"columns"`
	Rankings  []Ranking   `json:"rankings"`
	Standings []Standing  `json:"standings"`
}

type Column struct {
	Cells []Cell `json:"cells"`
}

type Cell struct {
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	Matches []Match `json:"matches"`
}

type Match struct {
	ID           string        `json:"id"`
	StructuralID string        `json:"structuralId"`
	Description  string        `json:"description"`
	State        string        `json:"state"`
	Type         string        `json:"type"`
	Flags        []string      `json:"flags"`
	Teams        []MatchTeam   `json:"matchTeams"`
	Destinations *Destinations `json:"destinations"`
}

type MatchTeam struct {
	TeamID string      `json:"teamId"`
	Name   string      `json:"name"`
	Code   string      `json:"code"`
	Slug   string      `json:"slug"`
	Image  string      `json:"image"`
	Result *TeamResult `json:"result"`
}

type Destinations struct {
	Win  Destination `json:"win"`
	Loss Destination `json:"loss"`
}

type Destination struct {
	Type         DestinationType `json:"type"`
	StructuralID *string         `json:"structuralId"`
	Slot         int             `json:"slot"`
}

type Ranking struct {
	Ordinal int    `json:"ordinal"`
	Teams   []Team `json:"teams"`
}

type Standing struct {
	Index     int  `json:"index"`
	Placement int  `json:"placement"`
	Team      Team `json:"team"`
}

type Team struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Code   string      `json:"code"`
	Slug   string      `json:"slug"`
	Image  string      `json:"image"`
	Record *TeamRecord `json:"record"`
	Result *TeamResult `json:"result"`
}

type TeamRecord struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Ties   int `json:"ties"`
}

type TeamResult struct {
	Outcome  *string `json:"outcome"`
	GameWins int     `json:"gameWins"`
}
