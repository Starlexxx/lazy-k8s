package utils

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate with ellipsis", "hello world", 8, "hello..."},
		{"very short maxLen", "hello", 3, "hel"},
		{"maxLen less than 3", "hello", 2, "he"},
		{"empty string", "", 5, ""},
		{"maxLen zero", "hello", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"pad short string", "hi", 5, "hi   "},
		{"no padding needed", "hello", 5, "hello"},
		{"string longer than length", "hello world", 5, "hello world"},
		{"empty string", "", 3, "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadRight(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"pad short string", "hi", 5, "   hi"},
		{"no padding needed", "hello", 5, "hello"},
		{"string longer than length", "hello world", 5, "hello world"},
		{"empty string", "", 3, "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadLeft(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("PadLeft(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{"simple wrap", "hello world foo", 12, []string{"hello world", "foo"}},
		{"no wrap needed", "hello", 20, []string{"hello"}},
		{"wrap multiple lines", "a b c d e f", 3, []string{"a b", "c d", "e f"}},
		{"empty text", "", 10, []string{}},
		{"zero width", "hello world", 0, []string{"hello world"}},
		{"negative width", "hello world", -1, []string{"hello world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.width)
			if len(result) != len(tt.expected) {
				t.Errorf(
					"WrapText(%q, %d) returned %d lines, want %d",
					tt.text,
					tt.width,
					len(result),
					len(tt.expected),
				)

				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf(
						"WrapText(%q, %d)[%d] = %q, want %q",
						tt.text,
						tt.width,
						i,
						line,
						tt.expected[i],
					)
				}
			}
		})
	}
}
