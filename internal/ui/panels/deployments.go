package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type DeploymentsPanel struct {
	BasePanel
	client      *k8s.Client
	styles      *theme.Styles
	deployments []appsv1.Deployment
	filtered    []appsv1.Deployment
}

func NewDeploymentsPanel(client *k8s.Client, styles *theme.Styles) *DeploymentsPanel {
	return &DeploymentsPanel{
		BasePanel: BasePanel{
			title:       "Deployments",
			shortcutKey: "3",
		},
		client: client,
		styles: styles,
	}
}

func (p *DeploymentsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *DeploymentsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			deploy := p.filtered[p.cursor]

			replicas := int32(0)
			if deploy.Spec.Replicas != nil {
				replicas = *deploy.Spec.Replicas
			}

			return p, func() tea.Msg {
				return ScaleRequestMsg{
					DeploymentName:  deploy.Name,
					Namespace:       deploy.Namespace,
					CurrentReplicas: replicas,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			return p, p.restartDeployment()
		case key.Matches(msg, key.NewBinding(key.WithKeys("R"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			deploy := p.filtered[p.cursor]

			return p, func() tea.Msg {
				return RollbackRequestMsg{
					DeploymentName: deploy.Name,
					Namespace:      deploy.Namespace,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("V"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			deploy := p.filtered[p.cursor]

			return p, func() tea.Msg {
				return DiffRequestMsg{
					DeploymentName: deploy.Name,
					Namespace:      deploy.Namespace,
				}
			}
		}

	case deploymentsLoadedMsg:
		p.deployments = msg.deployments
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *DeploymentsPanel) View() string {
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
		deploy := p.filtered[i]
		line := p.renderDeploymentLine(deploy, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *DeploymentsPanel) renderDeploymentLine(deploy appsv1.Deployment, selected bool) string {
	ready := k8s.GetDeploymentReadyCount(&deploy)

	readyStyle := p.styles.StatusRunning
	if deploy.Status.ReadyReplicas < *deploy.Spec.Replicas {
		readyStyle = p.styles.StatusPending
	}

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	if p.width > 80 {
		reserved := 28
		if p.width > 120 && p.allNs {
			reserved += 16
		}

		nameW := p.width - reserved
		if nameW < 10 {
			nameW = 10
		}

		line += utils.PadRight(
			utils.Truncate(deploy.Name, nameW), nameW,
		)
		line += " " + readyStyle.Render(utils.PadRight(ready, 7))

		age := utils.FormatAgeFromMeta(deploy.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)

		images := k8s.GetDeploymentImages(&deploy)
		if len(images) > 0 {
			imgStr := utils.Truncate(images[0], 30)
			line += " " + p.styles.Muted.Render(imgStr)
		}

		if p.width > 120 && p.allNs {
			line += " " + utils.Truncate(deploy.Namespace, 15)
		}
	} else {
		name := utils.Truncate(deploy.Name, p.width-15)
		line += name
		line = utils.PadRight(line, p.width-10)
		line += " " + readyStyle.Render(ready)
	}

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *DeploymentsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No deployment selected"
	}

	deploy := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Deployment: " + deploy.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Ready:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetDeploymentReadyCount(&deploy)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Up-to-date:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", deploy.Status.UpdatedReplicas)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Available:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", deploy.Status.AvailableReplicas)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(deploy.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(deploy.Namespace))
		b.WriteString("\n")
	}

	// Strategy
	b.WriteString(p.styles.DetailLabel.Render("Strategy:"))
	b.WriteString(p.styles.DetailValue.Render(string(deploy.Spec.Strategy.Type)))
	b.WriteString("\n")

	images := k8s.GetDeploymentImages(&deploy)
	if len(images) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Images:"))
		b.WriteString("\n")

		for _, img := range images {
			b.WriteString("  " + img + "\n")
		}
	}

	if len(deploy.Status.Conditions) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Conditions:"))
		b.WriteString("\n")

		for _, cond := range deploy.Status.Conditions {
			status := "False"
			if cond.Status == "True" {
				status = "True"
			}

			b.WriteString(fmt.Sprintf("  %s: %s\n", cond.Type, status))
		}
	}

	b.WriteString("\n")
	b.WriteString(
		p.styles.Muted.Render(
			"[s]cale [r]estart [R]ollback [V]ersion diff [d]escribe [y]aml [D]elete",
		),
	)

	return b.String()
}

func (p *DeploymentsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			deployments []appsv1.Deployment
			err         error
		)

		if p.allNs {
			deployments, err = p.client.ListDeploymentsAllNamespaces(ctx)
		} else {
			deployments, err = p.client.ListDeployments(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return deploymentsLoadedMsg{deployments: deployments}
	}
}

func (p *DeploymentsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	deploy := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteDeployment(ctx, deploy.Namespace, deploy.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted deployment: %s", deploy.Name)}
	}
}

func (p *DeploymentsPanel) restartDeployment() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	deploy := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.RestartDeployment(ctx, deploy.Namespace, deploy.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Restarted deployment: %s", deploy.Name)}
	}
}

func (p *DeploymentsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *DeploymentsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *DeploymentsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	deploy := p.filtered[p.cursor]

	data, err := yaml.Marshal(deploy)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *DeploymentsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	deploy := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:               %s\n", deploy.Name))
	b.WriteString(fmt.Sprintf("Namespace:          %s\n", deploy.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"CreationTimestamp:  %s\n",
			utils.FormatTimestampFromMeta(deploy.CreationTimestamp),
		),
	)
	b.WriteString(
		fmt.Sprintf(
			"Replicas:           %d desired | %d updated | %d total | %d available | %d unavailable\n",
			*deploy.Spec.Replicas,
			deploy.Status.UpdatedReplicas,
			deploy.Status.Replicas,
			deploy.Status.AvailableReplicas,
			deploy.Status.UnavailableReplicas,
		),
	)
	b.WriteString(fmt.Sprintf("Strategy:           %s\n", deploy.Spec.Strategy.Type))

	if len(deploy.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range deploy.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nPod Template:\n")

	for _, container := range deploy.Spec.Template.Spec.Containers {
		b.WriteString(fmt.Sprintf("  Container: %s\n", container.Name))
		b.WriteString(fmt.Sprintf("    Image:   %s\n", container.Image))

		if len(container.Ports) > 0 {
			b.WriteString("    Ports:   ")

			var ports []string
			for _, port := range container.Ports {
				ports = append(ports, fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol))
			}

			b.WriteString(strings.Join(ports, ", "))
			b.WriteString("\n")
		}
	}

	if len(deploy.Status.Conditions) > 0 {
		b.WriteString("\nConditions:\n")

		for _, cond := range deploy.Status.Conditions {
			b.WriteString(fmt.Sprintf("  %s: %s (%s)\n", cond.Type, cond.Status, cond.Reason))
		}
	}

	return b.String(), nil
}

func (p *DeploymentsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.deployments

		return
	}

	p.filtered = make([]appsv1.Deployment, 0)
	for _, deploy := range p.deployments {
		if strings.Contains(strings.ToLower(deploy.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, deploy)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *DeploymentsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type deploymentsLoadedMsg struct {
	deployments []appsv1.Deployment
}
