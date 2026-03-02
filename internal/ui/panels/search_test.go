package panels

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentsPanelSearchItems(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	replicas := int32(3)
	panel.deployments = []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx-web", Namespace: "default",
			},
			Spec:   appsv1.DeploymentSpec{Replicas: &replicas},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 3, Replicas: 3},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis-cache", Namespace: "default",
			},
			Spec:   appsv1.DeploymentSpec{Replicas: &replicas},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 3, Replicas: 3},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx-api", Namespace: "staging",
			},
			Spec:   appsv1.DeploymentSpec{Replicas: &replicas},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 2, Replicas: 3},
		},
	}

	// Empty query returns nil
	results := panel.SearchItems("")
	if results != nil {
		t.Error("expected nil for empty query")
	}

	// Matching query
	results = panel.SearchItems("nginx")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'nginx', got %d", len(results))
	}

	if results[0].Name != "nginx-web" {
		t.Errorf("expected first result 'nginx-web', got %q", results[0].Name)
	}

	if results[0].Kind != "Deployments" {
		t.Errorf("expected kind 'Deployments', got %q", results[0].Kind)
	}

	if results[1].Namespace != "staging" {
		t.Errorf(
			"expected second result namespace 'staging', got %q",
			results[1].Namespace,
		)
	}

	// Case-insensitive matching
	results = panel.SearchItems("REDIS")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'REDIS', got %d", len(results))
	}

	// No match
	results = panel.SearchItems("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestDeploymentsPanelNavigateTo(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	replicas := int32(1)
	panel.deployments = []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "app-a", Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{Replicas: &replicas},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "app-b", Namespace: "staging",
			},
			Spec: appsv1.DeploymentSpec{Replicas: &replicas},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "app-c", Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{Replicas: &replicas},
		},
	}

	// Apply filter so filtered = full list
	panel.SetFilter("")

	// Navigate to second item
	found := panel.NavigateTo("app-b", "staging")
	if !found {
		t.Error("expected NavigateTo to find app-b")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}

	// Navigate to third item
	found = panel.NavigateTo("app-c", "default")
	if !found {
		t.Error("expected NavigateTo to find app-c")
	}

	if panel.Cursor() != 2 {
		t.Errorf("expected cursor 2, got %d", panel.Cursor())
	}

	// Non-existent item
	found = panel.NavigateTo("app-x", "default")
	if found {
		t.Error("expected NavigateTo to return false for non-existent item")
	}

	// Wrong namespace
	found = panel.NavigateTo("app-a", "staging")
	if found {
		t.Error("expected NavigateTo to return false for wrong namespace")
	}
}

func TestPodsPanelSearchItems(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)

	panel.pods = []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "web-pod-1", Namespace: "default",
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "web-pod-2", Namespace: "staging",
			},
			Status: corev1.PodStatus{Phase: corev1.PodPending},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "db-pod", Namespace: "default",
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
	}

	results := panel.SearchItems("web")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'web', got %d", len(results))
	}

	if results[0].Kind != "Pods" {
		t.Errorf("expected kind 'Pods', got %q", results[0].Kind)
	}
}

func TestNodesPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNodesPanel(client, styles)

	panel.nodes = []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		},
	}

	// Search
	results := panel.SearchItems("node")
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Cluster-scoped: namespace should be empty
	if results[0].Namespace != "" {
		t.Errorf("expected empty namespace for nodes, got %q", results[0].Namespace)
	}

	// NavigateTo ignores namespace for cluster-scoped resources
	panel.SetFilter("")

	found := panel.NavigateTo("node-2", "")
	if !found {
		t.Error("expected NavigateTo to find node-2")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestEventsPanelSearchItems(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewEventsPanel(client, styles)

	panel.events = []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "event-1", Namespace: "default",
			},
			Reason:  "Pulled",
			Message: "Successfully pulled image nginx",
			InvolvedObject: corev1.ObjectReference{
				Name: "nginx-pod",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "event-2", Namespace: "default",
			},
			Reason:  "Failed",
			Message: "Back-off restarting failed container",
			InvolvedObject: corev1.ObjectReference{
				Name: "redis-pod",
			},
		},
	}

	// Search by involved object name
	results := panel.SearchItems("nginx")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'nginx', got %d", len(results))
	}

	// Search by reason
	results = panel.SearchItems("Failed")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Failed', got %d", len(results))
	}

	// Search by message
	results = panel.SearchItems("restarting")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'restarting', got %d", len(results))
	}
}

func TestServicesPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewServicesPanel(client, styles)

	panel.services = []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "web-svc", Namespace: "default",
			},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-svc", Namespace: "staging",
			},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer},
		},
	}

	results := panel.SearchItems("web")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'web', got %d", len(results))
	}

	if results[0].Kind != "Services" {
		t.Errorf("expected kind 'Services', got %q", results[0].Kind)
	}

	if results[0].Status != "ClusterIP" {
		t.Errorf("expected status 'ClusterIP', got %q", results[0].Status)
	}

	// NavigateTo
	panel.SetFilter("")

	found := panel.NavigateTo("api-svc", "staging")
	if !found {
		t.Error("expected NavigateTo to find api-svc")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}

	found = panel.NavigateTo("api-svc", "default")
	if found {
		t.Error("expected false for wrong namespace")
	}
}

func TestConfigMapsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewConfigMapsPanel(client, styles)

	panel.configmaps = []corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "app-config", Namespace: "default",
			},
			Data: map[string]string{"k1": "v1", "k2": "v2"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "db-config", Namespace: "default",
			},
			Data: map[string]string{"host": "localhost"},
		},
	}

	results := panel.SearchItems("app")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'app', got %d", len(results))
	}

	if results[0].Status != "2 keys" {
		t.Errorf("expected status '2 keys', got %q", results[0].Status)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("db-config", "default")
	if !found {
		t.Error("expected NavigateTo to find db-config")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestSecretsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewSecretsPanel(client, styles)

	panel.secrets = []corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tls-secret", Namespace: "default",
			},
			Type: corev1.SecretTypeTLS,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "app-secret", Namespace: "staging",
			},
			Type: corev1.SecretTypeOpaque,
		},
	}

	results := panel.SearchItems("tls")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'tls', got %d", len(results))
	}

	if results[0].Status != string(corev1.SecretTypeTLS) {
		t.Errorf(
			"expected status %q, got %q",
			corev1.SecretTypeTLS, results[0].Status,
		)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("app-secret", "staging")
	if !found {
		t.Error("expected NavigateTo to find app-secret")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestJobsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewJobsPanel(client, styles)

	panel.jobs = []batchv1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "backup-job", Namespace: "default",
			},
			Status: batchv1.JobStatus{Succeeded: 1},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "migrate-job", Namespace: "default",
			},
			Status: batchv1.JobStatus{Active: 1},
		},
	}

	results := panel.SearchItems("backup")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'backup', got %d", len(results))
	}

	if results[0].Kind != "Jobs" {
		t.Errorf("expected kind 'Jobs', got %q", results[0].Kind)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("migrate-job", "default")
	if !found {
		t.Error("expected NavigateTo to find migrate-job")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestCronJobsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nightly-backup", Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{Schedule: "0 0 * * *"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "hourly-sync", Namespace: "staging",
			},
			Spec: batchv1.CronJobSpec{Schedule: "0 * * * *"},
		},
	}

	results := panel.SearchItems("nightly")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'nightly', got %d", len(results))
	}

	if results[0].Kind != "CronJobs" {
		t.Errorf("expected kind 'CronJobs', got %q", results[0].Kind)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("hourly-sync", "staging")
	if !found {
		t.Error("expected NavigateTo to find hourly-sync")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestStatefulSetsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewStatefulSetsPanel(client, styles)

	replicas := int32(3)
	panel.statefulsets = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mysql-sts", Namespace: "default",
			},
			Spec:   appsv1.StatefulSetSpec{Replicas: &replicas},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 3, Replicas: 3},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis-sts", Namespace: "staging",
			},
			Spec:   appsv1.StatefulSetSpec{Replicas: &replicas},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 2, Replicas: 3},
		},
	}

	results := panel.SearchItems("mysql")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'mysql', got %d", len(results))
	}

	if results[0].Kind != "StatefulSets" {
		t.Errorf(
			"expected kind 'StatefulSets', got %q",
			results[0].Kind,
		)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("redis-sts", "staging")
	if !found {
		t.Error("expected NavigateTo to find redis-sts")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestDaemonSetsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDaemonSetsPanel(client, styles)

	panel.daemonsets = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluent-ds", Namespace: "kube-system",
			},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 3,
				NumberReady:            3,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prom-ds", Namespace: "monitoring",
			},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 3,
				NumberReady:            2,
			},
		},
	}

	results := panel.SearchItems("fluent")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'fluent', got %d", len(results))
	}

	if results[0].Kind != "DaemonSets" {
		t.Errorf("expected kind 'DaemonSets', got %q", results[0].Kind)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("prom-ds", "monitoring")
	if !found {
		t.Error("expected NavigateTo to find prom-ds")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestHPAPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewHPAPanel(client, styles)

	minReplicas := int32(1)
	panel.hpas = []autoscalingv2.HorizontalPodAutoscaler{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "web-hpa", Namespace: "default",
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				MinReplicas: &minReplicas,
				MaxReplicas: 10,
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 3,
				DesiredReplicas: 3,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-hpa", Namespace: "staging",
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				MinReplicas: &minReplicas,
				MaxReplicas: 5,
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 2,
				DesiredReplicas: 2,
			},
		},
	}

	results := panel.SearchItems("web")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'web', got %d", len(results))
	}

	if results[0].Kind != "HPAs" {
		t.Errorf("expected kind 'HPAs', got %q", results[0].Kind)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("api-hpa", "staging")
	if !found {
		t.Error("expected NavigateTo to find api-hpa")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestIngressPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewIngressPanel(client, styles)

	panel.ingresses = []networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "web-ingress", Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{Host: "example.com"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-ingress", Namespace: "staging",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{Host: "api.example.com"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bare-ingress", Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{},
		},
	}

	results := panel.SearchItems("web")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'web', got %d", len(results))
	}

	if results[0].Status != "example.com" {
		t.Errorf(
			"expected status 'example.com', got %q",
			results[0].Status,
		)
	}

	// Ingress with no rules should have empty status
	results = panel.SearchItems("bare")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'bare', got %d", len(results))
	}

	if results[0].Status != "" {
		t.Errorf("expected empty status, got %q", results[0].Status)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("api-ingress", "staging")
	if !found {
		t.Error("expected NavigateTo to find api-ingress")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestPVPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPVPanel(client, styles)

	panel.pvs = []corev1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pv-alpha"},
			Status: corev1.PersistentVolumeStatus{
				Phase: corev1.VolumeBound,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pv-beta"},
			Status: corev1.PersistentVolumeStatus{
				Phase: corev1.VolumeAvailable,
			},
		},
	}

	results := panel.SearchItems("alpha")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'alpha', got %d", len(results))
	}

	if results[0].Status != "Bound" {
		t.Errorf("expected status 'Bound', got %q", results[0].Status)
	}

	// Cluster-scoped: namespace should be empty
	if results[0].Namespace != "" {
		t.Errorf(
			"expected empty namespace for PV, got %q",
			results[0].Namespace,
		)
	}

	panel.SetFilter("")

	// Cluster-scoped NavigateTo ignores namespace
	found := panel.NavigateTo("pv-beta", "")
	if !found {
		t.Error("expected NavigateTo to find pv-beta")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestPVCPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPVCPanel(client, styles)

	panel.pvcs = []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "data-pvc", Namespace: "default",
			},
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimBound,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "logs-pvc", Namespace: "staging",
			},
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimPending,
			},
		},
	}

	results := panel.SearchItems("data")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'data', got %d", len(results))
	}

	if results[0].Status != "Bound" {
		t.Errorf("expected status 'Bound', got %q", results[0].Status)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("logs-pvc", "staging")
	if !found {
		t.Error("expected NavigateTo to find logs-pvc")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestNetworkPoliciesPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNetworkPoliciesPanel(client, styles)

	panel.networkPolicies = []networkingv1.NetworkPolicy{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "deny-all", Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{}, {},
				},
				Egress: []networkingv1.NetworkPolicyEgressRule{
					{},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "allow-web", Namespace: "staging",
			},
			Spec: networkingv1.NetworkPolicySpec{
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{},
				},
			},
		},
	}

	results := panel.SearchItems("deny")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'deny', got %d", len(results))
	}

	if results[0].Status != "2I/1E" {
		t.Errorf("expected status '2I/1E', got %q", results[0].Status)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("allow-web", "staging")
	if !found {
		t.Error("expected NavigateTo to find allow-web")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestServiceAccountsPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewServiceAccountsPanel(client, styles)

	panel.serviceAccounts = []corev1.ServiceAccount{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "deploy-sa", Namespace: "default",
			},
			Secrets: []corev1.ObjectReference{
				{Name: "deploy-sa-token"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "monitor-sa", Namespace: "monitoring",
			},
		},
	}

	results := panel.SearchItems("deploy")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'deploy', got %d", len(results))
	}

	if results[0].Kind != "ServiceAccounts" {
		t.Errorf(
			"expected kind 'ServiceAccounts', got %q",
			results[0].Kind,
		)
	}

	panel.SetFilter("")

	found := panel.NavigateTo("monitor-sa", "monitoring")
	if !found {
		t.Error("expected NavigateTo to find monitor-sa")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}

