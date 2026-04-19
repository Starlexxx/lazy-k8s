package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/Starlexxx/lazy-k8s/internal/config"
	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/components"
	"github.com/Starlexxx/lazy-k8s/internal/ui/panels"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

var (
	ErrInvalidReplicaCount = errors.New("replica count must be non-negative")
	ErrInvalidPortFormat   = errors.New("invalid port format, use local:remote or port")
	ErrMinReplicasTooLow   = errors.New("min replicas must be at least 1")
	ErrMaxReplicasTooLow   = errors.New("max replicas must be at least 1")
	ErrApplyFailed         = errors.New("kubectl apply failed")
)

type ViewMode int

const (
	ViewNormal ViewMode = iota
	ViewHelp
	ViewYaml
	ViewLogs
	ViewConfirm
	ViewInput
	ViewContextSwitch
	ViewNamespaceSwitch
	ViewContainerSelect
	ViewDiff
	ViewGlobalSearch
	ViewHistory
)

// borderLines is the number of lines used by panel borders (top + bottom).
const borderLines = 2

const metricsRefreshInterval = 10 * time.Second

type metricsTickMsg time.Time

type Model struct {
	// Core dependencies
	k8sClient *k8s.Client
	config    *config.Config
	styles    *theme.Styles
	keys      *theme.KeyMap

	// Dimensions
	width  int
	height int

	// Panels
	panels         []panels.Panel
	activePanelIdx int

	// Components
	header    *components.Header
	statusBar *components.StatusBar
	help      *components.Help
	confirm   *components.Confirm
	yamlView  *components.YamlViewer
	logView   *components.LogViewer
	diffView  *components.DiffViewer
	search    *components.Search
	input     *components.Input

	pendingInputAction func(value string) tea.Cmd

	// State
	viewMode     ViewMode
	lastError    string
	lastStatus   string
	showAllNs    bool
	zoomed       bool
	searchActive bool
	searchQuery  string

	// Context/namespace switching
	contextList        []string
	namespaceList      []string
	selectIdx          int
	switchFilter       string
	switchFiltered     []string
	switchFilterActive bool

	// Port forwarding
	portForwards map[string]*k8s.PortForwarder

	// Exec container selection
	execContainers []string
	execPodName    string
	execNamespace  string

	// Global search
	globalSearch        *components.GlobalSearch
	globalSearchResults []panels.SearchResult

	// Metrics
	metricsClient *k8s.MetricsClient

	// Operations history
	historyStore *components.HistoryStore
	historyView  *components.HistoryViewer
}

func NewModel(client *k8s.Client, cfg *config.Config) *Model {
	styles := theme.NewStyles(&cfg.Theme)
	keys := theme.NewKeyMap()

	m := &Model{
		k8sClient:    client,
		config:       cfg,
		styles:       styles,
		keys:         keys,
		viewMode:     ViewNormal,
		portForwards: make(map[string]*k8s.PortForwarder),
	}

	m.header = components.NewHeader(styles, client.CurrentContext(), client.CurrentNamespace())
	m.statusBar = components.NewStatusBar(styles)
	m.help = components.NewHelp(styles, keys)
	m.confirm = components.NewConfirm(styles)
	m.yamlView = components.NewYamlViewer(styles)
	m.logView = components.NewLogViewer(styles)
	m.diffView = components.NewDiffViewer(styles)
	m.search = components.NewSearch(styles)
	m.input = components.NewInput(styles)
	m.globalSearch = components.NewGlobalSearch(styles)
	m.historyStore = components.NewHistoryStore()
	m.historyView = components.NewHistoryViewer(styles, m.historyStore)

	// Initialize metrics client (optional - may fail if metrics-server not installed)
	metricsClient, err := client.NewMetricsClient()
	if err == nil {
		m.metricsClient = metricsClient
	}

	m.initPanels()

	return m
}

