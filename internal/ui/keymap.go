package ui

import "github.com/charmbracelet/bubbles/key"

// These key bindings are available across all the application.
type baseKeyMap struct {
	NextPage      key.Binding
	PrevPage      key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
	Quit          key.Binding
}

func newBaseKeyMap() baseKeyMap {
	return baseKeyMap{
		PrevPage: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next page"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}
