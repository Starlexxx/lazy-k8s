package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoadPods loads pods from the Kubernetes API and updates the pods view
func (a *App) LoadPods() error {
	// Clear the current list (except headers)
	for row := 1; row < a.PodsView.GetRowCount(); row++ {
		for col := 0; col < a.PodsView.GetColumnCount(); col++ {
			a.PodsView.SetCell(row, col, tview.NewTableCell(""))
		}
	}

	// Reset table to just the header row
	for i := a.PodsView.GetRowCount() - 1; i > 0; i-- {
		a.PodsView.RemoveRow(i)
	}

	// Add loading message
	a.PodsView.SetCell(1, 0, tview.NewTableCell("Loading pods...").SetTextColor(tcell.ColorWhite))
	a.TviewApp.Draw()

	// Get pods from API
	pods, err := a.Client.ClientSet.CoreV1().Pods(a.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		a.PodsView.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)).SetTextColor(tcell.ColorRed))
		return err
	}

	// Clear the loading message
	a.PodsView.RemoveRow(1)

	// If no pods found
	if len(pods.Items) == 0 {
		a.PodsView.SetCell(1, 0, tview.NewTableCell("No pods found").SetTextColor(tcell.ColorWhite))
		return nil
	}

	// Fill the pods table
	for i, pod := range pods.Items {
		row := i + 1

		// Ensure row exists
		for a.PodsView.GetRowCount() <= row {
			a.PodsView.SetCell(a.PodsView.GetRowCount(), 0, tview.NewTableCell(""))
		}

		// Add namespace
		a.PodsView.SetCell(row, 0, tview.NewTableCell(pod.Namespace).SetTextColor(tcell.ColorWhite))

		// Add name
		a.PodsView.SetCell(row, 1, tview.NewTableCell(pod.Name).SetTextColor(tcell.ColorWhite))

		// Add status
		status := string(pod.Status.Phase)
		color := tcell.ColorWhite
		switch pod.Status.Phase {
		case corev1.PodRunning:
			color = tcell.ColorGreen
		case corev1.PodPending:
			color = tcell.ColorYellow
		case corev1.PodFailed:
			color = tcell.ColorRed
		case corev1.PodSucceeded:
			color = tcell.ColorBlue
		}
		a.PodsView.SetCell(row, 2, tview.NewTableCell(status).SetTextColor(color))

		// Add ready
		ready := 0
		total := len(pod.Status.ContainerStatuses)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				ready++
			}
		}
		readyStr := fmt.Sprintf("%d/%d", ready, total)
		a.PodsView.SetCell(row, 3, tview.NewTableCell(readyStr).SetTextColor(tcell.ColorWhite))

		// Add restarts
		restarts := 0
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restarts += int(containerStatus.RestartCount)
		}
		a.PodsView.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%d", restarts)).SetTextColor(tcell.ColorWhite))

		// Add age
		age := formatAge(pod.CreationTimestamp.Time)
		a.PodsView.SetCell(row, 5, tview.NewTableCell(age).SetTextColor(tcell.ColorWhite))
	}

	return nil
}

// ShowPodDetails shows detailed information about the selected pod
func (a *App) ShowPodDetails() {
	row, _ := a.PodsView.GetSelection()
	if row <= 0 || row >= a.PodsView.GetRowCount() {
		return
	}

	nameCell := a.PodsView.GetCell(row, 1)
	if nameCell == nil {
		return
	}

	podName := nameCell.Text
	namespaceCell := a.PodsView.GetCell(row, 0)
	namespace := ""
	if namespaceCell != nil {
		namespace = namespaceCell.Text
	} else {
		namespace = a.Namespace
	}

	// Check if client is available
	if a.Client == nil {
		// In demo mode, show placeholder details
		a.showDemoPodDetails(podName, namespace)
		return
	}

	// Fetch pod details
	pod, err := a.Client.ClientSet.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		a.showPodDetailsError(err)
		return
	}

	// Create a text view for pod details
	detailsView := a.createPodDetailsView(pod)

	// Create a frame for the details
	frame := tview.NewFrame(detailsView).
		SetBorders(1, 1, 1, 1, 0, 0).
		AddText("Pod Details", true, tview.AlignCenter, tcell.ColorYellow).
		AddText("ESC: Close, ↑/↓: Scroll", false, tview.AlignCenter, tcell.ColorWhite)

	// Capture input events for the details view
	detailsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.Pages.RemovePage("podDetails")
			return nil
		}
		return event
	})

	// Add the details page
	a.Pages.AddPage("podDetails", frame, true, true)
}

