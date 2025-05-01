package ui

type BracketTemplate struct {
	Rounds []Round
}

type Round struct {
	Title   string
	Links   []Link
	Matches []Match
}

type LinkType int

const (
	LinkTypeHorizontal LinkType = iota
	LinkTypeZDown
	LinkTypeZUp
	LinkTypeLDown
	LinkTypeLUp
	LinkTypeReseed
)

type Link struct {
	Type   LinkType
	Height int
	Above  int
}

type DisplayType int

const (
	DisplayTypeUnknown DisplayType = iota
	DisplayTypeMatch
	DisplayTypeHorizontalLine
)

type Match struct {
	DisplayType DisplayType
	Height      int
	Above       int
	Label       string
}

var BracketCrossGroupBattles = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
	},
}

var Bracket4DEGroup = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 2},
			},
		},
		{
			Title: "ROUND 3",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 15},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 11},
			},
		},
	},
}

var Bracket6DE = BracketTemplate{
	Rounds: []Round{
		{
			Title: "Round 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "Round 2",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 2},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "Round 3",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 7},
				{Type: LinkTypeZDown, Height: 1, Above: 8},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 5},
				{DisplayType: DisplayTypeMatch, Above: 7},
			},
		},
		{
			Title: "Round 4",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 7},
				{Type: LinkTypeHorizontal, Above: 11},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 5},
				{DisplayType: DisplayTypeMatch, Above: 7},
			},
		},
		{
			Title: "FINALS",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 7},
				{Type: LinkTypeZUp, Height: 8, Above: 2},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 4},
			},
		},
	},
}

var Bracket6KOTHLCK = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 6},
			},
		},
		{
			Title: "ROUND 3",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 10},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 6},
			},
		},
		{
			Title: "ROUND 4",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 10},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 6},
			},
		},
		{
			Title: "ROUND 5",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 10},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 6},
			},
		},
	},
}

var Bracket4SE = BracketTemplate{
	Rounds: []Round{
		{
			Title: "SEMIFINALS",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "FINALS",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
			},
		},
	},
}

var Bracket8SE = BracketTemplate{
	Rounds: []Round{
		{
			Title: "QUARTERFINALS",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "SEMIFINALS",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
				{Type: LinkTypeZDown, Height: 1, Above: 6},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 6},
			},
		},
		{
			Title: "FINALS",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 4, Above: 7},
				{Type: LinkTypeZUp, Height: 4, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 9},
			},
		},
	},
}

var Bracket8DETop3 = BracketTemplate{
	Rounds: []Round{
		{
			Title: "Round 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "Round 2",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
				{Type: LinkTypeZDown, Height: 1, Above: 6},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 6},
				{DisplayType: DisplayTypeMatch, Above: 2},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "Round 3",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 4, Above: 7},
				{Type: LinkTypeZUp, Height: 4, Above: 1},
				{Type: LinkTypeHorizontal, Height: 1, Above: 8},
				{Type: LinkTypeHorizontal, Height: 1, Above: 6},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 9},
				{DisplayType: DisplayTypeMatch, Above: 8},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "LOWER BRACKET - ROUND 3",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 27},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 26},
			},
		},
	},
}

var Bracket8DE4Qual = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch, Label: "Losers' Bracket"},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 2, Above: 6},
				{Type: LinkTypeZUp, Height: 2, Above: 2},
				{Type: LinkTypeZDown, Height: 2, Above: 6},
				{Type: LinkTypeZUp, Height: 2, Above: 2},
				{Type: LinkTypeLDown, Height: 1, Above: 3},
				{Type: LinkTypeHorizontal, Height: 1, Above: 1},
				{Type: LinkTypeLDown, Height: 1, Above: 3},
				{Type: LinkTypeHorizontal, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 6},
				{DisplayType: DisplayTypeMatch, Above: 2},
				{DisplayType: DisplayTypeMatch},
			},
		},
	},
}

