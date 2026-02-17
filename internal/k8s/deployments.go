package k8s

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/yaml"
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

func (c *Client) ListDeployments(
	ctx context.Context,
	namespace string,
) ([]appsv1.Deployment, error) {
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

func (c *Client) GetDeployment(
	ctx context.Context,
	namespace, name string,
) (*appsv1.Deployment, error) {
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

func (c *Client) ScaleDeployment(
	ctx context.Context,
	namespace, name string,
	replicas int32,
) error {
	if namespace == "" {
		namespace = c.namespace
	}

	scale, err := c.clientset.AppsV1().
		Deployments(namespace).
		GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = c.clientset.AppsV1().
		Deployments(namespace).
		UpdateScale(ctx, name, scale, metav1.UpdateOptions{})

	return err
}

func (c *Client) RestartDeployment(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	// Kubernetes has no native restart API; a timestamp annotation forces the
	// deployment controller to perform a rolling update (matches kubectl rollout restart).
	patch := []byte(
		fmt.Sprintf(
			`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`,
			metav1.Now().Format("2006-01-02T15:04:05Z"),
		),
	)

	_, err := c.clientset.AppsV1().
		Deployments(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})

	return err
}

func (c *Client) UpdateDeployment(
	ctx context.Context,
	deployment *appsv1.Deployment,
) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().
		Deployments(deployment.Namespace).
		Update(ctx, deployment, metav1.UpdateOptions{})
}

var (
	ErrNoPreviousRevision = errors.New("no previous revision found for rollback")
	ErrRevisionNotFound   = errors.New("revision not found")
)

// RevisionInfo holds metadata about a single deployment revision
// extracted from the owning ReplicaSet.
type RevisionInfo struct {
	Revision  int64
	Name      string
	CreatedAt metav1.Time
	Template  corev1.PodTemplateSpec
}

// getOwnedReplicaSets returns all ReplicaSets owned by the named deployment,
// sorted by revision number descending (newest first).
func (c *Client) getOwnedReplicaSets(
	ctx context.Context,
	namespace, name string,
) ([]appsv1.ReplicaSet, error) {
	deployment, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	rsList, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list replica sets: %w", err)
	}

	var ownedRS []appsv1.ReplicaSet

	for _, rs := range rsList.Items {
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.Kind == "Deployment" && ownerRef.Name == name {
				ownedRS = append(ownedRS, rs)

				break
			}
		}
	}

	sort.Slice(ownedRS, func(i, j int) bool {
		return getRevision(&ownedRS[i]) > getRevision(&ownedRS[j])
	})

	return ownedRS, nil
}

func (c *Client) RollbackDeployment(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	ownedRS, err := c.getOwnedReplicaSets(ctx, namespace, name)
	if err != nil {
		return err
	}

	if len(ownedRS) < 2 {
		return ErrNoPreviousRevision
	}

	deployment, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	previousRS := ownedRS[1]
	deployment.Spec.Template = previousRS.Spec.Template

	_, err = c.UpdateDeployment(ctx, deployment)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

// ListDeploymentRevisions returns revision metadata for all ReplicaSets
// owned by the deployment, sorted newest-first.
func (c *Client) ListDeploymentRevisions(
	ctx context.Context,
	namespace, name string,
) ([]RevisionInfo, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	ownedRS, err := c.getOwnedReplicaSets(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	revisions := make([]RevisionInfo, 0, len(ownedRS))

	for i := range ownedRS {
		rs := &ownedRS[i]
		revisions = append(revisions, RevisionInfo{
			Revision:  getRevision(rs),
			Name:      rs.Name,
			CreatedAt: rs.CreationTimestamp,
			Template:  rs.Spec.Template,
		})
	}

	return revisions, nil
}

// GetRevisionYAML marshals the PodTemplateSpec for a specific revision to YAML.
// Returns ErrRevisionNotFound if no matching revision exists.
func (c *Client) GetRevisionYAML(
	ctx context.Context,
	namespace, name string,
	revision int64,
) (string, error) {
	revisions, err := c.ListDeploymentRevisions(ctx, namespace, name)
	if err != nil {
		return "", err
	}

	for _, rev := range revisions {
		if rev.Revision == revision {
			data, marshalErr := yaml.Marshal(rev.Template)
			if marshalErr != nil {
				return "", fmt.Errorf("failed to marshal revision template: %w", marshalErr)
			}

			return string(data), nil
		}
	}

	return "", ErrRevisionNotFound
}

func getRevision(rs *appsv1.ReplicaSet) int64 {
	if rs.Annotations == nil {
		return 0
	}

	revStr, ok := rs.Annotations["deployment.kubernetes.io/revision"]
	if !ok {
		return 0
	}

	rev, err := strconv.ParseInt(revStr, 10, 64)
	if err != nil {
		return 0
	}

	return rev
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
