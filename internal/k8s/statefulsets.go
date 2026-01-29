package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListStatefulSets(
	ctx context.Context,
	namespace string,
) ([]appsv1.StatefulSet, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListStatefulSetsAllNamespaces(ctx context.Context) ([]appsv1.StatefulSet, error) {
	list, err := c.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetStatefulSet(
	ctx context.Context,
	namespace, name string,
) (*appsv1.StatefulSet, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchStatefulSets(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AppsV1().StatefulSets(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteStatefulSet(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) ScaleStatefulSet(
	ctx context.Context,
	namespace, name string,
	replicas int32,
) error {
	if namespace == "" {
		namespace = c.namespace
	}

	scale, err := c.clientset.AppsV1().
		StatefulSets(namespace).
		GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = c.clientset.AppsV1().
		StatefulSets(namespace).
		UpdateScale(ctx, name, scale, metav1.UpdateOptions{})

	return err
}

func (c *Client) RestartStatefulSet(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	// Patch statefulset with a restart annotation to trigger rolling update
	restartedAt := metav1.Now().Format("2006-01-02T15:04:05Z")
	patchFmt := `{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`
	patch := fmt.Appendf(nil, patchFmt, restartedAt)

	_, err := c.clientset.AppsV1().
		StatefulSets(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})

	return err
}

// GetStatefulSetReadyCount returns ready/desired replicas as a string.
func GetStatefulSetReadyCount(sts *appsv1.StatefulSet) string {
	desired := int32(0)
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}

	return fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, desired)
}

// GetStatefulSetImages returns container images from the statefulset spec.
func GetStatefulSetImages(sts *appsv1.StatefulSet) []string {
	images := make([]string, 0)
	for _, container := range sts.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}

	return images
}
