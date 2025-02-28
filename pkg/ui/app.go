package ui

import (
	"github.com/Starlexxx/lazy-k8s/pkg/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ViewType represents the type of view currently displayed
type ViewType string

// View types
const (
	ViewPods  ViewType = "pods"
	ViewNodes ViewType = "nodes"
)

// App represents the main application UI
type App struct {
	Client    *client.K8sClient
	TviewApp  *tview.Application
	Pages     *tview.Pages
	StatusBar *tview.TextView
	InfoBar   *tview.TextView

	// Views
	PodsView    *tview.Table
	NodesView   *tview.Table
	CurrentView ViewType

	// Layout
	MainFlex *tview.Flex

	// State
	Namespace string
}

// NewApp creates a new UI application
func NewApp(client *client.K8sClient) *App {
	app := &App{
		Client:      client,
		TviewApp:    tview.NewApplication(),
		Pages:       tview.NewPages(),
		Namespace:   "default",
		CurrentView: ViewPods,
	}

	// Initialize UI components
	app.initUI()

	// Setup keybindings
	app.setupKeybindings()

	return app
}

// initUI initializes all UI components
func (a *App) initUI() {
	// Create header
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	header.SetText("[yellow]lazy-k8s [white]- Kubernetes Terminal UI")

	// Create status bar
	a.StatusBar = tview.NewTextView().
		SetDynamicColors(true)
	a.updateStatusBar()

	// Create info bar
	a.InfoBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.updateInfoBar()

	// Initialize views
	a.initPodsView()
	a.initNodesView()

	// Create main content
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Create main layout
	a.MainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 1, false).
		AddItem(mainContent, 0, 1, true).
		AddItem(a.InfoBar, 1, 1, false).
		AddItem(a.StatusBar, 1, 1, false)

	// Set initial view
	mainContent.AddItem(a.PodsView, 0, 1, true)

	// Add main layout to pages
	a.Pages.AddPage("main", a.MainFlex, true, true)

	// Set root
	a.TviewApp.SetRoot(a.Pages, true)
}

// initPodsView initializes the pods view
func (a *App) initPodsView() {
	a.PodsView = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	// Add headers
	headers := []string{"NAMESPACE", "NAME", "STATUS", "READY", "RESTARTS", "AGE"}
	for i, header := range headers {
		a.PodsView.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Placeholder for now
	a.PodsView.SetCell(1, 0, tview.NewTableCell("Loading pods...").SetTextColor(tcell.ColorWhite))

	a.PodsView.SetSelectedFunc(func(row, _ int) {
		// Handle pod selection
		if row > 0 {
			a.ShowPodDetails()
		}
	})
}

// initNodesView initializes the nodes view
func (a *App) initNodesView() {
	a.NodesView = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	// Add headers
	headers := []string{"NAME", "STATUS", "ROLES", "VERSION", "AGE"}
	for i, header := range headers {
		a.NodesView.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Placeholder for now
	a.NodesView.SetCell(1, 0, tview.NewTableCell("Loading nodes...").SetTextColor(tcell.ColorWhite))

	a.NodesView.SetSelectedFunc(func(row, _ int) {
		// Handle node selection
		if row > 0 {
			a.ShowNodeDetails()
		}
	})
}

// setupKeybindings configures all key bindings
func (a *App) setupKeybindings() {
	a.TviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Global key bindings
		if event.Key() == tcell.KeyEscape {
			a.TviewApp.Stop()
			return nil
		}

		// View-specific key bindings
		switch event.Rune() {
		case 'q':
			a.TviewApp.Stop()
			return nil
		case '1':
			a.switchToPodsView()
			return nil
		case '2':
			a.switchToNodesView()
			return nil
		case 'r':
			// Refresh the current view
			a.refreshCurrentView()
			return nil
		}

		return event
	})
}

// switchToPodsView switches to the pods view
func (a *App) switchToPodsView() {
	if a.CurrentView == ViewPods {
		return
	}

	mainContent, ok := a.MainFlex.GetItem(1).(*tview.Flex)
	if !ok {
		return
	}

	mainContent.Clear()
	mainContent.AddItem(a.PodsView, 0, 1, true)
	a.CurrentView = ViewPods
	a.updateInfoBar()
	a.TviewApp.SetFocus(a.PodsView)
}

