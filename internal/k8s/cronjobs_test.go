package k8s

import (
	"context"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestListCronJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cronjob-1",
				Namespace: "default",
			},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cronjob-2",
				Namespace: "default",
			},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-cronjob",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test listing cronjobs in default namespace
	cronjobs, err := client.ListCronJobs(ctx, "default")
	if err != nil {
		t.Fatalf("ListCronJobs returned unexpected error: %v", err)
	}

	if len(cronjobs) != 2 {
		t.Errorf("ListCronJobs returned %d cronjobs, want 2", len(cronjobs))
	}

	// Test listing cronjobs in other namespace
	cronjobs, err = client.ListCronJobs(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListCronJobs returned unexpected error: %v", err)
	}

	if len(cronjobs) != 1 {
		t.Errorf("ListCronJobs returned %d cronjobs, want 1", len(cronjobs))
	}

	// Test listing cronjobs with empty namespace (should use client's default)
	cronjobs, err = client.ListCronJobs(ctx, "")
	if err != nil {
		t.Fatalf("ListCronJobs returned unexpected error: %v", err)
	}

	if len(cronjobs) != 2 {
		t.Errorf("ListCronJobs with empty namespace returned %d cronjobs, want 2", len(cronjobs))
	}
}

func TestListCronJobsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cronjob-1",
				Namespace: "default",
			},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cronjob-2",
				Namespace: "kube-system",
			},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cronjob-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	cronjobs, err := client.ListCronJobsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListCronJobsAllNamespaces returned unexpected error: %v", err)
	}

	if len(cronjobs) != 3 {
		t.Errorf("ListCronJobsAllNamespaces returned %d cronjobs, want 3", len(cronjobs))
	}
}

func TestGetCronJob(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cronjob",
				Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "0 * * * *",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test getting existing cronjob
	cj, err := client.GetCronJob(ctx, "default", "test-cronjob")
	if err != nil {
		t.Fatalf("GetCronJob returned unexpected error: %v", err)
	}

	if cj.Name != "test-cronjob" {
		t.Errorf("GetCronJob returned cronjob with name %q, want %q", cj.Name, "test-cronjob")
	}

	if cj.Spec.Schedule != "0 * * * *" {
		t.Errorf(
			"GetCronJob returned cronjob with schedule %q, want %q",
			cj.Spec.Schedule,
			"0 * * * *",
		)
	}

	// Test getting non-existent cronjob
	_, err = client.GetCronJob(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetCronJob should have returned an error for non-existent cronjob")
	}
}

func TestDeleteCronJob(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cronjob",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Delete the cronjob
	err := client.DeleteCronJob(ctx, "default", "test-cronjob")
	if err != nil {
		t.Fatalf("DeleteCronJob returned unexpected error: %v", err)
	}

	// Verify cronjob is deleted
	_, err = client.GetCronJob(ctx, "default", "test-cronjob")
	if err == nil {
		t.Error("CronJob should have been deleted")
	}
}

func TestWatchCronJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cronjob",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test creating a watch
	watcher, err := client.WatchCronJobs(ctx, "default")
	if err != nil {
		t.Fatalf("WatchCronJobs returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	// Verify watch channel is available
	if watcher.ResultChan() == nil {
		t.Error("WatchCronJobs returned watcher with nil ResultChan")
	}
}

func TestTriggerCronJob(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cronjob",
				Namespace: "default",
				UID:       "test-uid",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "0 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Name: "main", Image: "busybox"},
								},
								RestartPolicy: corev1.RestartPolicyNever,
							},
						},
					},
				},
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Trigger the cronjob
	job, err := client.TriggerCronJob(ctx, "default", "test-cronjob")
	if err != nil {
		t.Fatalf("TriggerCronJob returned unexpected error: %v", err)
	}

	// Verify job was created
	if job == nil {
		t.Fatal("TriggerCronJob should have returned a job")
	}

	// Verify job has the manual annotation
	if job.Annotations["cronjob.kubernetes.io/instantiate"] != "manual" {
		t.Error("Created job should have manual instantiate annotation")
	}
}

func TestGetCronJobStatus(t *testing.T) {
	tests := []struct {
		name     string
		cronjob  *batchv1.CronJob
		expected string
	}{
		{
			name: "active cronjob",
			cronjob: &batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					Suspend: boolPtr(false),
				},
			},
			expected: "Active",
		},
		{
			name: "suspended cronjob",
			cronjob: &batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					Suspend: boolPtr(true),
				},
			},
			expected: "Suspended",
		},
		{
			name: "nil suspend field",
			cronjob: &batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					Suspend: nil,
				},
			},
			expected: "Active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCronJobStatus(tt.cronjob)
			if result != tt.expected {
				t.Errorf("GetCronJobStatus() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetCronJobLastSchedule(t *testing.T) {
	now := metav1.NewTime(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))

	tests := []struct {
		name     string
		cronjob  *batchv1.CronJob
		expected string
	}{
		{
			name: "has last schedule time",
			cronjob: &batchv1.CronJob{
				Status: batchv1.CronJobStatus{
					LastScheduleTime: &now,
				},
			},
			expected: "2024-01-15 10:30:00",
		},
		{
			name: "no last schedule time",
			cronjob: &batchv1.CronJob{
				Status: batchv1.CronJobStatus{
					LastScheduleTime: nil,
				},
			},
			expected: "Never",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCronJobLastSchedule(tt.cronjob)
			if result != tt.expected {
				t.Errorf("GetCronJobLastSchedule() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetCronJobActiveJobs(t *testing.T) {
	tests := []struct {
		name     string
		cronjob  *batchv1.CronJob
		expected int
	}{
		{
			name: "no active jobs",
			cronjob: &batchv1.CronJob{
				Status: batchv1.CronJobStatus{
					Active: []corev1.ObjectReference{},
				},
			},
			expected: 0,
		},
		{
			name: "some active jobs",
			cronjob: &batchv1.CronJob{
				Status: batchv1.CronJobStatus{
					Active: []corev1.ObjectReference{
						{Name: "job-1"},
						{Name: "job-2"},
						{Name: "job-3"},
					},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCronJobActiveJobs(tt.cronjob)
			if result != tt.expected {
				t.Errorf("GetCronJobActiveJobs() = %d, want %d", result, tt.expected)
			}
		})
	}
}
