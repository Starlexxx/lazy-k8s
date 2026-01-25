package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func createTestClient(clientset *fake.Clientset) *Client {
	return &Client{
		clientset: clientset,
		namespace: "default",
	}
}

// Note: The Client struct uses kubernetes.Interface which allows using fake.Clientset for testing

func TestListPods(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-ns-pod",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test listing pods in default namespace
	pods, err := client.ListPods(ctx, "default")
	if err != nil {
		t.Fatalf("ListPods returned unexpected error: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("ListPods returned %d pods, want 2", len(pods))
	}

	// Test listing pods in other namespace
	pods, err = client.ListPods(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListPods returned unexpected error: %v", err)
	}

	if len(pods) != 1 {
		t.Errorf("ListPods returned %d pods, want 1", len(pods))
	}

	// Test listing pods with empty namespace (should use client's default)
	pods, err = client.ListPods(ctx, "")
	if err != nil {
		t.Fatalf("ListPods returned unexpected error: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("ListPods with empty namespace returned %d pods, want 2", len(pods))
	}
}

func TestListPodsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: "default",
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: "kube-system",
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	pods, err := client.ListPodsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListPodsAllNamespaces returned unexpected error: %v", err)
	}

	if len(pods) != 3 {
		t.Errorf("ListPodsAllNamespaces returned %d pods, want 3", len(pods))
	}
}

func TestGetPod(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "main", Image: "nginx:latest"},
				},
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test getting existing pod
	pod, err := client.GetPod(ctx, "default", "test-pod")
	if err != nil {
		t.Fatalf("GetPod returned unexpected error: %v", err)
	}

	if pod.Name != "test-pod" {
		t.Errorf("GetPod returned pod with name %q, want %q", pod.Name, "test-pod")
	}

	if len(pod.Spec.Containers) != 1 {
		t.Errorf("GetPod returned pod with %d containers, want 1", len(pod.Spec.Containers))
	}

	// Test getting non-existent pod
	_, err = client.GetPod(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetPod should have returned an error for non-existent pod")
	}

	// Test getting pod with empty namespace (should use client's default)
	pod, err = client.GetPod(ctx, "", "test-pod")
	if err != nil {
		t.Fatalf("GetPod with empty namespace returned unexpected error: %v", err)
	}

	if pod.Name != "test-pod" {
		t.Errorf("GetPod returned pod with name %q, want %q", pod.Name, "test-pod")
	}
}

func TestDeletePod(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Delete the pod
	err := client.DeletePod(ctx, "default", "test-pod")
	if err != nil {
		t.Fatalf("DeletePod returned unexpected error: %v", err)
	}

	// Verify pod is deleted
	_, err = client.GetPod(ctx, "default", "test-pod")
	if err == nil {
		t.Error("Pod should have been deleted")
	}
}

func TestGetPodStatus(t *testing.T) {
	now := metav1.Now()

	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "running pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "terminating pod",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &now,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			expected: "Terminating",
		},
		{
			name: "pending pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			expected: "Pending",
		},
		{
			name: "container waiting - ImagePullBackOff",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ImagePullBackOff",
								},
							},
						},
					},
				},
			},
			expected: "ImagePullBackOff",
		},
		{
			name: "container waiting - CrashLoopBackOff",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					},
				},
			},
			expected: "CrashLoopBackOff",
		},
		{
			name: "container terminated - Error",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									Reason: "Error",
								},
							},
						},
					},
				},
			},
			expected: "Error",
		},
		{
			name: "succeeded pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
			expected: "Succeeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPodStatus(tt.pod)
			if result != tt.expected {
				t.Errorf("GetPodStatus() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetPodReadyCount(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "all containers ready",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container-1"},
						{Name: "container-2"},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: true},
					},
				},
			},
			expected: "2/2",
		},
		{
			name: "partial containers ready",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container-1"},
						{Name: "container-2"},
						{Name: "container-3"},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: false},
						{Ready: true},
					},
				},
			},
			expected: "2/3",
		},
		{
			name: "no containers ready",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container-1"},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: false},
					},
				},
			},
			expected: "0/1",
		},
		{
			name: "no container statuses yet",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container-1"},
						{Name: "container-2"},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			expected: "0/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPodReadyCount(tt.pod)
			if result != tt.expected {
				t.Errorf("GetPodReadyCount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetPodRestarts(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected int32
	}{
		{
			name: "no restarts",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{RestartCount: 0},
						{RestartCount: 0},
					},
				},
			},
			expected: 0,
		},
		{
			name: "some restarts",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{RestartCount: 3},
						{RestartCount: 2},
					},
				},
			},
			expected: 5,
		},
		{
			name: "single container restarts",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{RestartCount: 10},
					},
				},
			},
			expected: 10,
		},
		{
			name: "no container statuses",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPodRestarts(tt.pod)
			if result != tt.expected {
				t.Errorf("GetPodRestarts() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestWatchPods(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test creating a watch
	watcher, err := client.WatchPods(ctx, "default")
	if err != nil {
		t.Fatalf("WatchPods returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	// Verify watch channel is available
	if watcher.ResultChan() == nil {
		t.Error("WatchPods returned watcher with nil ResultChan")
	}
}
