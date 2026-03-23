package components

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHistoryViewer_Navigation(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type: OpScaleDeployment, Resource: "a", Namespace: "ns",
	})
	store.Add(OperationRecord{
		Type: OpRestartDeployment, Resource: "b", Namespace: "ns",
	})
	store.Add(OperationRecord{
		Type: OpDeleteResource, Resource: "c", Namespace: "ns",
	})

	viewer := NewHistoryViewer(styles, store)

	// Starts at cursor 0 (newest item)
	rec, ok := viewer.SelectedRecord()
	if !ok {
		t.Fatal("expected selected record")
	}

	if rec.Resource != "c" {
		t.Errorf("expected resource c (newest), got %s", rec.Resource)
	}

	// Move down
	viewer, _ = viewer.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	rec, _ = viewer.SelectedRecord()
	if rec.Resource != "b" {
		t.Errorf("expected resource b, got %s", rec.Resource)
	}

	// Move to bottom
	viewer, _ = viewer.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})

	rec, _ = viewer.SelectedRecord()
	if rec.Resource != "a" {
		t.Errorf("expected resource a (oldest), got %s", rec.Resource)
	}

	// Move to top
	viewer, _ = viewer.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	rec, _ = viewer.SelectedRecord()
	if rec.Resource != "c" {
		t.Errorf("expected resource c, got %s", rec.Resource)
	}
}

func TestHistoryViewer_UndoRequest(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type:      OpScaleDeployment,
		Resource:  "nginx",
		Namespace: "default",
		Undoable:  true,
		UndoData:  UndoData{PreviousReplicas: 2},
	})

	viewer := NewHistoryViewer(styles, store)

	// Press 'u' on undoable record
	_, cmd := viewer.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}},
	)

	if cmd == nil {
		t.Fatal("expected undo command to be returned")
	}

	msg := cmd()
	undoMsg, ok := msg.(UndoRequestMsg)

	if !ok {
		t.Fatalf("expected UndoRequestMsg, got %T", msg)
	}

	if undoMsg.RecordID != 0 {
		t.Errorf("expected record ID 0, got %d", undoMsg.RecordID)
	}
}

func TestHistoryViewer_UndoNotAvailableForUndone(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	id := store.Add(OperationRecord{
		Type:     OpScaleDeployment,
		Resource: "nginx",
		Undoable: true,
	})

	store.MarkUndone(id)

	viewer := NewHistoryViewer(styles, store)
	_, cmd := viewer.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}},
	)

	if cmd != nil {
		t.Error("expected nil cmd for already-undone record")
	}
}

func TestHistoryViewer_ViewEmpty(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()
	viewer := NewHistoryViewer(styles, store)

	output := viewer.View(80, 24)

	if !strings.Contains(output, "Operations History") {
		t.Error("expected title in output")
	}

	if !strings.Contains(output, "No operations recorded") {
		t.Error("expected empty state message")
	}
}

func TestHistoryViewer_ViewWithRecords(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type:      OpScaleDeployment,
		Resource:  "nginx",
		Namespace: "default",
		Timestamp: time.Date(2026, 1, 1, 12, 30, 0, 0, time.UTC),
		Undoable:  true,
	})

	viewer := NewHistoryViewer(styles, store)
	output := viewer.View(100, 24)

	if !strings.Contains(output, "12:30:00") {
		t.Error("expected timestamp in output")
	}

	if !strings.Contains(output, "Scale Deployment") {
		t.Error("expected operation label in output")
	}

	if !strings.Contains(output, "nginx") {
		t.Error("expected resource name in output")
	}

	if !strings.Contains(output, "[undo]") {
		t.Error("expected [undo] tag for undoable record")
	}
}

func TestHistoryViewer_ViewUndoneRecord(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	id := store.Add(OperationRecord{
		Type:      OpScaleDeployment,
		Resource:  "nginx",
		Namespace: "default",
		Timestamp: time.Date(2026, 1, 1, 12, 30, 0, 0, time.UTC),
		Undoable:  true,
	})

	store.MarkUndone(id)

	viewer := NewHistoryViewer(styles, store)
	output := viewer.View(100, 24)

	if !strings.Contains(output, "[undone]") {
		t.Error("expected [undone] tag for undone record")
	}
}

