package rift

type BracketTemplate struct {
	Rounds []Round `json:"rounds,omitempty"`
}

type Round struct {
	Title   string  `json:"title,omitempty"`
	Links   []Link  `json:"links,omitempty"`
	Matches []Match `json:"matches,omitempty"`
}

type LinkType string

const (
	LinkTypeHorizontal LinkType = "horizontal"
	LinkTypeZDown      LinkType = "z-down"
	LinkTypeZUp        LinkType = "z-up"
	LinkTypeLDown      LinkType = "l-down"
	LinkTypeLUp        LinkType = "l-up"
	LinkTypeReseed     LinkType = "reseed"
)

type Link struct {
	Type   LinkType `json:"type,omitempty"`
	Height int      `json:"height,omitempty"`
	Above  int      `json:"above,omitempty"`
}

type DisplayType string

const (
	DisplayTypeUnknown        DisplayType = "unknown"
	DisplayTypeMatch          DisplayType = "match"
	DisplayTypeHorizontalLine DisplayType = "horizontal-line"
)

type Match struct {
	DisplayType DisplayType `json:"displayType,omitempty"`
	Height      int         `json:"height,omitempty"`
	Above       int         `json:"above,omitempty"`
	Label       string      `json:"label,omitempty"`
}
