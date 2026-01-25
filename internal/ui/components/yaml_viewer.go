package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type YamlViewer struct {
	styles  *theme.Styles
	content string
	lines   []string
	offset  int
	width   int
	height  int
}

func NewYamlViewer(styles *theme.Styles) *YamlViewer {
	return &YamlViewer{
		styles: styles,
	}
}

func (y *YamlViewer) SetContent(content string) {
	y.content = content
	y.lines = strings.Split(content, "\n")
	y.offset = 0
}

func (y *YamlViewer) Update(msg tea.Msg) (*YamlViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if y.offset > 0 {
				y.offset--
			}
		case "down", "j":
			maxOffset := len(y.lines) - y.height + 6
			if maxOffset < 0 {
				maxOffset = 0
			}

			if y.offset < maxOffset {
				y.offset++
			}
		case "g":
			y.offset = 0
		case "G":
			maxOffset := len(y.lines) - y.height + 6
			if maxOffset < 0 {
				maxOffset = 0
			}

			y.offset = maxOffset
		case "pgup", "ctrl+u":
			y.offset -= y.height / 2
			if y.offset < 0 {
				y.offset = 0
			}
		case "pgdown", "ctrl+d":
			maxOffset := len(y.lines) - y.height + 6
			if maxOffset < 0 {
				maxOffset = 0
			}

			y.offset += y.height / 2
			if y.offset > maxOffset {
				y.offset = maxOffset
			}
		}
	}

	return y, nil
}

func (y *YamlViewer) View(width, height int) string {
	y.width = width
	y.height = height

	var b strings.Builder

	// Title bar
	title := y.styles.ModalTitle.Render("YAML Viewer")
	hint := y.styles.Muted.Render("↑/↓ scroll • g/G top/bottom • esc close")
	titleBar := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", hint)
	b.WriteString(titleBar)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	// Content area
	visibleHeight := height - 6
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	endIdx := y.offset + visibleHeight
	if endIdx > len(y.lines) {
		endIdx = len(y.lines)
	}

	// Syntax highlighting for YAML
	for i := y.offset; i < endIdx; i++ {
		line := y.lines[i]
		highlighted := y.highlightLine(line, width-6)
		b.WriteString(highlighted)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(y.lines) > visibleHeight {
		scrollPos := float64(y.offset) / float64(len(y.lines)-visibleHeight)
		indicator := strings.Repeat("─", int(float64(width-10)*scrollPos))
		indicator += "█"
		indicator += strings.Repeat("─", width-10-len(indicator))

		b.WriteString("\n")
		b.WriteString(y.styles.Muted.Render(indicator))
	}

	return y.styles.Modal.
		Width(width - 4).
		Height(height - 2).
		Render(b.String())
}

func (y *YamlViewer) highlightLine(line string, maxWidth int) string {
	if len(line) > maxWidth {
		line = line[:maxWidth-3] + "..."
	}

	// Simple YAML syntax highlighting
	keyStyle := lipgloss.NewStyle().Foreground(y.styles.Primary)
	valueStyle := lipgloss.NewStyle().Foreground(y.styles.Text)
	stringStyle := lipgloss.NewStyle().Foreground(y.styles.Secondary)
	commentStyle := lipgloss.NewStyle().Foreground(y.styles.MutedColor)

	trimmed := strings.TrimLeft(line, " ")
	indent := strings.Repeat(" ", len(line)-len(trimmed))

	// Comment
	if strings.HasPrefix(trimmed, "#") {
		return commentStyle.Render(line)
	}

	// Key: value
	if colonIdx := strings.Index(trimmed, ":"); colonIdx > 0 {
		key := trimmed[:colonIdx]
		rest := trimmed[colonIdx:]

		// Check if it's just a key with no value (nested object)
		if len(rest) == 1 ||
			(len(rest) > 1 && rest[1] == ' ' && len(strings.TrimSpace(rest[1:])) == 0) {
			return indent + keyStyle.Render(key) + valueStyle.Render(":")
		}

		value := rest[1:]
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		// String value (quoted)
		if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
			return indent + keyStyle.Render(
				key,
			) + valueStyle.Render(
				": ",
			) + stringStyle.Render(
				value,
			)
		}

		return indent + keyStyle.Render(key) + valueStyle.Render(": "+value)
	}

	// List item
	if strings.HasPrefix(trimmed, "- ") {
		return indent + valueStyle.Render("- ") + valueStyle.Render(trimmed[2:])
	}

	return valueStyle.Render(line)
}

func (y *YamlViewer) Content() string {
	return y.content
}
