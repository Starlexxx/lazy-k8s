package utils

import (
	"testing"
)

func TestFormatBytesShort(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 500, "500B"},
		{"kilobytes", 2 * KB, "2Ki"},
		{"megabytes", 5 * MB, "5Mi"},
		{"gigabytes", 3*GB + GB*7/10, "3.7Gi"},
		{"terabytes", 2*TB + TB/2, "2.5Ti"},
		{"petabytes", PB + PB*2/10, "1.2Pi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytesShort(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytesShort(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatCPU(t *testing.T) {
	tests := []struct {
		name       string
		millicores int64
		expected   string
	}{
		{"millicores less than 1000", 500, "500m"},
		{"exactly 1 core", 1000, "1"},
		{"fractional cores", 1500, "1.5"},
		{"multiple cores", 4000, "4"},
		{"zero", 0, "0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCPU(tt.millicores)
			if result != tt.expected {
				t.Errorf("FormatCPU(%d) = %q, want %q", tt.millicores, result, tt.expected)
			}
		})
	}
}

func TestFormatMemory(t *testing.T) {
	// FormatMemory is an alias for FormatBytesShort
	result := FormatMemory(5 * MB)

	expected := "5Mi"
	if result != expected {
		t.Errorf("FormatMemory(%d) = %q, want %q", 5*MB, result, expected)
	}
}
