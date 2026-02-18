package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDiffViewer(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	if viewer == nil {
		t.Fatal("NewDiffViewer returned nil")
	}

	if len(viewer.Lines()) != 0 {
		t.Error("New viewer should have no lines")
	}
}

func TestDiffViewerSetContent(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Test Diff", "line1\nline2\n", "line1\nline3\n")

	if len(viewer.Lines()) == 0 {
		t.Error("SetContent should produce diff lines")
	}
}

func TestDiffViewerIdenticalContent(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	content := "apiVersion: v1\nkind: Pod\n"
	viewer.SetContent("No changes", content, content)

	for _, line := range viewer.Lines() {
		if line.Type != DiffLineContext {
			t.Errorf(
				"Identical content should produce only context lines, got type %d for %q",
				line.Type, line.Text,
			)
		}
	}
}

func TestDiffViewerAddedLines(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Added", "line1\n", "line1\nline2\n")

	hasAdded := false

	for _, line := range viewer.Lines() {
		if line.Type == DiffLineAdded {
			hasAdded = true

			break
		}
	}

	if !hasAdded {
		t.Error("Diff should contain added lines")
	}
}

func TestDiffViewerRemovedLines(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Removed", "line1\nline2\n", "line1\n")

	hasRemoved := false

	for _, line := range viewer.Lines() {
		if line.Type == DiffLineRemoved {
			hasRemoved = true

			break
		}
	}

	if !hasRemoved {
		t.Error("Diff should contain removed lines")
	}
}

func TestDiffViewerView(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Test Diff", "old\n", "new\n")

	view := viewer.View(80, 24)

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Test Diff") {
		t.Error("View should contain title")
	}
}

func TestDiffViewerScrollDown(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Scroll", old, old)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 1 {
		t.Errorf("offset = %d, want 1", viewer.offset)
	}
}

func TestDiffViewerScrollUp(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Scroll", old, old)
	viewer.offset = 5

	msg := tea.KeyMsg{Type: tea.KeyUp}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 4 {
		t.Errorf("offset = %d, want 4", viewer.offset)
	}
}

func TestDiffViewerScrollUpAtTop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Top", "a\nb\n", "a\nb\n")
	viewer.offset = 0

	msg := tea.KeyMsg{Type: tea.KeyUp}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 0 {
		t.Errorf("offset = %d, want 0 (should not go negative)", viewer.offset)
	}
}

func TestDiffViewerJumpToTop(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Jump", old, old)
	viewer.offset = 50

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	viewer, _ = viewer.Update(msg)

	if viewer.offset != 0 {
		t.Errorf("offset = %d, want 0", viewer.offset)
	}
}

func TestDiffViewerJumpToBottom(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Jump", old, old)
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	viewer, _ = viewer.Update(msg)

	if viewer.offset == 0 {
		t.Error("offset should not be 0 after jumping to bottom")
	}
}

func TestDiffViewerSearchActivation(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Search", "line1\n", "line1\n")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	viewer, _ = viewer.Update(msg)

	if !viewer.searchActive {
		t.Error("Search should be active after pressing /")
	}
}

func TestDiffViewerSearchEscape(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Search", "a\n", "a\n")
	viewer.searchActive = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after pressing Escape")
	}
}

func TestDiffViewerSearchEnter(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Search", "hello\nworld\n", "hello\nworld\n")
	viewer.searchActive = true
	viewer.searchQuery = "world"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after pressing Enter")
	}

	if len(viewer.matchLines) == 0 {
		t.Error("Should have found matches")
	}
}

func TestDiffViewerSearchNextMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Match", "foo\nbar\nfoo\n", "foo\nbar\nfoo\n")
	viewer.searchQuery = "foo"
	viewer.performSearch()
	viewer.matchIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 1 {
		t.Errorf("matchIndex = %d, want 1", viewer.matchIndex)
	}
}

func TestDiffViewerSearchPrevMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Match", "foo\nbar\nfoo\n", "foo\nbar\nfoo\n")
	viewer.searchQuery = "foo"
	viewer.performSearch()
	viewer.matchIndex = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 0 {
		t.Errorf("matchIndex = %d, want 0", viewer.matchIndex)
	}
}

func TestDiffViewerSearchPrevMatchWrap(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Wrap", "foo\nbar\nfoo\n", "foo\nbar\nfoo\n")
	viewer.searchQuery = "foo"
	viewer.performSearch()
	viewer.matchIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	viewer, _ = viewer.Update(msg)

	if viewer.matchIndex != 1 {
		t.Errorf("matchIndex = %d, want 1 (wrapped)", viewer.matchIndex)
	}
}

func TestDiffViewerPageDown(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Page", old, old)
	viewer.height = 20

	msg := tea.KeyMsg{Type: tea.KeyCtrlD}
	viewer, _ = viewer.Update(msg)

	if viewer.offset == 0 {
		t.Error("offset should not be 0 after page down")
	}
}

func TestDiffViewerPageUp(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Page", old, old)
	viewer.height = 20
	viewer.offset = 30

	msg := tea.KeyMsg{Type: tea.KeyCtrlU}
	viewer, _ = viewer.Update(msg)

	if viewer.offset >= 30 {
		t.Errorf("offset = %d, should be less than 30 after page up", viewer.offset)
	}
}

func TestDiffViewerViewNarrowWidth(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Narrow", "old\n", "new\n")

	view := viewer.View(15, 10)

	if view == "" {
		t.Error("View should not be empty with narrow width")
	}
}

func TestDiffViewerViewSmallHeight(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Small", "a\nb\n", "a\nc\n")

	view := viewer.View(80, 5)

	if view == "" {
		t.Error("View should not be empty with small height")
	}
}

func TestDiffViewerEmptyContent(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Empty", "", "")

	view := viewer.View(80, 24)
	if view == "" {
		t.Error("View should not be empty for empty content")
	}
}

func TestDiffViewerRenderDiffLinePrefixes(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Prefixes", "removed\nkept\n", "kept\nadded\n")

	hasPlus := false
	hasMinus := false

	view := viewer.View(80, 24)

	if strings.Contains(view, "+ ") {
		hasPlus = true
	}

	if strings.Contains(view, "- ") {
		hasMinus = true
	}

	if !hasPlus {
		t.Error("View should contain + prefix for added lines")
	}

	if !hasMinus {
		t.Error("View should contain - prefix for removed lines")
	}
}

func TestDiffViewerIsCurrentMatch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.matchLines = []int{0, 3, 7}
	viewer.matchIndex = 1

	if !viewer.isCurrentMatch(3) {
		t.Error("Line 3 should be current match")
	}

	if viewer.isCurrentMatch(0) {
		t.Error("Line 0 should not be current match")
	}
}

func TestDiffViewerIsCurrentMatchEmpty(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.matchLines = []int{}

	if viewer.isCurrentMatch(0) {
		t.Error("Should return false for empty matchLines")
	}
}

func TestDiffViewerIsCurrentMatchOutOfRange(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.matchLines = []int{1}
	viewer.matchIndex = 5

	if viewer.isCurrentMatch(1) {
		t.Error("Should return false when matchIndex is out of range")
	}
}

func TestDiffViewerNonKeyMsg(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Test", "a\n", "b\n")

	// Send a non-key message (WindowSizeMsg)
	result, cmd := viewer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if result != viewer {
		t.Error("Non-key msg should return same viewer")
	}

	if cmd != nil {
		t.Error("Non-key msg should return nil cmd")
	}
}

func TestDiffViewerSearchEnterNoMatches(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Search", "hello\n", "hello\n")
	viewer.searchActive = true
	viewer.searchQuery = "nonexistent"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	viewer, _ = viewer.Update(msg)

	if viewer.searchActive {
		t.Error("Search should be inactive after Enter")
	}

	if len(viewer.matchLines) != 0 {
		t.Error("Should have no matches for nonexistent query")
	}
}

