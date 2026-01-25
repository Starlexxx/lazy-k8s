package theme

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	// Navigation
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Top       key.Binding
	Bottom    key.Binding
	NextPanel key.Binding
	PrevPanel key.Binding
	Enter     key.Binding
	Back      key.Binding

	// Panels (number keys)
	Panel1 key.Binding
	Panel2 key.Binding
	Panel3 key.Binding
	Panel4 key.Binding
	Panel5 key.Binding
	Panel6 key.Binding
	Panel7 key.Binding
	Panel8 key.Binding
	Panel9 key.Binding

	// Actions
	Quit         key.Binding
	Help         key.Binding
	Refresh      key.Binding
	Search       key.Binding
	ClearSearch  key.Binding
	Delete       key.Binding
	Describe     key.Binding
	Yaml         key.Binding
	Edit         key.Binding
	Logs         key.Binding
	Exec         key.Binding
	PortForward  key.Binding
	Scale        key.Binding
	Restart      key.Binding
	Rollback     key.Binding
	Copy         key.Binding
	Context      key.Binding
	Namespace    key.Binding
	AllNamespace key.Binding
	FollowLogs   key.Binding
}

func NewKeyMap() *KeyMap {
	return &KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "right"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),
		NextPanel: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		PrevPanel: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),

		// Panel shortcuts
		Panel1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "panel 1")),
		Panel2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "panel 2")),
		Panel3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "panel 3")),
		Panel4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "panel 4")),
		Panel5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "panel 5")),
		Panel6: key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "panel 6")),
		Panel7: key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "panel 7")),
		Panel8: key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "panel 8")),
		Panel9: key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "panel 9")),

		// Actions
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear search"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete"),
		),
		Describe: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "describe"),
		),
		Yaml: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "view yaml"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "logs"),
		),
		Exec: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "exec"),
		),
		PortForward: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "port-forward"),
		),
		Scale: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "scale"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		Rollback: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "rollback"),
		),
		Copy: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy yaml"),
		),
		Context: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "switch context"),
		),
		Namespace: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "switch namespace"),
		),
		AllNamespace: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "all namespaces"),
		),
		FollowLogs: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow logs"),
		),
	}
}

// ShortHelp returns key bindings for short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.NextPanel, k.Search}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.NextPanel, k.PrevPanel, k.Top, k.Bottom},
		{k.Enter, k.Back, k.Search, k.Refresh},
		{k.Describe, k.Yaml, k.Logs, k.Exec},
		{k.Delete, k.Scale, k.Restart, k.PortForward},
		{k.Context, k.Namespace, k.Copy, k.Help},
		{k.Quit},
	}
}