func (m *Model) initPanels() {
	m.panels = make([]panels.Panel, 0)

	for _, panelName := range m.config.Panels.Visible {
		switch panelName {
		case "namespaces":
			m.panels = append(m.panels, panels.NewNamespacesPanel(m.k8sClient, m.styles))
		case "pods":
			m.panels = append(m.panels, panels.NewPodsPanel(m.k8sClient, m.styles))
		case "deployments":
			m.panels = append(m.panels, panels.NewDeploymentsPanel(m.k8sClient, m.styles))
		case "services":
			m.panels = append(m.panels, panels.NewServicesPanel(m.k8sClient, m.styles))
		case "configmaps":
			m.panels = append(m.panels, panels.NewConfigMapsPanel(m.k8sClient, m.styles))
		case "secrets":
			m.panels = append(m.panels, panels.NewSecretsPanel(m.k8sClient, m.styles))
		case "nodes":
			m.panels = append(m.panels, panels.NewNodesPanel(m.k8sClient, m.styles))
		case "events":
			m.panels = append(m.panels, panels.NewEventsPanel(m.k8sClient, m.styles))
		case "jobs":
			m.panels = append(m.panels, panels.NewJobsPanel(m.k8sClient, m.styles))
		case "ingress", "ingresses":
			m.panels = append(m.panels, panels.NewIngressPanel(m.k8sClient, m.styles))
		case "pv", "persistentvolumes":
			m.panels = append(m.panels, panels.NewPVPanel(m.k8sClient, m.styles))
		case "pvc", "persistentvolumeclaims":
			m.panels = append(m.panels, panels.NewPVCPanel(m.k8sClient, m.styles))
		case "statefulsets", "sts":
			m.panels = append(m.panels, panels.NewStatefulSetsPanel(m.k8sClient, m.styles))
		case "daemonsets", "ds":
			m.panels = append(m.panels, panels.NewDaemonSetsPanel(m.k8sClient, m.styles))
		case "cronjobs", "cj":
			m.panels = append(m.panels, panels.NewCronJobsPanel(m.k8sClient, m.styles))
		case "hpa", "horizontalpodautoscalers":
			m.panels = append(m.panels, panels.NewHPAPanel(m.k8sClient, m.styles))
		case "networkpolicies", "netpol":
			m.panels = append(m.panels, panels.NewNetworkPoliciesPanel(m.k8sClient, m.styles))
		case "serviceaccounts", "sa":
			m.panels = append(m.panels, panels.NewServiceAccountsPanel(m.k8sClient, m.styles))
		}
	}

	if len(m.panels) > 0 {
		m.panels[0].SetFocused(true)
	}
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, panel := range m.panels {
		cmds = append(cmds, panel.Init())
	}

	if m.metricsClient != nil {
		cmds = append(cmds, m.fetchMetrics(), m.metricsTickCmd())
	}

	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()

		return m, nil

	case tea.KeyMsg:
		// Handle view-specific keys first
		switch m.viewMode {
		case ViewHelp:
			if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Help) ||
				msg.String() == "q" {
				m.viewMode = ViewNormal

				return m, nil
			}

			return m, nil

		case ViewYaml:
			if key.Matches(msg, m.keys.Back) {
				m.viewMode = ViewNormal

				return m, nil
			}

			var cmd tea.Cmd

			m.yamlView, cmd = m.yamlView.Update(msg)

			return m, cmd

		case ViewLogs:
			if key.Matches(msg, m.keys.Back) {
				m.viewMode = ViewNormal
				m.logView.Stop()

				return m, nil
			}

			var cmd tea.Cmd

			m.logView, cmd = m.logView.Update(msg)

			return m, cmd

		case ViewDiff:
			if key.Matches(msg, m.keys.Back) {
				m.viewMode = ViewNormal

				return m, nil
			}

			var cmd tea.Cmd

			m.diffView, cmd = m.diffView.Update(msg)

			return m, cmd

		case ViewConfirm:
			var cmd tea.Cmd

			m.confirm, cmd = m.confirm.Update(msg)
			if m.confirm.Done() {
				m.viewMode = ViewNormal
				if m.confirm.Confirmed() {
					return m, m.confirm.Action()
				}
			}

			return m, cmd

		case ViewContextSwitch:
			return m.handleContextSwitch(msg)

		case ViewNamespaceSwitch:
			return m.handleNamespaceSwitch(msg)

		case ViewContainerSelect:
			return m.handleContainerSelect(msg)

		case ViewInput:
			var cmd tea.Cmd

			m.input, cmd = m.input.Update(msg)

			return m, cmd

		case ViewGlobalSearch:
			return m.handleGlobalSearch(msg)

		case ViewHistory:
			if key.Matches(msg, m.keys.Back) {
				m.viewMode = ViewNormal

				return m, nil
			}

			var cmd tea.Cmd

			m.historyView, cmd = m.historyView.Update(msg)

			return m, cmd

		case ViewNormal:
			// Fall through to normal key handling below
		}

		if m.zoomed && !m.searchActive && key.Matches(msg, m.keys.Back) {
			m.zoomed = false
			m.updatePanelSizes()

			return m, nil
		}

		// Search mode
		if m.searchActive {
			if key.Matches(msg, m.keys.Back) {
				m.searchActive = false
				m.searchQuery = ""
				m.search.Clear()

				return m, nil
			}

			var cmd tea.Cmd

			m.search, cmd = m.search.Update(msg)
			m.searchQuery = m.search.Value()

			if len(m.panels) > m.activePanelIdx {
				m.panels[m.activePanelIdx].SetFilter(m.searchQuery)
			}

			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.stopAllPortForwards()

			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.viewMode = ViewHelp

			return m, nil

		case key.Matches(msg, m.keys.GlobalSearch):
			m.globalSearch.Reset()
			m.globalSearchResults = nil
			m.viewMode = ViewGlobalSearch

			return m, nil

		case key.Matches(msg, m.keys.History):
			m.historyView.Reset()
			m.viewMode = ViewHistory

			return m, nil

		case key.Matches(msg, m.keys.Search):
			m.searchActive = true
			m.search.Focus()

			return m, nil

		case key.Matches(msg, m.keys.Zoom):
			m.zoomed = !m.zoomed
			m.updatePanelSizes()

			return m, nil

		case key.Matches(msg, m.keys.NextPanel):
			m.nextPanel()

			return m, nil

		case key.Matches(msg, m.keys.PrevPanel):
			m.prevPanel()

			return m, nil

		case key.Matches(msg, m.keys.Context):
			return m.startContextSwitch()

		case key.Matches(msg, m.keys.Namespace):
			return m.startNamespaceSwitch()

		case key.Matches(msg, m.keys.Refresh):
			m.statusBar.SetMessage("Refreshed all panels")

			return m, m.refreshAllPanels()

		case key.Matches(msg, m.keys.AllNamespace):
			m.showAllNs = !m.showAllNs
			m.statusBar.SetMessage(fmt.Sprintf("All namespaces: %v", m.showAllNs))

			for _, panel := range m.panels {
				panel.SetAllNamespaces(m.showAllNs)
				cmds = append(cmds, panel.Refresh())
			}

			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.Panel1):
			m.selectPanel(0)
		case key.Matches(msg, m.keys.Panel2):
			m.selectPanel(1)
		case key.Matches(msg, m.keys.Panel3):
			m.selectPanel(2)
		case key.Matches(msg, m.keys.Panel4):
			m.selectPanel(3)
		case key.Matches(msg, m.keys.Panel5):
			m.selectPanel(4)
		case key.Matches(msg, m.keys.Panel6):
			m.selectPanel(5)
		case key.Matches(msg, m.keys.Panel7):
			m.selectPanel(6)
		case key.Matches(msg, m.keys.Panel8):
			m.selectPanel(7)
		case key.Matches(msg, m.keys.Panel9):
			m.selectPanel(8)

		case key.Matches(msg, m.keys.Yaml):
			return m.showYaml()

		case key.Matches(msg, m.keys.Logs):
			return m.showLogs()

		case key.Matches(msg, m.keys.Delete):
			return m.confirmDelete()

		case key.Matches(msg, m.keys.Describe):
			return m.showDescribe()

		case key.Matches(msg, m.keys.CopyName):
			return m.copyNameToClipboard()

		case key.Matches(msg, m.keys.Copy):
			return m.copyYamlToClipboard()

		case key.Matches(msg, m.keys.Edit):
			return m.editResource()

		default:
			if len(m.panels) > m.activePanelIdx {
				panel, cmd := m.panels[m.activePanelIdx].Update(msg)
				m.panels[m.activePanelIdx] = panel

				return m, cmd
			}
		}

	case components.LogLineMsg:
		var cmd tea.Cmd

		m.logView, cmd = m.logView.Update(msg)

		return m, cmd

	case panels.RefreshMsg:
		for i, panel := range m.panels {
			if panel.Title() == msg.PanelName {
				newPanel, cmd := panel.Update(msg)
				m.panels[i] = newPanel

				cmds = append(cmds, cmd)
			}
		}

		return m, tea.Batch(cmds...)

	case panels.ErrorMsg:
		m.lastError = msg.Error.Error()
		m.statusBar.SetError(m.lastError)

		return m, nil

	case panels.StatusMsg:
		m.lastStatus = msg.Message
		m.statusBar.SetMessage(m.lastStatus)

		return m, nil

	case panels.RefreshAllPanelsMsg:
		return m, m.refreshAllPanels()

	case panels.StatusWithRefreshMsg:
		m.lastStatus = msg.Message
		m.statusBar.SetMessage(m.lastStatus)
		m.header.SetNamespace(m.k8sClient.CurrentNamespace())

		return m, m.refreshAllPanels()

	case metricsTickMsg:
		return m, tea.Batch(m.metricsTickCmd(), m.fetchMetrics())

	case metricsLoadedMsg:
		var cmds []tea.Cmd

		for _, panel := range m.panels {
			if panel.Title() == "Pods" {
				_, cmd := panel.Update(panels.PodMetricsMsg{Metrics: msg.podMetrics})
				cmds = append(cmds, cmd)
			} else if panel.Title() == "Nodes" {
				_, cmd := panel.Update(panels.NodeMetricsMsg{Metrics: msg.nodeMetrics})
				cmds = append(cmds, cmd)
			}
		}

		return m, tea.Batch(cmds...)

	case components.InputSubmitMsg:
		m.viewMode = ViewNormal
		if m.pendingInputAction != nil {
			action := m.pendingInputAction
			m.pendingInputAction = nil

			return m, action(msg.Value)
		}

		return m, nil

	case components.InputCancelMsg:
		m.viewMode = ViewNormal
		m.pendingInputAction = nil

		return m, nil

	case panels.ScaleRequestMsg:
		description := fmt.Sprintf(
			"Enter new replica count for %s (current: %d)",
			msg.DeploymentName,
			msg.CurrentReplicas,
		)
		currentReplicas := msg.CurrentReplicas

		m.showInput(
			"Scale Deployment",
			description,
			strconv.Itoa(int(msg.CurrentReplicas)),
			func(value string) tea.Cmd {
				return m.scaleDeployment(
					msg.Namespace, msg.DeploymentName,
					value, currentReplicas,
				)
			},
		)
		m.input.SetValue(strconv.Itoa(int(msg.CurrentReplicas)))

		return m, nil

	case panels.RollbackRequestMsg:
		description := fmt.Sprintf(
			"Are you sure you want to rollback %s to the previous revision?",
			msg.DeploymentName,
		)
		m.confirm.Show(
			fmt.Sprintf("Rollback %s?", msg.DeploymentName),
			description,
			func() tea.Cmd {
				return m.rollbackDeployment(msg.Namespace, msg.DeploymentName)
			},
		)
		m.viewMode = ViewConfirm

		return m, nil

	case panels.DiffRequestMsg:
		return m, m.loadRevisionDiff(msg.Namespace, msg.DeploymentName)

	case diffLoadedMsg:
		m.diffView.SetContent(msg.title, msg.oldYAML, msg.newYAML)
		m.viewMode = ViewDiff

		return m, nil

	case panels.PortForwardRequestMsg:
		if len(msg.Ports) == 0 {
			m.statusBar.SetMessage("No ports exposed on this pod")

			return m, nil
		}

		defaultPort := msg.Ports[0]
		m.showInput(
			"Port Forward",
			fmt.Sprintf("Enter ports as local:remote (e.g., 8080:%d)", defaultPort),
			fmt.Sprintf("%d:%d", defaultPort, defaultPort),
			func(value string) tea.Cmd {
				return m.startPortForward(msg.Namespace, msg.PodName, value)
			},
		)
		m.input.SetValue(fmt.Sprintf("%d:%d", defaultPort, defaultPort))

		return m, nil

	case panels.ExecRequestMsg:
		if len(msg.Containers) == 0 {
			m.statusBar.SetError("No containers in this pod")

			return m, nil
		}

		if len(msg.Containers) == 1 {
			return m, m.execIntoPod(msg.Namespace, msg.PodName, msg.Containers[0])
		}

		m.execContainers = msg.Containers
		m.execPodName = msg.PodName
		m.execNamespace = msg.Namespace
		m.selectIdx = 0
		m.switchFilter = ""
		m.switchFiltered = msg.Containers
		m.viewMode = ViewContainerSelect

		return m, nil

	case panels.ScaleStatefulSetRequestMsg:
		description := fmt.Sprintf(
			"Enter new replica count for %s (current: %d)",
			msg.StatefulSetName,
			msg.CurrentReplicas,
		)
		currentReplicas := msg.CurrentReplicas

		m.showInput(
			"Scale StatefulSet",
			description,
			strconv.Itoa(int(msg.CurrentReplicas)),
			func(value string) tea.Cmd {
				return m.scaleStatefulSet(
					msg.Namespace, msg.StatefulSetName,
					value, currentReplicas,
				)
			},
		)
		m.input.SetValue(strconv.Itoa(int(msg.CurrentReplicas)))

		return m, nil

	case panels.EditHPAMinReplicasRequestMsg:
		description := fmt.Sprintf(
			"Enter new minimum replicas for %s (current: %d)",
			msg.HPAName,
			msg.MinReplicas,
		)
		currentMin := msg.MinReplicas

		m.showInput(
			"Edit HPA Min Replicas",
			description,
			strconv.Itoa(int(msg.MinReplicas)),
			func(value string) tea.Cmd {
				return m.updateHPAMinReplicas(
					msg.Namespace, msg.HPAName, value, currentMin,
				)
			},
		)
		m.input.SetValue(strconv.Itoa(int(msg.MinReplicas)))

		return m, nil

	case panels.EditHPAMaxReplicasRequestMsg:
		description := fmt.Sprintf(
			"Enter new maximum replicas for %s (current: %d)",
			msg.HPAName,
			msg.MaxReplicas,
		)
		currentMax := msg.MaxReplicas

		m.showInput(
			"Edit HPA Max Replicas",
			description,
			strconv.Itoa(int(msg.MaxReplicas)),
			func(value string) tea.Cmd {
				return m.updateHPAMaxReplicas(
					msg.Namespace, msg.HPAName, value, currentMax,
				)
			},
		)
		m.input.SetValue(strconv.Itoa(int(msg.MaxReplicas)))

		return m, nil

	case panels.RestartDeploymentRequestMsg:
		return m, m.restartDeployment(msg.Namespace, msg.DeploymentName)

	case panels.RestartStatefulSetRequestMsg:
		return m, m.restartStatefulSet(msg.Namespace, msg.StatefulSetName)

	case panels.RestartDaemonSetRequestMsg:
		return m, m.restartDaemonSet(msg.Namespace, msg.DaemonSetName)

	case panels.ToggleSuspendCronJobRequestMsg:
		return m, m.toggleSuspendCronJob(
			msg.Namespace, msg.CronJobName, msg.CurrentSuspend,
		)

	case panels.TriggerCronJobRequestMsg:
		return m, m.triggerCronJob(msg.Namespace, msg.CronJobName)

	case components.UndoRequestMsg:
		return m, m.handleUndo(msg.RecordID)
	}

	// Update ALL panels so they can process their respective loaded messages
	// (e.g., podsLoadedMsg, deploymentsLoadedMsg) after namespace/context switch
	for i := range m.panels {
		panel, cmd := m.panels[i].Update(msg)
		m.panels[i] = panel

		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content string

	switch m.viewMode {
	case ViewHelp:
		content = m.help.View(m.width, m.height)
	case ViewYaml:
		content = m.yamlView.View(m.width, m.height)
	case ViewLogs:
		content = m.logView.View(m.width, m.height)
	case ViewDiff:
		content = m.diffView.View(m.width, m.height)
	case ViewConfirm:
		content = m.renderNormalView()
		confirmView := m.confirm.View()
		content = m.overlayView(content, confirmView)
	case ViewContextSwitch, ViewNamespaceSwitch, ViewContainerSelect:
		content = m.renderSwitchView()
	case ViewGlobalSearch:
		content = m.renderGlobalSearchView()
	case ViewHistory:
		content = m.historyView.View(m.width, m.height)
	case ViewInput:
		content = m.renderNormalView()
		inputView := m.input.View()
		content = m.overlayView(content, inputView)
	case ViewNormal:
		content = m.renderNormalView()
	}

	return content
}

