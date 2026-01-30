package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type YamlViewer struct {
	styles       *theme.Styles
	content      string
	lines        []string
	offset       int
	width        int
	height       int
	searchActive bool
	searchInput  textinput.Model
	searchQuery  string
	matchLines   []int
	matchIndex   int
}

func NewYamlViewer(styles *theme.Styles) *YamlViewer {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 100
	ti.Width = 30
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)

	return &YamlViewer{
		styles:      styles,
		searchInput: ti,
		matchLines:  make([]int, 0),
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
		if y.searchActive {
			switch msg.String() {
			case "esc":
				y.searchActive = false
				y.searchInput.Blur()

				return y, nil
			case "enter":
				y.searchActive = false
				y.searchInput.Blur()
				y.performSearch()

				if len(y.matchLines) > 0 {
					y.matchIndex = 0
					y.offset = y.matchLines[0]
				}

				return y, nil
			default:
				var cmd tea.Cmd

				y.searchInput, cmd = y.searchInput.Update(msg)
				y.searchQuery = y.searchInput.Value()

				return y, cmd
			}
		}

		switch msg.String() {
		case "/":
			y.searchActive = true
			y.searchInput.Focus()
			y.searchInput.SetValue("")

			return y, nil
		case "n":
			if len(y.matchLines) > 0 {
				y.matchIndex = (y.matchIndex + 1) % len(y.matchLines)
				y.offset = y.matchLines[y.matchIndex]
			}
		case "N":
			if len(y.matchLines) > 0 {
				y.matchIndex--
				if y.matchIndex < 0 {
					y.matchIndex = len(y.matchLines) - 1
				}

				y.offset = y.matchLines[y.matchIndex]
			}
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

func (y *YamlViewer) performSearch() {
	y.matchLines = make([]int, 0)

	if y.searchQuery == "" {
		return
	}

	query := strings.ToLower(y.searchQuery)

	for i, line := range y.lines {
		if strings.Contains(strings.ToLower(line), query) {
			y.matchLines = append(y.matchLines, i)
		}
	}
}

func (y *YamlViewer) View(width, height int) string {
	y.width = width
	y.height = height

	var b strings.Builder

	title := y.styles.ModalTitle.Render("YAML Viewer")

	var hint string
	if y.searchActive {
		hint = y.styles.Muted.Render("enter search • esc cancel")
	} else {
		hint = y.styles.Muted.Render("/ search • n/N next/prev • ↑/↓ scroll • esc close")
	}

	titleBar := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", hint)
	b.WriteString(titleBar)
	b.WriteString("\n")

	if y.searchActive {
		y.searchInput.Width = width - 10
		b.WriteString(y.searchInput.View())
		b.WriteString("\n")
	} else if y.searchQuery != "" && len(y.matchLines) > 0 {
		matchInfo := y.styles.StatusValue.Render(
			lipgloss.NewStyle().Foreground(y.styles.Primary).Render(
				" [" + y.searchQuery + "] " +
					string(rune('0'+y.matchIndex+1)) + "/" +
					string(rune('0'+len(y.matchLines))) + " matches",
			),
		)
		b.WriteString(matchInfo)
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	visibleHeight := height - 7
	if y.searchActive || (y.searchQuery != "" && len(y.matchLines) > 0) {
		visibleHeight--
	}

	if visibleHeight < 1 {
		visibleHeight = 1
	}

	endIdx := y.offset + visibleHeight
	if endIdx > len(y.lines) {
		endIdx = len(y.lines)
	}

	for i := y.offset; i < endIdx; i++ {
		line := y.lines[i]
		highlighted := y.highlightLine(line, width-6)
		searchMatch := y.searchQuery != "" &&
			strings.Contains(strings.ToLower(line), strings.ToLower(y.searchQuery))

		if searchMatch {
			if y.isCurrentMatch(i) {
				highlighted = y.styles.ListItemFocused.Render("► " + highlighted)
			} else {
				highlighted = lipgloss.NewStyle().
					Background(lipgloss.Color("#3d4966")).
					Render(highlighted)
			}
		}

		b.WriteString(highlighted)
		b.WriteString("\n")
	}

	if len(y.lines) > visibleHeight && width > 12 {
		scrollPos := float64(y.offset) / float64(len(y.lines)-visibleHeight)
		barWidth := width - 10
		leftWidth := int(float64(barWidth) * scrollPos)

		if leftWidth < 0 {
			leftWidth = 0
		}

		if leftWidth > barWidth {
			leftWidth = barWidth
		}

		rightWidth := barWidth - leftWidth - 1
		if rightWidth < 0 {
			rightWidth = 0
		}

		indicator := strings.Repeat("─", leftWidth) + "█" + strings.Repeat("─", rightWidth)

		b.WriteString("\n")
		b.WriteString(y.styles.Muted.Render(indicator))
	}

	return y.styles.Modal.
		Width(width - 4).
		Height(height - 2).
		Render(b.String())
}

func (y *YamlViewer) isCurrentMatch(lineIdx int) bool {
	if len(y.matchLines) == 0 || y.matchIndex >= len(y.matchLines) {
		return false
	}

	return y.matchLines[y.matchIndex] == lineIdx
}

func (y *YamlViewer) highlightLine(line string, maxWidth int) string {
	if len(line) > maxWidth {
		line = line[:maxWidth-3] + "..."
	}

	keyStyle := lipgloss.NewStyle().Foreground(y.styles.Primary)
	valueStyle := lipgloss.NewStyle().Foreground(y.styles.Text)
	stringStyle := lipgloss.NewStyle().Foreground(y.styles.Secondary)
	commentStyle := lipgloss.NewStyle().Foreground(y.styles.MutedColor)

	trimmed := strings.TrimLeft(line, " ")
	indent := strings.Repeat(" ", len(line)-len(trimmed))

	if strings.HasPrefix(trimmed, "#") {
		return commentStyle.Render(line)
	}

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

	if strings.HasPrefix(trimmed, "- ") {
		return indent + valueStyle.Render("- ") + valueStyle.Render(trimmed[2:])
	}

	return valueStyle.Render(line)
}

func (y *YamlViewer) Content() string {
	return y.content
}
