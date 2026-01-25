package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type NodeInfo struct {
	Name             string
	Status           string
	Roles            string
	Age              string
	Version          string
	InternalIP       string
	ExternalIP       string
	OS               string
	KernelVersion    string
	ContainerRuntime string
	CPU              string
	Memory           string
}

func (c *Client) ListNodes(ctx context.Context) ([]corev1.Node, error) {
	list, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	return c.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchNodes(ctx context.Context) (watch.Interface, error) {
	return c.clientset.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{})
}

func GetNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}

			return "NotReady"
		}
	}

	return "Unknown"
}

func GetNodeRoles(node *corev1.Node) string {
	roles := ""

	for label := range node.Labels {
		switch label {
		case "node-role.kubernetes.io/master", "node-role.kubernetes.io/control-plane":
			if roles != "" {
				roles += ","
			}

			roles += "control-plane"
		case "node-role.kubernetes.io/worker":
			if roles != "" {
				roles += ","
			}

			roles += "worker"
		}
	}

	if roles == "" {
		roles = "<none>"
	}

	return roles
}

func GetNodeInternalIP(node *corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}

	return "<none>"
}

func GetNodeExternalIP(node *corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			return addr.Address
		}
	}

	return "<none>"
}

func GetNodeCapacity(node *corev1.Node) (cpu string, memory string) {
	cpuQty := node.Status.Capacity[corev1.ResourceCPU]
	memQty := node.Status.Capacity[corev1.ResourceMemory]

	cpu = cpuQty.String()
	memory = fmt.Sprintf("%.1fGi", float64(memQty.Value())/(1024*1024*1024))

	return cpu, memory
}
