package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type NodesPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	nodes    []corev1.Node
	filtered []corev1.Node
	metrics  map[string]NodeMetrics
}

func NewNodesPanel(client *k8s.Client, styles *theme.Styles) *NodesPanel {
	return &NodesPanel{
		BasePanel: BasePanel{
			title:       "Nodes",
			shortcutKey: "7",
		},
		client: client,
		styles: styles,
	}
}

func (p *NodesPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *NodesPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case nodesLoadedMsg:
		p.nodes = msg.nodes
		p.applyFilter()

		return p, nil

	case NodeMetricsMsg:
		p.metrics = msg.Metrics

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *NodesPanel) View() string {
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

	if p.width > 80 {
		b.WriteString(p.renderNodeHeader())
		b.WriteString("\n")

		visibleHeight--
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
		node := p.filtered[i]
		line := p.renderNodeLine(node, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *NodesPanel) renderNodeHeader() string {
	hasMetrics := len(p.metrics) > 0

	nameW := p.nodeNameWidth(hasMetrics)
	header := "  " + utils.PadRight("NAME", nameW)

	if hasMetrics {
		header += " " + utils.PadLeft("CPU", 5)
		header += " " + utils.PadLeft("MEM", 6)
	}

	header += " " + utils.PadRight("STATUS", 8)
	header += " " + utils.PadRight("ROLES", 15)
	header += " " + utils.PadRight("VERSION", 12)
	header += " " + utils.PadRight("AGE", 8)

	return p.styles.TableHeader.Render(
		utils.Truncate(header, p.width-2),
	)
}

func (p *NodesPanel) nodeNameWidth(hasMetrics bool) int {
	// Reserve space for: status(8) + roles(15) + version(12) + age(8) + padding
	reserved := 50

	if hasMetrics {
		reserved += 13
	}

	nameW := p.width - reserved
	if nameW < 10 {
		nameW = 10
	}

	return nameW
}

func (p *NodesPanel) renderNodeLine(node corev1.Node, selected bool) string {
	status := k8s.GetNodeStatus(&node)

	hasMetrics := false

	var cpuStr, memStr string

	if m, ok := p.metrics[node.Name]; ok {
		hasMetrics = true
		cpuStr = utils.FormatCPU(m.CPU)
		memStr = utils.FormatMemory(m.Memory)
	}

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	// Wide mode: extra columns (roles, version, age)
	if p.width > 80 {
		nameW := p.nodeNameWidth(hasMetrics)
		line += utils.PadRight(utils.Truncate(node.Name, nameW), nameW)

		if hasMetrics {
			line += " " + p.styles.Muted.Render(utils.PadLeft(cpuStr, 5))
			line += " " + p.styles.Muted.Render(utils.PadLeft(memStr, 6))
		}

		statusStyle := p.styles.GetStatusStyle(status)
		line += " " + statusStyle.Render(utils.PadRight(status, 8))

		roles := k8s.GetNodeRoles(&node)
		line += " " + utils.PadRight(utils.Truncate(roles, 15), 15)

		version := node.Status.NodeInfo.KubeletVersion
		line += " " + utils.PadRight(utils.Truncate(version, 12), 12)

		age := utils.FormatAgeFromMeta(node.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)

		if selected && p.focused {
			return p.styles.ListItemFocused.Render(line)
		} else if selected {
			return p.styles.ListItemSelected.Render(line)
		}

		return p.styles.ListItem.Render(line)
	}

	// Narrow mode: name + optional metrics + status
	nameWidth := p.width - 12

	if hasMetrics {
		nameWidth = p.width - 28
	}

	if nameWidth < 10 {
		nameWidth = 10
	}

	line += utils.Truncate(node.Name, nameWidth)

	if hasMetrics {
		line = utils.PadRight(line, p.width-25)
		line += " " + p.styles.Muted.Render(utils.PadLeft(cpuStr, 5))
		line += " " + p.styles.Muted.Render(utils.PadLeft(memStr, 6))
	} else {
		line = utils.PadRight(line, p.width-10)
	}

	statusStyle := p.styles.GetStatusStyle(status)
	line += " " + statusStyle.Render(status)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *NodesPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No node selected"
	}

	node := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Node: " + node.Name))
	b.WriteString("\n\n")

	status := k8s.GetNodeStatus(&node)

	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(status).Render(status))
	b.WriteString("\n")

	roles := k8s.GetNodeRoles(&node)

	b.WriteString(p.styles.DetailLabel.Render("Roles:"))
	b.WriteString(p.styles.DetailValue.Render(roles))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(node.CreationTimestamp)))
	b.WriteString("\n")

	// Version
	b.WriteString(p.styles.DetailLabel.Render("Version:"))
	b.WriteString(p.styles.DetailValue.Render(node.Status.NodeInfo.KubeletVersion))
	b.WriteString("\n")

	internalIP := k8s.GetNodeInternalIP(&node)

	b.WriteString(p.styles.DetailLabel.Render("Internal IP:"))
	b.WriteString(p.styles.DetailValue.Render(internalIP))
	b.WriteString("\n")

	externalIP := k8s.GetNodeExternalIP(&node)

	b.WriteString(p.styles.DetailLabel.Render("External IP:"))
	b.WriteString(p.styles.DetailValue.Render(externalIP))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(p.styles.DetailTitle.Render("System Info:"))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("OS:"))
	b.WriteString(
		p.styles.DetailValue.Render(
			fmt.Sprintf(
				"%s %s",
				node.Status.NodeInfo.OperatingSystem,
				node.Status.NodeInfo.OSImage,
			),
		),
	)
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Kernel:"))
	b.WriteString(p.styles.DetailValue.Render(node.Status.NodeInfo.KernelVersion))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Container:"))
	b.WriteString(p.styles.DetailValue.Render(node.Status.NodeInfo.ContainerRuntimeVersion))
	b.WriteString("\n")

	cpu, memory := k8s.GetNodeCapacity(&node)

	b.WriteString("\n")
	b.WriteString(p.styles.DetailTitle.Render("Capacity:"))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("CPU:"))
	b.WriteString(p.styles.DetailValue.Render(cpu))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Memory:"))
	b.WriteString(p.styles.DetailValue.Render(memory))
	b.WriteString("\n")

	// Show current usage if metrics available
	if m, ok := p.metrics[node.Name]; ok {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Current Usage:"))
		b.WriteString("\n")

		b.WriteString(p.styles.DetailLabel.Render("CPU:"))
		b.WriteString(p.styles.DetailValue.Render(utils.FormatCPU(m.CPU)))
		b.WriteString("\n")

		b.WriteString(p.styles.DetailLabel.Render("Memory:"))
		b.WriteString(p.styles.DetailValue.Render(utils.FormatMemory(m.Memory)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(p.styles.DetailTitle.Render("Conditions:"))
	b.WriteString("\n")

	for _, cond := range node.Status.Conditions {
		condStatus := "False"

		var condStyle lipgloss.Style

		if cond.Status == "True" {
			condStatus = "True"

			if cond.Type == corev1.NodeReady {
				condStyle = p.styles.StatusRunning
			} else {
				condStyle = p.styles.StatusFailed
			}
		} else if cond.Type == corev1.NodeReady {
			condStyle = p.styles.StatusFailed
		} else {
			condStyle = p.styles.StatusRunning
		}

		b.WriteString(fmt.Sprintf("  %s: %s\n", cond.Type, condStyle.Render(condStatus)))
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml"))

	return b.String()
}

func (p *NodesPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		nodes, err := p.client.ListNodes(ctx)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return nodesLoadedMsg{nodes: nodes}
	}
}

func (p *NodesPanel) Delete() tea.Cmd {
	// Nodes cannot be deleted via the API
	return func() tea.Msg {
		return StatusMsg{Message: "Cannot delete nodes through the API"}
	}
}

func (p *NodesPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *NodesPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *NodesPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	node := p.filtered[p.cursor]

	data, err := yaml.Marshal(node)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *NodesPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	node := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:               %s\n", node.Name))
	b.WriteString(fmt.Sprintf("Roles:              %s\n", k8s.GetNodeRoles(&node)))
	b.WriteString(
		fmt.Sprintf("Age:                %s\n", utils.FormatAgeFromMeta(node.CreationTimestamp)),
	)
	b.WriteString(fmt.Sprintf("Taints:             %d\n", len(node.Spec.Taints)))

	if len(node.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range node.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nSystem Info:\n")
	b.WriteString(fmt.Sprintf("  Machine ID:      %s\n", node.Status.NodeInfo.MachineID))
	b.WriteString(fmt.Sprintf("  System UUID:     %s\n", node.Status.NodeInfo.SystemUUID))
	b.WriteString(fmt.Sprintf("  Boot ID:         %s\n", node.Status.NodeInfo.BootID))
	b.WriteString(fmt.Sprintf("  Kernel Version:  %s\n", node.Status.NodeInfo.KernelVersion))
	b.WriteString(fmt.Sprintf("  OS Image:        %s\n", node.Status.NodeInfo.OSImage))
	b.WriteString(
		fmt.Sprintf("  Container Runtime: %s\n", node.Status.NodeInfo.ContainerRuntimeVersion),
	)
	b.WriteString(fmt.Sprintf("  Kubelet Version: %s\n", node.Status.NodeInfo.KubeletVersion))

	b.WriteString("\nCapacity:\n")

	for resource, qty := range node.Status.Capacity {
		b.WriteString(fmt.Sprintf("  %s: %s\n", resource, qty.String()))
	}

	b.WriteString("\nAllocatable:\n")

	for resource, qty := range node.Status.Allocatable {
		b.WriteString(fmt.Sprintf("  %s: %s\n", resource, qty.String()))
	}

	b.WriteString("\nConditions:\n")

	for _, cond := range node.Status.Conditions {
		b.WriteString(fmt.Sprintf("  %s: %s (%s)\n", cond.Type, cond.Status, cond.Reason))
	}

	return b.String(), nil
}

func (p *NodesPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.nodes

		return
	}

	p.filtered = make([]corev1.Node, 0)
	for _, node := range p.nodes {
		if strings.Contains(strings.ToLower(node.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, node)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *NodesPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type nodesLoadedMsg struct {
	nodes []corev1.Node
}