func (m *Model) renderNormalView() string {
	headerHeight := 1
	statusBarHeight := 1

	m.header.SetContext(m.k8sClient.CurrentContext())
	m.header.SetNamespace(m.k8sClient.CurrentNamespace())
	header := m.header.View(m.width)

	statusBar := m.statusBar.View(m.width)

	// Reserve space for header, status bar, and panel borders
	panelHeight := max(m.height-headerHeight-statusBarHeight, 3)

	// Search bar if active (takes 1 line + newline = 2 lines total)
	var searchView string

	if m.searchActive {
		searchView = m.search.View(m.width)
		panelHeight -= 2
	}

	panelsView := m.renderPanels(m.width, panelHeight)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	if m.searchActive {
		b.WriteString(searchView)
		b.WriteString("\n")
	}

	b.WriteString(panelsView)
	b.WriteString("\n")
	b.WriteString(statusBar)

	return b.String()
}

func (m *Model) renderPanels(width, height int) string {
	if len(m.panels) == 0 {
		return "No panels configured"
	}

	if m.zoomed {
		panelHeight := max(height-borderLines, 1)
		activePanel := m.panels[m.activePanelIdx]
		activePanel.SetSize(width, panelHeight)
		activePanel.SetFocused(true)

		return activePanel.View()
	}

	leftPanelWidth := width / 4
	rightPanelWidth := width - leftPanelWidth - 1

	numPanels := len(m.panels)
	borderOverhead := numPanels * borderLines

	availableHeight := max(height-borderOverhead, numPanels)

	panelHeight := availableHeight / numPanels

	var leftPanels []string

	for i, panel := range m.panels {
		panel.SetSize(leftPanelWidth, panelHeight)
		panel.SetFocused(i == m.activePanelIdx)
		leftPanels = append(leftPanels, panel.View())
	}

	leftView := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)

	detailHeight := max(height-borderLines, 1)

	rightView := m.renderDetailView(rightPanelWidth, detailHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}

func (m *Model) renderDetailView(width, height int) string {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return ""
	}

	activePanel := m.panels[m.activePanelIdx]
	detail := activePanel.DetailView(width, height)

	style := m.styles.Panel.Width(width).Height(height)

	return style.Render(detail)
}

// maxSwitchItems is the maximum number of items shown in the
// context/namespace/container selector to keep the modal compact.
const maxSwitchItems = 20

