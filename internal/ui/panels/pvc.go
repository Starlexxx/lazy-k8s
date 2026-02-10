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

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type PVCPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	pvcs     []corev1.PersistentVolumeClaim
	filtered []corev1.PersistentVolumeClaim
}

func NewPVCPanel(client *k8s.Client, styles *theme.Styles) *PVCPanel {
	return &PVCPanel{
		BasePanel: BasePanel{
			title:       "PersistentVolumeClaims",
			shortcutKey: "V",
		},
		client: client,
		styles: styles,
	}
}

func (p *PVCPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *PVCPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case pvcLoadedMsg:
		p.pvcs = msg.pvcs
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *PVCPanel) View() string {
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
		pvc := p.filtered[i]
		line := p.renderPVCLine(pvc, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *PVCPanel) renderPVCLine(pvc corev1.PersistentVolumeClaim, selected bool) string {
	name := utils.Truncate(pvc.Name, p.width-15)
	status := string(pvc.Status.Phase)

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

func (p *PVCPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No PVC selected"
	}

	pvc := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("PersistentVolumeClaim: " + pvc.Name))
	b.WriteString("\n\n")

	status := string(pvc.Status.Phase)

	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(status).Render(status))
	b.WriteString("\n")

	// Capacity (if bound)
	if len(pvc.Status.Capacity) > 0 {
		capacity := pvc.Status.Capacity[corev1.ResourceStorage]

		b.WriteString(p.styles.DetailLabel.Render("Capacity:"))
		b.WriteString(p.styles.DetailValue.Render(capacity.String()))
		b.WriteString("\n")
	}

	if pvc.Spec.Resources.Requests != nil {
		requested := pvc.Spec.Resources.Requests[corev1.ResourceStorage]

		b.WriteString(p.styles.DetailLabel.Render("Requested:"))
		b.WriteString(p.styles.DetailValue.Render(requested.String()))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Access Modes:"))

	modes := make([]string, 0)
	for _, mode := range pvc.Spec.AccessModes {
		modes = append(modes, string(mode))
	}

	b.WriteString(p.styles.DetailValue.Render(strings.Join(modes, ", ")))
	b.WriteString("\n")

	if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName != "" {
		b.WriteString(p.styles.DetailLabel.Render("Storage Class:"))
		b.WriteString(p.styles.DetailValue.Render(*pvc.Spec.StorageClassName))
		b.WriteString("\n")
	}

	if pvc.Spec.VolumeName != "" {
		b.WriteString(p.styles.DetailLabel.Render("Volume:"))
		b.WriteString(p.styles.DetailValue.Render(pvc.Spec.VolumeName))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(pvc.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(pvc.Namespace))
		b.WriteString("\n")
	}

	if len(pvc.Status.Conditions) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Conditions:"))
		b.WriteString("\n")

		for _, cond := range pvc.Status.Conditions {
			b.WriteString(fmt.Sprintf("  %s: %s\n", cond.Type, cond.Status))

			if cond.Message != "" {
				b.WriteString(fmt.Sprintf("    %s\n", cond.Message))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *PVCPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			pvcs *corev1.PersistentVolumeClaimList
			err  error
		)

		if p.allNs {
			pvcs, err = p.client.Clientset().
				CoreV1().
				PersistentVolumeClaims("").
				List(ctx, metav1.ListOptions{})
		} else {
			pvcs, err = p.client.Clientset().
				CoreV1().
				PersistentVolumeClaims(p.client.CurrentNamespace()).
				List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return pvcLoadedMsg{pvcs: pvcs.Items}
	}
}

func (p *PVCPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	pvc := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.Clientset().
			CoreV1().
			PersistentVolumeClaims(pvc.Namespace).
			Delete(ctx, pvc.Name, metav1.DeleteOptions{})
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted PVC: %s", pvc.Name)}
	}
}

func (p *PVCPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *PVCPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *PVCPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	pvc := p.filtered[p.cursor]

	data, err := yaml.Marshal(pvc)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *PVCPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	pvc := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:          %s\n", pvc.Name))
	b.WriteString(fmt.Sprintf("Namespace:     %s\n", pvc.Namespace))
	b.WriteString(fmt.Sprintf("Status:        %s\n", pvc.Status.Phase))

	if len(pvc.Status.Capacity) > 0 {
		capacity := pvc.Status.Capacity[corev1.ResourceStorage]
		b.WriteString(fmt.Sprintf("Capacity:      %s\n", capacity.String()))
	}

	modes := make([]string, 0)
	for _, mode := range pvc.Spec.AccessModes {
		modes = append(modes, string(mode))
	}

	b.WriteString(fmt.Sprintf("Access Modes:  %s\n", strings.Join(modes, ", ")))

	if pvc.Spec.StorageClassName != nil {
		b.WriteString(fmt.Sprintf("Storage Class: %s\n", *pvc.Spec.StorageClassName))
	}

	b.WriteString(fmt.Sprintf("Volume:        %s\n", pvc.Spec.VolumeName))

	if len(pvc.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range pvc.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	return b.String(), nil
}

func (p *PVCPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.pvcs

		return
	}

	p.filtered = make([]corev1.PersistentVolumeClaim, 0)
	for _, pvc := range p.pvcs {
		if strings.Contains(strings.ToLower(pvc.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, pvc)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *PVCPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type pvcLoadedMsg struct {
	pvcs []corev1.PersistentVolumeClaim
}
