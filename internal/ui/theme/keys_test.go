package theme

import (
	"slices"
	"testing"
)

func TestNewKeyMap(t *testing.T) {
	keys := NewKeyMap()

	if keys == nil {
		t.Fatal("NewKeyMap returned nil")
	}

	// Verify critical bindings are set
	bindings := []struct {
		name string
		keys []string
	}{
		{"Quit", keys.Quit.Keys()},
		{"Help", keys.Help.Keys()},
		{"Search", keys.Search.Keys()},
		{"Diff", keys.Diff.Keys()},
		{"Rollback", keys.Rollback.Keys()},
		{"Scale", keys.Scale.Keys()},
	}

	for _, b := range bindings {
		if len(b.keys) == 0 {
			t.Errorf("%s binding has no keys", b.name)
		}
	}
}

func TestNewKeyMapDiffBinding(t *testing.T) {
	keys := NewKeyMap()

	diffKeys := keys.Diff.Keys()

	if !slices.Contains(diffKeys, "V") {
		t.Errorf("Diff binding should contain 'V', got %v", diffKeys)
	}
}

func TestShortHelp(t *testing.T) {
	keys := NewKeyMap()

	bindings := keys.ShortHelp()

	if len(bindings) == 0 {
		t.Error("ShortHelp should return bindings")
	}
}

func TestFullHelp(t *testing.T) {
	keys := NewKeyMap()

	groups := keys.FullHelp()

	if len(groups) == 0 {
		t.Error("FullHelp should return binding groups")
	}

	// Verify Diff is included in one of the groups
	found := false

	for _, group := range groups {
		for _, binding := range group {
			if slices.Contains(binding.Keys(), "V") {
				found = true
			}
		}
	}

	if !found {
		t.Error("FullHelp should include the Diff (V) binding")
	}
}