func (m *Model) renderSwitchView() string {
	var title string

	switch m.viewMode {
	case ViewContextSwitch:
		title = "Switch Context"
	case ViewNamespaceSwitch:
		title = "Switch Namespace"
	case ViewContainerSelect:
		title = "Select Container"
	case ViewNormal, ViewHelp, ViewYaml, ViewLogs, ViewDiff, ViewConfirm, ViewInput,
		ViewGlobalSearch, ViewHistory:
		// These view modes don't use renderSwitchView
	}

	items := m.switchFiltered
	selectedIdx := m.selectIdx

	var b strings.Builder
	b.WriteString(m.styles.ModalTitle.Render(title))
	b.WriteString("\n\n")

	// Show filter input - vim-style: press / to activate filter mode
	if m.switchFilterActive {
		filterDisplay := m.switchFilter + "_"
		b.WriteString(m.styles.StatusKey.Render("  / "))
		b.WriteString(m.styles.StatusValue.Render(filterDisplay))
	} else if m.switchFilter != "" {
		b.WriteString(m.styles.Muted.Render("  / " + m.switchFilter))
	} else {
		b.WriteString(m.styles.Muted.Render("  / to filter, j/k to navigate"))
	}

	b.WriteString("\n\n")

	// Cap visible items to keep modal compact
	maxVisible := maxSwitchItems
	screenMax := m.height - 12

	if screenMax < maxVisible {
		maxVisible = screenMax
	}

	if maxVisible < 3 {
		maxVisible = 3
	}

	// Calculate scroll window to keep selected item visible
	startIdx := 0
	endIdx := len(items)

	if len(items) > maxVisible {
		halfVisible := maxVisible / 2
		startIdx = selectedIdx - halfVisible

		if startIdx < 0 {
			startIdx = 0
		}

		endIdx = startIdx + maxVisible
		if endIdx > len(items) {
			endIdx = len(items)
			startIdx = endIdx - maxVisible

			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	if startIdx > 0 {
		b.WriteString(m.styles.Muted.Render(fmt.Sprintf("  ↑ %d more", startIdx)))
		b.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		item := items[i]
		if i == selectedIdx {
			b.WriteString(m.styles.ListItemFocused.Render("> " + item))
		} else {
			b.WriteString(m.styles.ListItem.Render("  " + item))
		}

		b.WriteString("\n")
	}

	if endIdx < len(items) {
		b.WriteString(m.styles.Muted.Render(fmt.Sprintf("  ↓ %d more", len(items)-endIdx)))
		b.WriteString("\n")
	}

	if len(items) == 0 && m.switchFilter != "" {
		b.WriteString(m.styles.Muted.Render("  No matches"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render("↑/↓ navigate • enter select • esc cancel"))

	// Use wider modal to fit long names
	modalWidth := m.width / 2
	if modalWidth < 50 {
		modalWidth = 50
	}

	content := m.styles.Modal.Width(modalWidth).Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) overlayView(_, overlay string) string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(m.styles.Background))
}

func (m *Model) updatePanelSizes() {
	if len(m.panels) == 0 {
		return
	}

	if m.zoomed {
		panelHeight := max(m.height-borderLines-2, 1)
		m.panels[m.activePanelIdx].SetSize(m.width, panelHeight)

		return
	}

	leftPanelWidth := m.width / 4
	panelHeight := (m.height - 4) / len(m.panels)

	for _, panel := range m.panels {
		panel.SetSize(leftPanelWidth, panelHeight)
	}
}

func (m *Model) nextPanel() {
	if len(m.panels) == 0 {
		return
	}

	m.panels[m.activePanelIdx].SetFocused(false)
	m.activePanelIdx = (m.activePanelIdx + 1) % len(m.panels)
	m.panels[m.activePanelIdx].SetFocused(true)
	m.exitZoom()
}

func (m *Model) prevPanel() {
	if len(m.panels) == 0 {
		return
	}

	m.panels[m.activePanelIdx].SetFocused(false)
	m.activePanelIdx = (m.activePanelIdx - 1 + len(m.panels)) % len(m.panels)
	m.panels[m.activePanelIdx].SetFocused(true)
	m.exitZoom()
}

func (m *Model) selectPanel(idx int) {
	if idx < 0 || idx >= len(m.panels) {
		return
	}

	m.panels[m.activePanelIdx].SetFocused(false)
	m.activePanelIdx = idx
	m.panels[m.activePanelIdx].SetFocused(true)
	m.exitZoom()
}

func (m *Model) exitZoom() {
	if m.zoomed {
		m.zoomed = false
		m.updatePanelSizes()
	}
}

func (m *Model) startContextSwitch() (*Model, tea.Cmd) {
	m.contextList = m.k8sClient.GetContexts()
	m.switchFilter = ""
	m.switchFilterActive = false
	m.switchFiltered = m.contextList

	m.selectIdx = 0
	for i, ctx := range m.switchFiltered {
		if ctx == m.k8sClient.CurrentContext() {
			m.selectIdx = i

			break
		}
	}

	m.viewMode = ViewContextSwitch

	return m, nil
}

func (m *Model) startNamespaceSwitch() (*Model, tea.Cmd) {
	nsList, err := m.k8sClient.ListNamespaces(m.k8sClient.Context())
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to list namespaces: %v", err))

		return m, nil
	}

	m.namespaceList = make([]string, 0, len(nsList))
	for _, ns := range nsList {
		m.namespaceList = append(m.namespaceList, ns.Name)
	}

	m.switchFilter = ""
	m.switchFilterActive = false
	m.switchFiltered = m.namespaceList

	m.selectIdx = 0
	for i, ns := range m.switchFiltered {
		if ns == m.k8sClient.CurrentNamespace() {
			m.selectIdx = i

			break
		}
	}

	m.viewMode = ViewNamespaceSwitch

	return m, nil
}

func (m *Model) handleContextSwitch(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if m.switchFilterActive {
		switch msg.String() {
		case "esc":
			m.switchFilterActive = false
		case "enter":
			m.switchFilterActive = false
		case "backspace":
			if len(m.switchFilter) > 0 {
				m.switchFilter = m.switchFilter[:len(m.switchFilter)-1]
				m.applySwitchFilter(m.contextList)
			}
		default:
			if r := msg.Runes; len(r) == 1 && isValidFilterChar(r[0]) {
				m.switchFilter += string(r)
				m.applySwitchFilter(m.contextList)
			}
		}

		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		}
	case "down", "j":
		if m.selectIdx < len(m.switchFiltered)-1 {
			m.selectIdx++
		}
	case "/":
		m.switchFilterActive = true
	case "enter":
		if m.selectIdx < len(m.switchFiltered) {
			ctx := m.switchFiltered[m.selectIdx]

			if err := m.k8sClient.SwitchContext(ctx); err != nil {
				m.statusBar.SetError(fmt.Sprintf("Failed to switch context: %v", err))
			} else {
				m.header.SetContext(ctx)
				m.header.SetNamespace(m.k8sClient.CurrentNamespace())
				m.statusBar.SetMessage(fmt.Sprintf("Switched to context: %s", ctx))

				var cmds []tea.Cmd
				for _, panel := range m.panels {
					cmds = append(cmds, panel.Refresh())
				}

				m.viewMode = ViewNormal

				return m, tea.Batch(cmds...)
			}
		}

		m.viewMode = ViewNormal
	case "esc":
		if m.switchFilter != "" {
			m.switchFilter = ""
			m.applySwitchFilter(m.contextList)
		} else {
			m.viewMode = ViewNormal
		}
	}

	return m, nil
}

