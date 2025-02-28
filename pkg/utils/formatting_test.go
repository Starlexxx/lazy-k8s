package utils

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "less than a minute",
			duration: 30 * time.Second,
			expected: "just now",
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "hours",
			duration: 3 * time.Hour,
			expected: "3h",
		},
		{
			name:     "days",
			duration: 3 * 24 * time.Hour,
			expected: "3d",
		},
		{
			name:     "weeks",
			duration: 2 * 7 * 24 * time.Hour,
			expected: "2w",
		},
		{
			name:     "months",
			duration: 2 * 30 * 24 * time.Hour,
			expected: "2M",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			timestamp := time.Now().Add(-tc.duration)
			result := FormatAge(timestamp)
			if result != tc.expected {
				t.Errorf("FormatAge(%v) = %s, wanted %s", tc.duration, result, tc.expected)
			}
		})
	}
}

func TestFormatLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: "<none>",
		},
		{
			name: "one label",
			labels: map[string]string{
				"app": "nginx",
			},
			expected: "app=nginx",
		},
		{
			name: "multiple labels",
			labels: map[string]string{
				"app":         "nginx",
				"environment": "production",
				"tier":        "frontend",
			},
			expected: "app=nginx,environment=production,tier=frontend",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatLabels(tc.labels)
			// Check all possible combinations, as the order of labels in a map is not guaranteed
			allPossible := []string{
				"app=nginx,environment=production,tier=frontend",
				"app=nginx,tier=frontend,environment=production",
				"environment=production,app=nginx,tier=frontend",
				"environment=production,tier=frontend,app=nginx",
				"tier=frontend,app=nginx,environment=production",
				"tier=frontend,environment=production,app=nginx",
			}

			if tc.name == "multiple labels" {
				found := false
				for _, possible := range allPossible {
					if result == possible {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FormatLabels() = %s, wanted one of the combinations", result)
				}
			} else if result != tc.expected {
				t.Errorf("FormatLabels() = %s, wanted %s", result, tc.expected)
			}
		})
	}
}

func TestGetNodeConditionStatus(t *testing.T) {
	tests := []struct {
		name       string
		conditions []corev1.NodeCondition
		expected   string
	}{
		{
			name:       "empty conditions",
			conditions: []corev1.NodeCondition{},
			expected:   "Unknown",
		},
		{
			name: "node ready",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
			expected: "Ready",
		},
		{
			name: "node not ready",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			},
			expected: "NotReady",
		},
		{
			name: "node with multiple conditions",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeDiskPressure,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
			expected: "Ready",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetNodeConditionStatus(tc.conditions)
			if result != tc.expected {
				t.Errorf("GetNodeConditionStatus() = %s, wanted %s", result, tc.expected)
			}
		})
	}
}

func TestGetNodeRoles(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: "<none>",
		},
		{
			name: "node with master role",
			labels: map[string]string{
				"node-role.kubernetes.io/master": "",
			},
			expected: "master",
		},
		{
			name: "node with control-plane role",
			labels: map[string]string{
				"node-role.kubernetes.io/control-plane": "",
			},
			expected: "control-plane",
		},
		{
			name: "node with multiple roles",
			labels: map[string]string{
				"node-role.kubernetes.io/master":        "",
				"node-role.kubernetes.io/control-plane": "",
				"node-role.kubernetes.io/worker":        "",
			},
			expected: "master,control-plane,worker",
		},
		{
			name: "node with old role format",
			labels: map[string]string{
				"kubernetes.io/role": "worker",
			},
			expected: "worker",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetNodeRoles(tc.labels)

			// The order of roles may be different, as it depends on the map iteration order
			if tc.name == "node with multiple roles" {
				possibleResults := []string{
					"master,control-plane,worker",
					"master,worker,control-plane",
					"control-plane,master,worker",
					"control-plane,worker,master",
					"worker,master,control-plane",
					"worker,control-plane,master",
				}

				found := false
				for _, possible := range possibleResults {
					if result == possible {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("GetNodeRoles() = %s, wanted one of the combinations", result)
				}
			} else if result != tc.expected {
				t.Errorf("GetNodeRoles() = %s, wanted %s", result, tc.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		str      string
		expected bool
	}{
		{
			name:     "empty slice",
			slice:    []string{},
			str:      "test",
			expected: false,
		},
		{
			name:     "slice contains string",
			slice:    []string{"one", "test", "three"},
			str:      "test",
			expected: true,
		},
		{
			name:     "slice does not contain string",
			slice:    []string{"one", "two", "three"},
			str:      "test",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := contains(tc.slice, tc.str)
			if result != tc.expected {
				t.Errorf("contains() = %v, wanted %v", result, tc.expected)
			}
		})
	}
}
