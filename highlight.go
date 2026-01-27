package terma

// TextHighlight represents a styled region of text by grapheme index.
type TextHighlight struct {
	Start int       // Grapheme index (inclusive)
	End   int       // Grapheme index (exclusive)
	Style SpanStyle // Style to apply
}

// LineHighlight represents a styled line range (TextArea only).
type LineHighlight struct {
	StartLine int   // Line index (0-based, inclusive)
	EndLine   int   // Line index (exclusive), -1 means to end
	Style     Style // Full style (supports background color)
}

// Highlighter produces highlights for text content.
type Highlighter interface {
	Highlight(text string, graphemes []string) []TextHighlight
}

// HighlighterFunc adapts a function to the Highlighter interface.
type HighlighterFunc func(text string, graphemes []string) []TextHighlight

// Highlight implements the Highlighter interface.
func (f HighlighterFunc) Highlight(text string, graphemes []string) []TextHighlight {
	return f(text, graphemes)
}

// buildHighlightMap converts []TextHighlight to a per-grapheme lookup.
// Later highlights in the slice override earlier ones for overlapping ranges.
func buildHighlightMap(highlights []TextHighlight) map[int]SpanStyle {
	if len(highlights) == 0 {
		return nil
	}
	result := make(map[int]SpanStyle)
	for _, h := range highlights {
		for i := h.Start; i < h.End; i++ {
			result[i] = h.Style
		}
	}
	return result
}

// buildLineHighlightMap converts []LineHighlight to a per-line lookup.
// lineCount is the total number of lines in the content.
// Later highlights in the slice override earlier ones for overlapping lines.
func buildLineHighlightMap(highlights []LineHighlight, lineCount int) map[int]Style {
	if len(highlights) == 0 {
		return nil
	}
	result := make(map[int]Style)
	for _, h := range highlights {
		endLine := h.EndLine
		if endLine < 0 {
			endLine = lineCount
		}
		for i := h.StartLine; i < endLine && i < lineCount; i++ {
			result[i] = h.Style
		}
	}
	return result
}

// applySpanStyle merges a SpanStyle onto a base Style.
// Only non-zero values from the SpanStyle are applied.
func applySpanStyle(base Style, span SpanStyle) Style {
	result := base
	if span.Foreground.IsSet() {
		result.ForegroundColor = span.Foreground
	}
	if span.Background.IsSet() {
		result.BackgroundColor = span.Background
	}
	if span.Bold {
		result.Bold = true
	}
	if span.Faint {
		result.Faint = true
	}
	if span.Italic {
		result.Italic = true
	}
	if span.Underline != UnderlineNone {
		result.Underline = span.Underline
	}
	if span.UnderlineColor.IsSet() {
		result.UnderlineColor = span.UnderlineColor
	}
	if span.Blink {
		result.Blink = true
	}
	if span.Reverse {
		result.Reverse = true
	}
	if span.Conceal {
		result.Conceal = true
	}
	if span.Strikethrough {
		result.Strikethrough = true
	}
	return result
}