func (m *Model) handleNamespaceSwitch(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if m.switchFilterActive {
		switch msg.String() {
		case "esc":
			m.switchFilterActive = false
			m.switchFilter = ""
			m.applySwitchFilter(m.namespaceList)
		case "enter":
			m.switchFilterActive = false
		case "backspace":
			if len(m.switchFilter) > 0 {
				m.switchFilter = m.switchFilter[:len(m.switchFilter)-1]
				m.applySwitchFilter(m.namespaceList)
			}
		default:
			if r := msg.Runes; len(r) == 1 && isValidFilterChar(r[0]) {
				m.switchFilter += string(r)
				m.applySwitchFilter(m.namespaceList)
			}
		}

		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		}
	case "down", "j":
		if m.selectIdx < len(m.switchFiltered)-1 {
			m.selectIdx++
		}
	case "/":
		m.switchFilterActive = true
	case "enter":
		if m.selectIdx < len(m.switchFiltered) {
			ns := m.switchFiltered[m.selectIdx]
			m.k8sClient.SetNamespace(ns)
			m.header.SetNamespace(ns)
			m.statusBar.SetMessage(fmt.Sprintf("Switched to namespace: %s", ns))

			var cmds []tea.Cmd
			for _, panel := range m.panels {
				cmds = append(cmds, panel.Refresh())
			}

			m.viewMode = ViewNormal

			return m, tea.Batch(cmds...)
		}

		m.viewMode = ViewNormal
	case "esc":
		if m.switchFilter != "" {
			m.switchFilter = ""
			m.applySwitchFilter(m.namespaceList)
		} else {
			m.viewMode = ViewNormal
		}
	}

	return m, nil
}

// applySwitchFilter filters the source list by the current switchFilter
// and resets the selection cursor.
func (m *Model) applySwitchFilter(source []string) {
	if m.switchFilter == "" {
		m.switchFiltered = source
		m.selectIdx = 0

		return
	}

	query := strings.ToLower(m.switchFilter)
	m.switchFiltered = make([]string, 0)

	for _, item := range source {
		if strings.Contains(strings.ToLower(item), query) {
			m.switchFiltered = append(m.switchFiltered, item)
		}
	}

	m.selectIdx = 0
}

// isValidFilterChar returns true if the character is valid for filtering
// Kubernetes resource names (alphanumeric, hyphen, underscore, dot).
func isValidFilterChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.'
}

func (m *Model) showYaml() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]

	yaml, err := activePanel.GetSelectedYAML()
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to get YAML: %v", err))

		return m, nil
	}

	m.yamlView.SetContent(yaml)
	m.viewMode = ViewYaml

	return m, nil
}

func (m *Model) showLogs() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]

	// Single-pod logs keep the original code path; workload panels fan out
	// across every pod matched by the workload's pod selector.
	if podPanel, ok := activePanel.(*panels.PodsPanel); ok {
		pod := podPanel.SelectedPod()
		if pod == nil {
			return m, nil
		}

		m.viewMode = ViewLogs
		cmd := m.logView.Start(m.k8sClient, pod.Namespace, pod.Name, "")

		return m, cmd
	}

	kind, name, namespace, selector, ok := workloadLogTarget(activePanel.SelectedItem())
	if !ok {
		m.statusBar.SetMessage("Logs not available for this resource")

		return m, nil
	}

	if selector == "" {
		m.statusBar.SetError(fmt.Sprintf("%s %s has no pod selector", kind, name))

		return m, nil
	}

	m.viewMode = ViewLogs
	cmd := m.logView.StartMulti(m.k8sClient, namespace, kind, name, selector)

	return m, cmd
}

// workloadLogTarget extracts the workload kind/name/namespace/selector for a
// selected panel item. Returns ok=false when item is nil or isn't a workload
// type we know how to fan out logs over.
func workloadLogTarget(item any) (kind, name, namespace, selector string, ok bool) {
	if item == nil {
		return "", "", "", "", false
	}

	switch v := item.(type) {
	case *appsv1.Deployment:
		return "Deployment", v.Name, v.Namespace, k8s.GetDeploymentPodSelector(v), true
	case *appsv1.StatefulSet:
		return "StatefulSet", v.Name, v.Namespace, k8s.GetStatefulSetPodSelector(v), true
	case *appsv1.DaemonSet:
		return "DaemonSet", v.Name, v.Namespace, k8s.GetDaemonSetPodSelector(v), true
	case *batchv1.Job:
		return "Job", v.Name, v.Namespace, k8s.GetJobPodSelector(v), true
	}

	return "", "", "", "", false
}

func (m *Model) confirmDelete() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]

	name := activePanel.SelectedName()
	if name == "" {
		return m, nil
	}

	panelTitle := activePanel.Title()
	ns := m.k8sClient.CurrentNamespace()

	m.confirm.Show(
		fmt.Sprintf("Delete %s?", name),
		fmt.Sprintf(
			"Are you sure you want to delete %s? "+
				"This action cannot be undone.", name,
		),
		func() tea.Cmd {
			cmd := activePanel.Delete()

			m.historyStore.Add(components.OperationRecord{
				Type:      components.OpDeleteResource,
				Resource:  name,
				Namespace: ns,
				Message: fmt.Sprintf(
					"Deleted %s %s", panelTitle, name,
				),
			})

			return cmd
		},
	)
	m.viewMode = ViewConfirm

	return m, nil
}

func (m *Model) showDescribe() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]

	describe, err := activePanel.GetSelectedDescribe()
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to describe: %v", err))

		return m, nil
	}

	m.yamlView.SetContent(describe)
	m.viewMode = ViewYaml

	return m, nil
}

func (m *Model) showInput(title, description, placeholder string, action func(string) tea.Cmd) {
	m.input.Show(title, description, placeholder)
	m.pendingInputAction = action
	m.viewMode = ViewInput
}

// parseReplicaCount parses a replica string and validates it against
// a minimum value. Returns the parsed count or an error message.
func parseReplicaCount(
	replicaStr string,
	minValue int64,
	minErr error,
) (int32, tea.Msg) {
	replicas, err := strconv.ParseInt(replicaStr, 10, 32)
	if err != nil {
		return 0, panels.ErrorMsg{
			Error: fmt.Errorf("invalid replica count: %w", err),
		}
	}

	if replicas < minValue {
		return 0, panels.ErrorMsg{Error: minErr}
	}

	return int32(replicas), nil
}

func (m *Model) scaleDeployment(
	namespace, name, replicaStr string,
	currentReplicas int32,
) tea.Cmd {
	return func() tea.Msg {
		replicas, errMsg := parseReplicaCount(
			replicaStr, 0, ErrInvalidReplicaCount,
		)
		if errMsg != nil {
			return errMsg
		}

		ctx := context.Background()

		err := m.k8sClient.ScaleDeployment(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to scale deployment: %w", err),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpScaleDeployment,
			Resource:  name,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Scaled %s from %d to %d replicas",
				name, currentReplicas, replicas,
			),
			Undoable: true,
			UndoData: components.UndoData{
				PreviousReplicas: currentReplicas,
			},
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Scaled %s to %d replicas", name, replicas,
			),
		}
	}
}

func (m *Model) rollbackDeployment(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.k8sClient.RollbackDeployment(ctx, namespace, name); err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to rollback deployment: %w", err),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpRollbackDeployment,
			Resource:  name,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Rolled back %s to previous revision", name,
			),
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Rolled back %s to previous revision", name),
		}
	}
}

func (m *Model) loadRevisionDiff(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		revisions, err := m.k8sClient.ListDeploymentRevisions(ctx, namespace, name)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to list revisions: %w", err),
			}
		}

		if len(revisions) < 2 {
			return panels.StatusMsg{
				Message: fmt.Sprintf(
					"%s has fewer than 2 revisions — nothing to diff",
					name,
				),
			}
		}

		// Compare the two most recent revisions (current vs previous)
		currentYAML, err := m.k8sClient.GetRevisionYAML(
			ctx, namespace, name, revisions[0].Revision,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to get current revision YAML: %w", err),
			}
		}

		previousYAML, err := m.k8sClient.GetRevisionYAML(
			ctx, namespace, name, revisions[1].Revision,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to get previous revision YAML: %w", err),
			}
		}

		title := fmt.Sprintf(
			"Diff: %s (rev %d → %d)",
			name,
			revisions[1].Revision,
			revisions[0].Revision,
		)

		return diffLoadedMsg{
			title:   title,
			oldYAML: previousYAML,
			newYAML: currentYAML,
		}
	}
}

