package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type NetworkPoliciesPanel struct {
	BasePanel
	client          *k8s.Client
	styles          *theme.Styles
	networkPolicies []networkingv1.NetworkPolicy
	filtered        []networkingv1.NetworkPolicy
}

func NewNetworkPoliciesPanel(client *k8s.Client, styles *theme.Styles) *NetworkPoliciesPanel {
	return &NetworkPoliciesPanel{
		BasePanel: BasePanel{
			title:       "NetworkPolicies",
			shortcutKey: "0",
		},
		client: client,
		styles: styles,
	}
}

func (p *NetworkPoliciesPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *NetworkPoliciesPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case networkPoliciesLoadedMsg:
		p.networkPolicies = msg.networkPolicies
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *NetworkPoliciesPanel) View() string {
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
		np := p.filtered[i]
		line := p.renderNetworkPolicyLine(np, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *NetworkPoliciesPanel) renderNetworkPolicyLine(
	np networkingv1.NetworkPolicy,
	selected bool,
) string {
	name := utils.Truncate(np.Name, p.width-15)
	rules := k8s.GetNetworkPolicyRuleSummary(&np)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-20)
	line += " " + p.styles.StatusRunning.Render(rules)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *NetworkPoliciesPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No network policy selected"
	}

	np := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("NetworkPolicy: " + np.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Pod Selector:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetNetworkPolicyPodSelectorString(&np)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Policy Types:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetNetworkPolicyPolicyTypes(&np)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Ingress Rules:"))
	b.WriteString(
		p.styles.DetailValue.Render(fmt.Sprintf("%d", k8s.GetNetworkPolicyIngressRuleCount(&np))),
	)
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Egress Rules:"))
	b.WriteString(
		p.styles.DetailValue.Render(fmt.Sprintf("%d", k8s.GetNetworkPolicyEgressRuleCount(&np))),
	)
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(np.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(np.Namespace))
		b.WriteString("\n")
	}

	if len(np.Spec.Ingress) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Ingress Rules:"))
		b.WriteString("\n")

		for i, rule := range np.Spec.Ingress {
			b.WriteString(fmt.Sprintf("  Rule %d:\n", i+1))

			if len(rule.From) > 0 {
				b.WriteString(fmt.Sprintf("    From: %d source(s)\n", len(rule.From)))
			}

			if len(rule.Ports) > 0 {
				var ports []string

				for _, port := range rule.Ports {
					if port.Port != nil {
						ports = append(ports, port.Port.String())
					}
				}

				if len(ports) > 0 {
					b.WriteString(fmt.Sprintf("    Ports: %s\n", strings.Join(ports, ", ")))
				}
			}
		}
	}

	if len(np.Spec.Egress) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Egress Rules:"))
		b.WriteString("\n")

		for i, rule := range np.Spec.Egress {
			b.WriteString(fmt.Sprintf("  Rule %d:\n", i+1))

			if len(rule.To) > 0 {
				b.WriteString(fmt.Sprintf("    To: %d destination(s)\n", len(rule.To)))
			}

			if len(rule.Ports) > 0 {
				var ports []string

				for _, port := range rule.Ports {
					if port.Port != nil {
						ports = append(ports, port.Port.String())
					}
				}

				if len(ports) > 0 {
					b.WriteString(fmt.Sprintf("    Ports: %s\n", strings.Join(ports, ", ")))
				}
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *NetworkPoliciesPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			networkPolicies []networkingv1.NetworkPolicy
			err             error
		)

		if p.allNs {
			networkPolicies, err = p.client.ListNetworkPoliciesAllNamespaces(ctx)
		} else {
			networkPolicies, err = p.client.ListNetworkPolicies(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return networkPoliciesLoadedMsg{networkPolicies: networkPolicies}
	}
}

func (p *NetworkPoliciesPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	np := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteNetworkPolicy(ctx, np.Namespace, np.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted network policy: %s", np.Name)}
	}
}

func (p *NetworkPoliciesPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *NetworkPoliciesPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *NetworkPoliciesPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	np := p.filtered[p.cursor]

	data, err := yaml.Marshal(np)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *NetworkPoliciesPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	np := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:         %s\n", np.Name))
	b.WriteString(fmt.Sprintf("Namespace:    %s\n", np.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"Created:      %s\n",
			utils.FormatTimestampFromMeta(np.CreationTimestamp),
		),
	)
	b.WriteString(fmt.Sprintf("Pod Selector: %s\n", k8s.GetNetworkPolicyPodSelectorString(&np)))
	b.WriteString(fmt.Sprintf("Policy Types: %s\n", k8s.GetNetworkPolicyPolicyTypes(&np)))

	if len(np.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range np.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(np.Spec.Ingress) > 0 {
		b.WriteString("\nIngress Rules:\n")

		for i, rule := range np.Spec.Ingress {
			b.WriteString(fmt.Sprintf("  Rule %d:\n", i+1))
			b.WriteString(fmt.Sprintf("    From: %d source(s)\n", len(rule.From)))
			b.WriteString(fmt.Sprintf("    Ports: %d\n", len(rule.Ports)))
		}
	}

	if len(np.Spec.Egress) > 0 {
		b.WriteString("\nEgress Rules:\n")

		for i, rule := range np.Spec.Egress {
			b.WriteString(fmt.Sprintf("  Rule %d:\n", i+1))
			b.WriteString(fmt.Sprintf("    To: %d destination(s)\n", len(rule.To)))
			b.WriteString(fmt.Sprintf("    Ports: %d\n", len(rule.Ports)))
		}
	}

	return b.String(), nil
}

func (p *NetworkPoliciesPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.networkPolicies

		return
	}

	p.filtered = make([]networkingv1.NetworkPolicy, 0)
	for _, np := range p.networkPolicies {
		if strings.Contains(strings.ToLower(np.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, np)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *NetworkPoliciesPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type networkPoliciesLoadedMsg struct {
	networkPolicies []networkingv1.NetworkPolicy
}
