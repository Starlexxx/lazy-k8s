package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/lazyk8s/lazy-k8s/internal/config"
)

func createTestConfig() *config.ThemeConfig {
	return &config.ThemeConfig{
		PrimaryColor:    "#7aa2f7",
		SecondaryColor:  "#9ece6a",
		ErrorColor:      "#f7768e",
		WarningColor:    "#e0af68",
		BackgroundColor: "#1a1b26",
		TextColor:       "#c0caf5",
		BorderColor:     "#3b4261",
	}
}

func TestNewStyles(t *testing.T) {
	cfg := createTestConfig()

	styles := NewStyles(cfg)

	if styles == nil {
		t.Fatal("NewStyles returned nil")
	}

	// Verify colors are set correctly
	if styles.Primary != lipgloss.Color("#7aa2f7") {
		t.Errorf("Primary color = %v, want #7aa2f7", styles.Primary)
	}

	if styles.Secondary != lipgloss.Color("#9ece6a") {
		t.Errorf("Secondary color = %v, want #9ece6a", styles.Secondary)
	}

	if styles.Error != lipgloss.Color("#f7768e") {
		t.Errorf("Error color = %v, want #f7768e", styles.Error)
	}

	if styles.Warning != lipgloss.Color("#e0af68") {
		t.Errorf("Warning color = %v, want #e0af68", styles.Warning)
	}

	if styles.Background != lipgloss.Color("#1a1b26") {
		t.Errorf("Background color = %v, want #1a1b26", styles.Background)
	}

	if styles.Text != lipgloss.Color("#c0caf5") {
		t.Errorf("Text color = %v, want #c0caf5", styles.Text)
	}

	if styles.Border != lipgloss.Color("#3b4261") {
		t.Errorf("Border color = %v, want #3b4261", styles.Border)
	}
}

func TestStylesNotNil(t *testing.T) {
	cfg := createTestConfig()
	styles := NewStyles(cfg)

	// Verify all style fields are initialized
	styleFields := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Muted", styles.Muted},
		{"App", styles.App},
		{"Header", styles.Header},
		{"HeaderTitle", styles.HeaderTitle},
		{"HeaderContext", styles.HeaderContext},
		{"HeaderNamespace", styles.HeaderNamespace},
		{"HeaderHelp", styles.HeaderHelp},
		{"Panel", styles.Panel},
		{"PanelFocused", styles.PanelFocused},
		{"PanelTitle", styles.PanelTitle},
		{"PanelTitleActive", styles.PanelTitleActive},
		{"ListItem", styles.ListItem},
		{"ListItemSelected", styles.ListItemSelected},
		{"ListItemFocused", styles.ListItemFocused},
		{"StatusBar", styles.StatusBar},
		{"StatusKey", styles.StatusKey},
		{"StatusValue", styles.StatusValue},
		{"StatusError", styles.StatusError},
		{"StatusSuccess", styles.StatusSuccess},
		{"StatusWarning", styles.StatusWarning},
		{"TableHeader", styles.TableHeader},
		{"TableRow", styles.TableRow},
		{"TableCell", styles.TableCell},
		{"DetailTitle", styles.DetailTitle},
		{"DetailLabel", styles.DetailLabel},
		{"DetailValue", styles.DetailValue},
		{"Modal", styles.Modal},
		{"ModalTitle", styles.ModalTitle},
		{"ModalButton", styles.ModalButton},
		{"Input", styles.Input},
		{"InputPrompt", styles.InputPrompt},
		{"StatusRunning", styles.StatusRunning},
		{"StatusPending", styles.StatusPending},
		{"StatusFailed", styles.StatusFailed},
		{"StatusSucceeded", styles.StatusSucceeded},
		{"StatusUnknown", styles.StatusUnknown},
		{"StatusTerminating", styles.StatusTerminating},
	}

	for _, sf := range styleFields {
		// Just verify they can render without panic
		rendered := sf.style.Render("test")
		if rendered == "" {
			t.Errorf("%s style rendered empty string", sf.name)
		}
	}
}

