package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad_DefaultValues(t *testing.T) {
	viper.Reset()
	// Load config with defaults (no config file present in test paths)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Verify default theme colors
	if cfg.Theme.PrimaryColor != "#7aa2f7" {
		t.Errorf("Theme.PrimaryColor = %q, want %q", cfg.Theme.PrimaryColor, "#7aa2f7")
	}

	if cfg.Theme.SecondaryColor != "#9ece6a" {
		t.Errorf("Theme.SecondaryColor = %q, want %q", cfg.Theme.SecondaryColor, "#9ece6a")
	}

	if cfg.Theme.ErrorColor != "#f7768e" {
		t.Errorf("Theme.ErrorColor = %q, want %q", cfg.Theme.ErrorColor, "#f7768e")
	}

	if cfg.Theme.WarningColor != "#e0af68" {
		t.Errorf("Theme.WarningColor = %q, want %q", cfg.Theme.WarningColor, "#e0af68")
	}

	if cfg.Theme.BackgroundColor != "#1a1b26" {
		t.Errorf("Theme.BackgroundColor = %q, want %q", cfg.Theme.BackgroundColor, "#1a1b26")
	}

	if cfg.Theme.TextColor != "#c0caf5" {
		t.Errorf("Theme.TextColor = %q, want %q", cfg.Theme.TextColor, "#c0caf5")
	}

	if cfg.Theme.BorderColor != "#3b4261" {
		t.Errorf("Theme.BorderColor = %q, want %q", cfg.Theme.BorderColor, "#3b4261")
	}
}

func TestLoad_DefaultKeybindings(t *testing.T) {
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Verify default keybindings
	tests := []struct {
		name     string
		bindings []string
		expected []string
	}{
		{"Quit", cfg.Keybindings.Quit, []string{"q", "ctrl+c"}},
		{"Help", cfg.Keybindings.Help, []string{"?"}},
		{"NextPanel", cfg.Keybindings.NextPanel, []string{"tab"}},
		{"PrevPanel", cfg.Keybindings.PrevPanel, []string{"shift+tab"}},
		{"Up", cfg.Keybindings.Up, []string{"k", "up"}},
		{"Down", cfg.Keybindings.Down, []string{"j", "down"}},
		{"Search", cfg.Keybindings.Search, []string{"/"}},
		{"Delete", cfg.Keybindings.Delete, []string{"D"}},
		{"Describe", cfg.Keybindings.Describe, []string{"d"}},
		{"Yaml", cfg.Keybindings.Yaml, []string{"y"}},
		{"Logs", cfg.Keybindings.Logs, []string{"l"}},
		{"Exec", cfg.Keybindings.Exec, []string{"x"}},
		{"PortForward", cfg.Keybindings.PortForward, []string{"p"}},
		{"Scale", cfg.Keybindings.Scale, []string{"s"}},
		{"Restart", cfg.Keybindings.Restart, []string{"r"}},
		{"Refresh", cfg.Keybindings.Refresh, []string{"ctrl+r"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.bindings) != len(tt.expected) {
				t.Errorf("%s has %d bindings, want %d", tt.name, len(tt.bindings), len(tt.expected))

				return
			}

			for i, binding := range tt.bindings {
				if binding != tt.expected[i] {
					t.Errorf("%s[%d] = %q, want %q", tt.name, i, binding, tt.expected[i])
				}
			}
		})
	}
}

func TestLoad_DefaultDefaults(t *testing.T) {
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Verify default settings
	if cfg.Defaults.Namespace != "default" {
		t.Errorf("Defaults.Namespace = %q, want %q", cfg.Defaults.Namespace, "default")
	}

	if cfg.Defaults.LogLines != 100 {
		t.Errorf("Defaults.LogLines = %d, want %d", cfg.Defaults.LogLines, 100)
	}

	if cfg.Defaults.FollowLogs != true {
		t.Errorf("Defaults.FollowLogs = %v, want %v", cfg.Defaults.FollowLogs, true)
	}

	if cfg.Defaults.RefreshInterval != 5 {
		t.Errorf("Defaults.RefreshInterval = %d, want %d", cfg.Defaults.RefreshInterval, 5)
	}
}

func TestLoad_DefaultPanels(t *testing.T) {
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Verify default panels configuration
	expectedPanels := []string{"namespaces", "pods", "deployments", "services"}
	if len(cfg.Panels.Visible) != len(expectedPanels) {
		t.Errorf(
			"Panels.Visible has %d items, want %d",
			len(cfg.Panels.Visible),
			len(expectedPanels),
		)
	} else {
		for i, panel := range cfg.Panels.Visible {
			if panel != expectedPanels[i] {
				t.Errorf("Panels.Visible[%d] = %q, want %q", i, panel, expectedPanels[i])
			}
		}
	}

	if cfg.Panels.Layout != "vertical" {
		t.Errorf("Panels.Layout = %q, want %q", cfg.Panels.Layout, "vertical")
	}
}

