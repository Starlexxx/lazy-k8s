package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

type StatusBar struct {
	styles  *theme.Styles
	message string
	isError bool
}

func NewStatusBar(styles *theme.Styles) *StatusBar {
	return &StatusBar{
		styles: styles,
	}
}

func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
	s.isError = false
}

func (s *StatusBar) SetError(err string) {
	s.message = err
	s.isError = true
}

func (s *StatusBar) Clear() {
	s.message = ""
	s.isError = false
}

func (s *StatusBar) View(width int) string {
	var content string

	if s.message != "" {
		if s.isError {
			content = s.styles.StatusError.Render("Error: " + s.message)
		} else {
			content = s.styles.StatusValue.Render(s.message)
		}
	} else {
		hints := []string{
			s.styles.StatusKey.Render("q") + " quit",
			s.styles.StatusKey.Render("?") + " help",
			s.styles.StatusKey.Render("tab") + " next panel",
			s.styles.StatusKey.Render("/") + " search",
			s.styles.StatusKey.Render("c") + " context",
			s.styles.StatusKey.Render("n") + " namespace",
		}

		content = lipgloss.JoinHorizontal(lipgloss.Left,
			hints[0], " │ ",
			hints[1], " │ ",
			hints[2], " │ ",
			hints[3], " │ ",
			hints[4], " │ ",
			hints[5],
		)
	}

	return s.styles.StatusBar.Width(width).Render(content)
}
