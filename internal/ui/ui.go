package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/config"
	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/components"
	"github.com/lazyk8s/lazy-k8s/internal/ui/panels"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

var (
	// ErrInvalidReplicaCount is returned when the replica count is invalid.
	ErrInvalidReplicaCount = errors.New("replica count must be non-negative")
	// ErrInvalidPortFormat is returned when the port format is invalid.
	ErrInvalidPortFormat = errors.New("invalid port format, use local:remote or port")
	// ErrMinReplicasTooLow is returned when min replicas is less than 1.
	ErrMinReplicasTooLow = errors.New("min replicas must be at least 1")
	// ErrMaxReplicasTooLow is returned when max replicas is less than 1.
	ErrMaxReplicasTooLow = errors.New("max replicas must be at least 1")
	// ErrApplyFailed is returned when kubectl apply fails.
	ErrApplyFailed = errors.New("kubectl apply failed")
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
)

// borderLines is the number of lines used by panel borders (top + bottom).
const borderLines = 2

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
	search    *components.Search
	input     *components.Input

	// Input action callback - stores the function to call when input is submitted
	pendingInputAction func(value string) tea.Cmd

	// State
	viewMode     ViewMode
	lastError    string
	lastStatus   string
	showAllNs    bool
	searchActive bool
	searchQuery  string

	// Context/namespace switching
	contextList   []string
	namespaceList []string
	selectIdx     int

	// Port forwarding
	portForwards map[string]*k8s.PortForwarder

	// Exec container selection
	execContainers []string
	execPodName    string
	execNamespace  string
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

	// Initialize header and status bar
	m.header = components.NewHeader(styles, client.CurrentContext(), client.CurrentNamespace())
	m.statusBar = components.NewStatusBar(styles)
	m.help = components.NewHelp(styles, keys)
	m.confirm = components.NewConfirm(styles)
	m.yamlView = components.NewYamlViewer(styles)
	m.logView = components.NewLogViewer(styles)
	m.search = components.NewSearch(styles)
	m.input = components.NewInput(styles)

	// Initialize panels based on config
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

	// Set first panel as active
	if len(m.panels) > 0 {
		m.panels[0].SetFocused(true)
	}
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize all panels
	for _, panel := range m.panels {
		cmds = append(cmds, panel.Init())
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

		case ViewConfirm:
			var cmd tea.Cmd

			m.confirm, cmd = m.confirm.Update(msg)
			if m.confirm.Done() {
				m.viewMode = ViewNormal
				if m.confirm.Confirmed() {
					// Execute the confirmed action
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

		case ViewNormal:
			// Fall through to normal key handling below
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
			// Apply filter to current panel
			if len(m.panels) > m.activePanelIdx {
				m.panels[m.activePanelIdx].SetFilter(m.searchQuery)
			}

			return m, cmd
		}

		// Global keys
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.stopAllPortForwards()

			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.viewMode = ViewHelp

			return m, nil

		case key.Matches(msg, m.keys.Search):
			m.searchActive = true
			m.search.Focus()

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
			return m, m.refreshAllPanels()

		case key.Matches(msg, m.keys.AllNamespace):
			m.showAllNs = !m.showAllNs
			m.statusBar.SetMessage(fmt.Sprintf("All namespaces: %v", m.showAllNs))

			for _, panel := range m.panels {
				panel.SetAllNamespaces(m.showAllNs)
				cmds = append(cmds, panel.Refresh())
			}

			return m, tea.Batch(cmds...)

		// Panel number shortcuts
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

		// Resource actions
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
			// Pass to active panel
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

		return m, m.refreshAllPanels()

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
		m.showInput(
			"Scale Deployment",
			description,
			strconv.Itoa(int(msg.CurrentReplicas)),
			func(value string) tea.Cmd {
				return m.scaleDeployment(msg.Namespace, msg.DeploymentName, value)
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

		// Single container - exec directly
		if len(msg.Containers) == 1 {
			return m, m.execIntoPod(msg.Namespace, msg.PodName, msg.Containers[0])
		}

		// Multiple containers - show selection
		m.execContainers = msg.Containers
		m.execPodName = msg.PodName
		m.execNamespace = msg.Namespace
		m.selectIdx = 0
		m.viewMode = ViewContainerSelect

		return m, nil

	case panels.ScaleStatefulSetRequestMsg:
		description := fmt.Sprintf(
			"Enter new replica count for %s (current: %d)",
			msg.StatefulSetName,
			msg.CurrentReplicas,
		)
		m.showInput(
			"Scale StatefulSet",
			description,
			strconv.Itoa(int(msg.CurrentReplicas)),
			func(value string) tea.Cmd {
				return m.scaleStatefulSet(msg.Namespace, msg.StatefulSetName, value)
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
		m.showInput(
			"Edit HPA Min Replicas",
			description,
			strconv.Itoa(int(msg.MinReplicas)),
			func(value string) tea.Cmd {
				return m.updateHPAMinReplicas(msg.Namespace, msg.HPAName, value)
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
		m.showInput(
			"Edit HPA Max Replicas",
			description,
			strconv.Itoa(int(msg.MaxReplicas)),
			func(value string) tea.Cmd {
				return m.updateHPAMaxReplicas(msg.Namespace, msg.HPAName, value)
			},
		)
		m.input.SetValue(strconv.Itoa(int(msg.MaxReplicas)))

		return m, nil
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
	case ViewConfirm:
		content = m.renderNormalView()
		// Overlay confirm dialog
		confirmView := m.confirm.View()
		content = m.overlayView(content, confirmView)
	case ViewContextSwitch, ViewNamespaceSwitch, ViewContainerSelect:
		content = m.renderSwitchView()
	case ViewInput:
		content = m.renderNormalView()
		// Overlay input dialog
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

	// Render panels
	panelsView := m.renderPanels(m.width, panelHeight)

	// Combine views
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

	// Calculate panel widths
	leftPanelWidth := width / 4
	rightPanelWidth := width - leftPanelWidth - 1

	// Left side: stacked resource lists
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

	// Right side: detail view
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

func (m *Model) renderSwitchView() string {
	var (
		title       string
		items       []string
		selectedIdx int
	)

	switch m.viewMode {
	case ViewContextSwitch:
		title = "Switch Context"
		items = m.contextList
		selectedIdx = m.selectIdx
	case ViewNamespaceSwitch:
		title = "Switch Namespace"
		items = m.namespaceList
		selectedIdx = m.selectIdx
	case ViewContainerSelect:
		title = "Select Container"
		items = m.execContainers
		selectedIdx = m.selectIdx
	case ViewNormal, ViewHelp, ViewYaml, ViewLogs, ViewConfirm, ViewInput:
		// These view modes don't use renderSwitchView
	}

	var b strings.Builder
	b.WriteString(m.styles.ModalTitle.Render(title))
	b.WriteString("\n\n")

	for i, item := range items {
		if i == selectedIdx {
			b.WriteString(m.styles.ListItemFocused.Render("> " + item))
		} else {
			b.WriteString(m.styles.ListItem.Render("  " + item))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render("↑/↓ navigate • enter select • esc cancel"))

	content := m.styles.Modal.Width(40).Render(b.String())

	// Center the modal
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) overlayView(_, overlay string) string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(m.styles.Background))
}

func (m *Model) updatePanelSizes() {
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
}

func (m *Model) prevPanel() {
	if len(m.panels) == 0 {
		return
	}

	m.panels[m.activePanelIdx].SetFocused(false)
	m.activePanelIdx = (m.activePanelIdx - 1 + len(m.panels)) % len(m.panels)
	m.panels[m.activePanelIdx].SetFocused(true)
}

func (m *Model) selectPanel(idx int) {
	if idx < 0 || idx >= len(m.panels) {
		return
	}

	m.panels[m.activePanelIdx].SetFocused(false)
	m.activePanelIdx = idx
	m.panels[m.activePanelIdx].SetFocused(true)
}

func (m *Model) startContextSwitch() (*Model, tea.Cmd) {
	m.contextList = m.k8sClient.GetContexts()

	m.selectIdx = 0
	for i, ctx := range m.contextList {
		if ctx == m.k8sClient.CurrentContext() {
			m.selectIdx = i

			break
		}
	}

	m.viewMode = ViewContextSwitch

	return m, nil
}

func (m *Model) startNamespaceSwitch() (*Model, tea.Cmd) {
	// Fetch namespaces
	nsList, err := m.k8sClient.ListNamespaces(m.k8sClient.Context())
	if err != nil {
		m.statusBar.SetError(fmt.Sprintf("Failed to list namespaces: %v", err))

		return m, nil
	}

	m.namespaceList = make([]string, 0, len(nsList))
	for _, ns := range nsList {
		m.namespaceList = append(m.namespaceList, ns.Name)
	}

	m.selectIdx = 0
	for i, ns := range m.namespaceList {
		if ns == m.k8sClient.CurrentNamespace() {
			m.selectIdx = i

			break
		}
	}

	m.viewMode = ViewNamespaceSwitch

	return m, nil
}

func (m *Model) handleContextSwitch(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		}
	case "down", "j":
		if m.selectIdx < len(m.contextList)-1 {
			m.selectIdx++
		}
	case "enter":
		if m.selectIdx < len(m.contextList) {
			ctx := m.contextList[m.selectIdx]
			if err := m.k8sClient.SwitchContext(ctx); err != nil {
				m.statusBar.SetError(fmt.Sprintf("Failed to switch context: %v", err))
			} else {
				m.header.SetContext(ctx)
				m.header.SetNamespace(m.k8sClient.CurrentNamespace())
				m.statusBar.SetMessage(fmt.Sprintf("Switched to context: %s", ctx))
				// Refresh all panels
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
		m.viewMode = ViewNormal
	}

	return m, nil
}

func (m *Model) handleNamespaceSwitch(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		}
	case "down", "j":
		if m.selectIdx < len(m.namespaceList)-1 {
			m.selectIdx++
		}
	case "enter":
		if m.selectIdx < len(m.namespaceList) {
			ns := m.namespaceList[m.selectIdx]
			m.k8sClient.SetNamespace(ns)
			m.header.SetNamespace(ns)
			m.statusBar.SetMessage(fmt.Sprintf("Switched to namespace: %s", ns))
			// Refresh all panels
			var cmds []tea.Cmd
			for _, panel := range m.panels {
				cmds = append(cmds, panel.Refresh())
			}

			m.viewMode = ViewNormal

			return m, tea.Batch(cmds...)
		}

		m.viewMode = ViewNormal
	case "esc":
		m.viewMode = ViewNormal
	}

	return m, nil
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

	podPanel, ok := activePanel.(*panels.PodsPanel)
	if !ok {
		m.statusBar.SetMessage("Logs only available for pods")

		return m, nil
	}

	pod := podPanel.SelectedPod()
	if pod == nil {
		return m, nil
	}

	m.viewMode = ViewLogs
	cmd := m.logView.Start(m.k8sClient, pod.Namespace, pod.Name, "")

	return m, cmd
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

	m.confirm.Show(
		fmt.Sprintf("Delete %s?", name),
		fmt.Sprintf("Are you sure you want to delete %s? This action cannot be undone.", name),
		func() tea.Cmd {
			return activePanel.Delete()
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

func (m *Model) scaleDeployment(namespace, name, replicaStr string) tea.Cmd {
	return func() tea.Msg {
		replicas, err := strconv.ParseInt(replicaStr, 10, 32)
		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("invalid replica count: %w", err)}
		}

		if replicas < 0 {
			return panels.ErrorMsg{Error: ErrInvalidReplicaCount}
		}

		ctx := context.Background()
		if err := m.k8sClient.ScaleDeployment(ctx, namespace, name, int32(replicas)); err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("failed to scale deployment: %w", err)}
		}

		// Return status with refresh to show updated pods
		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Scaled %s to %d replicas", name, replicas),
		}
	}
}

func (m *Model) rollbackDeployment(namespace, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.k8sClient.RollbackDeployment(ctx, namespace, name); err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("failed to rollback deployment: %w", err)}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Rolled back %s to previous revision", name),
		}
	}
}

