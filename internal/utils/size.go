package utils

import (
	"fmt"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
	PB = TB * 1024
)

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

// FormatCPU converts millicores: values >= 1000m display as cores
// (e.g., "1.5"), otherwise as millicores (e.g., "100m").
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

func FormatMemory(bytes int64) string {
	return FormatBytesShort(bytes)
}
