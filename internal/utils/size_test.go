package utils

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 500, "500B"},
		{"kilobytes", 2 * KB, "2.00Ki"},
		{"megabytes", 5 * MB, "5.00Mi"},
		{"gigabytes", 3 * GB, "3.00Gi"},
		{"terabytes", 2 * TB, "2.00Ti"},
		{"petabytes", 1 * PB, "1.00Pi"},
		{"fractional GB", int64(1.5 * float64(GB)), "1.50Gi"},
		{"zero", 0, "0B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

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

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		shouldError bool
	}{
		{"kilobytes", "1Ki", 1024, false},
		{"megabytes", "1Mi", 1024 * 1024, false},
		{"gigabytes", "1Gi", 1024 * 1024 * 1024, false},
		{"plain number", "1000", 1000, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseQuantity(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseQuantity(%q) should have returned an error", tt.input)
				}

				return
			}

			if err != nil {
				t.Errorf("ParseQuantity(%q) returned unexpected error: %v", tt.input, err)

				return
			}

			if result != tt.expected {
				t.Errorf("ParseQuantity(%q) = %d, want %d", tt.input, result, tt.expected)
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

func TestParseCPU(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		shouldError bool
	}{
		{"millicores", "500m", 500, false},
		{"cores as integer", "2", 2000, false},
		{"cores as float", "1.5", 1500, false},
		{"with whitespace", "  100m  ", 100, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCPU(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseCPU(%q) should have returned an error", tt.input)
				}

				return
			}

			if err != nil {
				t.Errorf("ParseCPU(%q) returned unexpected error: %v", tt.input, err)

				return
			}

			if result != tt.expected {
				t.Errorf("ParseCPU(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		shouldError bool
	}{
		{"megabytes", "128Mi", 128 * 1024 * 1024, false},
		{"gigabytes", "1Gi", 1024 * 1024 * 1024, false},
		{"kilobytes", "512Ki", 512 * 1024, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMemory(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseMemory(%q) should have returned an error", tt.input)
				}

				return
			}

			if err != nil {
				t.Errorf("ParseMemory(%q) returned unexpected error: %v", tt.input, err)

				return
			}

			if result != tt.expected {
				t.Errorf("ParseMemory(%q) = %d, want %d", tt.input, result, tt.expected)
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

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"50 percent", 0.5, "50.0%"},
		{"100 percent", 1.0, "100.0%"},
		{"zero percent", 0.0, "0.0%"},
		{"fractional percent", 0.333, "33.3%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercentage(tt.value)
			if result != tt.expected {
				t.Errorf("FormatPercentage(%f) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		name     string
		used     int64
		total    int64
		expected float64
	}{
		{"half used", 50, 100, 0.5},
		{"all used", 100, 100, 1.0},
		{"none used", 0, 100, 0.0},
		{"zero total", 50, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePercentage(tt.used, tt.total)
			if result != tt.expected {
				t.Errorf(
					"CalculatePercentage(%d, %d) = %f, want %f",
					tt.used,
					tt.total,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestFormatResourceUsage(t *testing.T) {
	formatter := func(v int64) string { return FormatBytes(v) }

	result := FormatResourceUsage(512*MB, 1*GB, formatter)

	expected := "512.00Mi/1.00Gi (50.0%)"
	if result != expected {
		t.Errorf("FormatResourceUsage = %q, want %q", result, expected)
	}

	// Test with zero total
	result = FormatResourceUsage(100, 0, formatter)

	expected = "100B/0B (0.0%)"
	if result != expected {
		t.Errorf("FormatResourceUsage with zero total = %q, want %q", result, expected)
	}
}
