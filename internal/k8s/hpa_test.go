package k8s

import (
	"context"
	"testing"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListHPAs(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpa-1",
				Namespace: "default",
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpa-2",
				Namespace: "default",
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-hpa",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	hpas, err := client.ListHPAs(ctx, "default")
	if err != nil {
		t.Fatalf("ListHPAs returned unexpected error: %v", err)
	}

	if len(hpas) != 2 {
		t.Errorf("ListHPAs returned %d HPAs, want 2", len(hpas))
	}

	hpas, err = client.ListHPAs(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListHPAs returned unexpected error: %v", err)
	}

	if len(hpas) != 1 {
		t.Errorf("ListHPAs returned %d HPAs, want 1", len(hpas))
	}

	hpas, err = client.ListHPAs(ctx, "")
	if err != nil {
		t.Fatalf("ListHPAs returned unexpected error: %v", err)
	}

	if len(hpas) != 2 {
		t.Errorf("ListHPAs with empty namespace returned %d HPAs, want 2", len(hpas))
	}
}

func TestListHPAsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpa-1",
				Namespace: "default",
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpa-2",
				Namespace: "kube-system",
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpa-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	hpas, err := client.ListHPAsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListHPAsAllNamespaces returned unexpected error: %v", err)
	}

	if len(hpas) != 3 {
		t.Errorf("ListHPAsAllNamespaces returned %d HPAs, want 3", len(hpas))
	}
}

func TestGetHPA(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-hpa",
				Namespace: "default",
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				MinReplicas: int32Ptr(1),
				MaxReplicas: 10,
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	hpa, err := client.GetHPA(ctx, "default", "test-hpa")
	if err != nil {
		t.Fatalf("GetHPA returned unexpected error: %v", err)
	}

	if hpa.Name != "test-hpa" {
		t.Errorf("GetHPA returned HPA with name %q, want %q", hpa.Name, "test-hpa")
	}

	if hpa.Spec.MaxReplicas != 10 {
		t.Errorf("GetHPA returned HPA with max replicas %d, want 10", hpa.Spec.MaxReplicas)
	}

	_, err = client.GetHPA(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetHPA should have returned an error for non-existent HPA")
	}
}

func TestDeleteHPA(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-hpa",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	err := client.DeleteHPA(ctx, "default", "test-hpa")
	if err != nil {
		t.Fatalf("DeleteHPA returned unexpected error: %v", err)
	}

	_, err = client.GetHPA(ctx, "default", "test-hpa")
	if err == nil {
		t.Error("HPA should have been deleted")
	}
}

func TestWatchHPAs(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-hpa",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	watcher, err := client.WatchHPAs(ctx, "default")
	if err != nil {
		t.Fatalf("WatchHPAs returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	if watcher.ResultChan() == nil {
		t.Error("WatchHPAs returned watcher with nil ResultChan")
	}
}

func TestGetHPAReplicaCount(t *testing.T) {
	tests := []struct {
		name     string
		hpa      *autoscalingv2.HorizontalPodAutoscaler
		expected string
	}{
		{
			name: "with min replicas set",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(2),
					MaxReplicas: 10,
				},
				Status: autoscalingv2.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 5,
				},
			},
			expected: "5 (2-10)",
		},
		{
			name: "nil min replicas (defaults to 1)",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					MinReplicas: nil,
					MaxReplicas: 5,
				},
				Status: autoscalingv2.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 3,
				},
			},
			expected: "3 (1-5)",
		},
		{
			name: "at minimum",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(1),
					MaxReplicas: 10,
				},
				Status: autoscalingv2.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 1,
				},
			},
			expected: "1 (1-10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHPAReplicaCount(tt.hpa)
			if result != tt.expected {
				t.Errorf("GetHPAReplicaCount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetHPATargetRef(t *testing.T) {
	tests := []struct {
		name     string
		hpa      *autoscalingv2.HorizontalPodAutoscaler
		expected string
	}{
		{
			name: "deployment target",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-deployment",
					},
				},
			},
			expected: "Deployment/my-deployment",
		},
		{
			name: "statefulset target",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind: "StatefulSet",
						Name: "my-statefulset",
					},
				},
			},
			expected: "StatefulSet/my-statefulset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHPATargetRef(tt.hpa)
			if result != tt.expected {
				t.Errorf("GetHPATargetRef() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetHPAMetricsSummary(t *testing.T) {
	tests := []struct {
		name     string
		hpa      *autoscalingv2.HorizontalPodAutoscaler
		expected string
	}{
		{
			name: "no metrics",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					Metrics: []autoscalingv2.MetricSpec{},
				},
			},
			expected: "No metrics configured",
		},
		{
			name: "one metric",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					Metrics: []autoscalingv2.MetricSpec{
						{Type: autoscalingv2.ResourceMetricSourceType},
					},
				},
			},
			expected: "1 metric(s)",
		},
		{
			name: "multiple metrics",
			hpa: &autoscalingv2.HorizontalPodAutoscaler{
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					Metrics: []autoscalingv2.MetricSpec{
						{Type: autoscalingv2.ResourceMetricSourceType},
						{Type: autoscalingv2.ExternalMetricSourceType},
						{Type: autoscalingv2.PodsMetricSourceType},
					},
				},
			},
			expected: "3 metric(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHPAMetricsSummary(tt.hpa)
			if result != tt.expected {
				t.Errorf("GetHPAMetricsSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}
