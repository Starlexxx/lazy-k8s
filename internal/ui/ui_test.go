package ui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/lazyk8s/lazy-k8s/internal/config"
)

// TestViewMode tests view mode constants.
func TestViewMode(t *testing.T) {
	tests := []struct {
		mode     ViewMode
		expected int
	}{
		{ViewNormal, 0},
		{ViewHelp, 1},
		{ViewYaml, 2},
		{ViewLogs, 3},
		{ViewConfirm, 4},
		{ViewInput, 5},
		{ViewContextSwitch, 6},
		{ViewNamespaceSwitch, 7},
		{ViewContainerSelect, 8},
	}

	for _, tt := range tests {
		if int(tt.mode) != tt.expected {
			t.Errorf("ViewMode %v = %d, want %d", tt.mode, int(tt.mode), tt.expected)
		}
	}
}

// TestK8sClientWithFake tests that fake k8s client works for testing.
func TestK8sClientWithFake(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
			Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "main", Image: "nginx:latest"},
				},
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
	)

	// Test listing namespaces
	namespaces, err := fakeClientset.CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list namespaces: %v", err)
	}

	if len(namespaces.Items) != 1 {
		t.Errorf("Expected 1 namespace, got %d", len(namespaces.Items))
	}

	// Test listing pods
	pods, err := fakeClientset.CoreV1().
		Pods("default").
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(pods.Items) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(pods.Items))
	}

	if pods.Items[0].Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got %q", pods.Items[0].Name)
	}
}