// showPodDetailsError displays an error modal when pod details cannot be fetched
func (a *App) showPodDetailsError(err error) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Error fetching pod details: %v", err)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			a.Pages.RemovePage("podDetailsError")
		})
	a.Pages.AddPage("podDetailsError", modal, true, true)
}

// createPodDetailsView creates and populates a TextView with pod details
func (a *App) createPodDetailsView(pod *corev1.Pod) *tview.TextView {
	detailsView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			a.TviewApp.Draw()
		})

	// Add pod details
	fmt.Fprintf(detailsView, "[yellow]Pod: [white]%s\n", pod.Name)
	fmt.Fprintf(detailsView, "[yellow]Namespace: [white]%s\n", pod.Namespace)
	fmt.Fprintf(detailsView, "[yellow]Status: [white]%s\n", pod.Status.Phase)
	fmt.Fprintf(detailsView, "[yellow]Host IP: [white]%s\n", pod.Status.HostIP)
	fmt.Fprintf(detailsView, "[yellow]Pod IP: [white]%s\n", pod.Status.PodIP)
	fmt.Fprintf(detailsView, "[yellow]QoS: [white]%s\n", pod.Status.QOSClass)
	fmt.Fprintf(detailsView, "[yellow]Creation: [white]%s\n\n", pod.CreationTimestamp.Format(time.RFC3339))

	a.addLabelsToDetailsView(detailsView, pod.Labels)
	a.addContainersToDetailsView(detailsView, pod)

	return detailsView
}

// addLabelsToDetailsView adds label information to the details view
func (a *App) addLabelsToDetailsView(view *tview.TextView, labels map[string]string) {
	fmt.Fprintf(view, "[yellow]Labels:\n")
	if len(labels) == 0 {
		fmt.Fprintf(view, "  [white]<none>\n")
	} else {
		for key, value := range labels {
			fmt.Fprintf(view, "  [green]%s: [white]%s\n", key, value)
		}
	}
}

// addContainersToDetailsView adds container information to the details view
func (a *App) addContainersToDetailsView(view *tview.TextView, pod *corev1.Pod) {
	fmt.Fprintf(view, "\n[yellow]Containers:\n")
	for _, container := range pod.Spec.Containers {
		fmt.Fprintf(view, "  [green]%s:\n", container.Name)
		fmt.Fprintf(view, "    [yellow]Image: [white]%s\n", container.Image)

		// Get container status
		var containerStatus *corev1.ContainerStatus
		for i := range pod.Status.ContainerStatuses {
			if pod.Status.ContainerStatuses[i].Name == container.Name {
				containerStatus = &pod.Status.ContainerStatuses[i]
				break
			}
		}

		if containerStatus != nil {
			a.addContainerStatusToDetailsView(view, containerStatus)
		}

		fmt.Fprintf(view, "\n")
	}
}

