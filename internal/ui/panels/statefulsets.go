package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type StatefulSetsPanel struct {
	BasePanel
	client       *k8s.Client
	styles       *theme.Styles
	statefulsets []appsv1.StatefulSet
	filtered     []appsv1.StatefulSet
}

func NewStatefulSetsPanel(client *k8s.Client, styles *theme.Styles) *StatefulSetsPanel {
	return &StatefulSetsPanel{
		BasePanel: BasePanel{
			title:       "StatefulSets",
			shortcutKey: "0",
		},
		client: client,
		styles: styles,
	}
}

func (p *StatefulSetsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *StatefulSetsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

			sts := p.filtered[p.cursor]

			replicas := int32(0)
			if sts.Spec.Replicas != nil {
				replicas = *sts.Spec.Replicas
			}

			return p, func() tea.Msg {
				return ScaleStatefulSetRequestMsg{
					StatefulSetName: sts.Name,
					Namespace:       sts.Namespace,
					CurrentReplicas: replicas,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			return p, p.restartStatefulSet()
		}

	case statefulSetsLoadedMsg:
		p.statefulsets = msg.statefulsets
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *StatefulSetsPanel) View() string {
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
		sts := p.filtered[i]
		line := p.renderStatefulSetLine(sts, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *StatefulSetsPanel) renderStatefulSetLine(sts appsv1.StatefulSet, selected bool) string {
	name := utils.Truncate(sts.Name, p.width-15)
	ready := k8s.GetStatefulSetReadyCount(&sts)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-10)

	readyStyle := p.styles.StatusRunning

	desired := int32(0)
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}

	if sts.Status.ReadyReplicas < desired {
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

func (p *StatefulSetsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No statefulset selected"
	}

	sts := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("StatefulSet: " + sts.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Ready:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetStatefulSetReadyCount(&sts)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Current:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", sts.Status.CurrentReplicas)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Updated:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", sts.Status.UpdatedReplicas)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(sts.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(sts.Namespace))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Update Strategy:"))
	b.WriteString(p.styles.DetailValue.Render(string(sts.Spec.UpdateStrategy.Type)))
	b.WriteString("\n")

	if sts.Spec.ServiceName != "" {
		b.WriteString(p.styles.DetailLabel.Render("Service Name:"))
		b.WriteString(p.styles.DetailValue.Render(sts.Spec.ServiceName))
		b.WriteString("\n")
	}

	images := k8s.GetStatefulSetImages(&sts)
	if len(images) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Images:"))
		b.WriteString("\n")

		for _, img := range images {
			b.WriteString("  " + img + "\n")
		}
	}

	if len(sts.Spec.VolumeClaimTemplates) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Volume Claim Templates:"))
		b.WriteString("\n")

		for _, pvc := range sts.Spec.VolumeClaimTemplates {
			b.WriteString(fmt.Sprintf("  %s\n", pvc.Name))
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[s]cale [r]estart [d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *StatefulSetsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			statefulsets []appsv1.StatefulSet
			err          error
		)

		if p.allNs {
			statefulsets, err = p.client.ListStatefulSetsAllNamespaces(ctx)
		} else {
			statefulsets, err = p.client.ListStatefulSets(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return statefulSetsLoadedMsg{statefulsets: statefulsets}
	}
}

func (p *StatefulSetsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	sts := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteStatefulSet(ctx, sts.Namespace, sts.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted statefulset: %s", sts.Name)}
	}
}

func (p *StatefulSetsPanel) restartStatefulSet() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	sts := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.RestartStatefulSet(ctx, sts.Namespace, sts.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Restarted statefulset: %s", sts.Name)}
	}
}

func (p *StatefulSetsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *StatefulSetsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *StatefulSetsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	sts := p.filtered[p.cursor]

	data, err := yaml.Marshal(sts)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *StatefulSetsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	sts := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:               %s\n", sts.Name))
	b.WriteString(fmt.Sprintf("Namespace:          %s\n", sts.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"CreationTimestamp:  %s\n",
			utils.FormatTimestampFromMeta(sts.CreationTimestamp),
		),
	)

	desired := int32(0)
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}

	b.WriteString(
		fmt.Sprintf(
			"Replicas:           %d desired | %d current | %d ready | %d updated\n",
			desired,
			sts.Status.CurrentReplicas,
			sts.Status.ReadyReplicas,
			sts.Status.UpdatedReplicas,
		),
	)
	b.WriteString(fmt.Sprintf("Update Strategy:    %s\n", sts.Spec.UpdateStrategy.Type))
	b.WriteString(fmt.Sprintf("Service Name:       %s\n", sts.Spec.ServiceName))

	if len(sts.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range sts.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nPod Template:\n")

	for _, container := range sts.Spec.Template.Spec.Containers {
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

	if len(sts.Spec.VolumeClaimTemplates) > 0 {
		b.WriteString("\nVolume Claim Templates:\n")

		for _, pvc := range sts.Spec.VolumeClaimTemplates {
			b.WriteString(fmt.Sprintf("  Name:          %s\n", pvc.Name))

			if pvc.Spec.StorageClassName != nil {
				b.WriteString(fmt.Sprintf("  StorageClass:  %s\n", *pvc.Spec.StorageClassName))
			}
		}
	}

	return b.String(), nil
}

func (p *StatefulSetsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.statefulsets

		return
	}

	p.filtered = make([]appsv1.StatefulSet, 0)
	for _, sts := range p.statefulsets {
		if strings.Contains(strings.ToLower(sts.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, sts)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *StatefulSetsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type statefulSetsLoadedMsg struct {
	statefulsets []appsv1.StatefulSet
}