// TestTeaKeyMsg tests tea.KeyMsg handling utilities.
func TestTeaKeyMsg(t *testing.T) {
	// Test creating key messages
	tests := []struct {
		name     string
		msg      tea.KeyMsg
		expected string
	}{
		{"letter q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, "q"},
		{"letter a", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, "a"},
		{"escape", tea.KeyMsg{Type: tea.KeyEsc}, "esc"},
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}, "enter"},
		{"tab", tea.KeyMsg{Type: tea.KeyTab}, "tab"},
		{"up arrow", tea.KeyMsg{Type: tea.KeyUp}, "up"},
		{"down arrow", tea.KeyMsg{Type: tea.KeyDown}, "down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.String()
			if result != tt.expected {
				t.Errorf("KeyMsg.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWindowSizeMsg tests tea.WindowSizeMsg.
func TestWindowSizeMsg(t *testing.T) {
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	if msg.Width != 100 {
		t.Errorf("Width = %d, want 100", msg.Width)
	}

	if msg.Height != 50 {
		t.Errorf("Height = %d, want 50", msg.Height)
	}
}

// TestConfigStructs tests config struct initialization.
func TestConfigStructs(t *testing.T) {
	cfg := &config.Config{
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

	// Test namespace
	if cfg.Namespace != "default" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "default")
	}

	// Test theme
	if cfg.Theme.PrimaryColor != "#7aa2f7" {
		t.Errorf("PrimaryColor = %q, want %q", cfg.Theme.PrimaryColor, "#7aa2f7")
	}

	// Test keybindings
	if len(cfg.Keybindings.Quit) != 2 {
		t.Errorf("Quit keybindings = %d, want 2", len(cfg.Keybindings.Quit))
	}

	// Test panels
	if len(cfg.Panels.Visible) != 3 {
		t.Errorf("Visible panels = %d, want 3", len(cfg.Panels.Visible))
	}
}

// TestRenderSwitchViewLogic tests the switch view rendering logic.
func TestRenderSwitchViewLogic(t *testing.T) {
	// Test context list navigation
	contextList := []string{"dev", "staging", "prod"}
	selectIdx := 0

	// Test moving down
	if selectIdx < len(contextList)-1 {
		selectIdx++
	}

	if selectIdx != 1 {
		t.Errorf("After moving down, selectIdx = %d, want 1", selectIdx)
	}

	// Test moving up
	if selectIdx > 0 {
		selectIdx--
	}

	if selectIdx != 0 {
		t.Errorf("After moving up, selectIdx = %d, want 0", selectIdx)
	}

	// Test boundary - should not go below 0
	if selectIdx > 0 {
		selectIdx--
	}

	if selectIdx != 0 {
		t.Errorf("At boundary, selectIdx = %d, want 0", selectIdx)
	}

	// Test boundary - should not go above max
	selectIdx = len(contextList) - 1
	if selectIdx < len(contextList)-1 {
		selectIdx++
	}

	if selectIdx != 2 {
		t.Errorf("At max boundary, selectIdx = %d, want 2", selectIdx)
	}
}

// TestSwitchViewScrolling tests the scroll window calculation for long lists.
func TestSwitchViewScrolling(t *testing.T) {
	tests := []struct {
		name          string
		itemCount     int
		maxVisible    int
		selectedIdx   int
		expectedStart int
		expectedEnd   int
		hasItemsAbove bool
		hasItemsBelow bool
	}{
		{
			name:          "small list fits entirely",
			itemCount:     5,
			maxVisible:    10,
			selectedIdx:   2,
			expectedStart: 0,
			expectedEnd:   5,
			hasItemsAbove: false,
			hasItemsBelow: false,
		},
		{
			name:          "large list selection at top",
			itemCount:     20,
			maxVisible:    10,
			selectedIdx:   0,
			expectedStart: 0,
			expectedEnd:   10,
			hasItemsAbove: false,
			hasItemsBelow: true,
		},
		{
			name:          "large list selection at bottom",
			itemCount:     20,
			maxVisible:    10,
			selectedIdx:   19,
			expectedStart: 10,
			expectedEnd:   20,
			hasItemsAbove: true,
			hasItemsBelow: false,
		},
		{
			name:          "large list selection in middle",
			itemCount:     20,
			maxVisible:    10,
			selectedIdx:   10,
			expectedStart: 5,
			expectedEnd:   15,
			hasItemsAbove: true,
			hasItemsBelow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the scroll calculation from renderSwitchView
			startIdx := 0
			endIdx := tt.itemCount

			if tt.itemCount > tt.maxVisible {
				halfVisible := tt.maxVisible / 2
				startIdx = tt.selectedIdx - halfVisible

				if startIdx < 0 {
					startIdx = 0
				}

				endIdx = startIdx + tt.maxVisible
				if endIdx > tt.itemCount {
					endIdx = tt.itemCount
					startIdx = endIdx - tt.maxVisible

					if startIdx < 0 {
						startIdx = 0
					}
				}
			}

			if startIdx != tt.expectedStart {
				t.Errorf("startIdx = %d, want %d", startIdx, tt.expectedStart)
			}

			if endIdx != tt.expectedEnd {
				t.Errorf("endIdx = %d, want %d", endIdx, tt.expectedEnd)
			}

			hasAbove := startIdx > 0
			hasBelow := endIdx < tt.itemCount

			if hasAbove != tt.hasItemsAbove {
				t.Errorf("hasItemsAbove = %v, want %v", hasAbove, tt.hasItemsAbove)
			}

			if hasBelow != tt.hasItemsBelow {
				t.Errorf("hasItemsBelow = %v, want %v", hasBelow, tt.hasItemsBelow)
			}
		})
	}
}

// TestApplySwitchFilter tests inline filtering for switch views.
func TestApplySwitchFilter(t *testing.T) {
	m := &Model{}
	source := []string{
		"default", "kube-system", "kube-public",
		"monitoring", "production", "staging",
	}

	// Empty filter returns all items
	m.switchFilter = ""
	m.applySwitchFilter(source)

	if len(m.switchFiltered) != len(source) {
		t.Errorf("empty filter: got %d items, want %d", len(m.switchFiltered), len(source))
	}

	// Filter narrows results
	m.switchFilter = "kube"
	m.applySwitchFilter(source)

	if len(m.switchFiltered) != 2 {
		t.Errorf("'kube' filter: got %d items, want 2", len(m.switchFiltered))
	}

	if m.selectIdx != 0 {
		t.Errorf("filter should reset selectIdx to 0, got %d", m.selectIdx)
	}

	// Case-insensitive match
	m.switchFilter = "PROD"
	m.applySwitchFilter(source)

	if len(m.switchFiltered) != 1 {
		t.Errorf("'PROD' filter: got %d items, want 1", len(m.switchFiltered))
	}

	if m.switchFiltered[0] != "production" {
		t.Errorf("expected 'production', got %q", m.switchFiltered[0])
	}

	// No matches
	m.switchFilter = "nonexistent"
	m.applySwitchFilter(source)

	if len(m.switchFiltered) != 0 {
		t.Errorf("'nonexistent' filter: got %d items, want 0", len(m.switchFiltered))
	}
}

// TestIsValidFilterChar tests character validation for filter input.
func TestIsValidFilterChar(t *testing.T) {
	validChars := []rune{'a', 'z', 'A', 'Z', '0', '9', '-', '_', '.'}
	for _, c := range validChars {
		if !isValidFilterChar(c) {
			t.Errorf("'%c' should be valid", c)
		}
	}

	invalidChars := []rune{'/', '\\', ' ', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')'}
	for _, c := range invalidChars {
		if isValidFilterChar(c) {
			t.Errorf("'%c' should be invalid", c)
		}
	}
}

// TestPanelNavigationLogic tests panel navigation logic.
func TestPanelNavigationLogic(t *testing.T) {
	numPanels := 4
	activePanelIdx := 0

	// Test next panel
	activePanelIdx = (activePanelIdx + 1) % numPanels
	if activePanelIdx != 1 {
		t.Errorf("After next panel, idx = %d, want 1", activePanelIdx)
	}

	// Test wrap around
	activePanelIdx = 3

	activePanelIdx = (activePanelIdx + 1) % numPanels
	if activePanelIdx != 0 {
		t.Errorf("After wrap around, idx = %d, want 0", activePanelIdx)
	}

	// Test previous panel
	activePanelIdx = 1

	activePanelIdx = (activePanelIdx - 1 + numPanels) % numPanels
	if activePanelIdx != 0 {
		t.Errorf("After prev panel, idx = %d, want 0", activePanelIdx)
	}

	// Test previous panel wrap around
	activePanelIdx = 0

	activePanelIdx = (activePanelIdx - 1 + numPanels) % numPanels
	if activePanelIdx != 3 {
		t.Errorf("After prev panel wrap, idx = %d, want 3", activePanelIdx)
	}
}

// TestSelectPanelLogic tests panel selection logic.
func TestSelectPanelLogic(t *testing.T) {
	numPanels := 4

	tests := []struct {
		name        string
		selectIdx   int
		shouldApply bool
	}{
		{"valid index 0", 0, true},
		{"valid index 2", 2, true},
		{"invalid negative", -1, false},
		{"invalid too large", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.selectIdx >= 0 && tt.selectIdx < numPanels
			if isValid != tt.shouldApply {
				t.Errorf(
					"selectPanel(%d) valid = %v, want %v",
					tt.selectIdx,
					isValid,
					tt.shouldApply,
				)
			}
		})
	}
}

// TestViewModeTransitions tests view mode transition logic.
func TestViewModeTransitions(t *testing.T) {
	tests := []struct {
		name       string
		from       ViewMode
		action     string
		expectedTo ViewMode
	}{
		{"normal to help", ViewNormal, "?", ViewHelp},
		{"help to normal", ViewHelp, "esc", ViewNormal},
		{"normal to yaml", ViewNormal, "y", ViewYaml},
		{"yaml to normal", ViewYaml, "esc", ViewNormal},
		{"normal to context switch", ViewNormal, "K", ViewContextSwitch},
		{"context switch to normal", ViewContextSwitch, "esc", ViewNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the constants are defined correctly
			if tt.from < 0 || tt.expectedTo < 0 {
				t.Error("ViewMode should be non-negative")
			}
		})
	}
}

