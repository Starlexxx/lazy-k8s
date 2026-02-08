package panels

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

var ErrNoSelection = errors.New("no item selected")

type RefreshMsg struct {
	PanelName string
}

type ErrorMsg struct {
	Error error
}

type StatusMsg struct {
	Message string
}

// Used after mutating operations (scale, restart, etc.) to refresh all panels.
type RefreshAllPanelsMsg struct{}

// Displays a status notification and triggers a full panel refresh when processed.
type StatusWithRefreshMsg struct {
	Message string
}

type ScaleRequestMsg struct {
	DeploymentName  string
	Namespace       string
	CurrentReplicas int32
}

type RollbackRequestMsg struct {
	DeploymentName string
	Namespace      string
}

type PortForwardRequestMsg struct {
	PodName   string
	Namespace string
	Ports     []int32
}

type ExecRequestMsg struct {
	PodName    string
	Namespace  string
	Containers []string
}

type ScaleStatefulSetRequestMsg struct {
	StatefulSetName string
	Namespace       string
	CurrentReplicas int32
}

type EditHPAMinReplicasRequestMsg struct {
	HPAName     string
	Namespace   string
	MinReplicas int32
}

type EditHPAMaxReplicasRequestMsg struct {
	HPAName     string
	Namespace   string
	MaxReplicas int32
}

type PodMetricsMsg struct {
	Metrics map[string]PodMetrics
}

type PodMetrics struct {
	Name      string
	Namespace string
	CPU       int64 // in millicores
	Memory    int64 // in bytes
}

type NodeMetricsMsg struct {
	Metrics map[string]NodeMetrics
}

type NodeMetrics struct {
	Name   string
	CPU    int64 // in millicores
	Memory int64 // in bytes
}

type Panel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Panel, tea.Cmd)
	View() string
	DetailView(width, height int) string
	Title() string
	ShortcutKey() string
	SetSize(width, height int)
	SetFocused(focused bool)
	IsFocused() bool
	SelectedItem() interface{}
	SelectedName() string
	Refresh() tea.Cmd
	Delete() tea.Cmd
	SetFilter(query string)
	SetAllNamespaces(all bool)
	GetSelectedYAML() (string, error)
	GetSelectedDescribe() (string, error)
}

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
