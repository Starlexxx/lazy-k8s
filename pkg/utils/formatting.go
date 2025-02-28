package utils

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// FormatAge formats the timestamp as a user-friendly duration
func FormatAge(timestamp time.Time) string {
	duration := time.Since(timestamp)

	// Less than a minute
	if duration < time.Minute {
		return "just now"
	}

	// Less than an hour
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm", minutes)
	}

	// Less than a day
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh", hours)
	}

	// Less than a week
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}

	// Less than a month
	if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		return fmt.Sprintf("%dw", weeks)
	}

	// Months
	months := int(duration.Hours() / 24 / 30)
	return fmt.Sprintf("%dM", months)
}

// FormatLabels formats Kubernetes labels as a string (key1=value1,key2=value2)
func FormatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}

	formattedLabels := make([]string, 0, len(labels))
	for k, v := range labels {
		formattedLabels = append(formattedLabels, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(formattedLabels, ",")
}

// GetNodeConditionStatus extracts the node's ready condition status
func GetNodeConditionStatus(conditions []corev1.NodeCondition) string {
	for _, condition := range conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}

	return "Unknown"
}

// GetNodeRoles extracts the roles from node labels
func GetNodeRoles(labels map[string]string) string {
	const (
		labelNodeRolePrefix = "node-role.kubernetes.io/"
		labelNodeRole       = "kubernetes.io/role"
	)

	var roles []string

	// Check for the newer label format (node-role.kubernetes.io/<role>=)
	for label := range labels {
		if strings.HasPrefix(label, labelNodeRolePrefix) {
			role := strings.TrimPrefix(label, labelNodeRolePrefix)
			roles = append(roles, role)
		}
	}

	// Check for the older label format (kubernetes.io/role=<role>)
	if role, ok := labels[labelNodeRole]; ok {
		if !contains(roles, role) {
			roles = append(roles, role)
		}
	}

	// Special handling for master/control-plane
	if _, ok := labels["node-role.kubernetes.io/master"]; ok {
		if !contains(roles, "master") {
			roles = append(roles, "master")
		}
	}

	if _, ok := labels["node-role.kubernetes.io/control-plane"]; ok {
		if !contains(roles, "control-plane") {
			roles = append(roles, "control-plane")
		}
	}

	if len(roles) == 0 {
		return "<none>"
	}

	return strings.Join(roles, ",")
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
