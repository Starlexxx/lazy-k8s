package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewLogViewer(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	if viewer == nil {
		t.Fatal("NewLogViewer returned nil")
	}

	if !viewer.follow {
		t.Error("New viewer should have follow enabled")
	}

	if viewer.maxLines != 10000 {
		t.Errorf("maxLines = %d, want 10000", viewer.maxLines)
	}
}

func TestLogViewerStop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.Stop()

	if viewer.cancel != nil {
		t.Error("cancel should be nil after Stop()")
	}
}

func TestLogViewerView(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.pod = "test-pod"
	viewer.lines = []string{"log line 1", "log line 2"}

	view := viewer.View(80, 24)

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Logs:") {
		t.Error("View should contain 'Logs:' title")
	}
}

func TestLogViewerViewWithFollow(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.pod = "test-pod"
	viewer.follow = true

	view := viewer.View(80, 24)

	if !strings.Contains(view, "FOLLOW") {
		t.Error("View should contain FOLLOW indicator when following")
	}
}

func TestLogViewerScrollDown(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "log line"
	}

	viewer.lines = lines
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyDown}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 1 {
		t.Errorf("offset = %d, want 1", viewer.offset)
	}
}

func TestLogViewerScrollUp(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"line1", "line2", "line3"}
	viewer.offset = 2

	msg := tea.KeyMsg{Type: tea.KeyUp}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 1 {
		t.Errorf("offset = %d, want 1", viewer.offset)
	}

	if viewer.follow {
		t.Error("follow should be disabled after scrolling up")
	}
}

func TestLogViewerScrollUpAtTop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"line1", "line2"}
	viewer.offset = 0

	msg := tea.KeyMsg{Type: tea.KeyUp}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 0 {
		t.Errorf("offset = %d, want 0", viewer.offset)
	}
}

func TestLogViewerToggleFollow(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.follow = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	viewer, _ = viewer.Update(msg)

	if viewer.follow {
		t.Error("follow should be toggled off")
	}

	viewer, _ = viewer.Update(msg)

	if !viewer.follow {
		t.Error("follow should be toggled on")
	}
}

func TestLogViewerJumpToTop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = make([]string, 100)
	viewer.offset = 50

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 0 {
		t.Errorf("offset = %d, want 0", viewer.offset)
	}

	if viewer.follow {
		t.Error("follow should be disabled after jumping to top")
	}
}

func TestLogViewerJumpToBottom(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = make([]string, 100)
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	viewer, _ = viewer.Update(msg)

	if viewer.offset == 0 {
		t.Error("offset should not be 0 after jumping to bottom")
	}

	if !viewer.follow {
		t.Error("follow should be enabled after jumping to bottom")
	}
}

func TestLogViewerSearchActivation(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"line1", "line2"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	viewer, _ = viewer.Update(msg)

	if !viewer.searchActive {
		t.Error("Search should be active after pressing /")
	}
}

func TestLogViewerSearchEscape(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.searchActive = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after pressing Escape")
	}
}

func TestLogViewerSearchEnter(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"error occurred", "info message", "error again"}
	viewer.searchActive = true
	viewer.searchQuery = "error"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after pressing Enter")
	}

	if len(viewer.matchLines) != 2 {
		t.Errorf("matchLines = %d, want 2", len(viewer.matchLines))
	}

	if viewer.follow {
		t.Error("follow should be disabled after search")
	}
}

func TestLogViewerSearchNextMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"error", "info", "error"}
	viewer.searchQuery = "error"
	viewer.performSearch()
	viewer.matchIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 1 {
		t.Errorf("matchIndex = %d, want 1", viewer.matchIndex)
	}
}

func TestLogViewerSearchPrevMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"error", "info", "error"}
	viewer.searchQuery = "error"
	viewer.performSearch()
	viewer.matchIndex = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 0 {
		t.Errorf("matchIndex = %d, want 0", viewer.matchIndex)
	}
}

func TestLogViewerSearchPrevMatchWrap(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"error", "info", "error"}
	viewer.searchQuery = "error"
	viewer.performSearch()
	viewer.matchIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 1 {
		t.Errorf("matchIndex = %d, want 1 (wrapped)", viewer.matchIndex)
	}
}