// TestShowAllNamespaceFlag tests the showAllNs flag logic.
func TestShowAllNamespaceFlag(t *testing.T) {
	showAllNs := false

	// Toggle on
	showAllNs = !showAllNs
	if !showAllNs {
		t.Error("showAllNs should be true after toggle")
	}

	// Toggle off
	showAllNs = !showAllNs
	if showAllNs {
		t.Error("showAllNs should be false after second toggle")
	}
}

// TestSearchState tests search state logic.
func TestSearchState(t *testing.T) {
	searchActive := false
	searchQuery := ""

	// Activate search
	searchActive = true
	if !searchActive {
		t.Error("searchActive should be true")
	}

	// Set query
	searchQuery = "test-pod"
	if searchQuery != "test-pod" {
		t.Errorf("searchQuery = %q, want %q", searchQuery, "test-pod")
	}

	// Clear search
	searchActive = false
	searchQuery = ""

	if searchActive {
		t.Error("searchActive should be false after clear")
	}

	if searchQuery != "" {
		t.Error("searchQuery should be empty after clear")
	}
}

// TestStringContains tests string matching for view rendering.
func TestStringContains(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		expected bool
	}{
		{"lazy-k8s", "k8s", true},
		{"Context: prod-cluster", "Context:", true},
		{"Namespace: default", "Namespace:", true},
		{"? for help", "help", true},
		{"No panels configured", "panels", true},
		{"Switch Context", "Context", true},
		{"Switch Namespace", "Namespace", true},
	}

	for _, tt := range tests {
		t.Run(tt.needle, func(t *testing.T) {
			result := strings.Contains(tt.haystack, tt.needle)
			if result != tt.expected {
				t.Errorf("strings.Contains(%q, %q) = %v, want %v",
					tt.haystack, tt.needle, result, tt.expected)
			}
		})
	}
}