func TestDiffViewerSearchEmptyQuery(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Search", "a\n", "a\n")
	viewer.searchQuery = ""
	viewer.performSearch()

	if len(viewer.matchLines) != 0 {
		t.Error("Empty query should produce no matches")
	}
}

func TestDiffViewerViewWithSearchMatches(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Match View", "hello\nworld\n", "hello\nworld\n")
	viewer.searchQuery = "hello"
	viewer.performSearch()
	viewer.matchIndex = 0

	view := viewer.View(80, 24)

	if !strings.Contains(view, "1/1 matches") {
		t.Error("View should show match count when search has results")
	}
}

func TestDiffViewerViewWithActiveSearch(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Active Search", "test\n", "test\n")
	viewer.searchActive = true

	view := viewer.View(80, 24)

	if !strings.Contains(view, "esc cancel") {
		t.Error("View should show search hint when search is active")
	}
}

func TestDiffViewerSearchInputForwarding(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Input", "a\n", "a\n")
	viewer.searchActive = true

	// Type a character while search is active
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	viewer, _ = viewer.Update(msg)

	if !viewer.searchActive {
		t.Error("Search should remain active after typing")
	}
}

func TestDiffViewerRenderDiffLineHeader(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	// Manually set lines with a header type
	viewer.lines = []DiffLine{
		{Text: "header text", Type: DiffLineHeader},
		{Text: "context text", Type: DiffLineContext},
		{Text: "added text", Type: DiffLineAdded},
		{Text: "removed text", Type: DiffLineRemoved},
	}
	viewer.title = "All Types"

	view := viewer.View(80, 24)

	if view == "" {
		t.Error("View should render all line types without panic")
	}
}

func TestDiffViewerRenderLongLine(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	longLine := strings.Repeat("x", 200)
	viewer.SetContent("Long", longLine+"\n", "short\n")

	// Render at narrow width to trigger truncation
	view := viewer.View(40, 24)

	if view == "" {
		t.Error("View should handle long line truncation")
	}
}

func TestDiffViewerScrollbarRendering(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "line"
	}

	old := strings.Join(lines, "\n") + "\n"
	viewer.SetContent("Scrollbar", old, old)

	// Render with enough height to show scrollbar
	view := viewer.View(80, 20)

	if !strings.Contains(view, "█") {
		t.Error("View should contain scrollbar indicator for long content")
	}
}

func TestDiffViewerScrollbarNotShownForShortContent(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Short", "a\nb\n", "a\nb\n")

	view := viewer.View(80, 40)

	if strings.Contains(view, "█") {
		t.Error("Scrollbar should not appear when content fits in view")
	}
}

func TestDiffViewerSearchHighlightInView(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Highlight", "foo\nbar\nfoo\n", "foo\nbar\nfoo\n")
	viewer.searchQuery = "foo"
	viewer.performSearch()
	viewer.matchIndex = 0

	// Render should include the search highlight indicator
	view := viewer.View(80, 24)

	if !strings.Contains(view, "►") {
		t.Error("View should show ► indicator for current search match")
	}
}

func TestDiffViewerMaxOffsetSmallContent(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Small", "a\n", "a\n")
	viewer.height = 40

	if viewer.maxOffset() != 0 {
		t.Errorf("maxOffset = %d, want 0 for small content", viewer.maxOffset())
	}
}

func TestDiffViewerScrollDownAtBottom(t *testing.T) {
	styles := createTestStyles()
	viewer := NewDiffViewer(styles)

	viewer.SetContent("Bottom", "a\nb\n", "a\nb\n")
	viewer.height = 40
	viewer.offset = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	viewer, _ = viewer.Update(msg)

	// Should not scroll past max offset (which is 0 for short content)
	if viewer.offset != 0 {
		t.Errorf("offset = %d, should stay 0 for short content", viewer.offset)
	}
}
