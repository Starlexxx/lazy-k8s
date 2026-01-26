package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// createTestStyles is defined in header_test.go

func TestNewInput(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	if input == nil {
		t.Fatal("NewInput returned nil")
	}

	if input.IsActive() {
		t.Error("New input should not be active")
	}
}

func TestInputShow(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	input.Show("Test Title", "Test Description", "placeholder")

	if !input.IsActive() {
		t.Error("Input should be active after Show()")
	}

	if input.title != "Test Title" {
		t.Errorf("Input title = %q, want %q", input.title, "Test Title")
	}

	if input.description != "Test Description" {
		t.Errorf("Input description = %q, want %q", input.description, "Test Description")
	}
}

func TestInputHide(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	input.Show("Test", "Test", "placeholder")
	input.Hide()

	if input.IsActive() {
		t.Error("Input should not be active after Hide()")
	}
}

func TestInputUpdateEnter(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	input.Show("Test", "Test", "placeholder")
	input.SetValue("test-value")

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := input.Update(msg)

	if input.IsActive() {
		t.Error("Input should not be active after Enter")
	}

	// Execute the command and check the message
	if cmd == nil {
		t.Fatal("Update with Enter should return a command")
	}

	result := cmd()
	submitMsg, ok := result.(InputSubmitMsg)

	if !ok {
		t.Fatalf("Expected InputSubmitMsg, got %T", result)
	}

	if submitMsg.Value != "test-value" {
		t.Errorf("InputSubmitMsg.Value = %q, want %q", submitMsg.Value, "test-value")
	}
}

func TestInputUpdateEscape(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	input.Show("Test", "Test", "placeholder")

	// Simulate Escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := input.Update(msg)

	if input.IsActive() {
		t.Error("Input should not be active after Escape")
	}

	// Execute the command and check the message
	if cmd == nil {
		t.Fatal("Update with Escape should return a command")
	}

	result := cmd()
	_, ok := result.(InputCancelMsg)

	if !ok {
		t.Fatalf("Expected InputCancelMsg, got %T", result)
	}
}

func TestInputValue(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	input.SetValue("my-value")

	if input.Value() != "my-value" {
		t.Errorf("Input.Value() = %q, want %q", input.Value(), "my-value")
	}
}

func TestInputViewWhenInactive(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	view := input.View()
	if view != "" {
		t.Errorf("Inactive input should render empty string, got %q", view)
	}
}

func TestInputViewWhenActive(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	input.Show("Scale Deployment", "Enter replica count", "3")

	view := input.View()
	if view == "" {
		t.Error("Active input should render content")
	}
}

func TestInputUpdateWhenInactive(t *testing.T) {
	styles := createTestStyles()
	input := NewInput(styles)

	// Should not panic and should return nil command when inactive
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := input.Update(msg)

	if cmd != nil {
		t.Error("Update on inactive input should return nil command")
	}
}
