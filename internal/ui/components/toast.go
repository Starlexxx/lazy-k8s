package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

type Toast struct {
	styles    *theme.Styles
	message   string
	toastType ToastType
	visible   bool
	duration  time.Duration
}

type ToastHideMsg struct{}

func NewToast(styles *theme.Styles) *Toast {
	return &Toast{
		styles:   styles,
		duration: 3 * time.Second,
	}
}

func (t *Toast) Show(message string, toastType ToastType) tea.Cmd {
	t.message = message
	t.toastType = toastType
	t.visible = true

	return tea.Tick(t.duration, func(time.Time) tea.Msg {
		return ToastHideMsg{}
	})
}

func (t *Toast) Hide() {
	t.visible = false
}

func (t *Toast) Update(msg tea.Msg) (*Toast, tea.Cmd) {
	switch msg.(type) {
	case ToastHideMsg:
		t.visible = false
	}

	return t, nil
}

func (t *Toast) View(width int) string {
	if !t.visible {
		return ""
	}

	var (
		style lipgloss.Style
		icon  string
	)

	switch t.toastType {
	case ToastSuccess:
		style = lipgloss.NewStyle().
			Background(t.styles.Success).
			Foreground(t.styles.Background).
			Padding(0, 2)
		icon = "✓ "
	case ToastWarning:
		style = lipgloss.NewStyle().
			Background(t.styles.Warning).
			Foreground(t.styles.Background).
			Padding(0, 2)
		icon = "⚠ "
	case ToastError:
		style = lipgloss.NewStyle().
			Background(t.styles.Error).
			Foreground(t.styles.Background).
			Padding(0, 2)
		icon = "✗ "
	case ToastInfo:
		style = lipgloss.NewStyle().
			Background(t.styles.Primary).
			Foreground(t.styles.Background).
			Padding(0, 2)
		icon = "ℹ "
	}

	content := style.Render(icon + t.message)

	// Position at bottom-right
	return lipgloss.Place(width, 1, lipgloss.Right, lipgloss.Bottom, content)
}

func (t *Toast) IsVisible() bool {
	return t.visible
}
