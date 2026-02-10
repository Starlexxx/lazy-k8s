package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

type Help struct {
	styles *theme.Styles
	keys   *theme.KeyMap
}

func NewHelp(styles *theme.Styles, keys *theme.KeyMap) *Help {
	return &Help{
		styles: styles,
		keys:   keys,
	}
}

func (h *Help) View(width, height int) string {
	var b strings.Builder

	b.WriteString(h.styles.ModalTitle.Render("lazy-k8s - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	sections := []struct {
		title    string
		bindings []struct{ key, desc string }
	}{
		{
			title: "Navigation",
			bindings: []struct{ key, desc string }{
				{"↑/k", "Move up"},
				{"↓/j", "Move down"},
				{"g", "Go to top"},
				{"G", "Go to bottom"},
				{"Tab", "Next panel"},
				{"Shift+Tab", "Previous panel"},
				{"1-9", "Jump to panel"},
				{"Enter", "Select/expand"},
				{"Esc", "Back/cancel"},
			},
		},
		{
			title: "General",
			bindings: []struct{ key, desc string }{
				{"?", "Show help"},
				{"q/Ctrl+c", "Quit"},
				{"/", "Search/filter"},
				{"Ctrl+r", "Refresh"},
				{"K", "Switch context"},
				{"n", "Switch namespace"},
				{"A", "Toggle all namespaces"},
			},
		},
		{
			title: "Resource Actions",
			bindings: []struct{ key, desc string }{
				{"d", "Describe resource"},
				{"y", "View YAML"},
				{"e", "Edit resource"},
				{"D", "Delete (with confirm)"},
				{"c", "Copy name"},
				{"Ctrl+y", "Copy YAML"},
			},
		},
		{
			title: "Pod Actions",
			bindings: []struct{ key, desc string }{
				{"l", "View logs"},
				{"f", "Toggle follow logs"},
				{"x", "Exec into container"},
				{"p", "Port forward"},
			},
		},
		{
			title: "Deployment Actions",
			bindings: []struct{ key, desc string }{
				{"s", "Scale"},
				{"r", "Restart (rollout)"},
				{"R", "Rollback"},
			},
		},
	}

	keyStyle := h.styles.StatusKey
	descStyle := h.styles.StatusValue
	sectionStyle := lipgloss.NewStyle().
		MarginBottom(1).
		Width(35)

	var leftCol, rightCol strings.Builder

	for i, section := range sections {
		var col *strings.Builder
		if i%2 == 0 {
			col = &leftCol
		} else {
			col = &rightCol
		}

		col.WriteString(h.styles.DetailTitle.Render(section.title))
		col.WriteString("\n")

		for _, binding := range section.bindings {
			key := keyStyle.Width(12).Render(binding.key)
			desc := descStyle.Render(binding.desc)
			col.WriteString("  " + key + desc + "\n")
		}

		col.WriteString("\n")
	}

	left := sectionStyle.Render(leftCol.String())
	right := sectionStyle.Render(rightCol.String())

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right))

	b.WriteString("\n")
	b.WriteString(h.styles.Muted.Render("Press ? or Esc to close"))

	content := h.styles.Modal.
		Width(width - 10).
		MaxHeight(height - 4).
		Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