func TestGetStatusStyle(t *testing.T) {
	cfg := createTestConfig()
	styles := NewStyles(cfg)

	tests := []struct {
		status   string
		expected lipgloss.Style
	}{
		{"Running", styles.StatusRunning},
		{"Active", styles.StatusRunning},
		{"Ready", styles.StatusRunning},
		{"Bound", styles.StatusRunning},
		{"Pending", styles.StatusPending},
		{"ContainerCreating", styles.StatusPending},
		{"PodInitializing", styles.StatusPending},
		{"Failed", styles.StatusFailed},
		{"Error", styles.StatusFailed},
		{"CrashLoopBackOff", styles.StatusFailed},
		{"ImagePullBackOff", styles.StatusFailed},
		{"ErrImagePull", styles.StatusFailed},
		{"Succeeded", styles.StatusSucceeded},
		{"Completed", styles.StatusSucceeded},
		{"Terminating", styles.StatusTerminating},
		{"Unknown", styles.StatusUnknown},
		{"SomeRandomStatus", styles.StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := styles.GetStatusStyle(tt.status)
			// Compare rendered output since lipgloss.Style comparison is tricky
			if result.Render("test") != tt.expected.Render("test") {
				t.Errorf("GetStatusStyle(%q) rendered differently than expected", tt.status)
			}
		})
	}
}

func TestStylesCanRenderText(t *testing.T) {
	cfg := createTestConfig()
	styles := NewStyles(cfg)

	// Test that various styles can render text without panic
	testText := "Test Text"

	// Header styles
	_ = styles.Header.Render(testText)
	_ = styles.HeaderTitle.Render(testText)
	_ = styles.HeaderContext.Render(testText)
	_ = styles.HeaderNamespace.Render(testText)
	_ = styles.HeaderHelp.Render(testText)

	// Panel styles
	_ = styles.Panel.Render(testText)
	_ = styles.PanelFocused.Render(testText)
	_ = styles.PanelTitle.Render(testText)
	_ = styles.PanelTitleActive.Render(testText)

	// List styles
	_ = styles.ListItem.Render(testText)
	_ = styles.ListItemSelected.Render(testText)
	_ = styles.ListItemFocused.Render(testText)

	// Status bar styles
	_ = styles.StatusBar.Render(testText)
	_ = styles.StatusKey.Render(testText)
	_ = styles.StatusValue.Render(testText)
	_ = styles.StatusError.Render(testText)

	// Modal styles
	_ = styles.Modal.Render(testText)
	_ = styles.ModalTitle.Render(testText)
	_ = styles.ModalButton.Render(testText)

	// If we got here without panic, test passes
}

func TestStylesWithDifferentColors(t *testing.T) {
	// Test with different color schemes
	configs := []*config.ThemeConfig{
		{
			PrimaryColor:    "#ff0000",
			SecondaryColor:  "#00ff00",
			ErrorColor:      "#0000ff",
			WarningColor:    "#ffff00",
			BackgroundColor: "#000000",
			TextColor:       "#ffffff",
			BorderColor:     "#888888",
		},
		{
			PrimaryColor:    "#ffffff",
			SecondaryColor:  "#ffffff",
			ErrorColor:      "#ffffff",
			WarningColor:    "#ffffff",
			BackgroundColor: "#ffffff",
			TextColor:       "#000000",
			BorderColor:     "#000000",
		},
	}

	for i, cfg := range configs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			styles := NewStyles(cfg)
			if styles == nil {
				t.Fatal("NewStyles returned nil")
			}

			// Verify can render without panic
			_ = styles.Header.Render("test")
			_ = styles.Panel.Render("test")
			_ = styles.StatusBar.Render("test")
		})
	}
}

func TestStylesSuccess(t *testing.T) {
	cfg := createTestConfig()
	styles := NewStyles(cfg)

	// Success should be the same as Secondary color
	if styles.Success != styles.Secondary {
		t.Errorf("Success color should equal Secondary color")
	}
}

func TestMutedColor(t *testing.T) {
	cfg := createTestConfig()
	styles := NewStyles(cfg)

	// MutedColor should be set to a specific value
	if styles.MutedColor != lipgloss.Color("#565f89") {
		t.Errorf("MutedColor = %v, want #565f89", styles.MutedColor)
	}
}
