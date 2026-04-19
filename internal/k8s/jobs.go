package k8s

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetJobPodSelector returns the Job's pod selector as a string usable with
// List(ListOptions{LabelSelector: ...}). The API server auto-populates
// Spec.Selector with a controller-uid label, so matching is reliable even for
// jobs that omit an explicit selector in their manifest.
func GetJobPodSelector(job *batchv1.Job) string {
	if job.Spec.Selector == nil {
		return ""
	}

	return metav1.FormatLabelSelector(job.Spec.Selector)
}