func (m *Model) scaleStatefulSet(namespace, name, replicaStr string) tea.Cmd {
	return func() tea.Msg {
		replicas, err := strconv.ParseInt(replicaStr, 10, 32)
		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("invalid replica count: %w", err)}
		}

		if replicas < 0 {
			return panels.ErrorMsg{Error: ErrInvalidReplicaCount}
		}

		ctx := context.Background()
		if err := m.k8sClient.ScaleStatefulSet(ctx, namespace, name, int32(replicas)); err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("failed to scale statefulset: %w", err)}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Scaled %s to %d replicas", name, replicas),
		}
	}
}

func (m *Model) updateHPAMinReplicas(namespace, name, replicaStr string) tea.Cmd {
	return func() tea.Msg {
		replicas, err := strconv.ParseInt(replicaStr, 10, 32)
		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("invalid replica count: %w", err)}
		}

		if replicas < 1 {
			return panels.ErrorMsg{Error: ErrMinReplicasTooLow}
		}

		ctx := context.Background()

		err = m.k8sClient.UpdateHPAMinReplicas(ctx, namespace, name, int32(replicas))
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to update HPA min replicas: %w", err),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Updated %s min replicas to %d", name, replicas),
		}
	}
}

func (m *Model) updateHPAMaxReplicas(namespace, name, replicaStr string) tea.Cmd {
	return func() tea.Msg {
		replicas, err := strconv.ParseInt(replicaStr, 10, 32)
		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("invalid replica count: %w", err)}
		}

		if replicas < 1 {
			return panels.ErrorMsg{Error: ErrMaxReplicasTooLow}
		}

		ctx := context.Background()

		err = m.k8sClient.UpdateHPAMaxReplicas(ctx, namespace, name, int32(replicas))
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("failed to update HPA max replicas: %w", err),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Updated %s max replicas to %d", name, replicas),
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

		return panels.StatusMsg{
			Message: fmt.Sprintf("Port forwarding %d -> %s:%d", localPort, podName, remotePort),
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

	m.statusBar.SetMessage("Refreshed all panels")

	return tea.Batch(cmds...)
}

