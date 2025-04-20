package ui

import (
	"math/rand/v2"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type wukongSpinnerStyles struct {
	quote lipgloss.Style
}

func newDefaultWukongSpinnerStyle() wukongSpinnerStyles {
	return wukongSpinnerStyles{
		quote: lipgloss.NewStyle().Foreground(textPrimaryColor).Italic(true),
	}
}

type wukongSpinner struct {
	spinner.Model
	quote  string
	styles wukongSpinnerStyles
}

func newWukongSpinner() *wukongSpinner {
	return &wukongSpinner{
		Model:  spinner.New(spinner.WithSpinner(spinner.Monkey)),
		quote:  randomWukongQuote(),
		styles: newDefaultWukongSpinnerStyle(),
	}
}

func (s *wukongSpinner) Update(msg tea.Msg) (*wukongSpinner, tea.Cmd) {
	var cmd tea.Cmd
	s.Model, cmd = s.Model.Update(msg)
	return s, cmd
}

func (s *wukongSpinner) View() string {
	return s.Model.View() + " " + s.styles.quote.Render(s.quote)
}

func (s *wukongSpinner) refreshQuote() {
	s.quote = randomWukongQuote()
}

func randomWukongQuote() string {
	quotes := []string{
		"I will be the best.",
		"Every mistake is a lesson.",
		"Who questions my ability?",
		"Never settle for second.",
		"My journey's only beginning.",
		"Show me the path.",
		"You got it!",
		"My place is at the top.",
		"Adapt to all situations.",
		"Bring me a real challenge.",
		"If you don't fight, you can't make friends.",
		"Improve your skills! Then find me again.",
	}
	return quotes[rand.IntN(len(quotes))]
}
