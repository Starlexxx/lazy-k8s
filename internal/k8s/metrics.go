package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// PodMetrics contains CPU and memory usage for a pod.
type PodMetrics struct {
	Name      string
	Namespace string
	CPU       int64 // in millicores
	Memory    int64 // in bytes
}

// NodeMetrics contains CPU and memory usage for a node.
type NodeMetrics struct {
	Name   string
	CPU    int64 // in millicores
	Memory int64 // in bytes
}

// MetricsClient wraps the metrics-server client.
type MetricsClient struct {
	client metricsv.Interface
}

// NewMetricsClient creates a new metrics client from the k8s client.
func (c *Client) NewMetricsClient() (*MetricsClient, error) {
	metricsClient, err := metricsv.NewForConfig(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &MetricsClient{client: metricsClient}, nil
}

// GetPodMetrics returns metrics for all pods in the given namespace.
// Returns empty slice if metrics-server is not available.
func (m *MetricsClient) GetPodMetrics(
	ctx context.Context,
	namespace string,
) (map[string]PodMetrics, error) {
	result := make(map[string]PodMetrics)

	podMetricsList, err := m.client.MetricsV1beta1().
		PodMetricses(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		// Return empty map if metrics-server is not available
		return result, nil //nolint:nilerr
	}

	for _, pm := range podMetricsList.Items {
		var totalCPU, totalMemory int64

		for _, container := range pm.Containers {
			totalCPU += container.Usage.Cpu().MilliValue()
			totalMemory += container.Usage.Memory().Value()
		}

		key := pm.Namespace + "/" + pm.Name
		result[key] = PodMetrics{
			Name:      pm.Name,
			Namespace: pm.Namespace,
			CPU:       totalCPU,
			Memory:    totalMemory,
		}
	}

	return result, nil
}

// GetAllPodMetrics returns metrics for all pods across all namespaces.
func (m *MetricsClient) GetAllPodMetrics(ctx context.Context) (map[string]PodMetrics, error) {
	return m.GetPodMetrics(ctx, "")
}

// GetNodeMetrics returns metrics for all nodes.
// Returns empty slice if metrics-server is not available.
func (m *MetricsClient) GetNodeMetrics(ctx context.Context) (map[string]NodeMetrics, error) {
	result := make(map[string]NodeMetrics)

	nodeMetricsList, err := m.client.MetricsV1beta1().
		NodeMetricses().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		// Return empty map if metrics-server is not available
		return result, nil //nolint:nilerr
	}

	for _, nm := range nodeMetricsList.Items {
		result[nm.Name] = NodeMetrics{
			Name:   nm.Name,
			CPU:    nm.Usage.Cpu().MilliValue(),
			Memory: nm.Usage.Memory().Value(),
		}
	}

	return result, nil
}

// IsMetricsServerAvailable checks if metrics-server is available.
func (m *MetricsClient) IsMetricsServerAvailable(ctx context.Context) bool {
	_, err := m.client.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})

	return err == nil
}
