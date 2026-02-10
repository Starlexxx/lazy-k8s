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

type DaemonSetsPanel struct {
	BasePanel
	client     *k8s.Client
	styles     *theme.Styles
	daemonsets []appsv1.DaemonSet
	filtered   []appsv1.DaemonSet
}

func NewDaemonSetsPanel(client *k8s.Client, styles *theme.Styles) *DaemonSetsPanel {
	return &DaemonSetsPanel{
		BasePanel: BasePanel{
			title:       "DaemonSets",
			shortcutKey: "0",
		},
		client: client,
		styles: styles,
	}
}

func (p *DaemonSetsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *DaemonSetsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			return p, p.restartDaemonSet()
		}

	case daemonSetsLoadedMsg:
		p.daemonsets = msg.daemonsets
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *DaemonSetsPanel) View() string {
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
		ds := p.filtered[i]
		line := p.renderDaemonSetLine(ds, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *DaemonSetsPanel) renderDaemonSetLine(ds appsv1.DaemonSet, selected bool) string {
	name := utils.Truncate(ds.Name, p.width-15)
	ready := k8s.GetDaemonSetReadyCount(&ds)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-10)

	readyStyle := p.styles.StatusRunning
	if ds.Status.NumberReady < ds.Status.DesiredNumberScheduled {
		readyStyle = p.styles.StatusPending
	}

	line += " " + readyStyle.Render(ready)

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *DaemonSetsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No daemonset selected"
	}

	ds := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("DaemonSet: " + ds.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Desired:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", ds.Status.DesiredNumberScheduled)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Current:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", ds.Status.CurrentNumberScheduled)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Ready:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", ds.Status.NumberReady)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Up-to-date:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", ds.Status.UpdatedNumberScheduled)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Available:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", ds.Status.NumberAvailable)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(ds.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(ds.Namespace))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Update Strategy:"))
	b.WriteString(p.styles.DetailValue.Render(string(ds.Spec.UpdateStrategy.Type)))
	b.WriteString("\n")

	if len(ds.Spec.Template.Spec.NodeSelector) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Node Selector:"))
		b.WriteString("\n")

		for k, v := range ds.Spec.Template.Spec.NodeSelector {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	images := k8s.GetDaemonSetImages(&ds)
	if len(images) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Images:"))
		b.WriteString("\n")

		for _, img := range images {
			b.WriteString("  " + img + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[r]estart [d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *DaemonSetsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			daemonsets []appsv1.DaemonSet
			err        error
		)

		if p.allNs {
			daemonsets, err = p.client.ListDaemonSetsAllNamespaces(ctx)
		} else {
			daemonsets, err = p.client.ListDaemonSets(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return daemonSetsLoadedMsg{daemonsets: daemonsets}
	}
}

func (p *DaemonSetsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	ds := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteDaemonSet(ctx, ds.Namespace, ds.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted daemonset: %s", ds.Name)}
	}
}

func (p *DaemonSetsPanel) restartDaemonSet() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	ds := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.RestartDaemonSet(ctx, ds.Namespace, ds.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Restarted daemonset: %s", ds.Name)}
	}
}

func (p *DaemonSetsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *DaemonSetsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *DaemonSetsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	ds := p.filtered[p.cursor]

	data, err := yaml.Marshal(ds)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *DaemonSetsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	ds := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:               %s\n", ds.Name))
	b.WriteString(fmt.Sprintf("Namespace:          %s\n", ds.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"CreationTimestamp:  %s\n",
			utils.FormatTimestampFromMeta(ds.CreationTimestamp),
		),
	)
	b.WriteString(
		fmt.Sprintf(
			"Selector:           %s\n",
			ds.Spec.Selector.String(),
		),
	)
	b.WriteString(
		fmt.Sprintf(
			"Node-Selector:      %v\n",
			ds.Spec.Template.Spec.NodeSelector,
		),
	)
	b.WriteString(
		fmt.Sprintf(
			"Pods Status:        %d Running / %d Waiting / %d Succeeded / %d Failed\n",
			ds.Status.NumberReady,
			ds.Status.DesiredNumberScheduled-ds.Status.NumberReady,
			0,
			ds.Status.NumberMisscheduled,
		),
	)
	b.WriteString(fmt.Sprintf("Update Strategy:    %s\n", ds.Spec.UpdateStrategy.Type))

	if len(ds.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range ds.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nPod Template:\n")

	for _, container := range ds.Spec.Template.Spec.Containers {
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

	return b.String(), nil
}

func (p *DaemonSetsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.daemonsets

		return
	}

	p.filtered = make([]appsv1.DaemonSet, 0)
	for _, ds := range p.daemonsets {
		if strings.Contains(strings.ToLower(ds.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, ds)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *DaemonSetsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type daemonSetsLoadedMsg struct {
	daemonsets []appsv1.DaemonSet
}
