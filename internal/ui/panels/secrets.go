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
	name := utils.Truncate(secret.Name, p.width-15)
	secretType := utils.Truncate(string(secret.Type), 12)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-14)
	line += " " + secretType

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

func (p *SecretsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *SecretsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *SecretsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	secret := p.filtered[p.cursor]

	data, err := yaml.Marshal(secret)
	if err != nil {
		return "", err
	}

	return string(data), nil
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
	if p.filter == "" {
		p.filtered = p.secrets

		return
	}

	p.filtered = make([]corev1.Secret, 0)
	for _, secret := range p.secrets {
		if strings.Contains(strings.ToLower(secret.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, secret)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *SecretsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type secretsLoadedMsg struct {
	secrets []corev1.Secret
}
