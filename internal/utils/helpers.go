package utils

import (
	"fmt"
	"strings"
)

// Truncate appends "..." if s exceeds maxLen. If maxLen <= 3, hard-cuts without ellipsis.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

func PadRight(s string, length int) string {
	if len(s) >= length {
		return s
	}

	return s + strings.Repeat(" ", length-len(s))
}

func PadLeft(s string, length int) string {
	if len(s) >= length {
		return s
	}

	return strings.Repeat(" ", length-len(s)) + s
}

func Center(s string, width int) string {
	if len(s) >= width {
		return s
	}

	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

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

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

func FilterStrings(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0)

	for _, s := range slice {
		if predicate(s) {
			result = append(result, s)
		}
	}

	return result
}

func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func Clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}

	if value > maxVal {
		return maxVal
	}

	return value
}

// SplitNamespacedName returns empty namespace if no "/" is present.
func SplitNamespacedName(namespacedName string) (namespace, name string) {
	parts := strings.SplitN(namespacedName, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "", parts[0]
}

// JoinNamespacedName returns just the name if namespace is empty.
func JoinNamespacedName(namespace, name string) string {
	if namespace == "" {
		return name
	}

	return fmt.Sprintf("%s/%s", namespace, name)
}
