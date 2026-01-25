package k8s

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestClientSetNamespace(t *testing.T) {
	client := &Client{
		namespace: "default",
	}

	// Test setting namespace
	client.SetNamespace("my-namespace")
	if client.namespace != "my-namespace" {
		t.Errorf("SetNamespace() namespace = %q, want %q", client.namespace, "my-namespace")
	}

	// Test setting to empty string
	client.SetNamespace("")
	if client.namespace != "" {
		t.Errorf("SetNamespace() namespace = %q, want empty string", client.namespace)
	}
}

func TestClientCurrentNamespace(t *testing.T) {
	client := &Client{
		namespace: "test-namespace",
	}

	result := client.CurrentNamespace()
	if result != "test-namespace" {
		t.Errorf("CurrentNamespace() = %q, want %q", result, "test-namespace")
	}
}

func TestClientCurrentContext(t *testing.T) {
	client := &Client{
		contextName: "my-context",
	}

	result := client.CurrentContext()
	if result != "my-context" {
		t.Errorf("CurrentContext() = %q, want %q", result, "my-context")
	}
}

func TestClientClientset(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
	}

	result := client.Clientset()
	if result == nil {
		t.Error("Clientset() returned nil")
	}
}

func TestClientGetContexts(t *testing.T) {
	client := &Client{
		rawConfig: api.Config{
			Contexts: map[string]*api.Context{
				"context-1": {},
				"context-2": {},
				"context-3": {},
			},
		},
	}

	contexts := client.GetContexts()
	if len(contexts) != 3 {
		t.Errorf("GetContexts() returned %d contexts, want 3", len(contexts))
	}

	// Verify all contexts are present (order may vary)
	contextMap := make(map[string]bool)
	for _, ctx := range contexts {
		contextMap[ctx] = true
	}
	for _, expected := range []string{"context-1", "context-2", "context-3"} {
		if !contextMap[expected] {
			t.Errorf("GetContexts() missing expected context %q", expected)
		}
	}
}

func TestClientGetContextsEmpty(t *testing.T) {
	client := &Client{
		rawConfig: api.Config{
			Contexts: map[string]*api.Context{},
		},
	}

	contexts := client.GetContexts()
	if len(contexts) != 0 {
		t.Errorf("GetContexts() returned %d contexts, want 0", len(contexts))
	}
}

func TestClientContext(t *testing.T) {
	client := &Client{}

	ctx := client.Context()
	if ctx == nil {
		t.Error("Context() returned nil")
	}
}

func TestClientRestConfig(t *testing.T) {
	client := &Client{
		restConfig: nil, // Can be nil in this test
	}

	// Should not panic
	result := client.RestConfig()
	if result != nil {
		t.Error("RestConfig() should be nil when not set")
	}
}