func (m *Model) scaleStatefulSet(
	namespace, name, replicaStr string,
	currentReplicas int32,
) tea.Cmd {
	return func() tea.Msg {
		replicas, errMsg := parseReplicaCount(
			replicaStr, 0, ErrInvalidReplicaCount,
		)
		if errMsg != nil {
			return errMsg
		}

		ctx := context.Background()

		err := m.k8sClient.ScaleStatefulSet(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to scale statefulset: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpScaleStatefulSet,
			Resource:  name,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Scaled %s from %d to %d replicas",
				name, currentReplicas, replicas,
			),
			Undoable: true,
			UndoData: components.UndoData{
				PreviousReplicas: currentReplicas,
			},
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Scaled %s to %d replicas", name, replicas,
			),
		}
	}
}

func (m *Model) updateHPAMinReplicas(
	namespace, name, replicaStr string,
	currentMin int32,
) tea.Cmd {
	return func() tea.Msg {
		replicas, errMsg := parseReplicaCount(
			replicaStr, 1, ErrMinReplicasTooLow,
		)
		if errMsg != nil {
			return errMsg
		}

		ctx := context.Background()

		err := m.k8sClient.UpdateHPAMinReplicas(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to update HPA min replicas: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpEditHPAMin,
			Resource:  name,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Updated %s min replicas from %d to %d",
				name, currentMin, replicas,
			),
			Undoable: true,
			UndoData: components.UndoData{
				PreviousMinReplicas: currentMin,
			},
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Updated %s min replicas to %d", name, replicas,
			),
		}
	}
}

func (m *Model) updateHPAMaxReplicas(
	namespace, name, replicaStr string,
	currentMax int32,
) tea.Cmd {
	return func() tea.Msg {
		replicas, errMsg := parseReplicaCount(
			replicaStr, 1, ErrMaxReplicasTooLow,
		)
		if errMsg != nil {
			return errMsg
		}

		ctx := context.Background()

		err := m.k8sClient.UpdateHPAMaxReplicas(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to update HPA max replicas: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpEditHPAMax,
			Resource:  name,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Updated %s max replicas from %d to %d",
				name, currentMax, replicas,
			),
			Undoable: true,
			UndoData: components.UndoData{
				PreviousMaxReplicas: currentMax,
			},
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Updated %s max replicas to %d", name, replicas,
			),
		}
	}
}

func (m *Model) startPortForward(namespace, podName, portSpec string) tea.Cmd {
	return func() tea.Msg {
		var localPort, remotePort int

		_, err := fmt.Sscanf(portSpec, "%d:%d", &localPort, &remotePort)
		if err != nil {
			_, err = fmt.Sscanf(portSpec, "%d", &remotePort)
			if err != nil {
				return panels.ErrorMsg{Error: ErrInvalidPortFormat}
			}

			localPort = remotePort
		}

		stopCh := make(chan struct{})
		readyCh := make(chan struct{})

		opts := k8s.PortForwardOptions{
			Namespace:  namespace,
			PodName:    podName,
			LocalPort:  localPort,
			RemotePort: remotePort,
			StopCh:     stopCh,
			ReadyCh:    readyCh,
		}

		pf, err := m.k8sClient.NewPortForwarder(opts)
		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("failed to create port forwarder: %w", err)}
		}

		go func() {
			_ = pf.Start()
		}()

		<-readyCh

		key := fmt.Sprintf("%s/%s:%d", namespace, podName, remotePort)
		m.portForwards[key] = pf

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpPortForward,
			Resource:  podName,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Port forwarding %d -> %s:%d",
				localPort, podName, remotePort,
			),
		})

		return panels.StatusMsg{
			Message: fmt.Sprintf(
				"Port forwarding %d -> %s:%d",
				localPort, podName, remotePort,
			),
		}
	}
}

func (m *Model) stopAllPortForwards() {
	for key, pf := range m.portForwards {
		pf.Stop()
		delete(m.portForwards, key)
	}
}

func (m *Model) refreshAllPanels() tea.Cmd {
	var cmds []tea.Cmd

	for _, panel := range m.panels {
		cmds = append(cmds, panel.Refresh())
	}

	return tea.Batch(cmds...)
}

func (m *Model) handleContainerSelect(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if m.switchFilterActive {
		switch msg.String() {
		case "esc":
			m.switchFilterActive = false
			m.switchFilter = ""
			m.applySwitchFilter(m.execContainers)
		case "enter":
			m.switchFilterActive = false
		case "backspace":
			if len(m.switchFilter) > 0 {
				m.switchFilter = m.switchFilter[:len(m.switchFilter)-1]
				m.applySwitchFilter(m.execContainers)
			}
		default:
			if r := msg.Runes; len(r) == 1 && isValidFilterChar(r[0]) {
				m.switchFilter += string(r)
				m.applySwitchFilter(m.execContainers)
			}
		}

		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		}
	case "down", "j":
		if m.selectIdx < len(m.switchFiltered)-1 {
			m.selectIdx++
		}
	case "/":
		m.switchFilterActive = true
	case "enter":
		if m.selectIdx < len(m.switchFiltered) {
			container := m.switchFiltered[m.selectIdx]
			m.viewMode = ViewNormal

			return m, m.execIntoPod(m.execNamespace, m.execPodName, container)
		}

		m.viewMode = ViewNormal
	case "esc":
		if m.switchFilter != "" {
			m.switchFilter = ""
			m.applySwitchFilter(m.execContainers)
		} else {
			m.viewMode = ViewNormal
		}
	}

	return m, nil
}

func (m *Model) execIntoPod(namespace, podName, container string) tea.Cmd {
	m.historyStore.Add(components.OperationRecord{
		Type:      components.OpExec,
		Resource:  podName,
		Namespace: namespace,
		Message: fmt.Sprintf(
			"Exec into %s (container: %s)", podName, container,
		),
	})

	args := []string{
		"exec", "-it",
		"-n", namespace,
		podName,
		"-c", container,
		"--", "/bin/sh", "-c",
		"if command -v bash > /dev/null; then exec bash; else exec sh; fi",
	}

	return tea.ExecProcess(
		newKubectlCmd(args...),
		func(err error) tea.Msg {
			if err != nil {
				return panels.ErrorMsg{
					Error: fmt.Errorf("exec failed: %w", err),
				}
			}

			return panels.StatusMsg{
				Message: fmt.Sprintf(
					"Exited shell in %s/%s", podName, container,
				),
			}
		},
	)
}

func (m *Model) copyNameToClipboard() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]
	name := activePanel.SelectedName()

	if name == "" {
		m.statusBar.SetMessage("No resource selected")

		return m, nil
	}

	if err := clipboard.WriteAll(name); err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to copy: %v", err))

		return m, nil
	}

	m.statusBar.SetMessage(fmt.Sprintf("Copied '%s' to clipboard", name))

	return m, nil
}

func (m *Model) copyYamlToClipboard() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]

	yamlContent, err := activePanel.GetSelectedYAML()
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to get YAML: %v", err))

		return m, nil
	}

	if err := clipboard.WriteAll(yamlContent); err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to copy: %v", err))

		return m, nil
	}

	m.statusBar.SetMessage("Copied YAML to clipboard")

	return m, nil
}