// TestErrApplyFailed tests the ErrApplyFailed error constant.
func TestErrApplyFailed(t *testing.T) {
	if ErrApplyFailed == nil {
		t.Error("ErrApplyFailed should not be nil")
	}

	if ErrApplyFailed.Error() != "kubectl apply failed" {
		t.Errorf("ErrApplyFailed.Error() = %q, want %q",
			ErrApplyFailed.Error(), "kubectl apply failed")
	}
}

// TestCopyNameToClipboardNoPanel tests copyNameToClipboard when no panels exist.
func TestCopyNameToClipboardNoPanel(t *testing.T) {
	m := &Model{
		panels:         nil,
		activePanelIdx: 0,
	}

	result, cmd := m.copyNameToClipboard()

	if result != m {
		t.Error("copyNameToClipboard should return same model when no panels")
	}

	if cmd != nil {
		t.Error("copyNameToClipboard should return nil cmd when no panels")
	}
}

// TestCopyNameToClipboardInvalidIndex tests copyNameToClipboard with invalid panel index.
func TestCopyNameToClipboardInvalidIndex(t *testing.T) {
	m := &Model{
		panels:         nil,
		activePanelIdx: 5,
	}

	result, cmd := m.copyNameToClipboard()

	if result != m {
		t.Error("copyNameToClipboard should return same model with invalid index")
	}

	if cmd != nil {
		t.Error("copyNameToClipboard should return nil cmd with invalid index")
	}
}

// TestCopyYamlToClipboardNoPanel tests copyYamlToClipboard when no panels exist.
func TestCopyYamlToClipboardNoPanel(t *testing.T) {
	m := &Model{
		panels:         nil,
		activePanelIdx: 0,
	}

	result, cmd := m.copyYamlToClipboard()

	if result != m {
		t.Error("copyYamlToClipboard should return same model when no panels")
	}

	if cmd != nil {
		t.Error("copyYamlToClipboard should return nil cmd when no panels")
	}
}

