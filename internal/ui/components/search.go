package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type Search struct {
	styles *theme.Styles
	input  textinput.Model
}

func NewSearch(styles *theme.Styles) *Search {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	ti.Width = 30
	ti.Prompt = "/ "

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
	s.input.Width = width - 10
	return s.styles.Input.Width(width - 4).Render(s.input.View())
}
