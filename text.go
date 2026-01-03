package terma

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// WrapMode defines how text should wrap within available width.
type WrapMode int

const (
	// WrapSoft breaks at word boundaries (spaces), only breaking words if necessary (default).
	WrapSoft WrapMode = iota
	// WrapHard breaks at exact character boundary when line exceeds width.
	WrapHard
	// WrapNone disables wrapping - text is truncated if too long.
	WrapNone
)

// Text is a leaf widget that displays text content.
type Text struct {
	ID      string     // Optional unique identifier for the widget
	Content string     // Plain text (used if Spans is empty)
	Spans   []Span     // Rich text segments (takes precedence if non-empty)
	Width   Dimension  // Optional width (zero value = auto)
	Height  Dimension  // Optional height (zero value = auto)
	Wrap    WrapMode   // Wrapping mode (default = WrapSoft)
	Style   Style      // Optional styling (colors, inherited by spans)
	Click   func()     // Optional callback invoked when clicked
	Hover   func(bool) // Optional callback invoked when hover state changes
}

// Build returns itself as Text is a leaf widget.
func (t Text) Build(ctx BuildContext) Widget {
	return t
}

// WidgetID returns the text widget's unique identifier.
// Implements the Identifiable interface.
func (t Text) WidgetID() string {
	return t.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (t Text) OnClick() {
	if t.Click != nil {
		t.Click()
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (t Text) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// GetDimensions returns the width and height dimension preferences.
func (t Text) GetDimensions() (width, height Dimension) {
	return t.Width, t.Height
}

// GetStyle returns the style of the text widget.
func (t Text) GetStyle() Style {
	return t.Style
}

// textContent returns the effective text content.
// If Spans is non-empty, concatenates all span text; otherwise returns Content.
func (t Text) textContent() string {
	if len(t.Spans) > 0 {
		var sb strings.Builder
		for _, span := range t.Spans {
			sb.WriteString(span.Text)
		}
		return sb.String()
	}
	return t.Content
}

// wrapText wraps the given text content to fit within maxWidth based on the wrap mode.
func wrapText(content string, maxWidth int, mode WrapMode) []string {
	if maxWidth <= 0 || mode == WrapNone {
		return strings.Split(content, "\n")
	}

	inputLines := strings.Split(content, "\n")
	var result []string

	for _, line := range inputLines {
		lineWidth := ansi.StringWidth(line)

		// If line fits, add as-is
		if lineWidth <= maxWidth {
			result = append(result, line)
			continue
		}

		var wrapped string
		switch mode {
		case WrapHard:
			wrapped = ansi.Wrap(line, maxWidth, "")
		case WrapSoft:
			wrapped = ansi.Wordwrap(line, maxWidth, "")
			// Check if any resulting line exceeds maxWidth (long word scenario)
			wrappedLines := strings.Split(wrapped, "\n")
			var finalLines []string
			for _, wl := range wrappedLines {
				if ansi.StringWidth(wl) > maxWidth {
					// Word longer than maxWidth, hard-break it
					hardWrapped := ansi.Wrap(wl, maxWidth, "")
					finalLines = append(finalLines, strings.Split(hardWrapped, "\n")...)
				} else {
					finalLines = append(finalLines, wl)
				}
			}
			result = append(result, finalLines...)
			continue
		}

		result = append(result, strings.Split(wrapped, "\n")...)
	}

	return result
}

// Layout computes the size of the text widget.
func (t Text) Layout(ctx BuildContext, constraints Constraints) Size {
	content := t.textContent()

	// Determine the width we'll use for wrapping
	var wrapWidth int
	switch {
	case t.Width.IsCells():
		wrapWidth = t.Width.CellsValue()
	case t.Width.IsFr():
		wrapWidth = constraints.MaxWidth
	default: // Auto
		wrapWidth = constraints.MaxWidth
	}

	// Get lines (wrapped or not based on mode)
	lines := wrapText(content, wrapWidth, t.Wrap)

	naturalHeight := len(lines)
	naturalWidth := 0
	for _, line := range lines {
		lineWidth := ansi.StringWidth(line)
		if lineWidth > naturalWidth {
			naturalWidth = lineWidth
		}
	}

	// Determine width based on dimension type
	var width int
	switch {
	case t.Width.IsCells():
		width = t.Width.CellsValue()
	case t.Width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		width = naturalWidth
	}

	// Determine height based on dimension type
	var height int
	switch {
	case t.Height.IsCells():
		height = t.Height.CellsValue()
	case t.Height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		height = naturalHeight
	}

	// Clamp to constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}

	return Size{Width: width, Height: height}
}

// Render draws the text to the render context.
func (t Text) Render(ctx *RenderContext) {
	if len(t.Spans) > 0 {
		t.renderSpans(ctx)
	} else {
		t.renderPlain(ctx)
	}
}

// renderPlain renders plain text content.
func (t Text) renderPlain(ctx *RenderContext) {
	fgColor := t.Style.ForegroundColor
	if !fgColor.IsSet() {
		fgColor = ctx.buildContext.Theme().Text
	}
	style := Style{ForegroundColor: fgColor}

	// Get lines with wrapping applied
	lines := wrapText(t.Content, ctx.Width, t.Wrap)

	for i := 0; i < ctx.Height; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}
		// Truncate line if it exceeds width (fallback for WrapNone or edge cases)
		lineWidth := ansi.StringWidth(line)
		if lineWidth > ctx.Width {
			line = ansi.Truncate(line, ctx.Width, "")
			lineWidth = ctx.Width
		}
		// Pad line to fill the full width (for background colors)
		if lineWidth < ctx.Width {
			line = line + strings.Repeat(" ", ctx.Width-lineWidth)
		}
		ctx.DrawStyledText(0, i, line, style)
	}
}

// renderSpans renders rich text with multiple styled spans.
func (t Text) renderSpans(ctx *RenderContext) {
	fgColor := t.Style.ForegroundColor
	if !fgColor.IsSet() {
		fgColor = ctx.buildContext.Theme().Text
	}
	baseStyle := Style{ForegroundColor: fgColor}

	x, y := 0, 0

	for _, span := range t.Spans {
		parts := strings.Split(span.Text, "\n")
		for partIdx, part := range parts {
			// Handle explicit newline
			if partIdx > 0 {
				x = 0
				y++
				if y >= ctx.Height {
					return
				}
			}

			if len(part) == 0 {
				continue
			}

			// Process this part with wrapping
			remaining := part
			for len(remaining) > 0 {
				if y >= ctx.Height {
					return
				}

				availableWidth := ctx.Width - x
				if availableWidth <= 0 {
					x = 0
					y++
					availableWidth = ctx.Width
					if y >= ctx.Height {
						return
					}
				}

				partWidth := ansi.StringWidth(remaining)

				// If it fits or no wrapping, draw and continue
				if partWidth <= availableWidth || t.Wrap == WrapNone {
					chunk := remaining
					if partWidth > availableWidth {
						chunk = ansi.Truncate(remaining, availableWidth, "")
					}
					if len(chunk) > 0 {
						partSpan := Span{Text: chunk, Style: span.Style}
						ctx.DrawSpan(x, y, partSpan, baseStyle)
						x += ansi.StringWidth(chunk)
					}
					break
				}

				// Need to wrap - find break point
				chunk, rest := t.findWrapPoint(remaining, availableWidth)

				if len(chunk) > 0 {
					partSpan := Span{Text: chunk, Style: span.Style}
					ctx.DrawSpan(x, y, partSpan, baseStyle)
				}

				remaining = rest
				if len(remaining) > 0 {
					x = 0
					y++
				}
			}
		}
	}
}

// findWrapPoint finds where to break text for wrapping, returning the chunk to render
// and the remaining text.
func (t Text) findWrapPoint(text string, availableWidth int) (chunk, remaining string) {
	if t.Wrap == WrapHard {
		chunk = ansi.Truncate(text, availableWidth, "")
		remaining = text[len(chunk):]
		return
	}

	// Soft wrap: find last space within available width
	chunk = ansi.Truncate(text, availableWidth, "")
	lastSpace := strings.LastIndex(chunk, " ")

	if lastSpace > 0 {
		chunk = chunk[:lastSpace]
		remaining = strings.TrimPrefix(text[lastSpace:], " ")
	} else {
		// No space found, must break the word (fallback to hard wrap)
		remaining = text[len(chunk):]
	}
	return
}
