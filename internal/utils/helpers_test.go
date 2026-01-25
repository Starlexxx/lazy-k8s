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

func TestCenter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"center short string", "hi", 6, "  hi  "},
		{"center odd width", "hi", 5, " hi  "},
		{"no centering needed", "hello", 5, "hello"},
		{"string longer than width", "hello world", 5, "hello world"},
		{"empty string", "", 4, "    "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Center(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("Center(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
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
				t.Errorf("WrapText(%q, %d) returned %d lines, want %d", tt.text, tt.width, len(result), len(tt.expected))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("WrapText(%q, %d)[%d] = %q, want %q", tt.text, tt.width, i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"item exists", []string{"a", "b", "c"}, "b", true},
		{"item not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"empty item in slice", []string{"a", "", "b"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("Contains(%v, %q) = %v, want %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}

func TestFilterStrings(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		predicate func(string) bool
		expected  []string
	}{
		{
			"filter by length",
			[]string{"a", "bb", "ccc", "dd"},
			func(s string) bool { return len(s) > 1 },
			[]string{"bb", "ccc", "dd"},
		},
		{
			"filter all",
			[]string{"a", "b", "c"},
			func(s string) bool { return false },
			[]string{},
		},
		{
			"keep all",
			[]string{"a", "b", "c"},
			func(s string) bool { return true },
			[]string{"a", "b", "c"},
		},
		{
			"empty slice",
			[]string{},
			func(s string) bool { return true },
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterStrings(tt.slice, tt.predicate)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterStrings returned %d items, want %d", len(result), len(tt.expected))
				return
			}
			for i, item := range result {
				if item != tt.expected[i] {
					t.Errorf("FilterStrings[%d] = %q, want %q", i, item, tt.expected[i])
				}
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{-1, -5, -1},
		{0, 0, 0},
		{100, 100, 100},
	}

	for _, tt := range tests {
		result := Max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{-1, -5, -5},
		{0, 0, 0},
		{100, 100, 100},
	}

	for _, tt := range tests {
		result := Min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected int
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		result := Clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestSplitNamespacedName(t *testing.T) {
	tests := []struct {
		input             string
		expectedNamespace string
		expectedName      string
	}{
		{"default/my-pod", "default", "my-pod"},
		{"kube-system/coredns", "kube-system", "coredns"},
		{"my-pod", "", "my-pod"},
		{"ns/name/with/slashes", "ns", "name/with/slashes"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ns, name := SplitNamespacedName(tt.input)
			if ns != tt.expectedNamespace || name != tt.expectedName {
				t.Errorf("SplitNamespacedName(%q) = (%q, %q), want (%q, %q)",
					tt.input, ns, name, tt.expectedNamespace, tt.expectedName)
			}
		})
	}
}

func TestJoinNamespacedName(t *testing.T) {
	tests := []struct {
		namespace string
		name      string
		expected  string
	}{
		{"default", "my-pod", "default/my-pod"},
		{"", "my-pod", "my-pod"},
		{"kube-system", "coredns", "kube-system/coredns"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := JoinNamespacedName(tt.namespace, tt.name)
			if result != tt.expected {
				t.Errorf("JoinNamespacedName(%q, %q) = %q, want %q",
					tt.namespace, tt.name, result, tt.expected)
			}
		})
	}
}
