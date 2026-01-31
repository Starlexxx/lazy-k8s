package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSearch(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	if search == nil {
		t.Fatal("NewSearch returned nil")
	}

	if search.Value() != "" {
		t.Error("New search should have empty value")
	}
}

func TestSearchFocus(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	search.Focus()

	view := search.View(50)
	if view == "" {
		t.Error("Search view should not be empty after focus")
	}
}

func TestSearchBlur(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	search.Focus()
	search.Blur()

	view := search.View(50)
	if view == "" {
		t.Error("Search view should not be empty after blur")
	}
}

func TestSearchClear(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	search.SetValue("test")
	search.Clear()

	if search.Value() != "" {
		t.Errorf("Search.Value() = %q, want empty string", search.Value())
	}
}

func TestSearchSetValue(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	search.SetValue("my-search")

	if search.Value() != "my-search" {
		t.Errorf("Search.Value() = %q, want %q", search.Value(), "my-search")
	}
}

func TestSearchUpdate(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	search.Focus()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_, cmd := search.Update(msg)

	if cmd == nil {
		t.Log("Update returned nil command (expected for some key types)")
	}
}

func TestSearchView(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	view := search.View(80)

	if view == "" {
		t.Error("Search view should not be empty")
	}

	if !strings.Contains(view, "/") {
		t.Error("Search view should contain prompt character")
	}
}

func TestSearchViewNarrowWidth(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	view := search.View(20)

	if view == "" {
		t.Error("Search view should not be empty with narrow width")
	}
}

func TestSearchViewZeroWidth(t *testing.T) {
	styles := createTestStyles()
	search := NewSearch(styles)

	view := search.View(0)

	if view == "" {
		t.Error("Search view should not be empty with zero width")
	}
}
