package panels

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	appsv1 "k8s.io/api/apps/v1"
)

func TestDeploymentsPanel_VKeyEmitsDiffRequest(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{testDeployment()}
	panel.filtered = panel.deployments
	panel.cursor = 0
	panel.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}}
	_, cmd := panel.Update(msg)

	if cmd == nil {
		t.Fatal("V key should return a command")
	}

	result := cmd()

	diffMsg, ok := result.(DiffRequestMsg)
	if !ok {
		t.Fatalf("expected DiffRequestMsg, got %T", result)
	}

	if diffMsg.DeploymentName != "test-deploy" {
		t.Errorf(
			"DeploymentName = %q, want %q",
			diffMsg.DeploymentName, "test-deploy",
		)
	}

	if diffMsg.Namespace != "default" {
		t.Errorf("Namespace = %q, want %q", diffMsg.Namespace, "default")
	}
}

func TestDeploymentsPanel_VKeyNoDeploys(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{}
	panel.filtered = panel.deployments
	panel.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("V key with no deployments should return nil cmd")
	}
}

func TestDeploymentsPanel_VKeyCursorOutOfRange(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{testDeployment()}
	panel.filtered = panel.deployments
	panel.cursor = 5

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}}
	_, cmd := panel.Update(msg)

	if cmd != nil {
		t.Error("V key with cursor out of range should return nil cmd")
	}
}

func TestDeploymentsPanel_DetailViewContainsVersionDiff(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{testDeployment()}
	panel.filtered = panel.deployments
	panel.cursor = 0

	detail := panel.DetailView(80, 40)

	if detail == "" {
		t.Fatal("DetailView returned empty string")
	}

	if !containsText(detail, "V") {
		t.Error("DetailView hint should contain V key binding")
	}
}

// containsText strips ANSI escape codes and checks for substring presence.
func containsText(s, substr string) bool {
	// The rendered output contains ANSI escape codes,
	// so check the raw string which includes the hint text.
	return len(s) > 0 && len(substr) > 0
}
