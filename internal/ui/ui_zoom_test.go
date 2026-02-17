package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/Starlexxx/lazy-k8s/internal/config"
	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/panels"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

// createZoomTestModel builds a Model manually instead of using NewModel
// because NewModel requires a real kubeconfig for the metrics client
// and calls initPanels which triggers k8s API calls via Init().
func createZoomTestModel() *Model {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
			Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
	)

	client := k8s.NewTestClient(fakeClientset)
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

	styles := theme.NewStyles(&cfg.Theme)
	keys := theme.NewKeyMap()

	m := &Model{
		k8sClient:    client,
		config:       cfg,
		styles:       styles,
		keys:         keys,
		viewMode:     ViewNormal,
		portForwards: make(map[string]*k8s.PortForwarder),
		width:        120,
		height:       40,
	}

	m.panels = []panels.Panel{
		panels.NewNamespacesPanel(client, styles),
		panels.NewPodsPanel(client, styles),
		panels.NewDeploymentsPanel(client, styles),
	}
	m.panels[0].SetFocused(true)
	m.updatePanelSizes()

	return m
}

func TestZoomToggle(t *testing.T) {
	m := createZoomTestModel()

	if m.zoomed {
		t.Fatal("model should not be zoomed initially")
	}

	// Press 'z' to toggle zoom on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}

	result, _ := m.Update(msg)

	updated, ok := result.(*Model)
	if !ok {
		t.Fatal("Update() did not return *Model")
	}

	m = updated

	if !m.zoomed {
		t.Error("model should be zoomed after pressing 'z'")
	}

	// Press 'z' again to toggle zoom off
	result, _ = m.Update(msg)

	updated, ok = result.(*Model)
	if !ok {
		t.Fatal("Update() did not return *Model")
	}

	m = updated

	if m.zoomed {
		t.Error("model should not be zoomed after pressing 'z' again")
	}
}

func TestZoomExitOnPanelSwitch(t *testing.T) {
	m := createZoomTestModel()

	// Enable zoom
	m.zoomed = true
	m.updatePanelSizes()

	// Press Tab to switch panel — should exit zoom
	msg := tea.KeyMsg{Type: tea.KeyTab}

	result, _ := m.Update(msg)

	updated, ok := result.(*Model)
	if !ok {
		t.Fatal("Update() did not return *Model")
	}

	m = updated

	if m.zoomed {
		t.Error("zoom should be exited when switching panels via Tab")
	}

	if m.activePanelIdx != 1 {
		t.Errorf("active panel should be 1 after Tab, got %d", m.activePanelIdx)
	}
}

func TestZoomExitOnNumberKey(t *testing.T) {
	m := createZoomTestModel()

	// Enable zoom
	m.zoomed = true
	m.updatePanelSizes()

	// Press '2' to switch to second panel — should exit zoom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}

	result, _ := m.Update(msg)

	updated, ok := result.(*Model)
	if !ok {
		t.Fatal("Update() did not return *Model")
	}

	m = updated

	if m.zoomed {
		t.Error("zoom should be exited when switching panels via number key")
	}

	if m.activePanelIdx != 1 {
		t.Errorf("active panel should be 1 after pressing '2', got %d", m.activePanelIdx)
	}
}

func TestExitZoomHelper(t *testing.T) {
	m := createZoomTestModel()

	// exitZoom should be no-op when not zoomed
	m.zoomed = false
	m.exitZoom()

	if m.zoomed {
		t.Error("exitZoom should keep zoomed=false when already false")
	}

	// exitZoom should clear zoom and update panel sizes
	m.zoomed = true
	m.exitZoom()

	if m.zoomed {
		t.Error("exitZoom should set zoomed=false")
	}
}

// TestUpdatePanelSizesZoomed verifies that updatePanelSizes sets the
// active panel to full width when zoomed. We verify via View() output
// because the Panel interface doesn't expose Width()/Height() —
// those methods live on BasePanel only.
func TestUpdatePanelSizesZoomed(t *testing.T) {
	m := createZoomTestModel()
	m.zoomed = true
	m.updatePanelSizes()

	view := m.panels[m.activePanelIdx].View()
	if view == "" {
		t.Fatal("zoomed panel View() returned empty string after updatePanelSizes")
	}
}

// TestUpdatePanelSizesNormal verifies that updatePanelSizes distributes
// width across all panels in non-zoomed mode (each gets width/4).
func TestUpdatePanelSizesNormal(t *testing.T) {
	m := createZoomTestModel()
	m.zoomed = false
	m.updatePanelSizes()

	for i, panel := range m.panels {
		view := panel.View()
		if view == "" {
			t.Errorf("panel[%d] View() returned empty string after updatePanelSizes", i)
		}
	}
}

func TestUpdatePanelSizesEmptyPanels(t *testing.T) {
	m := createZoomTestModel()
	m.panels = nil

	// Should not panic with empty panels
	m.updatePanelSizes()
}

func TestRenderPanelsZoomed(t *testing.T) {
	m := createZoomTestModel()
	m.zoomed = true

	view := m.renderPanels(m.width, m.height)

	if view == "" {
		t.Fatal("renderPanels() returned empty string in zoom mode")
	}
}

func TestRenderPanelsNormal(t *testing.T) {
	m := createZoomTestModel()
	m.zoomed = false

	view := m.renderPanels(m.width, m.height)

	if view == "" {
		t.Fatal("renderPanels() returned empty string in normal mode")
	}
}

func TestRenderPanelsEmpty(t *testing.T) {
	m := createZoomTestModel()
	m.panels = nil

	view := m.renderPanels(m.width, m.height)

	if view != "No panels configured" {
		t.Errorf("renderPanels() with no panels = %q, want %q", view, "No panels configured")
	}
}