func (m *Model) editResource() (*Model, tea.Cmd) {
	if len(m.panels) == 0 || m.activePanelIdx >= len(m.panels) {
		return m, nil
	}

	activePanel := m.panels[m.activePanelIdx]
	name := activePanel.SelectedName()

	if name == "" {
		m.statusBar.SetMessage("No resource selected")

		return m, nil
	}

	yamlContent, err := activePanel.GetSelectedYAML()
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to get YAML: %v", err))

		return m, nil
	}

	tmpDir := os.TempDir()

	f, err := os.CreateTemp(tmpDir, fmt.Sprintf("lazy-k8s-%s-*.yaml", name))
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to create temp file: %v", err))

		return m, nil
	}

	tmpFile := f.Name()

	if _, err := f.WriteString(yamlContent); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpFile)

		m.statusBar.SetError(fmt.Sprintf("Failed to write temp file: %v", err))

		return m, nil
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpFile)

		m.statusBar.SetError(fmt.Sprintf("Failed to close temp file: %v", err))

		return m, nil
	}

	// Get editor from environment, fall back to vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}

	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, tmpFile) //nolint:gosec,noctx
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer func() { _ = os.Remove(tmpFile) }()

		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("editor failed: %w", err)}
		}

		applyCmd := exec.Command("kubectl", "apply", "-f", tmpFile) //nolint:gosec,noctx

		output, err := applyCmd.CombinedOutput()
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(string(output))),
			}
		}

		ns := m.k8sClient.CurrentNamespace()

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpEditResource,
			Resource:  name,
			Namespace: ns,
			Message: fmt.Sprintf(
				"Edited and applied changes to %s", name,
			),
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Applied changes to %s", name),
		}
	})
}

func newKubectlCmd(args ...string) *exec.Cmd {
	return exec.Command("kubectl", args...) //nolint:noctx
}

func (m *Model) restartDeployment(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.RestartDeployment(ctx, namespace, name)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to restart deployment: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpRestartDeployment,
			Resource:  name,
			Namespace: namespace,
			Message:   fmt.Sprintf("Restarted deployment %s", name),
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Restarted deployment: %s", name),
		}
	}
}

func (m *Model) restartStatefulSet(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.RestartStatefulSet(ctx, namespace, name)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to restart statefulset: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpRestartStatefulSet,
			Resource:  name,
			Namespace: namespace,
			Message:   fmt.Sprintf("Restarted statefulset %s", name),
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Restarted statefulset: %s", name),
		}
	}
}

func (m *Model) restartDaemonSet(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.RestartDaemonSet(ctx, namespace, name)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to restart daemonset: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpRestartDaemonSet,
			Resource:  name,
			Namespace: namespace,
			Message:   fmt.Sprintf("Restarted daemonset %s", name),
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Restarted daemonset: %s", name),
		}
	}
}

func (m *Model) toggleSuspendCronJob(
	namespace, name string,
	currentSuspend bool,
) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		newSuspend := !currentSuspend

		err := m.k8sClient.SuspendCronJob(
			ctx, namespace, name, newSuspend,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to toggle suspend cronjob: %w", err,
				),
			}
		}

		opType := components.OpSuspendCronJob
		action := "Suspended"

		if !newSuspend {
			opType = components.OpResumeCronJob
			action = "Resumed"
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      opType,
			Resource:  name,
			Namespace: namespace,
			Message:   fmt.Sprintf("%s cronjob %s", action, name),
			Undoable:  true,
			UndoData: components.UndoData{
				PreviousSuspend: currentSuspend,
			},
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("%s cronjob: %s", action, name),
		}
	}
}

func (m *Model) triggerCronJob(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		job, err := m.k8sClient.TriggerCronJob(ctx, namespace, name)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"failed to trigger cronjob: %w", err,
				),
			}
		}

		m.historyStore.Add(components.OperationRecord{
			Type:      components.OpTriggerCronJob,
			Resource:  name,
			Namespace: namespace,
			Message: fmt.Sprintf(
				"Triggered cronjob %s (job: %s)", name, job.Name,
			),
		})

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Triggered job: %s", job.Name),
		}
	}
}

// handleUndo reverses a previously recorded operation when possible.
func (m *Model) handleUndo(recordID int) tea.Cmd {
	rec, ok := m.historyStore.Get(recordID)
	if !ok || !rec.Undoable || rec.Undone {
		return nil
	}

	m.historyStore.MarkUndone(recordID)

	switch rec.Type {
	case components.OpScaleDeployment:
		return m.undoScaleDeployment(
			rec.Namespace, rec.Resource,
			rec.UndoData.PreviousReplicas,
		)
	case components.OpScaleStatefulSet:
		return m.undoScaleStatefulSet(
			rec.Namespace, rec.Resource,
			rec.UndoData.PreviousReplicas,
		)
	case components.OpSuspendCronJob, components.OpResumeCronJob:
		return m.toggleSuspendCronJob(
			rec.Namespace, rec.Resource,
			!rec.UndoData.PreviousSuspend,
		)
	case components.OpEditHPAMin:
		return m.undoHPAMinReplicas(
			rec.Namespace, rec.Resource,
			rec.UndoData.PreviousMinReplicas,
		)
	case components.OpEditHPAMax:
		return m.undoHPAMaxReplicas(
			rec.Namespace, rec.Resource,
			rec.UndoData.PreviousMaxReplicas,
		)
	case components.OpRestartDeployment,
		components.OpRestartStatefulSet,
		components.OpRestartDaemonSet,
		components.OpRollbackDeployment,
		components.OpDeleteResource,
		components.OpPortForward,
		components.OpExec,
		components.OpEditResource,
		components.OpTriggerCronJob:
		// These operations are not reversible
		return nil
	}

	return nil
}

func (m *Model) undoScaleDeployment(
	namespace, name string,
	replicas int32,
) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.ScaleDeployment(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("undo scale failed: %w", err),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Undone: restored %s to %d replicas",
				name, replicas,
			),
		}
	}
}

func (m *Model) undoScaleStatefulSet(
	namespace, name string,
	replicas int32,
) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.ScaleStatefulSet(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("undo scale failed: %w", err),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Undone: restored %s to %d replicas",
				name, replicas,
			),
		}
	}
}

func (m *Model) undoHPAMinReplicas(
	namespace, name string,
	replicas int32,
) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.UpdateHPAMinReplicas(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"undo HPA min replicas failed: %w", err,
				),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Undone: restored %s min replicas to %d",
				name, replicas,
			),
		}
	}
}

func (m *Model) undoHPAMaxReplicas(
	namespace, name string,
	replicas int32,
) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.k8sClient.UpdateHPAMaxReplicas(
			ctx, namespace, name, replicas,
		)
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf(
					"undo HPA max replicas failed: %w", err,
				),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf(
				"Undone: restored %s max replicas to %d",
				name, replicas,
			),
		}
	}
}

func (m *Model) metricsTickCmd() tea.Cmd {
	return tea.Tick(metricsRefreshInterval, func(t time.Time) tea.Msg {
		return metricsTickMsg(t)
	})
}

func (m *Model) fetchMetrics() tea.Cmd {
	if m.metricsClient == nil {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()

		var podMetrics map[string]panels.PodMetrics

		if m.showAllNs {
			k8sMetrics, _ := m.metricsClient.GetAllPodMetrics(ctx)
			podMetrics = convertPodMetrics(k8sMetrics)
		} else {
			k8sMetrics, _ := m.metricsClient.GetPodMetrics(ctx, m.k8sClient.CurrentNamespace())
			podMetrics = convertPodMetrics(k8sMetrics)
		}

		k8sNodeMetrics, _ := m.metricsClient.GetNodeMetrics(ctx)
		nodeMetrics := convertNodeMetrics(k8sNodeMetrics)

		return metricsLoadedMsg{
			podMetrics:  podMetrics,
			nodeMetrics: nodeMetrics,
		}
	}
}

type diffLoadedMsg struct {
	title   string
	oldYAML string
	newYAML string
}

type metricsLoadedMsg struct {
	podMetrics  map[string]panels.PodMetrics
	nodeMetrics map[string]panels.NodeMetrics
}

