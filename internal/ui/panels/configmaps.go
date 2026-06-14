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

type ConfigMapsPanel struct {
	BasePanel
	client     *k8s.Client
	styles     *theme.Styles
	configmaps []corev1.ConfigMap
	filtered   []corev1.ConfigMap
}

func NewConfigMapsPanel(client *k8s.Client, styles *theme.Styles) *ConfigMapsPanel {
	return &ConfigMapsPanel{
		BasePanel: BasePanel{
			title:       "ConfigMaps",
			shortcutKey: "5",
		},
		client: client,
		styles: styles,
	}
}

func (p *ConfigMapsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *ConfigMapsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
		}

	case configmapsLoadedMsg:
		p.configmaps = msg.configmaps
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *ConfigMapsPanel) View() string {
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
		cm := p.filtered[i]
		line := p.renderConfigMapLine(cm, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *ConfigMapsPanel) renderConfigMapLine(cm corev1.ConfigMap, selected bool) string {
	dataCount := fmt.Sprintf("%d", len(cm.Data))

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	if p.width > 80 {
		reserved := 22
		if p.width > 120 && p.allNs {
			reserved += 16
		}

		nameW := max(p.width-reserved, 10)

		line += utils.PadRight(
			utils.Truncate(cm.Name, nameW), nameW,
		)
		line += " " + utils.PadRight(dataCount, 5)

		age := utils.FormatAgeFromMeta(cm.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)

		if p.width > 120 && p.allNs {
			line += " " + utils.Truncate(cm.Namespace, 15)
		}
	} else {
		name := utils.Truncate(cm.Name, p.width-10)
		line += name
		line = utils.PadRight(line, p.width-6)
		line += " " + dataCount
	}

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *ConfigMapsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No configmap selected"
	}

	cm := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("ConfigMap: " + cm.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Data Keys:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", len(cm.Data))))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(cm.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(cm.Namespace))
		b.WriteString("\n")
	}

	if len(cm.Data) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Data:"))
		b.WriteString("\n")

		for k, v := range cm.Data {
			preview := utils.Truncate(strings.ReplaceAll(v, "\n", "\\n"), 50)
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, preview))
		}
	}

	if len(cm.BinaryData) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Binary Data:"))
		b.WriteString("\n")

		for k := range cm.BinaryData {
			b.WriteString(fmt.Sprintf("  %s: <binary>\n", k))
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [e]dit [D]elete"))

	return b.String()
}

func (p *ConfigMapsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			configmaps []corev1.ConfigMap
			err        error
		)

		if p.allNs {
			configmaps, err = p.client.ListConfigMapsAllNamespaces(ctx)
		} else {
			configmaps, err = p.client.ListConfigMaps(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return configmapsLoadedMsg{configmaps: configmaps}
	}
}

func (p *ConfigMapsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	cm := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteConfigMap(ctx, cm.Namespace, cm.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted configmap: %s", cm.Name)}
	}
}

func (p *ConfigMapsPanel) SelectedItem() any {
	item := selectedItem(p.filtered, p.cursor)
	if item == nil {
		return nil
	}

	return item
}

func (p *ConfigMapsPanel) SelectedName() string {
	return selectedName(p.filtered, p.cursor, func(cm corev1.ConfigMap) string { return cm.Name })
}

func (p *ConfigMapsPanel) GetSelectedYAML() (string, error) {
	return marshalSelectedYAML(p.filtered, p.cursor)
}

func (p *ConfigMapsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	cm := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:         %s\n", cm.Name))
	b.WriteString(fmt.Sprintf("Namespace:    %s\n", cm.Namespace))
	b.WriteString(fmt.Sprintf("Age:          %s\n", utils.FormatAgeFromMeta(cm.CreationTimestamp)))

	if len(cm.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range cm.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(cm.Data) > 0 {
		b.WriteString("\nData:\n")

		for k, v := range cm.Data {
			b.WriteString(fmt.Sprintf("====\n%s:\n----\n%s\n", k, v))
		}
	}

	return b.String(), nil
}

func (p *ConfigMapsPanel) applyFilter() {
	p.filtered = filterByName(
		p.configmaps, p.filter, func(cm corev1.ConfigMap) string { return cm.Name }, &p.cursor,
	)
}

func (p *ConfigMapsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

func (p *ConfigMapsPanel) SearchItems(query string) []SearchResult {
	return searchByName(
		p.configmaps,
		query,
		p.title,
		func(cm corev1.ConfigMap) string { return cm.Name },
		func(cm corev1.ConfigMap) string { return cm.Namespace },
		func(cm corev1.ConfigMap) string { return fmt.Sprintf("%d keys", len(cm.Data)) },
	)
}

func (p *ConfigMapsPanel) NavigateTo(name, namespace string) bool {
	return navigateTo(
		p.filtered,
		&p.cursor,
		func(cm corev1.ConfigMap) string { return cm.Name },
		func(cm corev1.ConfigMap) string { return cm.Namespace },
		name,
		namespace,
	)
}

type configmapsLoadedMsg struct {
	configmaps []corev1.ConfigMap
}
