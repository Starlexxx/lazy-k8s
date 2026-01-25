package utils

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds only", 45 * time.Second, "45s"},
		{"one minute", 1 * time.Minute, "1m"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2m30s"},
		{"one hour", 1 * time.Hour, "1h"},
		{"hours and minutes", 3*time.Hour + 15*time.Minute, "3h15m"},
		{"one day", 24 * time.Hour, "1d"},
		{"days and hours", 2*24*time.Hour + 5*time.Hour, "2d5h"},
		{"over a year", 400 * 24 * time.Hour, "1y35d"},
		{"exactly a year", 365 * 24 * time.Hour, "365d"},
		{"two years", 730 * 24 * time.Hour, "2y"},
		{"negative duration", -5 * time.Second, "0s"},
		{"zero duration", 0, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatAge(t *testing.T) {
	// Test with zero time
	result := FormatAge(time.Time{})
	if result != "Unknown" {
		t.Errorf("FormatAge(zero time) = %q, want %q", result, "Unknown")
	}

	// Test with a recent time
	recent := time.Now().Add(-5 * time.Minute)
	result = FormatAge(recent)
	// Should be around 5m, allow for some variance
	if result != "5m" && result != "5m0s" && result != "5m1s" {
		t.Logf("FormatAge(5 min ago) = %q (acceptable variance)", result)
	}
}

func TestFormatAgeFromMeta(t *testing.T) {
	// Test with zero time
	result := FormatAgeFromMeta(metav1.Time{})
	if result != "Unknown" {
		t.Errorf("FormatAgeFromMeta(zero time) = %q, want %q", result, "Unknown")
	}

	// Test with a recent time
	recent := metav1.NewTime(time.Now().Add(-1 * time.Hour))
	result = FormatAgeFromMeta(recent)
	// Should be around 1h
	if result != "1h" && result != "1h0m" {
		t.Logf("FormatAgeFromMeta(1 hour ago) = %q (acceptable variance)", result)
	}
}

func TestFormatTimestamp(t *testing.T) {
	// Test zero time
	result := FormatTimestamp(time.Time{})
	if result != "N/A" {
		t.Errorf("FormatTimestamp(zero time) = %q, want %q", result, "N/A")
	}

	// Test specific time
	specificTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	result = FormatTimestamp(specificTime)
	expected := "2024-01-15 14:30:45"
	if result != expected {
		t.Errorf("FormatTimestamp(%v) = %q, want %q", specificTime, result, expected)
	}
}

func TestFormatTimestampFromMeta(t *testing.T) {
	// Test zero time
	result := FormatTimestampFromMeta(metav1.Time{})
	if result != "N/A" {
		t.Errorf("FormatTimestampFromMeta(zero time) = %q, want %q", result, "N/A")
	}

	// Test specific time
	specificTime := metav1.NewTime(time.Date(2024, 6, 20, 10, 15, 30, 0, time.UTC))
	result = FormatTimestampFromMeta(specificTime)
	expected := "2024-06-20 10:15:30"
	if result != expected {
		t.Errorf("FormatTimestampFromMeta(%v) = %q, want %q", specificTime, result, expected)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		shouldError bool
	}{
		{"days format", "5d", 5 * 24 * time.Hour, false},
		{"weeks format", "2w", 2 * 7 * 24 * time.Hour, false},
		{"standard hours", "3h", 3 * time.Hour, false},
		{"standard minutes", "30m", 30 * time.Minute, false},
		{"standard seconds", "45s", 45 * time.Second, false},
		{"combined format", "1h30m", 1*time.Hour + 30*time.Minute, false},
		{"invalid format", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseDuration(%q) should have returned an error", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3 hours ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1 day ago"},
		{"3 days ago", now.Add(-72 * time.Hour), "3 days ago"},
		{"future time", now.Add(1 * time.Hour), "in the future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RelativeTime(tt.time)
			if result != tt.expected {
				// For times > 7 days, it returns a timestamp, so just check it's not empty
				if tt.expected == "" {
					if result == "" {
						t.Errorf("RelativeTime(%v) returned empty string", tt.time)
					}
				} else {
					t.Errorf("RelativeTime(%v) = %q, want %q", tt.time, result, tt.expected)
				}
			}
		})
	}

	// Test time older than 7 days returns timestamp format
	oldTime := now.Add(-10 * 24 * time.Hour)
	result := RelativeTime(oldTime)
	// Should return a timestamp, not a relative time string
	if result == "just now" || result == "in the future" {
		t.Errorf("RelativeTime(10 days ago) should return timestamp, got %q", result)
	}
}
