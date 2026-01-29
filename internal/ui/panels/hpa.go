package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type HPAPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	hpas     []autoscalingv2.HorizontalPodAutoscaler
	filtered []autoscalingv2.HorizontalPodAutoscaler
}

func NewHPAPanel(client *k8s.Client, styles *theme.Styles) *HPAPanel {
	return &HPAPanel{
		BasePanel: BasePanel{
			title:       "HPAs",
			shortcutKey: "0",
		},
		client: client,
		styles: styles,
	}
}

func (p *HPAPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *HPAPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("m"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			hpa := p.filtered[p.cursor]

			minReplicas := int32(1)
			if hpa.Spec.MinReplicas != nil {
				minReplicas = *hpa.Spec.MinReplicas
			}

			return p, func() tea.Msg {
				return EditHPAMinReplicasRequestMsg{
					HPAName:     hpa.Name,
					Namespace:   hpa.Namespace,
					MinReplicas: minReplicas,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("M"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			hpa := p.filtered[p.cursor]

			return p, func() tea.Msg {
				return EditHPAMaxReplicasRequestMsg{
					HPAName:     hpa.Name,
					Namespace:   hpa.Namespace,
					MaxReplicas: hpa.Spec.MaxReplicas,
				}
			}
		}

	case hpaLoadedMsg:
		p.hpas = msg.hpas
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *HPAPanel) View() string {
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
		hpa := p.filtered[i]
		line := p.renderHPALine(hpa, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *HPAPanel) renderHPALine(hpa autoscalingv2.HorizontalPodAutoscaler, selected bool) string {
	name := utils.Truncate(hpa.Name, p.width-18)
	replicas := k8s.GetHPAReplicaCount(&hpa)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-15)
	line += " " + p.styles.StatusRunning.Render(replicas)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *HPAPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No HPA selected"
	}

	hpa := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("HPA: " + hpa.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Target:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetHPATargetRef(&hpa)))
	b.WriteString("\n")

	minReplicas := int32(1)
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	b.WriteString(p.styles.DetailLabel.Render("Min Replicas:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", minReplicas)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Max Replicas:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", hpa.Spec.MaxReplicas)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Current:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", hpa.Status.CurrentReplicas)))
	b.WriteString("\n")

	if hpa.Status.DesiredReplicas != hpa.Status.CurrentReplicas {
		b.WriteString(p.styles.DetailLabel.Render("Desired:"))
		b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", hpa.Status.DesiredReplicas)))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(hpa.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(hpa.Namespace))
		b.WriteString("\n")
	}

	if len(hpa.Spec.Metrics) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Metrics:"))
		b.WriteString("\n")

		for _, metric := range hpa.Spec.Metrics {
			b.WriteString(fmt.Sprintf("  Type: %s\n", metric.Type))

			switch metric.Type {
			case autoscalingv2.ResourceMetricSourceType:
				if metric.Resource != nil {
					b.WriteString(fmt.Sprintf("    Resource: %s\n", metric.Resource.Name))

					if metric.Resource.Target.AverageUtilization != nil {
						b.WriteString(
							fmt.Sprintf(
								"    Target: %d%%\n",
								*metric.Resource.Target.AverageUtilization,
							),
						)
					}
				}
			case autoscalingv2.ExternalMetricSourceType:
				if metric.External != nil {
					b.WriteString(fmt.Sprintf("    Metric: %s\n", metric.External.Metric.Name))
				}
			case autoscalingv2.ObjectMetricSourceType,
				autoscalingv2.PodsMetricSourceType,
				autoscalingv2.ContainerResourceMetricSourceType:
				// Additional metric types
			}
		}
	}

	if len(hpa.Status.Conditions) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Conditions:"))
		b.WriteString("\n")

		for _, cond := range hpa.Status.Conditions {
			b.WriteString(fmt.Sprintf("  %s: %s\n", cond.Type, cond.Status))
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[m]in replicas [M]ax replicas [d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *HPAPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			hpas []autoscalingv2.HorizontalPodAutoscaler
			err  error
		)

		if p.allNs {
			hpas, err = p.client.ListHPAsAllNamespaces(ctx)
		} else {
			hpas, err = p.client.ListHPAs(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return hpaLoadedMsg{hpas: hpas}
	}
}

func (p *HPAPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	hpa := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteHPA(ctx, hpa.Namespace, hpa.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted HPA: %s", hpa.Name)}
	}
}

func (p *HPAPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *HPAPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *HPAPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	hpa := p.filtered[p.cursor]

	data, err := yaml.Marshal(hpa)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *HPAPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	hpa := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:               %s\n", hpa.Name))
	b.WriteString(fmt.Sprintf("Namespace:          %s\n", hpa.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"CreationTimestamp:  %s\n",
			utils.FormatTimestampFromMeta(hpa.CreationTimestamp),
		),
	)
	b.WriteString(
		fmt.Sprintf(
			"Reference:          %s/%s\n",
			hpa.Spec.ScaleTargetRef.Kind,
			hpa.Spec.ScaleTargetRef.Name,
		),
	)

	minReplicas := int32(1)
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	b.WriteString(fmt.Sprintf("Min Replicas:       %d\n", minReplicas))
	b.WriteString(fmt.Sprintf("Max Replicas:       %d\n", hpa.Spec.MaxReplicas))
	b.WriteString(fmt.Sprintf("Current Replicas:   %d\n", hpa.Status.CurrentReplicas))
	b.WriteString(fmt.Sprintf("Desired Replicas:   %d\n", hpa.Status.DesiredReplicas))

	if len(hpa.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range hpa.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(hpa.Spec.Metrics) > 0 {
		b.WriteString("\nMetrics:\n")

		for _, metric := range hpa.Spec.Metrics {
			b.WriteString(fmt.Sprintf("  - Type: %s\n", metric.Type))
		}
	}

	if len(hpa.Status.Conditions) > 0 {
		b.WriteString("\nConditions:\n")

		for _, cond := range hpa.Status.Conditions {
			b.WriteString(fmt.Sprintf("  %s: %s (%s)\n", cond.Type, cond.Status, cond.Reason))
		}
	}

	return b.String(), nil
}

func (p *HPAPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.hpas

		return
	}

	p.filtered = make([]autoscalingv2.HorizontalPodAutoscaler, 0)
	for _, hpa := range p.hpas {
		if strings.Contains(strings.ToLower(hpa.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, hpa)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *HPAPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type hpaLoadedMsg struct {
	hpas []autoscalingv2.HorizontalPodAutoscaler
}