// addContainerStatusToDetailsView adds container status information to the details view
func (a *App) addContainerStatusToDetailsView(view *tview.TextView, containerStatus *corev1.ContainerStatus) {
	fmt.Fprintf(view, "    [yellow]Ready: [white]%v\n", containerStatus.Ready)
	fmt.Fprintf(view, "    [yellow]Restarts: [white]%d\n", containerStatus.RestartCount)

	// Container state
	switch {
	case containerStatus.State.Running != nil:
		fmt.Fprintf(view, "    [yellow]State: [green]Running\n")
		fmt.Fprintf(view, "    [yellow]Started: [white]%s\n",
			containerStatus.State.Running.StartedAt.Format(time.RFC3339))
	case containerStatus.State.Terminated != nil:
		fmt.Fprintf(view, "    [yellow]State: [red]Terminated\n")
		fmt.Fprintf(view, "    [yellow]Reason: [white]%s\n",
			containerStatus.State.Terminated.Reason)
		fmt.Fprintf(view, "    [yellow]Exit Code: [white]%d\n",
			containerStatus.State.Terminated.ExitCode)
	case containerStatus.State.Waiting != nil:
		fmt.Fprintf(view, "    [yellow]State: [yellow]Waiting\n")
		fmt.Fprintf(view, "    [yellow]Reason: [white]%s\n",
			containerStatus.State.Waiting.Reason)
	}
}

// showDemoPodDetails shows a placeholder pod details page for demo mode
func (a *App) showDemoPodDetails(podName, namespace string) {
	// Create a text view for pod details
	detailsView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			a.TviewApp.Draw()
		})

	// Add pod details
	fmt.Fprintf(detailsView, "[yellow]Pod: [white]%s\n", podName)
	fmt.Fprintf(detailsView, "[yellow]Namespace: [white]%s\n", namespace)
	fmt.Fprintf(detailsView, "[yellow]Status: [white]Running\n")
	fmt.Fprintf(detailsView, "[yellow]Host IP: [white]10.0.0.1\n")
	fmt.Fprintf(detailsView, "[yellow]Pod IP: [white]10.1.0.1\n")
	fmt.Fprintf(detailsView, "[yellow]QoS: [white]Guaranteed\n")
	fmt.Fprintf(detailsView, "[yellow]Creation: [white]%s\n\n", time.Now().Add(-24*time.Hour).Format(time.RFC3339))

	fmt.Fprintf(detailsView, "[yellow]Labels:\n")
	fmt.Fprintf(detailsView, "  [green]app: [white]demo\n")
	fmt.Fprintf(detailsView, "  [green]environment: [white]demo\n")

	fmt.Fprintf(detailsView, "\n[yellow]Containers:\n")
	fmt.Fprintf(detailsView, "  [green]main:\n")
	fmt.Fprintf(detailsView, "    [yellow]Image: [white]nginx:latest\n")
	fmt.Fprintf(detailsView, "    [yellow]Ready: [white]true\n")
	fmt.Fprintf(detailsView, "    [yellow]Restarts: [white]0\n")
	fmt.Fprintf(detailsView, "    [yellow]State: [green]Running\n")
	fmt.Fprintf(detailsView, "    [yellow]Started: [white]%s\n",
		time.Now().Add(-23*time.Hour).Format(time.RFC3339))

	fmt.Fprintf(detailsView, "\n[red]DEMO MODE: No actual Kubernetes connection\n")

	// Create a frame for the details
	frame := tview.NewFrame(detailsView).
		SetBorders(1, 1, 1, 1, 0, 0).
		AddText("Pod Details (Demo)", true, tview.AlignCenter, tcell.ColorYellow).
		AddText("ESC: Close, ↑/↓: Scroll", false, tview.AlignCenter, tcell.ColorWhite)

	// Capture input events for the details view
	detailsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.Pages.RemovePage("podDetails")
			return nil
		}
		return event
	})

	// Add the details page
	a.Pages.AddPage("podDetails", frame, true, true)
}

// Helper function to format the age of resources
func formatAge(creationTime time.Time) string {
	duration := time.Since(creationTime)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	case duration < 30*24*time.Hour:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	case duration < 365*24*time.Hour:
		return fmt.Sprintf("%dM", int(duration.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy", int(duration.Hours()/(24*365)))
	}
}
