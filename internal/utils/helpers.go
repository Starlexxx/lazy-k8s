package utils

import (
	"fmt"
	"strings"
)

// Truncate truncates a string to the specified length, adding "..." if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

// PadRight pads a string with spaces on the right to reach the desired length.
func PadRight(s string, length int) string {
	if len(s) >= length {
		return s
	}

	return s + strings.Repeat(" ", length-len(s))
}

// PadLeft pads a string with spaces on the left to reach the desired length.
func PadLeft(s string, length int) string {
	if len(s) >= length {
		return s
	}

	return strings.Repeat(" ", length-len(s)) + s
}

// Center centers a string within the given width.
func Center(s string, width int) string {
	if len(s) >= width {
		return s
	}

	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// WrapText wraps text to fit within the specified width.
func WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string

	words := strings.Fields(text)
	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	lines = append(lines, currentLine)

	return lines
}

// Contains checks if a string slice contains a specific string.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

// FilterStrings filters a slice of strings based on a predicate.
func FilterStrings(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0)

	for _, s := range slice {
		if predicate(s) {
			result = append(result, s)
		}
	}

	return result
}

// Max returns the maximum of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Min returns the minimum of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// Clamp constrains a value to the given range.
func Clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}

	if value > maxVal {
		return maxVal
	}

	return value
}

// SplitNamespacedName splits a "namespace/name" string.
func SplitNamespacedName(namespacedName string) (namespace, name string) {
	parts := strings.SplitN(namespacedName, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "", parts[0]
}

// JoinNamespacedName joins namespace and name into "namespace/name" format.
func JoinNamespacedName(namespace, name string) string {
	if namespace == "" {
		return name
	}

	return fmt.Sprintf("%s/%s", namespace, name)
}