func TestHistoryViewer_Reset(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	for i := range 5 {
		store.Add(OperationRecord{
			Type:     OpDeleteResource,
			Resource: string(rune('a' + i)),
		})
	}

	viewer := NewHistoryViewer(styles, store)

	// Move cursor down
	viewer, _ = viewer.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
	)
	viewer, _ = viewer.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
	)

	viewer.Reset()

	if viewer.cursor != 0 {
		t.Errorf("expected cursor 0 after reset, got %d", viewer.cursor)
	}

	if viewer.offset != 0 {
		t.Errorf("expected offset 0 after reset, got %d", viewer.offset)
	}
}

func TestHistoryViewer_PageNavigation(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	// Add enough records to require scrolling
	for i := range 50 {
		store.Add(OperationRecord{
			Type:      OpDeleteResource,
			Resource:  string(rune('a' + (i % 26))),
			Namespace: "default",
			Timestamp: time.Now(),
		})
	}

	viewer := NewHistoryViewer(styles, store)

	// Must render once to set height so visibleHeight() works
	viewer.View(100, 20)

	// ctrl+d for page down (matching the handler in history_viewer.go)
	viewer, _ = viewer.Update(
		tea.KeyMsg{Type: tea.KeyCtrlD},
	)

	if viewer.cursor == 0 {
		t.Error("expected cursor to move after ctrl+d")
	}

	// ctrl+u for page up
	prevCursor := viewer.cursor

	viewer, _ = viewer.Update(
		tea.KeyMsg{Type: tea.KeyCtrlU},
	)

	if viewer.cursor >= prevCursor {
		t.Error("expected cursor to move up after ctrl+u")
	}
}

func TestHistoryViewer_ScrollbarRendered(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	// Add many records to trigger scrollbar
	for i := range 50 {
		store.Add(OperationRecord{
			Type:      OpDeleteResource,
			Resource:  string(rune('a' + (i % 26))),
			Namespace: "default",
			Timestamp: time.Now(),
		})
	}

	viewer := NewHistoryViewer(styles, store)
	output := viewer.View(100, 20)

	// Scrollbar uses "█" indicator
	if !strings.Contains(output, "█") {
		t.Error("expected scrollbar indicator in output")
	}
}

func TestHistoryViewer_NonUndoableNoCmd(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type:     OpDeleteResource,
		Resource: "pod",
		Undoable: false,
	})

	viewer := NewHistoryViewer(styles, store)
	_, cmd := viewer.Update(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}},
	)

	if cmd != nil {
		t.Error("expected nil cmd for non-undoable record")
	}
}

func TestHistoryViewer_NonKeyMsgIgnored(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()
	viewer := NewHistoryViewer(styles, store)

	// Non-key message should be ignored
	_, cmd := viewer.Update("not a key msg")

	if cmd != nil {
		t.Error("expected nil cmd for non-key message")
	}
}

func TestHistoryViewer_SelectedRecordEmpty(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()
	viewer := NewHistoryViewer(styles, store)

	_, ok := viewer.SelectedRecord()
	if ok {
		t.Error("expected ok=false for empty store")
	}
}

func TestHistoryViewer_RenderNonUndoableRecord(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type:      OpRestartDeployment,
		Resource:  "nginx",
		Namespace: "default",
		Timestamp: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		Undoable:  false,
	})

	viewer := NewHistoryViewer(styles, store)
	output := viewer.View(100, 24)

	if !strings.Contains(output, "Restart Deployment") {
		t.Error("expected operation label in output")
	}

	// Non-undoable should not have [undo] tag
	if strings.Contains(output, "[undo]") {
		t.Error("non-undoable record should not show [undo] tag")
	}
}

func TestHistoryViewer_LineTruncation(t *testing.T) {
	t.Parallel()

	styles := createTestStyles()
	store := NewHistoryStore()

	store.Add(OperationRecord{
		Type:      OpScaleDeployment,
		Resource:  "very-long-deployment-name-that-exceeds-normal-width",
		Namespace: "extremely-long-namespace-name",
		Timestamp: time.Now(),
		Undoable:  true,
	})

	viewer := NewHistoryViewer(styles, store)

	// Render at narrow width to trigger truncation
	output := viewer.View(50, 24)
	if !strings.Contains(output, "...") {
		t.Error("expected truncated line with '...'")
	}
}
