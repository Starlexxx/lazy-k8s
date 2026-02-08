//go:build integration

// Integration tests validate k8s client operations against a real cluster.
// These tests are excluded from normal test runs to avoid CI failures
// when no cluster is available.
//
// Prerequisites:
//
//	kind create cluster --name lazy-k8s-test
//	kubectl create namespace test-ns
//	kubectl create deployment nginx-test --image=nginx:alpine
//	kubectl create deployment app-test --image=busybox --replicas=2 -n test-ns -- sleep 3600
package k8s

import (
	"context"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// getTestClient creates a client for integration tests.
// Uses K8S_TEST_CONTEXT env var to allow testing against different clusters.
func getTestClient(t *testing.T) *Client {
	t.Helper()

	contextName := os.Getenv("K8S_TEST_CONTEXT")
	if contextName == "" {
		contextName = "kind-lazy-k8s-test"
	}

	client, err := NewClient("", contextName)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func TestIntegration_ClientConnection(t *testing.T) {
	client := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Clientset().Discovery().ServerVersion()
	if err != nil {
		t.Fatalf("Failed to get server version: %v", err)
	}

	if client.CurrentContext() != "kind-lazy-k8s-test" {
		t.Logf("Context: %s", client.CurrentContext())
	}

	if client.CurrentNamespace() != "default" {
		t.Errorf("Expected default namespace, got %s", client.CurrentNamespace())
	}

	_ = ctx
}

func TestIntegration_ListNamespaces(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("Failed to list namespaces: %v", err)
	}

	// Kind cluster creates default, kube-system, kube-public, kube-node-lease,
	// local-path-storage, plus our test-ns
	if len(namespaces) < 4 {
		t.Errorf("Expected at least 4 namespaces, got %d", len(namespaces))
	}

	found := false
	for _, ns := range namespaces {
		if ns.Name == "test-ns" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'test-ns' namespace")
	}
}

func TestIntegration_ListPods(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	// Deployments need time to create pods after cluster setup
	time.Sleep(2 * time.Second)

	pods, err := client.ListPods(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	t.Logf("Found %d pods in default namespace", len(pods))

	pods, err = client.ListPods(ctx, "test-ns")
	if err != nil {
		t.Fatalf("Failed to list pods in test-ns: %v", err)
	}

	t.Logf("Found %d pods in test-ns namespace", len(pods))
}

func TestIntegration_ListPodsAllNamespaces(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	pods, err := client.ListPodsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("Failed to list all pods: %v", err)
	}

	if len(pods) == 0 {
		t.Error("Expected at least some pods across all namespaces")
	}

	t.Logf("Found %d pods across all namespaces", len(pods))

	namespaces := make(map[string]int)
	for _, pod := range pods {
		namespaces[pod.Namespace]++
	}

	t.Logf("Pods by namespace: %v", namespaces)
}

func TestIntegration_ListDeployments(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	deployments, err := client.ListDeployments(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to list deployments: %v", err)
	}

	found := false
	for _, d := range deployments {
		if d.Name == "nginx-test" {
			found = true
			t.Logf("Found deployment: %s (replicas: %d/%d)",
				d.Name, d.Status.ReadyReplicas, *d.Spec.Replicas)
		}
	}
	if !found {
		t.Error("Expected to find 'nginx-test' deployment")
	}
}

func TestIntegration_ListDeploymentsAllNamespaces(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	deployments, err := client.ListDeploymentsAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("Failed to list all deployments: %v", err)
	}

	t.Logf("Found %d deployments across all namespaces", len(deployments))

	names := make([]string, 0)
	for _, d := range deployments {
		names = append(names, d.Namespace+"/"+d.Name)
	}
	t.Logf("Deployments: %v", names)
}

func TestIntegration_GetDeployment(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	deployment, err := client.GetDeployment(ctx, "default", "nginx-test")
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	if deployment.Name != "nginx-test" {
		t.Errorf("Expected deployment name 'nginx-test', got %s", deployment.Name)
	}

	t.Logf("Deployment: %s, Image: %s, Replicas: %d/%d",
		deployment.Name,
		deployment.Spec.Template.Spec.Containers[0].Image,
		deployment.Status.ReadyReplicas,
		*deployment.Spec.Replicas)
}

func TestIntegration_GetPod(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	pods, err := client.ListPods(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		t.Skip("No pods available for testing")
	}

	podName := pods[0].Name
	pod, err := client.GetPod(ctx, "default", podName)
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}

	if pod.Name != podName {
		t.Errorf("Expected pod name %s, got %s", podName, pod.Name)
	}

	t.Logf("Pod: %s, Status: %s, IP: %s",
		pod.Name, pod.Status.Phase, pod.Status.PodIP)
}

