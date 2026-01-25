package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazyk8s/lazy-k8s/internal/config"
	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/components"
	"github.com/lazyk8s/lazy-k8s/internal/ui/panels"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
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
)

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
	panels       []panels.Panel
	activePanelIdx int

	// Components
	header    *components.Header
	statusBar *components.StatusBar
	help      *components.Help
	confirm   *components.Confirm
	yamlView  *components.YamlViewer
	logView   *components.LogViewer
	search    *components.Search

	// State
	viewMode      ViewMode
	lastError     string
	lastStatus    string
	showAllNs     bool
	searchActive  bool
	searchQuery   string

	// Context/namespace switching
	contextList   []string
	namespaceList []string
	selectIdx     int
}

func NewModel(client *k8s.Client, cfg *config.Config) *Model {
	styles := theme.NewStyles(&cfg.Theme)
	keys := theme.NewKeyMap()

	m := &Model{
		k8sClient: client,
		config:    cfg,
		styles:    styles,
		keys:      keys,
		viewMode:  ViewNormal,
	}

	// Initialize header and status bar
	m.header = components.NewHeader(styles, client.CurrentContext(), client.CurrentNamespace())
	m.statusBar = components.NewStatusBar(styles)
	m.help = components.NewHelp(styles, keys)
	m.confirm = components.NewConfirm(styles)
	m.yamlView = components.NewYamlViewer(styles)
	m.logView = components.NewLogViewer(styles)
	m.search = components.NewSearch(styles)

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
			if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Help) || msg.String() == "q" {
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
					cmds = append(cmds, m.confirm.Action())
				}
			}
			return m, cmd

		case ViewContextSwitch:
			return m.handleContextSwitch(msg)

		case ViewNamespaceSwitch:
			return m.handleNamespaceSwitch(msg)
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
			if len(m.panels) > m.activePanelIdx {
				return m, m.panels[m.activePanelIdx].Refresh()
			}

		case key.Matches(msg, m.keys.AllNamespace):
			m.showAllNs = !m.showAllNs
			m.statusBar.SetMessage(fmt.Sprintf("All namespaces: %v", m.showAllNs))
			// Refresh all panels
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
		// Refresh the specified panel
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
	}

	// Update active panel
	if len(m.panels) > m.activePanelIdx {
		panel, cmd := m.panels[m.activePanelIdx].Update(msg)
		m.panels[m.activePanelIdx] = panel
		cmds = append(cmds, cmd)
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
	case ViewContextSwitch, ViewNamespaceSwitch:
		content = m.renderSwitchView()
	default:
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
	panelHeight := m.height - headerHeight - statusBarHeight
	if panelHeight < 3 {
		panelHeight = 3
	}

	// Search bar if active
	searchHeight := 0
	var searchView string
	if m.searchActive {
		searchHeight = 1
		searchView = m.search.View(m.width)
		panelHeight -= searchHeight
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
	// Each panel has 2 lines of border overhead (top + bottom)
	numPanels := len(m.panels)
	borderOverhead := numPanels * 2
	availableHeight := height - borderOverhead
	if availableHeight < numPanels {
		availableHeight = numPanels
	}
	panelHeight := availableHeight / numPanels

	var leftPanels []string
	for i, panel := range m.panels {
		panel.SetSize(leftPanelWidth, panelHeight)
		panel.SetFocused(i == m.activePanelIdx)
		leftPanels = append(leftPanels, panel.View())
	}

	leftView := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)

	// Right side: detail view (also has 2 lines border overhead)
	detailHeight := height - 2
	if detailHeight < 1 {
		detailHeight = 1
	}
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
	var title string
	var items []string
	var selectedIdx int

	if m.viewMode == ViewContextSwitch {
		title = "Switch Context"
		items = m.contextList
		selectedIdx = m.selectIdx
	} else {
		title = "Switch Namespace"
		items = m.namespaceList
		selectedIdx = m.selectIdx
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

func (m *Model) overlayView(base, overlay string) string {
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
