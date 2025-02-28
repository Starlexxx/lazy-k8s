package ui

import (
	"testing"

	"github.com/Starlexxx/lazy-k8s/pkg/client"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewApp(t *testing.T) {
	// Create a fake client
	fakeClientset := fake.NewSimpleClientset()
	fakeClient := &client.K8sClient{
		ClientSet:      fakeClientset,
		CurrentContext: "test-context",
	}

	// Create a new application
	app := NewApp(fakeClient)

	// Check that the application is successfully created
	if app == nil {
		t.Fatal("Application should not be nil")
	}

	// Check application settings
	if app.Client != fakeClient {
		t.Error("Client was not properly set in the application")
	}

	if app.TviewApp == nil {
		t.Error("TviewApp was not initialized")
	}

	if app.Pages == nil {
		t.Error("Pages was not initialized")
	}

	if app.Namespace != "default" {
		t.Errorf("Incorrect default namespace: %s, expected: default", app.Namespace)
	}

	if app.CurrentView != ViewPods {
		t.Errorf("Incorrect default view: %s, expected: pods", app.CurrentView)
	}
}

func TestInitUIComponents(t *testing.T) {
	// Create a fake client
	fakeClientset := fake.NewSimpleClientset()
	fakeClient := &client.K8sClient{
		ClientSet:      fakeClientset,
		CurrentContext: "test-context",
	}

	// Create a new application
	app := NewApp(fakeClient)

	// Check that UI components were initialized
	if app.PodsView == nil {
		t.Error("PodsView was not initialized")
	}

	if app.NodesView == nil {
		t.Error("NodesView was not initialized")
	}

	if app.StatusBar == nil {
		t.Error("StatusBar was not initialized")
	}

	if app.InfoBar == nil {
		t.Error("InfoBar was not initialized")
	}

	if app.MainFlex == nil {
		t.Error("MainFlex was not initialized")
	}

	// Check pod table headers
	if app.PodsView.GetCell(0, 0) == nil {
		t.Error("Pod table header was not initialized")
	} else if app.PodsView.GetCell(0, 0).Text != "NAMESPACE" {
		t.Errorf("Incorrect pod table header: %s, expected: NAMESPACE", app.PodsView.GetCell(0, 0).Text)
	}

	// Check node table headers
	if app.NodesView.GetCell(0, 0) == nil {
		t.Error("Node table header was not initialized")
	} else if app.NodesView.GetCell(0, 0).Text != "NAME" {
		t.Errorf("Incorrect node table header: %s, expected: NAME", app.NodesView.GetCell(0, 0).Text)
	}
}

func TestViewSwitching(t *testing.T) {
	// Create a fake client
	fakeClientset := fake.NewSimpleClientset()
	fakeClient := &client.K8sClient{
		ClientSet:      fakeClientset,
		CurrentContext: "test-context",
	}

	// Create a new application
	app := NewApp(fakeClient)

	// Check initial state
	if app.CurrentView != ViewPods {
		t.Errorf("Initial view should be ViewPods, but was %s", app.CurrentView)
	}

	// Switch to nodes view
	app.switchToNodesView()

	// Check that the view has changed
	if app.CurrentView != ViewNodes {
		t.Errorf("After switchToNodesView the view should be ViewNodes, but was %s", app.CurrentView)
	}

	// Switch back to pods view
	app.switchToPodsView()

	// Check that the view has changed back
	if app.CurrentView != ViewPods {
		t.Errorf("After switchToPodsView the view should be ViewPods, but was %s", app.CurrentView)
	}
}

func TestDemoMode(t *testing.T) {
	// Create an application without a client (demo mode)
	app := NewApp(nil)

	// Check that the application uses demo mode
	if app.Client != nil {
		t.Error("In demo mode Client should be nil")
	}

	// Load data in demo mode
	app.loadData()

	// Check that demo data was added to the pods table
	demoNamespaceCell := app.PodsView.GetCell(1, 0)
	if demoNamespaceCell == nil || demoNamespaceCell.Text != "demo-namespace" {
		t.Error("Demo data was not added to the pods table")
	}

	demoPodCell := app.PodsView.GetCell(1, 1)
	if demoPodCell == nil || demoPodCell.Text != "demo-pod" {
		t.Error("Demo data was not added to the pods table")
	}

	// Check that demo data was added to the nodes table
	demoNodeCell := app.NodesView.GetCell(1, 0)
	if demoNodeCell == nil || demoNodeCell.Text != "demo-node" {
		t.Error("Demo data was not added to the nodes table")
	}
}

func TestKeyBindings(t *testing.T) {
	app := NewApp(nil)

	// Check that the keyboard handler function has been registered
	if app.TviewApp.GetInputCapture() == nil {
		t.Error("Keyboard handler function was not registered")
	}

	// We would use simulateKeyEvent here in a real test situation,
	// but since we can't properly simulate key events in a unit test
	// without full tview initialization, we just check that the
	// input capture function exists
}