func TestNamespacesPanelSearchAndNavigate(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNamespacesPanel(client, styles)

	panel.namespaces = []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
			Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
			Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "staging"},
			Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
	}

	results := panel.SearchItems("kube")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'kube', got %d", len(results))
	}

	if results[0].Status != "Active" {
		t.Errorf(
			"expected status 'Active', got %q",
			results[0].Status,
		)
	}

	// Cluster-scoped: namespace should be empty
	if results[0].Namespace != "" {
		t.Errorf(
			"expected empty namespace for Namespaces, got %q",
			results[0].Namespace,
		)
	}

	panel.SetFilter("")

	// Cluster-scoped NavigateTo ignores namespace
	found := panel.NavigateTo("staging", "")
	if !found {
		t.Error("expected NavigateTo to find staging")
	}

	if panel.Cursor() != 2 {
		t.Errorf("expected cursor 2, got %d", panel.Cursor())
	}
}

func TestPodsPanelNavigateTo(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPodsPanel(client, styles)

	panel.pods = []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod-a", Namespace: "default",
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod-b", Namespace: "staging",
			},
			Status: corev1.PodStatus{Phase: corev1.PodPending},
		},
	}

	panel.SetFilter("")

	found := panel.NavigateTo("pod-b", "staging")
	if !found {
		t.Error("expected NavigateTo to find pod-b")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}

	found = panel.NavigateTo("pod-b", "default")
	if found {
		t.Error("expected false for wrong namespace")
	}
}

func TestEventsNavigateTo(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewEventsPanel(client, styles)

	panel.events = []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "evt-1", Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{Name: "pod-1"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "evt-2", Namespace: "staging",
			},
			InvolvedObject: corev1.ObjectReference{Name: "pod-2"},
		},
	}

	panel.SetFilter("")

	found := panel.NavigateTo("evt-2", "staging")
	if !found {
		t.Error("expected NavigateTo to find evt-2")
	}

	if panel.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", panel.Cursor())
	}
}
