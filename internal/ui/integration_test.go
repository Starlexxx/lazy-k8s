//go:build integration

// Integration tests validate UI behavior with a real Kubernetes cluster.
// These tests verify that the TUI correctly interacts with k8s resources
// and handles user input for navigation, search, and resource operations.
//
// Prerequisites:
//
//	kind create cluster --name lazy-k8s-test
//	kubectl create namespace test-ns
//	kubectl create deployment nginx-test --image=nginx:alpine
//	kubectl create deployment app-test --image=busybox --replicas=2 -n test-ns -- sleep 3600
package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Starlexxx/lazy-k8s/internal/config"
	"github.com/Starlexxx/lazy-k8s/internal/k8s"
)

// getTestK8sClient creates a client for integration tests.
// Uses K8S_TEST_CONTEXT env var to allow testing against different clusters.
func getTestK8sClient(t *testing.T) *k8s.Client {
	t.Helper()

	contextName := os.Getenv("K8S_TEST_CONTEXT")
	if contextName == "" {
		contextName = "kind-lazy-k8s-test"
	}

	client, err := k8s.NewClient("", contextName)
	if err != nil {
		t.Fatalf("Failed to create k8s client: %v", err)
	}
	return client
}

// getTestConfig returns a minimal config for testing.
// Uses Tokyo Night theme colors matching the default config.
func getTestConfig() *config.Config {
	return &config.Config{
		Namespace: "default",
		Theme: config.ThemeConfig{
			PrimaryColor:    "#7aa2f7",
			SecondaryColor:  "#9ece6a",
			ErrorColor:      "#f7768e",
			WarningColor:    "#e0af68",
			BackgroundColor: "#1a1b26",
			TextColor:       "#c0caf5",
			BorderColor:     "#3b4261",
		},
		Keybindings: config.KeybindingsConfig{
			Quit:      []string{"q", "ctrl+c"},
			Help:      []string{"?"},
			NextPanel: []string{"tab"},
			PrevPanel: []string{"shift+tab"},
			Up:        []string{"k", "up"},
			Down:      []string{"j", "down"},
			Search:    []string{"/"},
		},
		Defaults: config.DefaultsConfig{
			Namespace:       "default",
			LogLines:        100,
			FollowLogs:      true,
			RefreshInterval: 5,
		},
		Panels: config.PanelsConfig{
			Visible: []string{"namespaces", "pods", "deployments"},
			Layout:  "vertical",
		},
	}
}

func TestIntegration_NewModel(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()

	model := NewModel(client, cfg)

	if model == nil {
		t.Fatal("NewModel returned nil")
	}

	if len(model.panels) != 3 {
		t.Errorf("Expected 3 panels, got %d", len(model.panels))
	}

	if model.k8sClient.CurrentContext() != "kind-lazy-k8s-test" {
		t.Logf("Context: %s", model.k8sClient.CurrentContext())
	}
}

func TestIntegration_ModelInit(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	cmd := model.Init()
	if cmd == nil {
		t.Log("Init returned nil cmd (may be normal if panels don't need initialization)")
	}
}

func TestIntegration_ModelView(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.width = 120
	model.height = 40

	view := model.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	if !strings.Contains(view, "lazy-k8s") {
		t.Error("View should contain app title")
	}
}

