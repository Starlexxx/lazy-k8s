package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type Search struct {
	styles *theme.Styles
	input  textinput.Model
}

func NewSearch(styles *theme.Styles) *Search {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &Search{
		styles: styles,
		input:  ti,
	}
}

func (s *Search) Focus() {
	s.input.Focus()
}

func (s *Search) Blur() {
	s.input.Blur()
}

func (s *Search) Clear() {
	s.input.SetValue("")
}

func (s *Search) Value() string {
	return s.input.Value()
}

func (s *Search) SetValue(v string) {
	s.input.SetValue(v)
}

func (s *Search) Update(msg tea.Msg) (*Search, tea.Cmd) {
	var cmd tea.Cmd

	s.input, cmd = s.input.Update(msg)

	return s, cmd
}

func (s *Search) View(width int) string {
	s.input.Width = width - 6

	return lipgloss.NewStyle().
		Foreground(s.styles.Text).
		Padding(0, 1).
		Width(width - 2).
		Render(s.input.View())
}
