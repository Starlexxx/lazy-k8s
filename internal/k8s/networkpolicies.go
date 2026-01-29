package k8s

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListNetworkPolicies(
	ctx context.Context,
	namespace string,
) ([]networkingv1.NetworkPolicy, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.NetworkingV1().
		NetworkPolicies(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListNetworkPoliciesAllNamespaces(
	ctx context.Context,
) ([]networkingv1.NetworkPolicy, error) {
	list, err := c.clientset.NetworkingV1().
		NetworkPolicies("").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetNetworkPolicy(
	ctx context.Context,
	namespace, name string,
) (*networkingv1.NetworkPolicy, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.NetworkingV1().
		NetworkPolicies(namespace).
		Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchNetworkPolicies(
	ctx context.Context,
	namespace string,
) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.NetworkingV1().
		NetworkPolicies(namespace).
		Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteNetworkPolicy(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.NetworkingV1().
		NetworkPolicies(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
}

// GetNetworkPolicyIngressRuleCount returns the number of ingress rules.
func GetNetworkPolicyIngressRuleCount(np *networkingv1.NetworkPolicy) int {
	return len(np.Spec.Ingress)
}

// GetNetworkPolicyEgressRuleCount returns the number of egress rules.
func GetNetworkPolicyEgressRuleCount(np *networkingv1.NetworkPolicy) int {
	return len(np.Spec.Egress)
}

// GetNetworkPolicyPodSelectorString returns the pod selector as a string.
func GetNetworkPolicyPodSelectorString(np *networkingv1.NetworkPolicy) string {
	selector := np.Spec.PodSelector
	if len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0 {
		return "<all pods>"
	}

	return metav1.FormatLabelSelector(&selector)
}

// GetNetworkPolicyPolicyTypes returns the policy types as a string.
func GetNetworkPolicyPolicyTypes(np *networkingv1.NetworkPolicy) string {
	if len(np.Spec.PolicyTypes) == 0 {
		return "Ingress"
	}

	var types string

	for i, pt := range np.Spec.PolicyTypes {
		if i > 0 {
			types += ", "
		}

		types += string(pt)
	}

	return types
}

// GetNetworkPolicyRuleSummary returns a summary of ingress/egress rules.
func GetNetworkPolicyRuleSummary(np *networkingv1.NetworkPolicy) string {
	return fmt.Sprintf("Ingress: %d, Egress: %d",
		len(np.Spec.Ingress),
		len(np.Spec.Egress),
	)
}
