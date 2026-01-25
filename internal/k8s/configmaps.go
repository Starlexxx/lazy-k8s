package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListConfigMaps(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListConfigMapsAllNamespaces(ctx context.Context) ([]corev1.ConfigMap, error) {
	list, err := c.clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetConfigMap(
	ctx context.Context,
	namespace, name string,
) (*corev1.ConfigMap, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchConfigMaps(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) UpdateConfigMap(
	ctx context.Context,
	cm *corev1.ConfigMap,
) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(cm.Namespace).Update(ctx, cm, metav1.UpdateOptions{})
}

func (c *Client) CreateConfigMap(
	ctx context.Context,
	cm *corev1.ConfigMap,
) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, cm, metav1.CreateOptions{})
}
