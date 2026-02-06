package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type IngressPanel struct {
	BasePanel
	client    *k8s.Client
	styles    *theme.Styles
	ingresses []networkingv1.Ingress
	filtered  []networkingv1.Ingress
}

func NewIngressPanel(client *k8s.Client, styles *theme.Styles) *IngressPanel {
	return &IngressPanel{
		BasePanel: BasePanel{
			title:       "Ingresses",
			shortcutKey: "i",
		},
		client: client,
		styles: styles,
	}
}

func (p *IngressPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *IngressPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case ingressLoadedMsg:
		p.ingresses = msg.ingresses
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *IngressPanel) View() string {
	var b strings.Builder

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
		ing := p.filtered[i]
		line := p.renderIngressLine(ing, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *IngressPanel) renderIngressLine(ing networkingv1.Ingress, selected bool) string {
	name := utils.Truncate(ing.Name, p.width-15)
	hosts := p.getIngressHosts(&ing)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-15)
	line += " " + utils.Truncate(hosts, 12)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *IngressPanel) getIngressHosts(ing *networkingv1.Ingress) string {
	hosts := make([]string, 0)

	for _, rule := range ing.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, rule.Host)
		}
	}

	if len(hosts) == 0 {
		return "*"
	}

	return strings.Join(hosts, ",")
}

func (p *IngressPanel) getIngressAddress(ing *networkingv1.Ingress) string {
	addresses := make([]string, 0)

	for _, lb := range ing.Status.LoadBalancer.Ingress {
		if lb.IP != "" {
			addresses = append(addresses, lb.IP)
		} else if lb.Hostname != "" {
			addresses = append(addresses, lb.Hostname)
		}
	}

	if len(addresses) == 0 {
		return "<pending>"
	}

	return strings.Join(addresses, ",")
}

func (p *IngressPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No ingress selected"
	}

	ing := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Ingress: " + ing.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Hosts:"))
	b.WriteString(p.styles.DetailValue.Render(p.getIngressHosts(&ing)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Address:"))
	b.WriteString(p.styles.DetailValue.Render(p.getIngressAddress(&ing)))
	b.WriteString("\n")

	if ing.Spec.IngressClassName != nil {
		b.WriteString(p.styles.DetailLabel.Render("Class:"))
		b.WriteString(p.styles.DetailValue.Render(*ing.Spec.IngressClassName))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(ing.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(ing.Namespace))
		b.WriteString("\n")
	}

	if len(ing.Spec.Rules) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Rules:"))
		b.WriteString("\n")

		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = "*"
			}

			b.WriteString(fmt.Sprintf("  Host: %s\n", host))

			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					pathStr := path.Path
					if pathStr == "" {
						pathStr = "/"
					}

					pathType := "Prefix"
					if path.PathType != nil {
						pathType = string(*path.PathType)
					}

					backend := fmt.Sprintf(
						"%s:%v",
						path.Backend.Service.Name,
						path.Backend.Service.Port.Number,
					)
					b.WriteString(fmt.Sprintf("    %s (%s) -> %s\n", pathStr, pathType, backend))
				}
			}
		}
	}

	if len(ing.Spec.TLS) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("TLS:"))
		b.WriteString("\n")

		for _, tls := range ing.Spec.TLS {
			b.WriteString(fmt.Sprintf("  Secret: %s\n", tls.SecretName))
			b.WriteString(fmt.Sprintf("  Hosts:  %s\n", strings.Join(tls.Hosts, ", ")))
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *IngressPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			ingresses *networkingv1.IngressList
			err       error
		)

		if p.allNs {
			ingresses, err = p.client.Clientset().
				NetworkingV1().
				Ingresses("").
				List(ctx, metav1.ListOptions{})
		} else {
			ingresses, err = p.client.Clientset().
				NetworkingV1().
				Ingresses(p.client.CurrentNamespace()).
				List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return ingressLoadedMsg{ingresses: ingresses.Items}
	}
}

func (p *IngressPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	ing := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.Clientset().
			NetworkingV1().
			Ingresses(ing.Namespace).
			Delete(ctx, ing.Name, metav1.DeleteOptions{})
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted ingress: %s", ing.Name)}
	}
}

func (p *IngressPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *IngressPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *IngressPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	ing := p.filtered[p.cursor]

	data, err := yaml.Marshal(ing)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *IngressPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	ing := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:          %s\n", ing.Name))
	b.WriteString(fmt.Sprintf("Namespace:     %s\n", ing.Namespace))
	b.WriteString(fmt.Sprintf("Address:       %s\n", p.getIngressAddress(&ing)))
	b.WriteString(
		fmt.Sprintf("Age:           %s\n", utils.FormatAgeFromMeta(ing.CreationTimestamp)),
	)

	if len(ing.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range ing.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nRules:\n")

	for _, rule := range ing.Spec.Rules {
		host := rule.Host
		if host == "" {
			host = "*"
		}

		b.WriteString(fmt.Sprintf("  Host: %s\n", host))

		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				b.WriteString(fmt.Sprintf("    Path: %s -> %s:%v\n",
					path.Path,
					path.Backend.Service.Name,
					path.Backend.Service.Port.Number))
			}
		}
	}

	return b.String(), nil
}

func (p *IngressPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.ingresses

		return
	}

	p.filtered = make([]networkingv1.Ingress, 0)
	for _, ing := range p.ingresses {
		if strings.Contains(strings.ToLower(ing.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, ing)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *IngressPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type ingressLoadedMsg struct {
	ingresses []networkingv1.Ingress
}
