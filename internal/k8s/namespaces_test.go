package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-app",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListNamespaces returned unexpected error: %v", err)
	}

	if len(namespaces) != 3 {
		t.Errorf("ListNamespaces returned %d namespaces, want 3", len(namespaces))
	}

	foundNames := make(map[string]bool)
	for _, ns := range namespaces {
		foundNames[ns.Name] = true
	}

	for _, expected := range []string{"default", "kube-system", "my-app"} {
		if !foundNames[expected] {
			t.Errorf("Expected namespace %q not found", expected)
		}
	}
}

func TestGetNamespace(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
				Labels: map[string]string{
					"name": "default",
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	ns, err := client.GetNamespace(ctx, "default")
	if err != nil {
		t.Fatalf("GetNamespace returned unexpected error: %v", err)
	}

	if ns.Name != "default" {
		t.Errorf("GetNamespace returned namespace with name %q, want %q", ns.Name, "default")
	}

	if ns.Labels["name"] != "default" {
		t.Errorf("GetNamespace returned namespace with wrong label")
	}

	_, err = client.GetNamespace(ctx, "non-existent")
	if err == nil {
		t.Error("GetNamespace should have returned an error for non-existent namespace")
	}
}

func TestCreateNamespace(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	client := createTestClient(clientset)
	ctx := context.Background()

	ns, err := client.CreateNamespace(ctx, "new-namespace")
	if err != nil {
		t.Fatalf("CreateNamespace returned unexpected error: %v", err)
	}

	if ns.Name != "new-namespace" {
		t.Errorf(
			"CreateNamespace returned namespace with name %q, want %q",
			ns.Name,
			"new-namespace",
		)
	}

	created, err := client.GetNamespace(ctx, "new-namespace")
	if err != nil {
		t.Fatalf("GetNamespace returned unexpected error: %v", err)
	}

	if created.Name != "new-namespace" {
		t.Errorf("Created namespace has name %q, want %q", created.Name, "new-namespace")
	}
}

func TestDeleteNamespace(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "to-delete",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	_, err := client.GetNamespace(ctx, "to-delete")
	if err != nil {
		t.Fatalf("GetNamespace returned unexpected error: %v", err)
	}

	err = client.DeleteNamespace(ctx, "to-delete")
	if err != nil {
		t.Fatalf("DeleteNamespace returned unexpected error: %v", err)
	}

	_, err = client.GetNamespace(ctx, "to-delete")
	if err == nil {
		t.Error("Namespace should have been deleted")
	}
}

func TestWatchNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	watcher, err := client.WatchNamespaces(ctx)
	if err != nil {
		t.Fatalf("WatchNamespaces returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	if watcher.ResultChan() == nil {
		t.Error("WatchNamespaces returned watcher with nil ResultChan")
	}
}
