package components

import (
	"strings"
	"testing"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

func TestHelpViewContainsVersionDiff(t *testing.T) {
	styles := createTestStyles()
	keys := theme.NewKeyMap()

	help := NewHelp(styles, keys)

	view := help.View(100, 50)

	if !strings.Contains(view, "Version diff") {
		t.Error("Help view should contain 'Version diff' binding")
	}
}

func TestHelpViewContainsDeploymentActions(t *testing.T) {
	styles := createTestStyles()
	keys := theme.NewKeyMap()

	help := NewHelp(styles, keys)

	view := help.View(100, 50)

	if !strings.Contains(view, "Deployment Actions") {
		t.Error("Help view should contain 'Deployment Actions' section")
	}
}

func TestHelpViewNotEmpty(t *testing.T) {
	styles := createTestStyles()
	keys := theme.NewKeyMap()

	help := NewHelp(styles, keys)

	view := help.View(80, 40)

	if view == "" {
		t.Error("Help.View should not be empty")
	}
}
