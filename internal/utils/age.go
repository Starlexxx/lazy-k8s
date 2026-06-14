package utils

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FormatAge returns a compact age like "2d", "3h", or "45m" relative to now.
func FormatAge(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}

	return FormatDuration(time.Since(t))
}

func FormatAgeFromMeta(t metav1.Time) string {
	if t.IsZero() {
		return "Unknown"
	}

	return FormatDuration(time.Since(t.Time))
}

// FormatDuration shows at most two units of precision for readability (e.g., "2d3h", "45m12s").
func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 365 {
		years := days / 365

		remainingDays := days % 365
		if remainingDays > 0 {
			return fmt.Sprintf("%dy%dd", years, remainingDays)
		}

		return fmt.Sprintf("%dy", years)
	}

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}

		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}

		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		if seconds > 0 {
			return fmt.Sprintf("%dm%ds", minutes, seconds)
		}

		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%ds", seconds)
}

func FormatTimestampFromMeta(t metav1.Time) string {
	if t.IsZero() {
		return "N/A"
	}

	return t.Format("2006-01-02 15:04:05")
}
