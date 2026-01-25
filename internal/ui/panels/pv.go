package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type PVPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	pvs      []corev1.PersistentVolume
	filtered []corev1.PersistentVolume
}

func NewPVPanel(client *k8s.Client, styles *theme.Styles) *PVPanel {
	return &PVPanel{
		BasePanel: BasePanel{
			title:       "PersistentVolumes",
			shortcutKey: "v",
		},
		client: client,
		styles: styles,
	}
}

func (p *PVPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *PVPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case pvLoadedMsg:
		p.pvs = msg.pvs
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *PVPanel) View() string {
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
		pv := p.filtered[i]
		line := p.renderPVLine(pv, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *PVPanel) renderPVLine(pv corev1.PersistentVolume, selected bool) string {
	name := utils.Truncate(pv.Name, p.width-15)
	status := string(pv.Status.Phase)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

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

func (p *PVPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No PersistentVolume selected"
	}

	pv := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("PersistentVolume: " + pv.Name))
	b.WriteString("\n\n")

	status := string(pv.Status.Phase)

	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(status).Render(status))
	b.WriteString("\n")

	capacity := pv.Spec.Capacity[corev1.ResourceStorage]

	b.WriteString(p.styles.DetailLabel.Render("Capacity:"))
	b.WriteString(p.styles.DetailValue.Render(capacity.String()))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Access Modes:"))

	modes := make([]string, 0)
	for _, mode := range pv.Spec.AccessModes {
		modes = append(modes, string(mode))
	}

	b.WriteString(p.styles.DetailValue.Render(strings.Join(modes, ", ")))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Reclaim Policy:"))
	b.WriteString(p.styles.DetailValue.Render(string(pv.Spec.PersistentVolumeReclaimPolicy)))
	b.WriteString("\n")

	if pv.Spec.StorageClassName != "" {
		b.WriteString(p.styles.DetailLabel.Render("Storage Class:"))
		b.WriteString(p.styles.DetailValue.Render(pv.Spec.StorageClassName))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(pv.CreationTimestamp)))
	b.WriteString("\n")

	// Claim
	if pv.Spec.ClaimRef != nil {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Claim:"))
		b.WriteString("\n")
		b.WriteString(p.styles.DetailLabel.Render("Name:"))
		b.WriteString(p.styles.DetailValue.Render(pv.Spec.ClaimRef.Name))
		b.WriteString("\n")
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(pv.Spec.ClaimRef.Namespace))
		b.WriteString("\n")
	}

	// Source
	b.WriteString("\n")
	b.WriteString(p.styles.DetailTitle.Render("Source:"))
	b.WriteString("\n")

	if pv.Spec.HostPath != nil {
		b.WriteString(p.styles.DetailLabel.Render("HostPath:"))
		b.WriteString(p.styles.DetailValue.Render(pv.Spec.HostPath.Path))
		b.WriteString("\n")
	} else if pv.Spec.NFS != nil {
		b.WriteString(p.styles.DetailLabel.Render("NFS:"))
		b.WriteString(
			p.styles.DetailValue.Render(fmt.Sprintf("%s:%s", pv.Spec.NFS.Server, pv.Spec.NFS.Path)),
		)
		b.WriteString("\n")
	} else if pv.Spec.CSI != nil {
		b.WriteString(p.styles.DetailLabel.Render("CSI Driver:"))
		b.WriteString(p.styles.DetailValue.Render(pv.Spec.CSI.Driver))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *PVPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		pvs, err := p.client.Clientset().
			CoreV1().
			PersistentVolumes().
			List(ctx, metav1.ListOptions{})
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return pvLoadedMsg{pvs: pvs.Items}
	}
}

func (p *PVPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	pv := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.Clientset().
			CoreV1().
			PersistentVolumes().
			Delete(ctx, pv.Name, metav1.DeleteOptions{})
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted PV: %s", pv.Name)}
	}
}

func (p *PVPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *PVPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *PVPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	pv := p.filtered[p.cursor]

	data, err := yaml.Marshal(pv)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *PVPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	pv := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:            %s\n", pv.Name))
	b.WriteString(fmt.Sprintf("Status:          %s\n", pv.Status.Phase))

	capacity := pv.Spec.Capacity[corev1.ResourceStorage]
	b.WriteString(fmt.Sprintf("Capacity:        %s\n", capacity.String()))

	modes := make([]string, 0)
	for _, mode := range pv.Spec.AccessModes {
		modes = append(modes, string(mode))
	}

	b.WriteString(fmt.Sprintf("Access Modes:    %s\n", strings.Join(modes, ", ")))
	b.WriteString(fmt.Sprintf("Reclaim Policy:  %s\n", pv.Spec.PersistentVolumeReclaimPolicy))
	b.WriteString(fmt.Sprintf("Storage Class:   %s\n", pv.Spec.StorageClassName))

	if pv.Spec.ClaimRef != nil {
		b.WriteString(
			fmt.Sprintf(
				"Claim:           %s/%s\n",
				pv.Spec.ClaimRef.Namespace,
				pv.Spec.ClaimRef.Name,
			),
		)
	}

	if len(pv.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range pv.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	return b.String(), nil
}

func (p *PVPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.pvs

		return
	}

	p.filtered = make([]corev1.PersistentVolume, 0)
	for _, pv := range p.pvs {
		if strings.Contains(strings.ToLower(pv.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, pv)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *PVPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type pvLoadedMsg struct {
	pvs []corev1.PersistentVolume
}
