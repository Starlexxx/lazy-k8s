package panels

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

var ErrNoSelection = errors.New("no item selected")

// RefreshMsg is sent to request a panel refresh.
type RefreshMsg struct {
	PanelName string
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Error error
}

// StatusMsg is sent to display a status message.
type StatusMsg struct {
	Message string
}

// ScaleRequestMsg is sent when user requests to scale a deployment.
type ScaleRequestMsg struct {
	DeploymentName  string
	Namespace       string
	CurrentReplicas int32
}

// RollbackRequestMsg is sent when user requests to rollback a deployment.
type RollbackRequestMsg struct {
	DeploymentName string
	Namespace      string
}

// PortForwardRequestMsg is sent when user requests to port-forward to a pod.
type PortForwardRequestMsg struct {
	PodName   string
	Namespace string
	Ports     []int32
}

// ExecRequestMsg is sent when user requests to exec into a pod.
type ExecRequestMsg struct {
	PodName    string
	Namespace  string
	Containers []string
}

// Panel is the interface that all resource panels must implement.
type Panel interface {
	// Init initializes the panel and returns initial commands
	Init() tea.Cmd

	// Update handles messages and returns updated panel and commands
	Update(msg tea.Msg) (Panel, tea.Cmd)

	// View renders the panel's list view
	View() string

	// DetailView renders the detail view for the selected item
	DetailView(width, height int) string

	// Title returns the panel's title
	Title() string

	// ShortcutKey returns the keyboard shortcut for this panel
	ShortcutKey() string

	// SetSize sets the panel dimensions
	SetSize(width, height int)

	// SetFocused sets whether this panel is focused
	SetFocused(focused bool)

	// IsFocused returns whether this panel is focused
	IsFocused() bool

	// SelectedItem returns the currently selected item
	SelectedItem() interface{}

	// SelectedName returns the name of the selected item
	SelectedName() string

	// Refresh triggers a data refresh
	Refresh() tea.Cmd

	// Delete deletes the selected item
	Delete() tea.Cmd

	// SetFilter sets the search/filter query
	SetFilter(query string)

	// SetAllNamespaces sets whether to show all namespaces
	SetAllNamespaces(all bool)

	// GetSelectedYAML returns the YAML representation of the selected item
	GetSelectedYAML() (string, error)

	// GetSelectedDescribe returns the describe output for the selected item
	GetSelectedDescribe() (string, error)
}

// BasePanel provides common functionality for panels.
type BasePanel struct {
	title       string
	shortcutKey string
	width       int
	height      int
	focused     bool
	filter      string
	allNs       bool
	cursor      int
}

func (b *BasePanel) Title() string {
	return b.title
}

func (b *BasePanel) ShortcutKey() string {
	return b.shortcutKey
}

func (b *BasePanel) SetSize(width, height int) {
	b.width = width
	b.height = height
}

func (b *BasePanel) SetFocused(focused bool) {
	b.focused = focused
}

func (b *BasePanel) IsFocused() bool {
	return b.focused
}

func (b *BasePanel) SetFilter(query string) {
	b.filter = query
}

func (b *BasePanel) SetAllNamespaces(all bool) {
	b.allNs = all
}

func (b *BasePanel) MoveUp() {
	if b.cursor > 0 {
		b.cursor--
	}
}

func (b *BasePanel) MoveDown(maxItems int) {
	if b.cursor < maxItems-1 {
		b.cursor++
	}
}

func (b *BasePanel) MoveToTop() {
	b.cursor = 0
}

func (b *BasePanel) MoveToBottom(maxItems int) {
	b.cursor = maxItems - 1
	if b.cursor < 0 {
		b.cursor = 0
	}
}

func (b *BasePanel) Cursor() int {
	return b.cursor
}

func (b *BasePanel) SetCursor(cursor int) {
	b.cursor = cursor
}

func (b *BasePanel) Width() int {
	return b.width
}

func (b *BasePanel) Height() int {
	return b.height
}

func (b *BasePanel) Filter() string {
	return b.filter
}

func (b *BasePanel) AllNamespaces() bool {
	return b.allNs
}
