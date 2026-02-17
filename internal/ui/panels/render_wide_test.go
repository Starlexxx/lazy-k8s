package panels

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestDeploymentsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{testDeployment()}
	panel.filtered = panel.deployments
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-deploy") {
		t.Error("narrow view should contain deployment name")
	}
}

func TestDeploymentsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDeploymentsPanel(client, styles)

	panel.deployments = []appsv1.Deployment{testDeployment()}
	panel.filtered = panel.deployments
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-deploy") {
		t.Error("wide view should contain deployment name")
	}

	// Wide mode shows images
	if !strings.Contains(view, "myapp") {
		t.Error("wide view should contain container image")
	}
}

func TestServicesPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewServicesPanel(client, styles)

	panel.services = []corev1.Service{testService()}
	panel.filtered = panel.services
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-svc") {
		t.Error("narrow view should contain service name")
	}
}

func TestServicesPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewServicesPanel(client, styles)

	panel.services = []corev1.Service{testService()}
	panel.filtered = panel.services
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-svc") {
		t.Error("wide view should contain service name")
	}

	// Wide mode shows ClusterIP
	if !strings.Contains(view, "10.0.0.1") {
		t.Error("wide view should contain cluster IP")
	}
}

func TestEventsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewEventsPanel(client, styles)

	panel.events = []corev1.Event{testEvent()}
	panel.filtered = panel.events
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "Pulled") {
		t.Error("narrow view should contain event reason")
	}
}

func TestEventsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewEventsPanel(client, styles)

	panel.events = []corev1.Event{testEvent()}
	panel.filtered = panel.events
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	// Wide mode shows the object reference (Kind/Name)
	if !strings.Contains(view, "Pod") {
		t.Error("wide view should contain involved object kind")
	}
}

func TestPVCPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPVCPanel(client, styles)

	panel.pvcs = []corev1.PersistentVolumeClaim{testPVC()}
	panel.filtered = panel.pvcs
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-pvc") {
		t.Error("narrow view should contain PVC name")
	}
}

func TestPVCPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPVCPanel(client, styles)

	panel.pvcs = []corev1.PersistentVolumeClaim{testPVC()}
	panel.filtered = panel.pvcs
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-pvc") {
		t.Error("wide view should contain PVC name")
	}

	// Wide mode shows capacity
	if !strings.Contains(view, "10Gi") {
		t.Error("wide view should contain capacity")
	}
}

func TestCronJobsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{testCronJob()}
	panel.filtered = panel.cronjobs
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-cronjob") {
		t.Error("narrow view should contain cronjob name")
	}
}

func TestCronJobsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewCronJobsPanel(client, styles)

	panel.cronjobs = []batchv1.CronJob{testCronJob()}
	panel.filtered = panel.cronjobs
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-cronjob") {
		t.Error("wide view should contain cronjob name")
	}

	// Wide mode shows schedule
	if !strings.Contains(view, "*/5") {
		t.Error("wide view should contain cron schedule")
	}
}

func TestJobsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewJobsPanel(client, styles)

	panel.jobs = []batchv1.Job{testJob()}
	panel.filtered = panel.jobs
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-job") {
		t.Error("narrow view should contain job name")
	}
}

func TestJobsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewJobsPanel(client, styles)

	panel.jobs = []batchv1.Job{testJob()}
	panel.filtered = panel.jobs
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-job") {
		t.Error("wide view should contain job name")
	}
}

func TestHPAPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewHPAPanel(client, styles)

	panel.hpas = []autoscalingv2.HorizontalPodAutoscaler{testHPA()}
	panel.filtered = panel.hpas
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-hpa") {
		t.Error("narrow view should contain HPA name")
	}
}

func TestHPAPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewHPAPanel(client, styles)

	panel.hpas = []autoscalingv2.HorizontalPodAutoscaler{testHPA()}
	panel.filtered = panel.hpas
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-hpa") {
		t.Error("wide view should contain HPA name")
	}

	// Wide mode shows target ref (truncated to 15 chars: "Deployment/test")
	if !strings.Contains(view, "Deployment/") {
		t.Error("wide view should contain HPA target reference")
	}
}

func TestStatefulSetsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewStatefulSetsPanel(client, styles)

	panel.statefulsets = []appsv1.StatefulSet{testStatefulSet()}
	panel.filtered = panel.statefulsets
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-sts") {
		t.Error("narrow view should contain statefulset name")
	}
}

func TestStatefulSetsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewStatefulSetsPanel(client, styles)

	panel.statefulsets = []appsv1.StatefulSet{testStatefulSet()}
	panel.filtered = panel.statefulsets
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-sts") {
		t.Error("wide view should contain statefulset name")
	}

	// Wide mode shows ready count (e.g. "3/3")
	if !strings.Contains(view, "3/3") {
		t.Error("wide view should contain ready count")
	}
}

func TestDaemonSetsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDaemonSetsPanel(client, styles)

	panel.daemonsets = []appsv1.DaemonSet{testDaemonSet()}
	panel.filtered = panel.daemonsets
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-ds") {
		t.Error("narrow view should contain daemonset name")
	}
}

func TestDaemonSetsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewDaemonSetsPanel(client, styles)

	panel.daemonsets = []appsv1.DaemonSet{testDaemonSet()}
	panel.filtered = panel.daemonsets
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-ds") {
		t.Error("wide view should contain daemonset name")
	}

	// Wide mode shows ready count (e.g. "3/3")
	if !strings.Contains(view, "3/3") {
		t.Error("wide view should contain ready count")
	}
}

func TestConfigMapsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewConfigMapsPanel(client, styles)

	panel.configmaps = []corev1.ConfigMap{testConfigMap()}
	panel.filtered = panel.configmaps
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-cm") {
		t.Error("narrow view should contain configmap name")
	}
}

func TestConfigMapsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewConfigMapsPanel(client, styles)

	panel.configmaps = []corev1.ConfigMap{testConfigMap()}
	panel.filtered = panel.configmaps
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-cm") {
		t.Error("wide view should contain configmap name")
	}
}

func TestSecretsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewSecretsPanel(client, styles)

	panel.secrets = []corev1.Secret{testSecret()}
	panel.filtered = panel.secrets
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-secret") {
		t.Error("narrow view should contain secret name")
	}
}

func TestSecretsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewSecretsPanel(client, styles)

	panel.secrets = []corev1.Secret{testSecret()}
	panel.filtered = panel.secrets
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-secret") {
		t.Error("wide view should contain secret name")
	}

	// Wide mode shows secret type
	if !strings.Contains(view, "Opaque") {
		t.Error("wide view should contain secret type")
	}
}

func TestIngressPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewIngressPanel(client, styles)

	panel.ingresses = []networkingv1.Ingress{testIngress()}
	panel.filtered = panel.ingresses
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-ingress") {
		t.Error("narrow view should contain ingress name")
	}
}

func TestIngressPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewIngressPanel(client, styles)

	panel.ingresses = []networkingv1.Ingress{testIngress()}
	panel.filtered = panel.ingresses
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-ingress") {
		t.Error("wide view should contain ingress name")
	}

	// Wide mode shows hosts
	if !strings.Contains(view, "example.com") {
		t.Error("wide view should contain ingress host")
	}
}

func TestPVPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPVPanel(client, styles)

	panel.pvs = []corev1.PersistentVolume{testPV()}
	panel.filtered = panel.pvs
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-pv") {
		t.Error("narrow view should contain PV name")
	}
}

func TestPVPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewPVPanel(client, styles)

	panel.pvs = []corev1.PersistentVolume{testPV()}
	panel.filtered = panel.pvs
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-pv") {
		t.Error("wide view should contain PV name")
	}

	// Wide mode shows capacity
	if !strings.Contains(view, "10Gi") {
		t.Error("wide view should contain PV capacity")
	}
}

func TestNetworkPoliciesPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNetworkPoliciesPanel(client, styles)

	panel.networkPolicies = []networkingv1.NetworkPolicy{testNetworkPolicy()}
	panel.filtered = panel.networkPolicies
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-netpol") {
		t.Error("narrow view should contain network policy name")
	}
}

func TestNetworkPoliciesPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNetworkPoliciesPanel(client, styles)

	panel.networkPolicies = []networkingv1.NetworkPolicy{testNetworkPolicy()}
	panel.filtered = panel.networkPolicies
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-netpol") {
		t.Error("wide view should contain network policy name")
	}
}

func TestServiceAccountsPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewServiceAccountsPanel(client, styles)

	panel.serviceAccounts = []corev1.ServiceAccount{testServiceAccount()}
	panel.filtered = panel.serviceAccounts
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-sa") {
		t.Error("narrow view should contain service account name")
	}
}

func TestServiceAccountsPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewServiceAccountsPanel(client, styles)

	panel.serviceAccounts = []corev1.ServiceAccount{testServiceAccount()}
	panel.filtered = panel.serviceAccounts
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-sa") {
		t.Error("wide view should contain service account name")
	}
}

func TestNamespacesPanel_ViewNarrow(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNamespacesPanel(client, styles)

	panel.namespaces = []corev1.Namespace{testNamespace()}
	panel.filtered = panel.namespaces
	panel.SetSize(60, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-ns") {
		t.Error("narrow view should contain namespace name")
	}
}

func TestNamespacesPanel_ViewWide(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	panel := NewNamespacesPanel(client, styles)

	panel.namespaces = []corev1.Namespace{testNamespace()}
	panel.filtered = panel.namespaces
	panel.SetSize(150, 30)
	panel.SetFocused(true)

	view := panel.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}

	if !strings.Contains(view, "test-ns") {
		t.Error("wide view should contain namespace name")
	}
}

// TestAllPanels_WidthBoundary verifies that all 18 panels render without
// panicking at the exact threshold where rendering switches from narrow
// to wide mode. All panels use `width > 80` as the branch condition,
// so width=80 exercises narrow rendering and width=81 exercises wide.
func TestAllPanels_WidthBoundary(t *testing.T) {
	client := createTestK8sClient()
	styles := createTestStyles()
	tests := []struct {
		name  string
		panel Panel
	}{
		{"Pods", NewPodsPanel(client, styles)},
		{"Nodes", NewNodesPanel(client, styles)},
		{"Deployments", NewDeploymentsPanel(client, styles)},
		{"Services", NewServicesPanel(client, styles)},
		{"Events", NewEventsPanel(client, styles)},
		{"Namespaces", NewNamespacesPanel(client, styles)},
		{"ConfigMaps", NewConfigMapsPanel(client, styles)},
		{"Secrets", NewSecretsPanel(client, styles)},
		{"CronJobs", NewCronJobsPanel(client, styles)},
		{"Jobs", NewJobsPanel(client, styles)},
		{"HPA", NewHPAPanel(client, styles)},
		{"StatefulSets", NewStatefulSetsPanel(client, styles)},
		{"DaemonSets", NewDaemonSetsPanel(client, styles)},
		{"Ingress", NewIngressPanel(client, styles)},
		{"PV", NewPVPanel(client, styles)},
		{"PVC", NewPVCPanel(client, styles)},
		{"NetworkPolicies", NewNetworkPoliciesPanel(client, styles)},
		{"ServiceAccounts", NewServiceAccountsPanel(client, styles)},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_at_80", func(t *testing.T) {
			tt.panel.SetSize(80, 30)
			tt.panel.SetFocused(true)

			view := tt.panel.View()
			if view == "" {
				t.Fatal("View() returned empty string at width=80")
			}
		})

		t.Run(tt.name+"_at_81", func(t *testing.T) {
			tt.panel.SetSize(81, 30)
			tt.panel.SetFocused(true)

			view := tt.panel.View()
			if view == "" {
				t.Fatal("View() returned empty string at width=81")
			}
		})
	}
}
