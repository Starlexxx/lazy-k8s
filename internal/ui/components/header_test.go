package components

import (
	"strings"
	"testing"

	"github.com/Starlexxx/lazy-k8s/internal/config"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

func createTestStyles() *theme.Styles {
	cfg := &config.ThemeConfig{
		PrimaryColor:    "#7aa2f7",
		SecondaryColor:  "#9ece6a",
		ErrorColor:      "#f7768e",
		WarningColor:    "#e0af68",
		BackgroundColor: "#1a1b26",
		TextColor:       "#c0caf5",
		BorderColor:     "#3b4261",
	}

	return theme.NewStyles(cfg)
}

func TestNewHeader(t *testing.T) {
	styles := createTestStyles()

	header := NewHeader(styles, "my-context", "my-namespace")

	if header == nil {
		t.Fatal("NewHeader returned nil")
	}

	if header.context != "my-context" {
		t.Errorf("header.context = %q, want %q", header.context, "my-context")
	}

	if header.namespace != "my-namespace" {
		t.Errorf("header.namespace = %q, want %q", header.namespace, "my-namespace")
	}
}

func TestHeaderSetContext(t *testing.T) {
	styles := createTestStyles()
	header := NewHeader(styles, "old-context", "ns")

	header.SetContext("new-context")

	if header.context != "new-context" {
		t.Errorf("header.context = %q, want %q", header.context, "new-context")
	}
}

func TestHeaderSetNamespace(t *testing.T) {
	styles := createTestStyles()
	header := NewHeader(styles, "ctx", "old-namespace")

	header.SetNamespace("new-namespace")

	if header.namespace != "new-namespace" {
		t.Errorf("header.namespace = %q, want %q", header.namespace, "new-namespace")
	}
}

func TestHeaderView(t *testing.T) {
	styles := createTestStyles()
	header := NewHeader(styles, "prod-cluster", "kube-system")

	view := header.View(100)

	if !strings.Contains(view, "lazy-k8s") {
		t.Error("Header view should contain 'lazy-k8s' title")
	}

	if !strings.Contains(view, "Context:") || !strings.Contains(view, "prod-cluster") {
		t.Error("Header view should contain context information")
	}

	if !strings.Contains(view, "Namespace:") || !strings.Contains(view, "kube-system") {
		t.Error("Header view should contain namespace information")
	}

	if !strings.Contains(view, "? for help") {
		t.Error("Header view should contain help hint")
	}
}

func TestHeaderViewNarrowWidth(t *testing.T) {
	styles := createTestStyles()
	header := NewHeader(styles, "context", "namespace")

	// Test with very narrow width (should not panic)
	view := header.View(20)

	if view == "" {
		t.Error("Header view should not be empty even with narrow width")
	}
}

func TestHeaderViewZeroWidth(t *testing.T) {
	styles := createTestStyles()
	header := NewHeader(styles, "context", "namespace")

	// Test with zero width (should not panic)
	view := header.View(0)

	if view == "" {
		t.Error("Header view should not be empty even with zero width")
	}
}

func TestHeaderViewLongValues(t *testing.T) {
	styles := createTestStyles()

	// Test with very long context and namespace names
	longContext := "very-long-context-name-that-is-really-quite-long"
	longNamespace := "very-long-namespace-name-that-is-also-quite-long"
	header := NewHeader(styles, longContext, longNamespace)

	view := header.View(80)

	// Should contain the values (possibly truncated in display)
	if !strings.Contains(view, "lazy-k8s") {
		t.Error("Header view should contain 'lazy-k8s' title")
	}
}
