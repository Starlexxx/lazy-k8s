package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type ServiceInfo struct {
	Name       string
	Namespace  string
	Type       string
	ClusterIP  string
	ExternalIP string
	Ports      string
	Age        string
}

func (c *Client) ListServices(ctx context.Context, namespace string) ([]corev1.Service, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListServicesAllNamespaces(ctx context.Context) ([]corev1.Service, error) {
	list, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchServices(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Services(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteService(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) UpdateService(
	ctx context.Context,
	service *corev1.Service,
) (*corev1.Service, error) {
	return c.clientset.CoreV1().
		Services(service.Namespace).
		Update(ctx, service, metav1.UpdateOptions{})
}

func GetServicePorts(service *corev1.Service) string {
	ports := make([]string, 0, len(service.Spec.Ports))
	for _, port := range service.Spec.Ports {
		if port.NodePort > 0 {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
		}
	}

	return strings.Join(ports, ", ")
}

func GetServiceExternalIP(service *corev1.Service) string {
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		ingress := service.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			return ingress.IP
		}

		if ingress.Hostname != "" {
			return ingress.Hostname
		}
	}

	if len(service.Spec.ExternalIPs) > 0 {
		return strings.Join(service.Spec.ExternalIPs, ", ")
	}

	return "<none>"
}
