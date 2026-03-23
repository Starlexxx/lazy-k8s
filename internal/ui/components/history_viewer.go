package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

// UndoRequestMsg is returned by the history viewer when the user
// presses 'u' on an undoable operation entry.
type UndoRequestMsg struct {
	RecordID int
}

// HistoryViewer renders a scrollable full-screen list of past operations
// with undo support for reversible entries.
type HistoryViewer struct {
	styles *theme.Styles
	store  *HistoryStore
	cursor int
	offset int
	width  int
	height int
}

func NewHistoryViewer(
	styles *theme.Styles,
	store *HistoryStore,
) *HistoryViewer {
	return &HistoryViewer{
		styles: styles,
		store:  store,
	}
}

// Reset returns cursor and scroll to the top.
func (h *HistoryViewer) Reset() {
	h.cursor = 0
	h.offset = 0
}

func (h *HistoryViewer) Update(msg tea.Msg) (*HistoryViewer, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return h, nil
	}

	records := h.store.Records()

	switch keyMsg.String() {
	case "up", "k":
		if h.cursor > 0 {
			h.cursor--
			h.ensureVisible()
		}
	case "down", "j":
		if h.cursor < len(records)-1 {
			h.cursor++
			h.ensureVisible()
		}
	case "g":
		h.cursor = 0
		h.offset = 0
	case "G":
		h.cursor = max(len(records)-1, 0)
		h.ensureVisible()
	case "pgup", "ctrl+u":
		h.cursor -= h.visibleHeight() / 2
		if h.cursor < 0 {
			h.cursor = 0
		}

		h.ensureVisible()
	case "pgdown", "ctrl+d":
		h.cursor += h.visibleHeight() / 2
		if h.cursor >= len(records) {
			h.cursor = max(len(records)-1, 0)
		}

		h.ensureVisible()
	case "u":
		if h.cursor < len(records) {
			rec := records[h.cursor]
			if rec.Undoable && !rec.Undone {
				return h, func() tea.Msg {
					return UndoRequestMsg{RecordID: rec.ID}
				}
			}
		}
	}

	return h, nil
}

func (h *HistoryViewer) visibleHeight() int {
	// Title (1) + separator (1) + hint (1) + modal padding (~4)
	const overhead = 7

	return max(h.height-overhead, 1)
}

func (h *HistoryViewer) ensureVisible() {
	vis := h.visibleHeight()

	if h.cursor < h.offset {
		h.offset = h.cursor
	}

	if h.cursor >= h.offset+vis {
		h.offset = h.cursor - vis + 1
	}
}

func (h *HistoryViewer) View(width, height int) string {
	h.width = width
	h.height = height

	records := h.store.Records()

	var b strings.Builder

	title := h.styles.ModalTitle.Render("Operations History")
	countLabel := h.styles.Muted.Render(
		fmt.Sprintf("(%d operations)", len(records)),
	)
	hint := h.styles.Muted.Render(
		"↑/↓ navigate • u undo • esc close",
	)
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Center, title, " ", countLabel, "  ", hint,
	)

	b.WriteString(titleBar)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	if len(records) == 0 {
		b.WriteString(h.styles.Muted.Render("  No operations recorded yet."))
		b.WriteString("\n")
	} else {
		vis := h.visibleHeight()
		endIdx := min(h.offset+vis, len(records))

		maxLineWidth := width - 8

		for i := h.offset; i < endIdx; i++ {
			line := h.renderRecord(records[i], i == h.cursor, maxLineWidth)
			b.WriteString(line)
			b.WriteString("\n")
		}

		h.renderScrollbar(&b, vis, width, len(records))
	}

	return h.styles.Modal.
		Width(width - 4).
		Height(height - 2).
		Render(b.String())
}

func (h *HistoryViewer) renderRecord(
	rec OperationRecord,
	selected bool,
	maxWidth int,
) string {
	ts := rec.Timestamp.Format("15:04:05")
	label := operationLabel(rec.Type)

	var undoTag string
	if rec.Undone {
		undoTag = " [undone]"
	} else if rec.Undoable {
		undoTag = " [undo]"
	}

	line := fmt.Sprintf(
		"%s  %-22s  %s/%s%s",
		ts, label, rec.Namespace, rec.Resource, undoTag,
	)

	if len(line) > maxWidth {
		line = line[:maxWidth-3] + "..."
	}

	if selected {
		return h.styles.ListItemFocused.Render("► " + line)
	}

	if rec.Undone {
		return h.styles.Muted.Render("  " + line)
	}

	return lipgloss.NewStyle().
		Foreground(h.styles.Text).
		Render("  " + line)
}

func (h *HistoryViewer) renderScrollbar(
	b *strings.Builder,
	visibleHeight, width, totalItems int,
) {
	if totalItems <= visibleHeight || width <= 12 {
		return
	}

	scrollPos := float64(h.offset) / float64(totalItems-visibleHeight)
	barWidth := width - 10

	leftWidth := max(int(float64(barWidth)*scrollPos), 0)
	leftWidth = min(leftWidth, barWidth)

	rightWidth := max(barWidth-leftWidth-1, 0)

	indicator := strings.Repeat("─", leftWidth) + "█" +
		strings.Repeat("─", rightWidth)

	b.WriteString("\n")
	b.WriteString(h.styles.Muted.Render(indicator))
}

// SelectedRecord returns the record currently under the cursor.
func (h *HistoryViewer) SelectedRecord() (OperationRecord, bool) {
	records := h.store.Records()
	if h.cursor < 0 || h.cursor >= len(records) {
		return OperationRecord{}, false
	}

	return records[h.cursor], true
}
