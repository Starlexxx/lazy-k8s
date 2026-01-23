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

type ServicesPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	services []corev1.Service
	filtered []corev1.Service
}

func NewServicesPanel(client *k8s.Client, styles *theme.Styles) *ServicesPanel {
	return &ServicesPanel{
		BasePanel: BasePanel{
			title:       "Services",
			shortcutKey: "4",
		},
		client: client,
		styles: styles,
	}
}

func (p *ServicesPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *ServicesPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case servicesLoadedMsg:
		p.services = msg.services
		p.applyFilter()
		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *ServicesPanel) View() string {
	var b strings.Builder

	// Title
	title := fmt.Sprintf("%s [%s]", p.title, p.shortcutKey)
	if p.focused {
		b.WriteString(p.styles.PanelTitleActive.Render(title))
	} else {
		b.WriteString(p.styles.PanelTitle.Render(title))
	}
	b.WriteString("\n")

	visibleHeight := p.height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	startIdx := 0
	if p.cursor >= visibleHeight {
		startIdx = p.cursor - visibleHeight + 1
	}

	endIdx := startIdx + visibleHeight
	if endIdx > len(p.filtered) {
		endIdx = len(p.filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		svc := p.filtered[i]
		line := p.renderServiceLine(svc, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *ServicesPanel) renderServiceLine(svc corev1.Service, selected bool) string {
	name := utils.Truncate(svc.Name, p.width-15)
	svcType := string(svc.Spec.Type)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-12)
	line += " " + utils.Truncate(svcType, 10)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}
	return p.styles.ListItem.Render(line)
}

func (p *ServicesPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No service selected"
	}

	svc := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Service: " + svc.Name))
	b.WriteString("\n\n")

	// Basic info
	b.WriteString(p.styles.DetailLabel.Render("Type:"))
	b.WriteString(p.styles.DetailValue.Render(string(svc.Spec.Type)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Cluster IP:"))
	b.WriteString(p.styles.DetailValue.Render(svc.Spec.ClusterIP))
	b.WriteString("\n")

	externalIP := k8s.GetServiceExternalIP(&svc)
	b.WriteString(p.styles.DetailLabel.Render("External IP:"))
	b.WriteString(p.styles.DetailValue.Render(externalIP))
	b.WriteString("\n")

	ports := k8s.GetServicePorts(&svc)
	b.WriteString(p.styles.DetailLabel.Render("Ports:"))
	b.WriteString(p.styles.DetailValue.Render(ports))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(svc.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(svc.Namespace))
		b.WriteString("\n")
	}

	// Selector
	if len(svc.Spec.Selector) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Selector:"))
		b.WriteString("\n")
		for k, v := range svc.Spec.Selector {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	// Ports detail
	if len(svc.Spec.Ports) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Port Details:"))
		b.WriteString("\n")

		header := fmt.Sprintf("  %-15s %-10s %-10s %-10s", "NAME", "PORT", "TARGET", "PROTOCOL")
		b.WriteString(p.styles.TableHeader.Render(header))
		b.WriteString("\n")

		for _, port := range svc.Spec.Ports {
			name := port.Name
			if name == "" {
				name = "-"
			}
			row := fmt.Sprintf("  %-15s %-10d %-10s %-10s",
				utils.Truncate(name, 15),
				port.Port,
				port.TargetPort.String(),
				port.Protocol,
			)
			b.WriteString(p.styles.TableRow.Render(row))
			b.WriteString("\n")
		}
	}

	// Key hints
	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[p]ort-forward [d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *ServicesPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var services []corev1.Service
		var err error

		if p.allNs {
			services, err = p.client.ListServicesAllNamespaces(ctx)
		} else {
			services, err = p.client.ListServices(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}
		return servicesLoadedMsg{services: services}
	}
}

func (p *ServicesPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	svc := p.filtered[p.cursor]
	return func() tea.Msg {
		ctx := context.Background()
		err := p.client.DeleteService(ctx, svc.Namespace, svc.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}
		return StatusMsg{Message: fmt.Sprintf("Deleted service: %s", svc.Name)}
	}
}

func (p *ServicesPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}
	return &p.filtered[p.cursor]
}

func (p *ServicesPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}
	return p.filtered[p.cursor].Name
}

func (p *ServicesPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}
	svc := p.filtered[p.cursor]
	data, err := yaml.Marshal(svc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *ServicesPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}
	svc := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:              %s\n", svc.Name))
	b.WriteString(fmt.Sprintf("Namespace:         %s\n", svc.Namespace))
	b.WriteString(fmt.Sprintf("Type:              %s\n", svc.Spec.Type))
	b.WriteString(fmt.Sprintf("IP:                %s\n", svc.Spec.ClusterIP))
	b.WriteString(fmt.Sprintf("External IP:       %s\n", k8s.GetServiceExternalIP(&svc)))

	if len(svc.Labels) > 0 {
		b.WriteString("\nLabels:\n")
		for k, v := range svc.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(svc.Spec.Selector) > 0 {
		b.WriteString("\nSelector:\n")
		for k, v := range svc.Spec.Selector {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(svc.Spec.Ports) > 0 {
		b.WriteString("\nPorts:\n")
		for _, port := range svc.Spec.Ports {
			name := port.Name
			if name == "" {
				name = "<unnamed>"
			}
			b.WriteString(fmt.Sprintf("  %s  %d/%s -> %s\n", name, port.Port, port.Protocol, port.TargetPort.String()))
		}
	}

	return b.String(), nil
}

func (p *ServicesPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.services
		return
	}

	p.filtered = make([]corev1.Service, 0)
	for _, svc := range p.services {
		if strings.Contains(strings.ToLower(svc.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, svc)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *ServicesPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type servicesLoadedMsg struct {
	services []corev1.Service
}
