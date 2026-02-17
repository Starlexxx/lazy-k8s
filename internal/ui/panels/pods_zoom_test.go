package panels

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestPodsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)

	pod := testPod()
	panel.pods = []corev1.Pod{pod}
	panel.filtered = panel.pods
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if strings.Contains(view, "READY") {
		t.Error("narrow view should not contain READY header column")
	}

	if strings.Contains(view, "RESTARTS") {
		t.Error("narrow view should not contain RESTARTS header column")
	}

	if strings.Contains(view, "AGE") {
		t.Error("narrow view should not contain AGE header column")
	}
}

func TestPodsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)

	pod := testPod()
	panel.pods = []corev1.Pod{pod}
	panel.filtered = panel.pods
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	// Wide mode renders a header row with these columns
	if !strings.Contains(view, "NAME") {
		t.Error("wide view should contain NAME header column")
	}

	if !strings.Contains(view, "STATUS") {
		t.Error("wide view should contain STATUS header column")
	}

	if !strings.Contains(view, "READY") {
		t.Error("wide view should contain READY header column")
	}

	if !strings.Contains(view, "RESTARTS") {
		t.Error("wide view should contain RESTARTS header column")
	}

	if !strings.Contains(view, "AGE") {
		t.Error("wide view should contain AGE header column")
	}

	if !strings.Contains(view, "test-pod") {
		t.Error("wide view should contain pod name")
	}
}

// TestPodsPanel_ViewWideAllNs verifies the NAMESPACE column appears
// only when width > 120 AND allNs is enabled. Width=150 satisfies
// the >120 threshold that controls this extra column.
func TestPodsPanel_ViewWideAllNs(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)

	pod := testPod()
	panel.pods = []corev1.Pod{pod}
	panel.filtered = panel.pods
	panel.SetSize(150, 30)
	panel.SetFocused(true)
	panel.SetAllNamespaces(true)

	view := panel.View()

	if !strings.Contains(view, "NAMESPACE") {
		t.Error("wide allNs view should contain NAMESPACE header column")
	}
}

func TestPodsPanel_RenderPodHeader(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)
	panel.SetSize(150, 30)

	header := panel.renderPodHeader()

	if header == "" {
		t.Fatal("renderPodHeader() returned empty string")
	}

	expectedCols := []string{"NAME", "STATUS", "READY", "RESTARTS", "AGE"}
	for _, col := range expectedCols {
		if !strings.Contains(header, col) {
			t.Errorf("header should contain %q", col)
		}
	}
}

// TestPodsPanel_PodNameWidth verifies that the name column shrinks
// as more columns are added. The name column is dynamically sized:
// width minus reserved space for status/ready/restarts/age columns,
// with extra reservation when metrics or namespace columns are active.
func TestPodsPanel_PodNameWidth(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)

	panel.SetSize(150, 30)

	nameW := panel.podNameWidth(false)
	if nameW <= 0 {
		t.Errorf("podNameWidth(false) = %d, want > 0", nameW)
	}

	// Metrics add CPU+MEM columns, reducing available name space
	nameWMetrics := panel.podNameWidth(true)
	if nameWMetrics >= nameW {
		t.Errorf(
			"podNameWidth(true)=%d should be less than podNameWidth(false)=%d",
			nameWMetrics, nameW,
		)
	}

	// AllNs adds a NAMESPACE column, further reducing name space
	panel.SetAllNamespaces(true)

	nameWAllNs := panel.podNameWidth(false)
	if nameWAllNs >= nameW {
		t.Errorf(
			"podNameWidth with allNs=%d should be less than without=%d",
			nameWAllNs, nameW,
		)
	}
}

func TestPodsPanel_ViewEmpty(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string even with no pods")
	}
}
