package k8s

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) ListCronJobs(
	ctx context.Context,
	namespace string,
) ([]batchv1.CronJob, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	list, err := c.clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) ListCronJobsAllNamespaces(ctx context.Context) ([]batchv1.CronJob, error) {
	list, err := c.clientset.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) GetCronJob(
	ctx context.Context,
	namespace, name string,
) (*batchv1.CronJob, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) WatchCronJobs(ctx context.Context, namespace string) (watch.Interface, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.BatchV1().CronJobs(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *Client) DeleteCronJob(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	return c.clientset.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) TriggerCronJob(ctx context.Context, namespace, name string) (*batchv1.Job, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	cronJob, err := c.GetCronJob(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cronJob.Name + "-manual-",
			Namespace:    namespace,
			Labels:       cronJob.Spec.JobTemplate.Labels,
			Annotations: map[string]string{
				"cronjob.kubernetes.io/instantiate": "manual",
			},
			// Owner reference ensures garbage collection when the CronJob is deleted
			// and lets the CronJob controller count this in its active jobs list.
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cronJob.Name,
					UID:        cronJob.UID,
				},
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}

	return c.clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
}

func (c *Client) SuspendCronJob(ctx context.Context, namespace, name string, suspend bool) error {
	if namespace == "" {
		namespace = c.namespace
	}

	patch := fmt.Appendf(nil, `{"spec":{"suspend":%t}}`, suspend)

	_, err := c.clientset.BatchV1().
		CronJobs(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})

	return err
}

func GetCronJobStatus(cj *batchv1.CronJob) string {
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		return "Suspended"
	}

	return "Active"
}

func GetCronJobLastSchedule(cj *batchv1.CronJob) string {
	if cj.Status.LastScheduleTime == nil {
		return "Never"
	}

	return cj.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
}

func GetCronJobActiveJobs(cj *batchv1.CronJob) int {
	return len(cj.Status.Active)
}
