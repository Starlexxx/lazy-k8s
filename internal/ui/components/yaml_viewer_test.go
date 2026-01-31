package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewYamlViewer(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	if viewer == nil {
		t.Fatal("NewYamlViewer returned nil")
	}

	if viewer.Content() != "" {
		t.Error("New viewer should have empty content")
	}
}

func TestYamlViewerSetContent(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	yaml := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test"
	viewer.SetContent(yaml)

	if viewer.Content() != yaml {
		t.Errorf("Content() = %q, want %q", viewer.Content(), yaml)
	}
}

func TestYamlViewerView(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	yaml := "apiVersion: v1\nkind: Pod"
	viewer.SetContent(yaml)

	view := viewer.View(80, 24)

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "YAML Viewer") {
		t.Error("View should contain title")
	}
}

func TestYamlViewerScrollDown(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))

	msg := tea.KeyMsg{Type: tea.KeyDown}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 1 {
		t.Errorf("offset = %d, want 1", viewer.offset)
	}
}

func TestYamlViewerScrollUp(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))

	viewer.offset = 5

	msg := tea.KeyMsg{Type: tea.KeyUp}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 4 {
		t.Errorf("offset = %d, want 4", viewer.offset)
	}
}

func TestYamlViewerScrollUpAtTop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("line1\nline2")
	viewer.offset = 0

	msg := tea.KeyMsg{Type: tea.KeyUp}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 0 {
		t.Errorf("offset = %d, want 0 (should not go negative)", viewer.offset)
	}
}

func TestYamlViewerJumpToTop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))

	viewer.offset = 50

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 0 {
		t.Errorf("offset = %d, want 0", viewer.offset)
	}
}

func TestYamlViewerJumpToBottom(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	viewer, _ = viewer.Update(msg)

	if viewer.offset == 0 {
		t.Error("offset should not be 0 after jumping to bottom")
	}
}

func TestYamlViewerSearchActivation(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1\nkind: Pod")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	viewer, _ = viewer.Update(msg)

	if !viewer.searchActive {
		t.Error("Search should be active after pressing /")
	}
}

func TestYamlViewerSearchEscape(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1\nkind: Pod")
	viewer.searchActive = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after pressing Escape")
	}
}

func TestYamlViewerSearchEnter(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1\nkind: Pod\nname: test")
	viewer.searchActive = true
	viewer.searchQuery = "Pod"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after pressing Enter")
	}

	if len(viewer.matchLines) == 0 {
		t.Error("Should have found matches")
	}
}

func TestYamlViewerSearchNextMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("Pod\nService\nPod")
	viewer.searchQuery = "Pod"
	viewer.performSearch()
	viewer.matchIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 1 {
		t.Errorf("matchIndex = %d, want 1", viewer.matchIndex)
	}
}

func TestYamlViewerSearchPrevMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("Pod\nService\nPod")
	viewer.searchQuery = "Pod"
	viewer.performSearch()
	viewer.matchIndex = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 0 {
		t.Errorf("matchIndex = %d, want 0", viewer.matchIndex)
	}
}

func TestYamlViewerSearchPrevMatchWrap(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("Pod\nService\nPod")
	viewer.searchQuery = "Pod"
	viewer.performSearch()
	viewer.matchIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 1 {
		t.Errorf("matchIndex = %d, want 1 (wrapped)", viewer.matchIndex)
	}
}

func TestYamlViewerPerformSearchEmpty(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1")
	viewer.searchQuery = ""
	viewer.performSearch()

	if len(viewer.matchLines) != 0 {
		t.Error("Empty query should not match anything")
	}
}

func TestYamlViewerPerformSearchNoMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1")
	viewer.searchQuery = "nonexistent"
	viewer.performSearch()

	if len(viewer.matchLines) != 0 {
		t.Error("Should not find matches for nonexistent query")
	}
}

func TestYamlViewerPerformSearchCaseInsensitive(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1\nKIND: Pod")
	viewer.searchQuery = "kind"
	viewer.performSearch()

	if len(viewer.matchLines) != 1 {
		t.Errorf("matchLines = %d, want 1 (case insensitive)", len(viewer.matchLines))
	}
}

func TestYamlViewerIsCurrentMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.matchLines = []int{0, 2, 5}
	viewer.matchIndex = 1

	if !viewer.isCurrentMatch(2) {
		t.Error("Line 2 should be current match")
	}

	if viewer.isCurrentMatch(0) {
		t.Error("Line 0 should not be current match")
	}
}

func TestYamlViewerIsCurrentMatchEmpty(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.matchLines = []int{}

	if viewer.isCurrentMatch(0) {
		t.Error("Should return false for empty matchLines")
	}
}

func TestYamlViewerHighlightLine(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	tests := []struct {
		name     string
		line     string
		maxWidth int
	}{
		{"key value", "name: test", 50},
		{"comment", "# this is a comment", 50},
		{"list item", "- item1", 50},
		{"nested key", "metadata:", 50},
		{"quoted string", "name: \"test\"", 50},
		{"single quoted", "name: 'test'", 50},
		{"long line truncated", "verylongkey: verylongvaluethatwillbetruncated", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.highlightLine(tt.line, tt.maxWidth)
			if result == "" {
				t.Error("highlightLine should not return empty string")
			}
		})
	}
}

func TestYamlViewerPageDown(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyCtrlD}
	viewer, _ = viewer.Update(msg)

	if viewer.offset == 0 {
		t.Error("offset should not be 0 after page down")
	}
}

func TestYamlViewerPageUp(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))
	viewer.height = 20
	viewer.offset = 30

	msg := tea.KeyMsg{Type: tea.KeyCtrlU}
	viewer, _ = viewer.Update(msg)

	if viewer.offset >= 30 {
		t.Errorf("offset = %d, should be less than 30 after page up", viewer.offset)
	}
}

func TestYamlViewerViewNarrowWidth(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1")

	view := viewer.View(15, 10)

	if view == "" {
		t.Error("View should not be empty with narrow width")
	}
}

func TestYamlViewerViewSmallHeight(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	viewer.SetContent("apiVersion: v1\nkind: Pod")

	view := viewer.View(80, 5)

	if view == "" {
		t.Error("View should not be empty with small height")
	}
}

func TestYamlViewerScrollIndicatorSafeWidth(t *testing.T) {
	styles := createTestStyles()
	viewer := NewYamlViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	viewer.SetContent(strings.Join(lines, "\n"))

	view := viewer.View(12, 10)

	if view == "" {
		t.Error("View should not panic with width <= 12")
	}
}
