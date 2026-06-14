package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

// GlobalSearch manages the text input and cursor state for cross-resource search.
// Rendering is handled by ui.go to avoid import cycles with the panels package.
type GlobalSearch struct {
	styles *theme.Styles
	input  textinput.Model
	cursor int
	offset int
}

func NewGlobalSearch(styles *theme.Styles) *GlobalSearch {
	ti := textinput.New()
	ti.Placeholder = "type to search..."
	ti.CharLimit = 100
	ti.Width = 40
	ti.Prompt = "ctrl+f "
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(styles.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(styles.MutedColor)
	ti.Focus()

	return &GlobalSearch{
		styles: styles,
		input:  ti,
	}
}

// Reset clears input, cursor, and offset for a fresh search session.
func (g *GlobalSearch) Reset() {
	g.input.SetValue("")
	g.cursor = 0
	g.offset = 0
	g.input.Focus()
}

func (g *GlobalSearch) Query() string {
	return g.input.Value()
}

func (g *GlobalSearch) Cursor() int {
	return g.cursor
}

func (g *GlobalSearch) Offset() int {
	return g.offset
}

// SetVisibleHeight recalculates scroll offset bounds after the view renders.
func (g *GlobalSearch) SetVisibleHeight(visibleHeight int) {
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Keep cursor visible within the scroll window
	if g.cursor < g.offset {
		g.offset = g.cursor
	}

	if g.cursor >= g.offset+visibleHeight {
		g.offset = g.cursor - visibleHeight + 1
	}
}

func (g *GlobalSearch) InputView() string {
	return g.input.View()
}

// Update handles keyboard input and returns whether the query text changed.
func (g *GlobalSearch) Update(
	msg tea.Msg, resultCount int,
) (*GlobalSearch, tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return g, nil, false
	}

	switch keyMsg.String() {
	case "up", "ctrl+k":
		if g.cursor > 0 {
			g.cursor--
		}

		return g, nil, false

	case "down", "ctrl+j":
		if g.cursor < resultCount-1 {
			g.cursor++
		}

		return g, nil, false

	case "ctrl+u":
		g.cursor -= 10
		if g.cursor < 0 {
			g.cursor = 0
		}

		return g, nil, false

	case "ctrl+d":
		g.cursor += 10
		if g.cursor >= resultCount {
			g.cursor = max(resultCount-1, 0)
		}

		return g, nil, false
	}

	// Forward to text input
	oldValue := g.input.Value()

	var cmd tea.Cmd

	g.input, cmd = g.input.Update(msg)

	queryChanged := g.input.Value() != oldValue
	if queryChanged {
		// Reset cursor when query changes
		g.cursor = 0
		g.offset = 0
	}

	return g, cmd, queryChanged
}
