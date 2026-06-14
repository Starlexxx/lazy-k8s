package utils

import (
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
