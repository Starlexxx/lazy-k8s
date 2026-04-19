package ui

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWorkloadLogTarget(t *testing.T) {
	sel := &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "demo"},
	}

	tests := []struct {
		name         string
		item         any
		wantKind     string
		wantName     string
		wantNs       string
		wantSelector string
		wantOK       bool
	}{
		{
			name: "deployment",
			item: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "prod"},
				Spec:       appsv1.DeploymentSpec{Selector: sel},
			},
			wantKind:     "Deployment",
			wantName:     "web",
			wantNs:       "prod",
			wantSelector: "app=demo",
			wantOK:       true,
		},
		{
			name: "statefulset",
			item: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "data"},
				Spec:       appsv1.StatefulSetSpec{Selector: sel},
			},
			wantKind:     "StatefulSet",
			wantName:     "db",
			wantNs:       "data",
			wantSelector: "app=demo",
			wantOK:       true,
		},
		{
			name: "daemonset",
			item: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Name: "agent", Namespace: "kube-system"},
				Spec:       appsv1.DaemonSetSpec{Selector: sel},
			},
			wantKind:     "DaemonSet",
			wantName:     "agent",
			wantNs:       "kube-system",
			wantSelector: "app=demo",
			wantOK:       true,
		},
		{
			name: "job",
			item: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "migrate", Namespace: "prod"},
				Spec:       batchv1.JobSpec{Selector: sel},
			},
			wantKind:     "Job",
			wantName:     "migrate",
			wantNs:       "prod",
			wantSelector: "app=demo",
			wantOK:       true,
		},
		{
			name:   "nil item",
			item:   nil,
			wantOK: false,
		},
		{
			name:   "unsupported kind",
			item:   &appsv1.ReplicaSet{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, name, ns, selector, ok := workloadLogTarget(tt.item)

			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}

			if !ok {
				return
			}

			if kind != tt.wantKind {
				t.Errorf("kind = %q, want %q", kind, tt.wantKind)
			}

			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}

			if ns != tt.wantNs {
				t.Errorf("namespace = %q, want %q", ns, tt.wantNs)
			}

			if selector != tt.wantSelector {
				t.Errorf("selector = %q, want %q", selector, tt.wantSelector)
			}
		})
	}
}