var Bracket6DELCK = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{
				{Type: LinkTypeReseed},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 3",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 2},
			},
		},
		{
			Title: "ROUND 4",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 7},
				{Type: LinkTypeHorizontal, Height: 1, Above: 8},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 5},
				{DisplayType: DisplayTypeMatch, Above: 4},
			},
		},
		{
			Title: "FINALS",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 2, Above: 7},
				{Type: LinkTypeZUp, Height: 2, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 7},
			},
		},
	},
}

var Bracket4DE4LSeeds = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 18, Label: "Losers' Bracket"},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{
				{Type: LinkTypeLDown, Height: 1, Above: 20},
				{Type: LinkTypeHorizontal, Height: 1, Above: 2},
				{Type: LinkTypeLDown, Height: 1, Above: 4},
				{Type: LinkTypeHorizontal, Height: 1, Above: 2},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeHorizontalLine, Above: 5},
				{DisplayType: DisplayTypeHorizontalLine, Above: 4},
			},
		},
		{
			Title: "ROUND 3",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 7},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
				{Type: LinkTypeHorizontal, Height: 1, Above: 9},
				{Type: LinkTypeHorizontal, Height: 1, Above: 6},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 6},
				{DisplayType: DisplayTypeMatch, Above: 6},
				{DisplayType: DisplayTypeMatch, Above: 0},
			},
		},
		{
			Title: "ROUND 4",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 10},
				{Type: LinkTypeZDown, Height: 1, Above: 12},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 8},
				{DisplayType: DisplayTypeMatch, Above: 11},
			},
		},
		{
			Title: "ROUND 5",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 10},
				{Type: LinkTypeHorizontal, Height: 1, Above: 15},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 8},
				{DisplayType: DisplayTypeMatch, Above: 11},
			},
		},
		{
			Title: "GRAND FINAL",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 10},
				{Type: LinkTypeZUp, Height: 12, Above: 2},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 7},
			},
		},
	},
}

var BracketSE3Qual = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{
				{Type: LinkTypeHorizontal, Above: 4},
				{Type: LinkTypeHorizontal, Above: 6},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 3",
			Links: []Link{},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 12},
			},
		},
	},
}

var Bracket8DE = BracketTemplate{
	Rounds: []Round{
		{
			Title: "ROUND 1",
			Matches: []Match{
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 2",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 1, Above: 4},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
				{Type: LinkTypeZDown, Height: 1, Above: 6},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 3},
				{DisplayType: DisplayTypeMatch, Above: 6},
				{DisplayType: DisplayTypeMatch, Above: 2},
				{DisplayType: DisplayTypeMatch},
			},
		},
		{
			Title: "ROUND 3",
			Links: []Link{
				{Type: LinkTypeZDown, Height: 4, Above: 7},
				{Type: LinkTypeZUp, Height: 4, Above: 1},
				{Type: LinkTypeHorizontal, Height: 1, Above: 8},
				{Type: LinkTypeHorizontal, Height: 1, Above: 6},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 9},
				{DisplayType: DisplayTypeMatch, Above: 8},
				{DisplayType: DisplayTypeMatch, Above: 0},
			},
		},
		{
			Title: "ROUND 4",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 13},
				{Type: LinkTypeZDown, Height: 1, Above: 14},
				{Type: LinkTypeZUp, Height: 1, Above: 1},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 11},
				{DisplayType: DisplayTypeMatch, Above: 13},
			},
		},
		{
			Title: "ROUND 5",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 13},
				{Type: LinkTypeHorizontal, Height: 1, Above: 17},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeHorizontalLine, Above: 11},
				{DisplayType: DisplayTypeMatch, Above: 13},
			},
		},
		{
			Title: "GRAND FINAL",
			Links: []Link{
				{Type: LinkTypeHorizontal, Height: 1, Above: 13},
				{Type: LinkTypeZUp, Height: 14, Above: 2},
			},
			Matches: []Match{
				{DisplayType: DisplayTypeMatch, Above: 10},
			},
		},
	},
}
