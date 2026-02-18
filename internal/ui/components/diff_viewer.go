package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

// DiffLineType categorizes a line in a unified diff view.
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdded
	DiffLineRemoved
	DiffLineHeader
)

// DiffLine represents a single line in the diff output.
type DiffLine struct {
	Text string
	Type DiffLineType
}

// DiffViewer displays a unified diff between two YAML strings
// with colored additions/removals and search support.
type DiffViewer struct {
	styles       *theme.Styles
	title        string
	lines        []DiffLine
	offset       int
	width        int
	height       int
	searchActive bool
	searchInput  textinput.Model
	searchQuery  string
	matchLines   []int
	matchIndex   int
}

// NewDiffViewer creates a DiffViewer with search text input pre-configured.
func NewDiffViewer(styles *theme.Styles) *DiffViewer {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 100
	ti.Width = 30
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)

	return &DiffViewer{
		styles:      styles,
		searchInput: ti,
		matchLines:  make([]int, 0),
	}
}

// SetContent computes a line-level diff between oldYAML and newYAML
// and stores the result for rendering.
func (d *DiffViewer) SetContent(title, oldYAML, newYAML string) {
	d.title = title
	d.offset = 0
	d.searchQuery = ""
	d.matchLines = make([]int, 0)
	d.matchIndex = 0

	d.lines = computeDiff(oldYAML, newYAML)
}

// computeDiff produces line-level diff output using character-based diffing
// mapped back to lines via DiffLinesToChars / DiffCharsToLines.
func computeDiff(oldText, newText string) []DiffLine {
	dmp := diffmatchpatch.New()

	// Convert line-level changes to character representation,
	// diff at character level, then map back to lines for readable output.
	charA, charB, lineArray := dmp.DiffLinesToChars(oldText, newText)
	diffs := dmp.DiffMain(charA, charB, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	var result []DiffLine

	for _, d := range diffs {
		lines := strings.Split(strings.TrimRight(d.Text, "\n"), "\n")

		for _, line := range lines {
			switch d.Type {
			case diffmatchpatch.DiffEqual:
				result = append(result, DiffLine{
					Text: line,
					Type: DiffLineContext,
				})
			case diffmatchpatch.DiffInsert:
				result = append(result, DiffLine{
					Text: line,
					Type: DiffLineAdded,
				})
			case diffmatchpatch.DiffDelete:
				result = append(result, DiffLine{
					Text: line,
					Type: DiffLineRemoved,
				})
			}
		}
	}

	return result
}

func (d *DiffViewer) Update(msg tea.Msg) (*DiffViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if d.searchActive {
			return d.handleSearchKey(msg)
		}

		return d.handleNormalKey(msg)
	}

	return d, nil
}

func (d *DiffViewer) handleSearchKey(msg tea.KeyMsg) (*DiffViewer, tea.Cmd) {
	switch msg.String() {
	case "esc":
		d.searchActive = false
		d.searchInput.Blur()

		return d, nil
	case "enter":
		d.searchActive = false
		d.searchInput.Blur()
		d.performSearch()

		if len(d.matchLines) > 0 {
			d.matchIndex = 0
			d.offset = d.matchLines[0]
		}

		return d, nil
	default:
		var cmd tea.Cmd

		d.searchInput, cmd = d.searchInput.Update(msg)
		d.searchQuery = d.searchInput.Value()

		return d, cmd
	}
}

func (d *DiffViewer) handleNormalKey(msg tea.KeyMsg) (*DiffViewer, tea.Cmd) {
	switch msg.String() {
	case "/":
		d.searchActive = true
		d.searchInput.Focus()
		d.searchInput.SetValue("")

		return d, nil
	case "n":
		if len(d.matchLines) > 0 {
			d.matchIndex = (d.matchIndex + 1) % len(d.matchLines)
			d.offset = d.matchLines[d.matchIndex]
		}
	case "N":
		if len(d.matchLines) > 0 {
			d.matchIndex--
			if d.matchIndex < 0 {
				d.matchIndex = len(d.matchLines) - 1
			}

			d.offset = d.matchLines[d.matchIndex]
		}
	case "up", "k":
		if d.offset > 0 {
			d.offset--
		}
	case "down", "j":
		if d.offset < d.maxOffset() {
			d.offset++
		}
	case "g":
		d.offset = 0
	case "G":
		d.offset = d.maxOffset()
	case "pgup", "ctrl+u":
		d.offset -= d.height / 2
		if d.offset < 0 {
			d.offset = 0
		}
	case "pgdown", "ctrl+d":
		d.offset += d.height / 2
		if d.offset > d.maxOffset() {
			d.offset = d.maxOffset()
		}
	}

	return d, nil
}

