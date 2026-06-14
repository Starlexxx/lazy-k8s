package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
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
			if p.Cursor() < len(p.filtered) {
				ns := p.filtered[p.Cursor()]
				p.client.SetNamespace(ns.Name)

				return p, func() tea.Msg {
					return StatusWithRefreshMsg{
						Message: fmt.Sprintf("Switched to namespace: %s", ns.Name),
					}
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

	title := p.renderTitle()
	if p.focused {
		b.WriteString(p.styles.PanelTitleActive.Render(title))
	} else {
		b.WriteString(p.styles.PanelTitle.Render(title))
	}

	b.WriteString("\n")

	startIdx, endIdx := p.visibleWindow(len(p.filtered), 0)

	for i := startIdx; i < endIdx; i++ {
		ns := p.filtered[i]
		line := p.renderNamespaceLine(ns, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *NamespacesPanel) renderNamespaceLine(ns corev1.Namespace, selected bool) string {
	status := string(ns.Status.Phase)

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	if p.width > 80 {
		nameW := max(p.width-25, 10)

		line += utils.PadRight(
			utils.Truncate(ns.Name, nameW), nameW,
		)

		statusStyle := p.styles.GetStatusStyle(status)
		line += " " + statusStyle.Render(utils.PadRight(status, 8))

		age := utils.FormatAgeFromMeta(ns.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)
	} else {
		name := utils.Truncate(ns.Name, p.width-6)
		line += name
		line = utils.PadRight(line, p.width-10)

		statusStyle := p.styles.GetStatusStyle(status)
		line += " " + statusStyle.Render(status)
	}

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

	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(string(ns.Status.Phase)).Render(string(ns.Status.Phase)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(ns.CreationTimestamp)))
	b.WriteString("\n")

	if len(ns.Labels) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Labels:"))
		b.WriteString("\n")

		for k, v := range ns.Labels {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

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

func (p *NamespacesPanel) SelectedItem() any {
	item := selectedItem(p.filtered, p.cursor)
	if item == nil {
		return nil
	}

	return item
}

func (p *NamespacesPanel) SelectedName() string {
	return selectedName(p.filtered, p.cursor, func(ns corev1.Namespace) string { return ns.Name })
}

func (p *NamespacesPanel) GetSelectedYAML() (string, error) {
	return marshalSelectedYAML(p.filtered, p.cursor)
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
	p.filtered = filterByName(
		p.namespaces, p.filter, func(ns corev1.Namespace) string { return ns.Name }, &p.cursor,
	)
}

func (p *NamespacesPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

func (p *NamespacesPanel) SearchItems(query string) []SearchResult {
	// Namespaces are cluster-scoped: pass nil for the namespace callback.
	return searchByName(
		p.namespaces,
		query,
		p.title,
		func(ns corev1.Namespace) string { return ns.Name },
		nil,
		func(ns corev1.Namespace) string { return string(ns.Status.Phase) },
	)
}

func (p *NamespacesPanel) NavigateTo(name, _ string) bool {
	// Namespaces are cluster-scoped; namespace argument is always ignored.
	return navigateTo(
		p.filtered,
		&p.cursor,
		func(ns corev1.Namespace) string { return ns.Name },
		nil,
		name,
		"",
	)
}

type namespacesLoadedMsg struct {
	namespaces []corev1.Namespace
}