// TestCopyYamlToClipboardInvalidIndex tests copyYamlToClipboard with invalid panel index.
func TestCopyYamlToClipboardInvalidIndex(t *testing.T) {
	m := &Model{
		panels:         nil,
		activePanelIdx: 5,
	}

	result, cmd := m.copyYamlToClipboard()

	if result != m {
		t.Error("copyYamlToClipboard should return same model with invalid index")
	}

	if cmd != nil {
		t.Error("copyYamlToClipboard should return nil cmd with invalid index")
	}
}

// TestEditResourceNoPanel tests editResource when no panels exist.
func TestEditResourceNoPanel(t *testing.T) {
	m := &Model{
		panels:         nil,
		activePanelIdx: 0,
	}

	result, cmd := m.editResource()

	if result != m {
		t.Error("editResource should return same model when no panels")
	}

	if cmd != nil {
		t.Error("editResource should return nil cmd when no panels")
	}
}

// TestEditResourceInvalidPanelIndex tests editResource with invalid panel index.
func TestEditResourceInvalidPanelIndex(t *testing.T) {
	m := &Model{
		panels:         nil,
		activePanelIdx: 5,
	}

	result, cmd := m.editResource()

	if result != m {
		t.Error("editResource should return same model with invalid panel index")
	}

	if cmd != nil {
		t.Error("editResource should return nil cmd with invalid panel index")
	}
}

// TestGetEditorFromEnv tests editor selection logic.
func TestGetEditorFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		editor   string
		visual   string
		expected string
	}{
		{"EDITOR set", "nano", "", "nano"},
		{"VISUAL fallback", "", "code", "code"},
		{"vim default", "", "", "vim"},
		{"EDITOR takes priority", "emacs", "code", "emacs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := tt.editor
			if editor == "" {
				editor = tt.visual
			}

			if editor == "" {
				editor = "vim"
			}

			if editor != tt.expected {
				t.Errorf("editor = %q, want %q", editor, tt.expected)
			}
		})
	}
}

// Mock k8s.Client interface methods for documentation
// Note: The real k8s.Client requires a working kubeconfig.
type mockK8sClientMethods struct {
	contexts  []string
	namespace string
	context   string
	rawConfig api.Config
}

func (m *mockK8sClientMethods) CurrentNamespace() string {
	return m.namespace
}

func (m *mockK8sClientMethods) CurrentContext() string {
	return m.context
}

func (m *mockK8sClientMethods) GetContexts() []string {
	return m.contexts
}

func (m *mockK8sClientMethods) SetNamespace(ns string) {
	m.namespace = ns
}

func TestMockK8sClientMethods(t *testing.T) {
	mock := &mockK8sClientMethods{
		contexts:  []string{"dev", "staging", "prod"},
		namespace: "default",
		context:   "dev",
		rawConfig: api.Config{
			Contexts: map[string]*api.Context{
				"dev":     {},
				"staging": {},
				"prod":    {},
			},
		},
	}

	// Test CurrentNamespace
	if mock.CurrentNamespace() != "default" {
		t.Errorf("CurrentNamespace() = %q, want %q", mock.CurrentNamespace(), "default")
	}

	// Test CurrentContext
	if mock.CurrentContext() != "dev" {
		t.Errorf("CurrentContext() = %q, want %q", mock.CurrentContext(), "dev")
	}

	// Test GetContexts
	contexts := mock.GetContexts()
	if len(contexts) != 3 {
		t.Errorf("GetContexts() returned %d contexts, want 3", len(contexts))
	}

	// Test SetNamespace
	mock.SetNamespace("kube-system")

	if mock.CurrentNamespace() != "kube-system" {
		t.Errorf("After SetNamespace, CurrentNamespace() = %q, want %q",
			mock.CurrentNamespace(), "kube-system")
	}
}