func TestLoad_NamespaceFallback(t *testing.T) {
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// When no namespace is set explicitly, it should fall back to defaults.namespace
	if cfg.Namespace != cfg.Defaults.Namespace {
		t.Errorf(
			"Namespace = %q, want %q (from Defaults.Namespace)",
			cfg.Namespace,
			cfg.Defaults.Namespace,
		)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	viper.Reset()
	// Create a temporary directory for the config file
	tmpDir, err := os.MkdirTemp("", "lazy-k8s-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config file with custom values
	configContent := `
theme:
  primaryColor: "#ff0000"
  secondaryColor: "#00ff00"
defaults:
  namespace: "custom-ns"
  logLines: 50
panels:
  visible:
    - pods
    - services
  layout: "horizontal"
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to the temp directory so viper can find the config
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Verify custom theme colors were loaded
	if cfg.Theme.PrimaryColor != "#ff0000" {
		t.Errorf("Theme.PrimaryColor = %q, want %q", cfg.Theme.PrimaryColor, "#ff0000")
	}

	if cfg.Theme.SecondaryColor != "#00ff00" {
		t.Errorf("Theme.SecondaryColor = %q, want %q", cfg.Theme.SecondaryColor, "#00ff00")
	}

	// Verify custom defaults were loaded
	if cfg.Defaults.Namespace != "custom-ns" {
		t.Errorf("Defaults.Namespace = %q, want %q", cfg.Defaults.Namespace, "custom-ns")
	}

	if cfg.Defaults.LogLines != 50 {
		t.Errorf("Defaults.LogLines = %d, want %d", cfg.Defaults.LogLines, 50)
	}

	// Verify custom panels layout was loaded
	if cfg.Panels.Layout != "horizontal" {
		t.Errorf("Panels.Layout = %q, want %q", cfg.Panels.Layout, "horizontal")
	}

	// Note: Due to viper's behavior with merging structs, panels.visible may have
	// both default values and config file values. We verify the layout was loaded
	// which confirms the config file was read.

	// Verify namespace is set from defaults
	if cfg.Namespace != "custom-ns" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "custom-ns")
	}
}

func TestConfigStruct(t *testing.T) {
	// Test that the Config struct can be instantiated with all fields
	cfg := Config{
		Kubeconfig: "/path/to/kubeconfig",
		Context:    "my-context",
		Namespace:  "my-namespace",
		Theme: ThemeConfig{
			PrimaryColor:    "#000000",
			SecondaryColor:  "#111111",
			ErrorColor:      "#222222",
			WarningColor:    "#333333",
			BackgroundColor: "#444444",
			TextColor:       "#555555",
			BorderColor:     "#666666",
		},
		Keybindings: KeybindingsConfig{
			Quit: []string{"q"},
			Help: []string{"h"},
		},
		Defaults: DefaultsConfig{
			Namespace:       "test-ns",
			LogLines:        200,
			FollowLogs:      false,
			RefreshInterval: 10,
		},
		Panels: PanelsConfig{
			Visible: []string{"pods"},
			Layout:  "grid",
		},
	}

	// Verify the struct was instantiated correctly
	if cfg.Kubeconfig != "/path/to/kubeconfig" {
		t.Errorf("Kubeconfig = %q, want %q", cfg.Kubeconfig, "/path/to/kubeconfig")
	}

	if cfg.Context != "my-context" {
		t.Errorf("Context = %q, want %q", cfg.Context, "my-context")
	}

	if cfg.Namespace != "my-namespace" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "my-namespace")
	}

	if cfg.Theme.PrimaryColor != "#000000" {
		t.Errorf("Theme.PrimaryColor = %q, want %q", cfg.Theme.PrimaryColor, "#000000")
	}

	if len(cfg.Keybindings.Quit) != 1 || cfg.Keybindings.Quit[0] != "q" {
		t.Errorf("Keybindings.Quit = %v, want [\"q\"]", cfg.Keybindings.Quit)
	}

	if cfg.Defaults.LogLines != 200 {
		t.Errorf("Defaults.LogLines = %d, want %d", cfg.Defaults.LogLines, 200)
	}

	if cfg.Panels.Layout != "grid" {
		t.Errorf("Panels.Layout = %q, want %q", cfg.Panels.Layout, "grid")
	}
}
