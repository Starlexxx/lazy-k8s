package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type Confirm struct {
	styles    *theme.Styles
	title     string
	message   string
	action    func() tea.Cmd
	confirmed bool
	done      bool
	selected  int // 0 = No, 1 = Yes
}

func NewConfirm(styles *theme.Styles) *Confirm {
	return &Confirm{
		styles: styles,
	}
}

func (c *Confirm) Show(title, message string, action func() tea.Cmd) {
	c.title = title
	c.message = message
	c.action = action
	c.confirmed = false
	c.done = false
	c.selected = 0
}

func (c *Confirm) Update(msg tea.Msg) (*Confirm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			c.selected = 0
		case "right", "l":
			c.selected = 1
		case "tab":
			c.selected = (c.selected + 1) % 2
		case "enter":
			c.confirmed = c.selected == 1
			c.done = true
		case "esc", "n":
			c.confirmed = false
			c.done = true
		case "y":
			c.confirmed = true
			c.done = true
		}
	}

	return c, nil
}

func (c *Confirm) View() string {
	var b strings.Builder

	b.WriteString(c.styles.ModalTitle.Render(c.title))
	b.WriteString("\n\n")
	b.WriteString(c.message)
	b.WriteString("\n\n")

	// Buttons
	noStyle := lipgloss.NewStyle().
		Padding(0, 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c.styles.Border)

	yesStyle := lipgloss.NewStyle().
		Padding(0, 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c.styles.Border)

	if c.selected == 0 {
		noStyle = noStyle.
			BorderForeground(c.styles.Primary).
			Foreground(c.styles.Primary).
			Bold(true)
	} else {
		yesStyle = yesStyle.
			BorderForeground(c.styles.Error).
			Foreground(c.styles.Error).
			Bold(true)
	}

	noBtn := noStyle.Render("No")
	yesBtn := yesStyle.Render("Yes")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, noBtn, "  ", yesBtn)
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(c.styles.Muted.Render("←/→ to select • enter to confirm • esc to cancel"))

	return c.styles.Modal.Width(50).Render(b.String())
}

func (c *Confirm) Done() bool {
	return c.done
}

func (c *Confirm) Confirmed() bool {
	return c.confirmed
}

func (c *Confirm) Action() tea.Cmd {
	if c.action != nil {
		return c.action()
	}

	return nil
}

func (c *Confirm) Reset() {
	c.title = ""
	c.message = ""
	c.action = nil
	c.confirmed = false
	c.done = false
	c.selected = 0
}
