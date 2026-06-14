package panels

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	sigs_yaml "sigs.k8s.io/yaml"
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

type DiffRequestMsg struct {
	DeploymentName string
	Namespace      string
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

// RestartDeploymentRequestMsg is emitted by the deployments panel
// so ui.go can execute the restart and record it in history.
type RestartDeploymentRequestMsg struct {
	DeploymentName string
	Namespace      string
}

// RestartStatefulSetRequestMsg is emitted by the statefulsets panel.
type RestartStatefulSetRequestMsg struct {
	StatefulSetName string
	Namespace       string
}

// RestartDaemonSetRequestMsg is emitted by the daemonsets panel.
type RestartDaemonSetRequestMsg struct {
	DaemonSetName string
	Namespace     string
}

// ToggleSuspendCronJobRequestMsg is emitted by the cronjobs panel.
type ToggleSuspendCronJobRequestMsg struct {
	CronJobName    string
	Namespace      string
	CurrentSuspend bool
}

// TriggerCronJobRequestMsg is emitted by the cronjobs panel.
type TriggerCronJobRequestMsg struct {
	CronJobName string
	Namespace   string
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

// SearchResult represents a single resource match from cross-panel global search.
type SearchResult struct {
	Name      string
	Namespace string
	Kind      string // panel Title(), e.g. "Pods", "Deployments"
	Status    string // resource-specific status string
	PanelIdx  int    // set by ui.go when aggregating results
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
	SelectedItem() any
	SelectedName() string
	Refresh() tea.Cmd
	Delete() tea.Cmd
	SetFilter(query string)
	SetAllNamespaces(all bool)
	GetSelectedYAML() (string, error)
	GetSelectedDescribe() (string, error)
	// SearchItems returns matches from the unfiltered item list without mutating panel state.
	SearchItems(query string) []SearchResult
	// NavigateTo positions the cursor on the item matching name+namespace.
	NavigateTo(name, namespace string) bool
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
	b.cursor = max(maxItems-1, 0)
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

// filterByName returns items whose name contains filter (case-insensitive).
// If filter is empty the original slice is returned unchanged.
// The cursor is clamped to the new length so it never goes out of bounds.
func filterByName[T any](items []T, filter string, name func(T) string, cursor *int) []T {
	if filter == "" {
		return items
	}

	q := strings.ToLower(filter)
	out := make([]T, 0, len(items))

	for _, item := range items {
		if strings.Contains(strings.ToLower(name(item)), q) {
			out = append(out, item)
		}
	}

	if *cursor >= len(out) {
		*cursor = max(len(out)-1, 0)
	}

	return out
}

// visibleWindow computes the start/end indices of the visible slice window.
// extraHeaderRows should be 1 when a header row is rendered above the list
// (pods, nodes), 0 otherwise.
func (b *BasePanel) visibleWindow(total, extraHeaderRows int) (start, end int) {
	visible := max(b.height-3-extraHeaderRows, 1)

	start = 0
	if b.cursor >= visible {
		start = b.cursor - visible + 1
	}

	end = min(start+visible, total)

	return start, end
}

// renderTitle returns the panel title string, optionally with a [key] suffix.
// When shortcutKey is empty the bracket is omitted.
func (b *BasePanel) renderTitle() string {
	if b.shortcutKey == "" {
		return b.title
	}

	return b.title + " [" + b.shortcutKey + "]"
}

// marshalSelectedYAML marshals the item at cursor to YAML.
func marshalSelectedYAML[T any](items []T, cursor int) (string, error) {
	if cursor >= len(items) {
		return "", ErrNoSelection
	}

	data, err := sigs_yaml.Marshal(items[cursor])
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// selectedItem returns a pointer to the item at cursor, or nil.
func selectedItem[T any](items []T, cursor int) *T {
	if cursor >= len(items) {
		return nil
	}

	return &items[cursor]
}

// selectedName returns the name of the item at cursor via the name accessor.
func selectedName[T any](items []T, cursor int, name func(T) string) string {
	if cursor >= len(items) {
		return ""
	}

	return name(items[cursor])
}

// searchByName returns SearchResults for all items whose name contains query.
// status is a per-item callback to compute the Status field.
// namespace is a per-item callback; pass nil for cluster-scoped resources.
func searchByName[T any](
	items []T,
	query, kind string,
	name func(T) string,
	namespace func(T) string,
	status func(T) string,
) []SearchResult {
	if query == "" {
		return nil
	}

	q := strings.ToLower(query)

	var results []SearchResult

	for _, item := range items {
		if strings.Contains(strings.ToLower(name(item)), q) {
			ns := ""
			if namespace != nil {
				ns = namespace(item)
			}

			results = append(results, SearchResult{
				Name:      name(item),
				Namespace: ns,
				Kind:      kind,
				Status:    status(item),
			})
		}
	}

	return results
}

// navigateTo moves cursor to the first item matching name+namespace.
// For cluster-scoped resources pass nil for namespace and "" for targetNamespace.
func navigateTo[T any](
	items []T,
	cursor *int,
	name func(T) string,
	namespace func(T) string,
	targetName, targetNamespace string,
) bool {
	for i, item := range items {
		ns := ""
		if namespace != nil {
			ns = namespace(item)
		}

		if name(item) == targetName && ns == targetNamespace {
			*cursor = i

			return true
		}
	}

	return false
}
