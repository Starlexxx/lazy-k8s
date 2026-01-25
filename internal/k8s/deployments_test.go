package k8s

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func int32Ptr(i int32) *int32 {
	return &i
}

func TestListDeployments(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-1",
				Namespace: "default",
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-2",
				Namespace: "default",
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-deployment",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test listing deployments in default namespace
	deployments, err := client.ListDeployments(ctx, "default")
	if err != nil {
		t.Fatalf("ListDeployments returned unexpected error: %v", err)
	}
	if len(deployments) != 2 {
		t.Errorf("ListDeployments returned %d deployments, want 2", len(deployments))
	}

	// Test listing deployments in other namespace
	deployments, err = client.ListDeployments(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListDeployments returned unexpected error: %v", err)
	}
	if len(deployments) != 1 {
		t.Errorf("ListDeployments returned %d deployments, want 1", len(deployments))
	}

	// Test listing deployments with empty namespace (should use client's default)
	deployments, err = client.ListDeployments(ctx, "")
	if err != nil {
		t.Fatalf("ListDeployments returned unexpected error: %v", err)
	}
	if len(deployments) != 2 {
		t.Errorf("ListDeployments with empty namespace returned %d deployments, want 2", len(deployments))
	}
}

func TestListDeploymentsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-1",
				Namespace: "default",
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-2",
				Namespace: "kube-system",
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	deployments, err := client.ListDeploymentsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListDeploymentsAllNamespaces returned unexpected error: %v", err)
	}
	if len(deployments) != 3 {
		t.Errorf("ListDeploymentsAllNamespaces returned %d deployments, want 3", len(deployments))
	}
}

func TestGetDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(3),
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test getting existing deployment
	deployment, err := client.GetDeployment(ctx, "default", "test-deployment")
	if err != nil {
		t.Fatalf("GetDeployment returned unexpected error: %v", err)
	}
	if deployment.Name != "test-deployment" {
		t.Errorf("GetDeployment returned deployment with name %q, want %q", deployment.Name, "test-deployment")
	}
	if *deployment.Spec.Replicas != 3 {
		t.Errorf("GetDeployment returned deployment with %d replicas, want 3", *deployment.Spec.Replicas)
	}

	// Test getting non-existent deployment
	_, err = client.GetDeployment(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetDeployment should have returned an error for non-existent deployment")
	}
}

func TestDeleteDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Delete the deployment
	err := client.DeleteDeployment(ctx, "default", "test-deployment")
	if err != nil {
		t.Fatalf("DeleteDeployment returned unexpected error: %v", err)
	}

	// Verify deployment is deleted
	_, err = client.GetDeployment(ctx, "default", "test-deployment")
	if err == nil {
		t.Error("Deployment should have been deleted")
	}
}

// Note: TestScaleDeployment and TestRestartDeployment are skipped because the fake clientset
// doesn't fully support GetScale/UpdateScale and Patch operations with the same behavior as
// a real Kubernetes cluster. These operations require integration tests with a real cluster.

func TestUpdateDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Get the deployment
	deployment, err := client.GetDeployment(ctx, "default", "test-deployment")
	if err != nil {
		t.Fatalf("GetDeployment returned unexpected error: %v", err)
	}

	// Update the deployment
	deployment.Spec.Replicas = int32Ptr(10)
	deployment.Labels = map[string]string{"updated": "true"}

	updated, err := client.UpdateDeployment(ctx, deployment)
	if err != nil {
		t.Fatalf("UpdateDeployment returned unexpected error: %v", err)
	}

	if *updated.Spec.Replicas != 10 {
		t.Errorf("Updated deployment has %d replicas, want 10", *updated.Spec.Replicas)
	}
	if updated.Labels["updated"] != "true" {
		t.Errorf("Updated deployment should have label 'updated=true'")
	}
}

func TestWatchDeployments(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test creating a watch
	watcher, err := client.WatchDeployments(ctx, "default")
	if err != nil {
		t.Fatalf("WatchDeployments returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	// Verify watch channel is available
	if watcher.ResultChan() == nil {
		t.Error("WatchDeployments returned watcher with nil ResultChan")
	}
}

func TestGetDeploymentReadyCount(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		expected   string
	}{
		{
			name: "all replicas ready",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 3,
				},
			},
			expected: "3/3",
		},
		{
			name: "partial replicas ready",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(5),
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 2,
				},
			},
			expected: "2/5",
		},
		{
			name: "no replicas ready",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 0,
				},
			},
			expected: "0/3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDeploymentReadyCount(tt.deployment)
			if result != tt.expected {
				t.Errorf("GetDeploymentReadyCount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetDeploymentImages(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		expected   []string
	}{
		{
			name: "single container",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "main", Image: "nginx:1.21"},
							},
						},
					},
				},
			},
			expected: []string{"nginx:1.21"},
		},
		{
			name: "multiple containers",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "app", Image: "myapp:v1.0"},
								{Name: "sidecar", Image: "istio-proxy:1.15"},
								{Name: "logger", Image: "fluentd:latest"},
							},
						},
					},
				},
			},
			expected: []string{"myapp:v1.0", "istio-proxy:1.15", "fluentd:latest"},
		},
		{
			name: "no containers",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
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
			result := GetDeploymentImages(tt.deployment)
			if len(result) != len(tt.expected) {
				t.Errorf("GetDeploymentImages() returned %d images, want %d", len(result), len(tt.expected))
				return
			}
			for i, img := range result {
				if img != tt.expected[i] {
					t.Errorf("GetDeploymentImages()[%d] = %q, want %q", i, img, tt.expected[i])
				}
			}
		})
	}
}
