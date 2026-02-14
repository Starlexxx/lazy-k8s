package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type ServiceAccountsPanel struct {
	BasePanel
	client          *k8s.Client
	styles          *theme.Styles
	serviceAccounts []corev1.ServiceAccount
	filtered        []corev1.ServiceAccount
}

func NewServiceAccountsPanel(client *k8s.Client, styles *theme.Styles) *ServiceAccountsPanel {
	return &ServiceAccountsPanel{
		BasePanel: BasePanel{
			title:       "ServiceAccounts",
			shortcutKey: "0",
		},
		client: client,
		styles: styles,
	}
}

func (p *ServiceAccountsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *ServiceAccountsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case serviceAccountsLoadedMsg:
		p.serviceAccounts = msg.serviceAccounts
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *ServiceAccountsPanel) View() string {
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
		sa := p.filtered[i]
		line := p.renderServiceAccountLine(sa, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *ServiceAccountsPanel) renderServiceAccountLine(
	sa corev1.ServiceAccount,
	selected bool,
) string {
	secrets := k8s.GetServiceAccountSecretsSummary(&sa)

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	if p.width > 80 {
		reserved := 35
		if p.width > 120 && p.allNs {
			reserved += 16
		}

		nameW := p.width - reserved
		if nameW < 10 {
			nameW = 10
		}

		line += utils.PadRight(
			utils.Truncate(sa.Name, nameW), nameW,
		)
		line += " " + p.styles.StatusRunning.Render(
			utils.PadRight(secrets, 18),
		)

		age := utils.FormatAgeFromMeta(sa.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)

		if p.width > 120 && p.allNs {
			line += " " + utils.Truncate(sa.Namespace, 15)
		}
	} else {
		name := utils.Truncate(sa.Name, p.width-15)
		line += name
		line = utils.PadRight(line, p.width-25)
		line += " " + p.styles.StatusRunning.Render(secrets)
	}

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *ServiceAccountsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No service account selected"
	}

	sa := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("ServiceAccount: " + sa.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Secrets:"))
	b.WriteString(
		p.styles.DetailValue.Render(fmt.Sprintf("%d", k8s.GetServiceAccountSecretCount(&sa))),
	)
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Image Pull Secrets:"))
	b.WriteString(
		p.styles.DetailValue.Render(
			fmt.Sprintf("%d", k8s.GetServiceAccountImagePullSecretCount(&sa)),
		),
	)
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Automount Token:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetServiceAccountAutoMount(&sa)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(sa.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(sa.Namespace))
		b.WriteString("\n")
	}

	if len(sa.Secrets) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Secrets:"))
		b.WriteString("\n")

		for _, secret := range sa.Secrets {
			b.WriteString(fmt.Sprintf("  %s\n", secret.Name))
		}
	}

	if len(sa.ImagePullSecrets) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Image Pull Secrets:"))
		b.WriteString("\n")

		for _, secret := range sa.ImagePullSecrets {
			b.WriteString(fmt.Sprintf("  %s\n", secret.Name))
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [D]elete"))

	return b.String()
}

func (p *ServiceAccountsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			serviceAccounts []corev1.ServiceAccount
			err             error
		)

		if p.allNs {
			serviceAccounts, err = p.client.ListServiceAccountsAllNamespaces(ctx)
		} else {
			serviceAccounts, err = p.client.ListServiceAccounts(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return serviceAccountsLoadedMsg{serviceAccounts: serviceAccounts}
	}
}

func (p *ServiceAccountsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	sa := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteServiceAccount(ctx, sa.Namespace, sa.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted service account: %s", sa.Name)}
	}
}

func (p *ServiceAccountsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *ServiceAccountsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *ServiceAccountsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	sa := p.filtered[p.cursor]

	data, err := yaml.Marshal(sa)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *ServiceAccountsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	sa := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:                %s\n", sa.Name))
	b.WriteString(fmt.Sprintf("Namespace:           %s\n", sa.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"CreationTimestamp:   %s\n",
			utils.FormatTimestampFromMeta(sa.CreationTimestamp),
		),
	)
	b.WriteString(fmt.Sprintf("Automount Token:     %s\n", k8s.GetServiceAccountAutoMount(&sa)))

	if len(sa.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range sa.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(sa.Annotations) > 0 {
		b.WriteString("\nAnnotations:\n")

		for k, v := range sa.Annotations {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(sa.Secrets) > 0 {
		b.WriteString("\nMountable Secrets:\n")

		for _, secret := range sa.Secrets {
			b.WriteString(fmt.Sprintf("  %s\n", secret.Name))
		}
	}

	if len(sa.ImagePullSecrets) > 0 {
		b.WriteString("\nImage Pull Secrets:\n")

		for _, secret := range sa.ImagePullSecrets {
			b.WriteString(fmt.Sprintf("  %s\n", secret.Name))
		}
	}

	return b.String(), nil
}

func (p *ServiceAccountsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.serviceAccounts

		return
	}

	p.filtered = make([]corev1.ServiceAccount, 0)
	for _, sa := range p.serviceAccounts {
		if strings.Contains(strings.ToLower(sa.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, sa)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *ServiceAccountsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type serviceAccountsLoadedMsg struct {
	serviceAccounts []corev1.ServiceAccount
}
