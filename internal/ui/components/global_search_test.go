package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewGlobalSearch(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)

	if gs == nil {
		t.Fatal("expected non-nil GlobalSearch")
	}

	if gs.Query() != "" {
		t.Errorf("expected empty query, got %q", gs.Query())
	}

	if gs.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", gs.Cursor())
	}
}

func TestGlobalSearchReset(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)

	// Set some state
	gs.cursor = 5
	gs.offset = 3
	gs.input.SetValue("nginx")

	gs.Reset()

	if gs.Query() != "" {
		t.Errorf("expected empty query after reset, got %q", gs.Query())
	}

	if gs.Cursor() != 0 {
		t.Errorf("expected cursor 0 after reset, got %d", gs.Cursor())
	}

	if gs.Offset() != 0 {
		t.Errorf("expected offset 0 after reset, got %d", gs.Offset())
	}
}

func TestGlobalSearchCursorMovement(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)
	resultCount := 5

	// Move down
	gs, _, _ = gs.Update(tea.KeyMsg{Type: tea.KeyDown}, resultCount)

	if gs.Cursor() != 1 {
		t.Errorf("expected cursor 1 after down, got %d", gs.Cursor())
	}

	// Move down more
	gs, _, _ = gs.Update(tea.KeyMsg{Type: tea.KeyDown}, resultCount)
	gs, _, _ = gs.Update(tea.KeyMsg{Type: tea.KeyDown}, resultCount)

	if gs.Cursor() != 3 {
		t.Errorf("expected cursor 3 after 3 downs, got %d", gs.Cursor())
	}

	// Move up
	gs, _, _ = gs.Update(tea.KeyMsg{Type: tea.KeyUp}, resultCount)

	if gs.Cursor() != 2 {
		t.Errorf("expected cursor 2 after up, got %d", gs.Cursor())
	}

	// Won't go below 0
	gs.cursor = 0

	gs, _, _ = gs.Update(tea.KeyMsg{Type: tea.KeyUp}, resultCount)

	if gs.Cursor() != 0 {
		t.Errorf("expected cursor 0 at boundary, got %d", gs.Cursor())
	}

	// Won't go above max
	gs.cursor = resultCount - 1

	gs, _, _ = gs.Update(tea.KeyMsg{Type: tea.KeyDown}, resultCount)

	if gs.Cursor() != resultCount-1 {
		t.Errorf(
			"expected cursor %d at boundary, got %d",
			resultCount-1, gs.Cursor(),
		)
	}
}

func TestGlobalSearchQueryChanged(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)

	// Typing a character should trigger queryChanged
	_, _, changed := gs.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}, 0,
	)

	if !changed {
		t.Error("expected queryChanged=true after typing")
	}

	// Cursor movement should not trigger queryChanged
	_, _, changed = gs.Update(tea.KeyMsg{Type: tea.KeyDown}, 5)

	if changed {
		t.Error("expected queryChanged=false after cursor move")
	}
}

func TestGlobalSearchCursorResetsOnQueryChange(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)
	gs.cursor = 3

	// Typing resets cursor to 0
	gs, _, _ = gs.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, 10,
	)

	if gs.Cursor() != 0 {
		t.Errorf(
			"expected cursor reset to 0 on query change, got %d",
			gs.Cursor(),
		)
	}
}

func TestGlobalSearchSetVisibleHeight(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)
	gs.cursor = 15
	gs.offset = 0

	// Should scroll to make cursor visible
	gs.SetVisibleHeight(10)

	if gs.offset < 6 {
		t.Errorf(
			"expected offset >= 6 to make cursor 15 visible in 10 lines, got %d",
			gs.offset,
		)
	}

	// Cursor above viewport
	gs.cursor = 2
	gs.offset = 10

	gs.SetVisibleHeight(10)

	if gs.offset != 2 {
		t.Errorf("expected offset 2 to make cursor visible, got %d", gs.offset)
	}
}

func TestGlobalSearchPageMovement(t *testing.T) {
	styles := createTestStyles()
	gs := NewGlobalSearch(styles)
	resultCount := 25

	// ctrl+d jumps forward by 10
	gs, _, _ = gs.Update(
		tea.KeyMsg{Type: tea.KeyCtrlD}, resultCount,
	)

	if gs.Cursor() != 10 {
		t.Errorf("expected cursor 10 after ctrl+d, got %d", gs.Cursor())
	}

	// ctrl+u jumps back by 10
	gs, _, _ = gs.Update(
		tea.KeyMsg{Type: tea.KeyCtrlU}, resultCount,
	)

	if gs.Cursor() != 0 {
		t.Errorf("expected cursor 0 after ctrl+u, got %d", gs.Cursor())
	}
}
