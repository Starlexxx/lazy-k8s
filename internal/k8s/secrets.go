package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListSecrets(ctx context.Context, namespace string) ([]corev1.Secret, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListSecretsAllNamespaces(ctx context.Context) ([]corev1.Secret, error) {
	list, err := c.clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchSecrets(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Secrets(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteSecret(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) UpdateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().
		Secrets(secret.Namespace).
		Update(ctx, secret, metav1.UpdateOptions{})
}

func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().
		Secrets(secret.Namespace).
		Create(ctx, secret, metav1.CreateOptions{})
}
