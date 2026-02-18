package k8s

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListStatefulSets(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-1",
				Namespace: "default",
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-2",
				Namespace: "default",
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-statefulset",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	statefulsets, err := client.ListStatefulSets(ctx, "default")
	if err != nil {
		t.Fatalf("ListStatefulSets returned unexpected error: %v", err)
	}

	if len(statefulsets) != 2 {
		t.Errorf("ListStatefulSets returned %d statefulsets, want 2", len(statefulsets))
	}

	statefulsets, err = client.ListStatefulSets(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListStatefulSets returned unexpected error: %v", err)
	}

	if len(statefulsets) != 1 {
		t.Errorf("ListStatefulSets returned %d statefulsets, want 1", len(statefulsets))
	}

	statefulsets, err = client.ListStatefulSets(ctx, "")
	if err != nil {
		t.Fatalf("ListStatefulSets returned unexpected error: %v", err)
	}

	if len(statefulsets) != 2 {
		t.Errorf(
			"ListStatefulSets with empty namespace returned %d statefulsets, want 2",
			len(statefulsets),
		)
	}
}

func TestListStatefulSetsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-1",
				Namespace: "default",
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-2",
				Namespace: "kube-system",
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	statefulsets, err := client.ListStatefulSetsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListStatefulSetsAllNamespaces returned unexpected error: %v", err)
	}

	if len(statefulsets) != 3 {
		t.Errorf(
			"ListStatefulSetsAllNamespaces returned %d statefulsets, want 3",
			len(statefulsets),
		)
	}
}

func TestGetStatefulSet(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-statefulset",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: int32Ptr(3),
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	sts, err := client.GetStatefulSet(ctx, "default", "test-statefulset")
	if err != nil {
		t.Fatalf("GetStatefulSet returned unexpected error: %v", err)
	}

	if sts.Name != "test-statefulset" {
		t.Errorf(
			"GetStatefulSet returned statefulset with name %q, want %q",
			sts.Name,
			"test-statefulset",
		)
	}

	if *sts.Spec.Replicas != 3 {
		t.Errorf(
			"GetStatefulSet returned statefulset with %d replicas, want 3",
			*sts.Spec.Replicas,
		)
	}

	_, err = client.GetStatefulSet(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetStatefulSet should have returned an error for non-existent statefulset")
	}
}

func TestDeleteStatefulSet(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-statefulset",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	err := client.DeleteStatefulSet(ctx, "default", "test-statefulset")
	if err != nil {
		t.Fatalf("DeleteStatefulSet returned unexpected error: %v", err)
	}

	_, err = client.GetStatefulSet(ctx, "default", "test-statefulset")
	if err == nil {
		t.Error("StatefulSet should have been deleted")
	}
}

func TestWatchStatefulSets(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-statefulset",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	watcher, err := client.WatchStatefulSets(ctx, "default")
	if err != nil {
		t.Fatalf("WatchStatefulSets returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	if watcher.ResultChan() == nil {
		t.Error("WatchStatefulSets returned watcher with nil ResultChan")
	}
}

func TestGetStatefulSetReadyCount(t *testing.T) {
	tests := []struct {
		name        string
		statefulset *appsv1.StatefulSet
		expected    string
	}{
		{
			name: "all replicas ready",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32Ptr(3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 3,
				},
			},
			expected: "3/3",
		},
		{
			name: "partial replicas ready",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32Ptr(5),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 2,
				},
			},
			expected: "2/5",
		},
		{
			name: "no replicas ready",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32Ptr(3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 0,
				},
			},
			expected: "0/3",
		},
		{
			name: "nil replicas",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: nil,
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 0,
				},
			},
			expected: "0/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatefulSetReadyCount(tt.statefulset)
			if result != tt.expected {
				t.Errorf("GetStatefulSetReadyCount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetStatefulSetImages(t *testing.T) {
	tests := []struct {
		name        string
		statefulset *appsv1.StatefulSet
		expected    []string
	}{
		{
			name: "single container",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "main", Image: "postgres:14"},
							},
						},
					},
				},
			},
			expected: []string{"postgres:14"},
		},
		{
			name: "multiple containers",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "db", Image: "mysql:8"},
								{Name: "sidecar", Image: "backup:v1"},
							},
						},
					},
				},
			},
			expected: []string{"mysql:8", "backup:v1"},
		},
		{
			name: "no containers",
			statefulset: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
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
			result := GetStatefulSetImages(tt.statefulset)
			if len(result) != len(tt.expected) {
				t.Errorf(
					"GetStatefulSetImages() returned %d images, want %d",
					len(result),
					len(tt.expected),
				)

				return
			}

			for i, img := range result {
				if img != tt.expected[i] {
					t.Errorf("GetStatefulSetImages()[%d] = %q, want %q", i, img, tt.expected[i])
				}
			}
		})
	}
}
