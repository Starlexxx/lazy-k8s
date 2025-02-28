package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoadNodes loads nodes from the Kubernetes API and updates the nodes view
func (a *App) LoadNodes() error {
	// Clear the current list (except headers)
	for row := 1; row < a.NodesView.GetRowCount(); row++ {
		for col := 0; col < a.NodesView.GetColumnCount(); col++ {
			a.NodesView.SetCell(row, col, tview.NewTableCell(""))
		}
	}

	// Reset table to just the header row
	for i := a.NodesView.GetRowCount() - 1; i > 0; i-- {
		a.NodesView.RemoveRow(i)
	}

	// Add loading message
	a.NodesView.SetCell(1, 0, tview.NewTableCell("Loading nodes...").SetTextColor(tcell.ColorWhite))
	a.TviewApp.Draw()

	// Get nodes from API
	nodes, err := a.Client.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		a.NodesView.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)).SetTextColor(tcell.ColorRed))
		return err
	}

	// Clear the loading message
	a.NodesView.RemoveRow(1)

	// If no nodes found
	if len(nodes.Items) == 0 {
		a.NodesView.SetCell(1, 0, tview.NewTableCell("No nodes found").SetTextColor(tcell.ColorWhite))
		return nil
	}

	// Fill the nodes table
	for i, node := range nodes.Items {
		row := i + 1

		// Ensure row exists
		for a.NodesView.GetRowCount() <= row {
			a.NodesView.SetCell(a.NodesView.GetRowCount(), 0, tview.NewTableCell(""))
		}

		// Add name
		a.NodesView.SetCell(row, 0, tview.NewTableCell(node.Name).SetTextColor(tcell.ColorWhite))

		// Add status
		status := getNodeStatus(node)
		var color tcell.Color
		if status == "Ready" {
			color = tcell.ColorGreen
		} else {
			color = tcell.ColorRed
		}
		a.NodesView.SetCell(row, 1, tview.NewTableCell(status).SetTextColor(color))

		// Add roles
		roles := getNodeRoles(node)
		a.NodesView.SetCell(row, 2, tview.NewTableCell(roles).SetTextColor(tcell.ColorWhite))

		// Add version
		version := node.Status.NodeInfo.KubeletVersion
		a.NodesView.SetCell(row, 3, tview.NewTableCell(version).SetTextColor(tcell.ColorWhite))

		// Add age
		age := formatAge(node.CreationTimestamp.Time)
		a.NodesView.SetCell(row, 4, tview.NewTableCell(age).SetTextColor(tcell.ColorWhite))
	}

	return nil
}

// getNodeStatus returns the status of a node
func getNodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

// getNodeRoles returns the roles of a node
func getNodeRoles(node corev1.Node) string {
	roles := []string{}

	// Check common role labels
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		roles = append(roles, "master")
	}
	if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
		roles = append(roles, "control-plane")
	}
	if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
		roles = append(roles, "worker")
	}

	// Check for other role labels
	for label := range node.Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			if role != "master" && role != "control-plane" && role != "worker" {
				roles = append(roles, role)
			}
		}
	}

	if len(roles) == 0 {
		return "<none>"
	}

	return strings.Join(roles, ",")
}

