package k8s

import (
	"errors"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestNewPortForwarder(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := createTestClient(clientset)

	opts := PortForwardOptions{
		Namespace:  "default",
		PodName:    "test-pod",
		LocalPort:  8080,
		RemotePort: 80,
		StopCh:     make(chan struct{}),
		ReadyCh:    make(chan struct{}),
	}

	pf, err := client.NewPortForwarder(opts)
	if err != nil {
		t.Fatalf("NewPortForwarder returned unexpected error: %v", err)
	}

	if pf == nil {
		t.Fatal("NewPortForwarder returned nil")
	}
}

func TestNewPortForwarderDefaultNamespace(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := createTestClient(clientset)

	opts := PortForwardOptions{
		Namespace:  "", // Should default to client's namespace
		PodName:    "test-pod",
		LocalPort:  8080,
		RemotePort: 80,
	}

	pf, err := client.NewPortForwarder(opts)
	if err != nil {
		t.Fatalf("NewPortForwarder returned unexpected error: %v", err)
	}

	if pf.options.Namespace != "default" {
		t.Errorf("Namespace = %q, want %q", pf.options.Namespace, "default")
	}
}

func TestPortForwarderStop(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := createTestClient(clientset)

	stopCh := make(chan struct{})
	opts := PortForwardOptions{
		Namespace:  "default",
		PodName:    "test-pod",
		LocalPort:  8080,
		RemotePort: 80,
		StopCh:     stopCh,
		ReadyCh:    make(chan struct{}),
	}

	pf, err := client.NewPortForwarder(opts)
	if err != nil {
		t.Fatalf("NewPortForwarder returned unexpected error: %v", err)
	}

	// Stop should close the channel
	pf.Stop()

	// Verify channel is closed
	select {
	case <-stopCh:
		// Channel is closed as expected
	default:
		t.Error("Stop() should close the StopCh")
	}
}

func TestPortForwarderStopNilChannel(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := createTestClient(clientset)

	opts := PortForwardOptions{
		Namespace:  "default",
		PodName:    "test-pod",
		LocalPort:  8080,
		RemotePort: 80,
		StopCh:     nil, // nil channel
		ReadyCh:    make(chan struct{}),
	}

	pf, err := client.NewPortForwarder(opts)
	if err != nil {
		t.Fatalf("NewPortForwarder returned unexpected error: %v", err)
	}

	// Stop should not panic with nil channel
	pf.Stop()
}

func TestPortForwarderGetPortsNotStarted(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := createTestClient(clientset)

	opts := PortForwardOptions{
		Namespace:  "default",
		PodName:    "test-pod",
		LocalPort:  8080,
		RemotePort: 80,
	}

	pf, err := client.NewPortForwarder(opts)
	if err != nil {
		t.Fatalf("NewPortForwarder returned unexpected error: %v", err)
	}

	// GetPorts should return error when not started
	_, err = pf.GetPorts()
	if err == nil {
		t.Error("GetPorts should return error when forwarder not started")
	}

	if !errors.Is(err, ErrPortForwarderNotStarted) {
		t.Errorf("GetPorts error = %v, want ErrPortForwarderNotStarted", err)
	}
}

func TestPortForwarderAccessors(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := createTestClient(clientset)

	opts := PortForwardOptions{
		Namespace:  "my-namespace",
		PodName:    "my-pod",
		LocalPort:  9090,
		RemotePort: 8080,
	}

	pf, err := client.NewPortForwarder(opts)
	if err != nil {
		t.Fatalf("NewPortForwarder returned unexpected error: %v", err)
	}

	if pf.LocalPort() != 9090 {
		t.Errorf("LocalPort() = %d, want 9090", pf.LocalPort())
	}

	if pf.RemotePort() != 8080 {
		t.Errorf("RemotePort() = %d, want 8080", pf.RemotePort())
	}

	if pf.PodName() != "my-pod" {
		t.Errorf("PodName() = %q, want %q", pf.PodName(), "my-pod")
	}
}
