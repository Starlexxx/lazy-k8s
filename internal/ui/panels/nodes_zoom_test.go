package panels

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNodesPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNodesPanel(client, styles)

	node := testNode()
	panel.nodes = []corev1.Node{node}
	panel.filtered = panel.nodes
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	// Narrow mode should not have table header columns
	if strings.Contains(view, "ROLES") {
		t.Error("narrow view should not contain ROLES header column")
	}

	if strings.Contains(view, "VERSION") {
		t.Error("narrow view should not contain VERSION header column")
	}
}

func TestNodesPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNodesPanel(client, styles)

	node := testNode()
	panel.nodes = []corev1.Node{node}
	panel.filtered = panel.nodes
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	expectedCols := []string{"NAME", "STATUS", "ROLES", "VERSION", "AGE"}
	for _, col := range expectedCols {
		if !strings.Contains(view, col) {
			t.Errorf("wide view should contain %q header column", col)
		}
	}

	if !strings.Contains(view, "test-node") {
		t.Error("wide view should contain node name")
	}
}

func TestNodesPanel_RenderNodeHeader(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNodesPanel(client, styles)
	panel.SetSize(150, 30)

	header := panel.renderNodeHeader()

	if header == "" {
		t.Fatal("renderNodeHeader() returned empty string")
	}

	expectedCols := []string{"NAME", "STATUS", "ROLES", "VERSION", "AGE"}
	for _, col := range expectedCols {
		if !strings.Contains(header, col) {
			t.Errorf("header should contain %q", col)
		}
	}
}

// TestNodesPanel_NodeNameWidth verifies the name column shrinks when
// metrics columns are added. The reserved space for status/roles/version/age
// is fixed, and metrics add CPU+MEM columns that reduce available name width.
func TestNodesPanel_NodeNameWidth(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNodesPanel(client, styles)
	panel.SetSize(150, 30)

	nameW := panel.nodeNameWidth(false)
	if nameW <= 0 {
		t.Errorf("nodeNameWidth(false) = %d, want > 0", nameW)
	}

	nameWMetrics := panel.nodeNameWidth(true)
	if nameWMetrics >= nameW {
		t.Errorf(
			"nodeNameWidth(true)=%d should be less than nodeNameWidth(false)=%d",
			nameWMetrics, nameW,
		)
	}
}

func TestNodesPanel_ViewEmpty(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNodesPanel(client, styles)
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string even with no nodes")
	}
}