// ShowNodeDetails shows detailed information about the selected node
func (a *App) ShowNodeDetails() {
	row, _ := a.NodesView.GetSelection()
	if row <= 0 || row >= a.NodesView.GetRowCount() {
		return
	}

	nameCell := a.NodesView.GetCell(row, 0)
	if nameCell == nil {
		return
	}

	nodeName := nameCell.Text

	// Check if client is available
	if a.Client == nil {
		// In demo mode, show placeholder details
		a.showDemoNodeDetails(nodeName)
		return
	}

	// Fetch node details
	node, err := a.Client.ClientSet.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Error fetching node details: %v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(_ int, _ string) {
				a.Pages.RemovePage("nodeDetailsError")
			})
		a.Pages.AddPage("nodeDetailsError", modal, true, true)
		return
	}

	// Create a text view for node details
	detailsView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			a.TviewApp.Draw()
		})

	// Add node details
	fmt.Fprintf(detailsView, "[yellow]Node: [white]%s\n", node.Name)
	fmt.Fprintf(detailsView, "[yellow]Roles: [white]%s\n", getNodeRoles(*node))
	fmt.Fprintf(detailsView, "[yellow]Status: [white]%s\n", getNodeStatus(*node))
	fmt.Fprintf(detailsView, "[yellow]Creation: [white]%s\n\n", node.CreationTimestamp.Format(time.RFC3339))

	// Node info
	nodeInfo := node.Status.NodeInfo
	fmt.Fprintf(detailsView, "[yellow]Kubelet Version: [white]%s\n", nodeInfo.KubeletVersion)
	fmt.Fprintf(detailsView, "[yellow]OS: [white]%s\n", nodeInfo.OperatingSystem)
	fmt.Fprintf(detailsView, "[yellow]Architecture: [white]%s\n", nodeInfo.Architecture)
	fmt.Fprintf(detailsView, "[yellow]Container Runtime: [white]%s\n", nodeInfo.ContainerRuntimeVersion)
	fmt.Fprintf(detailsView, "[yellow]Kernel Version: [white]%s\n\n", nodeInfo.KernelVersion)

	// Node addresses
	fmt.Fprintf(detailsView, "[yellow]Addresses:\n")
	for _, address := range node.Status.Addresses {
		fmt.Fprintf(detailsView, "  [green]%s: [white]%s\n", address.Type, address.Address)
	}

	// Node conditions
	fmt.Fprintf(detailsView, "\n[yellow]Conditions:\n")
	for _, condition := range node.Status.Conditions {
		statusColor := "red"
		if condition.Status == corev1.ConditionTrue {
			statusColor = "green"
		}

		fmt.Fprintf(detailsView, "  [green]%s: [%s]%s\n",
			condition.Type, statusColor, condition.Status)
		fmt.Fprintf(detailsView, "    [yellow]Reason: [white]%s\n", condition.Reason)
		fmt.Fprintf(detailsView, "    [yellow]Message: [white]%s\n", condition.Message)
		fmt.Fprintf(detailsView, "    [yellow]Last Update: [white]%s\n",
			condition.LastTransitionTime.Format(time.RFC3339))
	}

	// Node capacity
	fmt.Fprintf(detailsView, "\n[yellow]Capacity:\n")
	fmt.Fprintf(detailsView, "  [green]CPU: [white]%s\n", node.Status.Capacity.Cpu().String())
	fmt.Fprintf(detailsView, "  [green]Memory: [white]%s\n", node.Status.Capacity.Memory().String())
	fmt.Fprintf(detailsView, "  [green]Pods: [white]%s\n", node.Status.Capacity.Pods().String())

	// Node allocatable
	fmt.Fprintf(detailsView, "\n[yellow]Allocatable:\n")
	fmt.Fprintf(detailsView, "  [green]CPU: [white]%s\n", node.Status.Allocatable.Cpu().String())
	fmt.Fprintf(detailsView, "  [green]Memory: [white]%s\n", node.Status.Allocatable.Memory().String())
	fmt.Fprintf(detailsView, "  [green]Pods: [white]%s\n", node.Status.Allocatable.Pods().String())

	// Labels
	fmt.Fprintf(detailsView, "\n[yellow]Labels:\n")
	if len(node.Labels) == 0 {
		fmt.Fprintf(detailsView, "  [white]<none>\n")
	} else {
		for key, value := range node.Labels {
			fmt.Fprintf(detailsView, "  [green]%s: [white]%s\n", key, value)
		}
	}

	// Create a frame for the details
	frame := tview.NewFrame(detailsView).
		SetBorders(1, 1, 1, 1, 0, 0).
		AddText("Node Details", true, tview.AlignCenter, tcell.ColorYellow).
		AddText("ESC: Close, ↑/↓: Scroll", false, tview.AlignCenter, tcell.ColorWhite)

	// Capture input events for the details view
	detailsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.Pages.RemovePage("nodeDetails")
			return nil
		}
		return event
	})

	// Add the details page
	a.Pages.AddPage("nodeDetails", frame, true, true)
}