func TestIntegration_WatchPods(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	watcher, err := client.WatchPods(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to watch pods: %v", err)
	}
	defer watcher.Stop()

	select {
	case event := <-watcher.ResultChan():
		if event.Object == nil {
			t.Error("Expected event object, got nil")
		} else {
			pod := event.Object.(*corev1.Pod)
			t.Logf("Watch event: %s - %s", event.Type, pod.Name)
		}
	case <-time.After(5 * time.Second):
		// No events is acceptable for a static cluster
		t.Log("No watch events received within timeout (this may be normal)")
	}
}

func TestIntegration_WatchDeployments(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	watcher, err := client.WatchDeployments(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to watch deployments: %v", err)
	}
	defer watcher.Stop()

	if watcher.ResultChan() == nil {
		t.Error("Expected result channel, got nil")
	}
}

func TestIntegration_CreateAndDeleteNamespace(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	// Timestamp ensures unique namespace name across parallel test runs
	testNs := "integration-test-ns-" + time.Now().Format("20060102150405")

	ns, err := client.CreateNamespace(ctx, testNs)
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	if ns.Name != testNs {
		t.Errorf("Expected namespace name %s, got %s", testNs, ns.Name)
	}

	t.Logf("Created namespace: %s", ns.Name)

	_, err = client.GetNamespace(ctx, testNs)
	if err != nil {
		t.Errorf("Failed to get created namespace: %v", err)
	}

	err = client.DeleteNamespace(ctx, testNs)
	if err != nil {
		t.Errorf("Failed to delete namespace: %v", err)
	}

	t.Logf("Deleted namespace: %s", testNs)
}

func TestIntegration_PodHelpers(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	pods, err := client.ListPods(ctx, "kube-system")
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		t.Skip("No pods in kube-system")
	}

	// Limit output to avoid test log noise
	for _, pod := range pods[:min(3, len(pods))] {
		status := GetPodStatus(&pod)
		ready := GetPodReadyCount(&pod)
		restarts := GetPodRestarts(&pod)

		t.Logf("Pod: %s, Status: %s, Ready: %s, Restarts: %d",
			pod.Name, status, ready, restarts)
	}
}

func TestIntegration_DeploymentHelpers(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	deployment, err := client.GetDeployment(ctx, "default", "nginx-test")
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	readyCount := GetDeploymentReadyCount(deployment)
	images := GetDeploymentImages(deployment)

	t.Logf("Deployment: %s", deployment.Name)
	t.Logf("  Ready: %s", readyCount)
	t.Logf("  Images: %v", images)

	if len(images) == 0 {
		t.Error("Expected at least one image")
	}
}

func TestIntegration_SwitchNamespace(t *testing.T) {
	client := getTestClient(t)

	if client.CurrentNamespace() != "default" {
		t.Errorf("Expected default namespace, got %s", client.CurrentNamespace())
	}

	client.SetNamespace("test-ns")
	if client.CurrentNamespace() != "test-ns" {
		t.Errorf("Expected test-ns namespace, got %s", client.CurrentNamespace())
	}

	ctx := context.Background()
	pods, err := client.ListPods(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	for _, pod := range pods {
		if pod.Namespace != "test-ns" {
			t.Errorf("Expected pod in test-ns, got %s", pod.Namespace)
		}
	}
}

func TestIntegration_GetContexts(t *testing.T) {
	client := getTestClient(t)

	contexts := client.GetContexts()
	if len(contexts) == 0 {
		t.Error("Expected at least one context")
	}

	t.Logf("Available contexts: %v", contexts)

	found := false
	for _, ctx := range contexts {
		if ctx == "kind-lazy-k8s-test" {
			found = true
			break
		}
	}
	if !found {
		// Different kubeconfig may not have the kind context
		t.Log("kind-lazy-k8s-test context not in list (may be using different kubeconfig)")
	}
}

func TestIntegration_ScaleDeployment(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	deployment, err := client.GetDeployment(ctx, "default", "nginx-test")
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}
	originalReplicas := *deployment.Spec.Replicas
	t.Logf("Original replicas: %d", originalReplicas)

	newReplicas := originalReplicas + 1
	err = client.ScaleDeployment(ctx, "default", "nginx-test", newReplicas)
	if err != nil {
		t.Fatalf("Failed to scale deployment: %v", err)
	}

	deployment, err = client.GetDeployment(ctx, "default", "nginx-test")
	if err != nil {
		t.Fatalf("Failed to get deployment after scale: %v", err)
	}
	if *deployment.Spec.Replicas != newReplicas {
		t.Errorf("Expected %d replicas, got %d", newReplicas, *deployment.Spec.Replicas)
	}
	t.Logf("Scaled to: %d replicas", *deployment.Spec.Replicas)

	// Restore original state to avoid affecting other tests
	err = client.ScaleDeployment(ctx, "default", "nginx-test", originalReplicas)
	if err != nil {
		t.Fatalf("Failed to scale deployment back: %v", err)
	}
	t.Logf("Scaled back to: %d replicas", originalReplicas)
}
