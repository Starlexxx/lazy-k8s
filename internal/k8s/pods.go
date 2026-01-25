package k8s

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodInfo struct {
	Name       string
	Namespace  string
	Status     string
	Ready      string
	Restarts   int32
	Age        string
	Node       string
	IP         string
	Containers []ContainerInfo
}

type ContainerInfo struct {
	Name     string
	Ready    bool
	Restarts int32
	State    string
	Image    string
}

func (c *Client) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListPodsAllNamespaces(ctx context.Context) ([]corev1.Pod, error) {
	list, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchPods(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeletePod(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) GetPodLogs(
	ctx context.Context,
	namespace, name, container string,
	follow bool,
	tailLines int64,
) (io.ReadCloser, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := &corev1.PodLogOptions{
		Follow: follow,
	}

	if container != "" {
		opts.Container = container
	}

	if tailLines > 0 {
		opts.TailLines = &tailLines
	}

	return c.clientset.CoreV1().Pods(namespace).GetLogs(name, opts).Stream(ctx)
}

func GetPodStatus(pod *corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return "Running"
		}
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			return cs.State.Waiting.Reason
		}

		if cs.State.Terminated != nil {
			return cs.State.Terminated.Reason
		}
	}

	return string(pod.Status.Phase)
}

func GetPodReadyCount(pod *corev1.Pod) string {
	ready := 0
	total := len(pod.Spec.Containers)

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}

	return fmt.Sprintf("%d/%d", ready, total)
}

func GetPodRestarts(pod *corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}

	return restarts
}
