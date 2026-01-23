package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type Header struct {
	styles    *theme.Styles
	context   string
	namespace string
}

func NewHeader(styles *theme.Styles, context, namespace string) *Header {
	return &Header{
		styles:    styles,
		context:   context,
		namespace: namespace,
	}
}

func (h *Header) SetContext(ctx string) {
	h.context = ctx
}

func (h *Header) SetNamespace(ns string) {
	h.namespace = ns
}

func (h *Header) View(width int) string {
	title := h.styles.HeaderTitle.Render("lazy-k8s")

	contextInfo := h.styles.HeaderContext.Render(fmt.Sprintf("Context: %s", h.context))
	nsInfo := h.styles.HeaderNamespace.Render(fmt.Sprintf("Namespace: %s", h.namespace))
	helpHint := h.styles.HeaderHelp.Render("? for help")

	// Calculate spacing
	leftPart := fmt.Sprintf("%s │ %s │ %s", title, contextInfo, nsInfo)
	rightPart := helpHint

	// Calculate padding
	padding := width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart) - 2
	if padding < 0 {
		padding = 1
	}

	spacer := lipgloss.NewStyle().Width(padding).Render("")

	return h.styles.Header.Width(width).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, leftPart, spacer, rightPart),
	)
}
