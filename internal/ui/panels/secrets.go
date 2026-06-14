package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type SecretsPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	secrets  []corev1.Secret
	filtered []corev1.Secret
}

func NewSecretsPanel(client *k8s.Client, styles *theme.Styles) *SecretsPanel {
	return &SecretsPanel{
		BasePanel: BasePanel{
			title:       "Secrets",
			shortcutKey: "6",
		},
		client: client,
		styles: styles,
	}
}

func (p *SecretsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *SecretsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case secretsLoadedMsg:
		p.secrets = msg.secrets
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *SecretsPanel) View() string {
	var b strings.Builder

	title := p.renderTitle()
	if p.focused {
		b.WriteString(p.styles.PanelTitleActive.Render(title))
	} else {
		b.WriteString(p.styles.PanelTitle.Render(title))
	}

	b.WriteString("\n")

	startIdx, endIdx := p.visibleWindow(len(p.filtered), 0)

	for i := startIdx; i < endIdx; i++ {
		secret := p.filtered[i]
		line := p.renderSecretLine(secret, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *SecretsPanel) renderSecretLine(secret corev1.Secret, selected bool) string {
	secretType := utils.Truncate(string(secret.Type), 12)

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	if p.width > 80 {
		reserved := 30
		if p.width > 120 && p.allNs {
			reserved += 16
		}

		nameW := max(p.width-reserved, 10)

		line += utils.PadRight(
			utils.Truncate(secret.Name, nameW), nameW,
		)
		line += " " + utils.PadRight(secretType, 12)

		age := utils.FormatAgeFromMeta(secret.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)

		if p.width > 120 && p.allNs {
			line += " " + utils.Truncate(secret.Namespace, 15)
		}
	} else {
		name := utils.Truncate(secret.Name, p.width-15)
		line += name
		line = utils.PadRight(line, p.width-14)
		line += " " + secretType
	}

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *SecretsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No secret selected"
	}

	secret := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Secret: " + secret.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Type:"))
	b.WriteString(p.styles.DetailValue.Render(string(secret.Type)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Data Keys:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", len(secret.Data))))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(secret.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(secret.Namespace))
		b.WriteString("\n")
	}

	// Data keys (values hidden for security)
	if len(secret.Data) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Data Keys:"))
		b.WriteString("\n")

		for k, v := range secret.Data {
			b.WriteString(fmt.Sprintf("  %s: %d bytes\n", k, len(v)))
		}
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [D]elete"))
	b.WriteString("\n")
	b.WriteString(p.styles.StatusWarning.Render("Note: Secret values are base64 encoded in YAML"))

	return b.String()
}

func (p *SecretsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			secrets []corev1.Secret
			err     error
		)

		if p.allNs {
			secrets, err = p.client.ListSecretsAllNamespaces(ctx)
		} else {
			secrets, err = p.client.ListSecrets(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return secretsLoadedMsg{secrets: secrets}
	}
}

func (p *SecretsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	secret := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteSecret(ctx, secret.Namespace, secret.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted secret: %s", secret.Name)}
	}
}

func (p *SecretsPanel) SelectedItem() any {
	item := selectedItem(p.filtered, p.cursor)
	if item == nil {
		return nil
	}

	return item
}

func (p *SecretsPanel) SelectedName() string {
	return selectedName(p.filtered, p.cursor, func(s corev1.Secret) string {
		return s.Name
	})
}

func (p *SecretsPanel) GetSelectedYAML() (string, error) {
	return marshalSelectedYAML(p.filtered, p.cursor)
}

func (p *SecretsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	secret := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:         %s\n", secret.Name))
	b.WriteString(fmt.Sprintf("Namespace:    %s\n", secret.Namespace))
	b.WriteString(fmt.Sprintf("Type:         %s\n", secret.Type))
	b.WriteString(
		fmt.Sprintf("Age:          %s\n", utils.FormatAgeFromMeta(secret.CreationTimestamp)),
	)

	if len(secret.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range secret.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	if len(secret.Data) > 0 {
		b.WriteString("\nData:\n")

		for k, v := range secret.Data {
			b.WriteString(fmt.Sprintf("  %s: %d bytes\n", k, len(v)))
		}
	}

	return b.String(), nil
}

func (p *SecretsPanel) applyFilter() {
	p.filtered = filterByName(
		p.secrets,
		p.filter,
		func(s corev1.Secret) string { return s.Name },
		&p.cursor,
	)
}

func (p *SecretsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

func (p *SecretsPanel) SearchItems(query string) []SearchResult {
	return searchByName(
		p.secrets,
		query,
		p.title,
		func(s corev1.Secret) string { return s.Name },
		func(s corev1.Secret) string { return s.Namespace },
		func(s corev1.Secret) string { return string(s.Type) },
	)
}

func (p *SecretsPanel) NavigateTo(name, namespace string) bool {
	return navigateTo(
		p.filtered,
		&p.cursor,
		func(s corev1.Secret) string { return s.Name },
		func(s corev1.Secret) string { return s.Namespace },
		name,
		namespace,
	)
}

type secretsLoadedMsg struct {
	secrets []corev1.Secret
}
