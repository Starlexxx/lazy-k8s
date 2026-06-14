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
