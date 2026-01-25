package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type NamespacesPanel struct {
	BasePanel
	client     *k8s.Client
	styles     *theme.Styles
	namespaces []corev1.Namespace
	filtered   []corev1.Namespace
}

func NewNamespacesPanel(client *k8s.Client, styles *theme.Styles) *NamespacesPanel {
	return &NamespacesPanel{
		BasePanel: BasePanel{
			title:       "Namespaces",
			shortcutKey: "1",
		},
		client: client,
		styles: styles,
	}
}

func (p *NamespacesPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *NamespacesPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			p.MoveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			p.MoveDown(len(p.filtered))
		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			p.MoveToTop()
		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			p.MoveToBottom(len(p.filtered))
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Switch to selected namespace
			if p.Cursor() < len(p.filtered) {
				ns := p.filtered[p.Cursor()]
				p.client.SetNamespace(ns.Name)

				return p, func() tea.Msg {
					return StatusMsg{Message: fmt.Sprintf("Switched to namespace: %s", ns.Name)}
				}
			}
		}

	case namespacesLoadedMsg:
		p.namespaces = msg.namespaces
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *NamespacesPanel) View() string {
	var b strings.Builder

	// Title
	title := fmt.Sprintf("%s [%s]", p.title, p.shortcutKey)
	if p.focused {
		b.WriteString(p.styles.PanelTitleActive.Render(title))
	} else {
		b.WriteString(p.styles.PanelTitle.Render(title))
	}

	b.WriteString("\n")

	// Calculate visible items
	visibleHeight := p.height - 3 // title + border
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Determine scroll position
	startIdx := 0
	if p.cursor >= visibleHeight {
		startIdx = p.cursor - visibleHeight + 1
	}

	endIdx := startIdx + visibleHeight
	if endIdx > len(p.filtered) {
		endIdx = len(p.filtered)
	}

	// Render items
	for i := startIdx; i < endIdx; i++ {
		ns := p.filtered[i]
		line := p.renderNamespaceLine(ns, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Apply panel style
	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *NamespacesPanel) renderNamespaceLine(ns corev1.Namespace, selected bool) string {
	name := utils.Truncate(ns.Name, p.width-6)
	status := string(ns.Status.Phase)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	// Pad and add status indicator
	line = utils.PadRight(line, p.width-10)

	statusStyle := p.styles.GetStatusStyle(status)
	line += " " + statusStyle.Render(status)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *NamespacesPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No namespace selected"
	}

	ns := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Namespace: " + ns.Name))
	b.WriteString("\n\n")

	// Status
	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(string(ns.Status.Phase)).Render(string(ns.Status.Phase)))
	b.WriteString("\n")

	// Age
	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(ns.CreationTimestamp)))
	b.WriteString("\n")

	// Labels
	if len(ns.Labels) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Labels:"))
		b.WriteString("\n")

		for k, v := range ns.Labels {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	// Annotations
	if len(ns.Annotations) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Annotations:"))
		b.WriteString("\n")

		for k, v := range ns.Annotations {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, utils.Truncate(v, width-len(k)-6)))
		}
	}

	return b.String()
}

func (p *NamespacesPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		namespaces, err := p.client.ListNamespaces(ctx)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return namespacesLoadedMsg{namespaces: namespaces}
	}
}

func (p *NamespacesPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	ns := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteNamespace(ctx, ns.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted namespace: %s", ns.Name)}
	}
}

func (p *NamespacesPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *NamespacesPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *NamespacesPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	ns := p.filtered[p.cursor]

	data, err := yaml.Marshal(ns)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *NamespacesPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	ns := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:         %s\n", ns.Name))
	b.WriteString(fmt.Sprintf("Status:       %s\n", ns.Status.Phase))
	b.WriteString(fmt.Sprintf("Age:          %s\n", utils.FormatAgeFromMeta(ns.CreationTimestamp)))

	if len(ns.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range ns.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(ns.Annotations) > 0 {
		b.WriteString("\nAnnotations:\n")

		for k, v := range ns.Annotations {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	return b.String(), nil
}

func (p *NamespacesPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.namespaces

		return
	}

	p.filtered = make([]corev1.Namespace, 0)
	for _, ns := range p.namespaces {
		if strings.Contains(strings.ToLower(ns.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, ns)
		}
	}

	// Reset cursor if out of bounds
	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *NamespacesPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type namespacesLoadedMsg struct {
	namespaces []corev1.Namespace
}
