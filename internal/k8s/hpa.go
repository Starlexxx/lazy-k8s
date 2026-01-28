package k8s

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListHPAs(
	ctx context.Context,
	namespace string,
) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListHPAsAllNamespaces(
	ctx context.Context,
) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	list, err := c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers("").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetHPA(
	ctx context.Context,
	namespace, name string,
) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers(namespace).
		Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchHPAs(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers(namespace).
		Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteHPA(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
}

// UpdateHPAMinReplicas updates the minimum replicas for an HPA.
func (c *Client) UpdateHPAMinReplicas(
	ctx context.Context,
	namespace, name string,
	minReplicas int32,
) error {
	if namespace == "" {
		namespace = c.namespace
	}

	patch := fmt.Appendf(nil, `{"spec":{"minReplicas":%d}}`, minReplicas)

	_, err := c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})

	return err
}

// UpdateHPAMaxReplicas updates the maximum replicas for an HPA.
func (c *Client) UpdateHPAMaxReplicas(
	ctx context.Context,
	namespace, name string,
	maxReplicas int32,
) error {
	if namespace == "" {
		namespace = c.namespace
	}

	patch := fmt.Appendf(nil, `{"spec":{"maxReplicas":%d}}`, maxReplicas)

	_, err := c.clientset.AutoscalingV2().
		HorizontalPodAutoscalers(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})

	return err
}

// GetHPAReplicaCount returns current/min/max replicas as a string.
func GetHPAReplicaCount(hpa *autoscalingv2.HorizontalPodAutoscaler) string {
	minReplicas := int32(1)
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	return fmt.Sprintf("%d (%d-%d)",
		hpa.Status.CurrentReplicas,
		minReplicas,
		hpa.Spec.MaxReplicas,
	)
}

// GetHPATargetRef returns the target reference as a string.
func GetHPATargetRef(hpa *autoscalingv2.HorizontalPodAutoscaler) string {
	return fmt.Sprintf("%s/%s",
		hpa.Spec.ScaleTargetRef.Kind,
		hpa.Spec.ScaleTargetRef.Name,
	)
}

// GetHPAMetricsSummary returns a summary of configured metrics.
func GetHPAMetricsSummary(hpa *autoscalingv2.HorizontalPodAutoscaler) string {
	if len(hpa.Spec.Metrics) == 0 {
		return "No metrics configured"
	}

	return fmt.Sprintf("%d metric(s)", len(hpa.Spec.Metrics))
}
