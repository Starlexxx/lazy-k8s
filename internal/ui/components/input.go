package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

type InputSubmitMsg struct {
	Value string
}

type InputCancelMsg struct{}

type Input struct {
	styles      *theme.Styles
	input       textinput.Model
	title       string
	description string
	active      bool
}

func NewInput(styles *theme.Styles) *Input {
	ti := textinput.New()
	ti.CharLimit = 100
	ti.Width = 40

	return &Input{
		styles: styles,
		input:  ti,
	}
}

func (i *Input) Show(title, description, placeholder string) {
	i.title = title
	i.description = description
	i.input.Placeholder = placeholder
	i.input.SetValue("")
	i.input.Focus()
	i.active = true
}

func (i *Input) Hide() {
	i.active = false
	i.input.Blur()
}

func (i *Input) IsActive() bool {
	return i.active
}

func (i *Input) Update(msg tea.Msg) (*Input, tea.Cmd) {
	if !i.active {
		return i, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			value := i.input.Value()
			i.Hide()

			return i, func() tea.Msg {
				return InputSubmitMsg{Value: value}
			}
		case "esc":
			i.Hide()

			return i, func() tea.Msg {
				return InputCancelMsg{}
			}
		}
	}

	var cmd tea.Cmd

	i.input, cmd = i.input.Update(msg)

	return i, cmd
}

func (i *Input) View() string {
	if !i.active {
		return ""
	}

	titleStyle := i.styles.ModalTitle
	descStyle := i.styles.Muted
	hintStyle := i.styles.Muted

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(i.title),
		"",
		descStyle.Render(i.description),
		"",
		i.styles.Input.Render(i.input.View()),
		"",
		hintStyle.Render("enter to submit • esc to cancel"),
	)

	return i.styles.Modal.Width(50).Render(content)
}

func (i *Input) Value() string {
	return i.input.Value()
}

func (i *Input) SetValue(v string) {
	i.input.SetValue(v)
}
