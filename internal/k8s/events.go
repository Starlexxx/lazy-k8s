package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type EventInfo struct {
	LastSeen  string
	Type      string
	Reason    string
	Object    string
	Message   string
	Count     int32
	Namespace string
}

func (c *Client) ListEvents(ctx context.Context, namespace string) ([]corev1.Event, error) {
	if namespace == "" {
		namespace = c.namespace
	}
	list, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) ListEventsAllNamespaces(ctx context.Context) ([]corev1.Event, error) {
	list, err := c.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) ListEventsForResource(ctx context.Context, namespace, kind, name string) ([]corev1.Event, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	fieldSelector := "involvedObject.name=" + name + ",involvedObject.kind=" + kind
	list, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) WatchEvents(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}
	return c.clientset.CoreV1().Events(namespace).Watch(ctx, metav1.ListOptions{})
}