func (m *Model) handleContainerSelect(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		}
	case "down", "j":
		if m.selectIdx < len(m.execContainers)-1 {
			m.selectIdx++
		}
	case "enter":
		if m.selectIdx < len(m.execContainers) {
			container := m.execContainers[m.selectIdx]
			m.viewMode = ViewNormal

			return m, m.execIntoPod(m.execNamespace, m.execPodName, container)
		}

		m.viewMode = ViewNormal
	case "esc":
		m.viewMode = ViewNormal
	}

	return m, nil
}

func (m *Model) execIntoPod(namespace, podName, container string) tea.Cmd {
	args := []string{
		"exec", "-it",
		"-n", namespace,
		podName,
		"-c", container,
		"--", "/bin/sh", "-c",
		"if command -v bash > /dev/null; then exec bash; else exec sh; fi",
	}

	c := tea.ExecProcess(
		newKubectlCmd(args...),
		func(err error) tea.Msg {
			if err != nil {
				return panels.ErrorMsg{Error: fmt.Errorf("exec failed: %w", err)}
			}

			return panels.StatusMsg{
				Message: fmt.Sprintf("Exited shell in %s/%s", podName, container),
			}
		},
	)

	return c
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

	// Create temp file for editing
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

	// Spawn editor process
	cmd := exec.Command(editor, tmpFile) //nolint:gosec,noctx
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer func() { _ = os.Remove(tmpFile) }()

		if err != nil {
			return panels.ErrorMsg{Error: fmt.Errorf("editor failed: %w", err)}
		}

		// Apply the edited YAML using kubectl
		applyCmd := exec.Command("kubectl", "apply", "-f", tmpFile) //nolint:gosec,noctx

		output, err := applyCmd.CombinedOutput()
		if err != nil {
			return panels.ErrorMsg{
				Error: fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(string(output))),
			}
		}

		return panels.StatusWithRefreshMsg{
			Message: fmt.Sprintf("Applied changes to %s", name),
		}
	})
}

func newKubectlCmd(args ...string) *exec.Cmd {
	return exec.Command("kubectl", args...) //nolint:noctx
}
