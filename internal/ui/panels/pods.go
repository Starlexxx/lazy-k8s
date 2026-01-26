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

type PodsPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	pods     []corev1.Pod
	filtered []corev1.Pod
}

func NewPodsPanel(client *k8s.Client, styles *theme.Styles) *PodsPanel {
	return &PodsPanel{
		BasePanel: BasePanel{
			title:       "Pods",
			shortcutKey: "2",
		},
		client: client,
		styles: styles,
	}
}

func (p *PodsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *PodsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("p"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			pod := p.filtered[p.cursor]

			var ports []int32

			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					ports = append(ports, port.ContainerPort)
				}
			}

			return p, func() tea.Msg {
				return PortForwardRequestMsg{
					PodName:   pod.Name,
					Namespace: pod.Namespace,
					Ports:     ports,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			pod := p.filtered[p.cursor]

			var containers []string
			for _, container := range pod.Spec.Containers {
				containers = append(containers, container.Name)
			}

			return p, func() tea.Msg {
				return ExecRequestMsg{
					PodName:    pod.Name,
					Namespace:  pod.Namespace,
					Containers: containers,
				}
			}
		}

	case podsLoadedMsg:
		p.pods = msg.pods
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *PodsPanel) View() string {
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
	visibleHeight := p.height - 3
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
		pod := p.filtered[i]
		line := p.renderPodLine(pod, i == p.cursor)
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

func (p *PodsPanel) renderPodLine(pod corev1.Pod, selected bool) string {
	name := utils.Truncate(pod.Name, p.width-15)
	status := k8s.GetPodStatus(&pod)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	// Pad and add status
	line = utils.PadRight(line, p.width-12)
	statusStyle := p.styles.GetStatusStyle(status)
	line += " " + statusStyle.Render(utils.Truncate(status, 10))

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *PodsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No pod selected"
	}

	pod := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Pod: " + pod.Name))
	b.WriteString("\n\n")

	// Basic info
	status := k8s.GetPodStatus(&pod)

	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(status).Render(status))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Ready:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetPodReadyCount(&pod)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Restarts:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", k8s.GetPodRestarts(&pod))))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(pod.CreationTimestamp)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Node:"))
	b.WriteString(p.styles.DetailValue.Render(pod.Spec.NodeName))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("IP:"))
	b.WriteString(p.styles.DetailValue.Render(pod.Status.PodIP))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(pod.Namespace))
		b.WriteString("\n")
	}

	// Containers
	b.WriteString("\n")
	b.WriteString(p.styles.DetailTitle.Render("Containers:"))
	b.WriteString("\n")

	// Table header
	header := fmt.Sprintf("  %-20s %-8s %-10s %-8s", "NAME", "READY", "STATUS", "RESTARTS")
	b.WriteString(p.styles.TableHeader.Render(header))
	b.WriteString("\n")

	for _, container := range pod.Spec.Containers {
		var cs *corev1.ContainerStatus

		for i := range pod.Status.ContainerStatuses {
			if pod.Status.ContainerStatuses[i].Name == container.Name {
				cs = &pod.Status.ContainerStatuses[i]

				break
			}
		}

		ready := "false"
		state := "Unknown"
		restarts := int32(0)

		if cs != nil {
			if cs.Ready {
				ready = "true"
			}

			restarts = cs.RestartCount

			if cs.State.Running != nil {
				state = "Running"
			} else if cs.State.Waiting != nil {
				state = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil {
				state = cs.State.Terminated.Reason
			}
		}

		row := fmt.Sprintf("  %-20s %-8s %-10s %-8d",
			utils.Truncate(container.Name, 20),
			ready,
			utils.Truncate(state, 10),
			restarts,
		)
		b.WriteString(p.styles.TableRow.Render(row))
		b.WriteString("\n")
	}

	// Key hints
	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[l]ogs [x]exec [p]ort-forward [d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *PodsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			pods []corev1.Pod
			err  error
		)

		if p.allNs {
			pods, err = p.client.ListPodsAllNamespaces(ctx)
		} else {
			pods, err = p.client.ListPods(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return podsLoadedMsg{pods: pods}
	}
}

func (p *PodsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	pod := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeletePod(ctx, pod.Namespace, pod.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted pod: %s", pod.Name)}
	}
}

func (p *PodsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *PodsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *PodsPanel) SelectedPod() *corev1.Pod {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *PodsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	pod := p.filtered[p.cursor]

	data, err := yaml.Marshal(pod)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *PodsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	pod := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:         %s\n", pod.Name))
	b.WriteString(fmt.Sprintf("Namespace:    %s\n", pod.Namespace))
	b.WriteString(fmt.Sprintf("Node:         %s\n", pod.Spec.NodeName))

	startTime := ""
	if pod.Status.StartTime != nil {
		startTime = utils.FormatTimestampFromMeta(*pod.Status.StartTime)
	}

	b.WriteString(fmt.Sprintf("Start Time:   %s\n", startTime))
	b.WriteString(fmt.Sprintf("Status:       %s\n", k8s.GetPodStatus(&pod)))
	b.WriteString(fmt.Sprintf("IP:           %s\n", pod.Status.PodIP))

	if len(pod.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range pod.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(pod.Annotations) > 0 {
		b.WriteString("\nAnnotations:\n")

		for k, v := range pod.Annotations {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nContainers:\n")

	for _, container := range pod.Spec.Containers {
		b.WriteString(fmt.Sprintf("  %s:\n", container.Name))
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

		// Find container status
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == container.Name {
				b.WriteString(fmt.Sprintf("    Ready:   %v\n", cs.Ready))
				b.WriteString(fmt.Sprintf("    Restarts: %d\n", cs.RestartCount))

				break
			}
		}
	}

	return b.String(), nil
}

func (p *PodsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.pods

		return
	}

	p.filtered = make([]corev1.Pod, 0)
	for _, pod := range p.pods {
		if strings.Contains(strings.ToLower(pod.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, pod)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *PodsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type podsLoadedMsg struct {
	pods []corev1.Pod
}
