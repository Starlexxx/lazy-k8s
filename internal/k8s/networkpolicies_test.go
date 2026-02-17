package k8s

import (
	"context"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListNetworkPolicies(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "netpol-1",
				Namespace: "default",
			},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "netpol-2",
				Namespace: "default",
			},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-netpol",
				Namespace: "other-namespace",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	netpols, err := client.ListNetworkPolicies(ctx, "default")
	if err != nil {
		t.Fatalf("ListNetworkPolicies returned unexpected error: %v", err)
	}

	if len(netpols) != 2 {
		t.Errorf("ListNetworkPolicies returned %d network policies, want 2", len(netpols))
	}

	netpols, err = client.ListNetworkPolicies(ctx, "other-namespace")
	if err != nil {
		t.Fatalf("ListNetworkPolicies returned unexpected error: %v", err)
	}

	if len(netpols) != 1 {
		t.Errorf("ListNetworkPolicies returned %d network policies, want 1", len(netpols))
	}

	netpols, err = client.ListNetworkPolicies(ctx, "")
	if err != nil {
		t.Fatalf("ListNetworkPolicies returned unexpected error: %v", err)
	}

	if len(netpols) != 2 {
		t.Errorf(
			"ListNetworkPolicies with empty namespace returned %d network policies, want 2",
			len(netpols),
		)
	}
}

func TestListNetworkPoliciesAllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "netpol-1",
				Namespace: "default",
			},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "netpol-2",
				Namespace: "kube-system",
			},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "netpol-3",
				Namespace: "my-app",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	netpols, err := client.ListNetworkPoliciesAllNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListNetworkPoliciesAllNamespaces returned unexpected error: %v", err)
	}

	if len(netpols) != 3 {
		t.Errorf(
			"ListNetworkPoliciesAllNamespaces returned %d network policies, want 3",
			len(netpols),
		)
	}
}

func TestGetNetworkPolicy(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-netpol",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	np, err := client.GetNetworkPolicy(ctx, "default", "test-netpol")
	if err != nil {
		t.Fatalf("GetNetworkPolicy returned unexpected error: %v", err)
	}

	if np.Name != "test-netpol" {
		t.Errorf(
			"GetNetworkPolicy returned network policy with name %q, want %q",
			np.Name,
			"test-netpol",
		)
	}

	_, err = client.GetNetworkPolicy(ctx, "default", "non-existent")
	if err == nil {
		t.Error("GetNetworkPolicy should have returned an error for non-existent network policy")
	}
}

func TestDeleteNetworkPolicy(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-netpol",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	err := client.DeleteNetworkPolicy(ctx, "default", "test-netpol")
	if err != nil {
		t.Fatalf("DeleteNetworkPolicy returned unexpected error: %v", err)
	}

	_, err = client.GetNetworkPolicy(ctx, "default", "test-netpol")
	if err == nil {
		t.Error("NetworkPolicy should have been deleted")
	}
}

func TestWatchNetworkPolicies(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-netpol",
				Namespace: "default",
			},
		},
	)

	client := createTestClient(clientset)
	ctx := context.Background()

	watcher, err := client.WatchNetworkPolicies(ctx, "default")
	if err != nil {
		t.Fatalf("WatchNetworkPolicies returned unexpected error: %v", err)
	}
	defer watcher.Stop()

	if watcher.ResultChan() == nil {
		t.Error("WatchNetworkPolicies returned watcher with nil ResultChan")
	}
}

func TestGetNetworkPolicyIngressRuleCount(t *testing.T) {
	tests := []struct {
		name     string
		np       *networkingv1.NetworkPolicy
		expected int
	}{
		{
			name: "no ingress rules",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Ingress: []networkingv1.NetworkPolicyIngressRule{},
				},
			},
			expected: 0,
		},
		{
			name: "some ingress rules",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{},
						{},
						{},
					},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNetworkPolicyIngressRuleCount(tt.np)
			if result != tt.expected {
				t.Errorf("GetNetworkPolicyIngressRuleCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetNetworkPolicyEgressRuleCount(t *testing.T) {
	tests := []struct {
		name     string
		np       *networkingv1.NetworkPolicy
		expected int
	}{
		{
			name: "no egress rules",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Egress: []networkingv1.NetworkPolicyEgressRule{},
				},
			},
			expected: 0,
		},
		{
			name: "some egress rules",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Egress: []networkingv1.NetworkPolicyEgressRule{
						{},
						{},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNetworkPolicyEgressRuleCount(tt.np)
			if result != tt.expected {
				t.Errorf("GetNetworkPolicyEgressRuleCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetNetworkPolicyPodSelectorString(t *testing.T) {
	tests := []struct {
		name     string
		np       *networkingv1.NetworkPolicy
		expected string
	}{
		{
			name: "empty selector (all pods)",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
				},
			},
			expected: "<all pods>",
		},
		{
			name: "with match labels",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "web",
						},
					},
				},
			},
			expected: "app=web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNetworkPolicyPodSelectorString(tt.np)
			if result != tt.expected {
				t.Errorf("GetNetworkPolicyPodSelectorString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetNetworkPolicyPolicyTypes(t *testing.T) {
	tests := []struct {
		name     string
		np       *networkingv1.NetworkPolicy
		expected string
	}{
		{
			name: "no policy types (defaults to Ingress)",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					PolicyTypes: []networkingv1.PolicyType{},
				},
			},
			expected: "Ingress",
		},
		{
			name: "ingress only",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
					},
				},
			},
			expected: "Ingress",
		},
		{
			name: "egress only",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeEgress,
					},
				},
			},
			expected: "Egress",
		},
		{
			name: "both ingress and egress",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
						networkingv1.PolicyTypeEgress,
					},
				},
			},
			expected: "Ingress, Egress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNetworkPolicyPolicyTypes(tt.np)
			if result != tt.expected {
				t.Errorf("GetNetworkPolicyPolicyTypes() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetNetworkPolicyRuleSummary(t *testing.T) {
	port80 := intstr.FromInt(80)

	tests := []struct {
		name     string
		np       *networkingv1.NetworkPolicy
		expected string
	}{
		{
			name: "no rules",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Ingress: []networkingv1.NetworkPolicyIngressRule{},
					Egress:  []networkingv1.NetworkPolicyEgressRule{},
				},
			},
			expected: "Ingress: 0, Egress: 0",
		},
		{
			name: "ingress only",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{Port: &port80},
							},
						},
						{},
					},
					Egress: []networkingv1.NetworkPolicyEgressRule{},
				},
			},
			expected: "Ingress: 2, Egress: 0",
		},
		{
			name: "both rules",
			np: &networkingv1.NetworkPolicy{
				Spec: networkingv1.NetworkPolicySpec{
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{},
					},
					Egress: []networkingv1.NetworkPolicyEgressRule{
						{},
						{},
						{},
					},
				},
			},
			expected: "Ingress: 1, Egress: 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNetworkPolicyRuleSummary(tt.np)
			if result != tt.expected {
				t.Errorf("GetNetworkPolicyRuleSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}
