package k8s

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetPodsLogSnapshotEmpty(t *testing.T) {
	client := createTestClient(fake.NewSimpleClientset())

	got, err := client.GetPodsLogSnapshot(context.Background(), "default", nil, LogOptions{})
	if err != nil {
		t.Fatalf("GetPodsLogSnapshot returned error: %v", err)
	}

	if got != "" {
		t.Errorf("GetPodsLogSnapshot with empty pod list = %q, want empty string", got)
	}
}

func TestGetPodsLogSnapshotPrefixesLines(t *testing.T) {
	// fake clientset returns "fake logs" for every GetLogs call; the test
	// verifies our prefixing wraps that content with the right pod names
	// rather than the content of the logs themselves.
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-a", Namespace: "default"},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-b", Namespace: "default"},
		},
	)
	client := createTestClient(clientset)

	got, err := client.GetPodsLogSnapshot(
		context.Background(),
		"default",
		[]string{"pod-a", "pod-b"},
		LogOptions{TailLines: 10},
	)
	if err != nil {
		t.Fatalf("GetPodsLogSnapshot returned error: %v", err)
	}

	if !strings.Contains(got, "[pod-a] ") {
		t.Errorf("GetPodsLogSnapshot missing pod-a prefix: %q", got)
	}

	if !strings.Contains(got, "[pod-b] ") {
		t.Errorf("GetPodsLogSnapshot missing pod-b prefix: %q", got)
	}
}
