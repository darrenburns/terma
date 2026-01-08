package layout

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// WrapMode controls text wrapping behavior.
type WrapMode int

const (
	// WrapNone disables wrapping - text may overflow available width.
	WrapNone WrapMode = iota
	// WrapWord wraps at word boundaries (spaces).
	// Falls back to character breaks for words longer than available width.
	WrapWord
	// WrapChar wraps at exact character boundaries.
	// Useful for CJK text or when precise character-level wrapping is needed.
	WrapChar
)

// MeasureText measures text content and returns content-box dimensions.
// It handles wrapping based on WrapMode and available width.
//
// Parameters:
//   - content: The text to measure
//   - wrap: Wrapping mode (WrapNone, WrapWord, WrapChar)
//   - maxWidth: Maximum width in cells (0 or negative = unbounded)
//
// Returns:
//   - width: The width of the widest line (in cells)
//   - height: Number of lines (after wrapping)
func MeasureText(content string, wrap WrapMode, maxWidth int) (width, height int) {
	if content == "" {
		return 0, 0
	}

	lines := wrapTextLines(content, wrap, maxWidth)

	height = len(lines)
	for _, line := range lines {
		lineWidth := ansi.StringWidth(line)
		if lineWidth > width {
			width = lineWidth
		}
	}

	return width, height
}

// wrapTextLines splits and wraps text into lines based on wrap mode and max width.
func wrapTextLines(content string, wrap WrapMode, maxWidth int) []string {
	// Split by explicit newlines first
	inputLines := strings.Split(content, "\n")

	// If no wrapping or unbounded, just return input lines
	if wrap == WrapNone || maxWidth <= 0 {
		return inputLines
	}

	var result []string
	for _, line := range inputLines {
		// If line fits, add as-is
		if ansi.StringWidth(line) <= maxWidth {
			result = append(result, line)
			continue
		}

		// Need to wrap this line
		switch wrap {
		case WrapChar:
			result = append(result, wrapLineByChar(line, maxWidth)...)
		case WrapWord:
			result = append(result, wrapLineByWord(line, maxWidth)...)
		}
	}

	return result
}

// wrapLineByChar wraps a single line at character boundaries.
func wrapLineByChar(line string, maxWidth int) []string {
	var result []string
	remaining := line

	for len(remaining) > 0 {
		if ansi.StringWidth(remaining) <= maxWidth {
			result = append(result, remaining)
			break
		}

		// Truncate to fit width
		chunk := ansi.Truncate(remaining, maxWidth, "")
		result = append(result, chunk)
		remaining = remaining[len(chunk):]
	}

	return result
}

// wrapLineByWord wraps a single line at word boundaries.
// Falls back to character breaks for words longer than maxWidth.
func wrapLineByWord(line string, maxWidth int) []string {
	// Use ansi.Wordwrap for word-boundary wrapping
	wrapped := ansi.Wordwrap(line, maxWidth, "")
	wrappedLines := strings.Split(wrapped, "\n")

	var result []string
	for _, wl := range wrappedLines {
		// If a word is longer than maxWidth, we need to break it
		if ansi.StringWidth(wl) > maxWidth {
			result = append(result, wrapLineByChar(wl, maxWidth)...)
		} else {
			result = append(result, wl)
		}
	}

	return result
}