// switchToNodesView switches to the nodes view
func (a *App) switchToNodesView() {
	if a.CurrentView == ViewNodes {
		return
	}

	mainContent, ok := a.MainFlex.GetItem(1).(*tview.Flex)
	if !ok {
		return
	}

	mainContent.Clear()
	mainContent.AddItem(a.NodesView, 0, 1, true)
	a.CurrentView = ViewNodes
	a.updateInfoBar()
	a.TviewApp.SetFocus(a.NodesView)
}

// refreshCurrentView refreshes the data in the current view
func (a *App) refreshCurrentView() {
	switch a.CurrentView {
	case ViewPods:
		if err := a.LoadPods(); err != nil {
			// Use StatusBar to show error
			a.StatusBar.SetText("[red]Error loading pods: " + err.Error())
		}
	case ViewNodes:
		if err := a.LoadNodes(); err != nil {
			// Use StatusBar to show error
			a.StatusBar.SetText("[red]Error loading nodes: " + err.Error())
		}
	}
}

// updateStatusBar updates the content of the status bar
func (a *App) updateStatusBar() {
	if a.Client != nil && a.Client.CurrentContext != "" {
		a.StatusBar.SetText(
			"[yellow]Context:[white] " + a.Client.CurrentContext +
				" | [yellow]Namespace:[white] " + a.Namespace +
				" | [yellow]Press[white] ? [yellow]for help",
		)
	} else {
		a.StatusBar.SetText("[red]Not connected to any Kubernetes cluster | [yellow]Demo Mode")
	}
}

// updateInfoBar updates the content of the info bar
func (a *App) updateInfoBar() {
	var helpText string
	switch a.CurrentView {
	case ViewPods:
		helpText = "[1]Pods [2]Nodes [r]Refresh [q]Quit | Press Enter on pod for details"
	case ViewNodes:
		helpText = "[1]Pods [2]Nodes [r]Refresh [q]Quit | Press Enter on node for details"
	}
	a.InfoBar.SetText(helpText)
}

// Run starts the UI application
func (a *App) Run() error {
	a.loadData()
	return a.TviewApp.Run()
}

// loadData loads initial data
func (a *App) loadData() {
	// Load pods and nodes data asynchronously
	if a.Client == nil {
		// In demo mode, just show placeholder data
		a.PodsView.SetCell(1, 0, tview.NewTableCell("demo-namespace").SetTextColor(tcell.ColorWhite))
		a.PodsView.SetCell(1, 1, tview.NewTableCell("demo-pod").SetTextColor(tcell.ColorWhite))
		a.PodsView.SetCell(1, 2, tview.NewTableCell("Running").SetTextColor(tcell.ColorGreen))
		a.PodsView.SetCell(1, 3, tview.NewTableCell("1/1").SetTextColor(tcell.ColorWhite))
		a.PodsView.SetCell(1, 4, tview.NewTableCell("0").SetTextColor(tcell.ColorWhite))
		a.PodsView.SetCell(1, 5, tview.NewTableCell("1d").SetTextColor(tcell.ColorWhite))

		a.NodesView.SetCell(1, 0, tview.NewTableCell("demo-node").SetTextColor(tcell.ColorWhite))
		a.NodesView.SetCell(1, 1, tview.NewTableCell("Ready").SetTextColor(tcell.ColorGreen))
		a.NodesView.SetCell(1, 2, tview.NewTableCell("control-plane").SetTextColor(tcell.ColorWhite))
		a.NodesView.SetCell(1, 3, tview.NewTableCell("v1.25.0").SetTextColor(tcell.ColorWhite))
		a.NodesView.SetCell(1, 4, tview.NewTableCell("30d").SetTextColor(tcell.ColorWhite))
		return
	}

	go func() {
		err := a.LoadPods()
		if err != nil {
			a.TviewApp.QueueUpdateDraw(func() {
				a.StatusBar.SetText("[red]Error loading pods: " + err.Error())
			})
		}
	}()

	go func() {
		err := a.LoadNodes()
		if err != nil {
			a.TviewApp.QueueUpdateDraw(func() {
				a.StatusBar.SetText("[red]Error loading nodes: " + err.Error())
			})
		}
	}()
}
