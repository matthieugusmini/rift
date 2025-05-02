// Package rift contains all the domain objects for the Rift app.
package rift

// BracketTemplate represents a template to display a bracket layout.
//
// This template structure is inspired by the LoL Fandom bracket templates.
// See: https://lol.fandom.com/wiki/Template:Bracket
type BracketTemplate struct {
	Rounds []Round `json:"rounds,omitempty"`
}

// Round represents a single round in the bracket.
type Round struct {
	// Title of the round.
	Title string `json:"title,omitempty"`

	// Links to the previous round.
	// These should be displayed before the matches, from top to bottom.
	Links []Link `json:"links,omitempty"`

	// Matches in the round, listed from top to bottom.
	Matches []Match `json:"matches,omitempty"`
}

// LinkType defines the type of link in the bracket.
type LinkType string

const (
	// LinkTypeHorizontal represents a horizontal line linking two rounds.
	LinkTypeHorizontal LinkType = "horizontal"

	// LinkTypeZDown represents a Z-shaped line going downward to link two rounds.
	LinkTypeZDown LinkType = "z-down"

	// LinkTypeZUp represents a Z-shaped line going upward to link two rounds.
	LinkTypeZUp LinkType = "z-up"

	// LinkTypeLDown represents an L-shaped line going downward to link two rounds.
	LinkTypeLDown LinkType = "l-down"

	// LinkTypeLUp represents an L-shaped line going upward to link two rounds.
	LinkTypeLUp LinkType = "l-up"

	// LinkTypeReseed represents a reseed link.
	LinkTypeReseed LinkType = "reseed"
)

// Link represents a link between two rounds of the bracket.
type Link struct {
	// Type of the link.
	Type LinkType `json:"type,omitempty"`

	// Height of the link.
	//
	// Applies only to the types: z-down, z-up, l-down, l-up.
	// It represents the number of vertical lines to add to the link.
	Height int `json:"height,omitempty"`

	// Above represents the number of newlines above the link.
	Above int `json:"above,omitempty"`
}

// DisplayType defines the type of display in the bracket.
type DisplayType string

const (
	// DisplayTypeMatch represents a match.
	DisplayTypeMatch DisplayType = "match"

	// DisplayTypeHorizontalLine represents a horizontal line (no match).
	DisplayTypeHorizontalLine DisplayType = "horizontal-line"
)

// Match represents a match in the bracket.
type Match struct {
	// DisplayType is the type of display for the match.
	// It offers more flexibility in the display, as sometimes you may not want
	// to display a match in certain rounds.
	DisplayType DisplayType `json:"displayType,omitempty"`

	// Above represents the number of newlines above the match.
	Above int `json:"above,omitempty"`

	// Label for the match (e.g., Lower bracket, Qualifier).
	Label string `json:"label,omitempty"`
}