// showDemoNodeDetails shows a placeholder node details page for demo mode
func (a *App) showDemoNodeDetails(nodeName string) {
	// Create a text view for node details
	detailsView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			a.TviewApp.Draw()
		})

	// Add node details
	fmt.Fprintf(detailsView, "[yellow]Node: [white]%s\n", nodeName)
	fmt.Fprintf(detailsView, "[yellow]Roles: [white]control-plane,master\n")
	fmt.Fprintf(detailsView, "[yellow]Status: [white]Ready\n")
	fmt.Fprintf(detailsView, "[yellow]Creation: [white]%s\n\n", time.Now().Add(-30*24*time.Hour).Format(time.RFC3339))

	// Node info
	fmt.Fprintf(detailsView, "[yellow]Kubelet Version: [white]v1.25.0\n")
	fmt.Fprintf(detailsView, "[yellow]OS: [white]linux\n")
	fmt.Fprintf(detailsView, "[yellow]Architecture: [white]amd64\n")
	fmt.Fprintf(detailsView, "[yellow]Container Runtime: [white]containerd://1.6.12\n")
	fmt.Fprintf(detailsView, "[yellow]Kernel Version: [white]5.15.0-generic\n\n")

	// Node addresses
	fmt.Fprintf(detailsView, "[yellow]Addresses:\n")
	fmt.Fprintf(detailsView, "  [green]InternalIP: [white]10.0.0.1\n")
	fmt.Fprintf(detailsView, "  [green]Hostname: [white]demo-node\n")

	// Node conditions
	fmt.Fprintf(detailsView, "\n[yellow]Conditions:\n")
	fmt.Fprintf(detailsView, "  [green]Ready: [green]True\n")
	fmt.Fprintf(detailsView, "    [yellow]Reason: [white]KubeletReady\n")
	fmt.Fprintf(detailsView, "    [yellow]Message: [white]kubelet is posting ready status\n")
	fmt.Fprintf(detailsView, "    [yellow]Last Update: [white]%s\n",
		time.Now().Add(-1*time.Hour).Format(time.RFC3339))

	// Node capacity
	fmt.Fprintf(detailsView, "\n[yellow]Capacity:\n")
	fmt.Fprintf(detailsView, "  [green]CPU: [white]4\n")
	fmt.Fprintf(detailsView, "  [green]Memory: [white]8Gi\n")
	fmt.Fprintf(detailsView, "  [green]Pods: [white]110\n")

	// Node allocatable
	fmt.Fprintf(detailsView, "\n[yellow]Allocatable:\n")
	fmt.Fprintf(detailsView, "  [green]CPU: [white]3800m\n")
	fmt.Fprintf(detailsView, "  [green]Memory: [white]7Gi\n")
	fmt.Fprintf(detailsView, "  [green]Pods: [white]110\n")

	// Labels
	fmt.Fprintf(detailsView, "\n[yellow]Labels:\n")
	fmt.Fprintf(detailsView, "  [green]node-role.kubernetes.io/control-plane: [white]\n")
	fmt.Fprintf(detailsView, "  [green]node-role.kubernetes.io/master: [white]\n")
	fmt.Fprintf(detailsView, "  [green]kubernetes.io/hostname: [white]demo-node\n")

	fmt.Fprintf(detailsView, "\n[red]DEMO MODE: No actual Kubernetes connection\n")

	// Create a frame for the details
	frame := tview.NewFrame(detailsView).
		SetBorders(1, 1, 1, 1, 0, 0).
		AddText("Node Details (Demo)", true, tview.AlignCenter, tcell.ColorYellow).
		AddText("ESC: Close, ↑/↓: Scroll", false, tview.AlignCenter, tcell.ColorWhite)

	// Capture input events for the details view
	detailsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.Pages.RemovePage("nodeDetails")
			return nil
		}
		return event
	})

	// Add the details page
	a.Pages.AddPage("nodeDetails", frame, true, true)
}
