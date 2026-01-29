package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListDaemonSets(
	ctx context.Context,
	namespace string,
) ([]appsv1.DaemonSet, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListDaemonSetsAllNamespaces(ctx context.Context) ([]appsv1.DaemonSet, error) {
	list, err := c.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetDaemonSet(
	ctx context.Context,
	namespace, name string,
) (*appsv1.DaemonSet, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchDaemonSets(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AppsV1().DaemonSets(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteDaemonSet(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) RestartDaemonSet(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	// Patch daemonset with a restart annotation to trigger rolling update
	restartedAt := metav1.Now().Format("2006-01-02T15:04:05Z")
	patchFmt := `{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`
	patch := fmt.Appendf(nil, patchFmt, restartedAt)

	_, err := c.clientset.AppsV1().
		DaemonSets(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})

	return err
}

// GetDaemonSetReadyCount returns ready/desired nodes as a string.
func GetDaemonSetReadyCount(ds *appsv1.DaemonSet) string {
	return fmt.Sprintf("%d/%d", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
}

// GetDaemonSetImages returns container images from the daemonset spec.
func GetDaemonSetImages(ds *appsv1.DaemonSet) []string {
	images := make([]string, 0)
	for _, container := range ds.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}

	return images
}
