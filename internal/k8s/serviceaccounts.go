package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListServiceAccounts(
	ctx context.Context,
	namespace string,
) ([]corev1.ServiceAccount, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListServiceAccountsAllNamespaces(
	ctx context.Context,
) ([]corev1.ServiceAccount, error) {
	list, err := c.clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetServiceAccount(
	ctx context.Context,
	namespace, name string,
) (*corev1.ServiceAccount, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchServiceAccounts(
	ctx context.Context,
	namespace string,
) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().ServiceAccounts(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteServiceAccount(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetServiceAccountSecretCount returns the number of secrets associated with the SA.
func GetServiceAccountSecretCount(sa *corev1.ServiceAccount) int {
	return len(sa.Secrets)
}

// GetServiceAccountImagePullSecretCount returns the number of image pull secrets.
func GetServiceAccountImagePullSecretCount(sa *corev1.ServiceAccount) int {
	return len(sa.ImagePullSecrets)
}

// GetServiceAccountAutoMount returns whether automount is enabled.
func GetServiceAccountAutoMount(sa *corev1.ServiceAccount) string {
	if sa.AutomountServiceAccountToken == nil {
		return "default"
	}

	if *sa.AutomountServiceAccountToken {
		return "true"
	}

	return "false"
}

// GetServiceAccountSecretsSummary returns a summary of secrets.
func GetServiceAccountSecretsSummary(sa *corev1.ServiceAccount) string {
	return fmt.Sprintf("Secrets: %d, ImagePullSecrets: %d",
		len(sa.Secrets),
		len(sa.ImagePullSecrets),
	)
}
