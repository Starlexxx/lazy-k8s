package utils

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
	PB = TB * 1024
)

// FormatBytes formats a byte count to a human-readable string
func FormatBytes(bytes int64) string {
	switch {
	case bytes >= PB:
		return fmt.Sprintf("%.2fPi", float64(bytes)/float64(PB))
	case bytes >= TB:
		return fmt.Sprintf("%.2fTi", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2fGi", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2fMi", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2fKi", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// FormatBytesShort formats bytes with fewer decimal places
func FormatBytesShort(bytes int64) string {
	switch {
	case bytes >= PB:
		return fmt.Sprintf("%.1fPi", float64(bytes)/float64(PB))
	case bytes >= TB:
		return fmt.Sprintf("%.1fTi", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1fGi", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0fMi", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0fKi", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// ParseQuantity parses a Kubernetes resource quantity string
func ParseQuantity(s string) (int64, error) {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return 0, err
	}
	return q.Value(), nil
}

// FormatQuantity formats a resource.Quantity to a human-readable string
func FormatQuantity(q resource.Quantity) string {
	return q.String()
}

// FormatCPU formats CPU quantity (in millicores) to a human-readable string
func FormatCPU(millicores int64) string {
	if millicores >= 1000 {
		cores := float64(millicores) / 1000
		if cores == float64(int64(cores)) {
			return fmt.Sprintf("%d", int64(cores))
		}
		return fmt.Sprintf("%.1f", cores)
	}
	return fmt.Sprintf("%dm", millicores)
}

// FormatMemory formats memory quantity to a human-readable string
func FormatMemory(bytes int64) string {
	return FormatBytesShort(bytes)
}

// ParseCPU parses a CPU string like "100m" or "2" to millicores
func ParseCPU(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "m") {
		val, err := strconv.ParseInt(strings.TrimSuffix(s, "m"), 10, 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(val * 1000), nil
}

// ParseMemory parses a memory string like "128Mi" or "1Gi" to bytes
func ParseMemory(s string) (int64, error) {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return 0, err
	}
	return q.Value(), nil
}

// FormatPercentage formats a float as a percentage string
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

// CalculatePercentage calculates percentage of used/total
func CalculatePercentage(used, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total)
}

// FormatResourceUsage formats resource usage as "used/total (percentage)"
func FormatResourceUsage(used, total int64, formatter func(int64) string) string {
	percentage := CalculatePercentage(used, total) * 100
	return fmt.Sprintf("%s/%s (%.1f%%)", formatter(used), formatter(total), percentage)
}