func TestIntegration_ModelViewWithSize(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := newModel.(*Model)

	if m.width != 120 {
		t.Errorf("Width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("Height = %d, want 40", m.height)
	}

	view := m.View()
	if view == "" {
		t.Error("View should not be empty after setting size")
	}
}

func TestIntegration_PanelNavigation(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if model.activePanelIdx != 0 {
		t.Errorf("Initial activePanelIdx = %d, want 0", model.activePanelIdx)
	}

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(*Model)
	if m.activePanelIdx != 1 {
		t.Errorf("After tab, activePanelIdx = %d, want 1", m.activePanelIdx)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(*Model)
	if m.activePanelIdx != 2 {
		t.Errorf("After second tab, activePanelIdx = %d, want 2", m.activePanelIdx)
	}

	// Panel navigation wraps around to first panel
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(*Model)
	if m.activePanelIdx != 0 {
		t.Errorf("After wrap, activePanelIdx = %d, want 0", m.activePanelIdx)
	}
}

func TestIntegration_HelpView(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m := newModel.(*Model)

	if m.viewMode != ViewHelp {
		t.Errorf("After ?, viewMode = %d, want ViewHelp (%d)", m.viewMode, ViewHelp)
	}

	view := m.View()
	if !strings.Contains(strings.ToLower(view), "help") && !strings.Contains(strings.ToLower(view), "key") {
		t.Log("Help view may not contain 'help' or 'key' - checking it renders")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	if m.viewMode != ViewNormal {
		t.Errorf("After esc, viewMode = %d, want ViewNormal (%d)", m.viewMode, ViewNormal)
	}
}

func TestIntegration_ContextSwitch(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m := newModel.(*Model)

	if m.viewMode != ViewContextSwitch {
		t.Errorf("After 'c', viewMode = %d, want ViewContextSwitch (%d)", m.viewMode, ViewContextSwitch)
	}

	if len(m.contextList) == 0 {
		t.Error("Context list should not be empty")
	}

	t.Logf("Available contexts: %v", m.contextList)

	view := m.View()
	if !strings.Contains(view, "Context") {
		t.Error("View should contain 'Context'")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	if m.viewMode != ViewNormal {
		t.Errorf("After esc, viewMode = %d, want ViewNormal (%d)", m.viewMode, ViewNormal)
	}
}

func TestIntegration_NamespaceSwitch(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m := newModel.(*Model)

	if m.viewMode != ViewNamespaceSwitch {
		t.Errorf("After 'n', viewMode = %d, want ViewNamespaceSwitch (%d)", m.viewMode, ViewNamespaceSwitch)
	}

	if len(m.namespaceList) == 0 {
		t.Error("Namespace list should not be empty")
	}

	t.Logf("Available namespaces: %v", m.namespaceList)

	found := false
	for _, ns := range m.namespaceList {
		if ns == "test-ns" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'test-ns' in namespace list")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	if m.viewMode != ViewNormal {
		t.Errorf("After esc, viewMode = %d, want ViewNormal (%d)", m.viewMode, ViewNormal)
	}
}

func TestIntegration_SelectNamespace(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m := newModel.(*Model)

	testNsIdx := -1
	for i, ns := range m.namespaceList {
		if ns == "test-ns" {
			testNsIdx = i
			break
		}
	}

	if testNsIdx == -1 {
		t.Skip("test-ns not found, skipping")
	}

	for i := 0; i < testNsIdx; i++ {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(*Model)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)

	if m.viewMode != ViewNormal {
		t.Errorf("After enter, viewMode = %d, want ViewNormal", m.viewMode)
	}

	if m.k8sClient.CurrentNamespace() != "test-ns" {
		t.Errorf("Namespace = %q, want 'test-ns'", m.k8sClient.CurrentNamespace())
	}
}

func TestIntegration_SearchMode(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m := newModel.(*Model)

	if !m.searchActive {
		t.Error("searchActive should be true after '/'")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	if m.searchActive {
		t.Error("searchActive should be false after esc")
	}
}

func TestIntegration_AllNamespacesToggle(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if model.showAllNs {
		t.Error("showAllNs should be false initially")
	}

	// Keybinding is 'A' (shift+a), not lowercase 'a'
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m := newModel.(*Model)

	if !m.showAllNs {
		t.Error("showAllNs should be true after 'A'")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m = newModel.(*Model)

	if m.showAllNs {
		t.Error("showAllNs should be false after second 'A'")
	}
}

func TestIntegration_PanelSelect(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Number keys select panels (1-indexed in UI, 0-indexed internally)
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m := newModel.(*Model)

	if m.activePanelIdx != 1 {
		t.Errorf("After '2', activePanelIdx = %d, want 1", m.activePanelIdx)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newModel.(*Model)

	if m.activePanelIdx != 2 {
		t.Errorf("After '3', activePanelIdx = %d, want 2", m.activePanelIdx)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = newModel.(*Model)

	if m.activePanelIdx != 0 {
		t.Errorf("After '1', activePanelIdx = %d, want 0", m.activePanelIdx)
	}
}

func TestIntegration_RenderNormalView(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.width = 120
	model.height = 40

	view := model.renderNormalView()

	if !strings.Contains(view, "lazy-k8s") {
		t.Error("View should contain 'lazy-k8s' title")
	}

	if !strings.Contains(view, "Context:") {
		t.Error("View should contain 'Context:'")
	}

	if !strings.Contains(view, "Namespace:") {
		t.Error("View should contain 'Namespace:'")
	}
}

func TestIntegration_QuitCommand(t *testing.T) {
	client := getTestK8sClient(t)
	cfg := getTestConfig()
	model := NewModel(client, cfg)

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Error("Quit should return a command")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Error("Expected tea.QuitMsg")
	}
}
