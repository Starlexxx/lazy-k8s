package panels

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/Starlexxx/lazy-k8s/internal/config"
	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

func createTestStyles() *theme.Styles {
	cfg := &config.ThemeConfig{
		PrimaryColor:    "#7aa2f7",
		SecondaryColor:  "#9ece6a",
		ErrorColor:      "#f7768e",
		WarningColor:    "#e0af68",
		BackgroundColor: "#1a1b26",
		TextColor:       "#c0caf5",
		BorderColor:     "#3b4261",
	}

	return theme.NewStyles(cfg)
}

// createTestK8sClient returns a client with an empty fake clientset.
// Tests inject data directly into panel fields (e.g. panel.pods)
// rather than going through the k8s API, so no seed objects are needed.
func createTestK8sClient() *k8s.Client {
	fakeClientset := fake.NewSimpleClientset()

	return k8s.NewTestClient(fakeClientset)
}

// testPod returns a running pod with ContainerStatuses populated.
// ContainerStatuses are required because GetPodReadyCount and
// GetPodRestarts read them for wide-mode column rendering.
func testPod() corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main", Image: "nginx:latest"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "main",
					Ready:        true,
					RestartCount: 0,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
		},
	}
}

// testNode returns a Ready node with the control-plane role label.
// The role label is needed because GetNodeRoles reads it for
// the ROLES column in wide-mode rendering.
func testNode() corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-node",
			CreationTimestamp: metav1.Now(),
			Labels: map[string]string{
				"node-role.kubernetes.io/control-plane": "",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: "v1.28.0",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}
}

func testDeployment() appsv1.Deployment {
	replicas := int32(3)

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-deploy",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "app", Image: "myapp:v1"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     3,
			Replicas:          3,
			AvailableReplicas: 3,
		},
	}
}

func testService() corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-svc",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "10.0.0.1",
			Ports: []corev1.ServicePort{
				{Port: 80, Protocol: corev1.ProtocolTCP},
			},
		},
	}
}

func testEvent() corev1.Event {
	return corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-event",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod",
			Name: "test-pod",
		},
		Reason:        "Pulled",
		Type:          "Normal",
		Message:       "Successfully pulled image",
		LastTimestamp: metav1.Now(),
	}
}

func testNamespace() corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-ns",
			CreationTimestamp: metav1.Now(),
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}
}

func testConfigMap() corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-cm",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
}

func testSecret() corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-secret",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte("secret"),
		},
	}
}

func testCronJob() batchv1.CronJob {
	return batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-cronjob",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: batchv1.CronJobSpec{
			Schedule:    "*/5 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{},
		},
	}
}

func testJob() batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-job",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Status: batchv1.JobStatus{
			Succeeded: 1,
		},
	}
}

func testHPA() autoscalingv2.HorizontalPodAutoscaler {
	minReplicas := int32(1)

	return autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-hpa",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: &minReplicas,
			MaxReplicas: 10,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: "test-deploy",
			},
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			CurrentReplicas: 3,
			DesiredReplicas: 3,
		},
	}
}

func testStatefulSet() appsv1.StatefulSet {
	replicas := int32(3)

	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-sts",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas: 3,
			Replicas:      3,
		},
	}
}

func testDaemonSet() appsv1.DaemonSet {
	return appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-ds",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			NumberReady:            3,
		},
	}
}

func testIngress() networkingv1.Ingress {
	return networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-ingress",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{Host: "example.com"},
			},
		},
	}
}

func testPV() corev1.PersistentVolume {
	return corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pv",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("10Gi"),
			},
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeBound,
		},
	}
}

func testPVC() corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pvc",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("10Gi"),
			},
		},
	}
}

func testNetworkPolicy() networkingv1.NetworkPolicy {
	return networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-netpol",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: networkingv1.NetworkPolicySpec{
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{},
			},
		},
	}
}

func testServiceAccount() corev1.ServiceAccount {
	return corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-sa",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Secrets: []corev1.ObjectReference{
			{Name: "test-sa-token"},
		},
	}
}