func (d *DiffViewer) maxOffset() int {
	offset := len(d.lines) - d.visibleHeight() + 1
	if offset < 0 {
		return 0
	}

	return offset
}

func (d *DiffViewer) visibleHeight() int {
	h := d.height - 7
	if d.searchActive || (d.searchQuery != "" && len(d.matchLines) > 0) {
		h--
	}

	if h < 1 {
		h = 1
	}

	return h
}

func (d *DiffViewer) performSearch() {
	d.matchLines = make([]int, 0)

	if d.searchQuery == "" {
		return
	}

	query := strings.ToLower(d.searchQuery)

	for i, line := range d.lines {
		if strings.Contains(strings.ToLower(line.Text), query) {
			d.matchLines = append(d.matchLines, i)
		}
	}
}

func (d *DiffViewer) View(width, height int) string {
	d.width = width
	d.height = height

	var b strings.Builder

	title := d.styles.ModalTitle.Render(d.title)

	var hint string
	if d.searchActive {
		hint = d.styles.Muted.Render("enter search • esc cancel")
	} else {
		hint = d.styles.Muted.Render(
			"/ search • n/N next/prev • ↑/↓ scroll • esc close",
		)
	}

	titleBar := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", hint)
	b.WriteString(titleBar)
	b.WriteString("\n")

	if d.searchActive {
		d.searchInput.Width = width - 10
		b.WriteString(d.searchInput.View())
		b.WriteString("\n")
	} else if d.searchQuery != "" && len(d.matchLines) > 0 {
		matchInfo := lipgloss.NewStyle().Foreground(d.styles.Primary).Render(
			fmt.Sprintf(
				" [%s] %d/%d matches",
				d.searchQuery,
				d.matchIndex+1,
				len(d.matchLines),
			),
		)
		b.WriteString(matchInfo)
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	visibleHeight := d.visibleHeight()

	endIdx := d.offset + visibleHeight
	if endIdx > len(d.lines) {
		endIdx = len(d.lines)
	}

	maxLineWidth := width - 8

	for i := d.offset; i < endIdx; i++ {
		line := d.lines[i]
		rendered := d.renderDiffLine(line, maxLineWidth, i)
		b.WriteString(rendered)
		b.WriteString("\n")
	}

	d.renderScrollbar(&b, visibleHeight, width)

	return d.styles.Modal.
		Width(width - 4).
		Height(height - 2).
		Render(b.String())
}

func (d *DiffViewer) renderDiffLine(line DiffLine, maxWidth, lineIdx int) string {
	text := line.Text
	if len(text) > maxWidth {
		text = text[:maxWidth-3] + "..."
	}

	var prefix string

	var style lipgloss.Style

	switch line.Type {
	case DiffLineAdded:
		prefix = "+ "
		style = d.styles.DiffAdded
	case DiffLineRemoved:
		prefix = "- "
		style = d.styles.DiffRemoved
	case DiffLineHeader:
		prefix = "  "
		style = d.styles.DetailTitle
	case DiffLineContext:
		prefix = "  "
		style = lipgloss.NewStyle().Foreground(d.styles.Text)
	}

	rendered := style.Render(prefix + text)

	// Highlight search matches
	if d.searchQuery != "" &&
		strings.Contains(strings.ToLower(line.Text), strings.ToLower(d.searchQuery)) {
		if d.isCurrentMatch(lineIdx) {
			rendered = d.styles.ListItemFocused.Render("► " + prefix + text)
		} else {
			rendered = lipgloss.NewStyle().
				Background(lipgloss.Color("#3d4966")).
				Render(prefix + text)
		}
	}

	return rendered
}

func (d *DiffViewer) renderScrollbar(b *strings.Builder, visibleHeight, width int) {
	if len(d.lines) <= visibleHeight || width <= 12 {
		return
	}

	scrollPos := float64(d.offset) / float64(len(d.lines)-visibleHeight)
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
	b.WriteString(d.styles.Muted.Render(indicator))
}

func (d *DiffViewer) isCurrentMatch(lineIdx int) bool {
	if len(d.matchLines) == 0 || d.matchIndex >= len(d.matchLines) {
		return false
	}

	return d.matchLines[d.matchIndex] == lineIdx
}

// Lines returns the computed diff lines (used in tests).
func (d *DiffViewer) Lines() []DiffLine {
	return d.lines
}
