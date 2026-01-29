package k8s

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListDaemonSets(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daemonset-1",
				Namespace: "default",
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daemonset-2",
				Namespace: "default",
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-daemonset",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test listing daemonsets in default namespace
	daemonsets, err := client.ListDaemonSets(ctx, "default")
	if err != nil {
		t.Fatalf("ListDaemonSets returned unexpected error: %v", err)
	}

	if len(daemonsets) != 2 {
		t.Errorf("ListDaemonSets returned %d daemonsets, want 2", len(daemonsets))
	}

	// Test listing daemonsets in other namespace
	daemonsets, err = client.ListDaemonSets(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListDaemonSets returned unexpected error: %v", err)
	}

	if len(daemonsets) != 1 {
		t.Errorf("ListDaemonSets returned %d daemonsets, want 1", len(daemonsets))
	}

	// Test listing daemonsets with empty namespace (should use client's default)
	daemonsets, err = client.ListDaemonSets(ctx, "")
	if err != nil {
		t.Fatalf("ListDaemonSets returned unexpected error: %v", err)
	}

	if len(daemonsets) != 2 {
		t.Errorf(
			"ListDaemonSets with empty namespace returned %d daemonsets, want 2",
			len(daemonsets),
		)
	}
}

func TestListDaemonSetsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daemonset-1",
				Namespace: "default",
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daemonset-2",
				Namespace: "kube-system",
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daemonset-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	daemonsets, err := client.ListDaemonSetsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListDaemonSetsAllNamespaces returned unexpected error: %v", err)
	}

	if len(daemonsets) != 3 {
		t.Errorf("ListDaemonSetsAllNamespaces returned %d daemonsets, want 3", len(daemonsets))
	}
}

func TestGetDaemonSet(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-daemonset",
				Namespace: "default",
			},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 3,
				NumberReady:            3,
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test getting existing daemonset
	ds, err := client.GetDaemonSet(ctx, "default", "test-daemonset")
	if err != nil {
		t.Fatalf("GetDaemonSet returned unexpected error: %v", err)
	}

	if ds.Name != "test-daemonset" {
		t.Errorf("GetDaemonSet returned daemonset with name %q, want %q", ds.Name, "test-daemonset")
	}

	if ds.Status.DesiredNumberScheduled != 3 {
		t.Errorf(
			"GetDaemonSet returned daemonset with %d desired, want 3",
			ds.Status.DesiredNumberScheduled,
		)
	}

	// Test getting non-existent daemonset
	_, err = client.GetDaemonSet(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetDaemonSet should have returned an error for non-existent daemonset")
	}
}

func TestDeleteDaemonSet(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-daemonset",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Delete the daemonset
	err := client.DeleteDaemonSet(ctx, "default", "test-daemonset")
	if err != nil {
		t.Fatalf("DeleteDaemonSet returned unexpected error: %v", err)
	}

	// Verify daemonset is deleted
	_, err = client.GetDaemonSet(ctx, "default", "test-daemonset")
	if err == nil {
		t.Error("DaemonSet should have been deleted")
	}
}

func TestWatchDaemonSets(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-daemonset",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test creating a watch
	watcher, err := client.WatchDaemonSets(ctx, "default")
	if err != nil {
		t.Fatalf("WatchDaemonSets returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	// Verify watch channel is available
	if watcher.ResultChan() == nil {
		t.Error("WatchDaemonSets returned watcher with nil ResultChan")
	}
}

func TestGetDaemonSetReadyCount(t *testing.T) {
	tests := []struct {
		name      string
		daemonset *appsv1.DaemonSet
		expected  string
	}{
		{
			name: "all nodes ready",
			daemonset: &appsv1.DaemonSet{
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: 3,
					NumberReady:            3,
				},
			},
			expected: "3/3",
		},
		{
			name: "partial nodes ready",
			daemonset: &appsv1.DaemonSet{
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            2,
				},
			},
			expected: "2/5",
		},
		{
			name: "no nodes ready",
			daemonset: &appsv1.DaemonSet{
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: 3,
					NumberReady:            0,
				},
			},
			expected: "0/3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDaemonSetReadyCount(tt.daemonset)
			if result != tt.expected {
				t.Errorf("GetDaemonSetReadyCount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetDaemonSetImages(t *testing.T) {
	tests := []struct {
		name      string
		daemonset *appsv1.DaemonSet
		expected  []string
	}{
		{
			name: "single container",
			daemonset: &appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "main", Image: "fluentd:v1.14"},
							},
						},
					},
				},
			},
			expected: []string{"fluentd:v1.14"},
		},
		{
			name: "multiple containers",
			daemonset: &appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "agent", Image: "datadog-agent:7"},
								{Name: "process", Image: "process-agent:7"},
							},
						},
					},
				},
			},
			expected: []string{"datadog-agent:7", "process-agent:7"},
		},
		{
			name: "no containers",
			daemonset: &appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{},
						},
					},
				},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDaemonSetImages(tt.daemonset)
			if len(result) != len(tt.expected) {
				t.Errorf(
					"GetDaemonSetImages() returned %d images, want %d",
					len(result),
					len(tt.expected),
				)

				return
			}

			for i, img := range result {
				if img != tt.expected[i] {
					t.Errorf("GetDaemonSetImages()[%d] = %q, want %q", i, img, tt.expected[i])
				}
			}
		})
	}
}
