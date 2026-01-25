package components

import (
	"strings"
	"testing"
)

func TestNewStatusBar(t *testing.T) {
	styles := createTestStyles()

	statusBar := NewStatusBar(styles)

	if statusBar == nil {
		t.Fatal("NewStatusBar returned nil")
	}

	if statusBar.message != "" {
		t.Errorf("Initial message = %q, want empty string", statusBar.message)
	}

	if statusBar.isError {
		t.Error("Initial isError should be false")
	}
}

func TestStatusBarSetMessage(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	statusBar.SetMessage("Test message")

	if statusBar.message != "Test message" {
		t.Errorf("message = %q, want %q", statusBar.message, "Test message")
	}

	if statusBar.isError {
		t.Error("isError should be false after SetMessage")
	}
}

func TestStatusBarSetError(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	statusBar.SetError("Connection failed")

	if statusBar.message != "Connection failed" {
		t.Errorf("message = %q, want %q", statusBar.message, "Connection failed")
	}

	if !statusBar.isError {
		t.Error("isError should be true after SetError")
	}
}

func TestStatusBarClear(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	// Set a message
	statusBar.SetMessage("Some message")

	// Clear it
	statusBar.Clear()

	if statusBar.message != "" {
		t.Errorf("After Clear, message = %q, want empty string", statusBar.message)
	}

	if statusBar.isError {
		t.Error("After Clear, isError should be false")
	}

	// Also test clearing an error
	statusBar.SetError("Some error")
	statusBar.Clear()

	if statusBar.message != "" {
		t.Errorf("After Clear (error), message = %q, want empty string", statusBar.message)
	}

	if statusBar.isError {
		t.Error("After Clear (error), isError should be false")
	}
}

func TestStatusBarViewDefault(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	view := statusBar.View(100)

	// Should contain default hints
	hints := []string{"quit", "help", "next panel", "search", "context", "namespace"}
	for _, hint := range hints {
		if !strings.Contains(view, hint) {
			t.Errorf("Default status bar view should contain %q hint", hint)
		}
	}
}

func TestStatusBarViewWithMessage(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	statusBar.SetMessage("Pod deleted successfully")
	view := statusBar.View(100)

	if !strings.Contains(view, "Pod deleted successfully") {
		t.Error("Status bar view should contain the message")
	}

	// Should not contain the default hints when showing a message
	// (they're replaced by the message)
}

func TestStatusBarViewWithError(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	statusBar.SetError("Failed to connect to cluster")
	view := statusBar.View(100)

	if !strings.Contains(view, "Error:") {
		t.Error("Error status bar view should contain 'Error:' prefix")
	}

	if !strings.Contains(view, "Failed to connect to cluster") {
		t.Error("Status bar view should contain the error message")
	}
}

func TestStatusBarViewNarrowWidth(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	// Test with narrow width (should not panic)
	view := statusBar.View(20)

	if view == "" {
		t.Error("Status bar view should not be empty even with narrow width")
	}
}

func TestStatusBarViewZeroWidth(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	// Test with zero width (should not panic)
	view := statusBar.View(0)

	// Should still render something
	if view == "" {
		t.Error("Status bar view should not be empty even with zero width")
	}
}

func TestStatusBarMessageOverridesError(t *testing.T) {
	styles := createTestStyles()
	statusBar := NewStatusBar(styles)

	// Set an error first
	statusBar.SetError("Error occurred")

	if !statusBar.isError {
		t.Error("isError should be true after SetError")
	}

	// Setting a message should clear the error flag
	statusBar.SetMessage("Success")

	if statusBar.isError {
		t.Error("isError should be false after SetMessage")
	}
}
