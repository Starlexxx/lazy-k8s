package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Starlexxx/lazy-k8s/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// captureOutput captures stdout for testing functions that write to stdout
func captureOutput(f func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	execErr := f()

	if err := w.Close(); err != nil {
		return "", err
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), execErr
}

// setupTestClient creates a test client with fake data
func setupTestClient() *client.K8sClient {
	// Create a fake client
	fakeClientset := fake.NewSimpleClientset()

	// Create a test Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			Labels: map[string]string{
				"app": "test",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
	}
	_, err := fakeClientset.CoreV1().Pods("default").Create(
		context.TODO(),
		pod,
		metav1.CreateOptions{},
	)
	if err != nil {
		// In tests, we can just panic since this should never fail with fake clientset
		panic("Failed to create test pod: " + err.Error())
	}

	// Create a test Node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-node",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
			Labels: map[string]string{
				"node-role.kubernetes.io/master": "",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: "v1.25.0",
			},
		},
	}
	_, err = fakeClientset.CoreV1().Nodes().Create(
		context.TODO(),
		node,
		metav1.CreateOptions{},
	)
	if err != nil {
		// In tests, we can just panic since this should never fail with fake clientset
		panic("Failed to create test node: " + err.Error())
	}

	return &client.K8sClient{
		ClientSet:      fakeClientset,
		CurrentContext: "test-context",
	}
}

func TestGet(t *testing.T) {
	testClient := setupTestClient()

	tests := []struct {
		name      string
		resource  string
		namespace string
		contains  []string
	}{
		{
			name:      "get pods",
			resource:  "pods",
			namespace: "default",
			contains:  []string{"NAMESPACE", "NAME", "STATUS", "test-pod", "Running"},
		},
		{
			name:      "get nodes",
			resource:  "nodes",
			namespace: "default",
			contains:  []string{"NAME", "STATUS", "ROLES", "test-node", "Ready", "master"},
		},
		{
			name:      "unknown resource",
			resource:  "unknown",
			namespace: "default",
			contains:  []string{"Unknown resource type"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := captureOutput(func() error {
				Get(testClient, tc.resource, tc.namespace)
				return nil
			})
			if err != nil {
				t.Fatalf("Failed to capture output: %v", err)
			}

			for _, s := range tc.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Expected output to contain '%s', but output was:\n%s", s, output)
				}
			}
		})
	}
}

func TestDisplayPodDetails(t *testing.T) {
	// Create a test Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			Labels: map[string]string{
				"app": "test",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
	}

	tests := []struct {
		name       string
		showLabels bool
		contains   []string
	}{
		{
			name:       "basic information without labels",
			showLabels: false,
			contains:   []string{"Name: test-pod", "Namespace: default", "Status: Running", "Containers:", "test-container", "nginx:latest"},
		},
		{
			name:       "information with labels",
			showLabels: true,
			contains:   []string{"Name: test-pod", "Labels:", "app: test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := captureOutput(func() error {
				// displayPodDetails returns an error that we need to return
				return displayPodDetails(pod, tc.showLabels)
			})
			if err != nil {
				t.Fatalf("Failed to display pod details: %v", err)
			}

			for _, s := range tc.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Expected output to contain '%s', but output was:\n%s", s, output)
				}
			}
		})
	}
}

func TestDisplayNodeDetails(t *testing.T) {
	// Create a test Node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-node",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
			Labels: map[string]string{
				"node-role.kubernetes.io/master": "",
				"kubernetes.io/hostname":         "test-node",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion:          "v1.25.0",
				OSImage:                 "Linux",
				Architecture:            "amd64",
				ContainerRuntimeVersion: "containerd://1.6.0",
			},
		},
	}

	tests := []struct {
		name       string
		showLabels bool
		contains   []string
	}{
		{
			name:       "basic information without labels",
			showLabels: false,
			contains:   []string{"Name: test-node", "Status: Ready", "Roles: master", "Kubernetes Version: v1.25.0"},
		},
		{
			name:       "information with labels",
			showLabels: true,
			contains:   []string{"Name: test-node", "Labels:", "node-role.kubernetes.io/master", "kubernetes.io/hostname"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := captureOutput(func() error {
				// displayNodeDetails returns an error that we need to return
				return displayNodeDetails(node, tc.showLabels)
			})
			if err != nil {
				t.Fatalf("Failed to display node details: %v", err)
			}

			for _, s := range tc.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Expected output to contain '%s', but output was:\n%s", s, output)
				}
			}
		})
	}
}

func TestNewGetCommand(t *testing.T) {
	testClient := setupTestClient()
	cmd := NewGetCommand(testClient)

	if cmd.Use != "get" {
		t.Errorf("Incorrect command usage: %s, expected: get", cmd.Use)
	}

	if len(cmd.Commands()) != 4 {
		t.Errorf("Incorrect number of subcommands: %d, expected: 4", len(cmd.Commands()))
	}

	// Check for all subcommands
	subcommands := []string{"pods", "nodes", "services", "deployments"}
	for _, sc := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Use == sc || strings.HasPrefix(c.Use, sc+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Subcommand '%s' not found", sc)
		}
	}
}
