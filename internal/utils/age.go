package utils

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FormatAge formats a time.Time to a human-readable age string (e.g., "2d", "3h", "45m")
func FormatAge(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return FormatDuration(time.Since(t))
}

// FormatAgeFromMeta formats a metav1.Time to a human-readable age string
func FormatAgeFromMeta(t metav1.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return FormatDuration(time.Since(t.Time))
}

// FormatDuration formats a duration to a human-readable string
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

// FormatTimestamp formats a time.Time to a standard timestamp string
func FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}

// FormatTimestampFromMeta formats a metav1.Time to a standard timestamp string
func FormatTimestampFromMeta(t metav1.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

// ParseDuration parses a string duration like "30m", "1h", "2d"
func ParseDuration(s string) (time.Duration, error) {
	// Handle day format
	if len(s) > 0 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}

	// Handle week format
	if len(s) > 0 && s[len(s)-1] == 'w' {
		var weeks int
		if _, err := fmt.Sscanf(s, "%dw", &weeks); err == nil {
			return time.Duration(weeks) * 7 * 24 * time.Hour, nil
		}
	}

	// Fall back to standard Go duration parsing
	return time.ParseDuration(s)
}

// RelativeTime returns a human-readable relative time string
func RelativeTime(t time.Time) string {
	now := time.Now()
	if t.After(now) {
		return "in the future"
	}

	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return FormatTimestamp(t)
	}
}
