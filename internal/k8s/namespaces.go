package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type NamespaceInfo struct {
	Name   string
	Status string
	Age    string
	Labels map[string]string
}

func (c *Client) ListNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	list, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchNamespaces(ctx context.Context) (watch.Interface, error) {
	return c.clientset.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{})
}

func (c *Client) CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
}

func (c *Client) DeleteNamespace(ctx context.Context, name string) error {
	return c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}
