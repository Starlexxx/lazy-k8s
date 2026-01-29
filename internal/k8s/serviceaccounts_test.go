package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListServiceAccounts(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-1",
				Namespace: "default",
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-2",
				Namespace: "default",
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-sa",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test listing service accounts in default namespace
	sas, err := client.ListServiceAccounts(ctx, "default")
	if err != nil {
		t.Fatalf("ListServiceAccounts returned unexpected error: %v", err)
	}

	if len(sas) != 2 {
		t.Errorf("ListServiceAccounts returned %d service accounts, want 2", len(sas))
	}

	// Test listing service accounts in other namespace
	sas, err = client.ListServiceAccounts(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListServiceAccounts returned unexpected error: %v", err)
	}

	if len(sas) != 1 {
		t.Errorf("ListServiceAccounts returned %d service accounts, want 1", len(sas))
	}

	// Test listing service accounts with empty namespace
	sas, err = client.ListServiceAccounts(ctx, "")
	if err != nil {
		t.Fatalf("ListServiceAccounts returned unexpected error: %v", err)
	}

	if len(sas) != 2 {
		t.Errorf(
			"ListServiceAccounts with empty namespace returned %d service accounts, want 2",
			len(sas),
		)
	}
}

func TestListServiceAccountsAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-1",
				Namespace: "default",
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-2",
				Namespace: "kube-system",
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	sas, err := client.ListServiceAccountsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListServiceAccountsAllNamespaces returned unexpected error: %v", err)
	}

	if len(sas) != 3 {
		t.Errorf("ListServiceAccountsAllNamespaces returned %d service accounts, want 3", len(sas))
	}
}

func TestGetServiceAccount(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "default",
			},
			Secrets: []corev1.ObjectReference{
				{Name: "test-sa-token"},
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test getting existing service account
	sa, err := client.GetServiceAccount(ctx, "default", "test-sa")
	if err != nil {
		t.Fatalf("GetServiceAccount returned unexpected error: %v", err)
	}

	if sa.Name != "test-sa" {
		t.Errorf(
			"GetServiceAccount returned service account with name %q, want %q",
			sa.Name,
			"test-sa",
		)
	}

	if len(sa.Secrets) != 1 {
		t.Errorf(
			"GetServiceAccount returned service account with %d secrets, want 1",
			len(sa.Secrets),
		)
	}

	// Test getting non-existent service account
	_, err = client.GetServiceAccount(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetServiceAccount should have returned an error for non-existent service account")
	}
}

func TestDeleteServiceAccount(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Delete the service account
	err := client.DeleteServiceAccount(ctx, "default", "test-sa")
	if err != nil {
		t.Fatalf("DeleteServiceAccount returned unexpected error: %v", err)
	}

	// Verify service account is deleted
	_, err = client.GetServiceAccount(ctx, "default", "test-sa")
	if err == nil {
		t.Error("ServiceAccount should have been deleted")
	}
}

func TestWatchServiceAccounts(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	// Test creating a watch
	watcher, err := client.WatchServiceAccounts(ctx, "default")
	if err != nil {
		t.Fatalf("WatchServiceAccounts returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	// Verify watch channel is available
	if watcher.ResultChan() == nil {
		t.Error("WatchServiceAccounts returned watcher with nil ResultChan")
	}
}

func TestGetServiceAccountSecretCount(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount *corev1.ServiceAccount
		expected       int
	}{
		{
			name: "no secrets",
			serviceAccount: &corev1.ServiceAccount{
				Secrets: []corev1.ObjectReference{},
			},
			expected: 0,
		},
		{
			name: "some secrets",
			serviceAccount: &corev1.ServiceAccount{
				Secrets: []corev1.ObjectReference{
					{Name: "secret-1"},
					{Name: "secret-2"},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetServiceAccountSecretCount(tt.serviceAccount)
			if result != tt.expected {
				t.Errorf("GetServiceAccountSecretCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetServiceAccountImagePullSecretCount(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount *corev1.ServiceAccount
		expected       int
	}{
		{
			name: "no image pull secrets",
			serviceAccount: &corev1.ServiceAccount{
				ImagePullSecrets: []corev1.LocalObjectReference{},
			},
			expected: 0,
		},
		{
			name: "some image pull secrets",
			serviceAccount: &corev1.ServiceAccount{
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "registry-secret-1"},
					{Name: "registry-secret-2"},
					{Name: "registry-secret-3"},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetServiceAccountImagePullSecretCount(tt.serviceAccount)
			if result != tt.expected {
				t.Errorf(
					"GetServiceAccountImagePullSecretCount() = %d, want %d",
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestGetServiceAccountAutoMount(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount *corev1.ServiceAccount
		expected       string
	}{
		{
			name: "nil automount (default)",
			serviceAccount: &corev1.ServiceAccount{
				AutomountServiceAccountToken: nil,
			},
			expected: "default",
		},
		{
			name: "automount enabled",
			serviceAccount: &corev1.ServiceAccount{
				AutomountServiceAccountToken: boolPtr(true),
			},
			expected: "true",
		},
		{
			name: "automount disabled",
			serviceAccount: &corev1.ServiceAccount{
				AutomountServiceAccountToken: boolPtr(false),
			},
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetServiceAccountAutoMount(tt.serviceAccount)
			if result != tt.expected {
				t.Errorf("GetServiceAccountAutoMount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetServiceAccountSecretsSummary(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount *corev1.ServiceAccount
		expected       string
	}{
		{
			name: "no secrets",
			serviceAccount: &corev1.ServiceAccount{
				Secrets:          []corev1.ObjectReference{},
				ImagePullSecrets: []corev1.LocalObjectReference{},
			},
			expected: "Secrets: 0, ImagePullSecrets: 0",
		},
		{
			name: "with secrets",
			serviceAccount: &corev1.ServiceAccount{
				Secrets: []corev1.ObjectReference{
					{Name: "secret-1"},
					{Name: "secret-2"},
				},
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "registry-secret"},
				},
			},
			expected: "Secrets: 2, ImagePullSecrets: 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetServiceAccountSecretsSummary(tt.serviceAccount)
			if result != tt.expected {
				t.Errorf("GetServiceAccountSecretsSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}
