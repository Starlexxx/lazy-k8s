package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type DeploymentInfo struct {
	Name      string
	Namespace string
	Ready     string
	UpToDate  int32
	Available int32
	Age       string
	Images    []string
}

func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]appsv1.Deployment, error) {
	if namespace == "" {
		namespace = c.namespace
	}
	list, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) ListDeploymentsAllNamespaces(ctx context.Context) ([]appsv1.Deployment, error) {
	list, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = c.namespace
	}
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchDeployments(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}
	return c.clientset.AppsV1().Deployments(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}
	return c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	if namespace == "" {
		namespace = c.namespace
	}

	scale, err := c.clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = c.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	return err
}

func (c *Client) RestartDeployment(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	// Patch deployment with a restart annotation
	patch := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, metav1.Now().Format("2006-01-02T15:04:05Z")))

	_, err := c.clientset.AppsV1().Deployments(namespace).Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *Client) UpdateDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

func GetDeploymentReadyCount(deployment *appsv1.Deployment) string {
	return fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
}

func GetDeploymentImages(deployment *appsv1.Deployment) []string {
	images := make([]string, 0)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}
	return images
}