func convertPodMetrics(k8sMetrics map[string]k8s.PodMetrics) map[string]panels.PodMetrics {
	result := make(map[string]panels.PodMetrics, len(k8sMetrics))
	for key, m := range k8sMetrics {
		result[key] = panels.PodMetrics{
			Name:      m.Name,
			Namespace: m.Namespace,
			CPU:       m.CPU,
			Memory:    m.Memory,
		}
	}

	return result
}

func convertNodeMetrics(k8sMetrics map[string]k8s.NodeMetrics) map[string]panels.NodeMetrics {
	result := make(map[string]panels.NodeMetrics, len(k8sMetrics))
	for key, m := range k8sMetrics {
		result[key] = panels.NodeMetrics{
			Name:   m.Name,
			CPU:    m.CPU,
			Memory: m.Memory,
		}
	}

	return result
}

// handleGlobalSearch processes key events while the global search modal is open.
func (m *Model) handleGlobalSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Back) {
		m.viewMode = ViewNormal

		return m, nil
	}

	if key.Matches(msg, m.keys.Enter) {
		m.navigateToSearchResult()

		return m, nil
	}

	var queryChanged bool

	m.globalSearch, _, queryChanged = m.globalSearch.Update(
		msg, len(m.globalSearchResults),
	)

	if queryChanged {
		m.aggregateSearchResults()
	}

	return m, nil
}

// aggregateSearchResults collects matches from all panels for the current query.
func (m *Model) aggregateSearchResults() {
	query := m.globalSearch.Query()
	m.globalSearchResults = nil

	for i, panel := range m.panels {
		results := panel.SearchItems(query)

		for j := range results {
			results[j].PanelIdx = i
		}

		m.globalSearchResults = append(m.globalSearchResults, results...)
	}
}

// navigateToSearchResult switches to the panel and item matching the selected result.
func (m *Model) navigateToSearchResult() {
	idx := m.globalSearch.Cursor()
	if idx < 0 || idx >= len(m.globalSearchResults) {
		return
	}

	result := m.globalSearchResults[idx]

	m.selectPanel(result.PanelIdx)

	// Clear any active per-panel filter so NavigateTo sees all items
	m.panels[result.PanelIdx].SetFilter("")
	m.searchQuery = ""
	m.searchActive = false

	m.panels[result.PanelIdx].NavigateTo(result.Name, result.Namespace)
	m.viewMode = ViewNormal
}

func (m *Model) renderGlobalSearchView() string {
	// Compute modal dimensions first so all internal widths are relative to it.
	// Modal border/padding consumes ~6 columns.
	const modalPadding = 6

	modalWidth := m.width * 3 / 4
	if modalWidth > m.width-2 {
		modalWidth = m.width - 2
	}

	if modalWidth < 30 {
		modalWidth = 30
	}

	innerWidth := modalWidth - modalPadding
	if innerWidth < 20 {
		innerWidth = 20
	}

	var b strings.Builder

	title := m.styles.ModalTitle.Render("Global Search")

	hint := m.styles.Muted.Render("↑/↓ • enter • esc")

	b.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Center, title, "  ", hint,
	))
	b.WriteString("\n")

	// Search input
	b.WriteString(m.globalSearch.InputView())
	b.WriteString("  ")
	b.WriteString(m.styles.Muted.Render(
		fmt.Sprintf("%d results", len(m.globalSearchResults)),
	))
	b.WriteString("\n")

	sepWidth := innerWidth
	if sepWidth < 1 {
		sepWidth = 1
	}

	b.WriteString(strings.Repeat("─", sepWidth))
	b.WriteString("\n")

	// Compute visible area
	modalHeight := m.height - 2
	headerLines := 3 // title + input + separator
	footerLines := 2 // scrollbar + border
	visibleHeight := modalHeight - headerLines - footerLines

	if visibleHeight < 3 {
		visibleHeight = 3
	}

	m.globalSearch.SetVisibleHeight(visibleHeight)
	cursorIdx := m.globalSearch.Cursor()

	// Build flat display list with group headers interleaved
	type displayLine struct {
		isHeader bool
		text     string
		resultID int // index into globalSearchResults, -1 for headers
	}

	var lines []displayLine

	currentKind := ""

	for i, r := range m.globalSearchResults {
		if r.Kind != currentKind {
			// Count how many results of this kind exist
			kindCount := 0

			for _, rr := range m.globalSearchResults {
				if rr.Kind == r.Kind {
					kindCount++
				}
			}

			currentKind = r.Kind
			lines = append(lines, displayLine{
				isHeader: true,
				text: fmt.Sprintf(
					"%s (%d)", r.Kind, kindCount,
				),
				resultID: -1,
			})
		}

		lines = append(lines, displayLine{
			isHeader: false,
			resultID: i,
		})
	}

	// Map cursor from result index to line index
	cursorLineIdx := 0

	for li, line := range lines {
		if !line.isHeader && line.resultID == cursorIdx {
			cursorLineIdx = li

			break
		}
	}

	// Scroll the lines view
	startLine := m.globalSearch.Offset()
	if cursorLineIdx < startLine {
		startLine = cursorLineIdx
	}

	if cursorLineIdx >= startLine+visibleHeight {
		startLine = cursorLineIdx - visibleHeight + 1
	}

	if startLine < 0 {
		startLine = 0
	}

	endLine := startLine + visibleHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Column widths adapt to available space.
	// Layout: "> name  namespace  status"
	nsW := 12
	statusW := 10

	if innerWidth < 60 {
		nsW = 8
		statusW = 8
	}

	// prefix(2) + name + gap(2) + ns + gap(2) + status
	maxNameW := innerWidth - 2 - 2 - nsW - 2 - statusW
	if maxNameW < 8 {
		maxNameW = 8
	}

	for li := startLine; li < endLine; li++ {
		line := lines[li]

		if line.isHeader {
			b.WriteString(m.styles.DetailTitle.Render(line.text))
			b.WriteString("\n")

			continue
		}

		r := m.globalSearchResults[line.resultID]
		isSelected := line.resultID == cursorIdx

		var prefix string
		if isSelected {
			prefix = "> "
		} else {
			prefix = "  "
		}

		name := r.Name
		if len(name) > maxNameW {
			name = name[:maxNameW-3] + "..."
		}

		ns := r.Namespace
		if len(ns) > nsW {
			ns = ns[:nsW-2] + ".."
		}

		status := r.Status
		if len(status) > statusW {
			status = status[:statusW-2] + ".."
		}

		entry := prefix + fmt.Sprintf(
			"%-*s  %-*s  %s",
			maxNameW, name, nsW, ns, status,
		)

		if isSelected {
			b.WriteString(m.styles.ListItemFocused.Render(entry))
		} else {
			b.WriteString(m.styles.ListItem.Render(entry))
		}

		b.WriteString("\n")
	}

	if len(m.globalSearchResults) == 0 && m.globalSearch.Query() != "" {
		b.WriteString(m.styles.Muted.Render("  No matches"))
		b.WriteString("\n")
	}

	// Scrollbar
	if len(lines) > visibleHeight {
		barWidth := innerWidth - 4
		if barWidth < 3 {
			barWidth = 3
		}

		scrollPos := float64(startLine) / float64(
			len(lines)-visibleHeight,
		)

		leftWidth := int(float64(barWidth) * scrollPos)
		if leftWidth < 0 {
			leftWidth = 0
		}

		if leftWidth > barWidth {
			leftWidth = barWidth
		}

		rightWidth := barWidth - leftWidth - 1
		if rightWidth < 0 {
			rightWidth = 0
		}

		indicator := strings.Repeat("─", leftWidth) +
			"█" + strings.Repeat("─", rightWidth)

		b.WriteString("\n")
		b.WriteString(m.styles.Muted.Render(indicator))
	}

	content := m.styles.Modal.
		Width(modalWidth).
		Render(b.String())

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}