func TestLogViewerPerformSearchEmpty(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"line1", "line2"}
	viewer.searchQuery = ""
	viewer.performSearch()

	if len(viewer.matchLines) != 0 {
		t.Error("Empty query should not match anything")
	}
}

func TestLogViewerPerformSearchCaseInsensitive(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = []string{"ERROR occurred", "info message"}
	viewer.searchQuery = "error"
	viewer.performSearch()

	if len(viewer.matchLines) != 1 {
		t.Errorf("matchLines = %d, want 1 (case insensitive)", len(viewer.matchLines))
	}
}

func TestLogViewerIsCurrentMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.matchLines = []int{0, 2, 5}
	viewer.matchIndex = 1

	if !viewer.isCurrentMatch(2) {
		t.Error("Line 2 should be current match")
	}

	if viewer.isCurrentMatch(0) {
		t.Error("Line 0 should not be current match")
	}
}

func TestLogViewerIsCurrentMatchEmpty(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.matchLines = []int{}

	if viewer.isCurrentMatch(0) {
		t.Error("Should return false for empty matchLines")
	}
}

func TestLogViewerHighlightLogLine(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	tests := []struct {
		name string
		line string
	}{
		{"error line", "2024-01-15T10:30:00 ERROR something failed"},
		{"fatal line", "fatal: connection refused"},
		{"panic line", "panic: runtime error"},
		{"warn line", "WARN: disk space low"},
		{"info line", "info: server started"},
		{"timestamp line", "2024-01-15T10:30:00Z info message here"},
		{"plain line", "just a regular log line"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.highlightLogLine(tt.line)
			if result == "" {
				t.Error("highlightLogLine should not return empty string")
			}
		})
	}
}

func TestLogViewerPageDown(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = make([]string, 100)
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyCtrlD}
	viewer, _ = viewer.Update(msg)

	if viewer.offset == 0 {
		t.Error("offset should not be 0 after page down")
	}
}

func TestLogViewerPageUp(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.lines = make([]string, 100)
	viewer.height = 20
	viewer.offset = 30

	msg := tea.KeyMsg{Type: tea.KeyCtrlU}
	viewer, _ = viewer.Update(msg)

	if viewer.offset >= 30 {
		t.Errorf("offset = %d, should be less than 30 after page up", viewer.offset)
	}

	if viewer.follow {
		t.Error("follow should be disabled after page up")
	}
}

func TestLogViewerUpdateLogLineMsg(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.follow = true
	viewer.height = 20

	msg := LogLineMsg{Line: "new log line\nanother line"}
	viewer, cmd := viewer.Update(msg)

	if len(viewer.lines) != 2 {
		t.Errorf("lines = %d, want 2", len(viewer.lines))
	}

	if cmd == nil {
		t.Error("Should return a tick command")
	}
}

func TestLogViewerUpdateLogLineMsgError(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	msg := LogLineMsg{Error: &testError{}}
	viewer, _ = viewer.Update(msg)

	if len(viewer.lines) != 1 {
		t.Errorf("lines = %d, want 1", len(viewer.lines))
	}

	if !strings.Contains(viewer.lines[0], "Error:") {
		t.Error("Should contain error message")
	}
}

func TestLogViewerMaxLines(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.maxLines = 5
	viewer.follow = true
	viewer.height = 20

	for i := 0; i < 10; i++ {
		msg := LogLineMsg{Line: "line"}
		viewer, _ = viewer.Update(msg)
	}

	if len(viewer.lines) > 5 {
		t.Errorf("lines = %d, should not exceed maxLines (5)", len(viewer.lines))
	}
}

func TestLogViewerViewNarrowWidth(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.pod = "test"
	viewer.lines = []string{"line1", "line2"}

	view := viewer.View(20, 10)

	if view == "" {
		t.Error("View should not be empty with narrow width")
	}
}

func TestLogViewerViewSmallHeight(t *testing.T) {
	styles := createTestStyles()
	viewer := NewLogViewer(styles)

	viewer.pod = "test"
	viewer.lines = []string{"line1", "line2"}

	view := viewer.View(80, 5)

	if view == "" {
		t.Error("View should not be empty with small height")
	}
}

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}
